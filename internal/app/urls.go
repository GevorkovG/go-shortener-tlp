package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/GevorkovG/go-shortener-tlp/internal/cookies"
	"go.uber.org/zap"
)

type RespURLs struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (a *App) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста
	userID, ok := r.Context().Value(cookies.SecretKey).(string)

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/app/urls.go  APIGetUserURLs %t userID %s", userID == "", userID)

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

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("userID %s, Number of URLs found %d", userID, len(userURLs))

	if len(userURLs) == 0 {
		zap.L().Info("No URLs found for user", zap.String("userID", userID))
		w.WriteHeader(http.StatusNoContent) // Возвращаем 204, если данных нет
		return
	}

	// Формируем ответ
	var links []RespURLs
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

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(doneCh chan struct{}, resultChs ...chan bool) chan bool {
	finalCh := make(chan bool)

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

	// Создаем каналы для каждого URL
	resultChs := make([]chan bool, len(shortURLs))
	for i := range resultChs {
		resultChs[i] = make(chan bool, 1)
	}

	// Запускаем горутины для удаления каждого URL
	for i, short := range shortURLs {
		go func(short string, resultCh chan bool) {
			err := a.Storage.MarkAsDeleted(userID, short)
			if err != nil {
				zap.L().Error("Failed to mark URL as deleted", zap.String("short", short), zap.Error(err))
				resultCh <- false
			} else {
				resultCh <- true
			}
			close(resultCh)
		}(short, resultChs[i])
	}

	// Объединяем результаты с помощью FanIn
	finalCh := fanIn(doneCh, resultChs...)

	// Ожидаем завершения всех горутин
	for success := range finalCh {
		if !success {
			zap.L().Warn("Failed to delete some URLs")
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
