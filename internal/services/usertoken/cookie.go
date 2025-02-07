package usertoken

import (
	"fmt"
	"log"

	"github.com/GevorkovG/go-shortener-tlp/internal/services/jwtstring"
	"github.com/golang-jwt/jwt"
)

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
