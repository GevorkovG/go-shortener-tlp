package app

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func Test_GetOriginalURL(t *testing.T) {

	//очищаем флаги командной строки
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	conf := config.NewCfg()
	app := NewApp(conf)

	resultURL := "https://yandex.ru"

	type want struct {
		code     int
		location string
	}

	type testInf struct {
		method string
		url    string
		testID string
		want   want
	}

	test := testInf{
		method: http.MethodGet,
		url:    "http://localhost:8080/vRFgdzs",
		testID: "vRFgdzs",
		want: want{
			code:     307,
			location: resultURL,
		},
	}

	app.Storage.SetURL(test.testID, resultURL)

	r := httptest.NewRequest(test.method, test.url, nil)

	w := httptest.NewRecorder()

	router := chi.NewRouteContext()

	router.URLParams.Add("id", test.testID)

	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, router))

	app.GetOriginURL(w, r)

	assert.Equal(t, test.want.code, w.Code, "Код ответа (307) не совпадает с ожидаемым")
	assert.Equal(t, test.want.location, w.Header().Get("Location"), "Location не совпадает с ожидаемым")

}

func Test_JSONGetShortURL(t *testing.T) {

	//очищаем флаги командной строки
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	type (
		want struct {
			code        int
			contentType string
			//location    string
		}

		testInf struct {
			method string
			want   want
		}
	)

	test := testInf{
		method: http.MethodPost,
		want: want{
			code:        201,
			contentType: "application/json",
		},
	}

	conf := config.NewCfg()
	app := NewApp(conf)

	r := httptest.NewRequest(test.method, "https://localhost:8080", strings.NewReader(`{"url": "https://yandex.ru"}`))

	w := httptest.NewRecorder()

	app.JSONGetShortURL(w, r)

	assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
	assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")

}

func Test_GetShortURL(t *testing.T) {

	//очищаем флаги командной строки
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	type want struct {
		code        int
		contentType string
		//location    string
	}
	testURL := "https://yandex.ru"

	type testInf struct {
		method string
		url    string
		want   want
	}

	test := testInf{
		method: http.MethodPost,
		url:    testURL,
		want: want{
			code:        201,
			contentType: "text/plain",
		},
	}

	conf := config.NewCfg()
	app := NewApp(conf)

	r := httptest.NewRequest(test.method, "https://localhost:8080", strings.NewReader(test.url))

	w := httptest.NewRecorder()

	app.GetShortURL(w, r)

	assert.Equal(t, test.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
	assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"), "Тип контента не совпадает с ожидаемым")

}
