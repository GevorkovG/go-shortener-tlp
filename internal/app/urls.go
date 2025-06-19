package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"go.uber.org/zap"
)

// RespURLs представляет структуру ответа с URL для API.
// Используется для сериализации в JSON при возврате списка URL пользователя.
type RespURLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

// APIGetUserURLs возвращает все URL, созданные текущим пользователем.
//
// Метод требует аутентификации - userID должен быть в контексте запроса.
// Возвращает:
// - 200 OK с массивом URL в формате JSON при успехе
// - 204 No Content если у пользователя нет сохраненных URL
// - 401 Unauthorized если пользователь не аутентифицирован
// - 500 Internal Server Error при ошибках сервера
//
// Пример успешного ответа:
//
//	[
//	  {
//	    "short_url": "http://short.ru/abc",
//	    "original_url": "https://original.com/long"
//	  }
//	]
func (a *App) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := r.Context().Value(cookies.SecretKey).(string)

	zap.L().Debug("internal/app/urls.go APIGetUserURLs",
		zap.Bool("empty UserID?", userID == ""),
		zap.String("userID", userID),
	)

	if !ok || userID == "" {
		zap.L().Warn("Unauthorized access attempt")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized) // Устанавиваем статус-код
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}
	zap.L().Info("UserID extracted from context", zap.String("userID", userID))

	// Получаем URL-адреса пользователя
	userURLs, err := a.Storage.GetAllByUserID(userID)
	if err != nil {
		zap.L().Error("Failed to get user URLs", zap.String("userID", userID), zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Устанавиваем статус-код
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal Server Error"})
		return
	}

	// Логируем количество найденных URL
	zap.L().Info("Number of URLs found", zap.Int("count", len(userURLs)))

	zap.L().Debug("internal/app/urls.go APIGetUserURLs",
		zap.String("userID", userID),
		zap.Int("Number of URLs found", len(userURLs)),
	)

	if len(userURLs) == 0 {
		zap.L().Info("No URLs found for user", zap.String("userID", userID))
		w.WriteHeader(http.StatusNoContent) // Возвращаем 204, если данных нет
		return
	}

	// Формируем ответ
	links := make([]RespURLs, 0, len(userURLs))
	for _, val := range userURLs {
		links = append(links, RespURLs{
			Short:    strings.TrimSpace(a.cfg.ResultURL + "/" + val.Short),
			Original: strings.TrimSpace(val.Original),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(links); err != nil {
		zap.L().Error("Failed to write response", zap.Error(err))
	}
}

// APIDeleteUserURLs асинхронно удаляет указанные URL пользователя.
//
// Принимает массив сокращенных URL в теле запроса (JSON).
// Удаление происходит асинхронно с использованием паттерна FanOut/FanIn.
// Возвращает:
// - 202 Accepted если запрос принят в обработку
// - 401 Unauthorized если пользователь не аутентифицирован
// - 400 Bad Request при неверном формате запроса
//
// Особенности:
// - URL удаляются в нескольких горутинах для повышения производительности
// - Ответ возвращается сразу после принятия запроса, не дожидаясь завершения удаления
// - Результаты удаления логируются
func (a *App) APIDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	var shortURLs []string
	userID, ok := r.Context().Value(cookies.SecretKey).(string)

	if !ok || userID == "" {
		zap.L().Warn("Unauthorized access attempt")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&shortURLs)
	if err != nil {
		zap.L().Error("Failed to decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Канал для завершения работы горутин
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Создаем несколько горутин для обработки URL (FanOut)
	channels := fanOut(doneCh, userID, shortURLs, a.Storage)

	// Объединяем результаты из всех горутин (FanIn)
	finalCh := fanIn(doneCh, channels...)

	// Ожидаем завершения всех горутин
	for success := range finalCh {
		if !success {
			zap.L().Warn("Failed to delete some URLs")
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// fanOut создает несколько горутин для обработки каждого URL.
func fanOut(doneCh chan struct{}, userID string, shortURLs []string, storage objects.Storage) []chan bool {
	// Количество горутин (можно настроить в зависимости от нагрузки)
	numWorkers := 5
	channels := make([]chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// Создаем канал для результатов
		resultCh := make(chan bool, 1)
		channels[i] = resultCh

		// Запускаем горутину для обработки URL
		go func(resultCh chan bool) {
			defer close(resultCh)

			for _, short := range shortURLs {
				log.Printf("fanOUT short: %s", short)
				select {
				case <-doneCh: // Проверяем сигнал завершения
					return
				default:
					err := storage.MarkAsDeleted(userID, short)
					if err != nil {
						log.Printf("fanOUT short: %s", err)
						zap.L().Error("Failed to mark URL as deleted", zap.String("short", short), zap.Error(err))
						resultCh <- false
					} else {
						resultCh <- true
					}
				}
			}
		}(resultCh)
	}

	return channels
}

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(doneCh chan struct{}, resultChs ...chan bool) chan bool {
	finalCh := make(chan bool)

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
