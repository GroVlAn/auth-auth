package e

import "errors"

var (
	ErrBlacklisted = errors.New("token is blacklisted")
)
