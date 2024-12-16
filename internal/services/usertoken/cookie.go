package usertoken

import (
	"fmt"
	"log"

	"github.com/GevorkovG/go-shortener-tlp/internal/services/jwtstring"
	"github.com/golang-jwt/jwt"
)

func GetUserID(tokenString string) (string, error) {
	claims := &jwtstring.Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtstring.SECRET_KEY), nil
		})
	if err != nil {
		log.Println(err)
		return "", err
	}

	return claims.UserID, nil
}

func ValidationToken(tokenString string) bool {

	claims := &jwtstring.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtstring.SECRET_KEY), nil
		})
	if err != nil {
		return false
	}

	if !token.Valid {
		return false
	}

	return true
}
