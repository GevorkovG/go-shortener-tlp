package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi"
)

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
	var status = http.StatusCreated

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link := &objects.Link{
		Short:    generateID(),
		Original: req.URL,
	}

	if err = a.Storage.Insert(link); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			link, err = a.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Don't get short URL", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			status = http.StatusConflict
		} else {
			zap.L().Error("Don't insert URL", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	result := Response{
		Result: a.cfg.ResultURL + "/" + link.Short,
	}

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(response)
	if err != nil {
		return
	}

}

func (a *App) GetShortURL(w http.ResponseWriter, r *http.Request) {

	var status = http.StatusCreated

	responseData, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
		return
	}
	if string(responseData) == "" {
		http.Error(w, "Empty POST request body!", http.StatusBadRequest)
		return
	}

	link := &objects.Link{
		Short:    generateID(),
		Original: string(responseData),
	}
	fmt.Println("link1: ", link)

	if err = a.Storage.Insert(link); err != nil {
		fmt.Println("------", err)
		if errors.Is(err, storage.ErrConflict) {
			link, err = a.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Don't get short URL", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			status = http.StatusConflict
		} else {
			zap.L().Error("Don't insert URL", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Println("end.. ")
	response := strings.TrimSpace(fmt.Sprintf(a.cfg.ResultURL+"/%s", link.Short))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	_, err = io.WriteString(w, response)
	if err != nil {
		return
	}
}

func (a *App) GetOriginalURL(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	link, err := a.Storage.GetOriginal(id)

	if err != nil {
		log.Println("Don't read data from table")
		log.Println(err)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", link.Original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
