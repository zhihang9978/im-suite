package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimiterEntry 限制器条目
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// RateLimiter 速率限制器
type RateLimiter struct {
	limiters    map[string]*rateLimiterEntry
	mu          sync.RWMutex
	rate        rate.Limit
	burst       int
	cleanupOnce sync.Once
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}

	// 只启动一次清理goroutine
	rl.cleanupOnce.Do(func() {
		go rl.cleanupRoutine()
	})

	return rl
}

// cleanupRoutine 清理过期限制器的goroutine
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup 清理不活跃的限制器
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiration := 10 * time.Minute

	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastUsed) > expiration {
			delete(rl.limiters, ip)
		}
	}
}

// GetLimiter 获取指定IP的限制器
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	// 先尝试读锁
	rl.mu.RLock()
	entry, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		// 更新最后使用时间
		rl.mu.Lock()
		entry.lastUsed = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	// 创建新的限制器（需要写锁）
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查，避免并发创建
	if entry, exists := rl.limiters[ip]; exists {
		entry.lastUsed = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = &rateLimiterEntry{
		limiter:  limiter,
		lastUsed: time.Now(),
	}

	return limiter
}

// 全局限制器实例（单例模式）
var globalRateLimiter *RateLimiter
var rateLimiterOnce sync.Once

// RateLimit 速率限制中间件
func RateLimit() gin.HandlerFunc {
	// 使用单例模式，只创建一次限制器
	rateLimiterOnce.Do(func() {
		globalRateLimiter = NewRateLimiter(10.0, 20) // 每秒10个请求，突发20个
	})

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !globalRateLimiter.GetLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
