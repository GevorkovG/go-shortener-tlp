// Main инициализирует и запускает HTTP/HTTPS сервер сокращения URL с полной конфигурацией.
// Функция выполняет:
//   - Загрузку конфигурации
//   - Инициализацию логгера
//   - Настройку роутера с middleware
//   - Запуск основного сервера
//   - Запуск pprof сервера для профилирования
//
// Конфигурация:
//   - Использует флаги командной строки и переменные окружения
//   - Логирует параметры при старте
//
// Middleware:
//   - Логирование запросов (LoggerMiddleware)
//   - Поддержка gzip сжатия (gzipMiddleware)
//   - Обработка cookies (cookies.Cookies)
//
// Роуты:
//
//	POST /api/shorten       - Сокращение URL через JSON API (JSONGetShortURL)
//	GET  /{id}              - Получение оригинального URL (GetOriginalURL)
//	GET  /ping              - Проверка доступности сервера (Ping)
//	POST /                  - Сокращение URL через форму (GetShortURL)
//	POST /api/shorten/batch - Пакетное сокращение URL (APIshortBatch)
//	GET  /api/user/urls     - Получение URL пользователя (APIGetUserURLs)
//	DELETE /api/user/urls   - Удаление URL пользователя (APIDeleteUserURLs)
//
// Особенности:
//   - Запускает параллельный pprof сервер на :6060
//   - Детально логирует параметры старта
//   - Использует zap для структурированного логгирования
//
// Пример запуска:
//
//	go run ./cmd/shortener/main.go
//
// Для профилирования:
//
//	go tool pprof http://localhost:6060/debug/pprof/profile
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	gr "github.com/GevorkovG/go-shortener-tlp/internal/grpc"
	"github.com/GevorkovG/go-shortener-tlp/internal/logger"
	"github.com/GevorkovG/go-shortener-tlp/internal/routes"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	conf := config.NewCfg()

	logger.InitLogger()

	newApp := app.NewApp(conf)

	GRPCServer := gr.NewServer(newApp.Storage, newApp)

	go func() {
		err := gr.Run(GRPCServer)
		if err != nil {
			log.Fatal("GRPC server not started", err)
		}
	}()

	// Создание HTTP сервера с таймаутами
	srv := &http.Server{
		Addr:         conf.Host,
		Handler:      routes.Router(newApp),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	pprofServer := &http.Server{Addr: ":6060"}

	// Создаем WaitGroup для ожидания завершения серверов
	var wg sync.WaitGroup

	// Создаем контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	// Запуск основного сервера
	wg.Add(1)
	go func() {
		defer wg.Done()

		zap.L().Info("Starting main server",
			zap.String("address", conf.Host),
			zap.Bool("https", conf.EnableHTTPS),
		)

		var err error
		if conf.EnableHTTPS {
			err = srv.ListenAndServeTLS("./certs/cert.pem", "./certs/key.pem")
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			zap.L().Error("Main server error", zap.Error(err))
			stop()
		}
	}()

	// Запуск pprof сервера
	wg.Add(1)
	go func() {
		defer wg.Done()

		zap.L().Info("Starting pprof server", zap.String("address", ":6060"))
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Pprof server error", zap.Error(err))
			stop()
		}
	}()

	// Ожидаем сигнал завершения
	<-ctx.Done()
	zap.L().Info("Received shutdown signal")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Main server shutdown error", zap.Error(err))
	}

	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Pprof server shutdown error", zap.Error(err))
	}

	wg.Wait()
	zap.L().Info("Server stopped gracefully")

}
