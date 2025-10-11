package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ScreenShareEnhancedService 屏幕共享增强服务
type ScreenShareEnhancedService struct {
	db            *gorm.DB
	webrtcService *WebRTCService
}

// NewScreenShareEnhancedService 创建屏幕共享增强服务
func NewScreenShareEnhancedService(webrtcService *WebRTCService) *ScreenShareEnhancedService {
	return &ScreenShareEnhancedService{
		db:            config.DB,
		webrtcService: webrtcService,
	}
}

// ScreenSharePermission 屏幕共享权限配置
type ScreenSharePermission struct {
	Role             string `json:"role"`              // 角色: user, admin, super_admin
	CanShare         bool   `json:"can_share"`         // 是否可以共享
	CanRecord        bool   `json:"can_record"`        // 是否可以录制
	MaxDuration      int    `json:"max_duration"`      // 最大时长（秒）0表示无限制
	MaxQuality       string `json:"max_quality"`       // 最大质量: high, medium, low
	RequiresApproval bool   `json:"requires_approval"` // 是否需要审批
}

// 预定义权限
var defaultPermissions = map[string]ScreenSharePermission{
	"user": {
		Role:             "user",
		CanShare:         true,
		CanRecord:        false,
		MaxDuration:      3600, // 1小时
		MaxQuality:       "medium",
		RequiresApproval: false,
	},
	"admin": {
		Role:             "admin",
		CanShare:         true,
		CanRecord:        true,
		MaxDuration:      7200, // 2小时
		MaxQuality:       "high",
		RequiresApproval: false,
	},
	"super_admin": {
		Role:             "super_admin",
		CanShare:         true,
		CanRecord:        true,
		MaxDuration:      0, // 无限制
		MaxQuality:       "high",
		RequiresApproval: false,
	},
}

// CheckSharePermission 检查屏幕共享权限
func (s *ScreenShareEnhancedService) CheckSharePermission(userID uint, quality string) error {
	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 获取权限配置
	permission, exists := defaultPermissions[user.Role]
	if !exists {
		permission = defaultPermissions["user"] // 默认使用普通用户权限
	}

	// 检查是否允许共享
	if !permission.CanShare {
		return errors.New("您没有屏幕共享权限")
	}

	// 检查质量限制
	if !s.isQualityAllowed(quality, permission.MaxQuality) {
		return fmt.Errorf("您的最高质量限制为: %s", permission.MaxQuality)
	}

	return nil
}

// isQualityAllowed 检查质量是否在允许范围内
func (s *ScreenShareEnhancedService) isQualityAllowed(requestQuality, maxQuality string) bool {
	qualityLevels := map[string]int{
		"low":    1,
		"medium": 2,
		"high":   3,
	}

	requestLevel := qualityLevels[requestQuality]
	maxLevel := qualityLevels[maxQuality]

	return requestLevel <= maxLevel
}

// StartShareSession 开始共享会话（增强版）
func (s *ScreenShareEnhancedService) StartShareSession(ctx context.Context, callID string, sharerID uint, sharerName string, quality string, withAudio bool) (*model.ScreenShareSession, error) {
	// 检查权限
	if err := s.CheckSharePermission(sharerID, quality); err != nil {
		return nil, err
	}

	// 创建会话记录
	session := model.ScreenShareSession{
		CallID:           callID,
		SharerID:         sharerID,
		SharerName:       sharerName,
		StartTime:        time.Now(),
		Quality:          quality,
		WithAudio:        withAudio,
		InitialQuality:   quality,
		QualityChanges:   0,
		ParticipantCount: 0,
		Status:           "active",
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, fmt.Errorf("创建会话记录失败: %w", err)
	}

	logrus.Infof("创建屏幕共享会话: ID=%d, CallID=%s, Sharer=%s", session.ID, callID, sharerName)

	return &session, nil
}

// EndShareSession 结束共享会话
func (s *ScreenShareEnhancedService) EndShareSession(ctx context.Context, callID string, endReason string) error {
	// 查找活跃会话
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("call_id = ? AND status = ?", callID, "active").
		First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未找到活跃的屏幕共享会话")
		}
		return err
	}

	// 更新会话
	endTime := time.Now()
	duration := int64(endTime.Sub(session.StartTime).Seconds())

	updates := map[string]interface{}{
		"end_time":   endTime,
		"duration":   duration,
		"status":     "ended",
		"end_reason": endReason,
	}

	if err := s.db.WithContext(ctx).Model(&session).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新会话失败: %w", err)
	}

	// 更新用户统计
	go s.updateUserStatistics(session.SharerID, duration)

	logrus.Infof("结束屏幕共享会话: ID=%d, Duration=%ds, Reason=%s", session.ID, duration, endReason)

	return nil
}

// RecordQualityChange 记录质量变更
func (s *ScreenShareEnhancedService) RecordQualityChange(ctx context.Context, callID string, fromQuality, toQuality, reason string, networkSpeed, cpuUsage float64) error {
	// 查找活跃会话
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("call_id = ? AND status = ?", callID, "active").
		First(&session).Error; err != nil {
		return err
	}

	// 创建质量变更记录
	change := model.ScreenShareQualityChange{
		SessionID:    session.ID,
		FromQuality:  fromQuality,
		ToQuality:    toQuality,
		ChangeTime:   time.Now(),
		ChangeReason: reason,
		NetworkSpeed: networkSpeed,
		CPUUsage:     cpuUsage,
	}

	if err := s.db.WithContext(ctx).Create(&change).Error; err != nil {
		return fmt.Errorf("创建质量变更记录失败: %w", err)
	}

	// 更新会话的质量变更次数
	s.db.WithContext(ctx).Model(&session).Update("quality_changes", gorm.Expr("quality_changes + 1"))
	s.db.WithContext(ctx).Model(&session).Update("quality", toQuality)

	logrus.Infof("记录质量变更: SessionID=%d, %s -> %s, Reason=%s", session.ID, fromQuality, toQuality, reason)

	return nil
}

// AddParticipant 添加参与者
func (s *ScreenShareEnhancedService) AddParticipant(ctx context.Context, callID string, userID uint, userName string) error {
	// 查找活跃会话
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("call_id = ? AND status = ?", callID, "active").
		First(&session).Error; err != nil {
		return err
	}

	// 检查是否已存在
	var existingParticipant model.ScreenShareParticipant
	err := s.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ? AND leave_time IS NULL", session.ID, userID).
		First(&existingParticipant).Error

	if err == nil {
		// 已经存在，不重复添加
		return nil
	}

	// 创建参与者记录
	participant := model.ScreenShareParticipant{
		SessionID: session.ID,
		UserID:    userID,
		UserName:  userName,
		JoinTime:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&participant).Error; err != nil {
		return fmt.Errorf("添加参与者失败: %w", err)
	}

	// 更新参与者数量
	s.db.WithContext(ctx).Model(&session).Update("participant_count", gorm.Expr("participant_count + 1"))

	logrus.Infof("添加屏幕共享参与者: SessionID=%d, User=%s", session.ID, userName)

	return nil
}

// RemoveParticipant 移除参与者
func (s *ScreenShareEnhancedService) RemoveParticipant(ctx context.Context, callID string, userID uint) error {
	// 查找活跃会话
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("call_id = ? AND status = ?", callID, "active").
		First(&session).Error; err != nil {
		return err
	}

	// 查找参与者
	var participant model.ScreenShareParticipant
	if err := s.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ? AND leave_time IS NULL", session.ID, userID).
		First(&participant).Error; err != nil {
		return err
	}

	// 更新离开时间和观看时长
	leaveTime := time.Now()
	viewDuration := int64(leaveTime.Sub(participant.JoinTime).Seconds())

	updates := map[string]interface{}{
		"leave_time":    leaveTime,
		"view_duration": viewDuration,
	}

	if err := s.db.WithContext(ctx).Model(&participant).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新参与者失败: %w", err)
	}

	logrus.Infof("移除屏幕共享参与者: SessionID=%d, UserID=%d, ViewDuration=%ds", session.ID, userID, viewDuration)

	return nil
}

// updateUserStatistics 更新用户统计信息
func (s *ScreenShareEnhancedService) updateUserStatistics(userID uint, duration int64) {
	ctx := context.Background()

	// 查找或创建统计记录
	var stats model.ScreenShareStatistics
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新统计
		stats = model.ScreenShareStatistics{
			UserID:        userID,
			TotalSessions: 1,
			TotalDuration: duration,
		}
		if err := s.db.WithContext(ctx).Create(&stats).Error; err != nil {
			logrus.Errorf("创建用户统计失败: %v", err)
			return
		}
	} else {
		// 更新统计
		updates := map[string]interface{}{
			"total_sessions": gorm.Expr("total_sessions + 1"),
			"total_duration": gorm.Expr("total_duration + ?", duration),
		}
		if err := s.db.WithContext(ctx).Model(&stats).Updates(updates).Error; err != nil {
			logrus.Errorf("更新用户统计失败: %v", err)
			return
		}

		// 重新计算平均值
		s.db.WithContext(ctx).First(&stats, stats.ID)
		stats.AverageDuration = float64(stats.TotalDuration) / float64(stats.TotalSessions)
		s.db.WithContext(ctx).Save(&stats)
	}

	now := time.Now()
	s.db.WithContext(ctx).Model(&stats).Update("last_share_time", now)

	logrus.Infof("更新用户统计: UserID=%d, TotalSessions=%d", userID, stats.TotalSessions)
}

// GetSessionHistory 获取会话历史
func (s *ScreenShareEnhancedService) GetSessionHistory(ctx context.Context, userID uint, limit, offset int) ([]model.ScreenShareSession, int64, error) {
	var sessions []model.ScreenShareSession
	var total int64

	query := s.db.WithContext(ctx).Where("sharer_id = ?", userID)

	// 获取总数
	if err := query.Model(&model.ScreenShareSession{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	if err := query.Order("start_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// GetUserStatistics 获取用户统计
func (s *ScreenShareEnhancedService) GetUserStatistics(ctx context.Context, userID uint) (*model.ScreenShareStatistics, error) {
	var stats model.ScreenShareStatistics
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回空统计
			return &model.ScreenShareStatistics{
				UserID: userID,
			}, nil
		}
		return nil, err
	}

	return &stats, nil
}

// GetSessionDetails 获取会话详情
func (s *ScreenShareEnhancedService) GetSessionDetails(ctx context.Context, sessionID uint) (*SessionDetailsResponse, error) {
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).Preload("Sharer").First(&session, sessionID).Error; err != nil {
		return nil, err
	}

	// 获取质量变更记录
	var qualityChanges []model.ScreenShareQualityChange
	s.db.WithContext(ctx).Where("session_id = ?", sessionID).
		Order("change_time ASC").
		Find(&qualityChanges)

	// 获取参与者列表
	var participants []model.ScreenShareParticipant
	s.db.WithContext(ctx).Where("session_id = ?", sessionID).
		Preload("User").
		Order("join_time ASC").
		Find(&participants)

	// 获取录制列表
	var recordings []model.ScreenShareRecording
	s.db.WithContext(ctx).Where("session_id = ?", sessionID).
		Order("start_time ASC").
		Find(&recordings)

	return &SessionDetailsResponse{
		Session:        session,
		QualityChanges: qualityChanges,
		Participants:   participants,
		Recordings:     recordings,
	}, nil
}

// SessionDetailsResponse 会话详情响应
type SessionDetailsResponse struct {
	Session        model.ScreenShareSession         `json:"session"`
	QualityChanges []model.ScreenShareQualityChange `json:"quality_changes"`
	Participants   []model.ScreenShareParticipant   `json:"participants"`
	Recordings     []model.ScreenShareRecording     `json:"recordings"`
}

// NetworkQualityMonitor 网络质量监控
type NetworkQualityMonitor struct {
	callID        string
	currentSpeed  float64 // Kbps
	avgSpeed      float64
	speedSamples  []float64
	cpuUsage      float64
	lastCheckTime time.Time
}

// CheckAndAdjustQuality 检查并调整质量
func (s *ScreenShareEnhancedService) CheckAndAdjustQuality(monitor *NetworkQualityMonitor, currentQuality string) (newQuality string, shouldChange bool) {
	// 根据网络速度推荐质量
	recommendedQuality := s.recommendQualityBySpeed(monitor.avgSpeed, monitor.cpuUsage)

	// 如果推荐质量与当前质量不同，建议切换
	if recommendedQuality != currentQuality {
		return recommendedQuality, true
	}

	return currentQuality, false
}

// recommendQualityBySpeed 根据速度推荐质量
func (s *ScreenShareEnhancedService) recommendQualityBySpeed(speed, cpuUsage float64) string {
	// 综合考虑网速和CPU使用率
	if speed > 3000 && cpuUsage < 70 { // > 3Mbps && CPU < 70%
		return "high"
	} else if speed > 1000 && cpuUsage < 80 { // > 1Mbps && CPU < 80%
		return "medium"
	}
	return "low"
}

// StartRecording 开始录制
func (s *ScreenShareEnhancedService) StartRecording(ctx context.Context, callID string, recorderID uint, format, quality string) (*model.ScreenShareRecording, error) {
	// 检查录制权限
	var user model.User
	if err := s.db.First(&user, recorderID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	permission, exists := defaultPermissions[user.Role]
	if !exists || !permission.CanRecord {
		return nil, errors.New("您没有录制权限")
	}

	// 查找活跃会话
	var session model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("call_id = ? AND status = ?", callID, "active").
		First(&session).Error; err != nil {
		return nil, errors.New("未找到活跃的屏幕共享会话")
	}

	// 创建录制记录
	recording := model.ScreenShareRecording{
		SessionID:  session.ID,
		RecorderID: recorderID,
		Format:     format,
		Quality:    quality,
		StartTime:  time.Now(),
		Status:     "recording",
	}

	if err := s.db.WithContext(ctx).Create(&recording).Error; err != nil {
		return nil, fmt.Errorf("创建录制记录失败: %w", err)
	}

	logrus.Infof("开始录制: RecordingID=%d, SessionID=%d, Format=%s", recording.ID, session.ID, format)

	return &recording, nil
}

// EndRecording 结束录制
func (s *ScreenShareEnhancedService) EndRecording(ctx context.Context, recordingID uint, filePath string, fileSize int64) error {
	var recording model.ScreenShareRecording
	if err := s.db.WithContext(ctx).First(&recording, recordingID).Error; err != nil {
		return err
	}

	endTime := time.Now()
	duration := int64(endTime.Sub(recording.StartTime).Seconds())

	updates := map[string]interface{}{
		"end_time":  endTime,
		"duration":  duration,
		"file_path": filePath,
		"file_size": fileSize,
		"status":    "completed",
	}

	if err := s.db.WithContext(ctx).Model(&recording).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新录制记录失败: %w", err)
	}

	logrus.Infof("结束录制: RecordingID=%d, Duration=%ds, Size=%d bytes", recordingID, duration, fileSize)

	return nil
}

// GetRecordings 获取录制列表
func (s *ScreenShareEnhancedService) GetRecordings(ctx context.Context, sessionID uint) ([]model.ScreenShareRecording, error) {
	var recordings []model.ScreenShareRecording
	if err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("start_time DESC").
		Find(&recordings).Error; err != nil {
		return nil, err
	}

	return recordings, nil
}

// ExportStatistics 导出统计数据
func (s *ScreenShareEnhancedService) ExportStatistics(ctx context.Context, userID uint, startTime, endTime time.Time) (string, error) {
	// 查询会话
	var sessions []model.ScreenShareSession
	if err := s.db.WithContext(ctx).
		Where("sharer_id = ? AND start_time >= ? AND start_time <= ?", userID, startTime, endTime).
		Find(&sessions).Error; err != nil {
		return "", err
	}

	// 组装数据
	data := map[string]interface{}{
		"user_id":    userID,
		"start_time": startTime,
		"end_time":   endTime,
		"sessions":   sessions,
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("导出失败: %w", err)
	}

	return string(jsonData), nil
}
