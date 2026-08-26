package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	authhandlers "todo/internal/auth/handlers"
	authrepo "todo/internal/auth/repositories"
	authsvc "todo/internal/auth/services"
	"todo/internal/config"
	"todo/internal/database/postgres"
	"todo/internal/database/redis"
	"todo/internal/middlewares"
	"todo/internal/security"
	todohandlers "todo/internal/todo/handlers"
	todorepo "todo/internal/todo/repositories"
	todosvc "todo/internal/todo/services"
	usershandlers "todo/internal/users/handlers"
	usersrepo "todo/internal/users/repositories"
	userssvc "todo/internal/users/services"

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
	const op = "app.setupRedis"

	redisClient, err := redis.NewClient(cfg.Redis.ConnTimeout, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.Username, cfg.Redis.DB, cfg.Redis.MinIdleConns, cfg.Redis.PoolSize, cfg.Redis.ConnMaxLifetime)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return redisClient, nil
}

func setupDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	const op = "app.setupDatabase"
	client, err := postgres.NewClient(cfg.Database.ConnTimeout, fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name), cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.MaxConnLifetime)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return client, nil
}

func setupRouter(cfg *config.Config, DbClient *pgxpool.Pool, redisClient *redisLib.Client) *chi.Mux {
	jwt := security.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	hasher := security.NewHasher(cfg.Hash.Time, cfg.Hash.Memory, cfg.Hash.KeyLen, cfg.Hash.SaltLength, cfg.Hash.Threads)

	taskRepo := todorepo.NewTaskRepository(DbClient, cfg.App.MaxTasksPerUser)
	cacheTaskRepo := todorepo.NewCacheTaskRepository(taskRepo, redisClient, cfg.Cache.TasksPrefix, cfg.Cache.UserTasksPrefix, cfg.Cache.CacheTaskTTL)
	usersRepo := usersrepo.NewUsersRepository(DbClient)
	whitelistRepo := authrepo.NewWhitelistRepository(redisClient, cfg.JWT.WhitelistPrefix)

	taskService := todosvc.NewTaskService(cacheTaskRepo)
	usersService := authsvc.NewAuthService(usersRepo, hasher, jwt, whitelistRepo)
	profileService := userssvc.NewProfileService(cacheTaskRepo, cfg.App.MaxTasksPerUser)

	taskHandler := todohandlers.NewTaskHandler(taskService, cfg.HTTP.HandlerTimeout)
	usersHandler := authhandlers.NewAuthHandler(usersService, cfg.HTTP.HandlerTimeout, cfg.HTTP.CookieSecure, cfg.JWT.RefreshTTL)
	profileHandler := usershandlers.NewProfileHandler(profileService, cfg.HTTP.HandlerTimeout)

	authMiddleware := middlewares.NewAuthMiddleware(jwt)
	corsMiddleware := middlewares.NewCORSMiddleware(cfg.HTTP.CORSUrl, cfg.HTTP.AllowHeaders, cfg.HTTP.AllowMethods, cfg.HTTP.AllowCredentials, cfg.HTTP.AccessControlMaxAge)

	mux := NewRouter(taskHandler, usersHandler, authMiddleware, corsMiddleware, profileHandler)

	return mux
}

func New(cfg *config.Config) (*App, error) {
	const op = "app.New"
	client, err := setupDatabase(cfg)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	redisClient, err := setupRedis(cfg)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
	const op = "app.App.Run"
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	const op = "app.App.Shutdown"
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	a.db.Close()

	if err := a.redis.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}
