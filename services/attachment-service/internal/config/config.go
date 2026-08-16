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
	Redis       RedisConfig
	Nats        messaging.Config
	Storage     StorageConfig
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

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

type StorageConfig struct {
	Provider string // "local" or "minio"
	Local    LocalStorageConfig
	Minio    MinioConfig
}

type LocalStorageConfig struct {
	BaseURL  string // Prefix URL unduhan lokal (e.g. "http://localhost:8080/public")
	LocalDir string // Direktori penyimpanan di disk lokal (e.g. "./public")
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

type TelemetryConfig struct {
	Enabled        bool
	JaegerEndpoint string
}

func Load() (*Config, error) {
	redisHost := envConfig.GetEnv("REDIS_HOST", "localhost")
	redisPort := envConfig.GetEnvInt("REDIS_PORT", 6379)

	natsHost := envConfig.GetEnv("IDENTITY_NATS_HOST", "nats://localhost")
	natsPort := envConfig.GetEnvInt("IDENTITY_NATS_PORT", 4222)

	jaegerHost := envConfig.GetEnv("JAEGER_HOST", "localhost")
	jaegerPort := envConfig.GetEnvInt("JAEGER_PORT", 4317)

	metricPort := envConfig.GetEnv("METRIC_PORT", "9010")

	return &Config{
		GRPC: GRPCConfig{
			Port: envConfig.GetEnvInt("GRPC_PORT", 50060),
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
			DB:       envConfig.GetEnvInt("REDIS_DB", 8),
		},
		Nats: messaging.Config{
			URL:                  fmt.Sprintf("%s:%d", natsHost, natsPort),
			MaxReconnects:        -1,
			ReconnectWait:        2 * time.Second,
			StreamName:           events.AttachmentStreamName,
			StreamSubjects:       []string{"attachment.>"},
			RetentionMaxAge:      30 * 24 * time.Hour,
			PublishMaxAttempts:   3,
			PublishBaseDelay:     100 * time.Millisecond,
			DefaultAckWait:       30 * time.Second,
			DefaultMaxDeliver:    5,
			DefaultMaxAckPending: 100,
		},
		Storage: StorageConfig{
			Provider: envConfig.GetEnv("STORAGE_PROVIDER", "local"),
			Local: LocalStorageConfig{
				BaseURL:  envConfig.GetEnv("STORAGE_BASE_URL", "http://localhost:8080/public"),
				LocalDir: envConfig.GetEnv("STORAGE_LOCAL_DIR", "./public"),
			},
			Minio: MinioConfig{
				Endpoint:  envConfig.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
				AccessKey: envConfig.GetEnv("MINIO_ACCESS_KEY", "minioadmin"),
				SecretKey: envConfig.GetEnv("MINIO_SECRET_KEY", "minioadmin"),
				UseSSL:    envConfig.GetEnvBool("MINIO_USE_SSL", false),
				Bucket:    envConfig.GetEnv("MINIO_BUCKET", "attachments"),
			},
		},
		Telemetry: TelemetryConfig{
			Enabled:        envConfig.GetEnvBool("TELEMETRY_ENABLED", false),
			JaegerEndpoint: fmt.Sprintf("%s:%d", jaegerHost, jaegerPort),
		},
		MetricsPort: fmt.Sprintf(":%s", metricPort),
	}, nil
}
