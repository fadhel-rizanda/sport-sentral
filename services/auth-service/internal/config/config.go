package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	envConfig "microservice-golang/shared/pkg/config"
)

type Config struct {
	GRPC        GRPCConfig
	UserService UserServiceConfig
	Redis       RedisConfig
	JWT         JWTConfig
}

type GRPCConfig struct {
	Port int
}

type UserServiceConfig struct {
	Address string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	grpcPort, err := strconv.Atoi(envConfig.GetEnv("GRPC_PORT", "50052"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRPC_PORT: %w", err)
	}

	redisDB, err := strconv.Atoi(envConfig.GetEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	accessTTL, err := time.ParseDuration(envConfig.GetEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(envConfig.GetEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	return &Config{
		GRPC: GRPCConfig{
			Port: grpcPort,
		},
		UserService: UserServiceConfig{
			Address: envConfig.GetEnv("USER_SERVICE_ADDRESS", "localhost:50051"),
		},
		Redis: RedisConfig{
			Address:  envConfig.GetEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: envConfig.GetEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			AccessSecret:  envConfig.MustGetEnv("JWT_ACCESS_SECRET"),
			RefreshSecret: envConfig.MustGetEnv("JWT_REFRESH_SECRET"),
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
		},
	}, nil
}
