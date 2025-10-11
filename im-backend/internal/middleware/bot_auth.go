package middleware

import (
	"net/http"
	"strings"
	"time"

	"zhihang-messenger/im-backend/internal/model"
	"zhihang-messenger/im-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// BotAuthMiddleware 机器人认证中间件
func BotAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取认证头
		authHeader := c.GetHeader("X-Bot-Auth")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少机器人认证信息",
			})
			c.Abort()
			return
		}

		// 解析 API Key 和 Secret
		// 格式: "Bot {api_key}:{api_secret}"
		if !strings.HasPrefix(authHeader, "Bot ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证格式",
			})
			c.Abort()
			return
		}

		credentials := strings.TrimPrefix(authHeader, "Bot ")
		parts := strings.Split(credentials, ":")
		if len(parts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证格式",
			})
			c.Abort()
			return
		}

		apiKey := parts[0]
		apiSecret := parts[1]

		// 验证机器人
		botService := service.NewBotService()
		bot, err := botService.ValidateBotAPIKey(c.Request.Context(), apiKey, apiSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// 检查速率限制
		if err := botService.CheckRateLimit(c.Request.Context(), bot); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// 将机器人信息存储到上下文
		c.Set("bot", &service.BotModel{Bot: bot})
		c.Set("bot_id", bot.ID)
		c.Set("bot_name", bot.Name)

		// 记录请求开始时间
		startTime := time.Now()

		// 继续处理请求
		c.Next()

		// 记录API调用
		duration := time.Since(startTime).Milliseconds()
		statusCode := c.Writer.Status()

		// 异步记录日志（带超时控制）
		botID := bot.ID
		apiPath := c.Request.URL.Path
		method := c.Request.Method
		
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			
			botService.RecordBotAPICall(
				ctx,
				botID,
				apiPath,
				method,
				statusCode,
				duration,
				"", // 请求体（可选）
				"", // 响应体（可选）
				"",
			)
		}()
	}
}

// BotPermissionMiddleware 机器人权限检查中间件
func BotPermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bot, exists := c.Get("bot")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "机器人认证失败",
			})
			c.Abort()
			return
		}

		botModel := bot.(*service.BotModel)
		botService := service.NewBotService()

		// 检查权限
		hasPermission := botService.CheckBotPermission(botModel.Bot, model.BotPermission(requiredPermission))
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "权限不足",
				"required_permission": requiredPermission,
				"message":             "机器人没有执行此操作的权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
