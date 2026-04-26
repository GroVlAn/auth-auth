package domain

import "time"

type Session struct {
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id"`
	RefreshJTI string    `json:"refresh_jti"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}
