package http_handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (h *HTTPHandler) cors(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   h.mConf.AllowedOrigins,
		AllowedMethods:   h.mConf.AllowedMethods,
		AllowedHeaders:   h.mConf.AllowedHeaders,
		ExposedHeaders:   h.mConf.ExposedHeaders,
		AllowCredentials: h.mConf.AllowCredentials,
		MaxAge:           h.mConf.MaxAge,
	}))
}
