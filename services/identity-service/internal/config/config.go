package config

import (
	"fmt"
	"time"

	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/mailer"
)

type Config struct {
	GRPC     GRPCConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Mailer   mailer.Config
	AppURL   string
}

type GRPCConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
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
	return &Config{
		GRPC: GRPCConfig{
			Port: envConfig.GetEnvInt("GRPC_PORT", 50051),
		},
		Database: DatabaseConfig{
			Host:     envConfig.GetEnv("DB_HOST", "localhost"),
			Port:     envConfig.GetEnvInt("DB_PORT", 5432),
			User:     envConfig.MustGetEnv("DB_USER"),
			Password: envConfig.MustGetEnv("DB_PASSWORD"),
			Name:     envConfig.MustGetEnv("DB_NAME"),
			SSLMode:  envConfig.GetEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Address:  envConfig.GetEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: envConfig.GetEnv("REDIS_PASSWORD", ""),
			DB:       envConfig.GetEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  envConfig.GetEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: envConfig.GetEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:     envConfig.GetEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    envConfig.GetEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		Mailer: mailer.Config{
			Host:     envConfig.GetEnv("MAILER_HOST", ""),
			Port:     envConfig.GetEnvInt("MAILER_PORT", 587),
			Username: envConfig.GetEnv("MAILER_USERNAME", ""),
			Password: envConfig.GetEnv("MAILER_PASSWORD", ""),
			From:     envConfig.GetEnv("MAILER_FROM", ""),
		},
		AppURL: envConfig.MustGetEnv("APP_URL"),
	}, nil
}
