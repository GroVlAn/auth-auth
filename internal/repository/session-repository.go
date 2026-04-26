package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/domain/e"
)

type SessionRepository struct {
	rc        redisClient
	rkBuilder rkBuilder
}

func NewSessionRepository(rc redisClient, rkBuilder rkBuilder) *SessionRepository {
	return &SessionRepository{
		rc:        rc,
		rkBuilder: rkBuilder,
	}
}

func (sr *SessionRepository) CreateSession(ctx context.Context, session domain.Session, exp time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return e.NewErrInternal(fmt.Errorf("db marshaling session: %w", err))
	}

	key := sr.rkBuilder.SessionKey(session.SessionID)

	if err := sr.rc.Set(ctx, key, data, exp); err != nil {
		return e.NewErrInternal(fmt.Errorf("creating session: %w", err))
	}

	return nil
}

func (sr *SessionRepository) Session(ctx context.Context, sessionID string) (domain.Session, error) {
	key := sr.rkBuilder.SessionKey(sessionID)

	val, err := sr.rc.Get(ctx, key)
	if err != nil {
		err = fmt.Errorf("getting session: %w", err)

		if errors.Is(err, e.ErrRedisNotFound) {
			return domain.Session{}, e.NewErrNotFound(
				err,
				"session not found",
			)
		}

		return domain.Session{}, err
	}

	var session domain.Session
	if err = json.Unmarshal([]byte(val), &session); err != nil {
		return domain.Session{}, e.NewErrInternal(fmt.Errorf("db unmarshaling session: %w", err))
	}

	return session, nil
}

func (sr *SessionRepository) Del(ctx context.Context, sessionID string) error {
	key := sr.rkBuilder.SessionKey(sessionID)

	if err := sr.rc.Del(ctx, key); err != nil {
		return e.NewErrInternal(fmt.Errorf("deleting session: %w", err))
	}

	return nil
}
