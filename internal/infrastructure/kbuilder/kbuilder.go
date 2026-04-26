package kbuilder

import (
	"strings"
)

const (
	sessionPref      string = "session"
	userSessionsPref string = "user_session"
	refreshTokenPref string = "refresh_token"
	blacklistPref    string = "blacklist"
	sep              string = ":"
)

type RedisKeyBuilder struct {
	pref string
	v    string
}

func New(pref, v string) *RedisKeyBuilder {
	return &RedisKeyBuilder{
		pref: pref,
		v:    v,
	}
}

func (k *RedisKeyBuilder) SessionKey(sessionID string) string {
	return k.build(k.pref, k.v, sessionPref, sessionID)
}

func (k *RedisKeyBuilder) RefreshKey(jti string) string {
	return k.build(k.pref, k.v, refreshTokenPref, jti)
}

func (k *RedisKeyBuilder) UserSessionsKey(userID string) string {
	return k.build(k.pref, k.v, userSessionsPref, userID)
}

func (k *RedisKeyBuilder) BlacklistKey(jti string) string {
	return k.build(k.pref, k.v, blacklistPref, jti)
}

func (k *RedisKeyBuilder) build(parts ...string) string {
	return strings.Join(parts, sep)
}
