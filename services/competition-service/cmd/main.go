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

	"microservice-golang/services/competition-service/internal/config"
	"microservice-golang/services/competition-service/internal/database"
	"microservice-golang/services/competition-service/internal/delivery/nats"
	"microservice-golang/services/competition-service/internal/handler"
	"microservice-golang/services/competition-service/internal/repository"
	replicatedRepo "microservice-golang/services/competition-service/internal/repository/replicated"
	"microservice-golang/services/competition-service/internal/usecase"
	replicatedUC "microservice-golang/services/competition-service/internal/usecase/replicated"
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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "competition-service")
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

	// ── Signal Context ────────────────────────────────────────────────────────
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ── Repositories ──────────────────────────────────────────────────────────
	rosterRepo := repository.NewRosterRepository(db)
	competitionRepo := repository.NewCompetitionRepository(db)
	competitionBranchRepo := repository.NewCompetitionBranchRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	statRepo := repository.NewMatchStatRepository(db)

	// ── Replicated Repositories ───────────────────────────────────────────────
	userRepo := replicatedRepo.NewUserRepository(db)
	sportRepo := replicatedRepo.NewSportRepository(db)
	statusRepo := replicatedRepo.NewStatusRepository(db)
	tagRepo := replicatedRepo.NewTagRepository(db)
	academyHoldingRepo := replicatedRepo.NewAcademyHoldingRepository(db)
	academyBranchRepo := replicatedRepo.NewAcademyBranchRepository(db)
	roleRepo := replicatedRepo.NewRoleRepository(db)
	permissionRepo := replicatedRepo.NewPermissionRepository(db)
	academyAdminRepo := replicatedRepo.NewAcademyAdminRepository(db)

	// ── Replicated Use Cases ──────────────────────────────────────────────────
	userSyncUC := replicatedUC.NewSyncUserUseCase(userRepo, log)
	sportSyncUC := replicatedUC.NewSyncSportUseCase(sportRepo, log)
	statusSyncUC := replicatedUC.NewSyncStatusUseCase(statusRepo, log)
	tagSyncUC := replicatedUC.NewSyncTagUseCase(tagRepo, log)
	holdingSyncUC := replicatedUC.NewSyncAcademyHoldingUseCase(academyHoldingRepo, log)
	branchSyncUC := replicatedUC.NewSyncAcademyBranchUseCase(academyBranchRepo, log)
	roleSyncUC := replicatedUC.NewSyncRoleUseCase(roleRepo, log)
	permissionSyncUC := replicatedUC.NewSyncPermissionUseCase(permissionRepo, log)
	adminSyncUC := replicatedUC.NewSyncAcademyAdminUseCase(academyAdminRepo, log)

	// ── NATS Subscribers ──────────────────────────────────────────────────────
	metaDurableName := envConfig.GetEnv("META_DURABLE_NAME", "competition-service-meta-sync")
	metaSub := nats.NewMetaSubscriber(
		statusSyncUC,
		tagSyncUC,
		natsClient,
		log,
		metaDurableName,
	)

	identityDurableName := envConfig.GetEnv("IDENTITY_DURABLE_NAME", "competition-service-identity-sync")
	identitySub := nats.NewIdentitySubscriber(
		userSyncUC,
		roleSyncUC,
		permissionSyncUC,
		natsClient,
		log,
		identityDurableName,
	)

	sportDurableName := envConfig.GetEnv("SPORT_DURABLE_NAME", "competition-service-sport-sync")
	sportSub := nats.NewSportSubscriber(
		sportSyncUC,
		natsClient,
		log,
		sportDurableName,
	)

	academyDurableName := envConfig.GetEnv("ACADEMY_DURABLE_NAME", "competition-service-academy-sync")
	academySub := nats.NewAcademySubscriber(
		holdingSyncUC,
		branchSyncUC,
		adminSyncUC,
		natsClient,
		log,
		academyDurableName,
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

	go func() {
		log.Info("starting academy event consumer")
		if err := academySub.Listen(ctx); err != nil {
			log.Error("academy consumer stopped", zap.Error(err))
		}
	}()

	// ── Use Cases ─────────────────────────────────────────────────────────────
	rosterUC := usecase.NewRosterUseCase(permissionRepo, rosterRepo)
	competitionUC := usecase.NewCompetitionUseCase(permissionRepo, competitionRepo, academyAdminRepo, academyBranchRepo, roleRepo)
	competitionBranchUC := usecase.NewCompetitionBranchUseCase(permissionRepo, competitionBranchRepo, competitionRepo, academyAdminRepo)
	matchUC := usecase.NewMatchUseCase(permissionRepo, matchRepo, competitionBranchRepo, competitionRepo, academyAdminRepo)
	statUC := usecase.NewStatUseCase(permissionRepo, statRepo, matchRepo, competitionBranchRepo, competitionRepo, sportRepo, academyAdminRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	rosterHandler := handler.NewRosterHandler(rosterUC)
	competitionHandler := handler.NewCompetitionHandler(competitionUC)
	competitionBranchHandler := handler.NewCompetitionBranchHandler(competitionBranchUC)
	matchHandler := handler.NewMatchHandler(matchUC)
	statHandler := handler.NewStatHandler(statUC)

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

	rosterHandler.RegisterGRPC(grpcServer)
	competitionHandler.RegisterGRPC(grpcServer)
	competitionBranchHandler.RegisterGRPC(grpcServer)
	matchHandler.RegisterGRPC(grpcServer)
	statHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	// ── Start gRPC Server ─────────────────────────────────────────────────────
	go func() {
		log.Info("competition-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	// Block until context is done (graceful shutdown signal)
	<-ctx.Done()
	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
