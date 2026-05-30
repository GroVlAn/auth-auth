package tokens

import (
	"fmt"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/golang-jwt/jwt"
)

type Tokens struct {
	secretKey string
}

func New(
	secretKey string,
) *Tokens {
	return &Tokens{
		secretKey: secretKey,
	}
}

func (t *Tokens) CreateRefreshToken(rc domain.RefreshClaims) (string, error) {
	payload := jwt.MapClaims{
		"sub": rc.SUB,
		"sid": rc.SID,
		"jti": rc.JTI,
		"iat": rc.IAT,
		"exp": rc.EXP,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token, err := jwtToken.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", fmt.Errorf("creating access token: %w", err)
	}

	return token, nil
}

func (t *Tokens) CreateAccessToken(ac domain.AccessClaims) (string, error) {
	payload := jwt.MapClaims{
		"refresh_token_id": ac.RefreshTokenID,
		"sub":              ac.SUB,
		"iat":              ac.IAT,
		"exp":              ac.EXP,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token, err := jwtToken.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", fmt.Errorf("creating access token: %w", err)
	}

	return token, nil
}

func (t *Tokens) ParseRefreshToken(token string) (domain.RefreshClaims, error) {
	tokenDetails, err := t.parseToken(token)
	if err != nil {
		return domain.RefreshClaims{}, fmt.Errorf("parsing token: %w", err)
	}

	tokenClaims := domain.RefreshClaims{
		SUB: tokenDetails["sub"].(string),
		SID: tokenDetails["sid"].(string),
		JTI: tokenDetails["jti"].(string),
		IAT: tokenDetails["IAT"].(int64),
		EXP: tokenDetails["EXP"].(int64),
	}

	if err := tokenClaims.CheckExpired(); err != nil {
		return domain.RefreshClaims{}, fmt.Errorf("checking expired token: %w", err)
	}

	return tokenClaims, nil
}

func (t *Tokens) ParseAccessToken(token string) (domain.AccessClaims, error) {
	tokenDetails, err := t.parseToken(token)
	if err != nil {
		return domain.AccessClaims{}, fmt.Errorf("parsing token: %w", err)
	}

	return domain.AccessClaims{
		RefreshTokenID: tokenDetails["refresh_token_id"].(string),
		SUB:            tokenDetails["sub"].(string),
		IAT:            tokenDetails["iat"].(int64),
		EXP:            tokenDetails["exp"].(int64),
	}, nil
}

func (t *Tokens) parseToken(token string) (jwt.MapClaims, error) {
	tokenClaims := jwt.MapClaims{}

	jwtToken, err := jwt.ParseWithClaims(
		token,
		tokenClaims,
		func(token *jwt.Token) (interface{}, error) {

			switch token.Method.Alg() {
			case jwt.SigningMethodHS256.Alg():
				return []byte(t.secretKey), nil
			default:
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		},
	)
	if err != nil {
		return jwt.MapClaims{}, fmt.Errorf("parsing access token: %w", err)
	}

	tokenDetails, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, e.ErrInvalidToken
	}

	return tokenDetails, nil
}
