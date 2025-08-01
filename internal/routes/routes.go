package routes

import (
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/logger"
	"github.com/GevorkovG/go-shortener-tlp/internal/middleware"
	"github.com/go-chi/chi"
)

// Router - http роутер
func Router(app *app.App) http.Handler {
	r := chi.NewRouter()

	r.Use(logger.LoggerMiddleware,
		middleware.GzipMiddleware,
		cookies.Cookies,
	)

	r.Post("/api/shorten", app.JSONGetShortURL)
	r.Get("/{id}", app.GetOriginalURL)
	r.Get("/ping", app.Ping)
	r.Post("/", app.GetShortURL)
	r.Post("/api/shorten/batch", app.APIshortBatch)
	r.Get("/api/user/urls", app.APIGetUserURLs)
	r.Delete("/api/user/urls", app.APIDeleteUserURLs)
	r.Get("/api/internal/stats", app.GetStats)

	return r
}
