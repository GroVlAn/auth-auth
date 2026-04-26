package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
)

type UserSessionRepository struct {
	rc        redisClient
	rkBuilder rkBuilder
}

func NewUserSessionRepository(rc redisClient, rkBuilder rkBuilder) *UserSessionRepository {
	return &UserSessionRepository{
		rc:        rc,
		rkBuilder: rkBuilder,
	}
}

func (ur *UserSessionRepository) AddUserSession(ctx context.Context, userID, sessionID string) error {
	key := ur.rkBuilder.UserSessionsKey(userID)

	if err := ur.rc.SAdd(ctx, key, sessionID); err != nil {
		return e.NewErrInternal(fmt.Errorf("creating user session: %w", err))
	}

	return nil
}

func (ur *UserSessionRepository) UserSessions(ctx context.Context, userID string) ([]string, error) {
	key := ur.rkBuilder.UserSessionsKey(userID)

	vals, err := ur.rc.SMembers(ctx, key)
	if err != nil {
		err := fmt.Errorf("getting user session: %w", err)

		if errors.Is(err, e.ErrRedisNotFound) {
			return nil, e.NewErrNotFound(err, "user session not found")
		}

		return nil, err
	}

	return vals, nil
}

func (ur *UserSessionRepository) RemoveUserSession(ctx context.Context, userID, sessionID string) error {
	key := ur.rkBuilder.UserSessionsKey(userID)

	if err := ur.rc.SRem(ctx, key, sessionID); err != nil {
		return e.NewErrInternal(fmt.Errorf("removing user session: %w", err))
	}

	return nil
}
