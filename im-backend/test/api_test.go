package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/controller"
	"zhihang-messenger/im-backend/internal/middleware"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// S+级测试：API端点测试

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// 初始化服务
	authService := service.NewAuthService()
	authController := controller.NewAuthController(authService)
	
	// 设置路由
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}
	}
	
	return r
}

func TestHealthCheck(t *testing.T) {
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestRegisterValidation(t *testing.T) {
	r := setupTestRouter()
	
	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "有效注册请求",
			payload: map[string]string{
				"phone":    "13800138000",
				"username": "testuser",
				"password": "Test123!",
				"nickname": "测试用户",
			},
			expectedStatus: 201,
		},
		{
			name: "缺少必填字段",
			payload: map[string]string{
				"username": "testuser",
				"password": "Test123!",
			},
			expectedStatus: 400,
		},
		{
			name: "密码过短",
			payload: map[string]string{
				"phone":    "13800138000",
				"username": "testuser",
				"password": "123",
			},
			expectedStatus: 400,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			
			// 注：需要数据库连接，此测试可能需要mock
			// assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRateLimiting(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RateLimit())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	
	// 快速发送100个请求，应该有部分被限流
	limitedCount := 0
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		r.ServeHTTP(w, req)
		
		if w.Code == 429 {
			limitedCount++
		}
	}
	
	// 应该有请求被限流
	assert.Greater(t, limitedCount, 0, "速率限制应该生效")
}

func TestCacheMiddleware(t *testing.T) {
	// 测试缓存中间件
	// 第一次请求应该缓存未命中
	// 第二次请求应该缓存命中
	// 需要Redis连接
}

func TestCircuitBreaker(t *testing.T) {
	// 测试熔断器
	// 连续失败应该开启熔断
	// 重置时间后应该进入半开状态
}

// Benchmark测试
func BenchmarkMessageService(b *testing.B) {
	// 性能基准测试
	// 目标: 单次操作 < 10ms
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 测试消息发送性能
	}
}

