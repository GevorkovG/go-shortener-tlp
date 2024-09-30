package app

import (
	"flag"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"
	logg "github.com/GevorkovG/go-shortener-tlp/internal/log"
	"github.com/go-chi/chi"
)

type App struct {
	cfg     *config.AppConfig
	storage *config.Storage
}

func NewApp(cfg *config.AppConfig) *App {

	return &App{
		cfg:     cfg,
		storage: config.NewStorage(),
	}
}

func Run() {
	conf := config.NewCfg()

	r := chi.NewRouter()
	r.Post("/", logg.WithLogging(conf.GetShortURL))
	r.Get("/{id}", logg.WithLogging(conf.GetOriginURL))
	flag.Parse()

	log.Fatal(http.ListenAndServe(conf.Host, r))

}
