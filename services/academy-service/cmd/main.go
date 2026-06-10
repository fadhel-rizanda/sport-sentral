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

	"microservice-golang/services/academy-service/internal/config"
	"microservice-golang/services/academy-service/internal/database"
	"microservice-golang/services/academy-service/internal/delivery/nats"
	"microservice-golang/services/academy-service/internal/handler"
	"microservice-golang/services/academy-service/internal/repository"
	repoReplicated "microservice-golang/services/academy-service/internal/repository/replicated"
	"microservice-golang/services/academy-service/internal/usecase"
	ucReplicated "microservice-golang/services/academy-service/internal/usecase/replicated"
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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "academy-service")
	serviceVersion := envConfig.GetEnv("SERVICE_VERSION", "0.0.1")

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(env).With(
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

	// ── Metrics ───────────────────────────────────────────────────────────────
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

	log.Info("Database URL check", zap.String("url", cfg.Database.PgDSN()))
	if err := database.RunExternalMigrations(cfg.Database.PgDSN()); err != nil {
		log.Fatal("database migration failed", zap.Error(err))
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}
	if err := database.CreateIndexes(db); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}

	// ── NATS ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(messaging.Config(cfg.Nats), log)
	if err != nil {
		log.Fatal("failed to connect to nats server", zap.Error(err))
	}
	defer natsClient.Drain()

	// ── Repositories ──────────────────────────────────────────────────────────
	academyHoldingRepo := repository.NewAcademyHoldingRepository(db)
	academyBranchRepo := repository.NewAcademyBranchRepository(db)
	academyAdminRepo := repository.NewAcademyAdminRepository(db)
	enrollmentRepo := repository.NewEnrollmentRepository(db)
	rosterRepo := repository.NewRosterRepository(db)

	// ── Replicated Repositories ───────────────────────────────────────────────
	statusRepo := repoReplicated.NewStatusRepository(db)
	tagRepo := repoReplicated.NewTagRepository(db)
	countryRepo := repoReplicated.NewCountryRepository(db)
	adminDivRepo := repoReplicated.NewAdministrativeDivisionRepository(db)
	userRepo := repoReplicated.NewUserRepository(db)
	roleRepo := repoReplicated.NewRoleRepository(db)
	permissionRepo := repoReplicated.NewPermissionRepository(db)
	sportRepo := repoReplicated.NewSportRepository(db)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	academyHoldingUC := usecase.NewAcademyHoldingUseCase(academyHoldingRepo)
	academyBranchUC := usecase.NewAcademyBranchUseCase(academyBranchRepo)
	academyAdminUC := usecase.NewAcademyAdminUseCase(academyAdminRepo)
	enrollmentUC := usecase.NewEnrollmentUseCase(enrollmentRepo)
	rosterUC := usecase.NewRosterUseCase(rosterRepo)

	// ── Replicated Use Cases ──────────────────────────────────────────────────
	statusSyncUC := ucReplicated.NewSyncStatusUseCase(statusRepo, log)
	tagSyncUC := ucReplicated.NewSyncTagUseCase(tagRepo, log)
	countrySyncUC := ucReplicated.NewSyncCountryUseCase(countryRepo, log)
	adminDivSyncUC := ucReplicated.NewSyncAdministrativeDivisionUseCase(adminDivRepo, log)
	userSyncUC := ucReplicated.NewSyncUserUseCase(userRepo, log)
	roleSyncUC := ucReplicated.NewSyncRoleUseCase(roleRepo, log)
	permissionSyncUC := ucReplicated.NewSyncPermissionUseCase(permissionRepo, log)
	sportSyncUC := ucReplicated.NewSyncSportUseCase(sportRepo, log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	academyHoldingHandler := handler.NewAcademyHoldingHandler(academyHoldingUC)
	academyBranchHandler := handler.NewAcademyBranchHandler(academyBranchUC)
	academyAdminHandler := handler.NewAcademyAdminHandler(academyAdminUC)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentUC)
	rosterHandler := handler.NewRosterHandler(rosterUC)

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

	academyHoldingHandler.RegisterGRPC(grpcServer)
	academyBranchHandler.RegisterGRPC(grpcServer)
	academyAdminHandler.RegisterGRPC(grpcServer)
	enrollmentHandler.RegisterGRPC(grpcServer)
	rosterHandler.RegisterGRPC(grpcServer)

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
		log.Info("academy-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	// ── NATS Subscribers ──────────────────────────────────────────────────────
	metaDurableName := envConfig.GetEnv("META_DURABLE_NAME", "academy-service-meta-sync")
	metaSub := nats.NewMetaSubscriber(
		statusSyncUC,
		tagSyncUC,
		countrySyncUC,
		adminDivSyncUC,
		natsClient,
		log,
		metaDurableName,
	)

	identityDurableName := envConfig.GetEnv("IDENTITY_DURABLE_NAME", "academy-service-identity-sync")
	identitySub := nats.NewIdentitySubscriber(
		userSyncUC,
		roleSyncUC,
		permissionSyncUC,
		natsClient,
		log,
		identityDurableName,
	)

	sportDurableName := envConfig.GetEnv("SPORT_DURABLE_NAME", "academy-service-sport-sync")
	sportSub := nats.NewSportSubscriber(
		sportSyncUC,
		natsClient,
		log,
		sportDurableName,
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

	go func() {
		log.Info("starting sport event consumer")
		if err := sportSub.Listen(ctx); err != nil {
			log.Error("sport consumer stopped", zap.Error(err))
		}
	}()

	// Block until context is done (graceful shutdown signal)
	<-ctx.Done()
	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
