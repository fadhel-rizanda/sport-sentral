package main

import (
	"context"
	"errors"
	"fmt"
	"microservice-golang/services/log-service/internal/config"
	"microservice-golang/services/log-service/internal/database"
	natsDelivery "microservice-golang/services/log-service/internal/delivery/nats"
	"microservice-golang/services/log-service/internal/handler"
	"microservice-golang/services/log-service/internal/repository"
	"microservice-golang/services/log-service/internal/usecase"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/messaging"
	"microservice-golang/shared/pkg/telemetry"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "log-service")
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

	if err := database.RunExternalMigrations(cfg.Database.PgDSN()); err != nil {
		log.Fatal("database migration failed", zap.Error(err))
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}
	if err := database.CreateIndexes(db); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}

	// ── Nats ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(cfg.Nats, log)
	if err != nil {
		log.Warn("nats server connection warning (service will proceed)", zap.Error(err))
	} else {
		defer natsClient.Drain()
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	auditRepo := repository.NewAuditLogRepository(db)
	activityRepo := repository.NewActivityLogRepository(db)

	// ── UseCases ──────────────────────────────────────────────────────────────
	auditUC := usecase.NewAuditLogUsecase(auditRepo)
	activityUC := usecase.NewActivityLogUsecase(activityRepo)
	natsSyncUC := usecase.NewNatsLogSyncUseCase(auditUC, activityUC, log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	logHandler := handler.NewLogHandler(auditUC, activityUC)

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
	logHandler.RegisterGRPC(grpcServer)
	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("log-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	// ── NATS Subscribers ──────────────────────────────────────────────────────
	if natsClient != nil {
		systemDurableName := envConfig.GetEnv("USER_DURABLE_NAME", "log-service-system-events")
		systemSub := natsDelivery.NewSystemEventSubscriber(natsSyncUC, natsClient, log, systemDurableName)

		activityDurableName := envConfig.GetEnv("ACTIVITY_DURABLE_NAME", "log-service-activity-events")
		activitySub := natsDelivery.NewActivityEventSubscriber(natsSyncUC, natsClient, log, activityDurableName)

		go func() {
			log.Info("starting system event subscriber for log service")
			if err := systemSub.Listen(ctx); err != nil {
				log.Error("system subscriber stopped", zap.Error(err))
			}
		}()

		go func() {
			log.Info("starting activity event subscriber for log service")
			if err := activitySub.Listen(ctx); err != nil {
				log.Error("activity subscriber stopped", zap.Error(err))
			}
		}()
	}

	<-ctx.Done()
	log.Info("shutting down log-service gracefully")
	grpcServer.GracefulStop()
}
