package app

import (
	"log"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

type App struct {
	cfg     *config.AppConfig
	Storage objects.Storage
}

type contextKey string

const Token contextKey = "token"

func NewApp(cfg *config.AppConfig) *App {
	var store objects.Storage

	switch {
	case cfg.DataBaseString != "":
		log.Printf("internal/app/app.go ValidationToken USE DataBase")
		db := database.InitDB(cfg.DataBaseString)
		store = storage.NewLinkStorage(db)
	case cfg.FilePATH != "":
		log.Printf("internal/app/app.go ValidationToken USE FilePATH ")
		store = storage.NewFileStorage(cfg.FilePATH)
	default:
		log.Printf("internal/app/app.go ValidationToken USE NewInMemoryStorage ")
		store = storage.NewInMemoryStorage()
	}

	return &App{
		cfg:     cfg,
		Storage: store,
	}
}

func (a *App) GetConfig() *config.AppConfig {
	return a.cfg
}
