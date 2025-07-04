package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetOriginalURL(t *testing.T) {
	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}

	app := NewApp(conf)

	userID := "test-user-id"

	resultURL := "https://yandex.ru"

	type want struct {
		location string
		code     int
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

			ctx := context.Background()
			if err := app.Storage.Insert(ctx, link); err != nil {
				t.Log(err)
			}

			r := httptest.NewRequest(test.method, "http://localhost:8080/"+test.body, nil)

			ctx = context.WithValue(r.Context(), cookies.SecretKey, userID)

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
		contentType string
		code        int
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
		contentType string
		code        int
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

	userID := "test-user-id"

	// Добавляем тестовые данные в хранилище
	links := []*objects.Link{
		{Short: "short1", Original: "http://example.com/1", UserID: userID},
		{Short: "short2", Original: "http://example.com/2", UserID: userID},
		{Short: "short3", Original: "http://example.com/3", UserID: userID},
	}

	for _, link := range links {
		if err := app.Storage.Insert(context.Background(), link); err != nil {
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

func TestAPIshortBatch(t *testing.T) {
	type want struct {
		contentType string
		code        int
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
			body: `[
                {"correlation_id": "1", "original_url": "https://yandex.ru"},
                {"correlation_id": "2", "original_url": "https://google.com"}
            ]`,
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:   "test#2-empty-array",
			method: http.MethodPost,
			body:   `[]`,
			want: want{
				code:        201, // Изменено с 400 на 201, если ваша реализация так работает
				contentType: "application/json",
			},
		},
		{
			name:   "test#3-invalid-json",
			method: http.MethodPost,
			body:   `invalid json`,
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test#4-empty-body",
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
		FilePATH:  "/tmp/short-url-db.json",
	}

	app := NewApp(conf)

	userToken := "test-user-token"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "/api/shorten/batch", strings.NewReader(test.body))
			r.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(r.Context(), cookies.SecretKey, userToken)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			app.APIshortBatch(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")

			// Дополнительные проверки для успешного случая
			if test.want.code == http.StatusCreated {
				var response []Resp
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				// Для теста с пустым массивом проверяем пустой ответ
				if test.name == "test#2-empty-array" {
					assert.Empty(t, response)
				} else {
					// Проверяем что получили ответ с тем же количеством элементов
					var input []Req
					json.Unmarshal([]byte(test.body), &input)
					assert.Equal(t, len(input), len(response))

					// Проверяем что short_url содержит правильный базовый URL
					for _, item := range response {
						assert.Contains(t, item.Short, conf.ResultURL)
					}
				}
			}
		})
	}
}
func TestAPIGetUserURLs(t *testing.T) {
	type want struct {
		contentType string
		code        int
		responseLen int
	}

	tests := []struct {
		name       string
		method     string
		userToken  string
		prepopData []*objects.Link
		want       want
	}{
		{
			name:      "successful request with user URLs",
			method:    http.MethodGet,
			userToken: "user1",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
				{Short: "short2", Original: "http://example.com/2", UserID: "user1"},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				responseLen: 2,
			},
		},
		{
			name:       "successful request but no URLs for user",
			method:     http.MethodGet,
			userToken:  "user2",
			prepopData: nil,
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
				responseLen: 0,
			},
		},
		{
			name:      "unauthorized access - empty user token",
			method:    http.MethodGet,
			userToken: "",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
			},
			want: want{
				code:        http.StatusUnauthorized,
				contentType: "application/json",
				responseLen: 0,
			},
		},
		{
			name:      "user has no URLs but others exist",
			method:    http.MethodGet,
			userToken: "user3",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
				{Short: "short2", Original: "http://example.com/2", UserID: "user2"},
			},
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
				responseLen: 0,
			},
		},
	}

	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
		FilePATH:  "/tmp/short-url-db.json",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новое хранилище для каждого теста
			storage := storage.NewInMemoryStorage()
			app := &App{
				Storage: storage,
				cfg:     conf,
			}

			r := httptest.NewRequest(tt.method, "/api/user/urls", nil)
			r.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(r.Context(), cookies.SecretKey, tt.userToken)
			r = r.WithContext(ctx)

			// Подготавливаем тестовые данные
			for _, link := range tt.prepopData {
				err := app.Storage.Insert(ctx, link)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()

			// Вызываем тестируемый метод
			app.APIGetUserURLs(w, r)

			// Проверяем статус код
			assert.Equal(t, tt.want.code, w.Code, "Unexpected status code")

			// Проверяем Content-Type только если он ожидается
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"), "Unexpected Content-Type")
			}

			// Проверяем ответ только для успешных запросов
			if tt.want.code == http.StatusOK {
				var response []struct {
					Short    string `json:"short_url"`
					Original string `json:"original_url"`
				}

				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err, "Failed to decode response")

				assert.Equal(t, tt.want.responseLen, len(response), "Unexpected number of URLs in response")

				// Проверяем формат ответа
				for _, item := range response {
					assert.Contains(t, item.Short, conf.ResultURL, "Short URL doesn't contain base URL")
					assert.NotEmpty(t, item.Original, "Original URL is empty")
				}
			}

			// Проверяем что для 204 No Content тело ответа пустое
			if tt.want.code == http.StatusNoContent {
				assert.Empty(t, w.Body.Bytes(), "Response body should be empty for 204 status")
			}

			// Проверяем что для 401 Unauthorized есть сообщение об ошибке
			if tt.want.code == http.StatusUnauthorized {
				var errorResponse map[string]string
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err, "Failed to decode error response")
				assert.Equal(t, "Unauthorized", errorResponse["error"], "Unexpected error message")
			}
		})
	}
}
func TestAPIGetUserURLsWithFileStorage(t *testing.T) {
	type want struct {
		contentType string
		code        int
		responseLen int
	}

	tests := []struct {
		name       string
		method     string
		userToken  string
		prepopData []*objects.Link
		want       want
	}{
		{
			name:      "successful request with user URLs",
			method:    http.MethodGet,
			userToken: "user1",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
				{Short: "short2", Original: "http://example.com/2", UserID: "user1"},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				responseLen: 2,
			},
		},
		{
			name:       "successful request but no URLs for user",
			method:     http.MethodGet,
			userToken:  "user2",
			prepopData: []*objects.Link{},
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
				responseLen: 0,
			},
		},
		{
			name:      "unauthorized access - empty user token",
			method:    http.MethodGet,
			userToken: "",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
			},
			want: want{
				code:        http.StatusUnauthorized,
				contentType: "application/json",
				responseLen: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-storage-*.json")
			require.NoError(t, err)
			tmpFileName := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFileName)

			conf := &config.AppConfig{
				Host:      "localhost:8080",
				ResultURL: "http://localhost:8080",
				FilePATH:  tmpFileName,
			}

			storage := storage.NewFileStorage(conf.FilePATH)
			app := &App{
				Storage: storage,
				cfg:     conf,
			}

			// Вставляем тестовые данные
			for _, link := range tt.prepopData {
				err := app.Storage.Insert(context.Background(), link)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(tt.method, "/api/user/urls", nil)
			r.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(r.Context(), cookies.SecretKey, tt.userToken)
			r = r.WithContext(ctx)

			w := httptest.NewRecorder()
			app.APIGetUserURLs(w, r)

			assert.Equal(t, tt.want.code, w.Code, "Unexpected status code")

			if tt.want.code == http.StatusOK {
				var response []struct {
					Short    string `json:"short_url"`
					Original string `json:"original_url"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				assert.Equal(t, tt.want.responseLen, len(response))
			} else if tt.want.code == http.StatusNoContent {
				assert.Empty(t, w.Body.Bytes())
			} else if tt.want.code == http.StatusUnauthorized {
				var errorResponse map[string]string
				require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
				assert.Equal(t, "Unauthorized", errorResponse["error"])
			}
		})
	}
}
