package storage

import (
	"errors"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"go.uber.org/zap"
)

type InMemoryStorage struct {
	urls    map[string]string // short -> original
	userIDs map[string]string // short -> userID
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls:    make(map[string]string),
		userIDs: make(map[string]string),
	}
}

func (s *InMemoryStorage) Load(data map[string]string) {
	s.urls = data
}

func (s *InMemoryStorage) Insert(link *objects.Link) error {
	zap.L().Info("MEMORY Inserting URL", zap.String("short", link.Short), zap.String("original", link.Original), zap.String("userID", link.UserID))
	s.urls[link.Short] = link.Original
	s.userIDs[link.Short] = link.UserID // Сохраняем userID
	return nil
}

func (s *InMemoryStorage) InsertLinks(links []*objects.Link) error {
	for _, link := range links {
		s.urls[link.Short] = link.Original
		s.userIDs[link.Short] = link.UserID
	}
	return nil
}

func (s *InMemoryStorage) GetOriginal(short string) (*objects.Link, error) {
	original, exists := s.urls[short]
	if !exists {
		return nil, errors.New("short URL not found")
	}
	return &objects.Link{
		Short:    short,
		Original: original,
		UserID:   s.userIDs[short],
	}, nil
}

func (s *InMemoryStorage) GetShort(original string) (*objects.Link, error) {
	for short, orig := range s.urls {
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

func (s *InMemoryStorage) GetAllByUserID(userID string) ([]objects.Link, error) {
	zap.L().Info("Getting URLs for user", zap.String("userID", userID))

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
