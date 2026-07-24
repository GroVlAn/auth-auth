package grpc_client

import (
	"context"

	api "github.com/GroVlAn/auth-api/user"
	"github.com/GroVlAn/auth-auth/internal/domain"
	"google.golang.org/grpc"
)

type UserGRPCClient struct {
	client api.UserServiceClient
}

func New(conn *grpc.ClientConn) *UserGRPCClient {
	return &UserGRPCClient{
		client: api.NewUserServiceClient(conn),
	}
}

func (uc *UserGRPCClient) GetUser(ctx context.Context, authUser domain.AuthUser) (domain.User, error) {
	u, err := uc.client.GetUser(ctx, &api.UserQuery{
		ID:       authUser.ID,
		Username: authUser.Username,
		Email:    authUser.Email,
	})
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Fullname:     u.Fullname,
		IsActive:     u.IsActive,
		IsSuperuser:  u.IsSuperuser,
		IsBanned:     u.IsBanned,
	}, nil
}
