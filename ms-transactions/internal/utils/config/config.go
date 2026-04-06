package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type (
	BaseConfig struct {
		Jwt              JwtConfig
		Postgres         PostgresConfig
		Port             int    `env:"PORT" envDefault:"3001"`
		EnableCORS       bool   `env:"ENABLE_CORS" envDefault:"true"`
		CorsAllowOrigins string `env:"CORS_ALLOW_ORIGINS" envDefault:"*"`
	}

	JwtConfig struct {
		Secret string `env:"JWT_SECRET,required" envDefault:"ILIACHALLENGE"`
	}

	PostgresConfig struct {
		Host           string `env:"POSTGRES_HOST,required" envDefault:"localhost"`
		Name           string `env:"POSTGRES_DB"        envDefault:"tunity"`
		Password       string `env:"POSTGRES_PASSWORD"  envDefault:"postgres"`
		User           string `env:"POSTGRES_USER"      envDefault:"postgres"`
		Port           int    `env:"POSTGRES_PORT"      envDefault:"5432"`
		Driver         string `env:"POSTGRES_DRIVER"    envDefault:"postgres"`
		Timeout        int    `env:"POSTGRES_TIMEOUT"         envDefault:"5"`
		IdleConnection int    `env:"POSTGRES_IDLE_CONNECTION" envDefault:"0"`
		LifeTime       int    `env:"POSTGRES_LIFE_TIME"       envDefault:"0"`
		OpenConnection int    `env:"POSTGRES_OPEN_CONNECTION" envDefault:"0"`
		RunMigration   bool   `env:"POSTGRES_MIGRATION"       envDefault:"1"`
	}
)

func (cfg PostgresConfig) GetDataBaseURL() string {
	baseURL := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Driver, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	if cfg.Timeout != 0 {
		baseURL += fmt.Sprintf("&connect_timeout=%d", cfg.Timeout)
	}

	return baseURL
}

func Load() (BaseConfig, error) {
	_ = godotenv.Load()

	cfg := BaseConfig{}
	if err := env.Parse(&cfg); err != nil {
		return BaseConfig{}, err
	}

	return cfg, nil
}
