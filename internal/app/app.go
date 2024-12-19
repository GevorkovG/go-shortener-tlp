package app

import (
	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

type App struct {
	cfg      *config.AppConfig
	DataBase *database.DBStore
	Storage  objects.Storage
}

type contextKey string

const Token contextKey = "token"

func NewApp(cfg *config.AppConfig) *App {

	return &App{
		cfg:      cfg,
		DataBase: database.InitDB(cfg.DataBaseString),
	}
}

func (a *App) GetConfig() *config.AppConfig {
	return a.cfg
}

func confDB(conn string) (*database.DBStore, error) {
	db := database.NewDB(conn)
	if err := db.Open(); err != nil {
		return nil, err
	}
	if err := db.PingDB(); err != nil {
		return nil, err
	}
	return db, nil
}

func (a *App) ConfigureStorage() {
	switch {
	case a.cfg.DataBaseString != "":
		a.Storage = storage.NewLinkStorage(a.DataBase)
	case a.cfg.FilePATH != "":
		a.Storage = storage.NewFileStorage(a.cfg.FilePATH)
	default:
		a.Storage = storage.NewInMemoryStorage()
	}
}
