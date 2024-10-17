package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"

	"github.com/go-chi/chi"
)

type Link struct {
	ID       int
	Short    string
	Original string
	Store    *database.DBStore
}

func (l *Link) CreateTable() error {
	if _, err := l.Store.DB.Exec("CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY , short CHAR (20), original CHAR (255));"); err != nil {
		return err
	}
	return nil
}

func (l *Link) Insert(link *Link) (*Link, error) {
	if err := l.CreateTable(); err != nil {
		return nil, err
	}
	if err := l.Store.DB.QueryRow(
		"INSERT INTO links (short, original) VALUES ($1,$2) RETURNING id",
		link.Short, link.Original,
	).Scan(&link.ID); err != nil {
		return nil, err
	}
	return link, nil
}

func (l *Link) GetOriginal(short string) (*Link, error) {

	if err := l.Store.DB.QueryRow("SELECT original FROM links WHERE short = $1", short).Scan(&l.Original); err != nil {
		log.Println(err)
		return nil, err
	}
	return l, nil
}

func generateID() string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	number := rand.Uint64()
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func (a *App) JSONGetShortURL(w http.ResponseWriter, r *http.Request) {

	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := generateID()

	linkModel := Link{
		Short:    id,
		Original: req.URL,
		Store:    a.DataBase,
	}

	if a.DBReady {
		_, err = linkModel.Insert(&linkModel)
		if err != nil {
			log.Println("Don't insert url!")
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if a.cfg.FilePATH != "" {

		fileStorage := storage.NewFileStorage()

		fileStorage.Short = id
		fileStorage.Original = req.URL

		err = storage.SaveToFile(fileStorage, a.cfg.FilePATH)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	a.Storage.SetURL(id, req.URL)

	result := Response{
		Result: a.cfg.ResultURL + "/" + id,
	}

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(response)
	if err != nil {
		return
	}

}

func (a *App) GetShortURL(w http.ResponseWriter, r *http.Request) {

	responseData, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
		return
	}
	if string(responseData) == "" {
		http.Error(w, "Empty POST request body!", http.StatusBadRequest)
		return
	}
	url := string(responseData)
	fmt.Println(url)
	id := generateID()

	linkModel := Link{
		Short:    id,
		Original: url,
		Store:    a.DataBase,
	}

	if a.DBReady {
		_, err = linkModel.Insert(&linkModel)
		if err != nil {
			log.Println("Don't insert url!")
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if a.cfg.FilePATH != "" {
		fileStorage := storage.NewFileStorage()

		fileStorage.Short = id
		fileStorage.Original = url

		err = storage.SaveToFile(fileStorage, a.cfg.FilePATH)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	a.Storage.SetURL(id, url)

	response := fmt.Sprintf(a.cfg.ResultURL+"/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = io.WriteString(w, response)
	if err != nil {
		return
	}
}

func (a *App) GetOriginalURL(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	linkModel := &Link{
		Store: a.DataBase,
	}
	var err error
	if a.DBReady {
		linkModel, err = linkModel.GetOriginal(id)
		if err != nil {
			log.Println("Didn't read data from table")
			log.Println(err)
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		}
	}
	if linkModel.Original == "" {
		linkModel.Original, err = a.Storage.GetURL(id)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
		}
	}

	w.Header().Set("Location", linkModel.Original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
