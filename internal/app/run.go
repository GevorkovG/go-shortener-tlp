package app

import (
	"flag"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"

	"github.com/go-chi/chi"
)

func Run() {
	r := chi.NewRouter()
	r.Post("/", GetShortURL)
	r.Get("/{id}", GetOriginURL)

	flag.Parse()

	log.Fatal(http.ListenAndServe(config.AppConfig.Host, r))
}
