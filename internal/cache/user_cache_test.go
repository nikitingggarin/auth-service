package cache

import (
	"sync"
	"testing"
	"time"

	"auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserCache_ConcurrentAccess(t *testing.T) {
	cache := NewUserCache(5 * time.Minute)

	var wg sync.WaitGroup

	// Тестируем concurrent запись
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			email := string(rune('a'+id)) + "@example.com"
			testUser := &models.User{
				ID:    uuid.New(),
				Email: email,
			}
			cache.Set(email, testUser)
		}(i)
	}

	// Тестируем concurrent чтение
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			email := string(rune('a'+id)) + "@example.com"
			cache.Get(email)
		}(i)
	}

	wg.Wait()

	// Проверяем что все данные сохранились
	for i := 0; i < 10; i++ {
		email := string(rune('a'+i)) + "@example.com"
		assert.NotNil(t, cache.Get(email), "User should be in cache")
	}
}

func TestUserCache_Expiration(t *testing.T) {
	cache := NewUserCache(100 * time.Millisecond) // Короткий TTL
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	cache.Set(user.Email, user)
	assert.NotNil(t, cache.Get(user.Email), "User should be in cache")

	// Ждем истечения TTL
	time.Sleep(150 * time.Millisecond)
	assert.Nil(t, cache.Get(user.Email), "User should be expired")
}

func TestUserCache_Delete(t *testing.T) {
	cache := NewUserCache(5 * time.Minute)
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	cache.Set(user.Email, user)
	assert.NotNil(t, cache.Get(user.Email), "User should be in cache")

	cache.Delete(user.Email)
	assert.Nil(t, cache.Get(user.Email), "User should be deleted from cache")
}
