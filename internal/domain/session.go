package domain

import "time"

type UserSession struct {
	ID           string    `json:"id"`
	UserAgent    string    `json:"user_agent"`
	IP           string    `json:"ip"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	Current      bool      `json:"current"`
}

type Session struct {
	ID         string `json:"session_id"`
	UserID     string `json:"user_id"`
	RefreshJTI string `json:"refresh_jti"`

	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`

	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}
