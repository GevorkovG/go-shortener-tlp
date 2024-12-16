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

const Token_Exp = time.Hour * 3
const Secret_Key = "sHoRtEnEr"

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(uuid string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(Token_Exp)),
		},
		// собственное утверждение
		UserID: uuid,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(Secret_Key))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}
