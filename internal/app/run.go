package app

import (
	"flag"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"
	logg "github.com/GevorkovG/go-shortener-tlp/internal/log"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"github.com/go-chi/chi"
)

type App struct {
	cfg     *config.AppConfig
	storage *storage.Storage
}

func NewApp(cfg *config.AppConfig) *App {

	return &App{
		cfg:     cfg,
		storage: storage.NewStorage(),
	}
}

func Run() {
	var cfg *config.AppConfig
	conf := config.NewCfg()
	newApp := NewApp(cfg)
	r := chi.NewRouter()
	r.Post("/", logg.WithLogging(newApp.GetShortURL))
	r.Get("/{id}", logg.WithLogging(newApp.GetOriginURL))
	flag.Parse()
	log.Fatal(http.ListenAndServe(conf.Host, r))

}
