package service

import (
	"sync"
	"time"
)

// NetworkQualityMonitor 网络质量监控器
type NetworkQualityMonitor struct {
	qualityHistory map[uint][]QualitySnapshot
	mutex          sync.RWMutex
}

// QualitySnapshot 质量快照
type QualitySnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	RTT            int       `json:"rtt"`
	PacketLoss     float64   `json:"packet_loss"`
	Jitter         float64   `json:"jitter"`
	Bandwidth      int       `json:"bandwidth"`
	NetworkType    string    `json:"network_type"`
	SignalStrength int       `json:"signal_strength"`
	QualityScore   float64   `json:"quality_score"`
}

// NetworkQualityLevel 网络质量等级
type NetworkQualityLevel int

const (
	QualityExcellent NetworkQualityLevel = iota
	QualityGood
	QualityFair
	QualityPoor
	QualityVeryPoor
)

// QualityThresholds 质量阈值
type QualityThresholds struct {
	RTTExcellent int // < 50ms
	RTTGood      int // < 100ms
	RTTPoor      int // < 200ms

	PacketLossExcellent float64 // < 1%
	PacketLossGood      float64 // < 3%
	PacketLossPoor      float64 // < 10%

	JitterExcellent float64 // < 10ms
	JitterGood      float64 // < 20ms
	JitterPoor      float64 // < 50ms

	BandwidthExcellent int // > 2000 kbps
	BandwidthGood      int // > 1000 kbps
	BandwidthPoor      int // > 500 kbps
}

// DefaultQualityThresholds 默认质量阈值
var DefaultQualityThresholds = QualityThresholds{
	RTTExcellent: 50,
	RTTGood:      100,
	RTTPoor:      200,

	PacketLossExcellent: 1.0,
	PacketLossGood:      3.0,
	PacketLossPoor:      10.0,

	JitterExcellent: 10.0,
	JitterGood:      20.0,
	JitterPoor:      50.0,

	BandwidthExcellent: 2000,
	BandwidthGood:      1000,
	BandwidthPoor:      500,
}

// NewNetworkQualityMonitor 创建网络质量监控器
func NewNetworkQualityMonitor() *NetworkQualityMonitor {
	return &NetworkQualityMonitor{
		qualityHistory: make(map[uint][]QualitySnapshot),
	}
}

// RecordNetworkStats 记录网络统计
func (nqm *NetworkQualityMonitor) RecordNetworkStats(userID uint, stats *NetworkStats) {
	nqm.mutex.Lock()
	defer nqm.mutex.Unlock()

	// 计算质量分数
	qualityScore := nqm.calculateQualityScore(stats)

	// 创建质量快照
	snapshot := QualitySnapshot{
		Timestamp:      time.Now(),
		RTT:            stats.RTT,
		PacketLoss:     stats.PacketLoss,
		Jitter:         stats.Jitter,
		Bandwidth:      stats.Bandwidth,
		NetworkType:    stats.NetworkType,
		SignalStrength: stats.SignalStrength,
		QualityScore:   qualityScore,
	}

	// 添加到历史记录
	if nqm.qualityHistory[userID] == nil {
		nqm.qualityHistory[userID] = make([]QualitySnapshot, 0)
	}
	nqm.qualityHistory[userID] = append(nqm.qualityHistory[userID], snapshot)

	// 限制历史记录数量（保留最近100条）
	if len(nqm.qualityHistory[userID]) > 100 {
		nqm.qualityHistory[userID] = nqm.qualityHistory[userID][1:]
	}
}

// CalculateQualityScore 计算质量分数
func (nqm *NetworkQualityMonitor) CalculateQualityScore(stats *NetworkStats) float64 {
	return nqm.calculateQualityScoreWithThresholds(stats, &DefaultQualityThresholds)
}

// calculateQualityScoreWithThresholds 使用指定阈值计算质量分数
func (nqm *NetworkQualityMonitor) calculateQualityScoreWithThresholds(stats *NetworkStats, thresholds *QualityThresholds) float64 {
	// RTT 分数 (权重: 30%)
	rttScore := nqm.calculateRTTScore(stats.RTT, thresholds)

	// 丢包率分数 (权重: 25%)
	packetLossScore := nqm.calculatePacketLossScore(stats.PacketLoss, thresholds)

	// 抖动分数 (权重: 20%)
	jitterScore := nqm.calculateJitterScore(stats.Jitter, thresholds)

	// 带宽分数 (权重: 15%)
	bandwidthScore := nqm.calculateBandwidthScore(stats.Bandwidth, thresholds)

	// 信号强度分数 (权重: 10%)
	signalScore := nqm.calculateSignalScore(stats.SignalStrength)

	// 加权计算总分
	totalScore := rttScore*0.3 + packetLossScore*0.25 + jitterScore*0.2 + bandwidthScore*0.15 + signalScore*0.1

	return totalScore
}

// calculateRTTScore 计算 RTT 分数
func (nqm *NetworkQualityMonitor) calculateRTTScore(rtt int, thresholds *QualityThresholds) float64 {
	if rtt <= thresholds.RTTExcellent {
		return 100.0
	} else if rtt <= thresholds.RTTGood {
		return 80.0 + (20.0*float64(thresholds.RTTGood-rtt))/float64(thresholds.RTTGood-thresholds.RTTExcellent)
	} else if rtt <= thresholds.RTTPoor {
		return 60.0 + (20.0*float64(thresholds.RTTPoor-rtt))/float64(thresholds.RTTPoor-thresholds.RTTGood)
	} else {
		// 超过200ms，分数急剧下降
		return maxQuality(0, 60.0-float64(rtt-thresholds.RTTPoor)*0.5)
	}
}

// calculatePacketLossScore 计算丢包率分数
func (nqm *NetworkQualityMonitor) calculatePacketLossScore(packetLoss float64, thresholds *QualityThresholds) float64 {
	if packetLoss <= thresholds.PacketLossExcellent {
		return 100.0
	} else if packetLoss <= thresholds.PacketLossGood {
		return 80.0 + (20.0*(thresholds.PacketLossGood-packetLoss))/(thresholds.PacketLossGood-thresholds.PacketLossExcellent)
	} else if packetLoss <= thresholds.PacketLossPoor {
		return 60.0 + (20.0*(thresholds.PacketLossPoor-packetLoss))/(thresholds.PacketLossPoor-thresholds.PacketLossGood)
	} else {
		// 超过10%丢包，分数急剧下降
		return maxQuality(0, 60.0-(packetLoss-thresholds.PacketLossPoor)*5)
	}
}

// calculateJitterScore 计算抖动分数
func (nqm *NetworkQualityMonitor) calculateJitterScore(jitter float64, thresholds *QualityThresholds) float64 {
	if jitter <= thresholds.JitterExcellent {
		return 100.0
	} else if jitter <= thresholds.JitterGood {
		return 80.0 + (20.0*(thresholds.JitterGood-jitter))/(thresholds.JitterGood-thresholds.JitterExcellent)
	} else if jitter <= thresholds.JitterPoor {
		return 60.0 + (20.0*(thresholds.JitterPoor-jitter))/(thresholds.JitterPoor-thresholds.JitterGood)
	} else {
		// 超过50ms抖动，分数急剧下降
		return maxQuality(0, 60.0-(jitter-thresholds.JitterPoor)*2)
	}
}

// calculateBandwidthScore 计算带宽分数
func (nqm *NetworkQualityMonitor) calculateBandwidthScore(bandwidth int, thresholds *QualityThresholds) float64 {
	if bandwidth >= thresholds.BandwidthExcellent {
		return 100.0
	} else if bandwidth >= thresholds.BandwidthGood {
		return 80.0 + (20.0*float64(bandwidth-thresholds.BandwidthGood))/float64(thresholds.BandwidthExcellent-thresholds.BandwidthGood)
	} else if bandwidth >= thresholds.BandwidthPoor {
		return 60.0 + (20.0*float64(bandwidth-thresholds.BandwidthPoor))/float64(thresholds.BandwidthGood-thresholds.BandwidthPoor)
	} else {
		// 低于500kbps，分数急剧下降
		return maxQuality(0, 60.0-float64(thresholds.BandwidthPoor-bandwidth)*0.1)
	}
}

// calculateSignalScore 计算信号强度分数
func (nqm *NetworkQualityMonitor) calculateSignalScore(signalStrength int) float64 {
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

// GetQualityLevel 获取网络质量等级
func (nqm *NetworkQualityMonitor) GetQualityLevel(score float64) NetworkQualityLevel {
	switch {
	case score >= 90:
		return QualityExcellent
	case score >= 75:
		return QualityGood
	case score >= 60:
		return QualityFair
	case score >= 40:
		return QualityPoor
	default:
		return QualityVeryPoor
	}
}

// GetQualityLevelName 获取质量等级名称
func (nqm *NetworkQualityMonitor) GetQualityLevelName(level NetworkQualityLevel) string {
	switch level {
	case QualityExcellent:
		return "优秀"
	case QualityGood:
		return "良好"
	case QualityFair:
		return "一般"
	case QualityPoor:
		return "较差"
	case QualityVeryPoor:
		return "很差"
	default:
		return "未知"
	}
}

// GetRecentQualityTrend 获取最近的质量趋势
func (nqm *NetworkQualityMonitor) GetRecentQualityTrend(userID uint, minutes int) ([]QualitySnapshot, error) {
	nqm.mutex.RLock()
	defer nqm.mutex.RUnlock()

	history, exists := nqm.qualityHistory[userID]
	if !exists {
		return []QualitySnapshot{}, nil
	}

	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var recentSnapshots []QualitySnapshot

	for _, snapshot := range history {
		if snapshot.Timestamp.After(cutoffTime) {
			recentSnapshots = append(recentSnapshots, snapshot)
		}
	}

	return recentSnapshots, nil
}

// GetAverageQuality 获取平均质量
func (nqm *NetworkQualityMonitor) GetAverageQuality(userID uint, minutes int) (float64, error) {
	snapshots, err := nqm.GetRecentQualityTrend(userID, minutes)
	if err != nil {
		return 0, err
	}

	if len(snapshots) == 0 {
		return 0, nil
	}

	var totalScore float64
	for _, snapshot := range snapshots {
		totalScore += snapshot.QualityScore
	}

	return totalScore / float64(len(snapshots)), nil
}

// IsNetworkStable 判断网络是否稳定
func (nqm *NetworkQualityMonitor) IsNetworkStable(userID uint, minutes int, threshold float64) (bool, error) {
	snapshots, err := nqm.GetRecentQualityTrend(userID, minutes)
	if err != nil {
		return false, err
	}

	if len(snapshots) < 3 {
		return true, nil // 数据不足，认为稳定
	}

	// 计算质量分数的标准差
	var totalScore float64
	for _, snapshot := range snapshots {
		totalScore += snapshot.QualityScore
	}
	avgScore := totalScore / float64(len(snapshots))

	var variance float64
	for _, snapshot := range snapshots {
		diff := snapshot.QualityScore - avgScore
		variance += diff * diff
	}
	stdDev := variance / float64(len(snapshots))

	// 标准差小于阈值认为稳定
	return stdDev < threshold, nil
}

// PredictNetworkQuality 预测网络质量
func (nqm *NetworkQualityMonitor) PredictNetworkQuality(userID uint) (float64, error) {
	// 获取最近5分钟的质量数据
	snapshots, err := nqm.GetRecentQualityTrend(userID, 5)
	if err != nil {
		return 0, err
	}

	if len(snapshots) < 3 {
		return 0, nil
	}

	// 简单的线性回归预测
	// 这里可以使用更复杂的机器学习算法
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(snapshots))

	for i, snapshot := range snapshots {
		x := float64(i)
		y := snapshot.QualityScore
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// 计算斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)

	// 预测下一个时间点的质量
	nextX := n
	predictedY := (sumY / n) + slope*(nextX-sumX/n)

	// 限制预测值在合理范围内
	if predictedY > 100 {
		predictedY = 100
	} else if predictedY < 0 {
		predictedY = 0
	}

	return predictedY, nil
}

// GetNetworkRecommendations 获取网络优化建议
func (nqm *NetworkQualityMonitor) GetNetworkRecommendations(userID uint) []string {
	nqm.mutex.RLock()
	defer nqm.mutex.RUnlock()

	history, exists := nqm.qualityHistory[userID]
	if !exists || len(history) == 0 {
		return []string{"暂无网络数据"}
	}

	latest := history[len(history)-1]
	var recommendations []string

	// 基于最新数据提供建议
	if latest.RTT > 200 {
		recommendations = append(recommendations, "延迟较高，建议切换到更稳定的网络")
	}

	if latest.PacketLoss > 5 {
		recommendations = append(recommendations, "丢包率较高，建议检查网络连接")
	}

	if latest.Jitter > 50 {
		recommendations = append(recommendations, "网络抖动严重，建议使用有线网络")
	}

	if latest.Bandwidth < 1000 {
		recommendations = append(recommendations, "带宽不足，建议关闭其他占用网络的应用程序")
	}

	if latest.SignalStrength < 30 {
		recommendations = append(recommendations, "信号强度较弱，建议靠近路由器或使用WiFi")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "网络质量良好")
	}

	return recommendations
}

// CleanupOldData 清理旧数据
func (nqm *NetworkQualityMonitor) CleanupOldData(maxAge time.Duration) {
	nqm.mutex.Lock()
	defer nqm.mutex.Unlock()

	cutoffTime := time.Now().Add(-maxAge)

	for userID, history := range nqm.qualityHistory {
		var filteredHistory []QualitySnapshot
		for _, snapshot := range history {
			if snapshot.Timestamp.After(cutoffTime) {
				filteredHistory = append(filteredHistory, snapshot)
			}
		}
		nqm.qualityHistory[userID] = filteredHistory

		// 如果用户没有历史数据，删除用户记录
		if len(filteredHistory) == 0 {
			delete(nqm.qualityHistory, userID)
		}
	}
}

// GetStatistics 获取统计信息
func (nqm *NetworkQualityMonitor) GetStatistics() map[string]interface{} {
	nqm.mutex.RLock()
	defer nqm.mutex.RUnlock()

	totalUsers := len(nqm.qualityHistory)
	var totalSnapshots int
	var totalScore float64

	for _, history := range nqm.qualityHistory {
		totalSnapshots += len(history)
		for _, snapshot := range history {
			totalScore += snapshot.QualityScore
		}
	}

	avgQuality := float64(0)
	if totalSnapshots > 0 {
		avgQuality = totalScore / float64(totalSnapshots)
	}

	return map[string]interface{}{
		"total_users":     totalUsers,
		"total_snapshots": totalSnapshots,
		"average_quality": avgQuality,
		"quality_level":   nqm.GetQualityLevelName(nqm.GetQualityLevel(avgQuality)),
	}
}

// maxQuality 返回两个浮点数中的较大值
func maxQuality(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
