package main

import (
	"context"
	"errors"
	"fmt"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/identity-service/internal/config"
	"microservice-golang/services/identity-service/internal/database"
	"microservice-golang/services/identity-service/internal/delivery/nats"
	"microservice-golang/services/identity-service/internal/handler"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/services/identity-service/internal/repository/replicated"
	"microservice-golang/services/identity-service/internal/usecase"
	replicated2 "microservice-golang/services/identity-service/internal/usecase/replicated"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/jwt"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/mailer"
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
	"google.golang.org/grpc/credentials/insecure"
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
	serviceName := envConfig.GetEnv("SERVICE_NAME", "identity-service")
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
	if err := database.Seed(db); err != nil {
		log.Fatal("failed to seed database", zap.Error(err))
	}

	// ── NATS ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(messaging.Config(cfg.MetaNats), log)
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

	// ── JWT ───────────────────────────────────────────────────────────────────
	jwtManager := jwt.NewManager(jwt.Config{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessTTL:     cfg.JWT.AccessTTL,
		RefreshTTL:    cfg.JWT.RefreshTTL,
	})

	// ── Mailer ────────────────────────────────────────────────────────────────
	mailerClient := mailer.New(cfg.Mailer)

	// ── gRPC Clients ──────────────────────────────────────────────────────────
	metaConn, err := grpc.NewClient(cfg.MetaService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Add tracing to client
	)
	if err != nil {
		log.Fatal("failed to connect to meta-service", zap.Error(err))
	}
	defer metaConn.Close()
	_ = metav1.NewStatusServiceClient(metaConn) // not yet implement

	// ── Event Publisher ───────────────────────────────────────────────────────
	userEventPublisher := nats.NewUserEventPublisher(natsClient, log)
	roleEventPublisher := nats.NewRoleEventPublisher(natsClient, log)
	permissionEventPublisher := nats.NewPermissionEventPublisher(natsClient, log)

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	statusCacheRepo := replicated.NewStatusRepository(db)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUseCase(
		userRepo,
		userRoleRepo,
		jwtManager,
		redisClient,
		cfg.JWT.RefreshTTL,
		statusCacheRepo,
	)
	userUC := usecase.NewUserUseCase(
		db,
		userRepo,
		userRoleRepo,
		roleRepo,
		mailerClient,
		redisWrapper,
		cfg.AppURL,
		log,
		statusCacheRepo,
		userEventPublisher,
	)
	profileUC := usecase.NewProfileUseCase(
		userRepo,
		userRoleRepo,
		roleRepo,
		statusCacheRepo,
	)
	roleUC := usecase.NewRoleUseCase(roleRepo, permissionRepo, userRepo, log, roleEventPublisher)
	permissionUC := usecase.NewPermissionUseCase(permissionRepo, permissionEventPublisher)
	statusSyncUC := replicated2.NewSyncStatusUseCase(statusCacheRepo, log)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)
	profileHandler := handler.NewProfileHandler(profileUC)
	userInternalHandler := handler.NewUserInternalHandler(userUC)
	roleHandler := handler.NewRoleHandler(roleUC)
	permissionHandler := handler.NewPermissionHandler(permissionUC)

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
	authHandler.RegisterGRPC(grpcServer)
	userHandler.RegisterGRPC(grpcServer)
	profileHandler.RegisterGRPC(grpcServer)
	userInternalHandler.RegisterGRPC(grpcServer)
	roleHandler.RegisterGRPC(grpcServer)
	permissionHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info("identity-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	log.Info("grpc server listening", zap.Int("port", cfg.GRPC.Port))

	// ── NATS Subscriber ────────────────────────────────────────────────────────────────
	statusDurableName := envConfig.GetEnv("STATUS_DURABLE_NAME", "identity-service-status-sync")
	metaSub := nats.NewMetaSubscriber(statusSyncUC, natsClient, log, statusDurableName)

	log.Info("starting status event consumer")
	if err := metaSub.Listen(ctx); err != nil {
		log.Error("consumer stopped", zap.Error(err))
	}

	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
