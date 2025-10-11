package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// MessagePushService 消息推送服务
type MessagePushService struct {
	db        *gorm.DB
	redis     *redis.Client
	ctx       context.Context
	pushQueue chan *PushTask
	workers   int
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// PushTask 推送任务
type PushTask struct {
	UserID      uint      `json:"user_id"`
	MessageID   uint      `json:"message_id"`
	MessageType string    `json:"message_type"`
	Content     string    `json:"content"`
	Priority    int       `json:"priority"` // 1-5, 5最高
	CreatedAt   time.Time `json:"created_at"`
}

// NewMessagePushService 创建消息推送服务
func NewMessagePushService() *MessagePushService {
	return &MessagePushService{
		db:        config.DB,
		redis:     config.Redis,
		ctx:       context.Background(),
		pushQueue: make(chan *PushTask, 1000),
		workers:   10,
		stopChan:  make(chan struct{}),
	}
}

// Start 启动推送服务
func (s *MessagePushService) Start() {
	logrus.Info("消息推送服务启动")

	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// Stop 停止推送服务
func (s *MessagePushService) Stop() {
	logrus.Info("正在停止消息推送服务...")
	close(s.stopChan)
	s.wg.Wait()
	logrus.Info("消息推送服务已停止")
}

// worker 推送工作协程
func (s *MessagePushService) worker(id int) {
	defer s.wg.Done()

	logrus.Debugf("推送工作协程 #%d 已启动", id)

	for {
		select {
		case task := <-s.pushQueue:
			s.processPushTask(task)
		case <-s.stopChan:
			logrus.Debugf("推送工作协程 #%d 已停止", id)
			return
		}
	}
}

// processPushTask 处理推送任务
func (s *MessagePushService) processPushTask(task *PushTask) {
	// 检查用户是否在线
	online, err := s.isUserOnline(task.UserID)
	if err != nil {
		logrus.Errorf("检查用户在线状态失败: %v", err)
		return
	}

	if online {
		// 用户在线，通过WebSocket实时推送
		s.pushViaWebSocket(task)
	} else {
		// 用户离线，使用离线推送
		s.pushViaOfflineChannel(task)
	}

	// 记录推送日志
	logrus.Debugf("推送消息 %d 给用户 %d，优先级: %d", task.MessageID, task.UserID, task.Priority)
}

// PushMessage 推送消息（批量）
func (s *MessagePushService) PushMessage(userIDs []uint, message *model.Message) error {
	for _, userID := range userIDs {
		// 跳过发送者自己
		if userID == message.SenderID {
			continue
		}

		task := &PushTask{
			UserID:      userID,
			MessageID:   message.ID,
			MessageType: message.MessageType,
			Content:     message.Content,
			Priority:    3,
			CreatedAt:   time.Now(),
		}

		// 加入推送队列
		select {
		case s.pushQueue <- task:
			// 成功加入队列
		default:
			// 队列已满，记录日志
			logrus.Warnf("推送队列已满，消息 %d 推送失败", message.ID)
		}
	}

	return nil
}

// isUserOnline 检查用户是否在线
func (s *MessagePushService) isUserOnline(userID uint) (bool, error) {
	if s.redis == nil {
		// Redis不可用，从数据库查询
		var user model.User
		if err := s.db.Select("online").First(&user, userID).Error; err != nil {
			return false, err
		}
		return user.Online, nil
	}

	// 从Redis检查
	key := fmt.Sprintf("user:online:%d", userID)
	exists, err := s.redis.Exists(s.ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// pushViaWebSocket 通过WebSocket推送
func (s *MessagePushService) pushViaWebSocket(task *PushTask) {
	if s.redis == nil {
		return
	}

	// 发布到Redis频道，由WebSocket服务器订阅
	channel := fmt.Sprintf("user:push:%d", task.UserID)
	data, _ := json.Marshal(task)
	s.redis.Publish(s.ctx, channel, data)
}

// pushViaOfflineChannel 离线推送
func (s *MessagePushService) pushViaOfflineChannel(task *PushTask) {
	// 这里可以集成第三方推送服务（极光推送、个推等）
	// 简化实现：保存到待推送列表
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf("user:offline:messages:%d", task.UserID)
	data, _ := json.Marshal(task)
	s.redis.RPush(s.ctx, key, data)
	s.redis.Expire(s.ctx, key, 7*24*time.Hour) // 7天过期
}

// BatchPush 批量推送
func (s *MessagePushService) BatchPush(messages []*model.Message, userIDs []uint) error {
	for _, message := range messages {
		if err := s.PushMessage(userIDs, message); err != nil {
			logrus.Errorf("批量推送失败: %v", err)
		}
	}
	return nil
}

// GetPendingPushCount 获取待推送消息数
func (s *MessagePushService) GetPendingPushCount() int {
	return len(s.pushQueue)
}
