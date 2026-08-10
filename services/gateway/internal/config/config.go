package config

import (
	"fmt"
	"time"

	envConfig "microservice-golang/shared/pkg/config"
)

type Config struct {
	App         AppConfig
	GRPC        GRPCClients
	Redis       RedisConfig
	RateLimit   RateLimitConfig
	Telemetry   TelemetryConfig
	JWT         JWTConfig
	MetricsPort string
}

type JWTConfig struct {
	AccessSecret string
}

type AppConfig struct {
	Port int
	Env  string
}

type GRPCClients struct {
	IdentityAddress    string
	MetaAddress        string
	AcademyAddress     string
	CompetitionAddress string
	SportAddress       string
	VenueAddress       string
	ScoutAddress       string
	LogAddress         string
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

type TelemetryConfig struct {
	Enabled        bool
	JaegerEndpoint string
}

func Load() (*Config, error) {
	redisHost := envConfig.GetEnv("REDIS_HOST", "localhost")
	redisPort := envConfig.GetEnvInt("REDIS_PORT", 6379)

	identityHost := envConfig.GetEnv("IDENTITY_SERVICE_HOST", "localhost")
	identityPort := envConfig.GetEnvInt("IDENTITY_SERVICE_PORT", 50051)

	metaHost := envConfig.GetEnv("META_SERVICE_HOST", "localhost")
	metaPort := envConfig.GetEnvInt("META_SERVICE_PORT", 50052)

	academyHost := envConfig.GetEnv("ACADEMY_SERVICE_HOST", "localhost")
	academyPort := envConfig.GetEnvInt("ACADEMY_SERVICE_PORT", 50053)

	competitionHost := envConfig.GetEnv("COMPETITION_SERVICE_HOST", "localhost")
	competitionPort := envConfig.GetEnvInt("COMPETITION_SERVICE_PORT", 50056)

	sportHost := envConfig.GetEnv("SPORT_SERVICE_HOST", "localhost")
	sportPort := envConfig.GetEnvInt("SPORT_SERVICE_PORT", 50055)

	venueHost := envConfig.GetEnv("VENUE_SERVICE_HOST", "localhost")
	venuePort := envConfig.GetEnvInt("VENUE_SERVICE_PORT", 50058)

	scoutHost := envConfig.GetEnv("SCOUT_SERVICE_HOST", "localhost")
	scoutPort := envConfig.GetEnvInt("SCOUT_SERVICE_PORT", 50057)

	logHost := envConfig.GetEnv("LOG_SERVICE_HOST", "localhost")
	logPort := envConfig.GetEnvInt("LOG_SERVICE_PORT", 50059)

	jaegerHost := envConfig.GetEnv("JAEGER_HOST", "localhost")
	jaegerPort := envConfig.GetEnvInt("JAEGER_PORT", 4317)

	metricPort := envConfig.GetEnv("METRIC_PORT", "9000")

	return &Config{
		App: AppConfig{
			Port: envConfig.GetEnvInt("APP_PORT", 8080),
			Env:  envConfig.GetEnv("APP_ENV", "development"),
		},
		GRPC: GRPCClients{
			IdentityAddress:    fmt.Sprintf("%s:%d", identityHost, identityPort),
			MetaAddress:        fmt.Sprintf("%s:%d", metaHost, metaPort),
			AcademyAddress:     fmt.Sprintf("%s:%d", academyHost, academyPort),
			CompetitionAddress: fmt.Sprintf("%s:%d", competitionHost, competitionPort),
			SportAddress:       fmt.Sprintf("%s:%d", sportHost, sportPort),
			VenueAddress:       fmt.Sprintf("%s:%d", venueHost, venuePort),
			ScoutAddress:       fmt.Sprintf("%s:%d", scoutHost, scoutPort),
			LogAddress:         fmt.Sprintf("%s:%d", logHost, logPort),
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
		Telemetry: TelemetryConfig{
			Enabled:        envConfig.GetEnvBool("TELEMETRY_ENABLED", false),
			JaegerEndpoint: fmt.Sprintf("%s:%d", jaegerHost, jaegerPort),
		},
		JWT: JWTConfig{
			AccessSecret: envConfig.GetEnv("JWT_ACCESS_SECRET", ""),
		},
		MetricsPort: fmt.Sprintf(":%s", metricPort),
	}, nil
}
