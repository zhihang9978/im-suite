package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"zhihang-messenger/im-backend/config"
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
		{
			messages.POST("/send", func(c *gin.Context) {
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
				
				var req services.SendMessageRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				message, err := messageService.SendMessage(user.ID, req)
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, message)
			})
			
			messages.GET("", func(c *gin.Context) {
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
				
				var req services.GetMessagesRequest
				if err := c.ShouldBindQuery(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				messages, err := messageService.GetMessages(user.ID, req)
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				
				c.JSON(200, messages)
			})
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
