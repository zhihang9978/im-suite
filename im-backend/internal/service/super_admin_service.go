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

// SuperAdminService 超级管理员服务
type SuperAdminService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// UserActivity 用户活动记录
type UserActivity struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	ActivityType string    `json:"activity_type"` // login, logout, send_message, join_group, etc.
	IPAddress    string    `json:"ip_address"`
	Device       string    `json:"device"`
	Location     string    `json:"location"`
	Details      string    `json:"details"`
	Timestamp    time.Time `json:"timestamp"`
}

// OnlineUser 在线用户信息
type OnlineUser struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar"`
	OnlineStatus string    `json:"online_status"` // online, away, busy, invisible
	IPAddress    string    `json:"ip_address"`
	Device       string    `json:"device"`
	Location     string    `json:"location"`
	LoginTime    time.Time `json:"login_time"`
	LastActivity time.Time `json:"last_activity"`
	SessionCount int       `json:"session_count"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	TotalUsers       int64   `json:"total_users"`
	OnlineUsers      int64   `json:"online_users"`
	TotalMessages    int64   `json:"total_messages"`
	TodayMessages    int64   `json:"today_messages"`
	TotalGroups      int64   `json:"total_groups"`
	ActiveGroups     int64   `json:"active_groups"`
	TotalFiles       int64   `json:"total_files"`
	StorageUsed      int64   `json:"storage_used"` // bytes
	BandwidthUsed    int64   `json:"bandwidth_used"` // bytes
	ServerLoad       float64 `json:"server_load"`
	MemoryUsage      float64 `json:"memory_usage"` // percentage
	CPUUsage         float64 `json:"cpu_usage"` // percentage
	DatabaseSize     int64   `json:"database_size"` // bytes
	RedisMemoryUsage int64   `json:"redis_memory_usage"` // bytes
	Timestamp        time.Time `json:"timestamp"`
}

// UserBehaviorAnalysis 用户行为分析
type UserBehaviorAnalysis struct {
	UserID              uint      `json:"user_id"`
	Username            string    `json:"username"`
	MessageCount        int64     `json:"message_count"`
	GroupCount          int64     `json:"group_count"`
	FileUploadCount     int64     `json:"file_upload_count"`
	OnlineTime          int64     `json:"online_time"` // seconds
	LastLoginTime       time.Time `json:"last_login_time"`
	LoginFrequency      int       `json:"login_frequency"` // times per day
	AverageSessionTime  int64     `json:"average_session_time"` // seconds
	RiskScore           float64   `json:"risk_score"` // 0-100
	ViolationCount      int       `json:"violation_count"`
	ReportedCount       int       `json:"reported_count"`
	IsBlacklisted       bool      `json:"is_blacklisted"`
	IsSuspicious        bool      `json:"is_suspicious"`
}

// ContentModeration 内容审核记录
type ContentModeration struct {
	ID            uint      `json:"id"`
	ContentType   string    `json:"content_type"` // message, file, group, user
	ContentID     uint      `json:"content_id"`
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	Content       string    `json:"content"`
	ModerationStatus string `json:"moderation_status"` // pending, approved, rejected, flagged
	ViolationType string    `json:"violation_type"` // spam, harassment, illegal, etc.
	Severity      string    `json:"severity"` // low, medium, high, critical
	ReviewerID    uint      `json:"reviewer_id"`
	ReviewerName  string    `json:"reviewer_name"`
	Action        string    `json:"action"` // none, warning, ban, delete
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
}

// NewSuperAdminService 创建超级管理员服务
func NewSuperAdminService() *SuperAdminService {
	return &SuperAdminService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// GetSystemStats 获取系统统计信息
func (s *SuperAdminService) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		Timestamp: time.Now(),
	}

	// 总用户数
	s.db.Model(&model.User{}).Count(&stats.TotalUsers)

	// 在线用户数
	onlineUsers, _ := s.redis.Keys(s.ctx, "user:online:*").Result()
	stats.OnlineUsers = int64(len(onlineUsers))

	// 总消息数
	s.db.Model(&model.Message{}).Count(&stats.TotalMessages)

	// 今日消息数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.Message{}).
		Where("created_at >= ?", today).
		Count(&stats.TodayMessages)

	// 总群组数
	s.db.Model(&model.Chat{}).Where("type = ?", "group").Count(&stats.TotalGroups)

	// 活跃群组数（最近7天有消息）
	weekAgo := time.Now().AddDate(0, 0, -7)
	s.db.Model(&model.Chat{}).
		Where("type = ? AND last_message_at >= ?", "group", weekAgo).
		Count(&stats.ActiveGroups)

	// 总文件数
	s.db.Model(&model.File{}).Count(&stats.TotalFiles)

	// 存储使用情况
	var storageSum struct {
		Total int64
	}
	s.db.Model(&model.File{}).
		Select("SUM(size) as total").
		Scan(&storageSum)
	stats.StorageUsed = storageSum.Total

	// 数据库大小
	var dbSize struct {
		Size int64
	}
	s.db.Raw(`
		SELECT SUM(data_length + index_length) as size
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&dbSize)
	stats.DatabaseSize = dbSize.Size

	// Redis内存使用
	info, _ := s.redis.Info(s.ctx, "memory").Result()
	// 解析info获取内存使用（简化版本）
	stats.RedisMemoryUsage = 0 // 实际需要解析info字符串

	return stats, nil
}

// GetOnlineUsers 获取在线用户列表
func (s *SuperAdminService) GetOnlineUsers(page, pageSize int) ([]OnlineUser, int64, error) {
	// 获取所有在线用户key
	keys, err := s.redis.Keys(s.ctx, "user:online:*").Result()
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(keys))
	onlineUsers := make([]OnlineUser, 0)

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(keys) {
		end = len(keys)
	}

	for i := start; i < end; i++ {
		key := keys[i]
		// 从Redis获取用户在线信息
		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var onlineUser OnlineUser
		if err := json.Unmarshal([]byte(data), &onlineUser); err != nil {
			continue
		}

		// 获取用户详细信息
		var user model.User
		if err := s.db.First(&user, onlineUser.UserID).Error; err == nil {
			onlineUser.Username = user.Username
			onlineUser.Nickname = user.Nickname
			onlineUser.Avatar = user.Avatar
		}

		onlineUsers = append(onlineUsers, onlineUser)
	}

	return onlineUsers, total, nil
}

// GetUserActivity 获取用户活动记录
func (s *SuperAdminService) GetUserActivity(userID uint, limit int) ([]UserActivity, error) {
	key := fmt.Sprintf("user:activity:%d", userID)
	
	// 从Redis获取最近的活动记录
	records, err := s.redis.LRange(s.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	activities := make([]UserActivity, 0, len(records))
	for _, record := range records {
		var activity UserActivity
		if err := json.Unmarshal([]byte(record), &activity); err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// RecordUserActivity 记录用户活动
func (s *SuperAdminService) RecordUserActivity(activity UserActivity) error {
	key := fmt.Sprintf("user:activity:%d", activity.UserID)
	
	data, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	// 添加到列表前面
	s.redis.LPush(s.ctx, key, data)
	
	// 只保留最近1000条记录
	s.redis.LTrim(s.ctx, key, 0, 999)
	
	// 设置过期时间30天
	s.redis.Expire(s.ctx, key, 30*24*time.Hour)

	return nil
}

// GetUserBehaviorAnalysis 获取用户行为分析
func (s *SuperAdminService) GetUserBehaviorAnalysis(userID uint) (*UserBehaviorAnalysis, error) {
	analysis := &UserBehaviorAnalysis{
		UserID: userID,
	}

	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	analysis.Username = user.Username
	analysis.LastLoginTime = user.LastSeen

	// 消息数量
	s.db.Model(&model.Message{}).
		Where("sender_id = ?", userID).
		Count(&analysis.MessageCount)

	// 群组数量
	s.db.Model(&model.ChatMember{}).
		Where("user_id = ?", userID).
		Count(&analysis.GroupCount)

	// 文件上传数量
	s.db.Model(&model.File{}).
		Where("uploader_id = ?", userID).
		Count(&analysis.FileUploadCount)

	// 违规记录
	s.db.Model(&model.UserWarning{}).
		Where("user_id = ?", userID).
		Count(&analysis.ViolationCount)

	// 被举报次数
	s.db.Model(&model.ContentReport{}).
		Where("reported_user_id = ?", userID).
		Count(&analysis.ReportedCount)

	// 计算风险分数
	analysis.RiskScore = s.calculateRiskScore(analysis)

	// 检查是否在黑名单
	// 实际应该从数据库查询
	analysis.IsBlacklisted = false

	// 判断是否可疑
	analysis.IsSuspicious = analysis.RiskScore > 60

	return analysis, nil
}

// calculateRiskScore 计算风险分数
func (s *SuperAdminService) calculateRiskScore(analysis *UserBehaviorAnalysis) float64 {
	score := 0.0

	// 违规记录权重
	if analysis.ViolationCount > 0 {
		score += float64(analysis.ViolationCount) * 10
	}

	// 被举报次数权重
	if analysis.ReportedCount > 0 {
		score += float64(analysis.ReportedCount) * 5
	}

	// 限制分数在0-100之间
	if score > 100 {
		score = 100
	}

	return score
}

// ForceLogoutUser 强制用户下线
func (s *SuperAdminService) ForceLogoutUser(userID uint, reason string) error {
	// 删除用户在线状态
	key := fmt.Sprintf("user:online:%d", userID)
	s.redis.Del(s.ctx, key)

	// 删除用户会话
	sessionKeys, _ := s.redis.Keys(s.ctx, fmt.Sprintf("session:%d:*", userID)).Result()
	if len(sessionKeys) > 0 {
		s.redis.Del(s.ctx, sessionKeys...)
	}

	// 记录操作日志
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"reason":  reason,
		"action":  "force_logout",
	}).Info("用户被强制下线")

	return nil
}

// BanUser 封禁用户
func (s *SuperAdminService) BanUser(userID uint, duration time.Duration, reason string, adminID uint) error {
	// 更新用户状态
	err := s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_banned":    true,
			"ban_until":    time.Now().Add(duration),
			"ban_reason":   reason,
		}).Error

	if err != nil {
		return err
	}

	// 强制下线
	s.ForceLogoutUser(userID, "user banned: "+reason)

	// 记录操作日志
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"duration": duration,
		"reason":   reason,
		"action":   "ban_user",
	}).Info("用户被封禁")

	return nil
}

// UnbanUser 解封用户
func (s *SuperAdminService) UnbanUser(userID uint, adminID uint) error {
	err := s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_banned":  false,
			"ban_until":  nil,
			"ban_reason": "",
		}).Error

	if err != nil {
		return err
	}

	// 记录操作日志
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"action":   "unban_user",
	}).Info("用户被解封")

	return nil
}

// DeleteUserAccount 删除用户账号
func (s *SuperAdminService) DeleteUserAccount(userID uint, adminID uint, reason string) error {
	// 软删除用户
	err := s.db.Where("id = ?", userID).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// 删除用户相关数据
	// 消息、文件、群组成员等（根据实际需求决定是否删除）

	// 记录操作日志
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"reason":   reason,
		"action":   "delete_account",
	}).Info("用户账号被删除")

	return nil
}

// GetContentModerationQueue 获取内容审核队列
func (s *SuperAdminService) GetContentModerationQueue(status string, page, pageSize int) ([]ContentModeration, int64, error) {
	var records []ContentModeration
	var total int64

	query := s.db.Model(&model.ContentReport{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&records).Error

	return records, total, err
}

// ModerateContent 审核内容
func (s *SuperAdminService) ModerateContent(contentID uint, action, reason string, reviewerID uint) error {
	// 更新审核记录
	now := time.Now()
	err := s.db.Model(&model.ContentReport{}).
		Where("id = ?", contentID).
		Updates(map[string]interface{}{
			"status":      "reviewed",
			"action":      action,
			"reason":      reason,
			"reviewer_id": reviewerID,
			"reviewed_at": now,
		}).Error

	if err != nil {
		return err
	}

	// 根据action执行相应操作
	// delete: 删除内容
	// warn: 警告用户
	// ban: 封禁用户

	logrus.WithFields(logrus.Fields{
		"content_id":  contentID,
		"reviewer_id": reviewerID,
		"action":      action,
		"reason":      reason,
	}).Info("内容审核完成")

	return nil
}

// GetSystemLogs 获取系统日志
func (s *SuperAdminService) GetSystemLogs(logType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 从Redis或数据库获取日志
	key := fmt.Sprintf("system:logs:%s", logType)
	
	total, _ := s.redis.LLen(s.ctx, key).Result()
	
	start := int64((page - 1) * pageSize)
	end := start + int64(pageSize) - 1
	
	logs, err := s.redis.LRange(s.ctx, key, start, end).Result()
	if err != nil {
		return nil, 0, err
	}

	results := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		var logData map[string]interface{}
		if err := json.Unmarshal([]byte(log), &logData); err == nil {
			results = append(results, logData)
		}
	}

	return results, total, nil
}

// BroadcastMessage 广播系统消息
func (s *SuperAdminService) BroadcastMessage(message string, targetType string, targetIDs []uint) error {
	// targetType: all, users, groups
	// 实现系统广播功能

	logrus.WithFields(logrus.Fields{
		"message":     message,
		"target_type": targetType,
		"target_ids":  targetIDs,
		"action":      "broadcast_message",
	}).Info("系统广播消息")

	return nil
}

// GetServerHealth 获取服务器健康状态
func (s *SuperAdminService) GetServerHealth() (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// 检查数据库连接
	sqlDB, err := s.db.DB()
	if err != nil {
		health["database"] = "error"
	} else {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "disconnected"
		} else {
			health["database"] = "healthy"
		}
	}

	// 检查Redis连接
	if err := s.redis.Ping(s.ctx).Err(); err != nil {
		health["redis"] = "disconnected"
	} else {
		health["redis"] = "healthy"
	}

	health["timestamp"] = time.Now()
	health["status"] = "running"

	return health, nil
}
