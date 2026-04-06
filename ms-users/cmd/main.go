package main

import (
	"context"
	"log/slog"
	"os"
	handler "users/internal/domain/http"
	"users/internal/domain/repository"
	"users/internal/domain/service"
	"users/internal/utils/config"
	"users/internal/utils/jwt"
	"users/internal/utils/postgres"

	httpkit "github.com/philippe-berto/httpkit"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	db, err := postgres.New(ctx, cfg.Postgres, "./migrations")
	if err != nil {
		panic(err)
	}

	repo, err := repository.New(db, logger)
	if err != nil {
		panic(err)
	}

	jwtClient := jwt.New(cfg.Jwt)

	svc := service.New(repo, jwtClient, logger)

	server := httpkit.New(cfg.Port, false, false, cfg.EnableCORS, cfg.CorsAllowOrigins)

	handler.Register(server.Router, svc, logger, jwtClient)

	if err := server.Start(); err != nil {
		logger.Error("failed to start server", "error", err)
	}
}
