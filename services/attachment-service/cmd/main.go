package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"microservice-golang/services/attachment-service/internal/config"
	"microservice-golang/services/attachment-service/internal/database"
	natsDelivery "microservice-golang/services/attachment-service/internal/delivery/nats"
	"microservice-golang/services/attachment-service/internal/handler"
	"microservice-golang/services/attachment-service/internal/repository"
	"microservice-golang/services/attachment-service/internal/storage"
	"microservice-golang/services/attachment-service/internal/usecase"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/messaging"
	"microservice-golang/shared/pkg/redisclient"
	"microservice-golang/shared/pkg/telemetry"

	"buf.build/go/protovalidate"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// ── Env ───────────────────────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "sport-sentral")
	appVersion := envConfig.GetEnv("APP_VERSION", "0.0.1")
	serviceName := envConfig.GetEnv("SERVICE_NAME", "attachment-service")
	serviceVersion := envConfig.GetEnv("SERVICE_VERSION", "0.0.1")

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(env).With(
		zap.String("appName", appName),
		zap.String("appVersion", appVersion),
		zap.String("serviceName", serviceName),
		zap.String("serviceVersion", serviceVersion),
		zap.String("env", env),
	)
	defer log.Sync()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config failed", zap.Error(err))
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
		log.Info("metrics server listening", zap.String("port", cfg.MetricsPort))
		if err := http.ListenAndServe(cfg.MetricsPort, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server stopped", zap.Error(err))
		}
	}()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.Database.GormDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("connect to database failed", zap.Error(err))
	}

	if err := telemetry.InitGORMTracing(db, serviceName); err != nil {
		log.Fatal("failed to initialize gorm telemetry", zap.Error(err))
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}
	if err := database.CreateIndexes(db); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	redisWrapper := redisclient.New(redisClient)

	// ── Nats ──────────────────────────────────────────────────────────────────
	var natsPublisher *natsDelivery.AttachmentEventPublisher
	natsClient, err := messaging.Connect(cfg.Nats, log)
	if err != nil {
		log.Warn("nats server connection warning (service will proceed without live events)", zap.Error(err))
	} else {
		defer natsClient.Drain()
		natsPublisher = natsDelivery.NewAttachmentEventPublisher(natsClient, log)
	}

	// ── Storage Provider ──────────────────────────────────────────────────────
	var storageProvider storage.StorageProvider
	if cfg.Storage.Provider == "minio" {
		log.Info("initializing MinIO storage provider", zap.String("endpoint", cfg.Storage.Minio.Endpoint))
		sp, err := storage.NewMinioStorage(cfg.Storage.Minio)
		if err != nil {
			log.Warn("failed to initialize MinIO storage provider, falling back to local storage", zap.Error(err))
			localSP, errLocal := storage.NewLocalStorage(cfg.Storage.Local)
			if errLocal != nil {
				log.Fatal("failed to initialize local storage provider", zap.Error(errLocal))
			}
			storageProvider = localSP
		} else {
			storageProvider = sp
		}
	} else {
		log.Info("initializing local storage provider", zap.String("dir", cfg.Storage.Local.LocalDir))
		sp, err := storage.NewLocalStorage(cfg.Storage.Local)
		if err != nil {
			log.Fatal("failed to initialize local storage provider", zap.Error(err))
		}
		storageProvider = sp
	}

	// ── Repositories & UseCases ───────────────────────────────────────────────
	attRepo := repository.NewAttachmentRepository(db, redisWrapper)
	attUC := usecase.NewAttachmentUseCase(attRepo, storageProvider, natsPublisher, cfg.Storage.Minio.Bucket)

	// ── Handlers ──────────────────────────────────────────────────────────────
	attHandler := handler.NewAttachmentHandler(attUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal("failed to initialize validator", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
			interceptor.UnaryValidator(v),
		),
	)
	attHandler.RegisterGRPC(grpcServer)
	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("attachment-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down attachment-service...")
	grpcServer.GracefulStop()
	log.Info("attachment-service exited cleanly")
}
