package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MessagePushService 消息推送服务
type MessagePushService struct {
	db          *gorm.DB
	redis       *redis.Client
	pushQueue   chan PushTask
	batchSize   int
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
}

// PushTask 推送任务
type PushTask struct {
	MessageID  uint            `json:"message_id"`
	ChatID     uint            `json:"chat_id"`
	SenderID   uint            `json:"sender_id"`
	Content    string          `json:"content"`
	Type       string          `json:"type"`
	Recipients []PushRecipient `json:"recipients"`
	Priority   int             `json:"priority"` // 1-高优先级, 2-普通, 3-低优先级
	Timestamp  time.Time       `json:"timestamp"`
}

// PushRecipient 推送接收者
type PushRecipient struct {
	UserID      uint      `json:"user_id"`
	DeviceToken string    `json:"device_token,omitempty"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
}

// PushBatch 批量推送
type PushBatch struct {
	Tasks       []PushTask `json:"tasks"`
	BatchID     string     `json:"batch_id"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// NewMessagePushService 创建消息推送服务
func NewMessagePushService() *MessagePushService {
	ctx, cancel := context.WithCancel(context.Background())

	return &MessagePushService{
		db:          config.DB,
		redis:       config.Redis,
		pushQueue:   make(chan PushTask, 10000), // 缓冲队列
		batchSize:   100,                        // 批量大小
		workerCount: 5,                          // 工作协程数
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动推送服务
func (s *MessagePushService) Start() {
	logrus.Info("启动消息推送服务")

	// 启动工作协程
	for i := 0; i < s.workerCount; i++ {
		go s.worker(i)
	}

	// 启动批量处理器
	go s.batchProcessor()

	// 启动去重清理器
	go s.deduplicationCleaner()
}

// Stop 停止推送服务
func (s *MessagePushService) Stop() {
	logrus.Info("停止消息推送服务")
	s.cancel()
	close(s.pushQueue)
}

// worker 工作协程
func (s *MessagePushService) worker(id int) {
	logrus.Infof("推送工作协程 %d 启动", id)

	for {
		select {
		case task := <-s.pushQueue:
			s.processPushTask(task)
		case <-s.ctx.Done():
			logrus.Infof("推送工作协程 %d 停止", id)
			return
		}
	}
}

// processPushTask 处理推送任务
func (s *MessagePushService) processPushTask(task PushTask) {
	start := time.Now()

	// 检查消息去重
	if s.isDuplicate(task) {
		logrus.Debugf("跳过重复消息: %d", task.MessageID)
		return
	}

	// 批量处理推送
	batch := s.createBatch(task)
	if batch != nil {
		s.sendBatchPush(batch)
	}

	// 记录推送去重
	s.markAsProcessed(task)

	// 记录性能指标
	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"message_id": task.MessageID,
		"chat_id":    task.ChatID,
		"duration":   duration.Milliseconds(),
		"recipients": len(task.Recipients),
	}).Info("消息推送完成")
}

// isDuplicate 检查是否重复消息
func (s *MessagePushService) isDuplicate(task PushTask) bool {
	key := fmt.Sprintf("push:dedup:%d:%d", task.MessageID, task.ChatID)

	// 检查Redis中是否存在
	exists, err := s.redis.Exists(s.ctx, key).Result()
	if err != nil {
		logrus.Errorf("检查推送去重失败: %v", err)
		return false
	}

	return exists > 0
}

// markAsProcessed 标记为已处理
func (s *MessagePushService) markAsProcessed(task PushTask) {
	key := fmt.Sprintf("push:dedup:%d:%d", task.MessageID, task.ChatID)

	// 设置5分钟过期时间
	err := s.redis.SetEX(s.ctx, key, "1", 5*time.Minute).Err()
	if err != nil {
		logrus.Errorf("标记推送去重失败: %v", err)
	}
}

// createBatch 创建批量推送
func (s *MessagePushService) createBatch(task PushTask) *PushBatch {
	// 按优先级分组
	highPriority := make([]PushTask, 0)
	normalPriority := make([]PushTask, 0)
	lowPriority := make([]PushTask, 0)

	switch task.Priority {
	case 1:
		highPriority = append(highPriority, task)
	case 2:
		normalPriority = append(normalPriority, task)
	case 3:
		lowPriority = append(lowPriority, task)
	}

	// 优先处理高优先级消息
	if len(highPriority) > 0 {
		return &PushBatch{
			Tasks:     highPriority,
			BatchID:   fmt.Sprintf("high_%d", time.Now().UnixNano()),
			CreatedAt: time.Now(),
		}
	}

	// 普通优先级消息
	if len(normalPriority) > 0 {
		return &PushBatch{
			Tasks:     normalPriority,
			BatchID:   fmt.Sprintf("normal_%d", time.Now().UnixNano()),
			CreatedAt: time.Now(),
		}
	}

	// 低优先级消息
	if len(lowPriority) > 0 {
		return &PushBatch{
			Tasks:     lowPriority,
			BatchID:   fmt.Sprintf("low_%d", time.Now().UnixNano()),
			CreatedAt: time.Now(),
		}
	}

	return nil
}

// batchProcessor 批量处理器
func (s *MessagePushService) batchProcessor() {
	ticker := time.NewTicker(100 * time.Millisecond) // 100ms批量处理间隔
	defer ticker.Stop()

	batch := make([]PushTask, 0, s.batchSize)

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = make([]PushTask, 0, s.batchSize)
			}
		case task := <-s.pushQueue:
			batch = append(batch, task)
			if len(batch) >= s.batchSize {
				s.processBatch(batch)
				batch = make([]PushTask, 0, s.batchSize)
			}
		case <-s.ctx.Done():
			// 处理剩余任务
			if len(batch) > 0 {
				s.processBatch(batch)
			}
			return
		}
	}
}

// processBatch 处理批量推送
func (s *MessagePushService) processBatch(tasks []PushTask) {
	if len(tasks) == 0 {
		return
	}

	start := time.Now()

	// 按聊天分组
	chatGroups := make(map[uint][]PushTask)
	for _, task := range tasks {
		chatGroups[task.ChatID] = append(chatGroups[task.ChatID], task)
	}

	// 并发处理各个聊天组
	var wg sync.WaitGroup
	for chatID, chatTasks := range chatGroups {
		wg.Add(1)
		go func(cID uint, cTasks []PushTask) {
			defer wg.Done()
			s.processChatBatch(cID, cTasks)
		}(chatID, chatTasks)
	}

	wg.Wait()

	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"batch_size": len(tasks),
		"duration":   duration.Milliseconds(),
		"chats":      len(chatGroups),
	}).Info("批量推送处理完成")
}

// processChatBatch 处理单个聊天的批量推送
func (s *MessagePushService) processChatBatch(chatID uint, tasks []PushTask) {
	// 获取聊天成员
	members, err := s.getChatMembers(chatID)
	if err != nil {
		logrus.Errorf("获取聊天成员失败: %v", err)
		return
	}

	// 为每个任务设置接收者
	for i := range tasks {
		tasks[i].Recipients = s.filterRecipients(members, tasks[i].SenderID)
	}

	// 发送推送
	for _, task := range tasks {
		s.sendPushNotification(task)
	}
}

// getChatMembers 获取聊天成员
func (s *MessagePushService) getChatMembers(chatID uint) ([]model.User, error) {
	var members []model.User

	err := s.db.Table("users").
		Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("chat_members.chat_id = ? AND chat_members.deleted_at IS NULL", chatID).
		Find(&members).Error

	return members, err
}

// filterRecipients 过滤接收者
func (s *MessagePushService) filterRecipients(members []model.User, senderID uint) []PushRecipient {
	recipients := make([]PushRecipient, 0, len(members))

	for _, member := range members {
		// 排除发送者
		if member.ID == senderID {
			continue
		}

		// 检查用户是否在线
		isOnline := s.isUserOnline(member.ID)

		recipient := PushRecipient{
			UserID:   member.ID,
			IsOnline: isOnline,
			LastSeen: member.LastSeen,
		}

		// 如果用户不在线，获取设备令牌
		if !isOnline {
			recipient.DeviceToken = s.getDeviceToken(member.ID)
		}

		recipients = append(recipients, recipient)
	}

	return recipients
}

// isUserOnline 检查用户是否在线
func (s *MessagePushService) isUserOnline(userID uint) bool {
	key := fmt.Sprintf("user:online:%d", userID)
	exists, err := s.redis.Exists(s.ctx, key).Result()
	return err == nil && exists > 0
}

// getDeviceToken 获取设备令牌
func (s *MessagePushService) getDeviceToken(userID uint) string {
	key := fmt.Sprintf("device:token:%d", userID)
	token, err := s.redis.Get(s.ctx, key).Result()
	if err != nil {
		return ""
	}
	return token
}

// sendPushNotification 发送推送通知
func (s *MessagePushService) sendPushNotification(task PushTask) {
	for _, recipient := range task.Recipients {
		if recipient.IsOnline {
			// 在线用户通过WebSocket推送
			s.sendWebSocketPush(recipient.UserID, task)
		} else if recipient.DeviceToken != "" {
			// 离线用户通过推送服务
			s.sendMobilePush(recipient.DeviceToken, task)
		}
	}
}

// sendWebSocketPush 发送WebSocket推送
func (s *MessagePushService) sendWebSocketPush(userID uint, task PushTask) {
	// 这里应该通过WebSocket连接发送
	// 实际实现需要与WebSocket Hub集成
	logrus.Debugf("发送WebSocket推送给用户 %d", userID)
}

// sendMobilePush 发送移动端推送
func (s *MessagePushService) sendMobilePush(deviceToken string, task PushTask) {
	// 这里应该调用第三方推送服务（如Firebase、APNs等）
	logrus.Debugf("发送移动端推送到设备 %s", deviceToken)
}

// sendBatchPush 发送批量推送
func (s *MessagePushService) sendBatchPush(batch *PushBatch) {
	if batch == nil || len(batch.Tasks) == 0 {
		return
	}

	now := time.Now()
	batch.ProcessedAt = &now

	// 记录批量推送统计
	s.recordBatchStats(batch)

	// 实际发送推送
	for _, task := range batch.Tasks {
		s.sendPushNotification(task)
	}
}

// recordBatchStats 记录批量推送统计
func (s *MessagePushService) recordBatchStats(batch *PushBatch) {
	stats := map[string]interface{}{
		"batch_id":     batch.BatchID,
		"task_count":   len(batch.Tasks),
		"created_at":   batch.CreatedAt,
		"processed_at": batch.ProcessedAt,
	}

	statsJSON, _ := json.Marshal(stats)
	s.redis.LPush(s.ctx, "push:stats", statsJSON)

	// 只保留最近1000条统计记录
	s.redis.LTrim(s.ctx, "push:stats", 0, 999)
}

// deduplicationCleaner 去重清理器
func (s *MessagePushService) deduplicationCleaner() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanExpiredDeduplication()
		case <-s.ctx.Done():
			return
		}
	}
}

// cleanExpiredDeduplication 清理过期的去重记录
func (s *MessagePushService) cleanExpiredDeduplication() {
	pattern := "push:dedup:*"
	keys, err := s.redis.Keys(s.ctx, pattern).Result()
	if err != nil {
		logrus.Errorf("获取去重键失败: %v", err)
		return
	}

	if len(keys) > 0 {
		s.redis.Del(s.ctx, keys...)
		logrus.Debugf("清理了 %d 个过期去重记录", len(keys))
	}
}

// QueuePush 队列推送任务
func (s *MessagePushService) QueuePush(messageID, chatID, senderID uint, content, msgType string, priority int) error {
	task := PushTask{
		MessageID: messageID,
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   content,
		Type:      msgType,
		Priority:  priority,
		Timestamp: time.Now(),
	}

	select {
	case s.pushQueue <- task:
		return nil
	default:
		return fmt.Errorf("推送队列已满")
	}
}

// GetPushStats 获取推送统计
func (s *MessagePushService) GetPushStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 队列长度
	queueLen := len(s.pushQueue)
	stats["queue_length"] = queueLen

	// 最近推送统计
	recentStats, err := s.redis.LRange(s.ctx, "push:stats", 0, 99).Result()
	if err != nil {
		return nil, err
	}

	stats["recent_batches"] = len(recentStats)
	stats["worker_count"] = s.workerCount
	stats["batch_size"] = s.batchSize

	return stats, nil
}
