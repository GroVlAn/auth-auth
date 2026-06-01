package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type sessionRepo interface {
	Create(ctx context.Context, session domain.Session) error
	RotateSession(ctx context.Context, session domain.Session, oldJTI, newJTI string) error
	Session(ctx context.Context, sessionID string) (domain.Session, error)
	Sessions(ctx context.Context, sessionIDs []string) ([]domain.Session, error)
	GetSessionIDByRefreshJTI(ctx context.Context, jti string) (string, error)
	UserSessions(ctx context.Context, userID string) ([]string, error)
	DelSession(ctx context.Context, session domain.Session) error
	DelRefreshToken(ctx context.Context, jti string) error
	DelAllSessions(ctx context.Context, userID string, sessions []domain.Session) error
	RemoveUserSession(ctx context.Context, userID, sessionID string) error
}

type blacklistRepo interface {
	AddToBlackList(ctx context.Context, jti string, exp time.Duration) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type tokenizer interface {
	CreateRefreshToken(rc domain.RefreshClaims) (string, error)
	CreateAccessToken(ac domain.AccessClaims) (string, error)
	ParseRefreshToken(token string) (domain.RefreshClaims, error)
	ParseAccessToken(token string) (domain.AccessClaims, error)
}

type Repos struct {
	sessionRepo   sessionRepo
	blacklistRepo blacklistRepo
}

type Deps struct {
	TokenRefreshEndTTL time.Duration
	TokenAccessEndTTL  time.Duration
}

type Service struct {
	tokenizer tokenizer
	Deps
	Repos
}

func New(repos Repos, tokenizer tokenizer, deps Deps) *Service {
	return &Service{
		tokenizer: tokenizer,
		Deps:      deps,
		Repos:     repos,
	}
}

func (s *Service) Authenticate(
	ctx context.Context,
	authUser domain.AuthUser,
	payload domain.UserPayload,
) (domain.RefreshToken, domain.AccessToken, error) {
	// TODO get User
	user := domain.User{
		ID:           "ff2i3f2jf2",
		Username:     "username",
		Email:        "example@example.com",
		PasswordHash: "f3ir2j32ijr",
		Fullname:     "LastName FirstName",
		IsActive:     true,
	}

	if err := s.verifyPassword(user.PasswordHash, authUser.Password); err != nil {
		return domain.RefreshToken{},
			domain.AccessToken{},
			e.NewErrUnauthorized(
				fmt.Errorf("verifying password: %w", err),
				"invalid login or password",
			)
	}

	session := s.createSession(user, payload)

	refToken, accToken, err := s.createTokens(user, session)
	if err != nil {
		return domain.RefreshToken{},
			domain.AccessToken{},
			e.NewErrInternal(fmt.Errorf("creating tokens: %w", err))
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return domain.RefreshToken{},
			domain.AccessToken{},
			e.NewErrInternal(fmt.Errorf("creating session: %w", err))
	}

	return refToken, accToken, nil
}

func (s *Service) RefreshSession(
	ctx context.Context,
	rfToken string,
) (
	newRefToken domain.RefreshToken,
	newAccToken domain.AccessToken,
	errRef error,
) {
	internalErr := func(err error, msg string) {
		errRef = e.NewErrInternal(
			fmt.Errorf("%s: %w", msg, err),
		)
	}
	unauthorizedErr := func(err error, logMsg, errMsg string) {
		errRef = e.NewErrUnauthorized(
			fmt.Errorf("%s: %w", logMsg, err),
			errMsg,
		)
	}

	parsedRefToken, err := s.tokenizer.ParseRefreshToken(rfToken)
	if err != nil {
		unauthorizedErr(
			err,
			"parsing refresh token",
			"invalid refresh token",
		)
		return
	}

	blacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, parsedRefToken.JTI)
	if err != nil {
		internalErr(err, "checking refresh token is blacklisted")
		return
	}

	if blacklisted {
		unauthorizedErr(
			e.ErrBlacklisted,
			"checking refresh token is blacklisted",
			e.ErrBlacklisted.Error(),
		)
		return
	}

	sessionID, err := s.sessionRepo.GetSessionIDByRefreshJTI(ctx, parsedRefToken.JTI)
	if err != nil {
		errRef = s.errGetSessionID(err)
		return
	}

	session, err := s.sessionRepo.Session(ctx, sessionID)
	if err != nil {
		unauthorizedErr(
			err,
			"getting session",
			"session by refresh token not found",
		)
		return
	}

	oldJTI := session.RefreshJTI
	if oldJTI != parsedRefToken.JTI {
		unauthorizedErr(
			e.ErrInvalidToken,
			"refresh token mismatch",
			"invalid refresh token",
		)
		return
	}

	session.RefreshJTI = uuid.NewString()

	// TODO get User
	user := domain.User{
		ID:           "ff2i3f2jf2",
		Username:     "username",
		Email:        "example@example.com",
		PasswordHash: "f3ir2j32ijr",
		Fullname:     "LastName FirstName",
		IsActive:     true,
	}

	newRefToken, newAccToken, err = s.createTokens(user, session)
	if err != nil {
		internalErr(err, "creating mew tokens")
		return
	}

	if err := s.sessionRepo.RotateSession(ctx, session, oldJTI, session.RefreshJTI); err != nil {
		internalErr(err, "updating session")
	}

	return
}

func (s *Service) Logout(ctx context.Context, rfToken string) error {
	parsedRefToken, err := s.tokenizer.ParseRefreshToken(rfToken)
	if err != nil {
		return e.NewErrUnauthorized(
			fmt.Errorf("parsed refresh token: %w", err),
			"invalid token",
		)
	}

	sessionID, err := s.sessionRepo.GetSessionIDByRefreshJTI(ctx, parsedRefToken.JTI)
	if err != nil {
		var errWrapper *e.ErrWrapper

		if errors.As(err, &errWrapper) {
			if errWrapper.ErrorType() == e.ErrorTypeNotFound {
				return nil
			}
		}

		return e.NewErrInternal(
			fmt.Errorf("getting session ID by refresh jti: %w", err),
		)
	}

	session, err := s.sessionRepo.Session(ctx, sessionID)
	if err != nil {
		return e.NewErrInternal(
			fmt.Errorf("getting session: %w", err),
		)
	}

	if err := s.sessionRepo.DelSession(ctx, session); err != nil {
		return e.NewErrInternal(
			fmt.Errorf("deleting session: %w", err),
		)
	}

	return nil
}

func (s *Service) LogoutAllSession(ctx context.Context, rfToken string) error {
	parsedRefToken, err := s.tokenizer.ParseRefreshToken(rfToken)
	if err != nil {
		return e.NewErrUnauthorized(
			fmt.Errorf("parsed refresh token: %w", err),
			"invalid token",
		)
	}

	sessionIDs, err := s.sessionRepo.UserSessions(ctx, parsedRefToken.SUB)
	if err != nil {
		return e.NewErrInternal(
			fmt.Errorf("getting user sessions: %w", err),
		)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	sessions, err := s.sessionRepo.Sessions(ctx, sessionIDs)
	if err != nil {
		return e.NewErrInternal(
			fmt.Errorf("getting sessions: %w", err),
		)
	}

	if err := s.sessionRepo.DelAllSessions(ctx, parsedRefToken.SUB, sessions); err != nil {
		return e.NewErrInternal(
			fmt.Errorf("deleting all sessions: %w", err),
		)
	}

	return nil
}

func (s *Service) GetUserSessions(
	ctx context.Context,
	rfToken string,
) ([]domain.UserSession, error) {
	parsedRefToken, err := s.tokenizer.ParseRefreshToken(rfToken)
	if err != nil {
		return nil, e.NewErrUnauthorized(
			fmt.Errorf("parsed refresh token: %w", err),
			"invalid token",
		)
	}

	sessionIDs, err := s.sessionRepo.UserSessions(ctx, parsedRefToken.SUB)
	if err != nil {
		return nil, e.NewErrInternal(
			fmt.Errorf("getting user sessions: %w", err),
		)
	}

	if len(sessionIDs) == 0 {
		return nil, nil
	}

	sessions, err := s.sessionRepo.Sessions(ctx, sessionIDs)
	if err != nil {
		return nil, e.NewErrInternal(
			fmt.Errorf("getting sessions: %w", err),
		)
	}

	result := make([]domain.UserSession, 0, len(sessions))

	for _, session := range sessions {
		userSession := domain.UserSession{
			ID:           session.ID,
			UserAgent:    session.UserAgent,
			IP:           session.IP,
			CreatedAt:    session.CreatedAt,
			LastActiveAt: session.LastActiveAt,
			Current:      session.ID == parsedRefToken.SID,
		}

		result = append(result, userSession)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].
			LastActiveAt.
			After(result[j].LastActiveAt)
	})

	return result, nil
}

func (s *Service) createSession(user domain.User, payload domain.UserPayload) domain.Session {
	session := domain.Session{
		UserID:     user.ID,
		ID:         uuid.NewString(),
		RefreshJTI: uuid.NewString(),
		UserAgent:  payload.UserAgent,
		IP:         payload.IP,
		CreatedAt:  time.Now(),
	}

	return session
}

func (s *Service) createTokens(user domain.User, session domain.Session) (domain.RefreshToken, domain.AccessToken, error) {
	refreshToken, err := s.createRefreshToken(user, session)
	if err != nil {
		return domain.RefreshToken{}, domain.AccessToken{}, fmt.Errorf("creating refresh token: %w", err)
	}

	accessToken, err := s.createAccessToken(refreshToken.ID, user)
	if err != nil {
		return domain.RefreshToken{}, domain.AccessToken{}, fmt.Errorf("creating access token: %w", err)
	}

	return refreshToken, accessToken, nil
}

func (s *Service) createRefreshToken(user domain.User, session domain.Session) (domain.RefreshToken, error) {
	refreshToken := domain.RefreshToken{}
	refreshToken.StartTTL = time.Now()
	refreshToken.EndTTL = refreshToken.StartTTL.Add(s.TokenRefreshEndTTL)

	token, err := s.tokenizer.CreateRefreshToken(domain.RefreshClaims{
		SUB: user.ID,
		SID: session.ID,
		JTI: session.RefreshJTI,
		IAT: refreshToken.StartTTL.Unix(),
		EXP: refreshToken.EndTTL.Unix(),
	})
	if err != nil {
		return domain.RefreshToken{}, fmt.Errorf("creating refresh token: %w", err)
	}

	refreshToken.ID = uuid.NewString()
	refreshToken.Token = token
	refreshToken.UserID = user.ID

	return refreshToken, nil
}

func (s *Service) createAccessToken(rfID string, user domain.User) (domain.AccessToken, error) {
	accessToken := domain.AccessToken{}
	accessToken.StartTTL = time.Now()
	accessToken.EndTTL = accessToken.StartTTL.Add(s.TokenAccessEndTTL)

	token, err := s.tokenizer.CreateAccessToken(domain.AccessClaims{
		RefreshTokenID: rfID,
		SUB:            user.ID,
		IAT:            accessToken.StartTTL.Unix(),
		EXP:            accessToken.EndTTL.Unix(),
	})
	if err != nil {
		return domain.AccessToken{}, fmt.Errorf("creating access token: %w", err)
	}

	accessToken.ID = uuid.NewString()
	accessToken.Token = token
	accessToken.UserID = user.ID

	return accessToken, nil
}

func (s *Service) verifyPassword(passwordHash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return e.NewErrUnauthorized(
			fmt.Errorf("comparing hash adn password: %w", err),
			"invalid password",
		)
	}
	if err != nil {
		return e.NewErrInternal(fmt.Errorf("comparing hash adn password: %w", err))
	}

	return nil
}

func (s *Service) errGetSessionID(err error) error {
	var errWrapper *e.ErrWrapper

	if errors.As(err, &errWrapper) {
		if errWrapper.ErrorType() == e.ErrorTypeNotFound {
			return e.NewErrUnauthorized(
				fmt.Errorf("getting session id by refresh token: %w", err),
				"session id by refresh token not found",
			)
		}
	}

	return e.NewErrInternal(
		fmt.Errorf("getting session id by refresh token: %w", err),
	)
}
