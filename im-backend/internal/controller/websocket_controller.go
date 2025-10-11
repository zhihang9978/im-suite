package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"zhihang-messenger/im-backend/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应该限制）
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketController WebSocket控制器
type WebSocketController struct {
	authService *service.AuthService
	connections map[uint]*websocket.Conn
	mutex       sync.RWMutex
}

// NewWebSocketController 创建WebSocket控制器
func NewWebSocketController(authService *service.AuthService) *WebSocketController {
	return &WebSocketController{
		authService: authService,
		connections: make(map[uint]*websocket.Conn),
	}
}

// HandleConnection 处理WebSocket连接
func (wsc *WebSocketController) HandleConnection(c *gin.Context) {
	// 从查询参数获取token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "缺少认证token",
		})
		return
	}

	// 验证token
	user, err := wsc.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "token无效",
		})
		return
	}

	// 升级到WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 保存连接
	wsc.mutex.Lock()
	wsc.connections[user.ID] = conn
	wsc.mutex.Unlock()

	// 清理连接
	defer func() {
		wsc.mutex.Lock()
		delete(wsc.connections, user.ID)
		wsc.mutex.Unlock()
	}()

	logrus.Infof("用户 %d (%s) WebSocket连接成功", user.ID, user.Username)

	// 发送欢迎消息
	welcomeMsg := map[string]interface{}{
		"type":    "welcome",
		"message": "WebSocket连接成功",
		"user_id": user.ID,
		"time":    time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 处理消息
	go func() {
		for {
			var msg map[string]interface{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("WebSocket读取错误: %v", err)
				}
				return
			}

			// 处理接收到的消息
			msgType, _ := msg["type"].(string)
			switch msgType {
			case "ping":
				// 响应pong
				conn.WriteJSON(map[string]interface{}{
					"type": "pong",
					"time": time.Now().Unix(),
				})
			case "message":
				// 处理消息（可以调用MessageService）
				logrus.Debugf("收到消息: %v", msg)
				// TODO: 处理实际消息逻辑
			default:
				// 回显消息
				conn.WriteJSON(map[string]interface{}{
					"type":     "echo",
					"original": msg,
					"time":     time.Now().Unix(),
				})
			}
		}
	}()

	// 保持连接并定期ping
	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

// BroadcastMessage 广播消息给所有在线用户
func (wsc *WebSocketController) BroadcastMessage(message interface{}) {
	wsc.mutex.RLock()
	defer wsc.mutex.RUnlock()

	for userID, conn := range wsc.connections {
		if err := conn.WriteJSON(message); err != nil {
			logrus.Errorf("发送消息给用户 %d 失败: %v", userID, err)
		}
	}
}

// SendToUser 发送消息给指定用户
func (wsc *WebSocketController) SendToUser(userID uint, message interface{}) error {
	wsc.mutex.RLock()
	conn, exists := wsc.connections[userID]
	wsc.mutex.RUnlock()

	if !exists {
		return nil // 用户不在线，忽略
	}

	return conn.WriteJSON(message)
}

