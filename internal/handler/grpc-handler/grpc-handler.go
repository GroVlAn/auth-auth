package grpc_handler

import (
	"context"
	"time"

	api "github.com/GroVlAn/auth-api/auth"
	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/rs/zerolog"
)

type service interface {
	Authenticate(
		ctx context.Context,
		authUser domain.AuthUser,
		payload domain.UserPayload,
	) (domain.RefreshToken, domain.AccessToken, error)
	RefreshSession(
		ctx context.Context,
		rfToken string,
	) (
		newRefToken domain.RefreshToken,
		newAccToken domain.AccessToken,
		errRef error,
	)
	Logout(ctx context.Context, rfToken string) error
	LogoutAllSession(ctx context.Context, rfToken string) error
	GetUserSessions(
		ctx context.Context,
		rfToken string,
	) ([]domain.UserSession, error)
}

type GRPCHandler struct {
	api.UnimplementedAuthServiceServer
	l              zerolog.Logger
	s              service
	defaultTimeout time.Duration
}

func New(l zerolog.Logger, s service, defTimeout time.Duration) *GRPCHandler {
	return &GRPCHandler{
		l:              l,
		s:              s,
		defaultTimeout: defTimeout,
	}
}
