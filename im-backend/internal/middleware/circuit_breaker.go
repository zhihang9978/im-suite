package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CircuitBreaker 熔断器（S+可靠性优化）
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	mu           sync.RWMutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(cb *CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查熔断器状态
		if !cb.Allow() {
			logrus.Warn("熔断器开启，拒绝请求: ", c.Request.URL.Path)
			c.JSON(503, gin.H{
				"error": "服务暂时不可用，请稍后重试",
				"code":  "SERVICE_UNAVAILABLE",
			})
			c.Abort()
			return
		}

		c.Next()

		// 记录请求结果
		if c.Writer.Status() >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}

// Allow 检查是否允许请求通过
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case "open":
		// 熔断开启，检查是否可以进入半开状态
		if now.Sub(cb.lastFailTime) > cb.resetTimeout {
			cb.state = "half-open"
			logrus.Info("熔断器进入半开状态")
			return true
		}
		return false

	case "half-open":
		// 半开状态，允许少量请求测试
		return true

	default: // "closed"
		return true
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == "half-open" {
		// 半开状态下失败，重新开启熔断
		cb.state = "open"
		logrus.Warn("熔断器重新开启")
	} else if cb.failures >= cb.maxFailures {
		// 失败次数超过阈值，开启熔断
		cb.state = "open"
		logrus.Warnf("熔断器开启，失败次数: %d", cb.failures)
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "half-open" {
		// 半开状态下成功，关闭熔断器
		cb.state = "closed"
		cb.failures = 0
		logrus.Info("熔断器关闭，服务恢复正常")
	} else if cb.state == "closed" {
		// 成功请求，重置失败计数
		if cb.failures > 0 {
			cb.failures--
		}
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

