package cache

import (
	"log"
	"sync"
	"time"

	"auth-service/internal/models"
)

type CacheItem struct {
	User      *models.User
	ExpiresAt time.Time
}

type UserCache struct {
	mu    sync.RWMutex
	users map[string]*CacheItem
	ttl   time.Duration
}

func NewUserCache(ttl time.Duration) *UserCache {
	return &UserCache{
		users: make(map[string]*CacheItem),
		ttl:   ttl,
	}
}

// Get возвращает пользователя из кеша (concurrent safe)
func (c *UserCache) Get(email string) *models.User {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.users[email]
	if !exists {
		return nil
	}

	if time.Now().After(item.ExpiresAt) {
		return nil
	}

	log.Printf("✅ Cache HIT for email: %s", email)
	return item.User
}

// Set сохраняет пользователя в кеш (concurrent safe)
func (c *UserCache) Set(email string, user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.users[email] = &CacheItem{
		User:      user,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	log.Printf("💾 Cache SET for email: %s", email)
}

// Delete удаляет пользователя из кеша (concurrent safe)
func (c *UserCache) Delete(email string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.users, email)
	log.Printf("🗑️ Cache DELETE for email: %s", email)
}
