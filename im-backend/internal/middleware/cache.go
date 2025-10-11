package middleware

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"

	"github.com/gin-gonic/gin"
)

// CacheMiddleware Redis缓存中间件（S+性能优化）
// 用于GET请求的响应缓存
func CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Redis不可用时跳过缓存（优雅降级）
		if config.Redis == nil {
			c.Next()
			return
		}

		// 生成缓存key（基于路径+查询参数）
		cacheKey := generateCacheKey(c)

		// 尝试从缓存获取
		cached, err := config.Redis.Get(c.Request.Context(), cacheKey).Result()
		if err == nil {
			// 缓存命中
			c.Header("X-Cache", "HIT")
			c.Data(200, "application/json", []byte(cached))
			c.Abort()
			return
		}

		// 缓存未命中，继续处理请求
		c.Header("X-Cache", "MISS")

		// 使用ResponseWriter包装，捕获响应
		blw := &bodyLogWriter{body: []byte{}, ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 只缓存成功的响应
		if c.Writer.Status() == 200 {
			// 异步写入缓存（不阻塞响应）
			// 复制数据避免竞态条件
			cacheData := make([]byte, len(blw.body))
			copy(cacheData, blw.body)
			
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				
				if err := config.Redis.Set(ctx, cacheKey, cacheData, ttl).Err(); err != nil {
					// 缓存写入失败不影响主流程，只记录日志
					// logrus.Debugf("缓存写入失败: %v", err)
				}
			}()
		}
	}
}

// generateCacheKey 生成缓存key
func generateCacheKey(c *gin.Context) string {
	// URL + 查询参数 + 用户ID（如果有）
	key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery

	// 添加用户ID（确保不同用户的缓存隔离）
	if userID, exists := c.Get("user_id"); exists {
		key = fmt.Sprintf("%s:user:%v", key, userID)
	}

	// MD5哈希（缩短key长度）
	hash := md5.Sum([]byte(key))
	return "api:cache:" + hex.EncodeToString(hash[:])
}

// bodyLogWriter 响应体捕获
type bodyLogWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}
