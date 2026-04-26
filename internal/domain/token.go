package domain

import "time"

type AccessToken struct {
	ID       string    `json:"-" db:"id"`
	Token    string    `json:"token" db:"token"`
	StartTTL time.Time `json:"start_ttl" db:"start_ttl"`
	EndTTL   time.Time `json:"end_ttl" db:"end_ttl"`
	UserID   string    `json:"-" db:"user_id"`
}

type RefreshToken struct {
	ID       string    `json:"-" db:"id"`
	Token    string    `json:"token" db:"token"`
	StartTTL time.Time `json:"start_ttl" db:"start_ttl"`
	EndTTL   time.Time `json:"end_ttl" db:"end_ttl"`
	UserID   string    `json:"-" db:"user_id"`
}
