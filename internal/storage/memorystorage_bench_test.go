package storage

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/google/uuid"
)

// Подготовка хранилища
func setupBenchStorage() *InMemoryStorage {
	storage := NewInMemoryStorage()
	return storage
}

func generateID() string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	number := rand.Uint64()
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}

func BenchmarkInsert(b *testing.B) {
	store := NewInMemoryStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Insert(context.Background(), &objects.Link{
			Short:    generateID(),
			Original: "https://example.com/" + generateID(),
			UserID:   uuid.New().String(),
		})
	}
}

func BenchmarkGetOriginal(b *testing.B) {
	storage := setupBenchStorage()
	short := generateID()

	storage.Insert(context.Background(), &objects.Link{
		Short:    short,
		Original: "https://example.com/testGetOriginal",
		UserID:   uuid.New().String(),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetOriginal(short)
	}
}

func BenchmarkGetShort(b *testing.B) {
	storage := setupBenchStorage()
	original := "https://example.com/testGetShort"

	storage.Insert(context.Background(), &objects.Link{
		Short:    generateID(),
		Original: original,
		UserID:   uuid.New().String(),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetShort(original)
	}
}

func BenchmarkGetAllByUserID(b *testing.B) {
	storage := setupBenchStorage()
	testUser := uuid.New().String()

	for i := 0; i < 10; i++ {
		storage.Insert(context.Background(), &objects.Link{
			Short:    generateID(),
			Original: "https://example.com/" + generateID(),
			UserID:   testUser,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetAllByUserID(testUser)
	}
}

func BenchmarkInsertLinks(b *testing.B) {
	storage := setupBenchStorage()
	links := make([]*objects.Link, 100)
	for i := range links {
		links[i] = &objects.Link{
			Short:    generateID(),
			Original: "https://example.com/" + generateID(),
			UserID:   uuid.New().String(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.InsertLinks(context.Background(), links)
	}
}
