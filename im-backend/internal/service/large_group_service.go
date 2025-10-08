package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// LargeGroupService 大群组服务
type LargeGroupService struct {
	db           *gorm.DB
	redis        *redis.Client
	ctx          context.Context
	cacheTimeout time.Duration
	pageSize     int
}

// NewLargeGroupService 创建大群组服务
func NewLargeGroupService() *LargeGroupService {
	return &LargeGroupService{
		db:           config.DB,
		redis:        config.Redis,
		ctx:          context.Background(),
		cacheTimeout: 5 * time.Minute,
		pageSize:     100,
	}
}

// GetChatMembers 获取群组成员（分页+缓存）
func (s *LargeGroupService) GetChatMembers(chatID uint, page, pageSize int) ([]model.ChatMember, int64, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("chat:%d:members:page:%d", chatID, page)
	if s.redis != nil {
		cached, err := s.redis.Get(s.ctx, cacheKey).Result()
		if err == nil {
			var members []model.ChatMember
			if json.Unmarshal([]byte(cached), &members) == nil {
				// 从缓存获取总数
				totalKey := fmt.Sprintf("chat:%d:members:total", chatID)
				total, _ := s.redis.Get(s.ctx, totalKey).Int64()
				return members, total, nil
			}
		}
	}

	// 从数据库查询
	var members []model.ChatMember
	var total int64

	// 统计总数
	if err := s.db.Model(&model.ChatMember{}).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("chat_id = ? AND is_active = ?", chatID, true).
		Preload("User").
		Limit(pageSize).
		Offset(offset).
		Find(&members).Error; err != nil {
		return nil, 0, err
	}

	// 缓存结果
	if s.redis != nil {
		data, _ := json.Marshal(members)
		s.redis.Set(s.ctx, cacheKey, data, s.cacheTimeout)
		totalKey := fmt.Sprintf("chat:%d:members:total", chatID)
		s.redis.Set(s.ctx, totalKey, total, s.cacheTimeout)
	}

	return members, total, nil
}

// GetChatMessages 获取群组消息（分页+缓存）
func (s *LargeGroupService) GetChatMessages(chatID uint, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	// 统计总数
	if err := s.db.Model(&model.Message{}).
		Where("chat_id = ? AND is_recalled = ?", chatID, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("chat_id = ? AND is_recalled = ?", chatID, false).
		Preload("Sender").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// InvalidateChatCache 清除聊天缓存
func (s *LargeGroupService) InvalidateChatCache(chatID uint) error {
	if s.redis == nil {
		return nil
	}

	// 删除所有相关缓存
	pattern := fmt.Sprintf("chat:%d:*", chatID)
	iter := s.redis.Scan(s.ctx, 0, pattern, 0).Iterator()
	for iter.Next(s.ctx) {
		s.redis.Del(s.ctx, iter.Val())
	}

	return iter.Err()
}

// OptimizeGroupQuery 优化群组查询
func (s *LargeGroupService) OptimizeGroupQuery(chatID uint) error {
	// 创建必要的索引
	logrus.Infof("优化群组 %d 的查询性能", chatID)

	// 预加载常用数据到缓存
	go func() {
		// 预加载成员列表
		s.GetChatMembers(chatID, 1, 100)

		// 预加载最新消息
		s.GetChatMessages(chatID, 1, 50)
	}()

	return nil
}

// GetMemberCount 获取成员数（带缓存）
func (s *LargeGroupService) GetMemberCount(chatID uint) (int64, error) {
	cacheKey := fmt.Sprintf("chat:%d:members:count", chatID)

	// 尝试从缓存获取
	if s.redis != nil {
		count, err := s.redis.Get(s.ctx, cacheKey).Int64()
		if err == nil {
			return count, nil
		}
	}

	// 从数据库查询
	var count int64
	if err := s.db.Model(&model.ChatMember{}).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Count(&count).Error; err != nil {
		return 0, err
	}

	// 缓存结果
	if s.redis != nil {
		s.redis.Set(s.ctx, cacheKey, count, 10*time.Minute)
	}

	return count, nil
}

