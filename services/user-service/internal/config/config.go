package config

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
	envConfig "microservice-golang/shared/pkg/config"
)

type Config struct {
	GRPC     GRPCConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Mailer   MailerConfig
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

type MailerConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	grpcPort, err := strconv.Atoi(envConfig.GetEnv("GRPC_PORT", "50051"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRPC_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(envConfig.GetEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	mailerPort, err := strconv.Atoi(envConfig.GetEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	appURL := envConfig.MustGetEnv("APP_URL")

	return &Config{
		GRPC: GRPCConfig{
			Port: grpcPort,
		},
		Database: DatabaseConfig{
			Host:     envConfig.GetEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     envConfig.MustGetEnv("DB_USER"),
			Password: envConfig.MustGetEnv("DB_PASSWORD"),
			Name:     envConfig.MustGetEnv("DB_NAME"),
			SSLMode:  envConfig.GetEnv("DB_SSLMODE", "disable"),
		},
		Mailer: MailerConfig{
			Host:     envConfig.MustGetEnv("SMTP_HOST"),
			Port:     mailerPort,
			Username: envConfig.MustGetEnv("SMTP_USERNAME"),
			Password: envConfig.MustGetEnv("SMTP_PASSWORD"),
			From:     envConfig.MustGetEnv("SMTP_FROM"),
		},
		AppURL: appURL,
	}, nil
}
