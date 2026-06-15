package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	shortifyv1 "shortify/server/api/gen/shortify/v1"
	"shortify/server/internal/auth"
	"shortify/server/internal/config"
	grpcserver "shortify/server/internal/handler/grpc"
	"shortify/server/internal/handler/rest"
	"shortify/server/internal/middleware"
	"shortify/server/internal/repository"
	"shortify/server/internal/service"
	"shortify/server/internal/worker"
	"shortify/server/pkg/db"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(database)
	linkRepo := repository.NewLinkRepository(database)
	clickRepo := repository.NewClickRepository(database)
	clickWorker := worker.NewClickWorker(clickRepo, 1000)
	workerCtx, stopWorker := context.WithCancel(context.Background())
	clickWorker.Start(workerCtx)

	tokenManager := auth.NewTokenManager(cfg.JWTSecret)
	authService := service.NewAuthService(userRepo, tokenManager)
	linkService := service.NewLinkService(linkRepo, clickRepo, clickWorker, cfg.BaseURL)

	authHandler := rest.NewAuthHandler(authService)
	linkHandler := rest.NewLinkHandler(linkService)

	mux := http.NewServeMux()
	rest.RegisterRoutes(
		mux,
		authHandler,
		linkHandler,
		middleware.Auth(tokenManager),
	)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: middleware.CORS(mux),
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor(tokenManager)),
	)
	shortifyv1.RegisterAuthServiceServer(grpcServer, grpcserver.NewAuthServer(authService))
	shortifyv1.RegisterLinkServiceServer(grpcServer, grpcserver.NewLinkServer(linkService))
	reflection.Register(grpcServer)

	go func() {
		logger.Info("REST started", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("REST stopped with error", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			logger.Error("gRPC listen failed", "err", err)
			os.Exit(1)
		}

		logger.Info("gRPC started", "addr", ":"+cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC stopped with error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(ctx)
	stopWorker()
	clickWorker.Stop()
	fmt.Println("shutdown complete")
}
