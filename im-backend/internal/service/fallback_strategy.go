package service

import (
	"fmt"
	"sync"
	"time"
	"zhihang-messenger/im-backend/internal/utils"

	"github.com/sirupsen/logrus"
)

// FallbackStrategy 降级策略服务
type FallbackStrategy struct {
	networkMonitor   *NetworkQualityMonitor
	bandwidthAdaptor *BandwidthAdaptor
	codecManager     *CodecManager
	activeCalls      map[string]*CallFallbackState
	mutex            sync.RWMutex
}

// CallFallbackState 通话降级状态
type CallFallbackState struct {
	CallID           string                    `json:"call_id"`
	CurrentQuality   string                    `json:"current_quality"`
	FallbackLevel    int                       `json:"fallback_level"`
	QualityHistory   []QualityTransition       `json:"quality_history"`
	NetworkHistory   []NetworkQualitySnapshot  `json:"network_history"`
	LastFallback     time.Time                 `json:"last_fallback"`
	FallbackCount    int                       `json:"fallback_count"`
	RecoveryAttempts int                       `json:"recovery_attempts"`
	IsRecovering     bool                      `json:"is_recovering"`
	FallbackReason   string                    `json:"fallback_reason"`
}

// QualityTransition 质量转换记录
type QualityTransition struct {
	Timestamp    time.Time `json:"timestamp"`
	FromQuality  string    `json:"from_quality"`
	ToQuality    string    `json:"to_quality"`
	Reason       string    `json:"reason"`
	NetworkStats *NetworkStats `json:"network_stats"`
	Success      bool      `json:"success"`
}

// NetworkQualitySnapshot 网络质量快照
type NetworkQualitySnapshot struct {
	Timestamp     time.Time `json:"timestamp"`
	RTT           int       `json:"rtt"`
	PacketLoss    float64   `json:"packet_loss"`
	Jitter        float64   `json:"jitter"`
	Bandwidth     int       `json:"bandwidth"`
	QualityScore  float64   `json:"quality_score"`
	QualityLevel  string    `json:"quality_level"`
}

// FallbackLevel 降级等级
type FallbackLevel int

const (
	LevelHighQuality FallbackLevel = iota    // 高质量
	LevelMediumQuality                       // 中等质量
	LevelLowQuality                          // 低质量
	LevelVeryLowQuality                      // 很低质量
	LevelAudioOnly                           // 仅音频
	LevelTextOnly                            // 仅文本
	LevelDisconnected                        // 断开连接
)

// FallbackReason 降级原因
type FallbackReason string

const (
	ReasonHighLatency     FallbackReason = "high_latency"
	ReasonHighPacketLoss  FallbackReason = "high_packet_loss"
	ReasonLowBandwidth    FallbackReason = "low_bandwidth"
	ReasonHighJitter      FallbackReason = "high_jitter"
	ReasonNetworkUnstable FallbackReason = "network_unstable"
	ReasonServerOverload  FallbackReason = "server_overload"
	ReasonUserRequest     FallbackReason = "user_request"
)

// QualityPreset 质量预设
type QualityPreset struct {
	Level           FallbackLevel `json:"level"`
	Name            string        `json:"name"`
	VideoBitrate    int           `json:"video_bitrate"`
	AudioBitrate    int           `json:"audio_bitrate"`
	Resolution      string        `json:"resolution"`
	FrameRate       int           `json:"frame_rate"`
	AudioCodec      string        `json:"audio_codec"`
	VideoCodec      string        `json:"video_codec"`
	Enabled         bool          `json:"enabled"`
	MinBandwidth    int           `json:"min_bandwidth"`
	MaxLatency      int           `json:"max_latency"`
	MaxPacketLoss   float64       `json:"max_packet_loss"`
	Description     string        `json:"description"`
}

// NewFallbackStrategy 创建降级策略服务
func NewFallbackStrategy(
	networkMonitor *NetworkQualityMonitor,
	bandwidthAdaptor *BandwidthAdaptor,
	codecManager *CodecManager,
) *FallbackStrategy {
	return &FallbackStrategy{
		networkMonitor:   networkMonitor,
		bandwidthAdaptor: bandwidthAdaptor,
		codecManager:     codecManager,
		activeCalls:      make(map[string]*CallFallbackState),
	}
}

// RegisterCall 注册通话
func (fs *FallbackStrategy) RegisterCall(callID string, initialQuality string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.activeCalls[callID] = &CallFallbackState{
		CallID:         callID,
		CurrentQuality: initialQuality,
		FallbackLevel:  int(LevelHighQuality),
		QualityHistory: make([]QualityTransition, 0),
		NetworkHistory: make([]NetworkQualitySnapshot, 0),
		LastFallback:   time.Now(),
		FallbackCount:  0,
		RecoveryAttempts: 0,
		IsRecovering:   false,
		FallbackReason: "",
	}

	logrus.WithFields(logrus.Fields{
		"call_id": callID,
		"quality": initialQuality,
	}).Info("通话降级策略注册成功")

	return nil
}

// UnregisterCall 注销通话
func (fs *FallbackStrategy) UnregisterCall(callID string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	delete(fs.activeCalls, callID)

	logrus.WithField("call_id", callID).Info("通话降级策略注销")
}

// EvaluateAndFallback 评估并执行降级
func (fs *FallbackStrategy) EvaluateAndFallback(callID string, networkStats *NetworkStats) (*QualityPreset, error) {
	fs.mutex.RLock()
	callState, exists := fs.activeCalls[callID]
	fs.mutex.RUnlock()

	if !exists {
		return nil, utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	// 记录网络统计
	fs.recordNetworkStats(callState, networkStats)

	// 评估网络质量
	qualityLevel := fs.assessNetworkQuality(networkStats)
	fallbackReason := fs.determineFallbackReason(networkStats, qualityLevel)

	// 检查是否需要降级
	shouldFallback, targetLevel := fs.shouldFallback(callState, networkStats, qualityLevel)

	if shouldFallback {
		return fs.executeFallback(callState, targetLevel, fallbackReason, networkStats)
	}

	// 检查是否可以恢复
	if fs.canRecover(callState, networkStats) {
		return fs.attemptRecovery(callState, networkStats)
	}

	// 返回当前质量设置
	return fs.getQualityPreset(callState.CurrentQuality), nil
}

// recordNetworkStats 记录网络统计
func (fs *FallbackStrategy) recordNetworkStats(callState *CallFallbackState, stats *NetworkStats) {
	snapshot := NetworkQualitySnapshot{
		Timestamp:    time.Now(),
		RTT:          stats.RTT,
		PacketLoss:   stats.PacketLoss,
		Jitter:       stats.Jitter,
		Bandwidth:    stats.Bandwidth,
		QualityScore: fs.networkMonitor.CalculateQualityScore(stats),
		QualityLevel: fs.getQualityLevelName(fs.assessNetworkQuality(stats)),
	}

	callState.NetworkHistory = append(callState.NetworkHistory, snapshot)

	// 限制历史记录数量
	if len(callState.NetworkHistory) > 50 {
		callState.NetworkHistory = callState.NetworkHistory[1:]
	}
}

// assessNetworkQuality 评估网络质量
func (fs *FallbackStrategy) assessNetworkQuality(stats *NetworkStats) NetworkQualityLevel {
	return fs.networkMonitor.GetQualityLevel(
		fs.networkMonitor.CalculateQualityScore(stats),
	)
}

// determineFallbackReason 确定降级原因
func (fs *FallbackStrategy) determineFallbackReason(stats *NetworkStats, quality NetworkQualityLevel) FallbackReason {
	// 根据网络条件确定降级原因
	if stats.RTT > 300 {
		return ReasonHighLatency
	}
	if stats.PacketLoss > 8 {
		return ReasonHighPacketLoss
	}
	if stats.Bandwidth < 800 {
		return ReasonLowBandwidth
	}
	if stats.Jitter > 60 {
		return ReasonHighJitter
	}
	if !stats.IsStable {
		return ReasonNetworkUnstable
	}

	return ReasonUserRequest
}

// shouldFallback 判断是否需要降级
func (fs *FallbackStrategy) shouldFallback(callState *CallFallbackState, stats *NetworkStats, quality NetworkQualityLevel) (bool, FallbackLevel) {
	currentLevel := FallbackLevel(callState.FallbackLevel)

	// 根据网络质量确定目标降级等级
	var targetLevel FallbackLevel
	switch quality {
	case QualityExcellent, QualityGood:
		// 网络质量好，不需要降级
		return false, currentLevel
	case QualityFair:
		targetLevel = LevelMediumQuality
	case QualityPoor:
		targetLevel = LevelLowQuality
	case QualityVeryPoor:
		targetLevel = LevelAudioOnly
	default:
		targetLevel = LevelAudioOnly
	}

	// 如果目标等级比当前等级低，需要降级
	if targetLevel > currentLevel {
		return true, targetLevel
	}

	// 检查特定条件
	if stats.RTT > 500 && currentLevel < LevelAudioOnly {
		return true, LevelAudioOnly
	}
	if stats.PacketLoss > 15 && currentLevel < LevelVeryLowQuality {
		return true, LevelVeryLowQuality
	}
	if stats.Bandwidth < 500 && currentLevel < LevelLowQuality {
		return true, LevelLowQuality
	}

	return false, currentLevel
}

// executeFallback 执行降级
func (fs *FallbackStrategy) executeFallback(callState *CallFallbackState, targetLevel FallbackLevel, reason FallbackReason, stats *NetworkStats) (*QualityPreset, error) {
	oldQuality := callState.CurrentQuality
	newQuality := fs.getQualityNameByLevel(targetLevel)

	// 创建质量转换记录
	transition := QualityTransition{
		Timestamp:    time.Now(),
		FromQuality:  oldQuality,
		ToQuality:    newQuality,
		Reason:       string(reason),
		NetworkStats: stats,
		Success:      true,
	}

	// 更新通话状态
	callState.CurrentQuality = newQuality
	callState.FallbackLevel = int(targetLevel)
	callState.LastFallback = time.Now()
	callState.FallbackCount++
	callState.FallbackReason = string(reason)
	callState.IsRecovering = false
	callState.RecoveryAttempts = 0

	// 记录质量转换历史
	callState.QualityHistory = append(callState.QualityHistory, transition)
	if len(callState.QualityHistory) > 20 {
		callState.QualityHistory = callState.QualityHistory[1:]
	}

	// 获取质量预设
	preset := fs.getQualityPreset(newQuality)

	logrus.WithFields(logrus.Fields{
		"call_id":     callState.CallID,
		"old_quality": oldQuality,
		"new_quality": newQuality,
		"reason":      reason,
		"fallback_count": callState.FallbackCount,
	}).Info("执行通话降级")

	return preset, nil
}

// canRecover 判断是否可以恢复
func (fs *FallbackStrategy) canRecover(callState *CallFallbackState, stats *NetworkStats) bool {
	// 如果正在恢复中，不再尝试恢复
	if callState.IsRecovering {
		return false
	}

	// 如果降级次数为0，不需要恢复
	if callState.FallbackCount == 0 {
		return false
	}

	// 检查网络质量是否改善
	quality := fs.assessNetworkQuality(stats)
	if quality >= QualityGood {
		return true
	}

	// 检查特定条件是否改善
	if stats.RTT < 100 && stats.PacketLoss < 2 && stats.Bandwidth > 1500 {
		return true
	}

	return false
}

// attemptRecovery 尝试恢复
func (fs *FallbackStrategy) attemptRecovery(callState *CallFallbackState, stats *NetworkStats) (*QualityPreset, error) {
	callState.IsRecovering = true
	callState.RecoveryAttempts++

	// 确定恢复目标
	var targetLevel FallbackLevel
	quality := fs.assessNetworkQuality(stats)

	switch quality {
	case QualityExcellent:
		targetLevel = LevelHighQuality
	case QualityGood:
		targetLevel = LevelMediumQuality
	case QualityFair:
		targetLevel = LevelLowQuality
	default:
		// 网络质量仍然不好，取消恢复
		callState.IsRecovering = false
		return fs.getQualityPreset(callState.CurrentQuality), nil
	}

	oldQuality := callState.CurrentQuality
	newQuality := fs.getQualityNameByLevel(targetLevel)

	// 创建恢复记录
	transition := QualityTransition{
		Timestamp:    time.Now(),
		FromQuality:  oldQuality,
		ToQuality:    newQuality,
		Reason:       "network_recovery",
		NetworkStats: stats,
		Success:      true,
	}

	// 更新通话状态
	callState.CurrentQuality = newQuality
	callState.FallbackLevel = int(targetLevel)
	callState.IsRecovering = false
	callState.RecoveryAttempts = 0

	// 记录恢复历史
	callState.QualityHistory = append(callState.QualityHistory, transition)

	// 获取质量预设
	preset := fs.getQualityPreset(newQuality)

	logrus.WithFields(logrus.Fields{
		"call_id":     callState.CallID,
		"old_quality": oldQuality,
		"new_quality": newQuality,
		"recovery_attempt": callState.RecoveryAttempts,
	}).Info("执行通话恢复")

	return preset, nil
}

// getQualityNameByLevel 根据等级获取质量名称
func (fs *FallbackStrategy) getQualityNameByLevel(level FallbackLevel) string {
	switch level {
	case LevelHighQuality:
		return "high"
	case LevelMediumQuality:
		return "medium"
	case LevelLowQuality:
		return "low"
	case LevelVeryLowQuality:
		return "very_low"
	case LevelAudioOnly:
		return "audio_only"
	case LevelTextOnly:
		return "text_only"
	case LevelDisconnected:
		return "disconnected"
	default:
		return "medium"
	}
}

// getQualityPreset 获取质量预设
func (fs *FallbackStrategy) getQualityPreset(quality string) *QualityPreset {
	switch quality {
	case "high":
		return &QualityPreset{
			Level:        LevelHighQuality,
			Name:         "高质量",
			VideoBitrate: 4000,
			AudioBitrate: 192,
			Resolution:   "1920x1080",
			FrameRate:    30,
			AudioCodec:   "opus",
			VideoCodec:   "vp8",
			Enabled:      true,
			MinBandwidth: 3000,
			MaxLatency:   100,
			MaxPacketLoss: 2,
			Description:  "高清视频通话，适合稳定网络环境",
		}
	case "medium":
		return &QualityPreset{
			Level:        LevelMediumQuality,
			Name:         "中等质量",
			VideoBitrate: 2000,
			AudioBitrate: 128,
			Resolution:   "1280x720",
			FrameRate:    30,
			AudioCodec:   "opus",
			VideoCodec:   "vp8",
			Enabled:      true,
			MinBandwidth: 1500,
			MaxLatency:   150,
			MaxPacketLoss: 5,
			Description:  "标清视频通话，适合一般网络环境",
		}
	case "low":
		return &QualityPreset{
			Level:        LevelLowQuality,
			Name:         "低质量",
			VideoBitrate: 1000,
			AudioBitrate: 96,
			Resolution:   "854x480",
			FrameRate:    24,
			AudioCodec:   "opus",
			VideoCodec:   "vp8",
			Enabled:      true,
			MinBandwidth: 800,
			MaxLatency:   200,
			MaxPacketLoss: 8,
			Description:  "低清视频通话，适合较差网络环境",
		}
	case "very_low":
		return &QualityPreset{
			Level:        LevelVeryLowQuality,
			Name:         "很低质量",
			VideoBitrate: 500,
			AudioBitrate: 64,
			Resolution:   "640x360",
			FrameRate:    15,
			AudioCodec:   "opus",
			VideoCodec:   "vp8",
			Enabled:      true,
			MinBandwidth: 400,
			MaxLatency:   300,
			MaxPacketLoss: 12,
			Description:  "极低清视频通话，适合很差网络环境",
		}
	case "audio_only":
		return &QualityPreset{
			Level:        LevelAudioOnly,
			Name:         "仅音频",
			VideoBitrate: 0,
			AudioBitrate: 64,
			Resolution:   "0x0",
			FrameRate:    0,
			AudioCodec:   "opus",
			VideoCodec:   "",
			Enabled:      true,
			MinBandwidth: 100,
			MaxLatency:   500,
			MaxPacketLoss: 20,
			Description:  "纯音频通话，适合网络极差环境",
		}
	default:
		return fs.getQualityPreset("medium")
	}
}

// getQualityLevelName 获取质量等级名称
func (fs *FallbackStrategy) getQualityLevelName(level NetworkQualityLevel) string {
	return fs.networkMonitor.GetQualityLevelName(level)
}

// GetCallFallbackState 获取通话降级状态
func (fs *FallbackStrategy) GetCallFallbackState(callID string) (*CallFallbackState, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	callState, exists := fs.activeCalls[callID]
	if !exists {
		return nil, utils.NewAppError(utils.ErrCodeMessageNotFound, "通话不存在")
	}

	return callState, nil
}

// GetFallbackStatistics 获取降级统计
func (fs *FallbackStrategy) GetFallbackStatistics() map[string]interface{} {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	stats := map[string]interface{}{
		"active_calls":        len(fs.activeCalls),
		"total_fallbacks":     0,
		"total_recoveries":    0,
		"quality_distribution": make(map[string]int),
		"reason_distribution": make(map[string]int),
	}

	qualityDist := stats["quality_distribution"].(map[string]int)
	reasonDist := stats["reason_distribution"].(map[string]int)

	for _, callState := range fs.activeCalls {
		stats["total_fallbacks"] = stats["total_fallbacks"].(int) + callState.FallbackCount
		stats["total_recoveries"] = stats["total_recoveries"].(int) + callState.RecoveryAttempts

		qualityDist[callState.CurrentQuality]++
		if callState.FallbackReason != "" {
			reasonDist[callState.FallbackReason]++
		}
	}

	return stats
}

// GetFallbackRecommendations 获取降级建议
func (fs *FallbackStrategy) GetFallbackRecommendations(callID string) []string {
	fs.mutex.RLock()
	callState, exists := fs.activeCalls[callID]
	fs.mutex.RUnlock()

	if !exists {
		return []string{"通话不存在"}
	}

	var recommendations []string

	// 基于降级历史提供建议
	if callState.FallbackCount > 3 {
		recommendations = append(recommendations, "通话质量不稳定，建议检查网络连接")
	}

	// 基于网络历史提供建议
	if len(callState.NetworkHistory) > 0 {
		recent := callState.NetworkHistory[len(callState.NetworkHistory)-1]
		
		if recent.RTT > 200 {
			recommendations = append(recommendations, "网络延迟较高，建议使用有线网络")
		}
		
		if recent.PacketLoss > 5 {
			recommendations = append(recommendations, "网络丢包率较高，建议检查网络连接")
		}
		
		if recent.Bandwidth < 1000 {
			recommendations = append(recommendations, "网络带宽不足，建议关闭其他占用网络的应用程序")
		}
	}

	// 基于当前质量提供建议
	switch callState.CurrentQuality {
	case "audio_only":
		recommendations = append(recommendations, "当前为音频通话模式，建议改善网络环境以启用视频")
	case "very_low":
		recommendations = append(recommendations, "当前为极低质量模式，建议切换到更稳定的网络")
	case "low":
		recommendations = append(recommendations, "当前为低质量模式，建议优化网络环境")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "通话质量良好")
	}

	return recommendations
}

// CleanupOldData 清理旧数据
func (fs *FallbackStrategy) CleanupOldData(maxAge time.Duration) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	cutoffTime := time.Now().Add(-maxAge)

	for callID, callState := range fs.activeCalls {
		// 清理旧的网络历史
		var filteredNetworkHistory []NetworkQualitySnapshot
		for _, snapshot := range callState.NetworkHistory {
			if snapshot.Timestamp.After(cutoffTime) {
				filteredNetworkHistory = append(filteredNetworkHistory, snapshot)
			}
		}
		callState.NetworkHistory = filteredNetworkHistory

		// 清理旧的质量历史
		var filteredQualityHistory []QualityTransition
		for _, transition := range callState.QualityHistory {
			if transition.Timestamp.After(cutoffTime) {
				filteredQualityHistory = append(filteredQualityHistory, transition)
			}
		}
		callState.QualityHistory = filteredQualityHistory

		// 如果没有历史数据且通话已结束，删除记录
		if len(filteredNetworkHistory) == 0 && len(filteredQualityHistory) == 0 {
			delete(fs.activeCalls, callID)
		}
	}
}
