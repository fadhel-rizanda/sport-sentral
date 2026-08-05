package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/config"
	"microservice-golang/services/gateway/internal/handler"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/router"
	"microservice-golang/shared/pkg/jwt"
	"microservice-golang/shared/pkg/telemetry"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	envConfig "microservice-golang/shared/pkg/config"
	zapLogger "microservice-golang/shared/pkg/logger"
)

func main() {
	// ── Env ───────────────────────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "sport-sentral")
	appVersion := envConfig.GetEnv("APP_VERSION", "0.0.1")
	serviceName := envConfig.GetEnv("SERVICE_NAME", "api-gateway")
	serviceVersion := envConfig.GetEnv("SERVICE_VERSION", "0.0.1")

	// ── Logger ────────────────────────────────────────────────────────────────
	log := zapLogger.New(env).With(
		zap.String("app_name", appName),
		zap.String("app_version", appVersion),
		zap.String("service", serviceName),
		zap.String("service_version", serviceVersion),
		zap.String("env", env),
	)
	defer log.Sync()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	// ── Telemetry ─────────────────────────────────────────────────────────────
	tel, err := telemetry.New(telemetry.Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    env,
		JaegerEndpoint: cfg.Telemetry.JaegerEndpoint,
		Enabled:        cfg.Telemetry.Enabled,
	}, log)
	if err != nil {
		log.Fatal("failed to init telemetry service", zap.Error(err))
	}
	defer tel.Shutdown(context.Background())

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		metricsAddr := cfg.MetricsPort
		log.Info("metrics server listening", zap.String("addr", metricsAddr))

		if err := http.ListenAndServe(metricsAddr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server stopped", zap.Error(err))
		}
	}()

	// ── gRPC Client ──────────────────────────────────────────────────
	identityClient, err := client.NewIdentityClient(cfg.GRPC.IdentityAddress)
	if err != nil {
		log.Fatal("failed to connect to identity-service", zap.Error(err))
	}
	defer identityClient.Close()

	metaClient, err := client.NewMetaClient(cfg.GRPC.MetaAddress)
	if err != nil {
		log.Fatal("failed to connect to meta-service", zap.Error(err))
	}
	defer metaClient.Close()

	academyClient, err := client.NewAcademyClient(cfg.GRPC.AcademyAddress)
	if err != nil {
		log.Fatal("failed to connect to academy-service", zap.Error(err))
	}
	defer academyClient.Close()

	competitionClient, err := client.NewCompetitionClient(cfg.GRPC.CompetitionAddress)
	if err != nil {
		log.Fatal("failed to connect to competition-service", zap.Error(err))
	}
	defer competitionClient.Close()

	sportClient, err := client.NewSportClient(cfg.GRPC.SportAddress)
	if err != nil {
		log.Fatal("failed to connect to sport-service", zap.Error(err))
	}
	defer sportClient.Close()

	venueClient, err := client.NewVenueClient(cfg.GRPC.VenueAddress)
	if err != nil {
		log.Fatal("failed to connect to venue-service", zap.Error(err))
	}
	defer venueClient.Close()

	logClient, err := client.NewLogClient(cfg.GRPC.LogAddress)
	if err != nil {
		log.Warn("failed to connect to log-service (optional gRPC client)", zap.Error(err))
	} else if logClient != nil {
		defer logClient.Close()
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// ── JWT Manager ───────────────────────────────────────────────────────────
	var jwtManager *jwt.Manager
	if cfg.JWT.AccessSecret != "" {
		jwtManager = jwt.NewManager(jwt.Config{
			AccessSecret: cfg.JWT.AccessSecret,
		})
	}

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(identityClient.Auth, identityClient.User)
	userHandler := handler.NewUserHandler(identityClient.User)
	profileHandler := handler.NewProfileHandler(identityClient.RBAC)
	statusHandler := handler.NewStatusHandler(metaClient.Status)
	tagHandler := handler.NewTagHandler(metaClient.Tag)
	roleHandler := handler.NewRoleHandler(identityClient.Role)
	permissionHandler := handler.NewPermissionHandler(identityClient.Permission)
	academyHandler := handler.NewAcademyHandler(academyClient, competitionClient)
	sportHandler := handler.NewSportHandler(sportClient)
	competitionHandler := handler.NewCompetitionHandler(competitionClient)
	venueHandler := handler.NewVenueHandler(venueClient)
	var logHandler *handler.LogHandler
	if logClient != nil {
		logHandler = handler.NewLogHandler(logClient.Log)
	}

	// ── Fiber ─────────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler(log),
	})

	// ── Global Middleware ─────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(otelfiber.Middleware())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(middleware.RateLimit(redisClient, cfg.RateLimit.Max, cfg.RateLimit.Expiration))

	// ── Routes ────────────────────────────────────────────────────────────────
	router.Setup(
		app,
		authHandler,
		userHandler,
		profileHandler,
		jwtManager,
		identityClient.Auth,
		statusHandler,
		tagHandler,
		roleHandler,
		permissionHandler,
		academyHandler,
		sportHandler,
		competitionHandler,
		venueHandler,
		logHandler,
	)

	// ── Start ─────────────────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("gateway listening", zap.String("port", strconv.Itoa(cfg.App.Port)))

	if err := app.Listen(addr); err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}
}
