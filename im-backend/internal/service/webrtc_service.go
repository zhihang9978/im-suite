package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"zhihang-messenger/im-backend/internal/model"
	"zhihang-messenger/im-backend/internal/utils"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebRTCService WebRTC 通话服务
type WebRTCService struct {
	db              *gorm.DB
	activeCalls     map[string]*CallSession
	callMutex       sync.RWMutex
	networkMonitor  *NetworkQualityMonitor
	codecManager    *CodecManager
	bandwidthAdaptor *BandwidthAdaptor
}

// CallSession 通话会话
type CallSession struct {
	ID           string                    `json:"id"`
	CallerID     uint                      `json:"caller_id"`
	CalleeID     uint                      `json:"callee_id"`
	Type         string                    `json:"type"` // audio, video
	Status       string                    `json:"status"` // ringing, connecting, connected, ended
	StartTime    time.Time                 `json:"start_time"`
	EndTime      *time.Time                `json:"end_time,omitempty"`
	Duration     int64                     `json:"duration,omitempty"`
	Quality      *CallQuality              `json:"quality,omitempty"`
	Connections  map[uint]*PeerConnection  `json:"connections"`
	NetworkStats map[uint]*NetworkStats    `json:"network_stats"`
	Mutex        sync.RWMutex              `json:"-"`
}

// PeerConnection 对等连接
type PeerConnection struct {
	UserID      uint                      `json:"user_id"`
	Conn        *websocket.Conn           `json:"-"`
	PC          *webrtc.PeerConnection    `json:"-"`
	AudioTrack  *webrtc.TrackLocalStaticSample `json:"-"`
	VideoTrack  *webrtc.TrackLocalStaticSample `json:"-"`
	IsMuted     bool                      `json:"is_muted"`
	IsVideoOff  bool                      `json:"is_video_off"`
	Bitrate     int                       `json:"bitrate"`
	Resolution  string                    `json:"resolution"`
	Codec       string                    `json:"codec"`
	LastPing    time.Time                 `json:"last_ping"`
	PingCount   int                       `json:"ping_count"`
}

// CallQuality 通话质量
type CallQuality struct {
	OverallScore     float64 `json:"overall_score"`     // 0-100
	AudioQuality     float64 `json:"audio_quality"`     // 0-100
	VideoQuality     float64 `json:"video_quality"`     // 0-100
	NetworkQuality   float64 `json:"network_quality"`   // 0-100
	Latency          int     `json:"latency"`           // ms
	PacketLoss       float64 `json:"packet_loss"`       // %
	Jitter           float64 `json:"jitter"`            // ms
	Bitrate          int     `json:"bitrate"`           // kbps
	Resolution       string  `json:"resolution"`
	FrameRate        int     `json:"frame_rate"`
	AudioCodec       string  `json:"audio_codec"`
	VideoCodec       string  `json:"video_codec"`
	AdaptiveQuality  bool    `json:"adaptive_quality"`
	QualityHistory   []QualitySnapshot `json:"quality_history"`
}

// QualitySnapshot 质量快照
type QualitySnapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Score     float64   `json:"score"`
	Latency   int       `json:"latency"`
	Bitrate   int       `json:"bitrate"`
}

// NetworkStats 网络统计
type NetworkStats struct {
	UserID           uint    `json:"user_id"`
	RTT              int     `json:"rtt"`              // Round Trip Time (ms)
	PacketLoss       float64 `json:"packet_loss"`      // %
	Jitter           float64 `json:"jitter"`           // ms
	Bandwidth        int     `json:"bandwidth"`        // kbps
	AvailableBitrate int     `json:"available_bitrate"` // kbps
	NetworkType      string  `json:"network_type"`     // wifi, 4g, 3g, 2g
	SignalStrength   int     `json:"signal_strength"`  // 0-100
	IsStable         bool    `json:"is_stable"`
	LastUpdate       time.Time `json:"last_update"`
}

// CallRequest 通话请求
type CallRequest struct {
	CalleeID uint   `json:"callee_id" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=audio video"`
}

// CallResponse 通话响应
type CallResponse struct {
	CallID   string        `json:"call_id"`
	Status   string        `json:"status"`
	ICEServers []ICEServer `json:"ice_servers"`
}

// ICEServer ICE 服务器配置
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// NewWebRTCService 创建 WebRTC 服务
func NewWebRTCService(db *gorm.DB) *WebRTCService {
	return &WebRTCService{
		db:              db,
		activeCalls:     make(map[string]*CallSession),
		networkMonitor:  NewNetworkQualityMonitor(),
		codecManager:    NewCodecManager(),
		bandwidthAdaptor: NewBandwidthAdaptor(),
	}
}

// StartCall 开始通话
func (s *WebRTCService) StartCall(callerID uint, req CallRequest) (*CallResponse, error) {
	// 验证被叫用户是否存在
	var callee model.User
	if err := s.db.First(&callee, req.CalleeID).Error; err != nil {
		return nil, utils.NewAppError(utils.ErrCodeUserNotFound, "被叫用户不存在")
	}

	// 检查用户是否在线
	if !callee.OnlineStatus {
		return nil, utils.NewAppError(utils.ErrCodeUserNotFound, "用户当前不在线")
	}

	// 检查是否已有活跃通话
	if s.hasActiveCall(callerID) {
		return nil, utils.NewAppError(utils.ErrCodeTooManyRequests, "您已有活跃的通话")
	}

	if s.hasActiveCall(req.CalleeID) {
		return nil, utils.NewAppError(utils.ErrCodeTooManyRequests, "对方正在通话中")
	}

	// 创建通话会话
	callID := s.generateCallID()
	callSession := &CallSession{
		ID:          callID,
		CallerID:    callerID,
		CalleeID:    req.CalleeID,
		Type:        req.Type,
		Status:      "ringing",
		StartTime:   time.Now(),
		Quality:     &CallQuality{},
		Connections: make(map[uint]*PeerConnection),
		NetworkStats: make(map[uint]*NetworkStats),
	}

	// 存储通话会话
	s.callMutex.Lock()
	s.activeCalls[callID] = callSession
	s.callMutex.Unlock()

	// 获取 ICE 服务器配置
	iceServers := s.getICEServers()

	// 记录通话开始
	if err := s.recordCallStart(callSession); err != nil {
		logrus.WithError(err).Error("记录通话开始失败")
	}

	// 启动通话质量监控
	go s.monitorCallQuality(callSession)

	return &CallResponse{
		CallID:     callID,
		Status:     "ringing",
		ICEServers: iceServers,
	}, nil
}

// AnswerCall 接听通话
func (s *WebRTCService) AnswerCall(callID string, userID uint) error {
	s.callMutex.Lock()
	call, exists := s.activeCalls[callID]
	s.callMutex.Unlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	if call.CalleeID != userID {
		return utils.NewAppError(utils.ErrCodePermissionDenied, "无权限接听此通话")
	}

	if call.Status != "ringing" {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "通话状态错误")
	}

	// 更新通话状态
	call.Mutex.Lock()
	call.Status = "connecting"
	call.Mutex.Unlock()

	// 记录通话接听
	if err := s.recordCallAnswer(call); err != nil {
		logrus.WithError(err).Error("记录通话接听失败")
	}

	return nil
}

// RejectCall 拒绝通话
func (s *WebRTCService) RejectCall(callID string, userID uint) error {
	s.callMutex.Lock()
	call, exists := s.activeCalls[callID]
	s.callMutex.Unlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	if call.CalleeID != userID {
		return utils.NewAppError(utils.ErrCodePermissionDenied, "无权限拒绝此通话")
	}

	// 结束通话
	return s.EndCall(callID, userID, "rejected")
}

// EndCall 结束通话
func (s *WebRTCService) EndCall(callID string, userID uint, reason string) error {
	s.callMutex.Lock()
	call, exists := s.activeCalls[callID]
	s.callMutex.Unlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	if call.CallerID != userID && call.CalleeID != userID {
		return utils.NewAppError(utils.ErrCodePermissionDenied, "无权限结束此通话")
	}

	// 更新通话状态
	call.Mutex.Lock()
	call.Status = "ended"
	now := time.Now()
	call.EndTime = &now
	call.Duration = int64(now.Sub(call.StartTime).Seconds())
	call.Mutex.Unlock()

	// 关闭所有连接
	for _, conn := range call.Connections {
		if conn.Conn != nil {
			conn.Conn.Close()
		}
		if conn.PC != nil {
			conn.PC.Close()
		}
	}

	// 记录通话结束
	if err := s.recordCallEnd(call, reason); err != nil {
		logrus.WithError(err).Error("记录通话结束失败")
	}

	// 移除活跃通话
	s.callMutex.Lock()
	delete(s.activeCalls, callID)
	s.callMutex.Unlock()

	return nil
}

// HandleWebRTCMessage 处理 WebRTC 消息
func (s *WebRTCService) HandleWebRTCMessage(callID string, userID uint, message map[string]interface{}) error {
	s.callMutex.RLock()
	call, exists := s.activeCalls[callID]
	s.callMutex.RUnlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	messageType, ok := message["type"].(string)
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "消息类型无效")
	}

	switch messageType {
	case "offer":
		return s.handleOffer(call, userID, message)
	case "answer":
		return s.handleAnswer(call, userID, message)
	case "ice-candidate":
		return s.handleICECandidate(call, userID, message)
	case "network-stats":
		return s.handleNetworkStats(call, userID, message)
	case "quality-report":
		return s.handleQualityReport(call, userID, message)
	case "mute":
		return s.handleMute(call, userID, message)
	case "video-toggle":
		return s.handleVideoToggle(call, userID, message)
	default:
		return utils.NewAppError(utils.ErrCodeInvalidParams, "未知的消息类型")
	}
}

// handleOffer 处理 Offer
func (s *WebRTCService) handleOffer(call *CallSession, userID uint, message map[string]interface{}) error {
	// 获取 SDP
	sdp, ok := message["sdp"].(string)
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "SDP 无效")
	}

	// 转发给其他参与者
	return s.broadcastToCall(call, userID, map[string]interface{}{
		"type": "offer",
		"from": userID,
		"sdp":  sdp,
	})
}

// handleAnswer 处理 Answer
func (s *WebRTCService) handleAnswer(call *CallSession, userID uint, message map[string]interface{}) error {
	// 获取 SDP
	sdp, ok := message["sdp"].(string)
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "SDP 无效")
	}

	// 转发给其他参与者
	return s.broadcastToCall(call, userID, map[string]interface{}{
		"type": "answer",
		"from": userID,
		"sdp":  sdp,
	})
}

// handleICECandidate 处理 ICE Candidate
func (s *WebRTCService) handleICECandidate(call *CallSession, userID uint, message map[string]interface{}) error {
	candidate, ok := message["candidate"].(map[string]interface{})
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "ICE Candidate 无效")
	}

	// 转发给其他参与者
	return s.broadcastToCall(call, userID, map[string]interface{}{
		"type":      "ice-candidate",
		"from":      userID,
		"candidate": candidate,
	})
}

// handleNetworkStats 处理网络统计
func (s *WebRTCService) handleNetworkStats(call *CallSession, userID uint, message map[string]interface{}) error {
	stats, ok := message["stats"].(map[string]interface{})
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "网络统计无效")
	}

	// 更新网络统计
	call.Mutex.Lock()
	networkStats := &NetworkStats{
		UserID:           userID,
		RTT:              int(stats["rtt"].(float64)),
		PacketLoss:       stats["packetLoss"].(float64),
		Jitter:           stats["jitter"].(float64),
		Bandwidth:        int(stats["bandwidth"].(float64)),
		AvailableBitrate: int(stats["availableBitrate"].(float64)),
		NetworkType:      stats["networkType"].(string),
		SignalStrength:   int(stats["signalStrength"].(float64)),
		IsStable:         stats["isStable"].(bool),
		LastUpdate:       time.Now(),
	}
	call.NetworkStats[userID] = networkStats
	call.Mutex.Unlock()

	// 根据网络质量调整通话参数
	s.adaptToNetworkQuality(call, userID, networkStats)

	return nil
}

// handleQualityReport 处理质量报告
func (s *WebRTCService) handleQualityReport(call *CallSession, userID uint, message map[string]interface{}) error {
	quality, ok := message["quality"].(map[string]interface{})
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "质量报告无效")
	}

	// 更新通话质量
	call.Mutex.Lock()
	call.Quality.AudioQuality = quality["audioQuality"].(float64)
	call.Quality.VideoQuality = quality["videoQuality"].(float64)
	call.Quality.Latency = int(quality["latency"].(float64))
	call.Quality.PacketLoss = quality["packetLoss"].(float64)
	call.Quality.Jitter = quality["jitter"].(float64)
	call.Quality.Bitrate = int(quality["bitrate"].(float64))
	call.Quality.LastUpdate = time.Now()

	// 添加质量历史记录
	snapshot := QualitySnapshot{
		Timestamp: time.Now(),
		Score:     call.Quality.OverallScore,
		Latency:   call.Quality.Latency,
		Bitrate:   call.Quality.Bitrate,
	}
	call.Quality.QualityHistory = append(call.Quality.QualityHistory, snapshot)
	
	// 限制历史记录数量
	if len(call.Quality.QualityHistory) > 100 {
		call.Quality.QualityHistory = call.Quality.QualityHistory[1:]
	}
	call.Mutex.Unlock()

	return nil
}

// handleMute 处理静音
func (s *WebRTCService) handleMute(call *CallSession, userID uint, message map[string]interface{}) error {
	muted, ok := message["muted"].(bool)
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "静音状态无效")
	}

	// 更新静音状态
	call.Mutex.Lock()
	if conn, exists := call.Connections[userID]; exists {
		conn.IsMuted = muted
	}
	call.Mutex.Unlock()

	// 广播静音状态
	return s.broadcastToCall(call, userID, map[string]interface{}{
		"type":   "mute",
		"from":   userID,
		"muted":  muted,
	})
}

// handleVideoToggle 处理视频开关
func (s *WebRTCService) handleVideoToggle(call *CallSession, userID uint, message map[string]interface{}) error {
	videoOff, ok := message["videoOff"].(bool)
	if !ok {
		return utils.NewAppError(utils.ErrCodeInvalidParams, "视频状态无效")
	}

	// 更新视频状态
	call.Mutex.Lock()
	if conn, exists := call.Connections[userID]; exists {
		conn.IsVideoOff = videoOff
	}
	call.Mutex.Unlock()

	// 广播视频状态
	return s.broadcastToCall(call, userID, map[string]interface{}{
		"type":     "video-toggle",
		"from":     userID,
		"videoOff": videoOff,
	})
}

// broadcastToCall 广播消息给通话中的所有参与者
func (s *WebRTCService) broadcastToCall(call *CallSession, fromUserID uint, message map[string]interface{}) error {
	call.Mutex.RLock()
	defer call.Mutex.RUnlock()

	for userID, conn := range call.Connections {
		if userID != fromUserID && conn.Conn != nil {
			if err := conn.Conn.WriteJSON(message); err != nil {
				logrus.WithError(err).Error("广播消息失败")
			}
		}
	}

	return nil
}

// 辅助方法

// hasActiveCall 检查用户是否有活跃通话
func (s *WebRTCService) hasActiveCall(userID uint) bool {
	s.callMutex.RLock()
	defer s.callMutex.RUnlock()

	for _, call := range s.activeCalls {
		if (call.CallerID == userID || call.CalleeID == userID) && call.Status != "ended" {
			return true
		}
	}
	return false
}

// generateCallID 生成通话ID
func (s *WebRTCService) generateCallID() string {
	return fmt.Sprintf("call_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// getICEServers 获取 ICE 服务器配置
func (s *WebRTCService) getICEServers() []ICEServer {
	// 这里应该从配置文件或环境变量获取真实的 STUN/TURN 服务器
	return []ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
		{
			URLs: []string{"stun:stun1.l.google.com:19302"},
		},
	}
}

// adaptToNetworkQuality 根据网络质量调整通话参数
func (s *WebRTCService) adaptToNetworkQuality(call *CallSession, userID uint, stats *NetworkStats) {
	// 根据网络质量调整码率和分辨率
	adaptation := s.bandwidthAdaptor.Adapt(stats)
	
	// 发送调整建议给客户端
	message := map[string]interface{}{
		"type": "quality-adaptation",
		"adaptation": map[string]interface{}{
			"videoBitrate": adaptation.VideoBitrate,
			"audioBitrate": adaptation.AudioBitrate,
			"resolution":   adaptation.Resolution,
			"frameRate":    adaptation.FrameRate,
			"codec":        adaptation.Codec,
		},
	}

	s.broadcastToCall(call, userID, message)
}

// monitorCallQuality 监控通话质量
func (s *WebRTCService) monitorCallQuality(call *CallSession) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			call.Mutex.RLock()
			status := call.Status
			call.Mutex.RUnlock()

			if status == "ended" {
				return
			}

			// 计算整体通话质量
			s.calculateOverallQuality(call)
		}
	}
}

// calculateOverallQuality 计算整体通话质量
func (s *WebRTCService) calculateOverallQuality(call *CallSession) {
	call.Mutex.Lock()
	defer call.Mutex.Unlock()

	if call.Quality == nil {
		call.Quality = &CallQuality{}
	}

	// 基于网络统计计算质量分数
	var totalScore float64
	var count int

	for _, stats := range call.NetworkStats {
		score := s.networkMonitor.CalculateQualityScore(stats)
		totalScore += score
		count++
	}

	if count > 0 {
		call.Quality.OverallScore = totalScore / float64(count)
		call.Quality.NetworkQuality = call.Quality.OverallScore
	}
}

// 数据库记录方法

// recordCallStart 记录通话开始
func (s *WebRTCService) recordCallStart(call *CallSession) error {
	callRecord := model.Call{
		CallID:   call.ID,
		CallerID: call.CallerID,
		CalleeID: call.CalleeID,
		Type:     call.Type,
		Status:   call.Status,
		StartTime: call.StartTime,
	}

	return s.db.Create(&callRecord).Error
}

// recordCallAnswer 记录通话接听
func (s *WebRTCService) recordCallAnswer(call *CallSession) error {
	return s.db.Model(&model.Call{}).
		Where("call_id = ?", call.ID).
		Update("status", "connected").Error
}

// recordCallEnd 记录通话结束
func (s *WebRTCService) recordCallEnd(call *CallSession, reason string) error {
	updates := map[string]interface{}{
		"status":   "ended",
		"end_time": call.EndTime,
		"duration": call.Duration,
		"reason":   reason,
	}

	if call.Quality != nil {
		qualityJSON, _ := json.Marshal(call.Quality)
		updates["quality_stats"] = string(qualityJSON)
	}

	return s.db.Model(&model.Call{}).
		Where("call_id = ?", call.ID).
		Updates(updates).Error
}
