package storage

import (
	"errors"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
)

type InMemoryStorage struct {
	urls map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls: make(map[string]string),
	}
}

func (s *InMemoryStorage) Load(data map[string]string) {
	s.urls = data
}

func (s *InMemoryStorage) Insert(link objects.Link) error {
	s.urls[link.Short] = link.Original
	return nil
}

func (s *InMemoryStorage) InsertLinks(links []objects.Link) error {

	for _, v := range links {
		s.urls[v.Short] = v.Original
	}
	return nil
}

func (s *InMemoryStorage) GetURL(short string) (objects.Link, error) {

	var ok bool
	link := objects.Link{
		Short: short,
	}
	link.Original, ok = s.urls[link.Short]
	if ok {
		return link, nil
	}
	return link, errors.New("id not found")
}
