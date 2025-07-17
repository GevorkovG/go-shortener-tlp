package storage

/*
import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/GevorkovG/go-shortener-tlp/internal/database"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type LinkStorageTestSuite struct {
	suite.Suite
	db       *sql.DB
	storage  *Link
	testData []*objects.Link
}

func (s *LinkStorageTestSuite) SetupSuite() {
	// Инициализация логгера
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// Настройки подключения
	dsn := os.Getenv("DATABASE_")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/shortener_test?sslmode=disable"
	}

	// Подключение к БД
	var err error
	s.db, err = sql.Open("pgx", dsn)
	require.NoError(s.T(), err, "Failed to connect to database")

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.db.PingContext(ctx)
	require.NoError(s.T(), err, "Failed to ping database")

	// Создаем таблицу перед всеми тестами
	err = createTestTable(s.db)
	require.NoError(s.T(), err, "Failed to create test table")

	// Инициализация хранилища
	dbStore := &database.DBStore{DB: s.db}
	s.storage = NewLinkStorage(dbStore)

	// Тестовые данные
	s.testData = []*objects.Link{
		{
			Short:    "abc123",
			Original: "https://example.com",
			UserID:   "user1",
		},
		{
			Short:    "def456",
			Original: "https://google.com",
			UserID:   "user1",
		},
		{
			Short:    "ghi789",
			Original: "https://github.com",
			UserID:   "user2",
		},
	}
}

func createTestTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			short VARCHAR(20) UNIQUE,
			original TEXT UNIQUE,
			userid VARCHAR(36),
			is_deleted BOOLEAN DEFAULT FALSE
		)`)
	return err
}

func (s *LinkStorageTestSuite) TearDownSuite() {
	if s.db != nil {
		// Очищаем таблицу после всех тестов
		_, _ = s.db.Exec("DROP TABLE IF EXISTS links")
		s.db.Close()
	}
}

func (s *LinkStorageTestSuite) SetupTest() {
	// Очищаем таблицу перед каждым тестом
	_, err := s.db.Exec("TRUNCATE TABLE links RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err, "Failed to truncate table")
}

func TestLinkStorageSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests")
	}
	suite.Run(t, new(LinkStorageTestSuite))
}

func (s *LinkStorageTestSuite) TestCreateTable() {
	// Удаляем таблицу для этого теста
	_, err := s.db.Exec("DROP TABLE IF EXISTS links")
	require.NoError(s.T(), err)

	// Проверяем создание таблицы
	err = s.storage.CreateTable(context.Background())
	assert.NoError(s.T(), err)

	// Проверяем что таблица существует
	var exists bool
	err = s.db.QueryRow(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'links')",
	).Scan(&exists)
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)
}

func (s *LinkStorageTestSuite) TestInsert() {
	ctx := context.Background()

	// Успешная вставка
	err := s.storage.Insert(ctx, s.testData[0])
	assert.NoError(s.T(), err)

	// Попытка вставить дубликат
	err = s.storage.Insert(ctx, s.testData[0])
	assert.Error(s.T(), err)
	assert.Equal(s.T(), ErrConflict, err)
}
func (s *LinkStorageTestSuite) TestGetOriginal() {
	ctx := context.Background()

	// Подготавливаем данные
	err := s.storage.Insert(ctx, s.testData[0])
	require.NoError(s.T(), err)

	// Успешное получение
	link, err := s.storage.GetOriginal(s.testData[0].Short)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testData[0].Original, link.Original)
	assert.Equal(s.T(), s.testData[0].Short, link.Short)
	assert.Equal(s.T(), s.testData[0].UserID, link.UserID)

	// Несуществующая ссылка
	_, err = s.storage.GetOriginal("nonexistent")
	assert.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, sql.ErrNoRows))
}

func (s *LinkStorageTestSuite) TestGetShort() {
	ctx := context.Background()

	// Подготавливаем данные
	err := s.storage.Insert(ctx, s.testData[0])
	require.NoError(s.T(), err)

	// Успешное получение
	link, err := s.storage.GetShort(s.testData[0].Original)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testData[0].Short, link.Short)
	assert.Equal(s.T(), s.testData[0].Original, link.Original)
	assert.Equal(s.T(), s.testData[0].UserID, link.UserID)

	// Несуществующая ссылка
	_, err = s.storage.GetShort("https://nonexistent.com")
	assert.Error(s.T(), err)
	assert.True(s.T(), errors.Is(err, sql.ErrNoRows))
}
func (s *LinkStorageTestSuite) TestGetAllByUserID() {
	ctx := context.Background()

	// Подготавливаем данные
	err := s.storage.InsertLinks(ctx, s.testData)
	require.NoError(s.T(), err)

	// Получаем ссылки user1
	links, err := s.storage.GetAllByUserID("user1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), links, 2)
}

func (s *LinkStorageTestSuite) TestMarkAsDeleted() {
	ctx := context.Background()

	// Подготавливаем данные
	err := s.storage.Insert(ctx, s.testData[0])
	require.NoError(s.T(), err)

	// Успешное удаление
	err = s.storage.MarkAsDeleted("user1", s.testData[0].Short)
	assert.NoError(s.T(), err)

	// Проверяем что ссылка помечена как удаленная
	var isDeleted bool
	err = s.db.QueryRow(
		"SELECT is_deleted FROM links WHERE short = $1", s.testData[0].Short,
	).Scan(&isDeleted)
	assert.NoError(s.T(), err)
	assert.True(s.T(), isDeleted)
}

func (s *LinkStorageTestSuite) TestPing() {
	// Проверяем что Ping возвращает nil при успешном подключении
	err := s.storage.Ping()
	assert.NoError(s.T(), err)
}
*/
