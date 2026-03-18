package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/internal/app"
	"todo/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("failed load config: %v", err)
	}

	application, err := app.New(cfg)

	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	shutdown := make(chan os.Signal, 1)
	serverShutdown := make(chan error, 1)

	signal.Notify(shutdown, syscall.SIGTERM, os.Interrupt)

	go func() {
		if err = application.Run(); err != nil {
			serverShutdown <- err

		}
	}()

	select {
	case err := <-serverShutdown:
		log.Printf("Не удалось запустить сервер: %v", err)
	case <-shutdown:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err = application.Shutdown(ctx); err != nil {
		log.Printf("Failed to shutdown app: %v", err)
	}

	fmt.Println("Сервер успешно завершил работу")

}
