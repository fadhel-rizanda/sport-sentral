package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	userv1 "microservice-golang/gen/user/v1"
	"microservice-golang/services/auth-service/internal/config"
	"microservice-golang/services/auth-service/internal/handler"
	"microservice-golang/services/auth-service/internal/usecase"
	"microservice-golang/shared/pkg/jwt"
)

func main() {
	// ── Environment Variables ─────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "microservice-golang")
	appVersion := envConfig.GetEnv("APP_VERSION", "0.0.1")
	serviceName := envConfig.GetEnv("SERVICE_NAME", "auth-service")
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

	// ── JWT Manager ───────────────────────────────────────────────────────────
	jwtManager := jwt.NewManager(jwt.Config{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessTTL:     cfg.JWT.AccessTTL,
		RefreshTTL:    cfg.JWT.RefreshTTL,
	})

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// ── User Service gRPC Client ──────────────────────────────────────────────
	userConn, err := grpc.NewClient(
		cfg.UserService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to user-service", zap.Error(err))
	}
	defer func(userConn *grpc.ClientConn) {
		err := userConn.Close()
		if err != nil {
			panic(err)
		}
	}(userConn)

	userClient := userv1.NewUserInternalServiceClient(userConn)

	// ── Usecase ───────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUseCase(userClient, jwtManager, redisClient, cfg.JWT.RefreshTTL)

	// ── Handler ───────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
		),
	)

	authHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	log.Info("auth-service gRPC server listening", zap.Int("port", cfg.GRPC.Port))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
