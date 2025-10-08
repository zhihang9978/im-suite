package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/controller"
	"zhihang-messenger/im-backend/internal/middleware"
	"zhihang-messenger/im-backend/internal/service"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("未找到.env文件，使用系统环境变量")
	}

	// 初始化日志
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// 初始化数据库
	if err := config.InitDB(); err != nil {
		logrus.Fatal("数据库初始化失败:", err)
	}

	// 自动迁移数据库
	if err := config.AutoMigrate(); err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 启动系统监控服务（如果需要）
	// systemMonitorService := service.NewSystemMonitorService()
	// go systemMonitorService.StartMonitoring()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.Security())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "im-backend",
			"version":   "1.3.0",
		})
	})

	// 指标端点
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"uptime": time.Since(time.Now()).String(),
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 初始化服务
		authService := service.NewAuthService()
		userManagementService := service.NewUserManagementService(config.DB)
		messageEncryptionService := service.NewMessageEncryptionService(config.DB)
		messageEnhancementService := service.NewMessageEnhancementService(config.DB)
		contentModerationService := service.NewContentModerationService(config.DB)
		themeService := service.NewThemeService(config.DB)
		groupMgmtService := service.NewGroupManagementService(config.DB)
		fileService := service.NewFileService()
		fileEncryptionService := service.NewFileEncryptionService()

		// 初始化控制器
		authController := controller.NewAuthController(authService)
		userMgmtController := controller.NewUserManagementController(userManagementService)
		messageEncryptionController := controller.NewMessageEncryptionController(messageEncryptionService)
		messageEnhancementController := controller.NewMessageEnhancementController(messageEnhancementService)
		contentModerationController := controller.NewContentModerationController(contentModerationService)
		themeController := controller.NewThemeController(themeService)
		groupMgmtController := controller.NewGroupManagementController(groupMgmtService)
		fileController := controller.NewFileController()

		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/register", authController.Register)
			auth.POST("/logout", authController.Logout)
			auth.POST("/refresh", authController.RefreshToken)
			auth.GET("/validate", authController.ValidateToken)
		}

		// 受保护的路由
		protected := api.Group("/")
		protected.Use(middleware.Auth())
		{
			// 用户管理
			users := protected.Group("/users")
			{
				users.GET("/me", func(c *gin.Context) {
					userID, _ := c.Get("user_id")
					c.JSON(http.StatusOK, gin.H{"user_id": userID})
				})
			}

			// 消息管理
			messages := protected.Group("/messages")
			{
				messages.POST("/", messageController.SendMessage)
				messages.GET("/", messageController.GetMessages)
				messages.GET("/:id", messageController.GetMessage)
				messages.DELETE("/:id", messageController.DeleteMessage)
				messages.POST("/:id/read", messageController.MarkAsRead)
			}

			// 高级消息功能
			advanced := protected.Group("/advanced")
			{
				advanced.POST("/recall", messageAdvancedController.RecallMessage)
				advanced.PUT("/edit", messageAdvancedController.EditMessage)
				advanced.POST("/forward", messageAdvancedController.ForwardMessage)
				advanced.POST("/search", messageAdvancedController.SearchMessages)
				advanced.POST("/schedule", messageAdvancedController.ScheduleMessage)
				advanced.GET("/scheduled", messageAdvancedController.GetScheduledMessages)
				advanced.DELETE("/scheduled/:id", messageAdvancedController.CancelScheduledMessage)
			}

			// 消息加密功能
			encryption := protected.Group("/encryption")
			{
				encryption.POST("/encrypt", messageEncryptionController.EncryptMessage)
				encryption.POST("/decrypt", messageEncryptionController.DecryptMessage)
			}

			// 消息增强功能
			enhancement := protected.Group("/enhancement")
			{
				enhancement.POST("/pin", messageEnhancementController.PinMessage)
				enhancement.POST("/unpin", messageEnhancementController.UnpinMessage)
				enhancement.POST("/mark", messageEnhancementController.MarkMessage)
				enhancement.POST("/unmark", messageEnhancementController.UnmarkMessage)
				enhancement.POST("/share", messageEnhancementController.ShareMessage)
				enhancement.GET("/status/:id", messageEnhancementController.GetMessageStatus)
			}

			// 群组管理
			groups := protected.Group("/groups")
			{
				groups.POST("/", groupMgmtController.CreateGroup)
				groups.GET("/", groupMgmtController.GetGroups)
				groups.GET("/:id", groupMgmtController.GetGroup)
				groups.PUT("/:id", groupMgmtController.UpdateGroup)
				groups.DELETE("/:id", groupMgmtController.DeleteGroup)
				groups.POST("/:id/members", groupMgmtController.AddMember)
				groups.DELETE("/:id/members/:user_id", groupMgmtController.RemoveMember)
				groups.POST("/:id/admins", groupMgmtController.PromoteAdmin)
				groups.DELETE("/:id/admins/:user_id", groupMgmtController.DemoteAdmin)
				groups.POST("/:id/invite", groupMgmtController.CreateInvite)
				groups.GET("/:id/invites", groupMgmtController.GetInvites)
				groups.POST("/join/:invite_code", groupMgmtController.JoinByInvite)
			}

			// 文件管理
			files := protected.Group("/files")
			{
				files.POST("/upload", fileController.UploadFile)
				files.GET("/:id", fileController.GetFile)
				files.GET("/:id/download", fileController.DownloadFile)
				files.GET("/:id/preview", fileController.GetFilePreview)
				files.DELETE("/:id", fileController.DeleteFile)
			}

			// 主题管理
			themes := protected.Group("/themes")
			{
				themes.GET("/", themeController.GetThemes)
				themes.GET("/:id", themeController.GetTheme)
				themes.POST("/", themeController.CreateTheme)
				themes.PUT("/:id", themeController.UpdateTheme)
				themes.DELETE("/:id", themeController.DeleteTheme)
				themes.POST("/:id/apply", themeController.ApplyTheme)
				themes.GET("/user/settings", themeController.GetUserThemeSettings)
				themes.PUT("/user/settings", themeController.UpdateUserThemeSettings)
			}

			// 内容审核
			moderation := protected.Group("/moderation")
			{
				moderation.POST("/report", contentModerationController.ReportContent)
				moderation.GET("/reports", contentModerationController.GetReports)
				moderation.POST("/reports/:id/review", contentModerationController.ReviewReport)
				moderation.POST("/filters", contentModerationController.CreateFilter)
				moderation.GET("/filters", contentModerationController.GetFilters)
				moderation.DELETE("/filters/:id", contentModerationController.DeleteFilter)
				moderation.GET("/statistics", contentModerationController.GetStatistics)
			}

		}
		
	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("服务器启动失败:", err)
		}
	}()

	logrus.Info("服务器启动成功，端口:", port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("服务器强制关闭:", err)
	}

	logrus.Info("服务器已关闭")
}
