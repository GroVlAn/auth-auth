package http_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/GroVlAn/auth-base/ew"
	"github.com/GroVlAn/auth-base/ew/httpx"
)

func (h *HTTPHandler) extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ew.New(
			ew.ErrorTypeUnauthorized,
			errors.New("authorization header is missing"),
		).Msg("authorization header is missing")
	}

	// Разделяем заголовок по пробелу
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ew.New(
			ew.ErrorTypeUnauthorized,
			errors.New("invalid authorization format"),
		).Msg("invalid authorization format")
	}

	return parts[1], nil
}

func (h *HTTPHandler) sendResponse(w http.ResponseWriter, res domain.Response, status int) {
	b, err := json.Marshal(res)
	if err != nil {
		h.l.Error().Err(err).Msg("failed marshal response")
	}

	w.WriteHeader(status)

	_, err = w.Write(b)
	if err != nil {
		h.l.Error().Err(err).Msg("failed write response")
	}
}

func (h *HTTPHandler) handleError(w http.ResponseWriter, err error) {
	respErr := httpx.HandleError(err)

	resp := domain.Response{
		Error: &domain.ErrorResponse{
			Code: respErr.Status,
			Text: respErr.Message,
		},
		Data: respErr.Fields,
	}

	h.l.Err(err).Msg(respErr.LogMsg)

	h.sendResponse(w, resp, respErr.Status)
}

func (h *HTTPHandler) handleDecodeBody(w http.ResponseWriter, err error) {
	ev := ew.NewErrValidation("failed read request body")

	switch e := err.(type) {
	case *json.SyntaxError:
		ev.AddField("body", fmt.Sprintf("invalid JSON syntax at offset %d", e.Offset))
	case *json.UnmarshalTypeError:
		ev.AddField(e.Field, fmt.Sprintf("expected %v but got %v", e.Type, e.Value))
	default:
		if errors.Is(err, io.EOF) {
			ev.AddField("body", "empty request body")
		} else {
			ev.AddField("body", err.Error())
		}
	}

	h.l.Error().Err(err).Msg("failed to decode request body")

	h.handleError(w, ev)
}

func (h *HTTPHandler) withBodyClose(body io.ReadCloser, fn func(io.ReadCloser)) {
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			h.l.Error().Err(err).Msg("failed to close request body")
		}
	}(body)

	fn(body)
}

func (h *HTTPHandler) sendInternalError(w http.ResponseWriter, logMessage string) {
	h.l.Error().Msg(logMessage)

	w.WriteHeader(http.StatusInternalServerError)

	res := domain.Response{
		Error: &domain.ErrorResponse{
			Code: http.StatusInternalServerError,
			Text: "internal server error",
		},
	}

	h.sendResponse(w, res, http.StatusInternalServerError)
}

func (h *HTTPHandler) userIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarder-For", "X-Real-IP"} {
		addresses := r.Header.Get(header)
		if addresses != "" {
			addressList := strings.Split(addresses, ",")
			ip := strings.TrimSpace(addressList[0])
			if ip != "" {
				return ip
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
