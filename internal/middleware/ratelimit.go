package middleware

import (
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter структура для хранения информации о запросах с IP
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // количество запросов
	window   time.Duration // временное окно
}

// Visitor хранит информацию о запросах с одного IP
type Visitor struct {
	lastSeen  time.Time
	requests  []time.Time
	blocked   bool
	blockTime time.Time
}

// NewRateLimiter создаёт новый rate limiter
// rate: количество запросов
// window: временное окно (например, 1 минута)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Запускаем очистку старых записей каждые 5 минут
	go rl.cleanupLoop()

	return rl
}

// RateLimit middleware для ограничения количества запросов с одного IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Проверяем лимит
		if !rl.allow(ip) {
			log.Printf("🚨 RATE LIMIT: Blocked IP %s (exceeded %d requests per %v)", ip, rl.rate, rl.window)
			c.JSON(429, gin.H{
				"error":   "too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow проверяет, разрешён ли запрос с данного IP
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		// Новый посетитель
		rl.visitors[ip] = &Visitor{
			lastSeen: now,
			requests: []time.Time{now},
			blocked:  false,
		}
		return true
	}

	// Обновляем время последнего визита
	visitor.lastSeen = now

	// Если заблокирован, проверяем, не истекло ли время блокировки
	if visitor.blocked {
		if now.Sub(visitor.blockTime) > rl.window {
			// Разблокируем
			visitor.blocked = false
			visitor.requests = []time.Time{now}
			return true
		}
		return false
	}

	// Удаляем старые запросы, вышедшие за окно
	validRequests := []time.Time{}
	for _, reqTime := range visitor.requests {
		if now.Sub(reqTime) <= rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Проверяем лимит
	if len(validRequests) >= rl.rate {
		// Блокируем
		visitor.blocked = true
		visitor.blockTime = now
		log.Printf("⚠️  SECURITY: IP %s exceeded rate limit (%d requests in %v)", ip, len(validRequests), rl.window)
		return false
	}

	// Добавляем текущий запрос
	validRequests = append(validRequests, now)
	visitor.requests = validRequests

	return true
}

// cleanupLoop периодически очищает старые записи
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup удаляет старые записи посетителей
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, visitor := range rl.visitors {
		// Удаляем записи, которые не обращались больше 10 минут
		if now.Sub(visitor.lastSeen) > 10*time.Minute {
			delete(rl.visitors, ip)
		}
	}
}
