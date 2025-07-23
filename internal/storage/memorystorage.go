// Package storage предоставляет in-memory реализацию хранилища для сервиса сокращения URL.
// Хранит данные в оперативной памяти без персистентности.
package storage

import (
	"context"
	"errors"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"go.uber.org/zap"
)

// InMemoryStorage реализует хранилище ссылок в оперативной памяти
type InMemoryStorage struct {
	urls    map[string]string // short -> original
	userIDs map[string]string // short -> userID
}

// NewInMemoryStorage создает новое in-memory хранилище
//
// Возвращает:
//   - *InMemoryStorage: инициализированное хранилище с пустыми мапами
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls:    make(map[string]string),
		userIDs: make(map[string]string),
	}
}

// Load загружает данные в хранилище из переданной мапы
//
// Параметры:
//   - data: маппинг short→original URL
//
// Особенности:
//   - Полностью заменяет текущие данные
//   - Не затрагивает информацию о пользователях
func (s *InMemoryStorage) Load(data map[string]string) {
	s.urls = data
}

// Insert добавляет новую ссылку в хранилище
//
// Параметры:
//   - ctx: контекст выполнения
//   - link: объект ссылки для добавления
//
// Возвращает:
//   - error: всегда nil
//
// Логирует:
//   - Информацию о добавляемой ссылке
func (s *InMemoryStorage) Insert(ctx context.Context, link *objects.Link) error {
	zap.L().Info("Inserting URL", zap.String("short", link.Short), zap.String("original", link.Original), zap.String("userID", link.UserID))

	s.urls[link.Short] = link.Original
	s.userIDs[link.Short] = link.UserID

	zap.L().Debug("internal/storage/memorystorage.go Insert",
		zap.String("userID", link.UserID),
		zap.String("original", link.Original),
	)

	return nil
}

// InsertLinks добавляет несколько ссылок атомарно
//
// Параметры:
//   - ctx: контекст выполнения
//   - links: массив ссылок для добавления
//
// Возвращает:
//   - error: всегда nil (ошибки невозможны в текущей реализации)
//
// Логирует:
//   - Начало и завершение операции
func (s *InMemoryStorage) InsertLinks(ctx context.Context, links []*objects.Link) error {
	zap.L().Info("MEMORY Inserting multiple URLs", zap.Any("links", links))

	for _, link := range links {
		s.urls[link.Short] = link.Original
		s.userIDs[link.Short] = link.UserID
	}

	zap.L().Info("MEMORY URLs inserted successfully", zap.Any("links", links))
	return nil
}

// GetOriginal возвращает оригинальный URL по его сокращенной версии
//
// Параметры:
//   - short: сокращенный URL
//
// Возвращает:
//   - *objects.Link: найденная ссылка с userID
//   - error: "short URL not found" если ссылка не существует
func (s *InMemoryStorage) GetOriginal(short string) (*objects.Link, error) {
	original, exists := s.urls[short]
	zap.L().Debug("internal/storage/memorystorage.go GetOriginal",
		zap.String("userID", s.userIDs[short]),
		zap.String("short", s.urls[short]),
		zap.String("original", original),
	)

	if !exists {
		return nil, errors.New("short URL not found")
	}
	return &objects.Link{
		Short:    short,
		Original: original,
		UserID:   s.userIDs[short],
	}, nil
}

// GetShort возвращает сокращенный URL по оригинальному
//
// Параметры:
//   - original: оригинальный URL
//
// Возвращает:
//   - *objects.Link: найденная ссылка с userID
//   - error: "original URL not found" если ссылка не существует
func (s *InMemoryStorage) GetShort(original string) (*objects.Link, error) {
	for short, orig := range s.urls {
		zap.L().Debug("internal/storage/memorystorage.go GetShort",
			zap.String("userID", s.userIDs[short]),
			zap.String("short", short),
			zap.String("original", original),
		)

		if orig == original {
			return &objects.Link{
				Short:    short,
				Original: original,
				UserID:   s.userIDs[short],
			}, nil
		}
	}
	return nil, errors.New("original URL not found")
}

// GetAllByUserID возвращает все ссылки принадлежащие указанному пользователю
//
// Параметры:
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - []objects.Link: массив ссылок пользователя (может быть пустым)
//   - error: всегда nil
func (s *InMemoryStorage) GetAllByUserID(userID string) ([]objects.Link, error) {
	zap.L().Info("Getting URLs for user", zap.String("userID", userID))
	zap.L().Debug("internal/storage/memorystorage.go GetAllByUserID",
		zap.String("UserID", userID),
	)

	var userLinks []objects.Link

	// Проходим по всем URL и фильтруем по userID
	for short, original := range s.urls {
		if s.userIDs[short] == userID { // Проверяем, что URL принадлежит userID
			userLinks = append(userLinks, objects.Link{
				Short:    short,
				Original: original,
				UserID:   userID,
			})
		}
	}

	zap.L().Info("Retrieved URLs for user", zap.String("userID", userID), zap.Any("userLinks", userLinks))

	if len(userLinks) == 0 {
		return nil, nil // Если URL не найдены, возвращаем nil
	}

	return userLinks, nil
}

// MarkAsDeleted помечает ссылку как удаленную
//
// Параметры:
//   - userID: идентификатор пользователя
//   - short: сокращенный URL
//
// Возвращает:
//   - error: "URL not found or user mismatch" если:
//   - ссылка не найдена
//   - ссылка принадлежит другому пользователю
//
// Особенности:
//   - Устанавливает original URL в пустую строку
//   - Сохраняет userID
func (s *InMemoryStorage) MarkAsDeleted(userID string, short string) error {
	if s.userIDs[short] == userID {
		s.urls[short] = ""        // Помечаем URL как удаленный
		s.userIDs[short] = userID // Сохраняем userID
		return nil
	}
	return errors.New("URL not found or user mismatch")
}

// Ping проверяет доступность хранилища
func (s *InMemoryStorage) Ping() error {
	return nil
}

// GetStats возвращает статистику сервиса:
//   - количество уникальных сокращённых URL
//   - количество уникальных пользователей
//
// Возвращает:
//   - urls: количество URL
//   - users: количество пользователей
//   - error: всегда nil (ошибки невозможны в текущей реализации)
func (s *InMemoryStorage) GetStats(ctx context.Context) (urls int, users int, err error) {
	// Считаем уникальные URL
	urls = len(s.urls)

	// Собираем уникальных пользователей
	uniqueUsers := make(map[string]struct{})
	for _, userID := range s.userIDs {
		if userID != "" { // Игнорируем пустые userID (если такие есть)
			uniqueUsers[userID] = struct{}{}
		}
	}
	users = len(uniqueUsers)

	zap.L().Debug("Storage stats",
		zap.Int("urls", urls),
		zap.Int("users", users),
	)

	return urls, users, nil
}
