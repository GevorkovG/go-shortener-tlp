package app

import (
	"encoding/json"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
	"go.uber.org/zap"
)

type RespURLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (a *App) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста с правильным типом ключа
	token, ok := r.Context().Value(cookies.ContextUserKey).(string)
	if !ok {
		zap.L().Warn("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := usertoken.GetUserID(token)
	if err != nil || userID == "" {
		zap.L().Warn("Unauthorized access attempt", zap.String("token", token))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userURLs, err := a.Storage.GetAllByUserID(userID)
	if err != nil {
		zap.L().Error("Failed to get user URLs", zap.String("userID", userID), zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(userURLs) == 0 {
		zap.L().Info("No URLs found for user", zap.String("userID", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		zap.L().Error("Failed to write response", zap.Error(err))
	}
}
