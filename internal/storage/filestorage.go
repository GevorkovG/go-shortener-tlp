package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"go.uber.org/zap"
)

type FileStorage struct {
	memStorage *InMemoryStorage
	filePATH   string
}

func NewFileStorage(path string) *FileStorage {
	fs := FileStorage{
		memStorage: NewInMemoryStorage(),
		filePATH:   path,
	}
	fs.ConfigureFileStorage()

	return &fs
}

func (fs *FileStorage) ConfigureFileStorage() {

	data, err := LoadFromFile(fs.filePATH)

	if err != nil {
		zap.L().Fatal("Don't load from file!", zap.Error(err))
	}

	fs.Load(data)
}

func SaveToFile(fs *objects.Link, fileName string) error {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(fs)
	return err
}

func AllSaveToFile(links []*objects.Link, fileName string) error {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, v := range links {
		err = encoder.Encode(v)
	}
	return err
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
		var d objects.Link
		// Декодируем строку из формата json
		err = json.Unmarshal(scanner.Bytes(), &d)
		if err != nil {
			zap.L().Error("error scan ", zap.Error(err))
		}

		data[d.Short] = d.Original
	}
	return data, nil
}

func (fs *FileStorage) Load(data map[string]string) {
	fs.memStorage.Load(data)
}

func (fs *FileStorage) Insert(ctx context.Context, link *objects.Link) error {
	zap.L().Info("FILE Inserting URL", zap.String("short", link.Short), zap.String("original", link.Original), zap.String("userID", link.UserID))

	err := fs.memStorage.Insert(ctx, link)
	if err != nil {
		return err
	}

	err2 := SaveToFile(link, fs.filePATH)
	if err2 != nil {
		return err2
	}

	zap.L().Info("FILE URL inserted successfully", zap.String("short", link.Short), zap.String("original", link.Original), zap.String("userID", link.UserID))
	return nil
}

func (fs *FileStorage) InsertLinks(ctx context.Context, links []*objects.Link) error {

	err := fs.memStorage.InsertLinks(ctx, links)
	if err != nil {
		return err
	}
	err2 := AllSaveToFile(links, fs.filePATH)
	if err2 != nil {
		return err2
	}
	return err
}

func (fs *FileStorage) GetOriginal(short string) (*objects.Link, error) {
	link, err := fs.memStorage.GetOriginal(short)
	if err != nil {
		zap.L().Error("Failed to get original URL", zap.String("short", short), zap.Error(err))
		return nil, err
	}

	// Логируем успешное получение оригинального URL
	zap.L().Info("Successfully retrieved original URL", zap.String("short", short), zap.String("original", link.Original))
	return link, nil
}

func (fs *FileStorage) GetShort(original string) (*objects.Link, error) {

	link, err := fs.memStorage.GetShort(original)

	if err != nil {
		zap.L().Error("Don't get short URL", zap.Error(err))
		return link, err
	}
	return link, nil
}

func (fs *FileStorage) GetAllByUserID(userID string) ([]objects.Link, error) {
	zap.L().Info("Getting URLs for user", zap.String("userID", userID))
	userLinks := make([]objects.Link, 0, len(fs.memStorage.urls))

	zap.L().Info("Querying user URLs from file storage", zap.String("userID", userID))

	// Проходим по всем ссылкам в памяти и фильтруем по userID
	for short, original := range fs.memStorage.urls {
		// Предполагаем, что userID сохраняется при вставке
		link := objects.Link{
			Short:    short,
			Original: original,
			UserID:   userID,
		}
		userLinks = append(userLinks, link)
	}

	zap.L().Info("User URLs retrieved from file storage", zap.String("userID", userID), zap.Any("userLinks", userLinks))

	if len(userLinks) == 0 {
		return nil, nil
	}

	return userLinks, nil
}

func (fs *FileStorage) MarkAsDeleted(userID string, short string) error {
	if fs.memStorage.userIDs[short] == userID {
		fs.memStorage.urls[short] = ""        // Помечаем URL как удаленный
		fs.memStorage.userIDs[short] = userID // Сохраняем userID
		return nil
	}
	return errors.New("URL not found or user mismatch")
}

func (fs *FileStorage) Ping() error {
	return nil
}
