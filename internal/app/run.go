package app

import (
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	logg "github.com/GevorkovG/go-shortener-tlp/internal/log"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// compressWriter реализует интерфейс http.ResponseWriter с поддержкой gzip-сжатия.
// Автоматически применяет сжатие для успешных ответов (statusCode < 300).
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает HTTP-заголовки ответа.
// Позволяет получать и модифицировать заголовки до записи тела ответа.
//
// Возвращает:
//   - http.Header: map-подобную структуру с заголовками ответа
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает сжатые данные в ответ.
// Автоматически применяет gzip-сжатие если статус ответа < 300.
//
// Параметры:
//   - p []byte: данные для записи
//
// Возвращает:
//   - int: количество записанных байт
//   - error: ошибка записи (если возникла)
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает код статуса HTTP ответа.
// Автоматически добавляет заголовок Content-Encoding: gzip для успешных ответов (statusCode < 300).
//
// Параметры:
//   - statusCode int: HTTP статус-код ответа
//
// Особенности:
//   - Заголовок Content-Encoding устанавливается только для успешных ответов
//   - Реальные заголовки записываются в нижележащий ResponseWriter
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close завершает операцию gzip-сжатия и освобождает ресурсы.
//
// Этот метод:
//   - Завершает запись сжатых данных
//   - Обеспечивает корректное завершение gzip-потока
//   - Освобождает ресурсы, связанные с компрессором
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader создает новый экземпляр compressReader для чтения сжатых gzip данных.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и распаковывает данные из сжатого потока.
//
// Параметры:
//   - p []byte: буфер для записи распакованных данных
//
// Возвращает:
//   - n int: количество прочитанных байт
//   - err error: ошибка чтения, включая:
//   - io.EOF при завершении потока
//   - gzip.ErrHeader при неверном формате gzip
//   - другие ошибки ввода-вывода
//
// Особенности:
//   - Данные автоматически распаковываются при чтении
//   - Сохраняет семантику стандартного io.Reader
//   - Может возвращать 0, nil при ожидании данных
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close освобождает ресурсы и закрывает оба потока:
//   - Нижележащий источник данных (io.ReadCloser)
//   - Gzip-распаковщик (*gzip.Reader)
//
// Возвращает:
//   - error: первая возникшая ошибка закрытия
//
// Гарантии:
//   - Всегда пытается закрыть оба потока, даже при ошибках
//   - Возвращает первую обнаруженную ошибку
//   - Последующие вызовы Read после Close возвращают ошибку
//
// Рекомендации:
//   - Всегда должен вызываться через defer после создания
//   - Не безопасен для конкурентного вызова
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer func(cw *compressWriter) {
				err := cw.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}(cw)
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer func(cr *compressReader) {
				err := cr.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}(cr)
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}

// Run инициализирует и запускает HTTP/HTTPS сервер сокращения URL с полной конфигурацией.
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
func Run() {

	conf := config.NewCfg()
	logg.InitLogger() // Инициализация логгера
	defer func() {
		// Принудительно сбрасываем буфер логов
		if err := zap.L().Sync(); err != nil {
			// Sync может возвращать ошибку для stderr (это нормально)
			log.Printf("Failed to sync zap logs: %v", err)
		}
		log.Println("Server shutdown completed")
	}()

	newApp := NewApp(conf)
	r := chi.NewRouter()
	r.Use(logg.LoggerMiddleware,
		gzipMiddleware,
		cookies.Cookies,
	)

	// Логируем информацию о запуске сервера
	logg.Logger.Info("Starting server",
		zap.String("host", conf.Host),
		zap.String("pprof_host", "localhost:6060"),
		zap.String("base_url", conf.ResultURL),
	)

	r.Post("/api/shorten", newApp.JSONGetShortURL)
	r.Get("/{id}", newApp.GetOriginalURL)
	r.Get("/ping", newApp.Ping)
	r.Post("/", newApp.GetShortURL)
	r.Post("/api/shorten/batch", newApp.APIshortBatch)
	r.Get("/api/user/urls", newApp.APIGetUserURLs)
	r.Delete("/api/user/urls", newApp.APIDeleteUserURLs)
	r.Get("/api/internal/stats", newApp.GetStats)

	// Создание HTTP сервера с таймаутами
	srv := &http.Server{
		Addr:         conf.Host,
		Handler:      r,
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
	wg.Add(2)
	go func() {
		defer wg.Done()

		zap.L().Info("Starting server",
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
			zap.L().Error("Server error", zap.Error(err))
			stop() // Инициируем shutdown при ошибке
		}
	}()

	// Запуск pprof сервера
	go func() {
		defer wg.Done()

		zap.L().Info("Starting pprof server", zap.String("address", ":6060"))
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Pprof server error", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения или ошибку сервера
	<-ctx.Done()
	zap.L().Info("Received shutdown signal")

	// Graceful shutdown с таймаутом
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Останавливаем основной сервер
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Main server shutdown error", zap.Error(err))
	}

	// Останавливаем pprof сервер
	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Pprof server shutdown error", zap.Error(err))
	}

	// Ждем завершения всех горутин
	wg.Wait()
	zap.L().Info("Server stopped gracefully")
}
