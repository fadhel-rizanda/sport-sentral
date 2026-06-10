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

	"microservice-golang/services/sport-service/internal/config"
	"microservice-golang/services/sport-service/internal/database"
	"microservice-golang/services/sport-service/internal/delivery/nats"
	"microservice-golang/services/sport-service/internal/handler"
	"microservice-golang/services/sport-service/internal/repository"
	"microservice-golang/services/sport-service/internal/repository/replicated"
	"microservice-golang/services/sport-service/internal/usecase"
	replicatedUC "microservice-golang/services/sport-service/internal/usecase/replicated"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/messaging"
	"microservice-golang/shared/pkg/telemetry"

	"buf.build/go/protovalidate"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "sport-service")
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

	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}
	if err := database.CreateIndexes(db); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}

	// ── NATS ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(cfg.Nats, log)
	if err != nil {
		log.Fatal("failed to connect to nats server", zap.Error(err))
	}
	defer natsClient.Drain()

	// ── Repositories ──────────────────────────────────────────────────────────
	sportRepo := repository.NewSportRepository(db)
	regulatorRepo := repository.NewRegulatorRepository(db)

	// ── Replicated Repositories ───────────────────────────────────────────────
	statusRepo := replicated.NewStatusRepository(db)
	tagRepo := replicated.NewTagRepository(db)
	userRepo := replicated.NewUserRepository(db)

	// ── Replicated Use Cases ──────────────────────────────────────────────────
	statusSyncUC := replicatedUC.NewSyncStatusUseCase(statusRepo, log)
	tagSyncUC := replicatedUC.NewSyncTagUseCase(tagRepo, log)
	userSyncUC := replicatedUC.NewSyncUserUseCase(userRepo, log)

	// ── Event Publisher ───────────────────────────────────────────────────────
	sportPublisher := nats.NewSportEventPublisher(natsClient, log)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	sportUC := usecase.NewSportUseCase(sportRepo, statusRepo, tagRepo, sportPublisher)
	regulatorUC := usecase.NewRegulatorUseCase(regulatorRepo, sportRepo, statusRepo, tagRepo, userRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	sportHandler := handler.NewSportHandler(sportUC)
	regulatorHandler := handler.NewRegulatorHandler(regulatorUC)

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

	sportHandler.RegisterGRPC(grpcServer)
	regulatorHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	// ── Signal Context ────────────────────────────────────────────────────────
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ── Start gRPC Server ─────────────────────────────────────────────────────
	go func() {
		log.Info("sport-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	// ── NATS Subscribers ──────────────────────────────────────────────────────
	metaDurableName := envConfig.GetEnv("META_DURABLE_NAME", "sport-service-meta-sync")
	metaSub := nats.NewMetaSubscriber(
		statusSyncUC,
		tagSyncUC,
		natsClient,
		log,
		metaDurableName,
	)

	identityDurableName := envConfig.GetEnv("IDENTITY_DURABLE_NAME", "sport-service-identity-sync")
	identitySub := nats.NewIdentitySubscriber(
		userSyncUC,
		natsClient,
		log,
		identityDurableName,
	)

	// Start subscribers in goroutines
	go func() {
		log.Info("starting meta event consumer")
		if err := metaSub.Listen(ctx); err != nil {
			log.Error("meta consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		log.Info("starting identity event consumer")
		if err := identitySub.Listen(ctx); err != nil {
			log.Error("identity consumer stopped", zap.Error(err))
		}
	}()

	// Block until context is done (graceful shutdown signal)
	<-ctx.Done()
	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
