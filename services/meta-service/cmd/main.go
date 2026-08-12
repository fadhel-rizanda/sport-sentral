package main

import (
	"context"
	"errors"
	"fmt"
	"microservice-golang/services/meta-service/internal/config"
	"microservice-golang/services/meta-service/internal/database"
	"microservice-golang/services/meta-service/internal/delivery/nats"
	"microservice-golang/services/meta-service/internal/handler"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/services/meta-service/internal/repository/replicated"
	"microservice-golang/services/meta-service/internal/usecase"
	replicated2 "microservice-golang/services/meta-service/internal/usecase/replicated"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/messaging"
	"microservice-golang/shared/pkg/redisclient"
	"microservice-golang/shared/pkg/telemetry"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "meta-service")
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

	// Add GORM tracing
	if err := telemetry.InitGORMTracing(db, serviceName); err != nil {
		log.Fatal("failed to initialize gorm telemetry", zap.Error(err))
	}

	if err := database.RunExternalMigrations(cfg.Database.PgDSN()); err != nil {
		log.Fatal("database migration failed", zap.Error(err))
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}
	if err := database.CreateIndexes(db); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}
	if err := database.Seed(db); err != nil {
		log.Fatal("failed to seed database", zap.Error(err))
	}

	// ── Nats ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(cfg.Nats, log)
	if err != nil {
		log.Fatal("failed to connect to nats server", zap.Error(err))
	}
	defer natsClient.Drain()

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	redisWrapper := redisclient.New(redisClient)

	// ── Event Publisher ───────────────────────────────────────────────────────
	statusEventPublisher := nats.NewStatusEventPublisher(natsClient, log)
	tagEventPublisher := nats.NewTagEventPublisher(natsClient, log)
	countryEventPublisher := nats.NewCountryEventPublisher(natsClient, log)
	adminDivEventPublisher := nats.NewAdministrativeDivisionPublisher(natsClient, log)

	// ── Repository ────────────────────────────────────────────────────────────
	statusRepo := repository.NewStatusRepository(db, redisWrapper)
	tagRepo := repository.NewTagRepository(db, redisWrapper)
	countryRepo := repository.NewCountryRepository(db, redisWrapper)
	adminDivRepo := repository.NewAdministrativeDivisionRepository(db, redisWrapper)
	userCacheRepo := replicated.NewUserRepository(db)

	// ── UseCase ───────────────────────────────────────────────────────────────
	statusUC := usecase.NewStatusUseCase(statusRepo, statusEventPublisher)
	tagUC := usecase.NewTagUseCase(tagRepo, tagEventPublisher)
	countryUC := usecase.NewCountryUseCase(countryRepo, countryEventPublisher)
	adminDivUC := usecase.NewAdministrativeDivisionUseCase(adminDivRepo, adminDivEventPublisher)
	userSyncUC := replicated2.NewUserSyncUseCase(userCacheRepo, log)

	err = database.SeedStatuses(statusUC)
	if err != nil {
		log.Fatal("failed to seed statuses", zap.Error(err))
	}

	// ── Handlers ──────────────────────────────────────────────────────────────
	statusHandler := handler.NewStatusHandler(statusUC)
	tagHandler := handler.NewTagHandler(tagUC)
	countryHandler := handler.NewCountryHandler(countryUC)
	adminDivHandler := handler.NewAdministrativeDivisionHandler(adminDivUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal("failed to initialize validator", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // Add tracing to server
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
			interceptor.UnaryValidator(v),
		),
	)
	statusHandler.RegisterGRPC(grpcServer)
	tagHandler.RegisterGRPC(grpcServer)
	countryHandler.RegisterGRPC(grpcServer)
	adminDivHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("meta-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	log.Info("grpc server listening", zap.Int("port", cfg.GRPC.Port))

	// ── NATS Subscriber ────────────────────────────────────────────────────────────────
	userDurableName := envConfig.GetEnv("USER_DURABLE_NAME", "meta-service-user-sync")
	identitySub := nats.NewIdentitySubscriber(userSyncUC, natsClient, log, userDurableName)

	log.Info("starting user event consumer")
	if err := identitySub.Listen(ctx); err != nil {
		log.Error("consumer stopped", zap.Error(err))
	}

	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
