package storage

import (
	"context"
	"os"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Инициализация логгера перед запуском тестов
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// Запуск тестов
	code := m.Run()

	// Очистка после тестов
	os.Exit(code)
}

func TestNewFileStorage(t *testing.T) {
	// Создаем временный файл для тестов
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Тестируем создание нового хранилища
	fs := NewFileStorage(tmpFile.Name())

	assert.NotNil(t, fs)
	assert.NotNil(t, fs.memStorage)
	assert.Equal(t, tmpFile.Name(), fs.filePATH)
}

func TestFileStorage_InsertAndGet(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	link := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}

	// Вставляем ссылку
	ctx := context.Background()
	err = fs.Insert(ctx, link)
	require.NoError(t, err)

	// Получаем ссылку по короткому URL
	retrievedLink, err := fs.GetOriginal(link.Short)
	require.NoError(t, err)
	assert.Equal(t, link.Original, retrievedLink.Original)
	assert.Equal(t, link.Short, retrievedLink.Short)
	assert.Equal(t, link.UserID, retrievedLink.UserID)

	// Получаем ссылку по оригинальному URL
	retrievedShortLink, err := fs.GetShort(link.Original)
	require.NoError(t, err)
	assert.Equal(t, link.Original, retrievedShortLink.Original)
	assert.Equal(t, link.Short, retrievedShortLink.Short)
}

func TestFileStorage_InsertLinks(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	links := []*objects.Link{
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
	}

	// Вставляем несколько ссылок
	ctx := context.Background()
	err = fs.InsertLinks(ctx, links)
	require.NoError(t, err)

	// Проверяем что ссылки сохранились
	for _, link := range links {
		retrievedLink, err := fs.GetOriginal(link.Short)
		require.NoError(t, err)
		assert.Equal(t, link.Original, retrievedLink.Original)
	}
}

func TestFileStorage_GetAllByUserID(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	user1Links := []*objects.Link{
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
	}
	user2Link := &objects.Link{
		Short:    "ghi789",
		Original: "https://github.com",
		UserID:   "user2",
	}

	// Вставляем ссылки
	ctx := context.Background()
	err = fs.InsertLinks(ctx, user1Links)
	require.NoError(t, err)
	err = fs.Insert(ctx, user2Link)
	require.NoError(t, err)

	// Получаем ссылки для user1
	links, err := fs.GetAllByUserID("user1")
	require.NoError(t, err)
	assert.Len(t, links, 2)

	// Проверяем содержимое
	for _, link := range links {
		assert.Equal(t, "user1", link.UserID)
		assert.Contains(t, []string{"abc123", "def456"}, link.Short)
		assert.Contains(t, []string{"https://example.com", "https://google.com"}, link.Original)
	}

	// Получаем ссылки для user2
	links, err = fs.GetAllByUserID("user2")
	require.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, user2Link.Short, links[0].Short)
	assert.Equal(t, user2Link.Original, links[0].Original)
}

func TestFileStorage_MarkAsDeleted(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	link := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}

	// Вставляем ссылку
	ctx := context.Background()
	err = fs.Insert(ctx, link)
	require.NoError(t, err)

	// Помечаем как удаленную
	err = fs.MarkAsDeleted("user1", link.Short)
	require.NoError(t, err)

	// Проверяем что ссылка помечена как удаленная
	retrievedLink, err := fs.GetOriginal(link.Short)
	require.NoError(t, err)
	assert.Empty(t, retrievedLink.Original)

	// Попытка пометить как удаленную чужую ссылку
	err = fs.MarkAsDeleted("user2", link.Short)
	assert.Error(t, err)
}

func TestFileStorage_Persistence(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Первое хранилище - вставляем данные
	fs1 := NewFileStorage(tmpFile.Name())
	link := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}
	ctx := context.Background()
	err = fs1.Insert(ctx, link)
	require.NoError(t, err)

	// Второе хранилище - проверяем что данные сохранились
	fs2 := NewFileStorage(tmpFile.Name())
	retrievedLink, err := fs2.GetOriginal(link.Short)
	require.NoError(t, err)
	assert.Equal(t, link.Original, retrievedLink.Original)
}

func TestFileStorage_GetOriginal(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	testLink := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}

	// Вставляем тестовую ссылку
	ctx := context.Background()
	err = fs.Insert(ctx, testLink)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		// Получаем существующую ссылку
		link, err := fs.GetOriginal(testLink.Short)
		require.NoError(t, err)
		assert.Equal(t, testLink.Original, link.Original)
		assert.Equal(t, testLink.Short, link.Short)
		assert.Equal(t, testLink.UserID, link.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		// Пытаемся получить несуществующую ссылку
		_, err := fs.GetOriginal("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, "short URL not found", err.Error())
	})

	t.Run("empty short URL", func(t *testing.T) {
		// Пытаемся получить с пустым short URL
		_, err := fs.GetOriginal("")
		assert.Error(t, err)
		assert.Equal(t, "short URL not found", err.Error())
	})
}

func TestFileStorage_GetShort(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fs := NewFileStorage(tmpFile.Name())

	// Тестовые данные
	testLink := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}

	// Вставляем тестовую ссылку
	ctx := context.Background()
	err = fs.Insert(ctx, testLink)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		// Получаем существующую ссылку
		link, err := fs.GetShort(testLink.Original)
		require.NoError(t, err)
		assert.Equal(t, testLink.Original, link.Original)
		assert.Equal(t, testLink.Short, link.Short)
		assert.Equal(t, testLink.UserID, link.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		// Пытаемся получить несуществующую ссылку
		_, err := fs.GetShort("https://nonexistent.com")
		assert.Error(t, err)
		assert.Equal(t, "original URL not found", err.Error())
	})

	t.Run("empty original URL", func(t *testing.T) {
		// Пытаемся получить с пустым original URL
		_, err := fs.GetShort("")
		assert.Error(t, err)
		assert.Equal(t, "original URL not found", err.Error())
	})
}

func TestSaveAndLoadFromFile(t *testing.T) {
	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Тестовые данные
	link := &objects.Link{
		Short:    "abc123",
		Original: "https://example.com",
		UserID:   "user1",
	}

	// Сохраняем в файл
	err = SaveToFile(link, tmpFile.Name())
	require.NoError(t, err)

	// Загружаем из файла
	data, err := LoadFromFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, link.Original, data[link.Short])
}

func TestFileStorage_Ping(t *testing.T) {
	// Создаем временный файл для теста
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Создаем хранилище с временным файлом
	fs := NewFileStorage(tmpFile.Name())

	// Проверяем что Ping возвращает nil (успешная проверка доступности)
	err = fs.Ping()
	assert.NoError(t, err, "Ping should always return nil for FileStorage")
}
