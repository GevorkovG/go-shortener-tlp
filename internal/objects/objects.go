package objects

import "context"

type Link struct {
	Short       string `json:"short_url"`
	Original    string `json:"original_url"`
	UserID      string `json:"-"`
	DeletedFlag bool   `json:"-"`
}

type Storage interface {
	Insert(ctx context.Context, link *Link) error
	InsertLinks(ctx context.Context, links []*Link) error
	GetOriginal(short string) (*Link, error)
	GetShort(original string) (*Link, error)
	GetAllByUserID(userID string) ([]Link, error)
	MarkAsDeleted(userID string, short string) error
	Ping() error
}
