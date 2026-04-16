package main

import (
	"fmt"
	"log"
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
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
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
		log.Fatalf("failed to connect to user-service: %v", err)
	}
	defer userConn.Close()

	userClient := userv1.NewUserInternalServiceClient(userConn)

	// ── Usecase ───────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUseCase(userClient, jwtManager, redisClient, cfg.JWT.RefreshTTL)

	// ── Handler ───────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer()

	authHandler.RegisterGRPC(grpcServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("auth-service gRPC server listening on :%d", cfg.GRPC.Port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
