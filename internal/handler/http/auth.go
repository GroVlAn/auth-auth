package http_handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"github.com/go-chi/chi"
)

const (
	authenticateEndpoint   = "/auth"
	refreshSessionEndpoint = "/auth/refresh"
	logoutEndpoint         = "/logout"
	logoutAllEndpoint      = "/logout/all"
	userSessionsEndpoint   = "/sessions"

	refreshCookieName = "refresh-token"
)

func (h *HTTPHandler) authRouter(r chi.Router) {
	r.Post(authenticateEndpoint, h.auth)
	r.Patch(refreshSessionEndpoint, h.refresh)
	r.Delete(logoutEndpoint, h.logout)
	r.Delete(logoutAllEndpoint, h.logoutAll)
	r.Get(userSessionsEndpoint, h.userSessions)
}

func (h *HTTPHandler) auth(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(body io.ReadCloser) {
		var authUser domain.AuthUser

		if err := json.NewDecoder(body).Decode(&authUser); err != nil {
			h.handleDecodeBody(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		userPayload := domain.UserPayload{
			UserAgent: r.UserAgent(),
			IP:        h.userIP(r),
		}

		rfToken, accToken, err := h.s.Authenticate(ctx, authUser, userPayload)
		if err != nil {
			status, res := h.handleError(err)

			h.sendResponse(w, res, status)
			return
		}

		refreshTokenCookie := http.Cookie{
			Name:     refreshCookieName,
			Value:    rfToken.Token,
			Expires:  rfToken.EndTTL,
			MaxAge:   rfToken.EndTTL.Second(),
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		}

		http.SetCookie(w, &refreshTokenCookie)

		res := domain.Response{}
		res.Data = map[string]any{
			"access_token": accToken.Token,
		}

		h.sendResponse(w, res, http.StatusOK)
	})
}

func (h *HTTPHandler) refresh(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(rc io.ReadCloser) {
		token := h.token(w, r)
		if len(token) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		rfToken, accToken, err := h.s.RefreshSession(ctx, token)
		if err != nil {
			status, res := h.handleError(err)

			h.sendResponse(w, res, status)
			return
		}

		refreshTokenCookie := http.Cookie{
			Name:     refreshCookieName,
			Value:    rfToken.Token,
			Expires:  rfToken.EndTTL,
			MaxAge:   rfToken.EndTTL.Second(),
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		}

		http.SetCookie(w, &refreshTokenCookie)

		res := domain.Response{}
		res.Data = map[string]any{
			"access_token": accToken.Token,
		}

		h.sendResponse(w, res, http.StatusOK)
	})
}

func (h *HTTPHandler) logout(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(rc io.ReadCloser) {
		token := h.token(w, r)
		if len(token) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		if err := h.s.Logout(ctx, token); err != nil {
			status, res := h.handleError(err)

			h.sendResponse(w, res, status)
			return
		}

		res := domain.Response{}
		res.Response = map[string]interface{}{
			"message": "access token for the current device has been revoked",
		}

		h.sendResponse(w, res, http.StatusOK)
	})
}

func (h *HTTPHandler) logoutAll(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(rc io.ReadCloser) {
		token := h.token(w, r)
		if len(token) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		if err := h.s.LogoutAllSession(ctx, token); err != nil {
			status, res := h.handleError(err)

			h.sendResponse(w, res, status)
			return
		}

		res := domain.Response{}
		res.Response = map[string]interface{}{
			"message": "access tokens for all devices have been revoked",
		}

		h.sendResponse(w, res, http.StatusOK)
	})
}

func (h *HTTPHandler) userSessions(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(rc io.ReadCloser) {
		token := h.token(w, r)
		if len(token) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		userSessions, err := h.s.GetUserSessions(ctx, token)
		if err != nil {
			status, res := h.handleError(err)

			h.sendResponse(w, res, status)
			return
		}

		res := domain.Response{
			Data: userSessions,
		}

		h.sendResponse(w, res, http.StatusOK)
	})
}

func (h *HTTPHandler) token(w http.ResponseWriter, r *http.Request) string {
	tokenCookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			status, res := h.handleError(e.NewErrUnauthorized(
				errors.New("authorization header is missing"),
				"authorization header is missing",
			))

			h.sendResponse(w, res, status)
			return ""
		}

		status, res := h.handleError(err)

		h.sendResponse(w, res, status)
		return ""
	}

	return tokenCookie.Value
}
