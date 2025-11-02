package telegram

import (
	"sync"
	"time"
)

// MembershipCache кэш для хранения статусов членства в группе
type MembershipCache struct {
	data map[int64]CacheEntry
	mu   sync.RWMutex
}

// CacheEntry запись в кэше с временем истечения
type CacheEntry struct {
	IsMember  bool
	ExpiresAt time.Time
}

// GlobalCache глобальный экземпляр кэша членства
var GlobalCache = &MembershipCache{
	data: make(map[int64]CacheEntry),
}

// Get возвращает закэшированный статус членства
func (c *MembershipCache) Get(userID int64) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[userID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return false, false
	}
	return entry.IsMember, true
}

// Set сохраняет статус членства в кэш
func (c *MembershipCache) Set(userID int64, isMember bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[userID] = CacheEntry{
		IsMember:  isMember,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Clear очищает устаревшие записи (вызывать периодически)
func (c *MembershipCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for userID, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, userID)
		}
	}
}

// StartCleanupRoutine запускает фоновую очистку кэша
func (c *MembershipCache) StartCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.Clear()
		}
	}()
}
