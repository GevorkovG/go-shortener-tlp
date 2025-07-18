// Package storage предоставляет реализацию хранилища для сервиса сокращения URL
// с использованием PostgreSQL в качестве бэкенда.
package storage

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// ErrConflict возвращается при попытке вставить дубликат URL
var ErrConflict = errors.New("conflict on inserting new record")

// Link реализует интерфейс Storage для работы с PostgreSQL
type Link struct {
	Store *database.DBStore // Подключение к базе данных
}

// NewLinkStorage создает новый экземпляр хранилища для работы с ссылками
func NewLinkStorage(db *database.DBStore) *Link {
	return &Link{
		Store: db,
	}
}

// IsDeleted проверяет, удален ли URL.
func IsDeleted(link *objects.Link) bool {
	// В зависимости от реализации хранилища, проверяем, удален ли URL.
	// Например, если URL помечен как удаленный, возвращаем true.
	// В данном примере предполагаем, что URL удален, если его Original пуст.
	return link.Original == ""
}

// CreateTable создает таблицу links если она не существует
//
// Параметры:
//   - ctx: контекст выполнения
//
// Возвращает:
//   - error: ошибка при создании таблицы
func (l *Link) CreateTable(ctx context.Context) error {
	if _, err := l.Store.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS links (id SERIAL PRIMARY KEY, short CHAR(20) UNIQUE, original CHAR(255) UNIQUE, userid CHAR(36), is_deleted BOOLEAN DEFAULT FALSE);"); err != nil {
		zap.L().Error("Failed to create table", zap.Error(err))
		return err
	}
	return nil
}

// Insert добавляет новую ссылку в хранилище
//
// Параметры:
//   - ctx: контекст выполнения
//   - link: объект ссылки для добавления
//
// Возвращает:
//   - ErrConflict: если ссылка уже существует
//   - error: другие ошибки базы данных
//
// Логирует:
//   - Информацию о добавляемой ссылке
//   - Конфликты при вставке
//   - Успешное завершение операции
func (l *Link) Insert(ctx context.Context, link *objects.Link) error {
	zap.L().Info("DB Inserting URL",
		zap.String("short", link.Short),
		zap.String("original", link.Original),
		zap.String("userID", link.UserID))

	if err := l.CreateTable(ctx); err != nil {
		return err
	}

	if _, err := l.Store.DB.ExecContext(ctx,
		"INSERT INTO links (short, original, userid) VALUES ($1, $2, $3)",
		link.Short, link.Original, link.UserID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				zap.L().Warn("Conflict on inserting new record",
					zap.String("short", link.Short),
					zap.String("original", link.Original))
				return ErrConflict
			}
		}
		zap.L().Error("Failed to insert link", zap.Error(err))
		return err
	}

	zap.L().Info("DB URL inserted successfully",
		zap.String("short", link.Short),
		zap.String("original", link.Original),
		zap.String("userID", link.UserID))
	return nil
}

// InsertLinks добавляет несколько ссылок в рамках транзакции
//
// Параметры:
//   - ctx: контекст выполнения
//   - links: массив ссылок для добавления
//
// Возвращает:
//   - error: ошибка при выполнении транзакции
//
// Особенности:
//   - Использует транзакцию для атомарности
//   - Прерывается при первой же ошибке
func (l *Link) InsertLinks(ctx context.Context, links []*objects.Link) error {
	if err := l.CreateTable(ctx); err != nil {
		return err
	}

	tx, err := l.Store.DB.Begin()
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO links (short, original, userid) VALUES ($1, $2, $3)")
	if err != nil {
		zap.L().Error("Failed to prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	for _, link := range links {
		if _, err := stmt.Exec(link.Short, link.Original, link.UserID); err != nil {
			zap.L().Error("Failed to insert link", zap.String("short", link.Short), zap.String("original", link.Original), zap.Error(err))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}

// GetOriginal возвращает оригинальный URL по его сокращенной версии
//
// Параметры:
//   - short: сокращенный URL
//
// Возвращает:
//   - *objects.Link: найденная ссылка
//   - error: ошибка при запросе
//
// Логирует:
//   - Ошибки при выполнении запроса
func (l *Link) GetOriginal(short string) (*objects.Link, error) {
	link := &objects.Link{Short: short}

	var (
		original  string
		userID    string
		isDeleted bool
	)

	err := l.Store.DB.QueryRow(
		"SELECT TRIM(original), TRIM(userid), is_deleted FROM links WHERE short = $1",
		strings.TrimSpace(short),
	).Scan(&original, &userID, &isDeleted)

	if err != nil {
		zap.L().Error("Failed to get original URL",
			zap.String("short", short),
			zap.Error(err))
		return nil, err
	}

	link.Original = original
	link.UserID = userID
	link.DeletedFlag = isDeleted

	return link, nil
}

// GetShort возвращает сокращенный URL по оригинальному
//
// Параметры:
//   - original: оригинальный URL
//
// Возвращает:
//   - *objects.Link: найденная ссылка
//   - error: ошибка при запросе
//
// Логирует:
//   - Ошибки при выполнении запроса
func (l *Link) GetShort(original string) (*objects.Link, error) {
	link := &objects.Link{Original: original}

	var (
		short  string
		userID string
	)

	err := l.Store.DB.QueryRow(
		"SELECT TRIM(short), TRIM(userid) FROM links WHERE original = $1",
		strings.TrimSpace(original),
	).Scan(&short, &userID)

	if err != nil {
		zap.L().Error("Failed to get short URL",
			zap.String("original", original),
			zap.Error(err))
		return nil, err
	}

	link.Short = short
	link.UserID = userID

	return link, nil
}

// GetAllByUserID возвращает все ссылки принадлежащие пользователю
//
// Параметры:
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - []objects.Link: массив ссылок пользователя
//   - error: ошибка при запросе
//
// Логирует:
//   - Начало и завершение операции
//   - Ошибки при выполнении запроса
func (l *Link) GetAllByUserID(userID string) ([]objects.Link, error) {
	zap.L().Info("Getting URLs for user", zap.String("userID", userID))
	var links []objects.Link

	zap.L().Info("Querying user URLs from database", zap.String("userID", userID))

	rows, err := l.Store.DB.Query("SELECT original, short FROM links WHERE userid = $1", userID)
	if err != nil {
		zap.L().Error("Failed to query user URLs", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var link objects.Link
		if err := rows.Scan(&link.Original, &link.Short); err != nil {
			zap.L().Error("Failed to scan row", zap.Error(err))
			return nil, err
		}
		links = append(links, link)
	}

	zap.L().Info("User URLs retrieved from database", zap.String("userID", userID), zap.Any("links", links))

	return links, nil
}

// MarkAsDeleted помечает ссылку как удаленную
//
// Параметры:
//   - userID: идентификатор пользователя
//   - short: сокращенный URL для удаления
//
// Возвращает:
//   - error: ошибка при выполнении операции
//
// Особенности:
//   - Проверяет принадлежность ссылки пользователю
//   - Использует транзакцию
//   - Логирует успешное выполнение
func (l *Link) MarkAsDeleted(userID string, short string) error {
	tx, err := l.Store.DB.Begin()
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		log.Printf("Failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE links SET is_deleted = TRUE WHERE short = $1 AND userid = $2")
	if err != nil {
		log.Printf("Failed to prepare statement")
		zap.L().Error("Failed to prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(short, userID); err != nil {
		log.Printf("Failed to mark URL as deleted---Short:%s    userID:%s   error: %s", short, userID, err)
		zap.L().Error("Failed to mark URL as deleted", zap.String("short", short), zap.String("userID", userID), zap.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Successfully deleted---Short:%s    userID:%s", short, userID)
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return err
	}
	log.Printf("Successfully deleted---Short:%s    userID:%s", short, userID)
	zap.L().Info("Successfully marked URL as deleted", zap.String("short", short), zap.String("userID", userID))
	return nil
}

// Ping проверяет соединение с базой данных
//
// Возвращает:
//   - error: ошибка если соединение недоступно
func (l *Link) Ping() error {

	return l.Store.PingDB()
}
