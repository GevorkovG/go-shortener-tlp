package objects

type Link struct {
	//ID       int    `json:"-"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
	UserID   string `json:"-"`
}

type Storage interface {
	Insert(link *Link) error
	InsertLinks(links []*Link) error
	GetOriginal(short string) (*Link, error)
	GetShort(original string) (*Link, error)
	GetAllByUserID(userID string) ([]Link, error)
}
