package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/services/scout-service/internal/config"
	"microservice-golang/services/scout-service/internal/database"
	"microservice-golang/services/scout-service/internal/delivery/nats"
	"microservice-golang/services/scout-service/internal/handler"
	"microservice-golang/services/scout-service/internal/repository"
	"microservice-golang/services/scout-service/internal/usecase"
	envConfig "microservice-golang/shared/pkg/config"
	"microservice-golang/shared/pkg/grpc/interceptor"
	"microservice-golang/shared/pkg/logger"
	"microservice-golang/shared/pkg/messaging"

	"buf.build/go/protovalidate"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// ── Env ───────────────────────────────────────────────────────────────────
	_ = godotenv.Load()
	env := envConfig.GetEnv("APP_ENV", "development")
	appName := envConfig.GetEnv("APP_NAME", "sport-sentral")
	appVersion := envConfig.GetEnv("APP_VERSION", "0.0.1")
	serviceName := envConfig.GetEnv("SERVICE_NAME", "scout-service")
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
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("connect to database failed", zap.Error(err))
	}

	// ── NATS ──────────────────────────────────────────────────────────────────
	natsClient, err := messaging.Connect(cfg.Nats, log)
	if err != nil {
		log.Fatal("failed to connect to nats server", zap.Error(err))
	}
	defer natsClient.Drain()

	// ── Repositories ──────────────────────────────────────────────────────────
	scoutRepo := repository.NewScoutRepository(db)

	// ── Use Cases ─────────────────────────────────────────────────────────────
	scoutUsecase := usecase.NewScoutUsecase(scoutRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	scoutHandler := handler.NewScoutHandler(scoutUsecase)

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

	scoutv1.RegisterScoutServiceServer(grpcServer, scoutHandler)
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
		log.Info("scout-service gRPC listening", zap.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server stopped", zap.Error(err))
			cancel()
		}
	}()

	// ── NATS Subscribers ──────────────────────────────────────────────────────
	competitionDurableName := envConfig.GetEnv("COMPETITION_DURABLE_NAME", "scout-service-competition-sync")
	compSub := nats.NewCompetitionSubscriber(
		scoutRepo,
		natsClient,
		log,
		competitionDurableName,
	)

	go func() {
		log.Info("starting competition event consumer")
		if err := compSub.Listen(ctx); err != nil {
			log.Error("competition consumer stopped", zap.Error(err))
		}
	}()

	// Block until context is done (graceful shutdown signal)
	<-ctx.Done()
	log.Info("shutting down gracefully")
	grpcServer.GracefulStop()
}
