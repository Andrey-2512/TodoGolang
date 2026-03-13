package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"todo/internal/auth"
	"todo/internal/config"
	router "todo/internal/delivery/http"
	"todo/internal/delivery/http/middlewares"
	delivery "todo/internal/delivery/http/v1"
	"todo/internal/repositories"
	"todo/internal/services"
	"todo/pkg/database/postgres"
	"todo/pkg/database/redis"

	redisLib "github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	server *http.Server
	db     *pgxpool.Pool
	redis  *redisLib.Client
}

func New(cfg *config.Config) (*App, error) {

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

	redisClient := redis.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.MinIdleConns, cfg.Redis.PoolSize, cfg.Redis.ConnMaxLifetime)

	jwt := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	hasher := auth.NewHasher(cfg.Hash.Time, cfg.Hash.Memory, cfg.Hash.KeyLen, cfg.Hash.Threads, cfg.Hash.SaltLength)

	taskRepo := repositories.NewTaskRepository(client)
	cacheTaskRepo := repositories.NewCacheTaskRepository(taskRepo, redisClient, cfg.Cache.CacheTaskTTL, cfg.Cache.TasksPrefix, cfg.Cache.UserTasksPrefix)
	usersRepo := repositories.NewUsersRepository(client)
	blacklistRepo := repositories.NewBlacklistRepository(redisClient, cfg.JWT.BlacklistPrefix)

	taskService := services.NewTaskService(cacheTaskRepo)
	usersService := services.NewUsersService(usersRepo, hasher, jwt, blacklistRepo)

	taskHandler := delivery.NewTaskHandler(taskService)
	usersHandler := delivery.NewUsersHandler(usersService)

	authMiddleware := middlewares.NewAuthMiddleware(jwt)
	corsMiddleware := middlewares.NewCORSMiddleware(cfg.HTTP.CORSUrl)

	mux := router.NewRouter(taskHandler, usersHandler, authMiddleware, corsMiddleware)

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
