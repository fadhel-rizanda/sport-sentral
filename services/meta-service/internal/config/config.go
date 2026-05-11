package config

import (
	"fmt"
	envConfig "microservice-golang/shared/pkg/config"
	"time"
)

type Config struct {
	GRPC         GRPCConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	IdentityNats NatsConfig
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

type NatsConfig struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration

	StreamName      string
	StreamSubjects  []string
	RetentionMaxAge time.Duration

	PublishMaxAttempts int
	PublishBaseDelay   time.Duration

	DefaultAckWait       time.Duration
	DefaultMaxDeliver    int
	DefaultMaxAckPending int
}

func (d *DatabaseConfig) GormDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (d *DatabaseConfig) PgDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func Load() (*Config, error) {
	redisHost := envConfig.GetEnv("REDIS_HOST", "localhost")
	redisPort := envConfig.GetEnvInt("REDIS_PORT", 6379)
	identityNatsHost := envConfig.GetEnv("IDENTITY_NATS_HOST", "nats://localhost")
	identityNatsPort := envConfig.GetEnvInt("IDENTITY_NATS_PORT", 4222)
	return &Config{
		GRPC: GRPCConfig{
			Port: envConfig.GetEnvInt("GRPC_PORT", 50052),
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
			Address:  fmt.Sprintf("%s:%d", redisHost, redisPort),
			Password: envConfig.GetEnv("REDIS_PASSWORD", ""),
			DB:       envConfig.GetEnvInt("REDIS_DB", 1),
		},
		IdentityNats: NatsConfig{
			URL:                  fmt.Sprintf("%s:%d", identityNatsHost, identityNatsPort),
			MaxReconnects:        -1,
			ReconnectWait:        2 * time.Second,
			StreamName:           "IDENTITY_EVENTS",
			StreamSubjects:       []string{"identity.user.*"},
			RetentionMaxAge:      7 * 24 * time.Hour,
			PublishMaxAttempts:   3,
			PublishBaseDelay:     100 * time.Millisecond,
			DefaultAckWait:       30 * time.Second,
			DefaultMaxDeliver:    5,
			DefaultMaxAckPending: 100,
		},
	}, nil
}
