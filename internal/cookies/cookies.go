package cookies

import (
	"context"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/services/jwtstring"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Определяем собственный тип для ключа контекста
type contextKey string

const ContextUserKey contextKey = "userID"

func Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		var userID string

		if err == nil {
			// Проверяем валидность токена
			if usertoken.ValidationToken(cookie.Value) {
				userID, err = usertoken.GetUserID(cookie.Value)
				if err != nil {
					zap.L().Error("Failed to get user ID from token", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}
		}

		// Если токен отсутствует или невалиден, создаем новый
		if userID == "" {
			userID = uuid.New().String()
			token, err := jwtstring.BuildJWTString(userID)
			if err != nil {
				zap.L().Error("Failed to build JWT string", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  "token",
				Value: token,
				Path:  "/",
			})
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), ContextUserKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
