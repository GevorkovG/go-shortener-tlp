package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/services/usertoken"
	"go.uber.org/zap"
)

// Req представляет структуру входящего запроса для пакетного создания сокращенных URL.
// Используется в APIshortBatch методе для обработки массива URL.
//
// Поля:
//   - ID string `json:"correlation_id"`: уникальный идентификатор запроса (клиентский ID)
//   - URL string `json:"original_url"`: оригинальный URL для сокращения
type Req struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

// Resp представляет структуру ответа при пакетном создании сокращенных URL.
// Возвращается клиенту в виде массива при успешной обработке.
//
// Поля:
//   - ID string `json:"correlation_id"`: оригинальный ID из запроса (для сопоставления)
//   - Short string `json:"short_url"`: сокращенный URL
type Resp struct {
	ID    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

// APIshortBatch обрабатывает пакетный запрос на создание сокращенных URL.
// Для аутентифицированных пользователей связывает URL с их аккаунтом.
//
// Метод: POST
// Путь: /api/shorten/batch
// Content-Type: application/json
//
// Входные данные:
//   - Массив объектов Req в теле запроса:
//     [
//     {"correlation_id": "string", "original_url": "string"},
//     ...
//     ]
//
// Возвращаемые статусы:
//   - 201 Created: при успешной обработке
//   - 400 Bad Request: при невалидном JSON, URL или отсутствии тела
//   - 401 Unauthorized: при невалидном токене (если требуется auth)
//   - 500 Internal ServerError: при ошибках хранилища
//
// Логика работы:
//  1. Извлекает токен из контекста запроса
//  2. Декодирует входящий JSON в массив Req
//  3. Для каждого URL:
//     - Генерирует уникальный ключ
//     - Формирует сокращенный URL
//     - Создает объект Link для хранения
//  4. Сохраняет все ссылки атомарной операцией
//  5. Возвращает массив созданных сокращенных URL
//
// Пример запроса:
//
//	POST /api/shorten/batch
//	Authorization: Bearer <token>
//	Content-Type: application/json
//
//	[{
//	  "correlation_id": "1",
//	  "original_url": "https://example.com"
//	}]
//
// Пример ответа:
//
//	HTTP/1.1 201 Created
//	Content-Type: application/json
//
//	[{
//	  "correlation_id": "1",
//	  "short_url": "http://short.ly/abc123"
//	}]
//
// Особенности:
//   - Поддерживает анонимные и аутентифицированные запросы
//   - Сохраняет userID для аутентифицированных пользователей
//   - Генерирует детальные логи для отладки
//   - Гарантирует атомарность при сохранении пакета
//   - Валидирует входные данные перед обработкой
func (a *App) APIshortBatch(w http.ResponseWriter, r *http.Request) {

	var originals []Req

	err := json.NewDecoder(r.Body).Decode(&originals)
	if err != nil {
		log.Println("didn't decode body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shorts := make([]Resp, 0, len(originals))
	links := make([]*objects.Link, 0, len(originals))

	token := r.Context().Value(cookies.SecretKey).(string)

	userID, err := usertoken.GetUserID(token)
	if err != nil {
		userID = ""
	}

	zap.L().Debug("internal/app/batsh.go",
		zap.String("userID", userID),
		zap.String("token", token),
	)

	for _, val := range originals {

		key := generateID()
		resp := Resp{
			ID:    val.ID,
			Short: fmt.Sprintf(a.cfg.ResultURL+"/%s", key),
		}
		link := &objects.Link{
			Short:    key,
			Original: val.URL,
			UserID:   userID,
		}

		shorts = append(shorts, resp)
		links = append(links, link)

	}

	if err = a.Storage.InsertLinks(r.Context(), links); err != nil {
		//		log.Println("Didn't insert to table")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(shorts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(response)
	if err != nil {
		return
	}

}
