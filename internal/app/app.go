package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"todo/internal/auth"
	"todo/internal/config"
	"todo/internal/middlewares"
	"todo/internal/todo"
	"todo/internal/users"
	"todo/pkg/database/postgres"
	"todo/pkg/database/redis"
	"todo/pkg/security"

	"github.com/go-chi/chi/v5"
	redisLib "github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	server *http.Server
	db     *pgxpool.Pool
	redis  *redisLib.Client
}

func setupRedis(cfg *config.Config) (*redisLib.Client, error) {
	redisClient, err := redis.NewRedisClient(cfg.Redis.ConnTimeout, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.Username, cfg.Redis.DB, cfg.Redis.MinIdleConns, cfg.Redis.PoolSize, cfg.Redis.ConnMaxLifetime)

	if err != nil {
		return nil, fmt.Errorf("failed to setup redis: %w", err)
	}

	return redisClient, nil
}

func setupDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	client, err := postgres.NewClient(cfg.Database.ConnTimeout, fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name), cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.MaxConnLifetime)

	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	err = postgres.SetMigrations(fmt.Sprintf("pgx5://%s:%s@%s:%d/%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to setup migrations: %w", err)
	}

	return client, nil
}

func setupRouter(cfg *config.Config, DbClient *pgxpool.Pool, redisClient *redisLib.Client) *chi.Mux {
	jwt := security.NewJWTManager(cfg.JWT)
	hasher := security.NewHasher(cfg.Hash)

	taskRepo := todo.NewTaskRepository(DbClient, cfg.App)
	cacheTaskRepo := todo.NewCacheTaskRepository(taskRepo, redisClient, cfg.Cache)
	usersRepo := users.NewUsersRepository(DbClient)
	whitelistRepo := auth.NewWhitelistRepository(redisClient, cfg.JWT)

	taskService := todo.NewTaskService(cacheTaskRepo)
	usersService := auth.NewAuthService(usersRepo, hasher, jwt, whitelistRepo)
	profileService := users.NewProfileService(cacheTaskRepo, cfg.App)

	taskHandler := todo.NewTaskHandler(taskService, cfg.HTTP)
	usersHandler := auth.NewAuthHandler(usersService, cfg.JWT, cfg.HTTP)
	profileHandler := users.NewProfileHandler(profileService, cfg.HTTP)

	authMiddleware := middlewares.NewAuthMiddleware(jwt)
	corsMiddleware := middlewares.NewCORSMiddleware(cfg.HTTP.CORSUrl)

	mux := NewRouter(taskHandler, usersHandler, authMiddleware, corsMiddleware, profileHandler)

	return mux
}

func New(cfg *config.Config) (*App, error) {

	client, err := setupDatabase(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	redisClient, err := setupRedis(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to setup redis: %w", err)
	}

	mux := setupRouter(cfg, client, redisClient)

	return &App{
		server: &http.Server{
			Addr:         cfg.HTTP.Addr,
			Handler:      mux,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		db:    client,
		redis: redisClient,
	}, nil

}

func (a *App) Run() error {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("app startup failed: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed close http requests: %w", err)
	}
	a.db.Close()

	if err := a.redis.Close(); err != nil {
		return fmt.Errorf("failed close redis %w", err)
	}

	return nil

}
