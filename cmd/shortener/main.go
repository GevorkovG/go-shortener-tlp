package main

import (
	"Praktikum_golang/sprint1/first/cmd/server/go-shortener-tlp/config"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var urls map[string]string

func generateID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetShortURL(w http.ResponseWriter, r *http.Request) {
	urls = make(map[string]string)
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
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := generateID()
	urls[id] = url
	response := fmt.Sprintf("http://localhost:8080/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(response))
}

func GetOriginURL(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, ok := urls[id]
	if !ok {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {

	urls = make(map[string]string)

	r := chi.NewRouter()
	r.Post("/", GetShortURL)
	r.Get("/{id}", GetOriginURL)

	flag.Parse()

	err := http.ListenAndServe(config.AppConfig.Host, r)
	if err != nil {
		panic(err)
	}
}
