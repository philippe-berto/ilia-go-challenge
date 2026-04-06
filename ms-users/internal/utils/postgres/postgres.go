package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
	"users/internal/utils/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type Client struct {
	client *sqlx.DB
}

func New(ctx context.Context, cfg config.PostgresConfig, migrationLocation string) (*Client, error) {
	databaseURL := cfg.GetDataBaseURL()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	driverName := "postgres"

	db, err := sqlx.Open(driverName, databaseURL)
	if err != nil {
		logger.Error("Database: fail to open connection", "error", err)
		return nil, err
	}

	if cfg.OpenConnection != 0 {
		db.SetMaxOpenConns(cfg.OpenConnection)
	}

	if cfg.IdleConnection != 0 {
		db.SetMaxIdleConns(cfg.IdleConnection)
	}

	if cfg.LifeTime != 0 {
		db.SetConnMaxLifetime(time.Duration(1) * time.Millisecond)
	}

	if err = db.PingContext(ctx); err != nil {
		logger.Error("Database: Could not ping the database", "error", err)
		return nil, err
	}

	migrations, err := runMigration(cfg, migrationLocation, logger)
	if err != nil {
		return nil, err
	}

	if err = closeMigration(migrations, logger); err != nil {
		return nil, err
	}

	return &Client{client: db}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("database: failed to close connection: %v", err)
	}
	return nil
}

func (c *Client) GetClient() *sqlx.DB {
	return c.client
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.PingContext(ctx)
}

func (c *Client) PrepareStatement(query string) (*sqlx.Stmt, error) {
	return c.client.Preparex(query)
}

func runMigration(cfg config.PostgresConfig, migrationLocation string, logger *slog.Logger) (*migrate.Migrate, error) {
	if !cfg.RunMigration || migrationLocation == "" {
		return nil, nil
	}

	logger.Info("Running migration")
	m, err := migrate.New("file://"+migrationLocation, cfg.GetDataBaseURL())
	if err != nil {
		logger.Error("Database: Error setting up migration connection", "error", err)
		return nil, err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
		logger.Error("Database: Error running migration", "error", err)
		return nil, err
	}

	return m, nil
}

func closeMigration(migrations *migrate.Migrate, logger *slog.Logger) error {
	if migrations == nil {
		return nil
	}

	sourceErr, dbErr := migrations.Close()
	if sourceErr != nil {
		logger.Error("Database: Error close migration source", "error", sourceErr)
		return fmt.Errorf("database: failed to close migration source")
	}

	if dbErr != nil {
		logger.Error("Database: Error close migration database", "error", dbErr)
		return fmt.Errorf("database: failed to close migration db")
	}

	return nil
}
