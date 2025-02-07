package cookies

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	Token_Exp             = time.Hour * 3
	Secret_Key contextKey = "supersecretkey"
)

func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(Token_Exp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(Secret_Key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(Secret_Key), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	return claims.UserID, nil
}

func Cookies(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie("token")
		if err == nil && cookie != nil {
			id, err := GetUserID(cookie.Value)
			if err == nil {
				userID = id
				//log.Printf("UserID: %s, Token: %s", userID, cookie.Value)
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			tokenString, err := BuildJWTString(userID)
			if err != nil {
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

		ctx := context.WithValue(r.Context(), Secret_Key, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

/*

//MyNEW----------------------------------
package cookies

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Определяем собственный тип для ключа контекста
type contextKey string

const ContextUserKey contextKey = "supersecretkey"

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(uuid string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: uuid,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// Получение из строки токена UserID
func GetUserID(tokenString string) string {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	// возвращаем ID пользователя в читаемом виде
	log.Printf("internal/cookies/cookies.go GetUserID %s\n", claims.UserID)
	return claims.UserID
}

// Проверка валидности токена
func GetUserId(tokenString string) string {
	claims := &Claims{}
	//token1, _ := jwt.ParseWithClaims()
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil {

		// DEBUG--------------------------------------------------------------------------------------------------
		log.Printf("internal/cookies/cookies.go err: %t", err)

		return "-1"
	}

	if !token.Valid {
		fmt.Println("internal/cookies/cookies.go Token is not valid")
		return "-1"
	}

	fmt.Printf("internal/cookies/cookies.go  Token is valid %s\n", claims.UserID)
	return claims.UserID
}

func Cookies(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var UserID string

		// DEBUG--------------------------------------------------------------------------------------------------
		//log.Printf("internal/cookies/cookies.go MAIN1 start")

		cookie, _ := r.Cookie("token")

		// DEBUG--------------------------------------------------------------------------------------------------
		//log.Printf("internal/cookies/cookies.go cookie==nil: ---%t--- \n", cookie == nil)

		if cookie != nil {
			if GetUserID(cookie.Value) != "-1" {
				UserID = GetUserID(cookie.Value)

				// DEBUG--------------------------------------------------------------------------------------------------
				log.Printf("internal/cookies/cookies.go cookie \n UserID %s\n Token %s", UserID, cookie.Value)
			}
		} else {
			//newToken
			UserID := uuid.New().String()
			tokenString, err := BuildJWTString(UserID)
			if err != nil {
				//zap.L().Error("Failed to build JWT string", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:  "token",
				Value: tokenString,
				Path:  "/",
			})

			// DEBUG--------------------------------------------------------------------------------------------------
			log.Printf("internal/cookies/cookies.go cookie=nil \n UserID %s\n Token %s", UserID, tokenString)

		}
		ctx := context.WithValue(r.Context(), ContextUserKey, UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
*/
/*
		//newToken
		userAIDI := uuid.New().String()

		tokenString, err := BuildJWTString(userAIDI)
		if err != nil {
			log.Fatal(err)
		}

		GetUserID(tokenString)
		fmt.Println("NEWTOKEN", tokenString)

		UserID := GetUserID(tokenString)

		// DEBUG--------------------------------------------------------------------------------------------------
		log.Printf("internal/cookies/cookies.go ++++++++UserID: %s\n", UserID)

		//fmt.Println(GetUserID(tokenString))

		//fmt.Println(GetUserID(tokenString))
		//log.Printf("internal/cookies/cookies.go MAIN1 middle")
		//fmt.Println(GetUserID(tokenString))

		// DEBUG--------------------------------------------------------------------------------------------------
		//log.Printf("internal/cookies/cookies.go MAIN1 finish")

		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: tokenString,
			Path:  "/",
		})

		ctx := context.WithValue(r.Context(), ContextUserKey, UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Cookies(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		var UserID string

		if cookie == nil || err != nil || !usertoken.ValidationToken(cookie.Value) {
			// gen new cookie
			// DEBUG--------------------------------------------------------------------------------------------------
			log.Printf("internal/cookies/cookies.go GENERATE NEW1 Token %t", (cookie == nil))
			UserID = uuid.New().String()
			token, err := jwtstring.BuildJWTString(UserID)
			if err != nil {
				zap.L().Error("Failed to build JWT string", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// DEBUG--------------------------------------------------------------------------------------------------
			log.Printf("internal/cookies/cookies.go GENERATE NEW2 UserID %s Token %s", UserID, token)

			// cookie = new cookie
		} else {
			UserID, _ = usertoken.GetUserID(cookie.Value)
			//DEBUG--------------------------------------------------------------------------------------------------
			log.Printf("internal/cookies/cookies.go ValidationToken UserID %s  cookie.Value %s", UserID, cookie.Value)

		}
		/*
			UserID, err = usertoken.GetUserID(cookie.Value)
			if err != nil {
				zap.L().Error("Failed to get user ID from token", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Если токен отсутствует или невалиден, создаем новый
			if UserID == "" {

				http.SetCookie(w, &http.Cookie{
					Name:  "token",
					Value: token,
					Path:  "/",
				})
				zap.L().Info("New Token created", zap.String("UserID", UserID), zap.String("token", token))
			}

			UserID, err = usertoken.GetUserID(cookie.Value)
			if err != nil {
				zap.L().Error("Failed to get user ID from token", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		//tuuuutututututu
		// Добавляем UserID в контекст запроса
		ctx := context.WithValue(r.Context(), ContextUserKey, UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
*/
