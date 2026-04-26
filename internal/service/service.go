package service

import (
	"context"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain"
)

type repo interface {
	CreateTokens(ctx context.Context, accToken domain.AccessToken, rfToken domain.RefreshToken, artID string) error
	CreateAccessToken(ctx context.Context, token domain.AccessToken) error
	AccessToken(ctx context.Context, token string) (domain.AccessToken, error)
	DeleteAccessToken(ctx context.Context, token string) error
	DeleteAllAccessTokens(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, token string) (domain.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteAllRefreshTokens(ctx context.Context, userID string) error
}

type AuthDeps struct {
	TokenRefreshEndTTL time.Duration
	TokenAccessEndTTL  time.Duration
	SecretKey          string
}

type AuthService struct {
	repo repo
	AuthDeps
}

func New(repo repo, deps AuthDeps) *AuthService {
	return &AuthService{
		repo:     repo,
		AuthDeps: deps,
	}
}
