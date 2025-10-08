package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	ActivityType string    `json:"activity_type"`
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
	OnlineStatus string    `json:"online_status"`
	IPAddress    string    `json:"ip_address"`
	Device       string    `json:"device"`
	Location     string    `json:"location"`
	LoginTime    time.Time `json:"login_time"`
	LastActivity time.Time `json:"last_activity"`
	SessionCount int       `json:"session_count"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	TotalUsers       int64     `json:"total_users"`
	OnlineUsers      int64     `json:"online_users"`
	TotalChats       int64     `json:"total_chats"`
	TotalMessages    int64     `json:"total_messages"`
	MessagesToday    int64     `json:"messages_today"`
	NewUsersToday    int64     `json:"new_users_today"`
	ActiveUsersToday int64     `json:"active_users_today"`
	TotalStorage     int64     `json:"total_storage"`
	ServerUptime     int64     `json:"server_uptime"`
	LastUpdateTime   time.Time `json:"last_update_time"`
}

// UserAnalysis 用户分析
type UserAnalysis struct {
	UserID         uint      `json:"user_id"`
	Username       string    `json:"username"`
	TotalMessages  int64     `json:"total_messages"`
	TotalChats     int64     `json:"total_chats"`
	LastActiveTime time.Time `json:"last_active_time"`
	RiskScore      float64   `json:"risk_score"`
	ViolationCount int64     `json:"violation_count"`
	ReportedCount  int64     `json:"reported_count"`
	IsBanned       bool      `json:"is_banned"`
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
		LastUpdateTime: time.Now(),
	}

	// 总用户数
	s.db.Model(&model.User{}).Count(&stats.TotalUsers)

	// 在线用户数
	s.db.Model(&model.User{}).Where("online = ?", true).Count(&stats.OnlineUsers)

	// 总聊天数
	s.db.Model(&model.Chat{}).Count(&stats.TotalChats)

	// 总消息数
	s.db.Model(&model.Message{}).Count(&stats.TotalMessages)

	// 今日消息数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.Message{}).Where("created_at >= ?", today).Count(&stats.MessagesToday)

	// 今日新用户数
	s.db.Model(&model.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday)

	// 今日活跃用户数
	s.db.Model(&model.User{}).Where("last_seen >= ?", today).Count(&stats.ActiveUsersToday)

	return stats, nil
}

// GetOnlineUsers 获取在线用户列表
func (s *SuperAdminService) GetOnlineUsers() ([]OnlineUser, error) {
	var users []model.User
	var onlineUsers []OnlineUser

	err := s.db.Where("online = ?", true).Find(&users).Error
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		onlineUser := OnlineUser{
			UserID:       user.ID,
			Username:     user.Username,
			Nickname:     user.Nickname,
			Avatar:       user.Avatar,
			OnlineStatus: "online",
			LoginTime:    user.CreatedAt,
			LastActivity: user.LastSeen,
		}
		onlineUsers = append(onlineUsers, onlineUser)
	}

	return onlineUsers, nil
}

// ForceLogoutUser 强制用户下线
func (s *SuperAdminService) ForceLogoutUser(adminID, userID uint) error {
	// 更新用户在线状态
	err := s.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"online": false,
		}).Error

	if err != nil {
		return err
	}

	// 记录管理员操作
	s.logAdminOperation(adminID, "force_logout", "user", userID, fmt.Sprintf("强制用户 %d 下线", userID))

	// 从Redis删除会话
	if s.redis != nil {
		sessionKey := fmt.Sprintf("user_session:%d", userID)
		s.redis.Del(s.ctx, sessionKey)
	}

	return nil
}

// BanUser 封禁用户
func (s *SuperAdminService) BanUser(adminID, userID uint, duration time.Duration, reason string) error {
	banUntil := time.Now().Add(duration)

	err := s.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_banned":  true,
			"ban_until":  banUntil,
			"ban_reason": reason,
		}).Error

	if err != nil {
		return err
	}

	// 记录操作
	details := fmt.Sprintf("封禁用户 %d，时长: %v，原因: %s", userID, duration, reason)
	s.logAdminOperation(adminID, "ban_user", "user", userID, details)

	return nil
}

// UnbanUser 解封用户
func (s *SuperAdminService) UnbanUser(adminID, userID uint) error {
	err := s.db.Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_banned":  false,
			"ban_until":  nil,
			"ban_reason": nil,
		}).Error

	if err != nil {
		return err
	}

	// 记录操作
	s.logAdminOperation(adminID, "unban_user", "user", userID, fmt.Sprintf("解封用户 %d", userID))

	return nil
}

// DeleteUser 删除用户
func (s *SuperAdminService) DeleteUser(adminID, userID uint) error {
	// 软删除用户
	err := s.db.Delete(&model.User{}, userID).Error
	if err != nil {
		return err
	}

	// 记录操作
	s.logAdminOperation(adminID, "delete_user", "user", userID, fmt.Sprintf("删除用户 %d", userID))

	return nil
}

// GetUserAnalysis 获取用户分析
func (s *SuperAdminService) GetUserAnalysis(userID uint) (*UserAnalysis, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	analysis := &UserAnalysis{
		UserID:         user.ID,
		Username:       user.Username,
		LastActiveTime: user.LastSeen,
		IsBanned:       user.IsBanned,
		RiskScore:      0.0,
	}

	// 统计消息数
	s.db.Model(&model.Message{}).Where("sender_id = ?", userID).Count(&analysis.TotalMessages)

	// 统计参与的聊天数
	s.db.Model(&model.ChatMember{}).Where("user_id = ?", userID).Count(&analysis.TotalChats)

	// 统计违规次数
	s.db.Model(&model.UserWarning{}).Where("user_id = ?", userID).Count(&analysis.ViolationCount)

	// 统计被举报次数
	s.db.Model(&model.ContentReport{}).Where("reported_user_id = ?", userID).Count(&analysis.ReportedCount)

	// 计算风险分数（简化）
	analysis.RiskScore = float64(analysis.ViolationCount)*10 + float64(analysis.ReportedCount)*5

	return analysis, nil
}

// logAdminOperation 记录管理员操作
func (s *SuperAdminService) logAdminOperation(adminID uint, operation, targetType string, targetID uint, details string) {
	log := model.AdminOperationLog{
		AdminID:       adminID,
		OperationType: operation,
		TargetType:    targetType,
		TargetID:      targetID,
		Action:        operation,
		Details:       details,
	}

	if err := s.db.Create(&log).Error; err != nil {
		logrus.Errorf("记录管理员操作失败: %v", err)
	}
}

// GetRecentActivities 获取最近活动
func (s *SuperAdminService) GetRecentActivities(limit int) ([]UserActivity, error) {
	var logs []model.AdminOperationLog
	var activities []UserActivity

	err := s.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		activity := UserActivity{
			UserID:       log.AdminID,
			ActivityType: log.OperationType,
			Details:      log.Details,
			Timestamp:    log.CreatedAt,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// BroadcastMessage 系统广播消息
func (s *SuperAdminService) BroadcastMessage(adminID uint, message string) error {
	// 记录系统广播
	s.logAdminOperation(adminID, "broadcast_message", "system", 0, message)

	// 通过Redis发布消息
	if s.redis != nil {
		payload := map[string]interface{}{
			"type":    "system_broadcast",
			"message": message,
			"time":    time.Now().Unix(),
		}

		data, _ := json.Marshal(payload)
		s.redis.Publish(s.ctx, "system:broadcast", data)
	}

	return nil
}
