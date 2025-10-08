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

// SuperAdminService 瓒呯骇绠＄悊鍛樻湇鍔?type SuperAdminService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// UserActivity 鐢ㄦ埛娲诲姩璁板綍
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

// OnlineUser 鍦ㄧ嚎鐢ㄦ埛淇℃伅
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

// SystemStats 绯荤粺缁熻淇℃伅
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

// UserBehaviorAnalysis 鐢ㄦ埛琛屼负鍒嗘瀽
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

// ContentModeration 鍐呭瀹℃牳璁板綍
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

// NewSuperAdminService 鍒涘缓瓒呯骇绠＄悊鍛樻湇鍔?func NewSuperAdminService() *SuperAdminService {
	return &SuperAdminService{
		db:    config.DB,
		redis: config.Redis,
		ctx:   context.Background(),
	}
}

// GetSystemStats 鑾峰彇绯荤粺缁熻淇℃伅
func (s *SuperAdminService) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		Timestamp: time.Now(),
	}

	// 鎬荤敤鎴锋暟
	s.db.Model(&model.User{}).Count(&stats.TotalUsers)

	// 鍦ㄧ嚎鐢ㄦ埛鏁?	onlineUsers, _ := s.redis.Keys(s.ctx, "user:online:*").Result()
	stats.OnlineUsers = int64(len(onlineUsers))

	// 鎬绘秷鎭暟
	s.db.Model(&model.Message{}).Count(&stats.TotalMessages)

	// 浠婃棩娑堟伅鏁?	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&model.Message{}).
		Where("created_at >= ?", today).
		Count(&stats.TodayMessages)

	// 鎬荤兢缁勬暟
	s.db.Model(&model.Chat{}).Where("type = ?", "group").Count(&stats.TotalGroups)

	// 娲昏穬缇ょ粍鏁帮紙鏈€杩?澶╂湁娑堟伅锛?	weekAgo := time.Now().AddDate(0, 0, -7)
	s.db.Model(&model.Chat{}).
		Where("type = ? AND last_message_at >= ?", "group", weekAgo).
		Count(&stats.ActiveGroups)

	// 鎬绘枃浠舵暟
	s.db.Model(&model.File{}).Count(&stats.TotalFiles)

	// 瀛樺偍浣跨敤鎯呭喌
	var storageSum struct {
		Total int64
	}
	s.db.Model(&model.File{}).
		Select("SUM(size) as total").
		Scan(&storageSum)
	stats.StorageUsed = storageSum.Total

	// 鏁版嵁搴撳ぇ灏?	var dbSize struct {
		Size int64
	}
	s.db.Raw(`
		SELECT SUM(data_length + index_length) as size
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&dbSize)
	stats.DatabaseSize = dbSize.Size

	// Redis鍐呭瓨浣跨敤
	info, _ := s.redis.Info(s.ctx, "memory").Result()
	// 瑙ｆ瀽info鑾峰彇鍐呭瓨浣跨敤锛堢畝鍖栫増鏈級
	stats.RedisMemoryUsage = 0 // 瀹為檯闇€瑕佽В鏋恑nfo瀛楃涓?
	return stats, nil
}

// GetOnlineUsers 鑾峰彇鍦ㄧ嚎鐢ㄦ埛鍒楄〃
func (s *SuperAdminService) GetOnlineUsers(page, pageSize int) ([]OnlineUser, int64, error) {
	// 鑾峰彇鎵€鏈夊湪绾跨敤鎴穔ey
	keys, err := s.redis.Keys(s.ctx, "user:online:*").Result()
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(keys))
	onlineUsers := make([]OnlineUser, 0)

	// 鍒嗛〉
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(keys) {
		end = len(keys)
	}

	for i := start; i < end; i++ {
		key := keys[i]
		// 浠嶳edis鑾峰彇鐢ㄦ埛鍦ㄧ嚎淇℃伅
		data, err := s.redis.Get(s.ctx, key).Result()
		if err != nil {
			continue
		}

		var onlineUser OnlineUser
		if err := json.Unmarshal([]byte(data), &onlineUser); err != nil {
			continue
		}

		// 鑾峰彇鐢ㄦ埛璇︾粏淇℃伅
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

// GetUserActivity 鑾峰彇鐢ㄦ埛娲诲姩璁板綍
func (s *SuperAdminService) GetUserActivity(userID uint, limit int) ([]UserActivity, error) {
	key := fmt.Sprintf("user:activity:%d", userID)
	
	// 浠嶳edis鑾峰彇鏈€杩戠殑娲诲姩璁板綍
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

// RecordUserActivity 璁板綍鐢ㄦ埛娲诲姩
func (s *SuperAdminService) RecordUserActivity(activity UserActivity) error {
	key := fmt.Sprintf("user:activity:%d", activity.UserID)
	
	data, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	// 娣诲姞鍒板垪琛ㄥ墠闈?	s.redis.LPush(s.ctx, key, data)
	
	// 鍙繚鐣欐渶杩?000鏉¤褰?	s.redis.LTrim(s.ctx, key, 0, 999)
	
	// 璁剧疆杩囨湡鏃堕棿30澶?	s.redis.Expire(s.ctx, key, 30*24*time.Hour)

	return nil
}

// GetUserBehaviorAnalysis 鑾峰彇鐢ㄦ埛琛屼负鍒嗘瀽
func (s *SuperAdminService) GetUserBehaviorAnalysis(userID uint) (*UserBehaviorAnalysis, error) {
	analysis := &UserBehaviorAnalysis{
		UserID: userID,
	}

	// 鑾峰彇鐢ㄦ埛淇℃伅
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	analysis.Username = user.Username
	analysis.LastLoginTime = user.LastSeen

	// 娑堟伅鏁伴噺
	s.db.Model(&model.Message{}).
		Where("sender_id = ?", userID).
		Count(&analysis.MessageCount)

	// 缇ょ粍鏁伴噺
	s.db.Model(&model.ChatMember{}).
		Where("user_id = ?", userID).
		Count(&analysis.GroupCount)

	// 鏂囦欢涓婁紶鏁伴噺
	s.db.Model(&model.File{}).
		Where("uploader_id = ?", userID).
		Count(&analysis.FileUploadCount)

	// 杩濊璁板綍
	s.db.Model(&model.UserWarning{}).
		Where("user_id = ?", userID).
		Count(&analysis.ViolationCount)

	// 琚妇鎶ユ鏁?	s.db.Model(&model.ContentReport{}).
		Where("reported_user_id = ?", userID).
		Count(&analysis.ReportedCount)

	// 璁＄畻椋庨櫓鍒嗘暟
	analysis.RiskScore = s.calculateRiskScore(analysis)

	// 妫€鏌ユ槸鍚﹀湪榛戝悕鍗?	// 瀹為檯搴旇浠庢暟鎹簱鏌ヨ
	analysis.IsBlacklisted = false

	// 鍒ゆ柇鏄惁鍙枒
	analysis.IsSuspicious = analysis.RiskScore > 60

	return analysis, nil
}

// calculateRiskScore 璁＄畻椋庨櫓鍒嗘暟
func (s *SuperAdminService) calculateRiskScore(analysis *UserBehaviorAnalysis) float64 {
	score := 0.0

	// 杩濊璁板綍鏉冮噸
	if analysis.ViolationCount > 0 {
		score += float64(analysis.ViolationCount) * 10
	}

	// 琚妇鎶ユ鏁版潈閲?	if analysis.ReportedCount > 0 {
		score += float64(analysis.ReportedCount) * 5
	}

	// 闄愬埗鍒嗘暟鍦?-100涔嬮棿
	if score > 100 {
		score = 100
	}

	return score
}

// ForceLogoutUser 寮哄埗鐢ㄦ埛涓嬬嚎
func (s *SuperAdminService) ForceLogoutUser(userID uint, reason string) error {
	// 鍒犻櫎鐢ㄦ埛鍦ㄧ嚎鐘舵€?	key := fmt.Sprintf("user:online:%d", userID)
	s.redis.Del(s.ctx, key)

	// 鍒犻櫎鐢ㄦ埛浼氳瘽
	sessionKeys, _ := s.redis.Keys(s.ctx, fmt.Sprintf("session:%d:*", userID)).Result()
	if len(sessionKeys) > 0 {
		s.redis.Del(s.ctx, sessionKeys...)
	}

	// 璁板綍鎿嶄綔鏃ュ織
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"reason":  reason,
		"action":  "force_logout",
	}).Info("鐢ㄦ埛琚己鍒朵笅绾?)

	return nil
}

// BanUser 灏佺鐢ㄦ埛
func (s *SuperAdminService) BanUser(userID uint, duration time.Duration, reason string, adminID uint) error {
	// 鏇存柊鐢ㄦ埛鐘舵€?	err := s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_banned":    true,
			"ban_until":    time.Now().Add(duration),
			"ban_reason":   reason,
		}).Error

	if err != nil {
		return err
	}

	// 寮哄埗涓嬬嚎
	s.ForceLogoutUser(userID, "user banned: "+reason)

	// 璁板綍鎿嶄綔鏃ュ織
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"duration": duration,
		"reason":   reason,
		"action":   "ban_user",
	}).Info("鐢ㄦ埛琚皝绂?)

	return nil
}

// UnbanUser 瑙ｅ皝鐢ㄦ埛
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

	// 璁板綍鎿嶄綔鏃ュ織
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"action":   "unban_user",
	}).Info("鐢ㄦ埛琚В灏?)

	return nil
}

// DeleteUserAccount 鍒犻櫎鐢ㄦ埛璐﹀彿
func (s *SuperAdminService) DeleteUserAccount(userID uint, adminID uint, reason string) error {
	// 杞垹闄ょ敤鎴?	err := s.db.Where("id = ?", userID).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// 鍒犻櫎鐢ㄦ埛鐩稿叧鏁版嵁
	// 娑堟伅銆佹枃浠躲€佺兢缁勬垚鍛樼瓑锛堟牴鎹疄闄呴渶姹傚喅瀹氭槸鍚﹀垹闄わ級

	// 璁板綍鎿嶄綔鏃ュ織
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"reason":   reason,
		"action":   "delete_account",
	}).Info("鐢ㄦ埛璐﹀彿琚垹闄?)

	return nil
}

// GetContentModerationQueue 鑾峰彇鍐呭瀹℃牳闃熷垪
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

// ModerateContent 瀹℃牳鍐呭
func (s *SuperAdminService) ModerateContent(contentID uint, action, reason string, reviewerID uint) error {
	// 鏇存柊瀹℃牳璁板綍
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

	// 鏍规嵁action鎵ц鐩稿簲鎿嶄綔
	// delete: 鍒犻櫎鍐呭
	// warn: 璀﹀憡鐢ㄦ埛
	// ban: 灏佺鐢ㄦ埛

	logrus.WithFields(logrus.Fields{
		"content_id":  contentID,
		"reviewer_id": reviewerID,
		"action":      action,
		"reason":      reason,
	}).Info("鍐呭瀹℃牳瀹屾垚")

	return nil
}

// GetSystemLogs 鑾峰彇绯荤粺鏃ュ織
func (s *SuperAdminService) GetSystemLogs(logType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 浠嶳edis鎴栨暟鎹簱鑾峰彇鏃ュ織
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

// BroadcastMessage 骞挎挱绯荤粺娑堟伅
func (s *SuperAdminService) BroadcastMessage(message string, targetType string, targetIDs []uint) error {
	// targetType: all, users, groups
	// 瀹炵幇绯荤粺骞挎挱鍔熻兘

	logrus.WithFields(logrus.Fields{
		"message":     message,
		"target_type": targetType,
		"target_ids":  targetIDs,
		"action":      "broadcast_message",
	}).Info("绯荤粺骞挎挱娑堟伅")

	return nil
}

// GetServerHealth 鑾峰彇鏈嶅姟鍣ㄥ仴搴风姸鎬?func (s *SuperAdminService) GetServerHealth() (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// 妫€鏌ユ暟鎹簱杩炴帴
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

	// 妫€鏌edis杩炴帴
	if err := s.redis.Ping(s.ctx).Err(); err != nil {
		health["redis"] = "disconnected"
	} else {
		health["redis"] = "healthy"
	}

	health["timestamp"] = time.Now()
	health["status"] = "running"

	return health, nil
}
