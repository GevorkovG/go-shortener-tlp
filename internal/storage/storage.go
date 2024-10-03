package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Storage struct {
	urls map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		urls: make(map[string]string),
	}
}

// -----------------------------------------
type InMemoryStorage struct {
	urls map[string]string
}

func (s *InMemoryStorage) Load(data map[string]string) {
	s.urls = data
}

func (s *Storage) SetURL(key, value string) {
	s.urls[key] = value
}

func (s *Storage) GetURL(key string) (string, error) {

	url, ok := s.urls[key]
	if ok {
		return url, nil
	}
	return "", errors.New("id not found")
}

type FileStorage struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func LoadFromFile(fileName string) (map[string]string, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	data := make(map[string]string)

	for scanner.Scan() {
		var d FileStorage
		// Декодируем строку из формата json
		err = json.Unmarshal(scanner.Bytes(), &d)
		if err != nil {
			log.Println(err)
		}

		data[d.Short] = d.Original
	}
	return data, nil
}
