package app_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/go-chi/chi/v5"
)

func ExampleApp_Ping() {
	// Инициализация конфигурации
	cfg := &config.AppConfig{
		Host:      "localhost:8080",
		ResultURL: "http://localhost:8080",
	}

	// Создание приложения
	app := app.NewApp(cfg)

	// Создаем chi роутер
	r := chi.NewRouter()

	// Добавляем обработчик ping (предполагая, что у app есть метод Ping)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		app.Ping(w, r) // Используем существующий метод Ping
	})

	// Создаем тестовый HTTP-сервер
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Создаем тестовый запрос
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Добавляем необходимые контексты
	ctx := context.WithValue(req.Context(), cookies.SecretKey, "test-user-id")
	routerCtx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, routerCtx))

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Проверяем ответ
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Service is healthy")
	} else {
		fmt.Println("Service is unavailable")
	}

	// Output:
	// Service is healthy
}
