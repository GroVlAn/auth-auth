package http_handler

import (
	"context"
	"net/http"
	"time"

	"github.com/GroVlAn/auth-auth/internal/domain"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type service interface {
	Authenticate(
		ctx context.Context,
		authUser domain.AuthUser,
		payload domain.UserPayload,
	) (domain.RefreshToken, domain.AccessToken, error)
	RefreshSession(
		ctx context.Context,
		rfToken string,
	) (
		newRefToken domain.RefreshToken,
		newAccToken domain.AccessToken,
		errRef error,
	)
	Logout(ctx context.Context, rfToken string) error
	LogoutAllSession(ctx context.Context, rfToken string) error
	GetUserSessions(
		ctx context.Context,
		rfToken string,
	) ([]domain.UserSession, error)
}

type Deps struct {
	BasePath       string
	DefaultTimeout time.Duration
}

type HTTPHandler struct {
	l zerolog.Logger
	s service
	Deps
}

func New(
	l zerolog.Logger,
	s service,
	deps Deps,
) *HTTPHandler {
	return &HTTPHandler{
		l:    l,
		s:    s,
		Deps: deps,
	}
}

func (h *HTTPHandler) Handler() *chi.Mux {
	r := chi.NewRouter()

	h.useMiddleware(r)

	r.Route("/", func(r chi.Router) {
		r.Get("/home", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome to the Home Page!"))
		})
	})

	r.Route(h.BasePath, func(r chi.Router) {
		h.authRouter(r)
	})

	return r
}
