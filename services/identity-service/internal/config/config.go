package config

import (
	"fmt"
	"time"

	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/mailer"
)

type Config struct {
	GRPC        GRPCConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Mailer      mailer.Config
	AppURL      string
	MetaService MetaServiceConfig
	MetaNats    NatsConfig
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

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type MetaServiceConfig struct {
	Address string
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

type TelemetryConfig struct {
	Enabled        bool
	JaegerEndpoint string
}

func Load() (*Config, error) {
	redisHost := envConfig.GetEnv("REDIS_HOST", "localhost")
	redisPort := envConfig.GetEnvInt("REDIS_PORT", 6379)

	metaServiceHost := envConfig.GetEnv("META_SERVICE_HOST", "localhost")
	metaServicePort := envConfig.GetEnvInt("META_SERVICE_PORT", 50052)

	identityNatsHost := envConfig.GetEnv("IDENTITY_NATS_HOST", "nats://localhost")
	identityNatsPort := envConfig.GetEnvInt("IDENTITY_NATS_PORT", 4222)

	jaegerHost := envConfig.GetEnv("JAEGER_HOST", "localhost")
	jaegerPort := envConfig.GetEnvInt("JAEGER_PORT", 4317)

	metricPort := envConfig.GetEnv("METRIC_PORT", "9001")

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
			Address:  fmt.Sprintf("%s:%d", redisHost, redisPort),
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
		MetaService: MetaServiceConfig{
			Address: fmt.Sprintf("%s:%d", metaServiceHost, metaServicePort),
		},
		MetaNats: NatsConfig{
			URL:                  fmt.Sprintf("%s:%d", identityNatsHost, identityNatsPort),
			MaxReconnects:        -1,
			ReconnectWait:        2 * time.Second,
			StreamName:           "META_EVENTS",
			StreamSubjects:       []string{"meta.status.*"},
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
