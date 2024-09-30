package app

import (
	"errors"
	"flag"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/log"
	"github.com/go-chi/chi"
)

type Storage struct {
	urls map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		urls: make(map[string]string),
	}
}

func (s *Storage) SetURL(key, value string) {
	s.urls[key] = value
}

func (s *Storage) GetURL(key string) (string, error) {

	url, ok := s.urls[key]
	if ok {
		return url, nil
	}
	return "", errors.New("id not found")
}

type AppConfig struct {
	Host      string
	ResultURL string
}
type App struct {
	cfg     *AppConfig
	storage *Storage
}

func NewApp(cfg *AppConfig) *App {

	return &App{
		cfg:     cfg,
		storage: NewStorage(),
	}
}

func Run() {
	conf := config.NewCfg()

	r := chi.NewRouter()
	r.Post("/", log.WithLogging(conf.GetShortURL))
	r.Get("/{id}", log.WithLogging(conf.GetOriginURL))
	flag.Parse()

	log.Fatal(http.ListenAndServe(conf.Host, r))

}
