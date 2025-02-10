package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func Test_GetOriginalURL(t *testing.T) {
	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := "test-user-id"

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
				UserID:   userID,
			}

			if err := app.Storage.Insert(link); err != nil {
				t.Log(err)
			}

			r := httptest.NewRequest(test.method, "http://localhost:8080/"+test.body, nil)

			ctx := context.WithValue(r.Context(), cookies.SecretKey, userID)

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

	userID := "test-user-id"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080/api/shorten", strings.NewReader(test.body))
			ctx := context.WithValue(r.Context(), cookies.SecretKey, userID)

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
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := "test-user-id"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080", strings.NewReader(test.body))

			ctx := context.WithValue(r.Context(), cookies.SecretKey, userID)

			router := chi.NewRouteContext()

			r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, router))

			w := httptest.NewRecorder()

			app.GetShortURL(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")
		})
	}
}

func TestAPIDeleteUserURLs(t *testing.T) {
	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}

	app := NewApp(conf)
	app.ConfigureStorage()

	userID := "test-user-id"

	// Добавляем тестовые данные в хранилище
	links := []*objects.Link{
		{Short: "short1", Original: "http://example.com/1", UserID: userID},
		{Short: "short2", Original: "http://example.com/2", UserID: userID},
		{Short: "short3", Original: "http://example.com/3", UserID: userID},
	}

	for _, link := range links {
		if err := app.Storage.Insert(link); err != nil {
			t.Fatal("Failed to insert link")
		}
	}

	// Тест на успешное удаление URL
	t.Run("successful deletion", func(t *testing.T) {
		shortURLs := []string{"short1", "short2"}

		body, err := json.Marshal(shortURLs)
		if err != nil {
			t.Fatal("Failed to marshal shortURLs")
		}

		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
		ctx := context.WithValue(r.Context(), cookies.SecretKey, userID)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()

		app.APIDeleteUserURLs(w, r)

		assert.Equal(t, http.StatusAccepted, w.Code, "Код ответа не совпадает с ожидаемым")

		// Проверяем, что URL действительно удалены
		for _, short := range shortURLs {
			link, err := app.Storage.GetOriginal(short)
			assert.NoError(t, err, "Ошибка при получении URL")
			assert.True(t, storage.IsDeleted(link), "URL не был удален")
		}
	})

	// Тест на попытку удаления URL другого пользователя
	t.Run("unauthorized deletion", func(t *testing.T) {
		shortURLs := []string{"short3"}

		body, err := json.Marshal(shortURLs)
		if err != nil {
			t.Fatal("Failed to marshal shortURLs")
		}

		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
		ctx := context.WithValue(r.Context(), cookies.SecretKey, "another-user-id")
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()

		app.APIDeleteUserURLs(w, r)

		assert.Equal(t, http.StatusAccepted, w.Code, "Код ответа не совпадает с ожидаемым")

		// Проверяем, что URL не был удален
		link, err := app.Storage.GetOriginal("short3")
		assert.NoError(t, err, "Ошибка при получении URL")
		assert.False(t, storage.IsDeleted(link), "URL был удален, хотя не должен был")
	})

	// Тест на попытку удаления несуществующего URL
	t.Run("delete non-existent URL", func(t *testing.T) {
		shortURLs := []string{"non-existent"}

		body, err := json.Marshal(shortURLs)
		if err != nil {
			t.Fatal("Failed to marshal shortURLs")
		}

		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
		ctx := context.WithValue(r.Context(), cookies.SecretKey, userID)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()

		app.APIDeleteUserURLs(w, r)

		assert.Equal(t, http.StatusAccepted, w.Code, "Код ответа не совпадает с ожидаемым")
	})
}
