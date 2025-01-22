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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized) // Устанавиваем статус-код
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}
	zap.L().Info("UserID extracted from context", zap.String("userID", userID))

	// Получаем URL-адреса пользователя
	userURLs, err := a.Storage.GetAllByUserID(userID)
	if err != nil {
		zap.L().Error("Failed to get user URLs", zap.String("userID", userID), zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Устанавиваем статус-код
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		return
	}
	zap.L().Info("User URLs retrieved from storage", zap.String("userID", userID), zap.Any("userURLs", userURLs))

	if len(userURLs) == 0 {
		zap.L().Info("No URLs found for user", zap.String("userID", userID))
		w.WriteHeader(http.StatusNoContent) // Возвращаем 204, если данных нет
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		zap.L().Error("Failed to write response", zap.Error(err))
	}
}
