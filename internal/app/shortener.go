package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

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

func generateID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type Request1 struct {
	URL string `json:"url"`
}

type Response1 struct {
	Result string `json:"result"`
}

func (a *App) JSONGetShortURL(w http.ResponseWriter, r *http.Request) {

	var req Request1

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := generateID()

	a.Storage.SetURL(id, req.URL)
	fileStorage := storage.NewFileStorage()

	fileStorage.Short = id
	fileStorage.Original = req.URL

	err = storage.SaveToFile(fileStorage, a.cfg.FilePATH)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := Response1{
		Result: a.cfg.ResultURL + "/" + id,
	}

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	url := string(responseData)
	if url == "" {
		http.Error(w, "Empty POST request body!", http.StatusBadRequest)
		return
	}

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

	fileStorage := storage.NewFileStorage()

	fileStorage.Short = id
	fileStorage.Original = url

	err = storage.SaveToFile(fileStorage, a.cfg.FilePATH)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf(a.cfg.ResultURL+"/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = io.WriteString(w, response)
	if err != nil {
		return
	}
}

func (a *App) GetOriginURL(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	url, err := a.Storage.GetURL(id)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
