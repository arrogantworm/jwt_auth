package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) RegisterRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	tokenMaker := h.TokenMaker

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.createUser)
		r.Post("/signin", h.loginUser)
		r.With(GetAuthMiddlewareFunc(tokenMaker)).Post("/logout", h.logoutUser)

		r.Route("/tokens", func(r chi.Router) {
			r.Use(GetAuthMiddlewareFunc(tokenMaker))
			r.Post("/renew", h.renewAccessToken)
			r.Post("/revoke", h.revokeSession)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/test", h.testHandler)
		r.With(GetAuthMiddlewareFunc(tokenMaker)).Get("/user", h.getUserInfo)
	})

	r.Route("/new-ip", func(r chi.Router) {
		r.Post("/", h.newIpReciever)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8000/swagger/doc.json"),
	))

	return r
}
