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

func InitDB(conn string) *DBStore {
	if conn == "" {
		return nil
	}
	db := NewDB(conn)
	if err := db.Open(); err != nil {
		log.Println("Don't connect DataBase")
		log.Fatal(err)
		return nil
	}
	if err := db.PingDB(); err != nil {
		log.Println("Don't ping DataBase")
		log.Fatal(err)
		return nil
	}
	return db
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
