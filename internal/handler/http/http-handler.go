package http_handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Deps struct {
	BasePath       string
	DefaultTimeout time.Duration
}

type HTTPHandler struct {
	l zerolog.Logger
	Deps
}

func New(
	l zerolog.Logger,
	deps Deps,
) *HTTPHandler {
	return &HTTPHandler{
		l:    l,
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

	return r
}
