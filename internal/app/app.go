package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kakitomeru/shared/config"
	"github.com/kakitomeru/shared/env"
	"github.com/kakitomeru/shared/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	"nota.auth/internal/api"
	"nota.auth/internal/interceptor"
	"nota.auth/internal/metric"
	"nota.auth/internal/repository"
	"nota.auth/internal/service"
	authpb "nota.auth/pkg/pb/v1"
)

type App struct {
	db     *gorm.DB
	cfg    *config.Auth
	port   string
	server *grpc.Server
}

func NewApp(db *gorm.DB, cfg *config.Auth) *App {
	return &App{
		db:   db,
		cfg:  cfg,
		port: env.GetAuthPort(),
	}
}

func (a *App) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	shutdownTracer, err := telemetry.NewTracerProvider(ctx, a.cfg.Name, env.GetOtelCollector())
	if err != nil {
		log.Fatalf("failed to create tracer provider: %v", err)
	}
	defer func() {
		if err := shutdownTracer(ctx); err != nil {
			log.Printf("failed to shutdown tracer provider: %v", err)
		}
	}()

	shutdownMeter, err := telemetry.NewMeterProvider(ctx, a.cfg.Name, env.GetOtelCollector())
	if err != nil {
		log.Fatalf("failed to create meter provider: %v", err)
	}
	defer func() {
		if err := shutdownMeter(ctx); err != nil {
			log.Printf("failed to shutdown meter provider: %v", err)
		}
	}()
	metric.Init()

	statsHandler, customInterceptor := telemetry.NewGRPCServerHandlers()

	authInterceptor := interceptor.AuthUnaryServerInterceptor(a.cfg.ProtectedRPC)

	a.server = grpc.NewServer(
		statsHandler,
		grpc.ChainUnaryInterceptor(
			customInterceptor,
			authInterceptor,
		),
	)

	sessionCfg, err := config.LoadSession()
	if err != nil {
		return fmt.Errorf("failed to load session config: %w", err)
	}

	jwtCfg, err := config.LoadJwt()
	if err != nil {
		return fmt.Errorf("failed to load jwt config: %w", err)
	}

	repo := repository.NewRepository(a.db, sessionCfg)
	service := service.NewService(repo, jwtCfg)

	authHandler := api.NewAuthServiceHandler(service)
	authpb.RegisterAuthServiceServer(a.server, authHandler)

	reflection.Register(a.server)

	go func() {
		lis, err := net.Listen("tcp", ":"+a.port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("Starting gRPC server on port %s", a.port)
		if err := a.server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return a.GracefulShutdown(ctx)
}

func (a *App) GracefulShutdown(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Shutdown requested via context")
	case <-quit:
		log.Println("Shutdown requested via signal")
	}

	log.Println("Shutting down gRPC server...")
	a.server.GracefulStop()
	log.Println("Server gracefully stopped")

	return nil
}
