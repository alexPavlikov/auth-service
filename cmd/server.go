package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexPavlikov/auth-service/internal/config"
	"github.com/alexPavlikov/auth-service/internal/postgres"
	"github.com/alexPavlikov/auth-service/internal/repository"
	"github.com/alexPavlikov/auth-service/internal/server"
	"github.com/alexPavlikov/auth-service/internal/server/locations"
	"github.com/alexPavlikov/auth-service/internal/service"
)

func Run() error {

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("cmd run error: %w", err)
	}

	t := time.Duration(cfg.Postgres.ConnectTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	db, err := postgres.Connect(ctx, *cfg)
	if err != nil {
		return fmt.Errorf("failed connect to postgres: %w", err)
	}

	repo := repository.New(db)
	service := service.New(repo)
	handler := locations.New(service)
	router := server.New(handler)

	slog.Info(fmt.Sprintf("server listen on %s:%d", cfg.Server.Path, cfg.Server.Port))

	srv := &http.Server{
		Addr:              cfg.Server.ToString(),
		Handler:           router.Build(),
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("start http serve error: %w", err)
	}

	return nil
}
