package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/GroVlAn/auth-api/user"
	"github.com/GroVlAn/auth-auth/internal/config"
	grpchandler "github.com/GroVlAn/auth-auth/internal/handler/grpc-handler"
	httphandler "github.com/GroVlAn/auth-auth/internal/handler/http"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/kbuilder"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/tokens"
	"github.com/GroVlAn/auth-auth/internal/repository"
	grpcserver "github.com/GroVlAn/auth-auth/internal/server/grpc-server"
	httpserver "github.com/GroVlAn/auth-auth/internal/server/http-server"
	"github.com/GroVlAn/auth-auth/internal/service"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	localConfigPath = "configs/config-local.yml"
)

func main() {
	timeStart := time.Now()

	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := config.LoadEnv(); err != nil {
		l.Fatal().Err(err).Msg("failed to load env variables")
	}
	configPath := flag.String("config", localConfigPath, "Path to the configuration file")
	flag.Parse()

	cfg, err := config.New(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load configuration")
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	kBuilder := kbuilder.New(cfg.KeyBuilder.Prev, cfg.KeyBuilder.Version)

	blRepo := repository.NewBlacklistRepository(rc, kBuilder)

	sessionRepo := repository.NewSessionRepository(rc, kBuilder, cfg.Redis.DefaultTimeout)

	tokenizer := tokens.New(cfg.Settings.SecretKey)

	con, err := grpc.NewClient(
		cfg.GRPC.UserApiHost+":"+cfg.GRPC.UserApiPort,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to grpc user service client")
	}

	grpcClient := api.NewUserServiceClient(con)

	s := service.New(
		service.Repos{
			BlacklistRepo: blRepo,
			SessionRepo:   sessionRepo,
		},
		tokenizer,
		grpcClient,
		service.Deps{
			TokenRefreshEndTTL: cfg.Settings.TokenRefreshEndTTL,
			TokenAccessEndTTL:  cfg.Settings.TokenAccessEndTTL,
		},
	)

	h := httphandler.New(
		l,
		s,
		httphandler.Deps{
			BasePath:       cfg.HTTP.BaseHTTPPath,
			DefaultTimeout: cfg.Settings.DefaultTimeout,
		},
	)

	gh := grpchandler.New(l, s, cfg.Settings.DefaultTimeout)

	hServer := httpserver.New(
		h.Handler(),
		httpserver.Settings{
			Port:              cfg.HTTP.Port,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
			ReadHeaderTimeout: time.Duration(cfg.HTTP.ReadHeaderTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		},
	)

	gServer := grpcserver.New(
		gh,
	)

	go func() {
		if err := hServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	go func() {
		l.Info().Msgf("grpc server started on port: %s", cfg.GRPC.Port)

		if err := gServer.ListenAndServe(cfg.GRPC.Port); err != nil {
			l.Fatal().Err(err).Msg("failed to start grpc server")
		}
	}()

	l.Info().Msgf("server start on port: %s load time: %v", cfg.HTTP.Port, time.Since(timeStart))

	<-ctx.Done()
	err = hServer.Shutdown(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to shutdown server")
	} else {
		l.Info().Msg("server shutdown gracefully")
	}
	gServer.Stop()
}
