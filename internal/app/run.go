package app

import (
	"flag"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func Run() {

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // flushes buffer, if any

	sugar := *logger.Sugar()

	r := chi.NewRouter()
	r.Post("/", GetShortURL)
	r.Get("/{id}", GetOriginURL)

	flag.Parse()

	if err := http.ListenAndServe(config.AppConfig.Host, r); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
