package domain

import (
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
)

type RefreshClaims struct {
	SUB string
	SID string
	JTI string
	IAT int64
	EXP int64
}

type AccessClaims struct {
	RefreshTokenID string
	SUB            string
	IAT            int64
	EXP            int64
}

func (rc RefreshClaims) CheckExpired() error {
	return checkExpiredToken(rc.EXP)
}

func (ac AccessClaims) CheckExpired() error {
	return checkExpiredToken(ac.EXP)
}

func checkExpiredToken(expToken int64) error {
	exp := time.Unix(int64(expToken), 0)
	now := time.Now()

	if now.After(exp) {
		return e.ErrTokenExpired
	}

	return nil
}
