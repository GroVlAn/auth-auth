package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/redis/go-redis/v9"
)

type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}

type SessionRepository struct {
	client *redis.Client
	keys   rkBuilder
	ttl    time.Duration
}

func NewSessionRepository(client *redis.Client, keys rkBuilder, timeout time.Duration) *SessionRepository {
	return &SessionRepository{
		client: client,
		keys:   keys,
		ttl:    timeout,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session domain.Session) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return e.NewErrInternal(fmt.Errorf("marshaling session: %w", err))
	}

	pipe := r.client.TxPipeline()

	pipe.Set(
		ctx,
		r.keys.SessionKey(session.ID),
		sessionData,
		r.ttl,
	)

	pipe.Set(
		ctx,
		r.keys.RefreshKey(session.RefreshJTI),
		session.ID,
		r.ttl,
	)

	pipe.SAdd(
		ctx,
		r.keys.UserSessionsKey(session.UserID),
		session.ID,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return e.NewErrInternal(fmt.Errorf("executing pipeline to create new session: %w", err))
	}

	return nil
}

func (r *SessionRepository) RotateSession(ctx context.Context, session domain.Session, oldJTI, newJTI string) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return e.NewErrInternal(fmt.Errorf("marshaling session: %w", err))
	}

	oldRefreshKey := r.keys.RefreshKey(oldJTI)

	err = r.client.Watch(ctx, func(tx *redis.Tx) error {
		val, err := tx.Get(ctx, oldRefreshKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return e.NewErrUnauthorized(
					fmt.Errorf("refresh token already rotated: %w", err),
					"refresh token already used",
				)
			}

			return e.NewErrInternal(
				fmt.Errorf("get refresh token: %w", err),
			)
		}

		if val != session.ID {
			return e.NewErrUnauthorized(
				fmt.Errorf("refresh token session mismatch"),
				"invalid refresh token",
			)
		}

		pipe := r.client.TxPipeline()

		pipe.Del(
			ctx,
			r.keys.RefreshKey(oldJTI),
		)

		pipe.Set(
			ctx,
			r.keys.RefreshKey(newJTI),
			session.ID,
			r.ttl,
		)

		pipe.Set(
			ctx,
			r.keys.SessionKey(session.ID),
			sessionData,
			r.ttl,
		)

		pipe.Set(
			ctx,
			r.keys.BlacklistKey(oldJTI),
			"1",
			r.ttl,
		)

		if _, err := pipe.Exec(ctx); err != nil {
			if errors.Is(err, redis.TxFailedErr) {
				return e.NewErrConflict(
					fmt.Errorf("refresh token race detected: %w", err),
					"refresh token already used",
				)
			}
			return e.NewErrInternal(
				fmt.Errorf("executing pipeline to update session: %w", err),
			)
		}

		return nil
	}, oldRefreshKey)

	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepository) Session(ctx context.Context, sessionID string) (domain.Session, error) {
	key := r.keys.SessionKey(sessionID)

	value, err := r.client.Get(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return domain.Session{}, e.NewErrNotFound(
			fmt.Errorf("getting session: %w", err),
			"session not found",
		)
	case err != nil:
		return domain.Session{}, e.NewErrInternal(fmt.Errorf("getting session: %w", err))
	default:
		var session domain.Session

		if err := json.Unmarshal([]byte(value), &session); err != nil {
			return domain.Session{}, e.NewErrInternal(fmt.Errorf("unmarshaling session: %w", err))
		}

		return session, nil
	}
}

func (r *SessionRepository) Sessions(ctx context.Context, sessionIDs []string) ([]domain.Session, error) {
	pipe := r.client.TxPipeline()

	cmds := make([]*redis.StringCmd, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		cmd := pipe.Get(
			ctx,
			r.keys.SessionKey(sessionID),
		)

		cmds = append(cmds, cmd)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, e.NewErrInternal(
			fmt.Errorf("executing pipeline to getting sessions: %w", err),
		)
	}

	sessions := make([]domain.Session, 0, len(cmds))

	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}
		var session domain.Session

		if err := json.Unmarshal([]byte(val), &session); err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *SessionRepository) GetSessionIDByRefreshJTI(ctx context.Context, jti string) (string, error) {
	key := r.keys.RefreshKey(jti)

	sessionID, err := r.client.Get(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", e.NewErrNotFound(
			fmt.Errorf("getting session id by refresh token: %w", err),
			"session id not found",
		)
	case err != nil:
		return "", e.NewErrInternal(fmt.Errorf("getting session id by refresh token: %w", err))
	default:
		return sessionID, nil
	}
}

func (r *SessionRepository) UserSessions(ctx context.Context, userID string) ([]string, error) {
	key := r.keys.UserSessionsKey(userID)

	sessionIDs, err := r.client.SMembers(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, e.NewErrNotFound(
			fmt.Errorf("getting user sessions: %w", err),
			"user sessions not found",
		)
	case err != nil:
		return nil, e.NewErrInternal(fmt.Errorf("getting user sessions: %w", err))
	default:
		return sessionIDs, nil
	}
}

func (r *SessionRepository) DelSession(ctx context.Context, session domain.Session) error {
	pipe := r.client.TxPipeline()

	pipe.Del(
		ctx,
		r.keys.SessionKey(session.ID),
	)

	pipe.Del(
		ctx,
		r.keys.RefreshKey(session.RefreshJTI),
	)

	pipe.SRem(
		ctx,
		r.keys.UserSessionsKey(session.UserID),
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return e.NewErrInternal(
			fmt.Errorf("executing pipeline to delete session: %w", err),
		)
	}

	return nil
}

func (r *SessionRepository) DelAllSessions(ctx context.Context, userID string, sessions []domain.Session) error {
	pipe := r.client.TxPipeline()

	for _, session := range sessions {
		pipe.Del(
			ctx,
			r.keys.SessionKey(session.ID),
		)

		pipe.Del(
			ctx,
			r.keys.RefreshKey(session.RefreshJTI),
		)
	}

	pipe.Del(
		ctx,
		r.keys.UserSessionsKey(userID),
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return e.NewErrInternal(
			fmt.Errorf("executing pipeline to delete all sessions: %w", err),
		)
	}

	return nil
}

func (r *SessionRepository) DelRefreshToken(ctx context.Context, jti string) error {
	key := r.keys.RefreshKey(jti)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return e.NewErrInternal(fmt.Errorf("deleting refresh token: %w", err))
	}

	return nil
}

func (r *SessionRepository) RemoveUserSession(ctx context.Context, userID, sessionID string) error {
	key := r.keys.UserSessionsKey(userID)

	if err := r.client.SRem(ctx, key, sessionID).Err(); err != nil {
		return e.NewErrInternal(fmt.Errorf("removing user session: %w", err))
	}

	return nil
}
