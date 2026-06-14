package grpc_handler

import (
	"context"

	api "github.com/GroVlAn/auth-api/auth"
	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/infrastructure/requestinfo"
	"github.com/GroVlAn/auth-base/ew/grpcx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GRPCHandler) Auth(ctx context.Context, reqAuthUser *api.AuthUser) (*api.Tokens, error) {
	authUser := domain.AuthUser{
		Username: reqAuthUser.Username,
		Email:    reqAuthUser.Email,
		Password: reqAuthUser.Password,
	}

	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	userPayload := domain.UserPayload{
		UserAgent: requestinfo.GetUserAgent(ctx),
		IP:        requestinfo.GetIP(ctx),
	}

	rfToken, accToken, err := h.s.Authenticate(ctx, authUser, userPayload)
	if err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Tokens{
		RefreshToken: &api.RefreshToken{
			Token:     rfToken.Token,
			Start_TTL: timestamppb.New(rfToken.StartTTL),
			End_TTL:   timestamppb.New(rfToken.EndTTL),
			User_ID:   rfToken.UserID,
		},
		AccessToken: &api.AccessToken{
			Token:     accToken.Token,
			Start_TTL: timestamppb.New(accToken.StartTTL),
			End_TTL:   timestamppb.New(accToken.EndTTL),
			User_ID:   accToken.UserID,
		},
	}, nil
}

func (h *GRPCHandler) Refresh(ctx context.Context, reqRefToken *api.RefreshToken) (*api.Tokens, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	rfToken, accToken, err := h.s.RefreshSession(ctx, reqRefToken.Token)
	if err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Tokens{
		RefreshToken: &api.RefreshToken{
			Token:     rfToken.Token,
			Start_TTL: timestamppb.New(rfToken.StartTTL),
			End_TTL:   timestamppb.New(rfToken.EndTTL),
			User_ID:   rfToken.UserID,
		},
		AccessToken: &api.AccessToken{
			Token:     accToken.Token,
			Start_TTL: timestamppb.New(accToken.StartTTL),
			End_TTL:   timestamppb.New(accToken.EndTTL),
			User_ID:   accToken.UserID,
		},
	}, nil
}

func (h *GRPCHandler) Logout(ctx context.Context, reqRefToken *api.RefreshToken) (*api.Success, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.Logout(ctx, reqRefToken.Token); err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Success{
		Success: true,
	}, nil
}

func (h *GRPCHandler) LogoutAll(ctx context.Context, reqRefToken *api.RefreshToken) (*api.Success, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.LogoutAllSession(ctx, reqRefToken.Token); err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Success{
		Success: true,
	}, nil
}

func (h *GRPCHandler) GetUserSessions(ctx context.Context, reqRefToken *api.RefreshToken) (*api.UserSessions, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	userSessions, err := h.s.GetUserSessions(ctx, reqRefToken.Token)
	if err != nil {
		return nil, grpcx.HandleError(err)
	}

	respUserSessions := make([]*api.UserSession, 0, len(userSessions))

	for _, session := range userSessions {
		respSession := &api.UserSession{
			ID:           session.ID,
			UserAgent:    session.UserAgent,
			IP:           session.IP,
			CreatedAt:    timestamppb.New(session.CreatedAt),
			LastActiveAt: timestamppb.New(session.LastActiveAt),
			Current:      session.Current,
		}

		respUserSessions = append(respUserSessions, respSession)
	}

	return &api.UserSessions{
		UserSessions: respUserSessions,
	}, nil
}
