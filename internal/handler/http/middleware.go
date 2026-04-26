package http_handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (h *HTTPHandler) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func (h *HTTPHandler) useMiddleware(r *chi.Mux) {
	r.Use(h.Cors)
}
