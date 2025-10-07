package service

import (
	"encoding/json"
	"sync"
	"time"
	"zhihang-messenger/im-backend/internal/utils"

	"github.com/sirupsen/logrus"
)

// CallQualityStatsService 通话质量统计服务
type CallQualityStatsService struct {
	db              *gorm.DB
	activeStats     map[string]*CallQualityStats
	statsMutex      sync.RWMutex
	aggregatedStats *AggregatedStats
	reportTimer     *time.Timer
}

// CallQualityStats 通话质量统计
type CallQualityStats struct {
	CallID           string                    `json:"call_id"`
	UserID           uint                      `json:"user_id"`
	StartTime        time.Time                 `json:"start_time"`
	EndTime          *time.Time                `json:"end_time,omitempty"`
	Duration         int64                     `json:"duration"` // 秒
	Status           string                    `json:"status"`
	
	// 网络质量统计
	NetworkStats     NetworkQualitySummary    `json:"network_stats"`
	
	// 音视频质量统计
	AudioStats       AudioQualitySummary      `json:"audio_stats"`
	VideoStats       VideoQualitySummary      `json:"video_stats"`
	
	// 用户体验统计
	UserExperience   UserExperienceSummary    `json:"user_experience"`
	
	// 质量事件
	QualityEvents    []QualityEvent           `json:"quality_events"`
	
	// 实时统计
	RealTimeStats    RealTimeStats            `json:"real_time_stats"`
	
	// 统计更新时间
	LastUpdate       time.Time                 `json:"last_update"`
}

// NetworkQualitySummary 网络质量摘要
type NetworkQualitySummary struct {
	AverageRTT         float64 `json:"average_rtt"`
	MaxRTT             int     `json:"max_rtt"`
	MinRTT             int     `json:"min_rtt"`
	AveragePacketLoss  float64 `json:"average_packet_loss"`
	MaxPacketLoss      float64 `json:"max_packet_loss"`
	AverageJitter      float64 `json:"average_jitter"`
	MaxJitter          float64 `json:"max_jitter"`
	AverageBandwidth   float64 `json:"average_bandwidth"`
	MinBandwidth       int     `json:"min_bandwidth"`
	NetworkStability   float64 `json:"network_stability"`
	QualityScore       float64 `json:"quality_score"`
	NetworkType        string  `json:"network_type"`
}

// AudioQualitySummary 音频质量摘要
type AudioQualitySummary struct {
	AverageBitrate     float64 `json:"average_bitrate"`
	AverageVolume      float64 `json:"average_volume"`
	AverageNoiseLevel  float64 `json:"average_noise_level"`
	EchoLevel          float64 `json:"echo_level"`
	AudioDropouts      int     `json:"audio_dropouts"`
	CodecUsed          string  `json:"codec_used"`
	QualityScore       float64 `json:"quality_score"`
	Latency            int     `json:"latency"`
}

// VideoQualitySummary 视频质量摘要
type VideoQualitySummary struct {
	AverageBitrate     float64 `json:"average_bitrate"`
	AverageFrameRate   float64 `json:"average_frame_rate"`
	AverageResolution  string  `json:"average_resolution"`
	FrameDrops         int     `json:"frame_drops"`
	KeyFrameDrops      int     `json:"key_frame_drops"`
	CodecUsed          string  `json:"codec_used"`
	QualityScore       float64 `json:"quality_score"`
	Latency            int     `json:"latency"`
}

// UserExperienceSummary 用户体验摘要
type UserExperienceSummary struct {
	OverallScore       float64 `json:"overall_score"`
	AudioClarity       float64 `json:"audio_clarity"`
	VideoClarity       float64 `json:"video_clarity"`
	ConnectionStability float64 `json:"connection_stability"`
	Responsiveness     float64 `json:"responsiveness"`
	UserSatisfaction   float64 `json:"user_satisfaction"`
	Complaints         int     `json:"complaints"`
	Compliments        int     `json:"compliments"`
}

// QualityEvent 质量事件
type QualityEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	EventType    string    `json:"event_type"`
	Severity     string    `json:"severity"`
	Description  string    `json:"description"`
	NetworkStats *NetworkStats `json:"network_stats,omitempty"`
	Duration     int64     `json:"duration,omitempty"`
}

// RealTimeStats 实时统计
type RealTimeStats struct {
	CurrentRTT          int     `json:"current_rtt"`
	CurrentPacketLoss   float64 `json:"current_packet_loss"`
	CurrentJitter       float64 `json:"current_jitter"`
	CurrentBandwidth    int     `json:"current_bandwidth"`
	CurrentAudioBitrate int     `json:"current_audio_bitrate"`
	CurrentVideoBitrate int     `json:"current_video_bitrate"`
	CurrentFrameRate    int     `json:"current_frame_rate"`
	CurrentResolution   string  `json:"current_resolution"`
	QualityLevel        string  `json:"quality_level"`
	LastUpdate          time.Time `json:"last_update"`
}

// AggregatedStats 聚合统计
type AggregatedStats struct {
	TotalCalls          int64   `json:"total_calls"`
	SuccessfulCalls     int64   `json:"successful_calls"`
	FailedCalls         int64   `json:"failed_calls"`
	AverageCallDuration float64 `json:"average_call_duration"`
	AverageQuality      float64 `json:"average_quality"`
	NetworkIssues       int64   `json:"network_issues"`
	AudioIssues         int64   `json:"audio_issues"`
	VideoIssues         int64   `json:"video_issues"`
	UserComplaints      int64   `json:"user_complaints"`
	LastUpdated         time.Time `json:"last_updated"`
}

// NewCallQualityStatsService 创建通话质量统计服务
func NewCallQualityStatsService(db *gorm.DB) *CallQualityStatsService {
	service := &CallQualityStatsService{
		db:              db,
		activeStats:     make(map[string]*CallQualityStats),
		aggregatedStats: &AggregatedStats{},
	}
	
	// 启动定期报告
	service.startPeriodicReporting()
	
	return service
}

// StartCallStats 开始通话统计
func (cqs *CallQualityStatsService) StartCallStats(callID string, userID uint) error {
	cqs.statsMutex.Lock()
	defer cqs.statsMutex.Unlock()

	stats := &CallQualityStats{
		CallID:        callID,
		UserID:        userID,
		StartTime:     time.Now(),
		Status:        "active",
		NetworkStats:  NetworkQualitySummary{},
		AudioStats:    AudioQualitySummary{},
		VideoStats:    VideoQualitySummary{},
		UserExperience: UserExperienceSummary{},
		QualityEvents: make([]QualityEvent, 0),
		RealTimeStats: RealTimeStats{},
		LastUpdate:    time.Now(),
	}

	cqs.activeStats[callID] = stats

	logrus.WithFields(logrus.Fields{
		"call_id": callID,
		"user_id": userID,
	}).Info("开始通话质量统计")

	return nil
}

// EndCallStats 结束通话统计
func (cqs *CallQualityStatsService) EndCallStats(callID string, reason string) error {
	cqs.statsMutex.Lock()
	defer cqs.statsMutex.Unlock()

	stats, exists := cqs.activeStats[callID]
	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	now := time.Now()
	stats.EndTime = &now
	stats.Duration = int64(now.Sub(stats.StartTime).Seconds())
	stats.Status = "ended"
	stats.LastUpdate = now

	// 添加结束事件
	cqs.addQualityEvent(stats, "call_ended", "info", fmt.Sprintf("通话结束: %s", reason), nil)

	// 计算最终统计
	cqs.calculateFinalStats(stats)

	// 保存到数据库
	if err := cqs.saveCallStats(stats); err != nil {
		logrus.WithError(err).Error("保存通话统计失败")
	}

	// 更新聚合统计
	cqs.updateAggregatedStats(stats)

	// 从活跃统计中移除
	delete(cqs.activeStats, callID)

	logrus.WithFields(logrus.Fields{
		"call_id": callID,
		"duration": stats.Duration,
		"quality_score": stats.UserExperience.OverallScore,
	}).Info("结束通话质量统计")

	return nil
}

// UpdateNetworkStats 更新网络统计
func (cqs *CallQualityStatsService) UpdateNetworkStats(callID string, stats *NetworkStats) error {
	cqs.statsMutex.RLock()
	callStats, exists := cqs.activeStats[callID]
	cqs.statsMutex.RUnlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	// 更新实时统计
	callStats.RealTimeStats.CurrentRTT = stats.RTT
	callStats.RealTimeStats.CurrentPacketLoss = stats.PacketLoss
	callStats.RealTimeStats.CurrentJitter = stats.Jitter
	callStats.RealTimeStats.CurrentBandwidth = stats.Bandwidth
	callStats.RealTimeStats.LastUpdate = time.Now()

	// 更新网络质量摘要
	cqs.updateNetworkSummary(callStats, stats)

	// 检查质量事件
	cqs.checkNetworkQualityEvents(callStats, stats)

	callStats.LastUpdate = time.Now()

	return nil
}

// UpdateAudioStats 更新音频统计
func (cqs *CallQualityStatsService) UpdateAudioStats(callID string, audioStats map[string]interface{}) error {
	cqs.statsMutex.RLock()
	callStats, exists := cqs.activeStats[callID]
	cqs.statsMutex.RUnlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	// 更新实时统计
	if bitrate, ok := audioStats["bitrate"].(int); ok {
		callStats.RealTimeStats.CurrentAudioBitrate = bitrate
	}
	callStats.RealTimeStats.LastUpdate = time.Now()

	// 更新音频质量摘要
	cqs.updateAudioSummary(callStats, audioStats)

	callStats.LastUpdate = time.Now()

	return nil
}

// UpdateVideoStats 更新视频统计
func (cqs *CallQualityStatsService) UpdateVideoStats(callID string, videoStats map[string]interface{}) error {
	cqs.statsMutex.RLock()
	callStats, exists := cqs.activeStats[callID]
	cqs.statsMutex.RUnlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	// 更新实时统计
	if bitrate, ok := videoStats["bitrate"].(int); ok {
		callStats.RealTimeStats.CurrentVideoBitrate = bitrate
	}
	if frameRate, ok := videoStats["frame_rate"].(int); ok {
		callStats.RealTimeStats.CurrentFrameRate = frameRate
	}
	if resolution, ok := videoStats["resolution"].(string); ok {
		callStats.RealTimeStats.CurrentResolution = resolution
	}
	callStats.RealTimeStats.LastUpdate = time.Now()

	// 更新视频质量摘要
	cqs.updateVideoSummary(callStats, videoStats)

	callStats.LastUpdate = time.Now()

	return nil
}

// AddQualityEvent 添加质量事件
func (cqs *CallQualityStatsService) AddQualityEvent(callID, eventType, severity, description string, networkStats *NetworkStats) error {
	cqs.statsMutex.RLock()
	callStats, exists := cqs.activeStats[callID]
	cqs.statsMutex.RUnlock()

	if !exists {
		return utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	cqs.addQualityEvent(callStats, eventType, severity, description, networkStats)
	callStats.LastUpdate = time.Now()

	return nil
}

// addQualityEvent 添加质量事件（内部方法）
func (cqs *CallQualityStatsService) addQualityEvent(stats *CallQualityStats, eventType, severity, description string, networkStats *NetworkStats) {
	event := QualityEvent{
		Timestamp:    time.Now(),
		EventType:    eventType,
		Severity:     severity,
		Description:  description,
		NetworkStats: networkStats,
	}

	stats.QualityEvents = append(stats.QualityEvents, event)

	// 限制事件数量
	if len(stats.QualityEvents) > 100 {
		stats.QualityEvents = stats.QualityEvents[1:]
	}
}

// updateNetworkSummary 更新网络质量摘要
func (cqs *CallQualityStatsService) updateNetworkSummary(stats *CallQualityStats, networkStats *NetworkStats) {
	// 这里应该实现滑动窗口或历史数据统计
	// 为了简化，这里只更新当前值
	
	if networkStats.RTT > stats.NetworkStats.MaxRTT {
		stats.NetworkStats.MaxRTT = networkStats.RTT
	}
	if networkStats.RTT < stats.NetworkStats.MinRTT || stats.NetworkStats.MinRTT == 0 {
		stats.NetworkStats.MinRTT = networkStats.RTT
	}
	
	if networkStats.PacketLoss > stats.NetworkStats.MaxPacketLoss {
		stats.NetworkStats.MaxPacketLoss = networkStats.PacketLoss
	}
	
	if networkStats.Jitter > stats.NetworkStats.MaxJitter {
		stats.NetworkStats.MaxJitter = networkStats.Jitter
	}
	
	if networkStats.Bandwidth < stats.NetworkStats.MinBandwidth || stats.NetworkStats.MinBandwidth == 0 {
		stats.NetworkStats.MinBandwidth = networkStats.Bandwidth
	}
	
	stats.NetworkStats.NetworkType = networkStats.NetworkType
}

// updateAudioSummary 更新音频质量摘要
func (cqs *CallQualityStatsService) updateAudioSummary(stats *CallQualityStats, audioStats map[string]interface{}) {
	if bitrate, ok := audioStats["bitrate"].(int); ok {
		// 更新平均码率
		if stats.AudioStats.AverageBitrate == 0 {
			stats.AudioStats.AverageBitrate = float64(bitrate)
		} else {
			stats.AudioStats.AverageBitrate = (stats.AudioStats.AverageBitrate + float64(bitrate)) / 2
		}
	}
	
	if codec, ok := audioStats["codec"].(string); ok {
		stats.AudioStats.CodecUsed = codec
	}
	
	if latency, ok := audioStats["latency"].(int); ok {
		stats.AudioStats.Latency = latency
	}
}

// updateVideoSummary 更新视频质量摘要
func (cqs *CallQualityStatsService) updateVideoSummary(stats *CallQualityStats, videoStats map[string]interface{}) {
	if bitrate, ok := videoStats["bitrate"].(int); ok {
		if stats.VideoStats.AverageBitrate == 0 {
			stats.VideoStats.AverageBitrate = float64(bitrate)
		} else {
			stats.VideoStats.AverageBitrate = (stats.VideoStats.AverageBitrate + float64(bitrate)) / 2
		}
	}
	
	if frameRate, ok := videoStats["frame_rate"].(int); ok {
		if stats.VideoStats.AverageFrameRate == 0 {
			stats.VideoStats.AverageFrameRate = float64(frameRate)
		} else {
			stats.VideoStats.AverageFrameRate = (stats.VideoStats.AverageFrameRate + float64(frameRate)) / 2
		}
	}
	
	if resolution, ok := videoStats["resolution"].(string); ok {
		stats.VideoStats.AverageResolution = resolution
	}
	
	if codec, ok := videoStats["codec"].(string); ok {
		stats.VideoStats.CodecUsed = codec
	}
	
	if latency, ok := videoStats["latency"].(int); ok {
		stats.VideoStats.Latency = latency
	}
}

// checkNetworkQualityEvents 检查网络质量事件
func (cqs *CallQualityStatsService) checkNetworkQualityEvents(stats *CallQualityStats, networkStats *NetworkStats) {
	// 检查高延迟
	if networkStats.RTT > 300 {
		cqs.addQualityEvent(stats, "high_latency", "warning", 
			fmt.Sprintf("网络延迟过高: %dms", networkStats.RTT), networkStats)
	}
	
	// 检查高丢包率
	if networkStats.PacketLoss > 5 {
		cqs.addQualityEvent(stats, "high_packet_loss", "warning", 
			fmt.Sprintf("网络丢包率过高: %.2f%%", networkStats.PacketLoss), networkStats)
	}
	
	// 检查高抖动
	if networkStats.Jitter > 50 {
		cqs.addQualityEvent(stats, "high_jitter", "warning", 
			fmt.Sprintf("网络抖动过高: %.2fms", networkStats.Jitter), networkStats)
	}
	
	// 检查低带宽
	if networkStats.Bandwidth < 500 {
		cqs.addQualityEvent(stats, "low_bandwidth", "warning", 
			fmt.Sprintf("网络带宽过低: %dkbps", networkStats.Bandwidth), networkStats)
	}
	
	// 检查网络不稳定
	if !networkStats.IsStable {
		cqs.addQualityEvent(stats, "unstable_network", "error", "网络连接不稳定", networkStats)
	}
}

// calculateFinalStats 计算最终统计
func (cqs *CallQualityStatsService) calculateFinalStats(stats *CallQualityStats) {
	// 计算网络质量分数
	stats.NetworkStats.QualityScore = cqs.calculateNetworkQualityScore(stats.NetworkStats)
	
	// 计算音频质量分数
	stats.AudioStats.QualityScore = cqs.calculateAudioQualityScore(stats.AudioStats)
	
	// 计算视频质量分数
	stats.VideoStats.QualityScore = cqs.calculateVideoQualityScore(stats.VideoStats)
	
	// 计算整体用户体验分数
	stats.UserExperience.OverallScore = cqs.calculateOverallExperienceScore(stats)
}

// calculateNetworkQualityScore 计算网络质量分数
func (cqs *CallQualityStatsService) calculateNetworkQualityScore(stats NetworkQualitySummary) float64 {
	score := 100.0
	
	// RTT 评分
	if stats.AverageRTT > 200 {
		score -= 30
	} else if stats.AverageRTT > 100 {
		score -= 20
	} else if stats.AverageRTT > 50 {
		score -= 10
	}
	
	// 丢包率评分
	if stats.AveragePacketLoss > 5 {
		score -= 25
	} else if stats.AveragePacketLoss > 2 {
		score -= 15
	} else if stats.AveragePacketLoss > 1 {
		score -= 5
	}
	
	// 抖动评分
	if stats.AverageJitter > 30 {
		score -= 20
	} else if stats.AverageJitter > 20 {
		score -= 10
	} else if stats.AverageJitter > 10 {
		score -= 5
	}
	
	return math.Max(0, score)
}

// calculateAudioQualityScore 计算音频质量分数
func (cqs *CallQualityStatsService) calculateAudioQualityScore(stats AudioQualitySummary) float64 {
	score := 100.0
	
	// 基于码率评分
	if stats.AverageBitrate < 64 {
		score -= 20
	} else if stats.AverageBitrate < 96 {
		score -= 10
	}
	
	// 基于延迟评分
	if stats.Latency > 200 {
		score -= 15
	} else if stats.Latency > 100 {
		score -= 10
	}
	
	// 基于音频中断评分
	if stats.AudioDropouts > 10 {
		score -= 20
	} else if stats.AudioDropouts > 5 {
		score -= 10
	}
	
	return math.Max(0, score)
}

// calculateVideoQualityScore 计算视频质量分数
func (cqs *CallQualityStatsService) calculateVideoQualityScore(stats VideoQualitySummary) float64 {
	score := 100.0
	
	// 基于码率评分
	if stats.AverageBitrate < 500 {
		score -= 25
	} else if stats.AverageBitrate < 1000 {
		score -= 15
	} else if stats.AverageBitrate < 2000 {
		score -= 5
	}
	
	// 基于帧率评分
	if stats.AverageFrameRate < 15 {
		score -= 20
	} else if stats.AverageFrameRate < 24 {
		score -= 10
	}
	
	// 基于帧丢失评分
	if stats.FrameDrops > 50 {
		score -= 20
	} else if stats.FrameDrops > 20 {
		score -= 10
	}
	
	return math.Max(0, score)
}

// calculateOverallExperienceScore 计算整体体验分数
func (cqs *CallQualityStatsService) calculateOverallExperienceScore(stats *CallQualityStats) float64 {
	// 加权计算整体分数
	networkWeight := 0.3
	audioWeight := 0.3
	videoWeight := 0.2
	eventWeight := 0.2
	
	networkScore := stats.NetworkStats.QualityScore
	audioScore := stats.AudioStats.QualityScore
	videoScore := stats.VideoStats.QualityScore
	
	// 基于质量事件调整分数
	eventPenalty := 0.0
	for _, event := range stats.QualityEvents {
		switch event.Severity {
		case "error":
			eventPenalty += 10
		case "warning":
			eventPenalty += 5
		case "info":
			eventPenalty += 1
		}
	}
	
	overallScore := (networkScore*networkWeight + audioScore*audioWeight + videoScore*videoWeight) - eventPenalty
	
	return math.Max(0, overallScore)
}

// saveCallStats 保存通话统计到数据库
func (cqs *CallQualityStatsService) saveCallStats(stats *CallQualityStats) error {
	// 这里应该实现数据库保存逻辑
	// 为了简化，这里只记录日志
	
	statsJSON, _ := json.Marshal(stats)
	logrus.WithFields(logrus.Fields{
		"call_id": stats.CallID,
		"user_id": stats.UserID,
		"duration": stats.Duration,
		"quality_score": stats.UserExperience.OverallScore,
		"stats": string(statsJSON),
	}).Info("保存通话质量统计")
	
	return nil
}

// updateAggregatedStats 更新聚合统计
func (cqs *CallQualityStatsService) updateAggregatedStats(stats *CallQualityStats) {
	cqs.aggregatedStats.TotalCalls++
	
	if stats.Status == "ended" {
		cqs.aggregatedStats.SuccessfulCalls++
		cqs.aggregatedStats.AverageCallDuration = 
			(cqs.aggregatedStats.AverageCallDuration*float64(cqs.aggregatedStats.SuccessfulCalls-1) + float64(stats.Duration)) / 
			float64(cqs.aggregatedStats.SuccessfulCalls)
	} else {
		cqs.aggregatedStats.FailedCalls++
	}
	
	// 更新平均质量
	cqs.aggregatedStats.AverageQuality = 
		(cqs.aggregatedStats.AverageQuality*float64(cqs.aggregatedStats.SuccessfulCalls-1) + stats.UserExperience.OverallScore) / 
		float64(cqs.aggregatedStats.SuccessfulCalls)
	
	// 统计质量问题
	for _, event := range stats.QualityEvents {
		switch event.EventType {
		case "high_latency", "high_packet_loss", "high_jitter", "low_bandwidth", "unstable_network":
			cqs.aggregatedStats.NetworkIssues++
		case "audio_dropout", "audio_quality_poor":
			cqs.aggregatedStats.AudioIssues++
		case "video_frame_drop", "video_quality_poor":
			cqs.aggregatedStats.VideoIssues++
		}
	}
	
	cqs.aggregatedStats.LastUpdated = time.Now()
}

// startPeriodicReporting 启动定期报告
func (cqs *CallQualityStatsService) startPeriodicReporting() {
	cqs.reportTimer = time.NewTimer(5 * time.Minute)
	
	go func() {
		for {
			select {
			case <-cqs.reportTimer.C:
				cqs.generateQualityReport()
				cqs.reportTimer.Reset(5 * time.Minute)
			}
		}
	}()
}

// generateQualityReport 生成质量报告
func (cqs *CallQualityStatsService) generateQualityReport() {
	cqs.statsMutex.RLock()
	activeCount := len(cqs.activeStats)
	cqs.statsMutex.RUnlock()
	
	logrus.WithFields(logrus.Fields{
		"active_calls": activeCount,
		"total_calls": cqs.aggregatedStats.TotalCalls,
		"success_rate": float64(cqs.aggregatedStats.SuccessfulCalls) / float64(cqs.aggregatedStats.TotalCalls) * 100,
		"average_quality": cqs.aggregatedStats.AverageQuality,
		"network_issues": cqs.aggregatedStats.NetworkIssues,
		"audio_issues": cqs.aggregatedStats.AudioIssues,
		"video_issues": cqs.aggregatedStats.VideoIssues,
	}).Info("通话质量报告")
}

// GetCallStats 获取通话统计
func (cqs *CallQualityStatsService) GetCallStats(callID string) (*CallQualityStats, error) {
	cqs.statsMutex.RLock()
	defer cqs.statsMutex.RUnlock()

	stats, exists := cqs.activeStats[callID]
	if !exists {
		return nil, utils.NewAppError(utils.ErrCodeMessageNotFound, "通话统计不存在")
	}

	return stats, nil
}

// GetAggregatedStats 获取聚合统计
func (cqs *CallQualityStatsService) GetAggregatedStats() *AggregatedStats {
	return cqs.aggregatedStats
}

// GetActiveCallsCount 获取活跃通话数量
func (cqs *CallQualityStatsService) GetActiveCallsCount() int {
	cqs.statsMutex.RLock()
	defer cqs.statsMutex.RUnlock()

	return len(cqs.activeStats)
}
