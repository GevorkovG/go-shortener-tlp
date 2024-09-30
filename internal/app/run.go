package app

import (
	"errors"
	"flag"
	"net/http"
	"time"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

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

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // flushes buffer, if any

	sugar := *logger.Sugar()

	r := chi.NewRouter()
	r.Post("/", WithLogging(GetShortURL))
	r.Get("/{id}", GetOriginURL)

	flag.Parse()

	sugar.Infow(
		"Starting server",
		"addr", config.AppConfig.Host,
	)

	if err := http.ListenAndServe(config.AppConfig.Host, r); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	}
	return logFn
}
