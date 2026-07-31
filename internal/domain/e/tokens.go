package e

import "errors"

var (
	ErrInvalidToken = errors.New("invalid jwt token")
	ErrTokenExpired = errors.New("token is expired")
)
