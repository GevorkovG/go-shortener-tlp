package app

import (
	"log"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/dbmodel"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

type App struct {
	cfg      *config.AppConfig
	DataBase *database.DBStore
	Storage  objects.Storage
}

func NewApp(cfg *config.AppConfig) *App {

	return &App{
		cfg:     cfg,
		Storage: ConfigureStorage(cfg),
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

func ConfigureStorage(conf *config.AppConfig) objects.Storage {
	if conf.DataBaseString != "" {
		db, err := confDB(conf.DataBaseString)
		if err == nil {
			return &dbmodel.Link{
				Store: db,
			}
		} else {
			log.Fatal(err)
			return nil
		}

	} else if conf.FilePATH != "" {
		store := storage.NewFileStorage(conf.FilePATH)

		data, err := storage.LoadFromFile(conf.FilePATH)

		store.Load(data)

		if err != nil {
			log.Println("Don't load from file!")
			log.Fatal(err)
			return nil
		}

		store.Load(data)

		return store
	}
	return storage.NewInMemoryStorage()
}
