package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"microservice-golang/services/gateway/internal/client"
	"microservice-golang/services/gateway/internal/config"
	"microservice-golang/services/gateway/internal/handler"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/response"
	"microservice-golang/services/gateway/internal/router"
	"microservice-golang/shared/pkg/jwt"
	"microservice-golang/shared/pkg/telemetry"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

func normalizeFilePayload(data []byte) []byte {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return data
	}

	// Strip data URL prefix if present: data:image/png;base64,...
	if bytes.HasPrefix(trimmed, []byte("data:")) {
		if idx := bytes.Index(trimmed, []byte(",")); idx != -1 {
			trimmed = trimmed[idx+1:]
		}
	}

	// Attempt base64 decode if payload is base64 string
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(trimmed)))
	n, err := base64.StdEncoding.Decode(decoded, trimmed)
	if err == nil && n > 0 {
		decoded = decoded[:n]
		if bytes.HasPrefix(decoded, []byte("\x89PNG\r\n\x1a\n")) ||
			bytes.HasPrefix(decoded, []byte("\xFF\xD8\xFF")) ||
			bytes.HasPrefix(decoded, []byte("GIF8")) ||
			bytes.HasPrefix(decoded, []byte("%PDF")) ||
			(len(decoded) > 12 && bytes.Equal(decoded[0:4], []byte("RIFF")) && bytes.Equal(decoded[8:12], []byte("WEBP"))) {
			return decoded
		}
	}
	return data
}

// @title                      Sport-Sentral API Gateway
// @version                    1.0
// @description                Centralized API Documentation for Sport-Sentral Microservices Gateway (Auth, RBAC, Academy, Venue, Sport, Competition, Scout, Attachment, Log).
// @termsOfService             http://swagger.io/terms/

// @contact.name               Sport-Sentral Development Team
// @contact.email              dev@sportsentral.id

// @license.name               Apache 2.0
// @license.url                http://www.apache.org/licenses/LICENSE-2.0.html

// @host                       localhost:8000
// @BasePath                   /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter JWT token in the format: `Bearer <access_token>`
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

	scoutClient, err := client.NewScoutClient(cfg.GRPC.ScoutAddress)
	if err != nil {
		log.Warn("failed to connect to scout-service (optional gRPC client)", zap.Error(err))
	} else if scoutClient != nil {
		defer scoutClient.Close()
	}

	logClient, err := client.NewLogClient(cfg.GRPC.LogAddress)
	if err != nil {
		log.Warn("failed to connect to log-service (optional gRPC client)", zap.Error(err))
	} else if logClient != nil {
		defer logClient.Close()
	}

	attachmentClient, err := client.NewAttachmentClient(cfg.GRPC.AttachmentAddress)
	if err != nil {
		log.Warn("failed to connect to attachment-service (optional gRPC client)", zap.Error(err))
	} else if attachmentClient != nil {
		defer attachmentClient.Close()
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
	scoutHandler := handler.NewScoutHandler(scoutClient)
	var logHandler *handler.LogHandler
	if logClient != nil {
		logHandler = handler.NewLogHandler(logClient.Log)
	} else {
		logHandler = handler.NewLogHandler(nil)
	}

	var attachmentHandler *handler.AttachmentHandler
	if attachmentClient != nil {
		attachmentHandler = handler.NewAttachmentHandler(attachmentClient.Attachment)
	} else {
		attachmentHandler = handler.NewAttachmentHandler(nil)
	}

	// ── Fiber ─────────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		ReadBufferSize: cfg.App.ReadBufferSize,
		ErrorHandler:   handler.ErrorHandler(log),
	})

	// ── Global Middleware ─────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(otelfiber.Middleware())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(middleware.RateLimit(redisClient, cfg.RateLimit.Max, cfg.RateLimit.Expiration))

	// ── Static Files (Local Attachments) ─────────────────────────────────────
	publicDir := envConfig.GetEnv("PUBLIC_DIR", "../attachment-service/public")
	if _, err := os.Stat(publicDir); err != nil {
		publicDir = "./public"
	}
	_ = os.MkdirAll(publicDir, 0755)

	// Serve static files on GET
	app.Get("/public/*", func(c *fiber.Ctx) error {
		relPath := c.Params("*")
		targetPath := filepath.Join(publicDir, relPath)
		content, err := os.ReadFile(targetPath)
		if err != nil {
			return response.Error(c, fiber.StatusNotFound, "File not found")
		}

		normalized := normalizeFilePayload(content)
		if len(normalized) != len(content) {
			_ = os.WriteFile(targetPath, normalized, 0644)
		}

		contentType := http.DetectContentType(normalized)
		c.Set(fiber.HeaderContentType, contentType)
		return c.Send(normalized)
	})

	// Local development support for presigned PUT upload
	app.Put("/public/*", func(c *fiber.Ctx) error {
		relPath := c.Params("*")
		targetPath := filepath.Join(publicDir, relPath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to create storage directory")
		}
		body := normalizeFilePayload(c.Body())
		if err := os.WriteFile(targetPath, body, 0644); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to save file")
		}
		return c.SendStatus(fiber.StatusOK)
	})

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
		scoutHandler,
		logHandler,
		attachmentHandler,
	)

	// ── Start ─────────────────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("gateway listening", zap.String("port", strconv.Itoa(cfg.App.Port)))

	if err := app.Listen(addr); err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}
}
