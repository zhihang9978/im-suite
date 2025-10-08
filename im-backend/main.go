package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/controller"
	"zhihang-messenger/im-backend/internal/middleware"
	"zhihang-messenger/im-backend/internal/service"
	"zhihang-messenger/im-backend/services"
)

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源
	},
}

// WebSocket 连接管理器
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
			logrus.Info("客户端连接")
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
				logrus.Info("客户端断开")
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(h.clients, conn)
					conn.Close()
				}
			}
		}
	}
}

var hub = newHub()

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("未找到 .env 文件，使用默认配置")
	}

	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("志航密信后端启动中...")

	// 初始化数据库
	if err := config.InitDatabase(); err != nil {
		logrus.Fatal("数据库初始化失败:", err)
	}

	// 自动迁移数据库表
	if err := config.AutoMigrate(); err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 启动 WebSocket Hub
	go hub.run()

	// 创建 Gin 路由
	r := gin.Default()

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true, "message": "志航密信后端运行正常"})
		})

		// 初始化服务
		authService := services.NewAuthService()
		messageService := services.NewMessageService()
		
		// 初始化高级消息服务
		messageAdvancedService := service.NewMessageAdvancedService(config.DB)
		messageEncryptionService := service.NewMessageEncryptionService(config.DB)
		schedulerService := service.NewSchedulerService(config.DB, messageAdvancedService, messageEncryptionService)
		
		// 初始化群组管理服务
		chatPermissionService := service.NewChatPermissionService(config.DB)
		chatAnnouncementService := service.NewChatAnnouncementService(config.DB)
		chatStatisticsService := service.NewChatStatisticsService(config.DB)
		chatBackupService := service.NewChatBackupService(config.DB)
		
		// 启动定时任务
		ctx := context.Background()
		go schedulerService.StartScheduler(ctx)
		
		// 初始化控制器
		messageAdvancedController := controller.NewMessageAdvancedController(messageAdvancedService)
		messageEncryptionController := controller.NewMessageEncryptionController(messageEncryptionService)
		chatManagementController := controller.NewChatManagementController(
			chatPermissionService,
			chatAnnouncementService,
			chatStatisticsService,
			chatBackupService,
		)

		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				var req services.LoginRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				resp, err := authService.Login(req)
				if err != nil {
					c.JSON(401, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, resp)
			})
			
			auth.POST("/register", func(c *gin.Context) {
				var req services.RegisterRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				resp, err := authService.Register(req)
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, resp)
			})
			
			auth.POST("/refresh", func(c *gin.Context) {
				var req services.RefreshRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				resp, err := authService.RefreshToken(req)
				if err != nil {
					c.JSON(401, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, resp)
			})
			
			auth.POST("/logout", func(c *gin.Context) {
				token := c.GetHeader("Authorization")
				if token == "" {
					c.JSON(400, gin.H{"error": "缺少认证令牌"})
					return
				}
				
				if err := authService.Logout(token); err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, gin.H{"message": "登出成功"})
			})
		}

		// 用户相关
		users := api.Group("/users")
		{
			users.GET("/me", func(c *gin.Context) {
				token := c.GetHeader("Authorization")
				if token == "" {
					c.JSON(401, gin.H{"error": "缺少认证令牌"})
					return
				}
				
				user, err := authService.ValidateToken(token)
				if err != nil {
					c.JSON(401, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, user)
			})
			
			users.PUT("/me", func(c *gin.Context) {
				token := c.GetHeader("Authorization")
				if token == "" {
					c.JSON(401, gin.H{"error": "缺少认证令牌"})
					return
				}
				
				user, err := authService.ValidateToken(token)
				if err != nil {
					c.JSON(401, gin.H{"error": err.Error()})
					return
				}
				
				var updateData struct {
					Nickname string `json:"nickname"`
					Bio      string `json:"bio"`
					Avatar   string `json:"avatar"`
				}
				
				if err := c.ShouldBindJSON(&updateData); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				// 更新用户信息
				config.DB.Model(user).Updates(updateData)
				
				c.JSON(200, gin.H{"message": "用户信息更新成功"})
			})
		}

		// 消息相关
		messages := api.Group("/messages")
		messages.Use(middleware.AuthMiddleware()) // 添加认证中间件
		{
			// 基础消息功能
			messages.POST("/send", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				
				var req services.SendMessageRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				message, err := messageService.SendMessage(userID, req)
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, message)
			})
			
			messages.GET("", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				
				var req services.GetMessagesRequest
				if err := c.ShouldBindQuery(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				messages, err := messageService.GetMessages(userID, req)
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, messages)
			})

			// 高级消息功能
			messages.POST("/recall", messageAdvancedController.RecallMessage)
			messages.POST("/edit", messageAdvancedController.EditMessage)
			messages.POST("/forward", messageAdvancedController.ForwardMessage)
			messages.POST("/schedule", messageAdvancedController.ScheduleMessage)
			messages.GET("/search", messageAdvancedController.SearchMessages)
			messages.GET("/:message_id/edit-history", messageAdvancedController.GetMessageEditHistory)
			messages.DELETE("/:message_id/schedule", messageAdvancedController.CancelScheduledMessage)
			messages.GET("/scheduled", messageAdvancedController.GetScheduledMessages)
			messages.POST("/reply", messageAdvancedController.ReplyToMessage)
		}

		// 消息加密相关
		encryption := api.Group("/encryption")
		encryption.Use(middleware.AuthMiddleware())
		{
			encryption.POST("/encrypt", messageEncryptionController.EncryptMessage)
			encryption.POST("/decrypt", messageEncryptionController.DecryptMessage)
			encryption.GET("/:message_id/info", messageEncryptionController.GetEncryptedMessageInfo)
			encryption.POST("/self-destruct", messageEncryptionController.SetMessageSelfDestruct)
		}

		// 群组管理相关
		chats := api.Group("/chats")
		chats.Use(middleware.AuthMiddleware())
		{
			// 权限管理
			chats.POST("/:chat_id/permissions", chatManagementController.SetChatPermissions)
			chats.GET("/:chat_id/permissions", chatManagementController.GetChatPermissions)
			chats.POST("/:chat_id/members/mute", chatManagementController.MuteMember)
			chats.DELETE("/:chat_id/members/:user_id/mute", chatManagementController.UnmuteMember)
			chats.POST("/:chat_id/members/ban", chatManagementController.BanMember)
			chats.DELETE("/:chat_id/members/:user_id/ban", chatManagementController.UnbanMember)
			chats.POST("/:chat_id/members/promote", chatManagementController.PromoteMember)
			chats.POST("/:chat_id/members/:user_id/demote", chatManagementController.DemoteMember)
			chats.GET("/:chat_id/members", chatManagementController.GetChatMembers)

			// 公告管理
			chats.POST("/:chat_id/announcements", chatManagementController.CreateAnnouncement)
			chats.PUT("/announcements/:announcement_id", chatManagementController.UpdateAnnouncement)
			chats.DELETE("/announcements/:announcement_id", chatManagementController.DeleteAnnouncement)
			chats.GET("/:chat_id/announcements", chatManagementController.GetChatAnnouncements)
			chats.GET("/:chat_id/announcements/pinned", chatManagementController.GetPinnedAnnouncement)
			chats.POST("/announcements/:announcement_id/pin", chatManagementController.PinAnnouncement)
			chats.DELETE("/announcements/:announcement_id/pin", chatManagementController.UnpinAnnouncement)

			// 规则管理
			chats.POST("/:chat_id/rules", chatManagementController.CreateRule)
			chats.PUT("/rules/:rule_id", chatManagementController.UpdateRule)
			chats.DELETE("/rules/:rule_id", chatManagementController.DeleteRule)
			chats.GET("/:chat_id/rules", chatManagementController.GetChatRules)

			// 统计分析
			chats.GET("/:chat_id/statistics", chatManagementController.GetChatStatistics)

			// 备份恢复
			chats.POST("/:chat_id/backups", chatManagementController.CreateBackup)
			chats.POST("/backups/:backup_id/restore", chatManagementController.RestoreBackup)
			chats.GET("/:chat_id/backups", chatManagementController.GetBackupList)
			chats.DELETE("/backups/:backup_id", chatManagementController.DeleteBackup)
		}
	}

	// WebSocket 连接处理
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logrus.Error("WebSocket 升级失败:", err)
			return
		}
		defer conn.Close()

		hub.register <- conn
		defer func() { hub.unregister <- conn }()

		// 处理 WebSocket 消息
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logrus.Error("WebSocket 读取错误:", err)
				break
			}
			
			// 广播消息给所有客户端
			hub.broadcast <- message
		}
	})

	// 获取端口号，默认 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.Info("志航密信后端启动完成，端口:", port)
	logrus.Info("API 文档: http://localhost:" + port + "/api/ping")
	logrus.Info("WebSocket: ws://localhost:" + port + "/ws")
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
