package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"net"

	"microservice-golang/services/user-service/internal/config"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/handler"
	"microservice-golang/services/user-service/internal/repository"
	"microservice-golang/services/user-service/internal/usecase"
)

func main() {
	// ── Environment Variables ─────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "microservice-golang")
	appVersion := envConfig.GetEnv("APP_VERSION", "0.0.1")
	serviceName := envConfig.GetEnv("SERVICE_NAME", "user-service")
	serviceVersion := envConfig.GetEnv("SERVICE_VERSION", "0.0.1")

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(env).With(
		zap.String("app_name", appName),
		zap.String("app_version", appVersion),
		zap.String("service", serviceName),
		zap.String("service_version", serviceVersion),
		zap.String("env", env),
	)

	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			panic(err)
		}
	}(log)

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(
		&entity.Permission{},
		&entity.Role{},
		&entity.User{},
	); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// ── Repository ────────────────────────────────────────────────────────────
	userRepo := repository.NewGormUserRepository(db)
	roleRepo := repository.NewGormRoleRepository(db)
	permissionRepo := repository.NewGormPermissionRepository(db)

	// ── Usecase ───────────────────────────────────────────────────────────────
	userUC := usecase.NewUserUseCase(userRepo)
	roleUC := usecase.NewRoleUseCase(roleRepo, permissionRepo, userRepo)
	permissionUC := usecase.NewPermissionUseCase(permissionRepo)

	// ── Handler ───────────────────────────────────────────────────────────────
	userHandler := handler.NewUserHandler(userUC)
	roleHandler := handler.NewRoleHandler(roleUC, permissionUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
		),
	)

	userHandler.RegisterGRPC(grpcServer)
	roleHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	log.Info("user-service gRPC server listening", zap.Int("port", cfg.GRPC.Port))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
