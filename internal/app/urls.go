package app

import (
	"encoding/json"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"go.uber.org/zap"
)

type RespURLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (a *App) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := r.Context().Value(cookies.ContextUserKey).(string)
	if !ok || userID == "" {
		zap.L().Warn("Unauthorized access attempt")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	zap.L().Info("UserID extracted from context", zap.String("userID", userID))

	// Получаем URL-адреса пользователя
	userURLs, err := a.Storage.GetAllByUserID(userID)
	if err != nil {
		zap.L().Error("Failed to get user URLs", zap.String("userID", userID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	zap.L().Info("User URLs retrieved from storage", zap.String("userID", userID), zap.Any("userURLs", userURLs))

	if len(userURLs) == 0 {
		zap.L().Info("No URLs found for user", zap.String("userID", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Формируем ответ
	var links []RespURLs
	for _, val := range userURLs {
		links = append(links, RespURLs{
			Short:    a.cfg.ResultURL + "/" + val.Short,
			Original: val.Original,
		})
	}

	response, err := json.Marshal(links)
	if err != nil {
		zap.L().Error("Failed to marshal user URLs", zap.String("userID", userID), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		zap.L().Error("Failed to write response", zap.Error(err))
	}
}
