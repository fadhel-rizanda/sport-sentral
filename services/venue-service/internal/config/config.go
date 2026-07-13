package config

import (
	"fmt"
	"time"

	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/events"
	"microservice-golang/shared/pkg/messaging"
)

type Config struct {
	GRPC        GRPCConfig
	Database    DatabaseConfig
	Nats        messaging.Config
	Telemetry   TelemetryConfig
	MetricsPort string
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

type TelemetryConfig struct {
	Enabled        bool
	JaegerEndpoint string
}

func Load() (*Config, error) {
	natsHost := envConfig.GetEnv("IDENTITY_NATS_HOST", "nats://localhost")
	natsPort := envConfig.GetEnvInt("IDENTITY_NATS_PORT", 4222)

	jaegerHost := envConfig.GetEnv("JAEGER_HOST", "localhost")
	jaegerPort := envConfig.GetEnvInt("JAEGER_PORT", 4317)

	metricPort := envConfig.GetEnv("METRIC_PORT", "9008")

	return &Config{
		GRPC: GRPCConfig{
			Port: envConfig.GetEnvInt("GRPC_PORT", 50058),
		},
		Database: DatabaseConfig{
			Host:     envConfig.GetEnv("DB_HOST", "localhost"),
			Port:     envConfig.GetEnvInt("DB_PORT", 5432),
			User:     envConfig.MustGetEnv("DB_USER"),
			Password: envConfig.MustGetEnv("DB_PASSWORD"),
			Name:     envConfig.MustGetEnv("DB_NAME"),
			SSLMode:  envConfig.GetEnv("DB_SSLMODE", "disable"),
		},
		Nats: messaging.Config{
			URL:                  fmt.Sprintf("%s:%d", natsHost, natsPort),
			MaxReconnects:        -1,
			ReconnectWait:        2 * time.Second,
			StreamName:           events.CourtStreamName,
			StreamSubjects:       []string{"court.court.*", "court.booking.*"},
			RetentionMaxAge:      7 * 24 * time.Hour,
			PublishMaxAttempts:   3,
			PublishBaseDelay:     100 * time.Millisecond,
			DefaultAckWait:       30 * time.Second,
			DefaultMaxDeliver:    5,
			DefaultMaxAckPending: 100,
		},
		Telemetry: TelemetryConfig{
			Enabled:        envConfig.GetEnvBool("TELEMETRY_ENABLED", false),
			JaegerEndpoint: fmt.Sprintf("%s:%d", jaegerHost, jaegerPort),
		},
		MetricsPort: fmt.Sprintf(":%s", metricPort),
	}, nil
}
