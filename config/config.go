package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"os"
	"time"
)

func NewConfig() *Config {
	config, err := env.ParseAs[Config]()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "❌ Failed to parse config: %v\n", err)

		os.Exit(1)
	}
	return &config
}

const (
	EnvironmentProduction = "prod"
	EnvironmentStage      = "stage"
)

type Config struct {
	HTTP  HTTPConfig
	Auth  AuthConfig
	PG    DBConfig
	Redis RedisConfig
	App   AppConfig
}

type HTTPConfig struct {
	Host             string `env:"HTTP_HOST,required"`
	Port             string `env:"HTTP_PORT,required"`
	AllowOrigins     string `env:"ALLOW_ORIGINS,required"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS,required"`
}

type AuthConfig struct {
	SecretSignKey   string        `env:"SECRET_SIGN_KEY,required"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL,required"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL,required"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	DB       int    `env:"REDIS_DB,required"`
	Port     string `env:"REDIS_PORT,required"`
	Password string `env:"REDIS_PASSWORD,required"`
}

type AppConfig struct {
	Environment   string `env:"ENVIRONMENT,required"`
	SolanaNodeURL string `env:"SOLANA_NODE_URL,required"`
}
