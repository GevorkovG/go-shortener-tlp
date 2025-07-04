// Package objects определяет основные структуры данных и интерфейсы хранилища
// для сервиса сокращения URL
package objects

import "context"

// Link представляет структуру для хранения информации о URL.
// Используется для хранения как оригинальных, так и сокращенных URL
type Link struct {
	Short       string `json:"short_url"`    //Сокращенный URL
	Original    string `json:"original_url"` //Оригинальный URL
	UserID      string `json:"-"`            //ID пользователя
	DeletedFlag bool   `json:"-"`            //Флаг удаления
}

// Storage определяет интерфейс для работы с хранилищем URL.
// Реализации должны поддерживать все указанные методы.
type Storage interface {
	Insert(ctx context.Context, link *Link) error
	InsertLinks(ctx context.Context, links []*Link) error
	GetOriginal(short string) (*Link, error)
	GetShort(original string) (*Link, error)
	GetAllByUserID(userID string) ([]Link, error)
	MarkAsDeleted(userID string, short string) error
	Ping() error
}
