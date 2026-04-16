package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"microservice-golang/services/user-service/internal/config"
	"microservice-golang/services/user-service/internal/entity"
	"microservice-golang/services/user-service/internal/handler"
	"microservice-golang/services/user-service/internal/repository"
	"microservice-golang/services/user-service/internal/usecase"
)

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&entity.Permission{},
		&entity.Role{},
		&entity.User{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
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
	grpcServer := grpc.NewServer()
	/*
		grpcServer := grpc.NewServer(
				grpc.UnaryInterceptor(grpc.ChainUnaryInterceptor(
					grpc_recovery.UnaryServerInterceptor(),
					//grpc_logging.LoggingInterceptor(),
					//ValidationInterceptor(),
				)),
			)
	*/

	userHandler.RegisterGRPC(grpcServer)
	roleHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("user-service gRPC server listening on :%d", cfg.GRPC.Port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
