// Package log предоставляет кастомную реализацию логгера для приложения.
// Поддерживает различные уровни логирования (Debug, Info, Warn, Error)
// и конфигурируемые выходные потоки.
package log

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData хранит информацию о HTTP-ответе
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter оборачивает http.ResponseWriter для захвата статуса и размера ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write переопределяет метод Write для захвата размера ответа
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader переопределяет метод WriteHeader для захвата статуса ответа
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// ZapLogger *zap.Logger
var ZapLogger *zap.Logger

// InitLogger инициализирует логгер
func InitLogger() {
	ZapLogger, _ = zap.NewDevelopment()
}

// LoggerMiddleware — middleware для логирования HTTP-запросов
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обёртку для захвата статуса и размера ответа
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(&lw, r)

		// Логируем информацию о запросе
		ZapLogger.Sugar().Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", time.Since(start),
			"size", responseData.size, // получаем перехваченный размер ответа
			"loc", w.Header().Get("Location"),
		)
	})
}
