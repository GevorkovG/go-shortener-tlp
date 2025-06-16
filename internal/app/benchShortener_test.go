package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/go-chi/chi"
)

func Benchmark_GetOriginalURL(b *testing.B) {
	// Инициализация
	conf := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}
	//app := NewApp(conf)
	userID := "test-user-id"
	resultURL := "https://yandex.ru"
	shortURL := "vRFgdzs"

	// Подготовка данных (делается один раз)
	link := &objects.Link{
		Short:    shortURL,
		Original: resultURL,
		UserID:   userID,
	}
	ctx := context.Background()
	if err := app.Storage.Insert(ctx, link); err != nil {
		b.Fatal(err)
	}

	// Сбрасываем таймер перед основным циклом
	b.ResetTimer()

	// Основной цикл бенчмарка (выполняется b.N раз)
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+shortURL, nil)
		ctx = context.WithValue(r.Context(), cookies.SecretKey, userID)
		w := httptest.NewRecorder()
		router := chi.NewRouteContext()
		router.URLParams.Add("id", shortURL)
		r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, router))

		app.GetOriginalURL(w, r)
		// В бенчмарках обычно не проверяют assertions, только если не для верификации
		if w.Code != http.StatusTemporaryRedirect {
			b.Errorf("unexpected status code: got %v want %v", w.Code, http.StatusTemporaryRedirect)
		}
	}
}
