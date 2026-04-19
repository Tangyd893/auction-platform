package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"auction-platform/internal/config"
	"auction-platform/internal/interceptor"
	"auction-platform/internal/repository"
	"auction-platform/internal/service"
	pb "auction-platform/proto/gen/auction"
)

func main() {
	// 初始化配置
	config.Init("config.yaml")
	cfg := config.Get()

	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.App.Env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.App.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Str("env", cfg.App.Env).Msg("Starting Auction Platform")

	// 初始化数据库
	db, err := repository.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Database connected")

	// 运行迁移
	if err := repository.RunMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Migrations completed")

	// 初始化 Redis
	redisClient, err := repository.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Redis connected")

	// 初始化 repositories
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)
	bidRepo := repository.NewBidRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cacheRepo := repository.NewCacheRepository(redisClient)

	// 初始化 services
	authService := service.NewAuthService(userRepo, &cfg.JWT)
	itemService := service.NewItemService(itemRepo, cacheRepo)
	bidService := service.NewBidService(bidRepo, itemRepo, cacheRepo)
	orderService := service.NewOrderService(orderRepo, itemRepo)
	userService := service.NewUserService(userRepo)

	// 启动 gRPC 服务器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor()),
		grpc.StreamInterceptor(interceptor.StreamServerInterceptor()),
	)

	pb.RegisterAuctionServiceServer(grpcServer, service.NewAuctionService(
		authService, itemService, bidService, orderService, userService,
	))

	// 注册反射服务（用于 grpcurl 等工具）
	reflection.Register(grpcServer)

	// 启动 HTTP 服务器（用于 HTTP REST API + Prometheus metrics）
	mux := http.NewServeMux()
	setupHTTPRoutes(mux, grpcServer, authService)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port),
		Handler: mux,
	}

	// 启动 gRPC
	grpcLis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen on gRPC port")
	}

	// 启动 HTTP
	go func() {
		log.Info().Int("port", cfg.Server.HTTP.Port).Msg("HTTP server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// 启动 Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info().Int("port", 9090).Msg("Prometheus metrics server starting")
		http.ListenAndServe(":9090", nil)
	}()

	log.Info().Int("port", cfg.Server.GRPC.Port).Msg("gRPC server starting")

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown failed")
	}

	grpcServer.GracefulStop()

	log.Info().Msg("Server exited")
}

func setupHTTPRoutes(mux *http.ServeMux, grpcServer *grpc.Server, authService *service.AuthService) {
	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// gRPC Gateway 可以在这里接入，或使用 grpcweb
	// 目前先用 HTTP handler 直接处理
}
