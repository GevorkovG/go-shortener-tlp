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

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
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
	var UserID string

	// Извлекаем UserID из контекста
	token := r.Context().Value(cookies.SecretKey)
	if token != nil {
		var err error
		UserID, err = usertoken.GetUserID(token.(string))
		if err != nil {
			zap.L().Warn("Failed to get UserID from token, proceeding without it", zap.Error(err))
			UserID = ""
		}
	}

	// Декодируем тело запроса
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем объект Link
	link := &objects.Link{
		Short:    generateID(),
		Original: req.URL,
		UserID:   UserID, // Устанавливаем UserID
	}

	// Сохраняем ссылку в хранилище
	if err = a.Storage.Insert(link); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			// Если URL уже существует, получаем существующий короткий URL
			link, err = a.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Failed to get short URL", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			status = http.StatusConflict
		} else {
			zap.L().Error("Failed to insert URL", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Формируем ответ
	result := Response{
		Result: strings.TrimSpace(fmt.Sprintf("%s/%s", a.cfg.ResultURL, link.Short)),
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
		zap.L().Error("Failed to write response", zap.Error(err))
		return
	}
}

func (a *App) GetShortURL(w http.ResponseWriter, r *http.Request) {
	var status = http.StatusCreated

	// Извлекаем UserID из контекста
	userID, ok := r.Context().Value(cookies.SecretKey).(string)
	if !ok || userID == "" {
		// Если UserID не найден в контексте, логируем ошибку и продолжаем без него
		zap.L().Error("UserID not found in context")
		userID = "" // Устанавливаем пустой UserID
	}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/app/shortener.go  UserID: %s ", userID)

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
		UserID:   userID, // Устанавливаем UserID
	}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/app/shortener.go GetShortURL Original:%s UserID:%s Short: %s", link.Original, link.UserID, link.Short)

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

	response := strings.TrimSpace(fmt.Sprintf("%s/%s", a.cfg.ResultURL, link.Short))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	_, err = io.WriteString(w, response)
	if err != nil {
		zap.L().Error("Didn't write response", zap.Error(err))
		return
	}
}

func (a *App) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	link, err := a.Storage.GetOriginal(id)
	log.Printf("GetOriginalURL short:%s %t", link.Short, link.DeletedFlag)
	if err != nil {
		zap.L().Error("Failed to get original URL", zap.String("id", id), zap.Error(err))
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Проверяем, удален ли URL
	if link.DeletedFlag || link.Original == "" {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", link.Original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) Ping(w http.ResponseWriter, _ *http.Request) {
	if err := a.Storage.Ping(); err != nil {
		log.Println("Storage ping failed:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
