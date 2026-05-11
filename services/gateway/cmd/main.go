package main

import (
	"fmt"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/config"
	"microservice-golang/services/gateway/internal/handler"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/router"
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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "gateway")
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

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(identityClient.Auth, identityClient.User)
	userHandler := handler.NewUserHandler(identityClient.User)
	profileHandler := handler.NewProfileHandler(identityClient.RBAC)
	statusHandler := handler.NewStatusHandler(metaClient.Status)
	tagHandler := handler.NewTagHandler(metaClient.Tag)

	// ── Fiber ─────────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler(log),
	})

	// ── Global Middleware ─────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(middleware.RateLimit(redisClient, cfg.RateLimit.Max, cfg.RateLimit.Expiration))

	// ── Routes ────────────────────────────────────────────────────────────────
	router.Setup(
		app,
		authHandler,
		userHandler,
		profileHandler,
		identityClient.Auth,
		statusHandler,
		tagHandler,
	)

	// ── Start ─────────────────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("gateway listening", zap.String("port", strconv.Itoa(cfg.App.Port)))

	if err := app.Listen(addr); err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}
}
