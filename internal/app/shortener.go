package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi"
)

func generateID() string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	number := rand.Uint64()
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}

// Request представляет входящий запрос на сокращение URL.
// Используется в JSON API эндпоинтах.
//
// Поля:
//   - URL string `json:"url"`: оригинальный URL для сокращения.
//
// Пример:
//
//	{
//	  "url": "https://example.com/very/long/url"
//	}
type Request struct {
	URL string `json:"url"`
}

// Response представляет ответ сервера с сокращенным URL.
// Возвращается клиенту при успешном сокращении.
//
// Поля:
//   - Result string `json:"result"`: сокращенный URL в формате:
//     http(s)://<домен>/<короткий-идентификатор>
//
// Пример:
//
//	{
//	  "result": "http://short.ly/abc123"
//	}
type Response struct {
	Result string `json:"result"`
}

// JSONGetShortURL обрабатывает запрос на сокращение URL в JSON формате.
// Эндпоинт: POST /api/shorten
//
// Входные данные (Request):
//
//	{
//	  "url": "string"  // URL для сокращения (обязательное поле)
//	}
//
// Возможные ответы:
//   - 201 Created: URL успешно сокращен
//     {
//     "result": "string"  // сокращенный URL
//     }
//   - 400 Bad Request: невалидный JSON или URL
//   - 409 Conflict: URL уже был сокращен ранее
//   - 500 Internal Server Error: ошибка сервера
//
// Логика работы:
//  1. Извлекает UserID из токена аутентификации (если есть)
//  2. Проверяет валидность входного JSON
//  3. Генерирует уникальный идентификатор для URL
//  4. Сохраняет связь URL-идентификатор в хранилище:
//     - Если URL уже существует, возвращает существующий сокращенный URL
//  5. Возвращает сокращенный URL в формате: {базовый_URL}/{идентификатор}
//
// Пример запроса:
//
//	POST /api/shorten
//	Content-Type: application/json
//	Authorization: Bearer <token>
//
//	{"url": "https://example.com/very/long/path"}
//
// Пример ответа:
//
//	HTTP/1.1 201 Created
//	Content-Type: application/json
//
//	{"result": "http://short.ly/abc123"}
//
// Особенности:
//   - Поддерживает аутентификацию через токен
//   - Автоматически обрабатывает конфликты (дубликаты URL)
//   - Логирует все ошибки и предупреждения
//   - Возвращает полный URL (с базовым доменом)
func (a *App) JSONGetShortURL(w http.ResponseWriter, r *http.Request) {
	var req Request
	var status = http.StatusCreated
	var UserID string

	// Извлекаем UserID из контекста
	token := r.Context().Value(cookies.SecretKey)
	if token != nil {
		var err error
		UserID, err = usertoken.GetUserID(token.(string))
		if err != nil {
			zap.L().Warn("Failed to get UserID from token, proceeding without it", zap.Error(err))
			UserID = ""
		}
	}

	// Декодируем тело запроса
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем объект Link
	link := &objects.Link{
		Short:    generateID(),
		Original: req.URL,
		UserID:   UserID, // Устанавливаем UserID
	}

	// Сохраняем ссылку в хранилище
	if err = a.Storage.Insert(r.Context(), link); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			// Если URL уже существует, получаем существующий короткий URL
			link, err = a.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Failed to get short URL", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			status = http.StatusConflict
		} else {
			zap.L().Error("Failed to insert URL", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Формируем ответ
	result := Response{
		Result: strings.TrimSpace(fmt.Sprintf("%s/%s", a.cfg.ResultURL, link.Short)),
	}

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(response)
	if err != nil {
		zap.L().Error("Failed to write response", zap.Error(err))
		return
	}
}

// GetShortURL обрабатывает запросы на создание коротких URL.
//
// Метод ожидает POST-запрос с оригинальным URL в теле запроса.
// При успешном выполнении возвращает сокращенный URL со статусом 201 (Created).
// Если URL уже существует, возвращает существующий сокращенный URL со статусом 409 (Conflict).
// Требуется аутентификация пользователя через контекст (userID может быть пустым для неавторизованных пользователей).
//
// Пример запроса:
//
//	POST /api/shorten
//	Тело запроса: "https://example.com"
//
// Пример ответа:
//
//	"http://short.url/abc123" (Статус 201 или 409)
//
// Возможные ошибки:
//   - 400: Неверный формат запроса или пустое тело
//   - 500: Внутренняя ошибка сервера
func (a *App) GetShortURL(w http.ResponseWriter, r *http.Request) {
	var status = http.StatusCreated

	// Извлекаем UserID из контекста
	userID, ok := r.Context().Value(cookies.SecretKey).(string)
	if !ok || userID == "" {
		// Если UserID не найден в контексте, логируем ошибку и продолжаем без него
		zap.L().Error("UserID not found in context")
		userID = "" // Устанавливаем пустой UserID
	}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/app/shortener.go  UserID: %s ", userID)

	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusBadRequest)
		return
	}
	if string(responseData) == "" {
		http.Error(w, "Empty POST request body!", http.StatusBadRequest)
		return
	}

	link := &objects.Link{
		Short:    generateID(),
		Original: string(responseData),
		UserID:   userID, // Устанавливаем UserID
	}

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/app/shortener.go GetShortURL Original:%s UserID:%s Short: %s", link.Original, link.UserID, link.Short)

	if err = a.Storage.Insert(r.Context(), link); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			link, err = a.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Don't get short URL", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			status = http.StatusConflict
		} else {
			zap.L().Error("Don't insert URL", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := strings.TrimSpace(fmt.Sprintf("%s/%s", a.cfg.ResultURL, link.Short))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	_, err = io.WriteString(w, response)
	if err != nil {
		zap.L().Error("Didn't write response", zap.Error(err))
		return
	}
}

// GetOriginalURL обрабатывает запросы на перенаправление по короткому URL.
//
// При успешном выполнении возвращает 307 (Temporary Redirect) с Location на оригинальный URL.
// Если URL помечен как удаленный, возвращает 410 (Gone).
// Если URL не найден, возвращает 400 (Bad Request).
//
// Пример запроса:
//
//	GET /abc123
//
// Пример ответа:
//
//	Заголовок Location: "https://example.com"
//	Статус: 307
//
// Возможные ошибки:
//   - 400: Неверный идентификатор короткого URL
//   - 410: URL был удален
func (a *App) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	link, err := a.Storage.GetOriginal(id)
	log.Printf("GetOriginalURL short:%s %t", link.Short, link.DeletedFlag)
	if err != nil {
		zap.L().Error("Failed to get original URL", zap.String("id", id), zap.Error(err))
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Проверяем, удален ли URL
	if link.DeletedFlag || link.Original == "" {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", link.Original)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Ping проверяет доступность хранилища.
//
// Возвращает 200 (OK) если хранилище доступно, 500 (Internal Server Error) в противном случае.
//
// Пример запроса:
//
//	GET /ping
//
// Пример ответа:
//
//	Статус: 200 OK или 500 Internal Server Error
func (a *App) Ping(w http.ResponseWriter, _ *http.Request) {
	if err := a.Storage.Ping(); err != nil {
		log.Println("Storage ping failed:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
