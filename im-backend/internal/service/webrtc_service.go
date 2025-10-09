package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"zhihang-messenger/im-backend/config"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WebRTCService WebRTC通话服务
type WebRTCService struct {
	db          *gorm.DB
	activeCalls map[string]*CallSession
	callMutex   sync.RWMutex
	ctx         context.Context
}

// CallSession 通话会话
type CallSession struct {
	ID            string                   `json:"id"`
	CallerID      uint                     `json:"caller_id"`
	CalleeID      uint                     `json:"callee_id"`
	Type          string                   `json:"type"`   // audio, video, screen_share
	Status        string                   `json:"status"` // ringing, connecting, connected, ended
	StartTime     time.Time                `json:"start_time"`
	EndTime       *time.Time               `json:"end_time,omitempty"`
	Duration      int64                    `json:"duration,omitempty"`
	Peers         map[uint]*PeerConnection `json:"peers"`
	ScreenSharing *ScreenShareInfo         `json:"screen_sharing,omitempty"` // 屏幕共享信息
	Mutex         sync.RWMutex             `json:"-"`
}

// ScreenShareInfo 屏幕共享信息
type ScreenShareInfo struct {
	SharerID   uint      `json:"sharer_id"`   // 共享者用户ID
	SharerName string    `json:"sharer_name"` // 共享者名称
	IsActive   bool      `json:"is_active"`   // 是否正在共享
	StartTime  time.Time `json:"start_time"`  // 开始时间
	Quality    string    `json:"quality"`     // 质量: high, medium, low
	WithAudio  bool      `json:"with_audio"`  // 是否包含音频
}

// PeerConnection 对等连接
type PeerConnection struct {
	UserID          uint                   `json:"user_id"`
	Conn            *websocket.Conn        `json:"-"`
	PC              *webrtc.PeerConnection `json:"-"`
	ScreenSharePC   *webrtc.PeerConnection `json:"-"` // 屏幕共享专用连接
	IsMuted         bool                   `json:"is_muted"`
	IsVideoOff      bool                   `json:"is_video_off"`
	IsScreenSharing bool                   `json:"is_screen_sharing"` // 是否正在共享屏幕
	JoinTime        time.Time              `json:"join_time"`
}

// NewWebRTCService 创建WebRTC服务
func NewWebRTCService() *WebRTCService {
	return &WebRTCService{
		db:          config.DB,
		activeCalls: make(map[string]*CallSession),
		ctx:         context.Background(),
	}
}

// CreateCall 创建通话
func (s *WebRTCService) CreateCall(callerID, calleeID uint, callType string) (*CallSession, error) {
	s.callMutex.Lock()
	defer s.callMutex.Unlock()

	callID := fmt.Sprintf("call_%d_%d_%d", callerID, calleeID, time.Now().Unix())

	session := &CallSession{
		ID:        callID,
		CallerID:  callerID,
		CalleeID:  calleeID,
		Type:      callType,
		Status:    "ringing",
		StartTime: time.Now(),
		Peers:     make(map[uint]*PeerConnection),
	}

	s.activeCalls[callID] = session

	logrus.Infof("创建通话: %s, 类型: %s, 呼叫者: %d, 被叫者: %d", callID, callType, callerID, calleeID)

	return session, nil
}

// JoinCall 加入通话
func (s *WebRTCService) JoinCall(callID string, userID uint, conn *websocket.Conn) error {
	s.callMutex.RLock()
	session, exists := s.activeCalls[callID]
	s.callMutex.RUnlock()

	if !exists {
		return fmt.Errorf("通话不存在: %s", callID)
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// 创建WebRTC配置
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// 创建PeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("创建PeerConnection失败: %w", err)
	}

	peer := &PeerConnection{
		UserID:   userID,
		Conn:     conn,
		PC:       pc,
		JoinTime: time.Now(),
	}

	session.Peers[userID] = peer
	session.Status = "connected"

	logrus.Infof("用户 %d 加入通话 %s", userID, callID)

	return nil
}

// EndCall 结束通话
func (s *WebRTCService) EndCall(callID string) error {
	s.callMutex.Lock()
	defer s.callMutex.Unlock()

	session, exists := s.activeCalls[callID]
	if !exists {
		return fmt.Errorf("通话不存在: %s", callID)
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// 关闭所有PeerConnection
	for _, peer := range session.Peers {
		if peer.PC != nil {
			peer.PC.Close()
		}
	}

	// 计算通话时长
	endTime := time.Now()
	session.EndTime = &endTime
	session.Duration = int64(endTime.Sub(session.StartTime).Seconds())
	session.Status = "ended"

	// 从活跃通话列表中移除
	delete(s.activeCalls, callID)

	logrus.Infof("通话结束: %s, 时长: %d秒", callID, session.Duration)

	return nil
}

// GetActiveCall 获取活跃通话
func (s *WebRTCService) GetActiveCall(callID string) (*CallSession, error) {
	s.callMutex.RLock()
	defer s.callMutex.RUnlock()

	session, exists := s.activeCalls[callID]
	if !exists {
		return nil, fmt.Errorf("通话不存在: %s", callID)
	}

	return session, nil
}

// GetActiveCalls 获取所有活跃通话
func (s *WebRTCService) GetActiveCalls() []*CallSession {
	s.callMutex.RLock()
	defer s.callMutex.RUnlock()

	calls := make([]*CallSession, 0, len(s.activeCalls))
	for _, session := range s.activeCalls {
		calls = append(calls, session)
	}

	return calls
}

// HandleSignaling 处理信令消息
func (s *WebRTCService) HandleSignaling(callID string, userID uint, signalType string, payload interface{}) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.RLock()
	peer, exists := session.Peers[userID]
	session.Mutex.RUnlock()

	if !exists {
		return fmt.Errorf("用户未加入通话")
	}

	// 根据信令类型处理
	switch signalType {
	case "offer":
		return s.handleOffer(peer, payload)
	case "answer":
		return s.handleAnswer(peer, payload)
	case "ice_candidate":
		return s.handleICECandidate(peer, payload)
	case "screen_share_offer":
		return s.handleScreenShareOffer(peer, payload)
	case "screen_share_answer":
		return s.handleScreenShareAnswer(peer, payload)
	case "screen_share_ice_candidate":
		return s.handleScreenShareICECandidate(peer, payload)
	default:
		return fmt.Errorf("未知的信令类型: %s", signalType)
	}
}

// handleOffer 处理Offer
func (s *WebRTCService) handleOffer(peer *PeerConnection, payload interface{}) error {
	offerData, _ := json.Marshal(payload)
	var offer webrtc.SessionDescription
	if err := json.Unmarshal(offerData, &offer); err != nil {
		return err
	}

	if err := peer.PC.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("设置远程描述失败: %w", err)
	}

	return nil
}

// handleAnswer 处理Answer
func (s *WebRTCService) handleAnswer(peer *PeerConnection, payload interface{}) error {
	answerData, _ := json.Marshal(payload)
	var answer webrtc.SessionDescription
	if err := json.Unmarshal(answerData, &answer); err != nil {
		return err
	}

	if err := peer.PC.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("设置远程描述失败: %w", err)
	}

	return nil
}

// handleICECandidate 处理ICE候选
func (s *WebRTCService) handleICECandidate(peer *PeerConnection, payload interface{}) error {
	candidateData, _ := json.Marshal(payload)
	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal(candidateData, &candidate); err != nil {
		return err
	}

	if err := peer.PC.AddICECandidate(candidate); err != nil {
		return fmt.Errorf("添加ICE候选失败: %w", err)
	}

	return nil
}

// ToggleMute 切换静音
func (s *WebRTCService) ToggleMute(callID string, userID uint) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	peer, exists := session.Peers[userID]
	if !exists {
		return fmt.Errorf("用户未加入通话")
	}

	peer.IsMuted = !peer.IsMuted
	logrus.Infof("用户 %d 切换静音状态: %v", userID, peer.IsMuted)

	return nil
}

// ToggleVideo 切换视频
func (s *WebRTCService) ToggleVideo(callID string, userID uint) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	peer, exists := session.Peers[userID]
	if !exists {
		return fmt.Errorf("用户未加入通话")
	}

	peer.IsVideoOff = !peer.IsVideoOff
	logrus.Infof("用户 %d 切换视频状态: %v", userID, peer.IsVideoOff)

	return nil
}

// GetCallStats 获取通话统计
func (s *WebRTCService) GetCallStats(callID string) (map[string]interface{}, error) {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return nil, err
	}

	session.Mutex.RLock()
	defer session.Mutex.RUnlock()

	stats := map[string]interface{}{
		"call_id":        session.ID,
		"type":           session.Type,
		"status":         session.Status,
		"start_time":     session.StartTime,
		"peer_count":     len(session.Peers),
		"duration":       time.Since(session.StartTime).Seconds(),
		"screen_sharing": session.ScreenSharing,
	}

	return stats, nil
}

// StartScreenShare 开始屏幕共享
func (s *WebRTCService) StartScreenShare(callID string, userID uint, userName string, quality string, withAudio bool) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// 检查用户是否在通话中
	peer, exists := session.Peers[userID]
	if !exists {
		return fmt.Errorf("用户未加入通话")
	}

	// 检查是否已有人在共享
	if session.ScreenSharing != nil && session.ScreenSharing.IsActive {
		return fmt.Errorf("已有用户正在共享屏幕，共享者: %s", session.ScreenSharing.SharerName)
	}

	// 创建屏幕共享专用PeerConnection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	screenSharePC, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return fmt.Errorf("创建屏幕共享连接失败: %w", err)
	}

	peer.ScreenSharePC = screenSharePC
	peer.IsScreenSharing = true

	// 设置屏幕共享信息
	session.ScreenSharing = &ScreenShareInfo{
		SharerID:   userID,
		SharerName: userName,
		IsActive:   true,
		StartTime:  time.Now(),
		Quality:    quality,
		WithAudio:  withAudio,
	}

	logrus.Infof("用户 %d (%s) 开始屏幕共享，通话: %s, 质量: %s, 音频: %v",
		userID, userName, callID, quality, withAudio)

	return nil
}

// StopScreenShare 停止屏幕共享
func (s *WebRTCService) StopScreenShare(callID string, userID uint) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// 检查是否是共享者
	if session.ScreenSharing == nil || session.ScreenSharing.SharerID != userID {
		return fmt.Errorf("您没有正在进行的屏幕共享")
	}

	peer, exists := session.Peers[userID]
	if !exists {
		return fmt.Errorf("用户未加入通话")
	}

	// 关闭屏幕共享连接
	if peer.ScreenSharePC != nil {
		peer.ScreenSharePC.Close()
		peer.ScreenSharePC = nil
	}

	peer.IsScreenSharing = false

	// 清除屏幕共享信息
	duration := time.Since(session.ScreenSharing.StartTime).Seconds()
	session.ScreenSharing = nil

	logrus.Infof("用户 %d 停止屏幕共享，通话: %s, 共享时长: %.0f秒", userID, callID, duration)

	return nil
}

// GetScreenShareStatus 获取屏幕共享状态
func (s *WebRTCService) GetScreenShareStatus(callID string) (*ScreenShareInfo, error) {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return nil, err
	}

	session.Mutex.RLock()
	defer session.Mutex.RUnlock()

	return session.ScreenSharing, nil
}

// handleScreenShareOffer 处理屏幕共享Offer
func (s *WebRTCService) handleScreenShareOffer(peer *PeerConnection, payload interface{}) error {
	offerData, _ := json.Marshal(payload)
	var offer webrtc.SessionDescription
	if err := json.Unmarshal(offerData, &offer); err != nil {
		return err
	}

	// 如果还没有屏幕共享连接，创建一个
	if peer.ScreenSharePC == nil {
		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		pc, err := webrtc.NewPeerConnection(config)
		if err != nil {
			return fmt.Errorf("创建屏幕共享连接失败: %w", err)
		}
		peer.ScreenSharePC = pc
	}

	if err := peer.ScreenSharePC.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("设置屏幕共享远程描述失败: %w", err)
	}

	return nil
}

// handleScreenShareAnswer 处理屏幕共享Answer
func (s *WebRTCService) handleScreenShareAnswer(peer *PeerConnection, payload interface{}) error {
	answerData, _ := json.Marshal(payload)
	var answer webrtc.SessionDescription
	if err := json.Unmarshal(answerData, &answer); err != nil {
		return err
	}

	if peer.ScreenSharePC == nil {
		return fmt.Errorf("屏幕共享连接不存在")
	}

	if err := peer.ScreenSharePC.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("设置屏幕共享远程描述失败: %w", err)
	}

	return nil
}

// handleScreenShareICECandidate 处理屏幕共享ICE候选
func (s *WebRTCService) handleScreenShareICECandidate(peer *PeerConnection, payload interface{}) error {
	candidateData, _ := json.Marshal(payload)
	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal(candidateData, &candidate); err != nil {
		return err
	}

	if peer.ScreenSharePC == nil {
		return fmt.Errorf("屏幕共享连接不存在")
	}

	if err := peer.ScreenSharePC.AddICECandidate(candidate); err != nil {
		return fmt.Errorf("添加屏幕共享ICE候选失败: %w", err)
	}

	return nil
}

// ChangeScreenShareQuality 更改屏幕共享质量
func (s *WebRTCService) ChangeScreenShareQuality(callID string, userID uint, quality string) error {
	session, err := s.GetActiveCall(callID)
	if err != nil {
		return err
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// 检查是否是共享者
	if session.ScreenSharing == nil || session.ScreenSharing.SharerID != userID {
		return fmt.Errorf("您没有正在进行的屏幕共享")
	}

	// 验证质量参数
	if quality != "high" && quality != "medium" && quality != "low" {
		return fmt.Errorf("无效的质量参数，应为: high, medium, low")
	}

	session.ScreenSharing.Quality = quality

	logrus.Infof("用户 %d 更改屏幕共享质量为: %s", userID, quality)

	return nil
}
