package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GroVlAn/auth-auth/internal/config"
	grpcHandler "github.com/GroVlAn/auth-auth/internal/handler/grpc-handler"
	httpHandler "github.com/GroVlAn/auth-auth/internal/handler/http-handler"
	grpcClient "github.com/GroVlAn/auth-auth/internal/infrastructure/grpc-client"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/kbuilder"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/secrets"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/tokens"
	vaultClient "github.com/GroVlAn/auth-auth/internal/infrastructure/vault-client"
	"github.com/GroVlAn/auth-auth/internal/repository"
	grpcServer "github.com/GroVlAn/auth-auth/internal/server/grpc-server"
	httpServer "github.com/GroVlAn/auth-auth/internal/server/http-server"
	"github.com/GroVlAn/auth-auth/internal/service"
	"github.com/GroVlAn/auth-base/crypto"
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

	vc, err := vaultClient.New(vaultClient.Conf{
		SecretToken: cfg.Vault.SecretToken,
		Address:     cfg.Vault.Address,
		Mount:       cfg.Vault.Mount,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load vault client")
	}

	provider := secrets.New(vc, secrets.Paths{
		Token:  cfg.VaultPaths.Token,
		Redis:  cfg.VaultPaths.Redis,
		Hasher: cfg.VaultPaths.Hasher,
	})

	scrt, err := provider.Load(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed load secrets")
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     scrt.Redis.Host + ":" + scrt.Redis.Addr,
		Password: scrt.Redis.Password,
		DB:       scrt.Redis.DB,
	})

	kBuilder := kbuilder.New(cfg.KeyBuilder.Prev, cfg.KeyBuilder.Version)

	blRepo := repository.NewBlacklistRepository(rc, kBuilder)

	sessionRepo := repository.NewSessionRepository(rc, kBuilder, cfg.Redis.DefaultTimeout)

	tokenizer := tokens.New(scrt.Token.SecretKey)

	conn, err := grpc.NewClient(
		cfg.GRPC.UserApiHost+":"+cfg.GRPC.UserApiPort,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to grpc user service client")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			l.Error().Err(err).Msg("failed close grpc connection")
		}
	}()

	grpcClient := grpcClient.New(conn)

	hasher := crypto.New(crypto.Deps{
		Time:    scrt.Hasher.Time,
		Memory:  scrt.Hasher.Memory,
		Threads: scrt.Hasher.Threads,
		KeyLen:  scrt.Hasher.KeyLen,
		SaltLen: scrt.Hasher.SaltLen,
	})

	s := service.New(
		service.Repos{
			BlacklistRepo: blRepo,
			SessionRepo:   sessionRepo,
		},
		tokenizer,
		grpcClient,
		hasher,
		service.Deps{
			TokenRefreshEndTTL: time.Duration(scrt.Token.TokenRefreshEndTTL) * time.Second,
			TokenAccessEndTTL:  time.Duration(scrt.Token.TokenAccessEndTTL) * time.Second,
		},
	)

	h := httpHandler.New(
		l,
		s,
		httpHandler.Deps{
			BasePath:       cfg.HTTP.BaseHTTPPath,
			DefaultTimeout: cfg.Settings.DefaultTimeout,
		},
	)

	gh := grpcHandler.New(l, s, cfg.Settings.DefaultTimeout)

	hServer := httpServer.New(
		h.Handler(),
		httpServer.Settings{
			Port:              cfg.HTTP.Port,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
			ReadHeaderTimeout: time.Duration(cfg.HTTP.ReadHeaderTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		},
	)

	gServer := grpcServer.New(
		gh,
	)

	errCh := make(chan error, 2)

	go func() {
		l.Info().Msgf("starting http server on port: %s", cfg.HTTP.Port)

		err := hServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		l.Info().Msgf("starting grpc server on port: %s", cfg.GRPC.Port)

		if err := gServer.ListenAndServe(cfg.GRPC.Port); err != nil {
			errCh <- err
		}

	}()

	l.Info().
		Dur("startup_time", time.Since(timeStart)).
		Str("http_port", cfg.HTTP.Port).
		Str("grpc_port", cfg.GRPC.Port).
		Msg("server started")

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := hServer.Shutdown(shutdownCtx); err != nil {
			l.Error().Err(err).Msg("failed to shutdown server")
		} else {
			l.Info().Msg("server shutdown gracefully")
		}
		gServer.Stop()
	}

	select {
	case <-ctx.Done():
		shutdown()
	case err := <-errCh:
		if err != nil {
			l.Error().Err(err).Msg("server exited with error")

			shutdown()
		}
	}
}
