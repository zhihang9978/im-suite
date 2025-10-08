package service

import (
	"sync"
	"time"
)

// BandwidthAdaptor 带宽自适应器
type BandwidthAdaptor struct {
	adaptationHistory map[uint][]AdaptationRecord
	mutex             sync.RWMutex
}

// AdaptationRecord 自适应记录
type AdaptationRecord struct {
	Timestamp    time.Time           `json:"timestamp"`
	NetworkStats *NetworkStats       `json:"network_stats"`
	Adaptation   *AdaptationSettings `json:"adaptation"`
	QualityScore float64             `json:"quality_score"`
	Success      bool                `json:"success"`
}

// AdaptationSettings 自适应设置
type AdaptationSettings struct {
	VideoBitrate int    `json:"video_bitrate"` // kbps
	AudioBitrate int    `json:"audio_bitrate"` // kbps
	Resolution   string `json:"resolution"`    // 如: 1920x1080, 1280x720, 640x480
	FrameRate    int    `json:"frame_rate"`    // fps
	Codec        string `json:"codec"`         // 编解码器
	AudioCodec   string `json:"audio_codec"`   // 音频编解码器
	VideoCodec   string `json:"video_codec"`   // 视频编解码器
	AdaptiveMode string `json:"adaptive_mode"` // auto, manual, conservative, aggressive
	QualityLevel string `json:"quality_level"` // high, medium, low, very_low
}

// QualityPresets 质量预设
var QualityPresets = map[string]*AdaptationSettings{
	"ultra_hd": {
		VideoBitrate: 8000,
		AudioBitrate: 256,
		Resolution:   "3840x2160",
		FrameRate:    30,
		VideoCodec:   "VP9",
		AudioCodec:   "OPUS",
		QualityLevel: "ultra_high",
	},
	"full_hd": {
		VideoBitrate: 4000,
		AudioBitrate: 192,
		Resolution:   "1920x1080",
		FrameRate:    30,
		VideoCodec:   "VP8",
		AudioCodec:   "OPUS",
		QualityLevel: "high",
	},
	"hd": {
		VideoBitrate: 2000,
		AudioBitrate: 128,
		Resolution:   "1280x720",
		FrameRate:    30,
		VideoCodec:   "VP8",
		AudioCodec:   "OPUS",
		QualityLevel: "medium",
	},
	"sd": {
		VideoBitrate: 1000,
		AudioBitrate: 96,
		Resolution:   "854x480",
		FrameRate:    24,
		VideoCodec:   "VP8",
		AudioCodec:   "OPUS",
		QualityLevel: "low",
	},
	"low": {
		VideoBitrate: 500,
		AudioBitrate: 64,
		Resolution:   "640x360",
		FrameRate:    15,
		VideoCodec:   "VP8",
		AudioCodec:   "OPUS",
		QualityLevel: "very_low",
	},
	"audio_only": {
		VideoBitrate: 0,
		AudioBitrate: 128,
		Resolution:   "0x0",
		FrameRate:    0,
		VideoCodec:   "",
		AudioCodec:   "OPUS",
		QualityLevel: "audio_only",
	},
}

// NewBandwidthAdaptor 创建带宽自适应器
func NewBandwidthAdaptor() *BandwidthAdaptor {
	return &BandwidthAdaptor{
		adaptationHistory: make(map[uint][]AdaptationRecord),
	}
}

// Adapt 根据网络质量自适应调整
func (ba *BandwidthAdaptor) Adapt(stats *NetworkStats) *AdaptationSettings {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	// 计算网络质量分数
	qualityScore := ba.calculateNetworkQualityScore(stats)

	// 根据质量分数选择预设
	var selectedPreset string
	switch {
	case qualityScore >= 90:
		selectedPreset = "full_hd"
	case qualityScore >= 75:
		selectedPreset = "hd"
	case qualityScore >= 60:
		selectedPreset = "sd"
	case qualityScore >= 40:
		selectedPreset = "low"
	case qualityScore >= 20:
		selectedPreset = "audio_only"
	default:
		selectedPreset = "audio_only"
	}

	// 获取基础设置
	baseSettings := QualityPresets[selectedPreset]

	// 根据网络条件进一步调整
	adaptedSettings := ba.fineTuneSettings(baseSettings, stats)

	// 记录自适应历史
	ba.recordAdaptation(stats.UserID, stats, adaptedSettings, qualityScore, true)

	return adaptedSettings
}

// calculateNetworkQualityScore 计算网络质量分数
func (ba *BandwidthAdaptor) calculateNetworkQualityScore(stats *NetworkStats) float64 {
	// 延迟分数 (权重: 30%)
	rttScore := ba.calculateRTTScore(stats.RTT)

	// 丢包率分数 (权重: 25%)
	packetLossScore := ba.calculatePacketLossScore(stats.PacketLoss)

	// 带宽分数 (权重: 25%)
	bandwidthScore := ba.calculateBandwidthScore(stats.Bandwidth)

	// 信号强度分数 (权重: 20%)
	signalScore := ba.calculateSignalScore(stats.SignalStrength)

	// 加权计算总分
	totalScore := rttScore*0.3 + packetLossScore*0.25 + bandwidthScore*0.25 + signalScore*0.2

	return totalScore
}

// calculateRTTScore 计算RTT分数
func (ba *BandwidthAdaptor) calculateRTTScore(rtt int) float64 {
	switch {
	case rtt <= 50:
		return 100.0
	case rtt <= 100:
		return 80.0 + (20.0*float64(100-rtt))/50.0
	case rtt <= 200:
		return 60.0 + (20.0*float64(200-rtt))/100.0
	case rtt <= 500:
		return 40.0 + (20.0*float64(500-rtt))/300.0
	default:
		return maxFloat(0, 40.0-float64(rtt-500)*0.1)
	}
}

// calculatePacketLossScore 计算丢包率分数
func (ba *BandwidthAdaptor) calculatePacketLossScore(packetLoss float64) float64 {
	switch {
	case packetLoss <= 1.0:
		return 100.0
	case packetLoss <= 3.0:
		return 80.0 + (20.0*(3.0-packetLoss))/2.0
	case packetLoss <= 10.0:
		return 60.0 + (20.0*(10.0-packetLoss))/7.0
	default:
		return maxFloat(0, 60.0-(packetLoss-10.0)*5.0)
	}
}

// calculateBandwidthScore 计算带宽分数
func (ba *BandwidthAdaptor) calculateBandwidthScore(bandwidth int) float64 {
	switch {
	case bandwidth >= 5000:
		return 100.0
	case bandwidth >= 2000:
		return 80.0 + (20.0*float64(bandwidth-2000))/3000.0
	case bandwidth >= 1000:
		return 60.0 + (20.0*float64(bandwidth-1000))/1000.0
	case bandwidth >= 500:
		return 40.0 + (20.0*float64(bandwidth-500))/500.0
	case bandwidth >= 200:
		return 20.0 + (20.0*float64(bandwidth-200))/300.0
	default:
		return maxFloat(0, float64(bandwidth)*0.1)
	}
}

// calculateSignalScore 计算信号强度分数
func (ba *BandwidthAdaptor) calculateSignalScore(signalStrength int) float64 {
	if signalStrength >= 80 {
		return 100.0
	} else if signalStrength >= 60 {
		return 80.0 + (20.0*float64(signalStrength-60))/20.0
	} else if signalStrength >= 40 {
		return 60.0 + (20.0*float64(signalStrength-40))/20.0
	} else if signalStrength >= 20 {
		return 40.0 + (20.0*float64(signalStrength-20))/20.0
	} else {
		return float64(signalStrength) * 2.0
	}
}

// fineTuneSettings 根据网络条件微调设置
func (ba *BandwidthAdaptor) fineTuneSettings(base *AdaptationSettings, stats *NetworkStats) *AdaptationSettings {
	// 复制基础设置
	adapted := &AdaptationSettings{
		VideoBitrate: base.VideoBitrate,
		AudioBitrate: base.AudioBitrate,
		Resolution:   base.Resolution,
		FrameRate:    base.FrameRate,
		Codec:        base.Codec,
		AudioCodec:   base.AudioCodec,
		VideoCodec:   base.VideoCodec,
		AdaptiveMode: "auto",
		QualityLevel: base.QualityLevel,
	}

	// 根据带宽调整视频码率
	if stats.Bandwidth > 0 {
		// 预留20%带宽给音频和其他开销
		availableVideoBitrate := int(float64(stats.Bandwidth) * 0.8)
		if adapted.VideoBitrate > availableVideoBitrate {
			adapted.VideoBitrate = availableVideoBitrate
		}
	}

	// 根据延迟调整帧率
	if stats.RTT > 200 {
		// 高延迟时降低帧率
		if adapted.FrameRate > 15 {
			adapted.FrameRate = 15
		}
	}

	// 根据丢包率调整编解码器
	if stats.PacketLoss > 5.0 {
		// 高丢包率时使用更鲁棒的编解码器
		adapted.VideoCodec = "VP8" // VP8在高丢包率下表现更好
	}

	// 根据网络类型调整
	switch stats.NetworkType {
	case "2g":
		adapted.VideoBitrate = min(adapted.VideoBitrate, 300)
		adapted.FrameRate = min(adapted.FrameRate, 10)
		adapted.Resolution = "320x240"
	case "3g":
		adapted.VideoBitrate = min(adapted.VideoBitrate, 800)
		adapted.FrameRate = min(adapted.FrameRate, 20)
		adapted.Resolution = "640x480"
	case "4g":
		// 4G网络可以支持更高的质量
		if stats.Bandwidth > 2000 {
			adapted.VideoBitrate = min(adapted.VideoBitrate, 3000)
		}
	case "wifi":
		// WiFi网络通常比较稳定
		if stats.SignalStrength > 70 {
			adapted.VideoBitrate = min(adapted.VideoBitrate, 4000)
		}
	}

	// 确保设置合理
	ba.validateSettings(adapted)

	return adapted
}

// validateSettings 验证和修正设置
func (ba *BandwidthAdaptor) validateSettings(settings *AdaptationSettings) {
	// 确保码率不为负
	if settings.VideoBitrate < 0 {
		settings.VideoBitrate = 0
	}
	if settings.AudioBitrate < 0 {
		settings.AudioBitrate = 64
	}

	// 确保帧率合理
	if settings.FrameRate < 0 {
		settings.FrameRate = 0
	} else if settings.FrameRate > 60 {
		settings.FrameRate = 60
	}

	// 确保分辨率格式正确
	if settings.Resolution == "" {
		settings.Resolution = "640x480"
	}
}

// recordAdaptation 记录自适应历史
func (ba *BandwidthAdaptor) recordAdaptation(userID uint, stats *NetworkStats, adaptation *AdaptationSettings, qualityScore float64, success bool) {
	record := AdaptationRecord{
		Timestamp:    time.Now(),
		NetworkStats: stats,
		Adaptation:   adaptation,
		QualityScore: qualityScore,
		Success:      success,
	}

	if ba.adaptationHistory[userID] == nil {
		ba.adaptationHistory[userID] = make([]AdaptationRecord, 0)
	}
	ba.adaptationHistory[userID] = append(ba.adaptationHistory[userID], record)

	// 限制历史记录数量
	if len(ba.adaptationHistory[userID]) > 50 {
		ba.adaptationHistory[userID] = ba.adaptationHistory[userID][1:]
	}
}

// GetAdaptationHistory 获取自适应历史
func (ba *BandwidthAdaptor) GetAdaptationHistory(userID uint, minutes int) ([]AdaptationRecord, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	history, exists := ba.adaptationHistory[userID]
	if !exists {
		return []AdaptationRecord{}, nil
	}

	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var recentRecords []AdaptationRecord

	for _, record := range history {
		if record.Timestamp.After(cutoffTime) {
			recentRecords = append(recentRecords, record)
		}
	}

	return recentRecords, nil
}

// GetOptimalSettings 获取最优设置
func (ba *BandwidthAdaptor) GetOptimalSettings(userID uint) (*AdaptationSettings, error) {
	history, err := ba.GetAdaptationHistory(userID, 10) // 最近10分钟
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return QualityPresets["hd"], nil // 默认HD质量
	}

	// 找到成功率最高的设置
	var bestRecord AdaptationRecord
	var maxScore float64

	for _, record := range history {
		if record.Success && record.QualityScore > maxScore {
			maxScore = record.QualityScore
			bestRecord = record
		}
	}

	if bestRecord.Adaptation != nil {
		return bestRecord.Adaptation, nil
	}

	return QualityPresets["hd"], nil
}

// GetAdaptationStats 获取自适应统计
func (ba *BandwidthAdaptor) GetAdaptationStats(userID uint) map[string]interface{} {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	history, exists := ba.adaptationHistory[userID]
	if !exists {
		return map[string]interface{}{
			"total_adaptations": 0,
			"success_rate":      0.0,
			"average_quality":   0.0,
		}
	}

	totalAdaptations := len(history)
	var successCount int
	var totalQualityScore float64

	for _, record := range history {
		if record.Success {
			successCount++
		}
		totalQualityScore += record.QualityScore
	}

	successRate := float64(0)
	if totalAdaptations > 0 {
		successRate = float64(successCount) / float64(totalAdaptations) * 100
	}

	averageQuality := float64(0)
	if totalAdaptations > 0 {
		averageQuality = totalQualityScore / float64(totalAdaptations)
	}

	return map[string]interface{}{
		"total_adaptations": totalAdaptations,
		"success_rate":      successRate,
		"average_quality":   averageQuality,
		"last_adaptation":   history[len(history)-1].Timestamp,
	}
}

// CleanupOldData 清理旧数据
func (ba *BandwidthAdaptor) CleanupOldData(maxAge time.Duration) {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	cutoffTime := time.Now().Add(-maxAge)

	for userID, history := range ba.adaptationHistory {
		var filteredHistory []AdaptationRecord
		for _, record := range history {
			if record.Timestamp.After(cutoffTime) {
				filteredHistory = append(filteredHistory, record)
			}
		}
		ba.adaptationHistory[userID] = filteredHistory

		// 如果用户没有历史数据，删除用户记录
		if len(filteredHistory) == 0 {
			delete(ba.adaptationHistory, userID)
		}
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxFloat 返回两个浮点数中的较大值
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
