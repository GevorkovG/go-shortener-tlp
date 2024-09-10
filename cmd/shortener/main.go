package main

import (
	"Praktikum_golang/sprint1/first/cmd/server/go-shortener-tlp/config"
	"Praktikum_golang/sprint1/first/cmd/server/go-shortener-tlp/internal/app"
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	r := chi.NewRouter()
	r.Post("/", app.GetShortURL)
	r.Get("/{id}", app.GetOriginURL)

	flag.Parse()

	log.Fatal(http.ListenAndServe(config.AppConfig.Host, r))
}
