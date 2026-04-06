package main

import (
	"context"
	"log/slog"
	"os"
	handler "transactions/internal/domain/http"
	"transactions/internal/domain/repository"
	"transactions/internal/domain/service"
	"transactions/internal/utils/config"
	"transactions/internal/utils/jwt"
	"transactions/internal/utils/postgres"

	httpkit "github.com/philippe-berto/httpkit"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	config, err := config.Load()
	if err != nil {
		panic(err)
	}

	db, err := postgres.New(ctx, config.Postgres, "./migrations")
	if err != nil {
		panic(err)
	}

	repo, err := repository.New(db, logger)
	if err != nil {
		panic(err)
	}

	service := service.New(repo, logger)

	jwtClient := jwt.New(config.Jwt)

	server := httpkit.New(config.Port, false, false, config.EnableCORS, config.CorsAllowOrigins)

	handler.Register(server.Router, service, logger, jwtClient)

	if err := server.Start(); err != nil {
		logger.Error("failed to start server", "error", err)
	}
}
