package config

import (
	"fmt"
	"time"

	envConfig "microservice-golang/shared/pkg/config"
)

type Config struct {
	App       AppConfig
	GRPC      GRPCClients
	Redis     RedisConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Port int
	Env  string
}

type GRPCClients struct {
	IdentityAddress string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

func Load() (*Config, error) {
	redisHost := envConfig.GetEnv("REDIS_HOST", "localhost")
	redisPort := envConfig.GetEnvInt("REDIS_PORT", 6379)
	return &Config{
		App: AppConfig{
			Port: envConfig.GetEnvInt("APP_PORT", 8080),
			Env:  envConfig.GetEnv("APP_ENV", "development"),
		},
		GRPC: GRPCClients{
			IdentityAddress: envConfig.GetEnv("IDENTITY_SERVICE_ADDRESS", "localhost:50051"),
		},
		Redis: RedisConfig{
			Address:  fmt.Sprintf("%s:%d", redisHost, redisPort),
			Password: envConfig.GetEnv("REDIS_PASSWORD", ""),
			DB:       envConfig.GetEnvInt("REDIS_DB", 0),
		},
		RateLimit: RateLimitConfig{
			Max:        envConfig.GetEnvInt("RATE_LIMIT_MAX", 100),
			Expiration: envConfig.GetEnvDuration("RATE_LIMIT_EXPIRATION", 1*time.Minute),
		},
	}, nil
}
