// Пакет usertoken предоставляет функции для работы с JWT токенами пользователей.
// Включает методы для извлечения UserID и валидации токенов.
package usertoken

import (
	"fmt"
	"log"

	"github.com/GevorkovG/go-shortener-tlp/internal/services/jwtstring"
	"github.com/golang-jwt/jwt"
)

// GetUserID извлекает идентификатор пользователя из JWT токена.
//
// Параметры:
//   - tokenString: строка с JWT токеном
//
// Возвращает:
//   - string: идентификатор пользователя
//   - error: ошибка если токен невалидный, содержит:
//   - сообщение о неверном методе подписи
//   - "invalid token" для невалидных токенов
//   - ошибки парсинга токена
//
// Пример использования:
//
//	userID, err := GetUserID("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
func GetUserID(tokenString string) (string, error) {
	claims := &jwtstring.Claims{}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/services/usertoken/cookie.go ValidationToken tokenString %s ", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtstring.SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

// ValidationToken проверяет валидность JWT токена.
//
// Параметры:
//   - tokenString: строка с JWT токеном для проверки
//
// Возвращает:
//   - bool: true если токен валиден, false в противном случае
//
// Особенности:
//   - Проверяет метод подписи (должен быть HMAC)
//   - Проверяет срок действия токена
//   - Логирует процесс проверки (в debug режиме)
func ValidationToken(tokenString string) bool {
	claims := &jwtstring.Claims{}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/services/usertoken/cookie.go ValidationToken tokenString %s ", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtstring.SecretKey), nil
	})
	if err != nil {
		return false
	}

	return token.Valid
}
