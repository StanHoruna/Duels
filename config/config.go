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
	Environment        string `env:"ENVIRONMENT,required"`
	SolanaNodeURL      string `env:"SOLANA_URL,required"`
	SolanaWSNodeURL    string `env:"SOLANA_WS_URL,required"`
	SolanaQuickNodeAPI string `env:"SOLANA_QUICKNODE_API,required"`

	ShareImageAPI    string `env:"SHARE_IMAGE_API,required"`
	USDCMintAddress  string `env:"USDC_MINT_ADDRESS,required"`
	USDCMintDecimals uint8  `env:"USDC_MINT_DECIMALS,required"`

	SolanaAdminPrivateKey string `env:"SOLANA_ADMIN_PRIVATE_KEY,required"`
	ContractAddress       string `env:"CONTRACT_ADDRESS,required"`
	ContractAddressApi    string `env:"CONTRACT_ADDRESS_API,required"`
}
