package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

var urls map[string]string

func shortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
			return
		}
		url := string(body)
		id := generateID()
		urls[id] = url
		response := fmt.Sprintf("http://localhost:8080/%s", id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(response))

	} else if r.Method == http.MethodGet {
		id := r.URL.Path[1:]
		url := urls[id]
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func generateID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	urls = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
