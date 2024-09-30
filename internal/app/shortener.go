package app

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi"
)

func generateID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (a *App) GetShortURL(w http.ResponseWriter, r *http.Request) {
	//urls = make(map[string]string)
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
	a.storage.SetURL(id, url)
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
	url, err := a.storage.GetURL(id)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
