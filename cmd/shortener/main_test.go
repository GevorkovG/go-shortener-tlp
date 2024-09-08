package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testUrls map[string]string

func Test_shortenURL(t *testing.T) {
	urls = make(map[string]string)

	testUrls = make(map[string]string)
	type want struct {
		contentType string
		statusCode  int
	}

	testURL := "https://yandex.ru"

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "test 1 (yandex)",
			url:  testURL,
			want: want{
				statusCode:  201,
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "https://localhost:8080", strings.NewReader(test.url))
			w := httptest.NewRecorder()

			shortenURL(w, req)

			url := w.Body.String()

			testUrls[testURL] = url

			assert.Equal(t, test.want.statusCode, w.Code, "Код ответа не совпадает с ожидаемым")

			t.Log("w.Body: " + url)

			req2 := httptest.NewRequest(http.MethodGet, url, nil)

			w2 := httptest.NewRecorder()

			shortenURL(w2, req2)

			t.Log("w2.Location: " + w2.Header().Get("Location"))

			assert.Equal(t, 307, w2.Code, "Код ответа (307) не совпадает c ожидаемым")
			assert.Equal(t, testURL, w2.Header().Get("Location"), "Location не совпадает с ожидаемым")

		})
	}
}
