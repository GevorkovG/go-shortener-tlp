// Package jwtstring предоставляет функционал для работы с JWT-токенами.
// Включает методы для генерации, подписи и верификации токенов.
package jwtstring

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// TokenExp определяет срок действия JWT токена.
// Текущее значение - 3 часа.
const TokenExp = time.Hour * 3

// SecretKey содержит секретный ключ для подписи JWT токенов.
// ВНИМАНИЕ: В production среде должен заменяться на значение из защищенного
// хранилища (environment variables, secret manager и т.п.)
const SecretKey = "sHoRtEnEr"

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(uuid string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: uuid,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}
