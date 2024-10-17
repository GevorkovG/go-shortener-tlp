package app

import (
	"errors"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

type App struct {
	cfg      *config.AppConfig
	Storage  *storage.InMemoryStorage
	DataBase *database.DBStore
	DBReady  bool
}

func NewApp(cfg *config.AppConfig) *App {
	return &App{
		cfg:     cfg,
		Storage: storage.NewInMemoryStorage(),
	}
}

func (a *App) ConfigureDB() error {
	if a.cfg.DataBaseString != "" {
		db := database.NewDB(a.cfg.DataBaseString)
		if err := db.Open(); err != nil {
			return err
		}
		a.DataBase = db
		a.DBReady = true
		return nil
	}
	return errors.New("dataBaseString is empty")
}
