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
	if err := config.InitDatabase(); err != nil {
		logrus.Fatal("数据库初始化失败:", err)
	}

	// 自动迁移数据库
	if err := config.AutoMigrate(); err != nil {
		logrus.Fatal("数据库迁移失败:", err)
	}

	// 初始化Redis
	if err := config.InitRedis(); err != nil {
		logrus.Warn("Redis初始化失败（非致命错误）:", err)
	}

	// 启动系统监控服务
	systemMonitorService := service.NewSystemMonitorService()
	go systemMonitorService.StartMonitoring()

	// 设置Gin模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)

	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimit())
	r.Use(middleware.Security())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "zhihang-messenger-backend",
			"version":   "1.3.1",
		})
	})

	// 指标端点
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// ============================================
		// 初始化所有服务
		// ============================================
		authService := service.NewAuthService()
		userManagementService := service.NewUserManagementService(config.DB)
		messageEncryptionService := service.NewMessageEncryptionService(config.DB)
		messageEnhancementService := service.NewMessageEnhancementService(config.DB)
		contentModerationService := service.NewContentModerationService(config.DB)
		themeService := service.NewThemeService(config.DB)
		groupMgmtService := service.NewGroupManagementService(config.DB)
		chatPermissionService := service.NewChatPermissionService(config.DB)
		chatAnnouncementService := service.NewChatAnnouncementService(config.DB)
		chatStatisticsService := service.NewChatStatisticsService(config.DB)
		chatBackupService := service.NewChatBackupService(config.DB)
		_ = service.NewFileEncryptionService()

		// ============================================
		// 初始化所有控制器
		// ============================================
		authController := controller.NewAuthController(authService)
		userMgmtController := controller.NewUserManagementController(userManagementService)
		messageEncryptionController := controller.NewMessageEncryptionController(messageEncryptionService)
		messageEnhancementController := controller.NewMessageEnhancementController(messageEnhancementService)
		contentModerationController := controller.NewContentModerationController(contentModerationService)
		themeController := controller.NewThemeController(themeService)
		groupMgmtController := controller.NewGroupManagementController(groupMgmtService)
		chatMgmtController := controller.NewChatManagementController(
			chatPermissionService,
			chatAnnouncementService,
			chatStatisticsService,
			chatBackupService,
		)
		fileController := controller.NewFileController()
		superAdminController := controller.NewSuperAdminController()

		// ============================================
		// 认证路由（公开）
		// ============================================
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/register", authController.Register)
			auth.POST("/logout", authController.Logout)
			auth.POST("/refresh", authController.RefreshToken)
			auth.GET("/validate", authController.ValidateToken)
		}

		// ============================================
		// 受保护的路由（需要登录）
		// ============================================
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// ------------------------------------
			// 用户管理
			// ------------------------------------
			users := protected.Group("/users")
			{
				users.POST("/:id/blacklist", userMgmtController.AddToBlacklist)
				users.DELETE("/:id/blacklist/:blacklist_id", userMgmtController.RemoveFromBlacklist)
				users.GET("/:id/blacklist", userMgmtController.GetBlacklist)
				users.GET("/:id/activity", userMgmtController.GetUserActivity)
				users.POST("/:id/restrictions", userMgmtController.SetUserRestriction)
				users.GET("/:id/restrictions", userMgmtController.GetUserRestrictions)
				users.POST("/:id/ban", userMgmtController.BanUser)
				users.POST("/:id/unban", userMgmtController.UnbanUser)
				users.GET("/:id/stats", userMgmtController.GetUserStats)
				users.GET("/suspicious", userMgmtController.GetSuspiciousUsers)
				users.POST("/cleanup-blacklist", userMgmtController.CleanupExpiredBlacklist)
				users.GET("/:id/restrictions/check", userMgmtController.CheckUserRestriction)
				users.POST("/:id/restrictions/increment", userMgmtController.IncrementUserRestriction)
			}

			// ------------------------------------
			// 文件管理
			// ------------------------------------
			files := protected.Group("/files")
			{
				files.POST("/upload", fileController.UploadFile)
				files.POST("/upload/chunk", fileController.UploadChunk)
				files.GET("/:file_id", fileController.GetFile)
				files.GET("/:file_id/download", fileController.DownloadFile)
				files.GET("/:file_id/preview", fileController.GetFilePreview)
				files.GET("/:file_id/versions", fileController.GetFileVersions)
				files.POST("/:file_id/versions", fileController.CreateFileVersion)
				files.DELETE("/:file_id", fileController.DeleteFile)
			}

			// ------------------------------------
			// 消息加密
			// ------------------------------------
			encryption := protected.Group("/encryption")
			{
				encryption.POST("/messages", messageEncryptionController.EncryptMessage)
				encryption.POST("/decrypt", messageEncryptionController.DecryptMessage)
				encryption.GET("/messages/:id/info", messageEncryptionController.GetEncryptedMessageInfo)
				encryption.POST("/messages/:id/self-destruct", messageEncryptionController.SetMessageSelfDestruct)
			}

			// ------------------------------------
			// 消息增强
			// ------------------------------------
			enhancement := protected.Group("/enhancement")
			{
				enhancement.POST("/messages/:id/pin", messageEnhancementController.PinMessage)
				enhancement.DELETE("/messages/:id/pin", messageEnhancementController.UnpinMessage)
				enhancement.POST("/messages/:id/mark", messageEnhancementController.MarkMessage)
				enhancement.DELETE("/messages/:id/mark", messageEnhancementController.UnmarkMessage)
				enhancement.POST("/messages/:id/reply", messageEnhancementController.ReplyToMessage)
				enhancement.POST("/messages/:id/share", messageEnhancementController.ShareMessage)
				enhancement.POST("/messages/:id/status", messageEnhancementController.UpdateMessageStatus)
				enhancement.GET("/messages/:id/reply-chain", messageEnhancementController.GetMessageReplyChain)
				enhancement.GET("/messages/pinned", messageEnhancementController.GetPinnedMessages)
				enhancement.GET("/messages/marked", messageEnhancementController.GetMarkedMessages)
				enhancement.GET("/messages/:id/status", messageEnhancementController.GetMessageStatus)
				enhancement.GET("/messages/:id/share-history", messageEnhancementController.GetMessageShareHistory)
			}

			// ------------------------------------
			// 群组管理
			// ------------------------------------
			groups := protected.Group("/groups")
			{
				groups.POST("/invites", groupMgmtController.CreateInvite)
				groups.POST("/invites/use", groupMgmtController.UseInvite)
				groups.DELETE("/invites/:id", groupMgmtController.RevokeInvite)
				groups.GET("/:id/invites", groupMgmtController.GetChatInvites)
				groups.POST("/:id/join-requests/:request_id/approve", groupMgmtController.ApproveJoinRequest)
				groups.GET("/:id/join-requests/pending", groupMgmtController.GetPendingJoinRequests)
				groups.POST("/:id/members/:user_id/promote", groupMgmtController.PromoteMember)
				groups.POST("/:id/members/:user_id/demote", groupMgmtController.DemoteMember)
				groups.GET("/:id/admins", groupMgmtController.GetChatAdmins)
				groups.GET("/:id/audit-logs", groupMgmtController.GetAuditLogs)
			}

			// ------------------------------------
			// 聊天管理
			// ------------------------------------
			chats := protected.Group("/chats")
			{
				chats.POST("/:id/permissions", chatMgmtController.SetChatPermissions)
				chats.GET("/:id/permissions", chatMgmtController.GetChatPermissions)
				chats.POST("/:id/members/:user_id/mute", chatMgmtController.MuteMember)
				chats.POST("/:id/members/:user_id/unmute", chatMgmtController.UnmuteMember)
				chats.POST("/:id/members/:user_id/ban", chatMgmtController.BanMember)
				chats.POST("/:id/members/:user_id/unban", chatMgmtController.UnbanMember)
				chats.POST("/:id/members/:user_id/promote", chatMgmtController.PromoteMember)
				chats.POST("/:id/members/:user_id/demote", chatMgmtController.DemoteMember)
				chats.GET("/:id/members", chatMgmtController.GetChatMembers)
				chats.POST("/:id/announcements", chatMgmtController.CreateAnnouncement)
				chats.PUT("/:id/announcements/:announcement_id", chatMgmtController.UpdateAnnouncement)
				chats.DELETE("/:id/announcements/:announcement_id", chatMgmtController.DeleteAnnouncement)
				chats.GET("/:id/announcements", chatMgmtController.GetChatAnnouncements)
				chats.GET("/:id/announcements/pinned", chatMgmtController.GetPinnedAnnouncement)
				chats.POST("/:id/announcements/:announcement_id/pin", chatMgmtController.PinAnnouncement)
				chats.DELETE("/:id/announcements/:announcement_id/pin", chatMgmtController.UnpinAnnouncement)
				chats.POST("/:id/rules", chatMgmtController.CreateRule)
				chats.PUT("/:id/rules/:rule_id", chatMgmtController.UpdateRule)
				chats.DELETE("/:id/rules/:rule_id", chatMgmtController.DeleteRule)
				chats.GET("/:id/rules", chatMgmtController.GetChatRules)
				chats.GET("/:id/statistics", chatMgmtController.GetChatStatistics)
				chats.POST("/:id/backup", chatMgmtController.CreateBackup)
				chats.POST("/:id/backup/:backup_id/restore", chatMgmtController.RestoreBackup)
				chats.GET("/:id/backups", chatMgmtController.GetBackupList)
				chats.DELETE("/:id/backups/:backup_id", chatMgmtController.DeleteBackup)
			}

			// ------------------------------------
			// 主题管理
			// ------------------------------------
			themes := protected.Group("/themes")
			{
				themes.POST("/", themeController.CreateTheme)
				themes.GET("/:id", themeController.GetTheme)
				themes.GET("/", themeController.ListThemes)
				themes.POST("/user/:id", themeController.UpdateUserTheme)
				themes.GET("/user/:id", themeController.GetUserTheme)
				themes.POST("/initialize", themeController.InitializeBuiltInThemes)
			}

			// ------------------------------------
			// 内容审核
			// ------------------------------------
			moderation := protected.Group("/moderation")
			{
				moderation.POST("/reports", contentModerationController.ReportContent)
				moderation.GET("/reports/pending", contentModerationController.GetPendingReports)
				moderation.GET("/reports/:id", contentModerationController.GetReportDetail)
				moderation.POST("/reports/:id/handle", contentModerationController.HandleReport)
				moderation.POST("/filters", contentModerationController.CreateFilter)
				moderation.GET("/users/:id/warnings", contentModerationController.GetUserWarnings)
				moderation.GET("/statistics", contentModerationController.GetStatistics)
				moderation.POST("/content/check", contentModerationController.CheckContent)
			}
		}

		// ============================================
		// 超级管理员路由（需要超级管理员权限）
		// ============================================
		superAdmin := api.Group("/super-admin")
		superAdmin.Use(middleware.AuthMiddleware())
		superAdmin.Use(middleware.SuperAdmin())
		{
			superAdminController.SetupRoutes(superAdmin)
		}
	}

	// 启动HTTP服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 在goroutine中启动服务器
	go func() {
		logrus.Infof("🚀 志航密信后端服务启动成功，监听端口: %s", port)
		logrus.Info("📌 可用功能:")
		logrus.Info("  ✅ 用户认证 (/api/auth)")
		logrus.Info("  ✅ 用户管理 (/api/users)")
		logrus.Info("  ✅ 文件管理 (/api/files)")
		logrus.Info("  ✅ 消息加密 (/api/encryption)")
		logrus.Info("  ✅ 消息增强 (/api/enhancement)")
		logrus.Info("  ✅ 群组管理 (/api/groups)")
		logrus.Info("  ✅ 聊天管理 (/api/chats)")
		logrus.Info("  ✅ 主题管理 (/api/themes)")
		logrus.Info("  ✅ 内容审核 (/api/moderation)")
		logrus.Info("  ✅ 超级管理员 (/api/super-admin)")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("服务器强制关闭:", err)
	}

	// 关闭数据库连接
	if config.DB != nil {
		if sqlDB, err := config.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// 关闭Redis连接
	if config.Redis != nil {
		config.Redis.Close()
	}

	logrus.Info("✅ 服务器已安全关闭")
}
