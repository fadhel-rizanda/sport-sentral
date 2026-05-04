package main

import (
	"buf.build/go/protovalidate"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"microservice-golang/services/meta-service/internal/config"
	"microservice-golang/services/meta-service/internal/database"
	"microservice-golang/services/meta-service/internal/handler"
	"microservice-golang/services/meta-service/internal/repository"
	"microservice-golang/services/meta-service/internal/usecase"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/redisclient"
	"net"
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

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.Database.GormDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("connect to database failed", zap.Error(err))
	}

	if err := database.RunExternalMigrations(cfg.Database.PgDSN()); err != nil {
		log.Fatal("database migration failed", zap.Error(err))
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate failed", zap.Error(err))
	}

	if err := database.Seed(db); err != nil {
		log.Fatal("failed to seed database", zap.Error(err))
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	// TODO integrasiin redis
	_ = redisclient.New(redisClient)

	// ── Repository ────────────────────────────────────────────────────────────
	statusRepo := repository.NewStatusRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// ── UseCase ───────────────────────────────────────────────────────────────
	statusUC := usecase.NewStatusUseCase(statusRepo)
	tagUC := usecase.NewTagUseCase(tagRepo)

	// ── Handler ───────────────────────────────────────────────────────────────
	statusHandler := handler.NewStatusHandler(statusUC)
	tagHandler := handler.NewTagHandler(tagUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	v, err := protovalidate.New()
	if err != nil {
		log.Fatal("failed to initialize validator", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
			interceptor.UnaryValidator(v),
		),
	)

	statusHandler.RegisterGRPC(grpcServer)
	tagHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	log.Info("grpc server listening", zap.Int("port", cfg.GRPC.Port))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
