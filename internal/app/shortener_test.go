package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func Test_GetOriginalURL(t *testing.T) {

	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
		//FilePATH:  "/tmp/short-url-db.json",
	}

	app := NewApp(conf)

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

			app.Storage.SetURL(test.body, resultURL)

			r := httptest.NewRequest(test.method, "http://localhost:8080/"+test.body, nil)

			w := httptest.NewRecorder()

			router := chi.NewRouteContext()

			router.URLParams.Add("id", test.body)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, router))

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
		//FilePATH:  "/tmp/short-url-db.json",
	}

	app := NewApp(conf)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080/api/shorten", strings.NewReader(test.body))

			w := httptest.NewRecorder()

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
		//FilePATH:  "/tmp/short-url-db.json",
	}

	app := NewApp(conf)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "https://localhost:8080", strings.NewReader(test.body))

			w := httptest.NewRecorder()

			app.GetShortURL(w, r)

			assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")
		})
	}
}
