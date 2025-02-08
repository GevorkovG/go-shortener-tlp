package storage

import (
	"errors"
	"log"

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
	zap.L().Info("Inserting URL", zap.String("short", link.Short), zap.String("original", link.Original), zap.String("userID", link.UserID))

	s.urls[link.Short] = link.Original
	s.userIDs[link.Short] = link.UserID

	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/storage/memorystorage.go Insert Original:%s UserID:%s", link.Original, link.UserID)

	return nil
}

func (s *InMemoryStorage) InsertLinks(links []*objects.Link) error {
	zap.L().Info("MEMORY Inserting multiple URLs", zap.Any("links", links))

	for _, link := range links {
		s.urls[link.Short] = link.Original
		s.userIDs[link.Short] = link.UserID
	}

	zap.L().Info("MEMORY URLs inserted successfully", zap.Any("links", links))
	return nil
}

func (s *InMemoryStorage) GetOriginal(short string) (*objects.Link, error) {
	original, exists := s.urls[short]
	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/storage/memorystorage.go GetOriginal Original:%s Short:%s UserID:%s", original, s.urls[short], s.userIDs[short])

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
		//DEBUG--------------------------------------------------------------------------------------------------
		log.Printf("--------------internal/storage/memorystorage.go GetShort userID %s SHORT %s ORIGIN %s", s.userIDs[short], short, original)
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
	//DEBUG--------------------------------------------------------------------------------------------------
	log.Printf("internal/storage/memorystorage.go GetAllByUserID userID %s", userID)

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

func (s *InMemoryStorage) MarkAsDeleted(userID string, short string) error {
	if s.userIDs[short] == userID {
		s.urls[short] = ""        // Помечаем URL как удаленный
		s.userIDs[short] = userID // Сохраняем userID
		return nil
	}
	return errors.New("URL not found or user mismatch")
}
