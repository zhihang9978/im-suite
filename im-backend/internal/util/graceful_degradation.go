package util

import (
	"context"
	"sync"
	"time"

	"zhihang-messenger/im-backend/config"

	"github.com/sirupsen/logrus"
)

// GracefulDegradation 优雅降级工具（S+可靠性）

// GetWithFallback 从缓存获取，失败时降级到数据库
func GetWithFallback(ctx context.Context, cacheKey string, dbQuery func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// 1. 尝试从Redis缓存获取
	if config.Redis != nil {
		cached, err := config.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			logrus.Debugf("缓存命中: %s", cacheKey)
			return cached, nil
		}
		logrus.Debugf("缓存未命中: %s", cacheKey)
	}

	// 2. Redis不可用或缓存未命中，查询数据库
	result, err := dbQuery()
	if err != nil {
		return nil, err
	}

	// 3. 异步写入缓存（不阻塞响应）
	if config.Redis != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("写入缓存失败（已忽略）: %v", r)
				}
			}()

			// 使用独立的context，避免请求取消影响缓存写入
			bgCtx := context.Background()
			config.Redis.Set(bgCtx, cacheKey, result, ttl)
			logrus.Debugf("缓存已更新: %s", cacheKey)
		}()
	}

	return result, nil
}

// InvalidateCache 删除缓存
func InvalidateCache(ctx context.Context, cacheKey string) {
	if config.Redis != nil {
		config.Redis.Del(ctx, cacheKey)
		logrus.Debugf("缓存已删除: %s", cacheKey)
	}
}

// InvalidateCachePattern 批量删除缓存（匹配模式）
func InvalidateCachePattern(ctx context.Context, pattern string) {
	if config.Redis != nil {
		keys, err := config.Redis.Keys(ctx, pattern).Result()
		if err != nil {
			logrus.Errorf("查询缓存key失败: %v", err)
			return
		}

		if len(keys) > 0 {
			config.Redis.Del(ctx, keys...)
			logrus.Debugf("批量删除缓存: %d个key", len(keys))
		}
	}
}

// RetryWithBackoff 指数退避重试（S+可靠性）
func RetryWithBackoff(maxRetries int, operation func() error) error {
	var err error
	backoff := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			logrus.Warnf("操作失败，%v后重试（第%d/%d次）: %v", backoff, i+1, maxRetries, err)
			time.Sleep(backoff)
			backoff *= 2 // 指数退避
		}
	}

	return err
}

// HealthCheck 健康检查工具
type HealthCheck struct {
	checks map[string]func() bool
	mu     sync.RWMutex
}

// NewHealthCheck 创建健康检查
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		checks: make(map[string]func() bool),
	}
}

// Register 注册健康检查项
func (hc *HealthCheck) Register(name string, check func() bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// Check 执行所有健康检查
func (hc *HealthCheck) Check() map[string]bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]bool)
	for name, check := range hc.checks {
		results[name] = check()
	}
	return results
}

// IsHealthy 检查系统是否健康
func (hc *HealthCheck) IsHealthy() bool {
	results := hc.Check()
	for _, healthy := range results {
		if !healthy {
			return false
		}
	}
	return true
}
