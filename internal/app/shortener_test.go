package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/jwtstring"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func Test_GetOriginalURL(t *testing.T) {

	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
		// FilePATH: "/tmp/short-url-db.json",
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := uuid.New().String()

	cookieString, err := jwtstring.BuildJWTString(userID)
	if err != nil {
		t.Log("Didn't create cookie string")
	}

	resultURL := "https://yandex.ru"

	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "test#1-ok",
			method: http.MethodPost,
			body:   "vRFgdzs",
			want: want{
				code:     307,
				location: resultURL,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			link := &objects.Link{
				Short:    test.body,
				Original: resultURL,
			}

			if err := app.Storage.Insert(link); err != nil {
				t.Log(err)
			}

			r := httptest.NewRequest(test.method, "http://localhost:8080/"+test.body, nil)

			ctx := context.WithValue(r.Context(), cookies.ContextUserKey, cookieString)

			w := httptest.NewRecorder()

			router := chi.NewRouteContext()

			router.URLParams.Add("id", test.body)

			r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, router))

			app.GetOriginalURL(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа (307) не совпадает с ожидаемым")
			assert.Equal(t, test.want.location, w.Header().Get("Location"), "Location не совпадает с ожидаемым")
		})
	}
}

func Test_JSONGetShortURL(t *testing.T) {

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "test#1-ok",
			method: http.MethodPost,
			body:   `{"url": "https://yandex.ru"}`,
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:   "test#-fail",
			method: http.MethodPost,
			body:   "sdfqwed",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
		FilePATH:  "/tmp/short-url-db.json",
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := uuid.New().String()

	cookieString, err := jwtstring.BuildJWTString(userID)
	if err != nil {
		t.Log("Didn't create cookie string")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080/api/shorten", strings.NewReader(test.body))
			ctx := context.WithValue(r.Context(), cookies.ContextUserKey, cookieString)

			w := httptest.NewRecorder()

			router := chi.NewRouteContext()

			r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, router))

			app.JSONGetShortURL(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")
		})
	}

}

func Test_GetShortURL(t *testing.T) {

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "test#1-ok",
			method: http.MethodPost,
			body:   "https://yandex.ru",
			want: want{
				code:        201,
				contentType: "text/plain",
			},
		},
		{
			name:   "test#2-fail",
			method: http.MethodPost,
			body:   "",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
		// FilePATH: "/tmp/short-url-db.json",
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := uuid.New().String()

	cookieString, err := jwtstring.BuildJWTString(userID)
	if err != nil {
		t.Log("Didn't create cookie string")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080", strings.NewReader(test.body))

			ctx := context.WithValue(r.Context(), cookies.ContextUserKey, cookieString)

			router := chi.NewRouteContext()

			r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, router))

			w := httptest.NewRecorder()

			app.GetShortURL(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")
		})
	}
}

func TestAPIGetUserURLs(t *testing.T) {
	// Настройка тестового окружения
	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}
	app := NewApp(conf)
	app.ConfigureStorage()

	// Создаем тестового пользователя
	userID := uuid.New().String()
	cookieString, err := jwtstring.BuildJWTString(userID)

	t.Logf("cookieString for UserID: %s", cookieString)

	if err != nil {
		t.Fatal("Failed to create cookie string")
	}

	// Добавляем тестовые данные в хранилище
	link := &objects.Link{
		Short:    "testShort",
		Original: "http://example.com",
		UserID:   userID,
	}
	if err := app.Storage.Insert(link); err != nil {
		t.Fatal("Failed to insert link")
	}
	t.Logf("UserID added: %s", link.UserID)

	// Проверяем, что данные добавлены в хранилище
	userURLs, err := app.Storage.GetAllByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get URLs for user: %v", err)
	}
	if len(userURLs) == 0 {
		t.Fatal("No URLs found in storage for the user")
	}

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), cookies.ContextUserKey, cookieString)
	req = req.WithContext(ctx)

	// Логируем userID из контекста
	userIDFromContext, ok := req.Context().Value(cookies.ContextUserKey).(string)
	if !ok {
		t.Fatal("Failed to get userID from context")
	}
	t.Logf("UserID from context: %s", userIDFromContext)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Вызываем хендлер
	app.APIGetUserURLs(w, req)

	// Проверяем ответ
	res := w.Result()
	defer res.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Логируем тело ответа
	t.Logf("Response body: %s", string(body))

	// Логируем статус-код
	t.Logf("Status code: %d", res.StatusCode)

	// Логируем заголовки ответа
	t.Logf("Response headers: %v", res.Header)

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Декодируем JSON только если статус-код 200
	if res.StatusCode == http.StatusOK {
		var links []RespURLs
		if err := json.Unmarshal(body, &links); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		if len(links) == 0 {
			t.Error("Expected non-empty response")
		}

		// Проверяем содержимое ответа
		expected := RespURLs{
			Short:    "http://localhost:8080/testShort",
			Original: "http://example.com",
		}
		if links[0] != expected {
			t.Errorf("Expected %v, got %v", expected, links[0])
		}
	}
}
