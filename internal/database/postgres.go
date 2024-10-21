package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStore struct {
	DatabaseConf string
	DB           *sql.DB
}

func NewDB(conf string) *DBStore {
	return &DBStore{
		DatabaseConf: conf,
	}
}

func (store *DBStore) Open() error {

	db, err := sql.Open("pgx", store.DatabaseConf)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	store.DB = db
	return nil
}

func (store *DBStore) Close() {
	store.DB.Close()
}

func (store *DBStore) PingDB() error {
	if err := store.DB.Ping(); err != nil {
		log.Println("don't ping Database")
		return err
	}
	return nil
}
