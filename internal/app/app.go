// Package app предоставляет основную логику работы сервера сокращения URL.
// Включает обработчики HTTP-запросов, middleware для сжатия данных и управления cookies,
// а также интеграцию с хранилищем данных и системой логирования.
package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

// App представляет основное приложение с его зависимостями.
// Содержит:
//   - Конфигурацию приложения
//   - Хранилище данных
type App struct {
	cfg     *config.AppConfig
	Storage objects.Storage
}

type contextKey string

// Token - ключ контекста для хранения и извлечения токена аутентификации.
// Используется в middleware и обработчиках для передачи токена между слоями приложения.
const Token contextKey = "token"

// NewApp создает и инициализирует новый экземпляр приложения с заданной конфигурацией.
// В зависимости от параметров конфигурации выбирается соответствующее хранилище:
//   - Если указана строка подключения к БД (DataBaseString), используется PostgreSQL хранилище
//   - Если указан путь к файлу (FilePATH), используется файловое хранилище
//   - Если не указано ни то, ни другое, используется хранилище в памяти
//
// Параметры:
//   - cfg *config.AppConfig: конфигурация приложения, должна содержать:
//   - DataBaseString: строка подключения к PostgreSQL (необязательно)
//   - FilePATH: путь к файлу для хранения данных (необязательно)
//
// Возвращает:
//   - *App: указатель на созданный экземпляр приложения, содержащий:
//   - cfg: переданную конфигурацию
//   - Storage: инициализированное хранилище данных
//
// Пример использования:
//
//	config := &config.AppConfig{
//	    DataBaseString: "postgres://user:pass@localhost/db",
//	}
//	app := NewApp(config)
//
// Примечания:
//   - Функция логирует выбранный тип хранилища
//   - Приоритет выбора хранилища: БД > Файл > Память
//   - Переданная конфигурация сохраняется по ссылке, изменения в cfg после создания
//     приложения будут влиять на его работу
func NewApp(cfg *config.AppConfig) *App {
	var store objects.Storage

	switch {
	case cfg.DataBaseString != "":
		log.Printf("internal/app/app.go ValidationToken USE DataBase")
		db := database.InitDB(cfg.DataBaseString)
		store = storage.NewLinkStorage(db)
	case cfg.FilePATH != "":
		log.Printf("internal/app/app.go ValidationToken USE FilePATH ")
		store = storage.NewFileStorage(cfg.FilePATH)
	default:
		log.Printf("internal/app/app.go ValidationToken USE NewInMemoryStorage ")
		store = storage.NewInMemoryStorage()
	}

	return &App{
		cfg:     cfg,
		Storage: store,
	}
}

// GetConfig возвращает текущую конфигурацию приложения.
func (a *App) GetConfig() *config.AppConfig {
	return a.cfg
}

// GetStats возвращает статистику сервиса
func (a *App) GetStats(w http.ResponseWriter, r *http.Request) {
	// Получаем статистику из хранилища
	urls, users, err := a.Storage.GetStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	stats := struct {
		URLs  int `json:"urls"`
		Users int `json:"users"`
	}{
		URLs:  urls,
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
