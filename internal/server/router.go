package server

import (
	"ndk/internal/handler"
	"ndk/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.HealthCheck)
		r.Get("/cert", handler.GetCerts)
	})

	return r
}
