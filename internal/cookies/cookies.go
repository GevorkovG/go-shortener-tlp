package cookies

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	TokenExp             = time.Hour * 3
	SecretKey contextKey = "supersecretkey"
)

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

			//log.Printf("New UserID: %s, Token: %s", userID, tokenString)
		}

		ctx := context.WithValue(r.Context(), SecretKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
