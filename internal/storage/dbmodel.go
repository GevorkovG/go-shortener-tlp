package storage

import (
	"errors"
	"fmt"
	"log"

	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

var ErrConflict = errors.New("conflict on inserting new record")

type Link struct {
	Store *database.DBStore
}

func NewLinkStorage(db *database.DBStore) *Link {
	return &Link{
		Store: db,
	}
}

func (l *Link) CreateTable() error {
	if _, err := l.Store.DB.Exec("CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY , short CHAR (20), original CHAR (255));"); err != nil {
		return err
	}
	return nil
}

func (l *Link) Insert(link *objects.Link) error {
	if err := l.CreateTable(); err != nil {
		return err
	}
	fmt.Println("here")
	if _, err := l.Store.DB.Exec(
		"INSERT INTO links (short, original) VALUES ($1,$2)",
		link.Short, link.Original); err != nil {
		var pgErr *pgconn.PgError
		fmt.Println("*", err)
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				err = ErrConflict
			}
		}
		return err
	}
	return nil
}

func (l *Link) InsertLinks(links []*objects.Link) error {
	if err := l.CreateTable(); err != nil {
		return err
	}
	tx, err := l.Store.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO links (short, original) VALUES($1,$2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range links {
		_, err := stmt.Exec(v.Short, v.Original)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (l *Link) GetOriginal(short string) (*objects.Link, error) {

	link := &objects.Link{
		Short: short,
	}
	if err := l.Store.DB.QueryRow("SELECT original FROM links WHERE short = $1", link.Short).Scan(&link.Original); err != nil {
		log.Println(err)
		return link, err
	}
	return link, nil
}

func (l *Link) GetShort(original string) (*objects.Link, error) {

	link := objects.Link{
		Original: original,
	}
	if err := l.Store.DB.QueryRow("SELECT short FROM links WHERE original = $1", link.Original).Scan(&link.Short); err != nil {
		zap.L().Error("Don't get short URL", zap.Error(err))
		return &link, err
	}
	return &link, nil
}
