// Пакет cookies предоставляет функционал для работы с JWT-аутентификацией
// через HTTP cookies.
package cookies

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// contextKey тип для ключей контекста (обертывает string)
type contextKey string

// Claims представляет кастомные JWT-claims с идентификатором пользователя.
// Наследует стандартные зарегистрированные claims JWT
type Claims struct {
	jwt.RegisteredClaims
	UserID string // Уникальный идентификатор пользователя
}

// Константы
const (
	// TokenExp - время жизни JWT токена (3 часа)
	TokenExp = time.Hour * 3

	// SecretKey ключ для доступа к данным в контексте
	SecretKey contextKey = "supersecretkey"
)

// BuildJWTString создает JWT токен для указанного пользователя.
//
// Параметры:
//   - userID: строка с идентификатором пользователя
//
// Возвращает:
//   - string: подписанный JWT токен
//   - error: ошибка если не удалось подписать токен
//
// Пример использования:
//
//	token, err := BuildJWTString("123e4567-e89b-12d3-a456-426614174000")
func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID извлекает идентификатор пользователя из JWT токена.
//
// Параметры:
//   - tokenString: строка с JWT токеном
//
// Возвращает:
//   - string: идентификатор пользователя
//   - error: ошибка если токен невалидный или не содержит UserID
//
// Пример использования:
//
//	userID, err := GetUserID("eyJhbGciOiJIUzI1NiIsI...")
func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	return claims.UserID, nil
}

// Cookies middleware для обработки аутентификации через JWT в cookies.
//
// Функционал:
//   - Проверяет наличие валидного токена в cookie "token"
//   - Если токен валиден - извлекает UserID
//   - Если токена нет/невалиден - генерирует новый UserID и токен
//   - Добавляет UserID в контекст запроса
//   - Устанавливает cookie с токеном для новых пользователей
//
// Возможные ошибки:
//   - 500 Internal Server Error при ошибке генерации токена
//
// Пример использования:
//
//	router.Use(Cookies)
func Cookies(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, _ := r.Cookie("token")
		if cookie != nil {
			id, err := GetUserID(cookie.Value)
			if err == nil {
				userID = id
			} else {
				zap.L().Info("Failed to create UserID", zap.String("cookie.value", cookie.Value))
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			tokenString, err := BuildJWTString(userID)
			if err != nil {
				zap.L().Error("Failed to create a new token", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  "token",
				Value: tokenString,
				Path:  "/",
			})
		}

		ctx := context.WithValue(r.Context(), SecretKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
