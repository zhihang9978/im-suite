package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// ChatStatisticsService 群组统计服务
type ChatStatisticsService struct {
	db *gorm.DB
}

// NewChatStatisticsService 创建群组统计服务
func NewChatStatisticsService(db *gorm.DB) *ChatStatisticsService {
	return &ChatStatisticsService{
		db: db,
	}
}

// StatisticsRequest 统计请求
type StatisticsRequest struct {
	ChatID    uint      `json:"chat_id" binding:"required"`
	DateFrom  *time.Time `json:"date_from,omitempty"`
	DateTo    *time.Time `json:"date_to,omitempty"`
	GroupBy   string    `json:"group_by,omitempty"` // hour, day, week, month
}

// StatisticsResponse 统计响应
type StatisticsResponse struct {
	ChatID                uint                    `json:"chat_id"`
	TotalMembers          int64                   `json:"total_members"`
	ActiveMembers         int64                   `json:"active_members"`
	TotalMessages         int64                   `json:"total_messages"`
	MessagesToday         int64                   `json:"messages_today"`
	MessagesThisWeek      int64                   `json:"messages_this_week"`
	MessagesThisMonth     int64                   `json:"messages_this_month"`
	TotalFiles            int64                   `json:"total_files"`
	TotalImages           int64                   `json:"total_images"`
	TotalVideos           int64                   `json:"total_videos"`
	TotalAudios           int64                   `json:"total_audios"`
	TotalVoiceCalls       int                     `json:"total_voice_calls"`
	TotalVideoCalls       int                     `json:"total_video_calls"`
	AverageMessageLength  float64                 `json:"average_message_length"`
	PeakActivityHour      int                     `json:"peak_activity_hour"`
	LastActivityAt        time.Time               `json:"last_activity_at"`
	MessageTrends         []MessageTrendData      `json:"message_trends,omitempty"`
	MemberActivity        []MemberActivityData    `json:"member_activity,omitempty"`
	MessageTypeDistribution []MessageTypeData     `json:"message_type_distribution,omitempty"`
	TopActiveMembers      []TopMemberData         `json:"top_active_members,omitempty"`
}

// MessageTrendData 消息趋势数据
type MessageTrendData struct {
	Date      string `json:"date"`
	Count     int    `json:"count"`
	Hour      *int   `json:"hour,omitempty"`
	Day       *int   `json:"day,omitempty"`
	Week      *int   `json:"week,omitempty"`
	Month     *int   `json:"month,omitempty"`
}

// MemberActivityData 成员活跃度数据
type MemberActivityData struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	MessageCount int    `json:"message_count"`
	LastActive   string `json:"last_active"`
	JoinDate     string `json:"join_date"`
}

// MessageTypeData 消息类型分布数据
type MessageTypeData struct {
	MessageType string `json:"message_type"`
	Count       int    `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// TopMemberData 最活跃成员数据
type TopMemberData struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	MessageCount int    `json:"message_count"`
	Rank         int    `json:"rank"`
}

// GetChatStatistics 获取群组统计信息
func (s *ChatStatisticsService) GetChatStatistics(ctx context.Context, userID uint, req *StatisticsRequest) (*StatisticsResponse, error) {
	// 检查用户是否为群成员
	if !s.isChatMember(ctx, req.ChatID, userID) {
		return nil, fmt.Errorf("不是群成员")
	}

	// 获取基础统计信息
	stats, err := s.getBasicStatistics(ctx, req.ChatID)
	if err != nil {
		return nil, fmt.Errorf("获取基础统计失败: %w", err)
	}

	// 获取消息趋势
	messageTrends, err := s.getMessageTrends(ctx, req.ChatID, req.DateFrom, req.DateTo, req.GroupBy)
	if err != nil {
		return nil, fmt.Errorf("获取消息趋势失败: %w", err)
	}

	// 获取成员活跃度
	memberActivity, err := s.getMemberActivity(ctx, req.ChatID, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("获取成员活跃度失败: %w", err)
	}

	// 获取消息类型分布
	messageTypeDistribution, err := s.getMessageTypeDistribution(ctx, req.ChatID, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("获取消息类型分布失败: %w", err)
	}

	// 获取最活跃成员
	topActiveMembers, err := s.getTopActiveMembers(ctx, req.ChatID, req.DateFrom, req.DateTo, 10)
	if err != nil {
		return nil, fmt.Errorf("获取最活跃成员失败: %w", err)
	}

	response := &StatisticsResponse{
		ChatID:                  req.ChatID,
		TotalMembers:            stats.TotalMembers,
		ActiveMembers:           stats.ActiveMembers,
		TotalMessages:           stats.TotalMessages,
		MessagesToday:           stats.MessagesToday,
		MessagesThisWeek:        stats.MessagesThisWeek,
		MessagesThisMonth:       stats.MessagesThisMonth,
		TotalFiles:              stats.TotalFiles,
		TotalImages:             stats.TotalImages,
		TotalVideos:             stats.TotalVideos,
		TotalAudios:             stats.TotalAudios,
		TotalVoiceCalls:         stats.TotalVoiceCalls,
		TotalVideoCalls:         stats.TotalVideoCalls,
		AverageMessageLength:    stats.AverageMessageLength,
		PeakActivityHour:        stats.PeakActivityHour,
		LastActivityAt:          stats.LastActivityAt,
		MessageTrends:           messageTrends,
		MemberActivity:          memberActivity,
		MessageTypeDistribution: messageTypeDistribution,
		TopActiveMembers:        topActiveMembers,
	}

	return response, nil
}

// UpdateChatStatistics 更新群组统计信息（定时任务调用）
func (s *ChatStatisticsService) UpdateChatStatistics(ctx context.Context, chatID uint) error {
	// 获取基础统计
	stats, err := s.calculateBasicStatistics(ctx, chatID)
	if err != nil {
		return fmt.Errorf("计算基础统计失败: %w", err)
	}

	// 查找或创建统计记录
	var chatStats model.ChatStatistics
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&chatStats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的统计记录
			chatStats = model.ChatStatistics{
				ChatID: chatID,
			}
		} else {
			return fmt.Errorf("查询统计记录失败: %w", err)
		}
	}

	// 更新统计信息
	updates := map[string]interface{}{
		"total_members":            stats.TotalMembers,
		"active_members":           stats.ActiveMembers,
		"total_messages":           stats.TotalMessages,
		"messages_today":           stats.MessagesToday,
		"messages_this_week":       stats.MessagesThisWeek,
		"messages_this_month":      stats.MessagesThisMonth,
		"total_files":              stats.TotalFiles,
		"total_images":             stats.TotalImages,
		"total_videos":             stats.TotalVideos,
		"total_audios":             stats.TotalAudios,
		"total_voice_calls":        stats.TotalVoiceCalls,
		"total_video_calls":        stats.TotalVideoCalls,
		"average_message_length":   stats.AverageMessageLength,
		"peak_activity_hour":       stats.PeakActivityHour,
		"last_activity_at":         stats.LastActivityAt,
		"updated_at":               time.Now(),
	}

	if chatStats.ID == 0 {
		// 创建新记录
		if err := s.db.WithContext(ctx).Create(&chatStats).Error; err != nil {
			return fmt.Errorf("创建统计记录失败: %w", err)
		}
	} else {
		// 更新现有记录
		if err := s.db.WithContext(ctx).Model(&chatStats).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新统计记录失败: %w", err)
		}
	}

	return nil
}

// getBasicStatistics 获取基础统计信息
func (s *ChatStatisticsService) getBasicStatistics(ctx context.Context, chatID uint) (*model.ChatStatistics, error) {
	var stats model.ChatStatistics
	if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有统计记录，计算并创建
			if err := s.UpdateChatStatistics(ctx, chatID); err != nil {
				return nil, err
			}
			// 重新查询
			if err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).First(&stats).Error; err != nil {
				return nil, fmt.Errorf("查询统计记录失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询统计记录失败: %w", err)
		}
	}

	return &stats, nil
}

// calculateBasicStatistics 计算基础统计信息
func (s *ChatStatisticsService) calculateBasicStatistics(ctx context.Context, chatID uint) (*model.ChatStatistics, error) {
	stats := &model.ChatStatistics{
		ChatID: chatID,
	}

	// 计算总成员数
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND is_active = ?", chatID, true).
		Count(&stats.TotalMembers)

	// 计算活跃成员数（最近7天有消息）
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	s.db.WithContext(ctx).Model(&model.Message{}).
		Select("COUNT(DISTINCT sender_id)").
		Where("chat_id = ? AND created_at >= ?", chatID, sevenDaysAgo).
		Count(&stats.ActiveMembers)

	// 计算总消息数
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND is_recalled = ?", chatID, false).
		Count(&stats.TotalMessages)

	// 计算今日消息数
	today := time.Now().Truncate(24 * time.Hour)
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND created_at >= ? AND is_recalled = ?", chatID, today, false).
		Count(&stats.MessagesToday)

	// 计算本周消息数
	weekStart := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -int(time.Now().Weekday()))
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND created_at >= ? AND is_recalled = ?", chatID, weekStart, false).
		Count(&stats.MessagesThisWeek)

	// 计算本月消息数
	monthStart := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND created_at >= ? AND is_recalled = ?", chatID, monthStart, false).
		Count(&stats.MessagesThisMonth)

	// 计算各种类型消息数量
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND message_type = ? AND is_recalled = ?", chatID, "file", false).
		Count(&stats.TotalFiles)

	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND message_type = ? AND is_recalled = ?", chatID, "image", false).
		Count(&stats.TotalImages)

	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND message_type = ? AND is_recalled = ?", chatID, "video", false).
		Count(&stats.TotalVideos)

	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND message_type = ? AND is_recalled = ?", chatID, "audio", false).
		Count(&stats.TotalAudios)

	// 计算平均消息长度
	var avgLength float64
	s.db.WithContext(ctx).Model(&model.Message{}).
		Select("AVG(LENGTH(content))").
		Where("chat_id = ? AND message_type = ? AND is_recalled = ?", chatID, "text", false).
		Scan(&avgLength)
	stats.AverageMessageLength = avgLength

	// 计算最活跃时段
	var peakHour int
	s.db.WithContext(ctx).Model(&model.Message{}).
		Select("HOUR(created_at) as hour").
		Where("chat_id = ? AND is_recalled = ?", chatID, false).
		Group("HOUR(created_at)").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&peakHour)
	stats.PeakActivityHour = peakHour

	// 获取最后活动时间
	var lastActivity time.Time
	s.db.WithContext(ctx).Model(&model.Message{}).
		Select("MAX(created_at)").
		Where("chat_id = ?", chatID).
		Scan(&lastActivity)
	stats.LastActivityAt = lastActivity

	return stats, nil
}

// getMessageTrends 获取消息趋势
func (s *ChatStatisticsService) getMessageTrends(ctx context.Context, chatID uint, dateFrom, dateTo *time.Time, groupBy string) ([]MessageTrendData, error) {
	var trends []MessageTrendData

	// 设置默认时间范围
	if dateFrom == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		dateFrom = &thirtyDaysAgo
	}
	if dateTo == nil {
		now := time.Now()
		dateTo = &now
	}

	// 根据分组方式构建查询
	var query string
	var args []interface{}

	switch groupBy {
	case "hour":
		query = `SELECT DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as date, 
		                HOUR(created_at) as hour,
		                COUNT(*) as count 
		         FROM messages 
		         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
		         GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')
		         ORDER BY date`
		args = []interface{}{chatID, *dateFrom, *dateTo, false}
	case "day":
		query = `SELECT DATE(created_at) as date, 
		                DAY(created_at) as day,
		                COUNT(*) as count 
		         FROM messages 
		         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
		         GROUP BY DATE(created_at)
		         ORDER BY date`
		args = []interface{}{chatID, *dateFrom, *dateTo, false}
	case "week":
		query = `SELECT YEARWEEK(created_at) as date, 
		                WEEK(created_at) as week,
		                COUNT(*) as count 
		         FROM messages 
		         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
		         GROUP BY YEARWEEK(created_at)
		         ORDER BY date`
		args = []interface{}{chatID, *dateFrom, *dateTo, false}
	case "month":
		query = `SELECT DATE_FORMAT(created_at, '%Y-%m') as date, 
		                MONTH(created_at) as month,
		                COUNT(*) as count 
		         FROM messages 
		         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
		         GROUP BY DATE_FORMAT(created_at, '%Y-%m')
		         ORDER BY date`
		args = []interface{}{chatID, *dateFrom, *dateTo, false}
	default:
		// 默认按天分组
		query = `SELECT DATE(created_at) as date, 
		                DAY(created_at) as day,
		                COUNT(*) as count 
		         FROM messages 
		         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
		         GROUP BY DATE(created_at)
		         ORDER BY date`
		args = []interface{}{chatID, *dateFrom, *dateTo, false}
	}

	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询消息趋势失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trend MessageTrendData
		var day, week, month *int
		var hour *int

		if err := rows.Scan(&trend.Date, &day, &hour, &week, &month, &trend.Count); err != nil {
			return nil, fmt.Errorf("扫描趋势数据失败: %w", err)
		}

		switch groupBy {
		case "hour":
			trend.Hour = hour
		case "day":
			trend.Day = day
		case "week":
			trend.Week = week
		case "month":
			trend.Month = month
		}

		trends = append(trends, trend)
	}

	return trends, nil
}

// getMemberActivity 获取成员活跃度
func (s *ChatStatisticsService) getMemberActivity(ctx context.Context, chatID uint, dateFrom, dateTo *time.Time) ([]MemberActivityData, error) {
	var activity []MemberActivityData

	// 设置默认时间范围
	if dateFrom == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		dateFrom = &thirtyDaysAgo
	}
	if dateTo == nil {
		now := time.Now()
		dateTo = &now
	}

	query := `SELECT m.user_id, u.username, u.nickname, 
	                COUNT(m.id) as message_count,
	                MAX(m.created_at) as last_active,
	                cm.joined_at as join_date
	         FROM messages m
	         JOIN users u ON m.sender_id = u.id
	         JOIN chat_members cm ON m.sender_id = cm.user_id AND m.chat_id = cm.chat_id
	         WHERE m.chat_id = ? AND m.created_at >= ? AND m.created_at <= ? AND m.is_recalled = ?
	         GROUP BY m.sender_id, u.username, u.nickname, cm.joined_at
	         ORDER BY message_count DESC`

	rows, err := s.db.WithContext(ctx).Raw(query, chatID, *dateFrom, *dateTo, false).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询成员活跃度失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member MemberActivityData
		var lastActive, joinDate time.Time

		if err := rows.Scan(&member.UserID, &member.Username, &member.Nickname, 
			&member.MessageCount, &lastActive, &joinDate); err != nil {
			return nil, fmt.Errorf("扫描成员活跃度数据失败: %w", err)
		}

		member.LastActive = lastActive.Format("2006-01-02 15:04:05")
		member.JoinDate = joinDate.Format("2006-01-02")

		activity = append(activity, member)
	}

	return activity, nil
}

// getMessageTypeDistribution 获取消息类型分布
func (s *ChatStatisticsService) getMessageTypeDistribution(ctx context.Context, chatID uint, dateFrom, dateTo *time.Time) ([]MessageTypeData, error) {
	var distribution []MessageTypeData

	// 设置默认时间范围
	if dateFrom == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		dateFrom = &thirtyDaysAgo
	}
	if dateTo == nil {
		now := time.Now()
		dateTo = &now
	}

	// 获取总数
	var total int64
	s.db.WithContext(ctx).Model(&model.Message{}).
		Where("chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ?", 
			chatID, *dateFrom, *dateTo, false).
		Count(&total)

	if total == 0 {
		return distribution, nil
	}

	// 获取各类型消息数量
	query := `SELECT message_type, COUNT(*) as count 
	         FROM messages 
	         WHERE chat_id = ? AND created_at >= ? AND created_at <= ? AND is_recalled = ? 
	         GROUP BY message_type 
	         ORDER BY count DESC`

	rows, err := s.db.WithContext(ctx).Raw(query, chatID, *dateFrom, *dateTo, false).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询消息类型分布失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var typeData MessageTypeData

		if err := rows.Scan(&typeData.MessageType, &typeData.Count); err != nil {
			return nil, fmt.Errorf("扫描消息类型数据失败: %w", err)
		}

		typeData.Percentage = float64(typeData.Count) / float64(total) * 100

		distribution = append(distribution, typeData)
	}

	return distribution, nil
}

// getTopActiveMembers 获取最活跃成员
func (s *ChatStatisticsService) getTopActiveMembers(ctx context.Context, chatID uint, dateFrom, dateTo *time.Time, limit int) ([]TopMemberData, error) {
	var topMembers []TopMemberData

	// 设置默认时间范围
	if dateFrom == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		dateFrom = &thirtyDaysAgo
	}
	if dateTo == nil {
		now := time.Now()
		dateTo = &now
	}

	query := `SELECT m.sender_id, u.username, u.nickname, 
	                COUNT(m.id) as message_count
	         FROM messages m
	         JOIN users u ON m.sender_id = u.id
	         WHERE m.chat_id = ? AND m.created_at >= ? AND m.created_at <= ? AND m.is_recalled = ?
	         GROUP BY m.sender_id, u.username, u.nickname
	         ORDER BY message_count DESC
	         LIMIT ?`

	rows, err := s.db.WithContext(ctx).Raw(query, chatID, *dateFrom, *dateTo, false, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询最活跃成员失败: %w", err)
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var member TopMemberData

		if err := rows.Scan(&member.UserID, &member.Username, &member.Nickname, &member.MessageCount); err != nil {
			return nil, fmt.Errorf("扫描最活跃成员数据失败: %w", err)
		}

		member.Rank = rank
		topMembers = append(topMembers, member)
		rank++
	}

	return topMembers, nil
}

// 辅助方法

// isChatMember 检查用户是否为群成员
func (s *ChatStatisticsService) isChatMember(ctx context.Context, chatID uint, userID uint) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND is_active = ?", chatID, userID, true).
		Count(&count)
	return count > 0
}
