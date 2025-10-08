package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// LargeGroupService 大群组性能优化服务
type LargeGroupService struct {
	db           *gorm.DB
	redis        *redis.Client
	ctx          context.Context
	cacheTimeout time.Duration
	pageSize     int
}

// GroupCache 群组缓存
type GroupCache struct {
	ChatID        uint      `json:"chat_id"`
	MemberCount   int       `json:"member_count"`
	LastMessageID uint      `json:"last_message_id"`
	LastMessageAt time.Time `json:"last_message_at"`
	CachedAt      time.Time `json:"cached_at"`
}

// MemberCache 成员缓存
type MemberCache struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
	LastSeen time.Time `json:"last_seen"`
}

// MessageCache 消息缓存
type MessageCache struct {
	ID        uint      `json:"id"`
	ChatID    uint      `json:"chat_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLargeGroupService 创建大群组服务
func NewLargeGroupService() *LargeGroupService {
	return &LargeGroupService{
		db:           config.DB,
		redis:        config.Redis,
		ctx:          context.Background(),
		cacheTimeout: 5 * time.Minute,
		pageSize:     50,
	}
}

// GetChatInfo 获取聊天信息（带缓存）
func (s *LargeGroupService) GetChatInfo(chatID uint) (*model.Chat, error) {
	// 尝试从缓存获取
	cached, err := s.getChatFromCache(chatID)
	if err == nil && cached != nil {
		// 转换为完整模型
		chat := &model.Chat{
			ID:            cached.ChatID,
			LastMessageID: &cached.LastMessageID,
			LastMessageAt: &cached.LastMessageAt,
		}
		return chat, nil
	}

	// 从数据库获取
	var chat model.Chat
	err = s.db.First(&chat, chatID).Error
	if err != nil {
		return nil, err
	}

	// 缓存结果
	s.cacheChat(chat)

	return &chat, nil
}

// getChatFromCache 从缓存获取聊天信息
func (s *LargeGroupService) getChatFromCache(chatID uint) (*GroupCache, error) {
	key := fmt.Sprintf("chat:info:%d", chatID)

	result, err := s.redis.Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cache GroupCache
	err = json.Unmarshal([]byte(result), &cache)
	if err != nil {
		return nil, err
	}

	// 检查缓存是否过期
	if time.Since(cache.CachedAt) > s.cacheTimeout {
		return nil, fmt.Errorf("缓存过期")
	}

	return &cache, nil
}

// cacheChat 缓存聊天信息
func (s *LargeGroupService) cacheChat(chat model.Chat) {
	cache := GroupCache{
		ChatID:        chat.ID,
		LastMessageID: 0,
		LastMessageAt: time.Now(),
		CachedAt:      time.Now(),
	}

	if chat.LastMessageID != nil {
		cache.LastMessageID = *chat.LastMessageID
	}

	if chat.LastMessageAt != nil {
		cache.LastMessageAt = *chat.LastMessageAt
	}

	// 获取成员数量
	memberCount, _ := s.getMemberCount(chat.ID)
	cache.MemberCount = memberCount

	key := fmt.Sprintf("chat:info:%d", chat.ID)
	data, _ := json.Marshal(cache)

	s.redis.SetEX(s.ctx, key, data, s.cacheTimeout)
}

// GetMembers 分页获取群成员
func (s *LargeGroupService) GetMembers(chatID uint, page, pageSize int) ([]MemberCache, int64, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 尝试从缓存获取
	cachedMembers, err := s.getMembersFromCache(chatID, page)
	if err == nil && len(cachedMembers) > 0 {
		// 获取总数
		total, _ := s.getMemberCount(chatID)
		return cachedMembers, total, nil
	}

	// 从数据库获取
	var members []model.ChatMember
	var total int64

	// 获取总数
	s.db.Model(&model.ChatMember{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Count(&total)

	// 获取分页数据
	err = s.db.Table("chat_members").
		Select("chat_members.*, users.username, users.nickname, users.avatar, users.last_seen, users.is_active").
		Joins("JOIN users ON chat_members.user_id = users.id").
		Where("chat_members.chat_id = ? AND chat_members.deleted_at IS NULL", chatID).
		Order("chat_members.joined_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&members).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为缓存格式
	memberCaches := make([]MemberCache, 0, len(members))
	for _, member := range members {
		memberCache := MemberCache{
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
		}

		// 获取用户信息
		var user model.User
		if err := s.db.First(&user, member.UserID).Error; err == nil {
			memberCache.Username = user.Username
			memberCache.Nickname = user.Nickname
			memberCache.Avatar = user.Avatar
			memberCache.IsActive = user.IsActive
			memberCache.LastSeen = user.LastSeen
		}

		memberCaches = append(memberCaches, memberCache)
	}

	// 缓存结果
	s.cacheMembers(chatID, page, memberCaches)

	return memberCaches, total, nil
}

// getMembersFromCache 从缓存获取成员列表
func (s *LargeGroupService) getMembersFromCache(chatID uint, page int) ([]MemberCache, error) {
	key := fmt.Sprintf("chat:members:%d:page:%d", chatID, page)

	result, err := s.redis.Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var members []MemberCache
	err = json.Unmarshal([]byte(result), &members)
	if err != nil {
		return nil, err
	}

	return members, nil
}

// cacheMembers 缓存成员列表
func (s *LargeGroupService) cacheMembers(chatID uint, page int, members []MemberCache) {
	key := fmt.Sprintf("chat:members:%d:page:%d", chatID, page)
	data, _ := json.Marshal(members)

	s.redis.SetEX(s.ctx, key, data, s.cacheTimeout)
}

// GetMessages 分页获取消息
func (s *LargeGroupService) GetMessages(chatID uint, page, pageSize int) ([]MessageCache, int64, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 尝试从缓存获取
	cachedMessages, err := s.getMessagesFromCache(chatID, page)
	if err == nil && len(cachedMessages) > 0 {
		// 获取总数
		total, _ := s.getMessageCount(chatID)
		return cachedMessages, total, nil
	}

	// 从数据库获取
	var messages []model.Message
	var total int64

	// 获取总数
	s.db.Model(&model.Message{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Count(&total)

	// 获取分页数据
	err = s.db.Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, 0, err
	}

	// 转换为缓存格式
	messageCaches := make([]MessageCache, 0, len(messages))
	for _, message := range messages {
		messageCaches = append(messageCaches, MessageCache{
			ID:        message.ID,
			ChatID:    message.ChatID,
			UserID:    message.UserID,
			Content:   message.Content,
			Type:      message.Type,
			CreatedAt: message.CreatedAt,
		})
	}

	// 缓存结果
	s.cacheMessages(chatID, page, messageCaches)

	return messageCaches, total, nil
}

// getMessagesFromCache 从缓存获取消息列表
func (s *LargeGroupService) getMessagesFromCache(chatID uint, page int) ([]MessageCache, error) {
	key := fmt.Sprintf("chat:messages:%d:page:%d", chatID, page)

	result, err := s.redis.Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var messages []MessageCache
	err = json.Unmarshal([]byte(result), &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// cacheMessages 缓存消息列表
func (s *LargeGroupService) cacheMessages(chatID uint, page int, messages []MessageCache) {
	key := fmt.Sprintf("chat:messages:%d:page:%d", chatID, page)
	data, _ := json.Marshal(messages)

	s.redis.SetEX(s.ctx, key, data, s.cacheTimeout)
}

// getMemberCount 获取成员数量（带缓存）
func (s *LargeGroupService) getMemberCount(chatID uint) (int, error) {
	key := fmt.Sprintf("chat:member_count:%d", chatID)

	// 尝试从缓存获取
	result, err := s.redis.Get(s.ctx, key).Result()
	if err == nil {
		count, err := strconv.Atoi(result)
		if err == nil {
			return count, nil
		}
	}

	// 从数据库获取
	var count int64
	err = s.db.Model(&model.ChatMember{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// 缓存结果
	s.redis.SetEX(s.ctx, key, strconv.Itoa(int(count)), s.cacheTimeout)

	return int(count), nil
}

// getMessageCount 获取消息数量（带缓存）
func (s *LargeGroupService) getMessageCount(chatID uint) (int64, error) {
	key := fmt.Sprintf("chat:message_count:%d", chatID)

	// 尝试从缓存获取
	result, err := s.redis.Get(s.ctx, key).Result()
	if err == nil {
		count, err := strconv.ParseInt(result, 10, 64)
		if err == nil {
			return count, nil
		}
	}

	// 从数据库获取
	var count int64
	err = s.db.Model(&model.Message{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// 缓存结果
	s.redis.SetEX(s.ctx, key, strconv.FormatInt(count, 10), s.cacheTimeout)

	return count, nil
}

// InvalidateCache 使缓存失效
func (s *LargeGroupService) InvalidateCache(chatID uint) {
	patterns := []string{
		fmt.Sprintf("chat:info:%d", chatID),
		fmt.Sprintf("chat:member_count:%d", chatID),
		fmt.Sprintf("chat:message_count:%d", chatID),
		fmt.Sprintf("chat:members:%d:*", chatID),
		fmt.Sprintf("chat:messages:%d:*", chatID),
	}

	for _, pattern := range patterns {
		keys, err := s.redis.Keys(s.ctx, pattern).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			s.redis.Del(s.ctx, keys...)
		}
	}

	logrus.Infof("已清理聊天 %d 的缓存", chatID)
}

// OptimizeDatabase 数据库优化
func (s *LargeGroupService) OptimizeDatabase() error {
	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_chat_members_chat_id ON chat_members(chat_id)",
		"CREATE INDEX IF NOT EXISTS idx_chat_members_user_id ON chat_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_chat_id_created_at ON messages(chat_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
	}

	for _, index := range indexes {
		err := s.db.Exec(index).Error
		if err != nil {
			logrus.Errorf("创建索引失败: %s, 错误: %v", index, err)
		} else {
			logrus.Infof("创建索引成功: %s", index)
		}
	}

	return nil
}

// CleanupInactiveMembers 清理不活跃成员
func (s *LargeGroupService) CleanupInactiveMembers(chatID uint, inactiveDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -inactiveDays)

	// 标记不活跃成员为删除状态
	result := s.db.Model(&model.ChatMember{}).
		Where("chat_id = ? AND last_seen < ? AND deleted_at IS NULL", chatID, cutoffDate).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return result.Error
	}

	logrus.Infof("清理了聊天 %d 中 %d 个不活跃成员", chatID, result.RowsAffected)

	// 清理相关缓存
	s.InvalidateCache(chatID)

	return nil
}

// GetChatStatistics 获取聊天统计信息
func (s *LargeGroupService) GetChatStatistics(chatID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 成员数量
	memberCount, err := s.getMemberCount(chatID)
	if err != nil {
		return nil, err
	}
	stats["member_count"] = memberCount

	// 消息数量
	messageCount, err := s.getMessageCount(chatID)
	if err != nil {
		return nil, err
	}
	stats["message_count"] = messageCount

	// 今日消息数量
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	s.db.Model(&model.Message{}).
		Where("chat_id = ? AND created_at >= ? AND deleted_at IS NULL", chatID, today).
		Count(&todayCount)
	stats["today_message_count"] = todayCount

	// 活跃成员数量（最近7天）
	weekAgo := time.Now().AddDate(0, 0, -7)
	var activeCount int64
	s.db.Model(&model.User{}).
		Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("chat_members.chat_id = ? AND users.last_seen >= ? AND chat_members.deleted_at IS NULL", chatID, weekAgo).
		Count(&activeCount)
	stats["active_member_count"] = activeCount

	return stats, nil
}
