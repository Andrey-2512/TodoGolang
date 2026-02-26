package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	migrations "todo"
	"todo/internal/auth"
	"todo/internal/config"
	router "todo/internal/delivery/http"
	"todo/internal/delivery/http/middlewares"
	delivery "todo/internal/delivery/http/v1"
	"todo/internal/repositories"
	"todo/internal/services"
	"todo/pkg/database/postgres"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("failed load config: %v", err)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)

	client, err := postgres.NewClient(dbURL, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	if err != nil {
		log.Fatalf("Failed to up database: %v", err)
	}
	defer client.Close()
	err = migrations.SetMigrations("pgx5" + dbURL[len("postgres"):])
	if err != nil {
		log.Printf("Failed to setup migrations %v", err)
	}

	jwt := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	hasher := auth.NewHasher(cfg.Hash.Time, cfg.Hash.Memory, cfg.Hash.KeyLen, cfg.Hash.Threads, cfg.Hash.SaltLength)

	taskRepo := repositories.NewTaskRepository(client)
	usersRepo := repositories.NewUsersRepository(client)

	taskService := services.NewTaskService(taskRepo)
	usersService := services.NewUsersService(usersRepo, hasher, jwt)

	taskHandler := delivery.NewTaskHandler(taskService)
	usersHandler := delivery.NewUsersHandler(usersService)

	authMiddleware := middlewares.NewAuthMiddleware(jwt)
	corsMiddleware := middlewares.NewCORSMiddleware(cfg.HTTP.CORSUrl)

	mainRouter := router.NewRouter(taskHandler, usersHandler, authMiddleware, corsMiddleware)

	shutdown := make(chan os.Signal, 1)
	ErrShutdown := make(chan error, 1)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      mainRouter,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	go func() {
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			ErrShutdown <- err
		}

	}()

	signal.Notify(shutdown, syscall.SIGTERM, os.Interrupt)
	select {
	case err := <-ErrShutdown:
		log.Printf("Не удалось запустить сервер: %v", err)
	case <-shutdown:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)

	if err != nil {
		log.Println("Сервер завершил работу по таймауту")
	}
	fmt.Println("Сервер успешно завершил работу")

}
