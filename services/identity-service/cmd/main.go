package main

import (
	"errors"
	"fmt"
	"microservice-golang/shared/pkg/redisclient"
	"net"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"microservice-golang/services/identity-service/internal/config"
	"microservice-golang/services/identity-service/internal/entity"
	"microservice-golang/services/identity-service/internal/handler"
	"microservice-golang/services/identity-service/internal/repository"
	"microservice-golang/services/identity-service/internal/usecase"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/jwt"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/mailer"
)

func main() {
	// ── Env ───────────────────────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "sportcentral")
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

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	if err := migrate(db); err != nil {
		log.Fatal("failed to migrate database", zap.Error(err))
	}

	if err := seed(db); err != nil {
		log.Fatal("failed to seed database", zap.Error(err))
	}

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

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	statusRepo := repository.NewStatusRepository(db)

	// ── Usecases ──────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUseCase(userRepo, userRoleRepo, jwtManager, redisClient, cfg.JWT.RefreshTTL)
	userUC := usecase.NewUserUseCase(userRepo, userRoleRepo, roleRepo, statusRepo, mailerClient, redisWrapper, cfg.AppURL)
	profileUC := usecase.NewProfileUseCase(userRepo, userRoleRepo, roleRepo, statusRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)
	profileHandler := handler.NewProfileHandler(profileUC)
	userInternalHandler := handler.NewUserInternalHandler(userUC)

	// ── gRPC Server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryLogger(log),
			interceptor.UnaryRecovery(log),
		),
	)

	authHandler.RegisterGRPC(grpcServer)
	userHandler.RegisterGRPC(grpcServer)
	profileHandler.RegisterGRPC(grpcServer)
	userInternalHandler.RegisterGRPC(grpcServer)

	reflection.Register(grpcServer)

	// ── Listen ────────────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	log.Info("identity-service gRPC listening", zap.Int("port", cfg.GRPC.Port))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}

// ─── Migration ────────────────────────────────────────────────────────────────

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Status{},
		&entity.Permission{},
		&entity.Role{},
		&entity.User{},
		&entity.UserRole{},
	)
}

// ─── Seed ─────────────────────────────────────────────────────────────────────

func seed(db *gorm.DB) error {
	if err := seedStatuses(db); err != nil {
		return err
	}
	return seedRoles(db)
}

func seedStatuses(db *gorm.DB) error {
	statuses := []entity.Status{
		{Type: entity.StatusTypeUser, Name: entity.UserStatusActive},
		{Type: entity.StatusTypeUser, Name: entity.UserStatusPending},
		{Type: entity.StatusTypeUser, Name: entity.UserStatusBanned},
		{Type: entity.StatusTypeUserRole, Name: entity.UserRoleStatusActive},
		{Type: entity.StatusTypeUserRole, Name: entity.UserRoleStatusPending},
	}

	for _, s := range statuses {
		var existing entity.Status
		err := db.Where("type = ? AND name = ?", s.Type, s.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.ID = uuid.New()
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []entity.Role{
		{Name: entity.RoleAthlete, Description: "Default role. Can join academy, competitions, book courts."},
		{Name: entity.RoleScout, Description: "Talent finder. Freemium access to athlete profiles and leaderboard."},
		{Name: entity.RoleCourtOwner, Description: "Manages courts. Requires admin verification."},
		{Name: entity.RoleAcademyAdmin, Description: "Manages academies. Requires admin verification."},
		{Name: entity.RoleRegulator, Description: "Official sport body. Assigned by platform admin only."},
		{Name: entity.RolePlatformAdmin, Description: "Internal platform administrator."},
	}

	for _, r := range roles {
		var existing entity.Role
		err := db.Where("name = ?", r.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.ID = uuid.New()
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
