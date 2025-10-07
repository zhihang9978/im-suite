package service

import (
	"sort"
	"sync"
)

// CodecManager 编解码器管理器
type CodecManager struct {
	supportedCodecs map[string]*CodecInfo
	preferences     map[string][]string // 按优先级排序的编解码器列表
	mutex           sync.RWMutex
}

// CodecInfo 编解码器信息
type CodecInfo struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`            // audio, video
	MimeType        string  `json:"mime_type"`
	ClockRate       int     `json:"clock_rate"`
	Channels        int     `json:"channels"`
	Bitrate         int     `json:"bitrate"`
	Complexity      int     `json:"complexity"`
	Latency         int     `json:"latency"`         // ms
	CPUUsage        float64 `json:"cpu_usage"`       // 0-100
	Bandwidth       int     `json:"bandwidth"`       // kbps
	PacketLoss      float64 `json:"packet_loss"`     // %
	NetworkAdaptive bool    `json:"network_adaptive"`
	HardwareAccel   bool    `json:"hardware_accel"`
	BrowserSupport  map[string]bool `json:"browser_support"`
	MobileSupport   bool    `json:"mobile_support"`
	Description     string  `json:"description"`
}

// CodecPreferences 编解码器偏好设置
type CodecPreferences struct {
	NetworkType      string `json:"network_type"`      // wifi, 4g, 3g, 2g
	Bandwidth        int    `json:"bandwidth"`         // kbps
	Latency          int    `json:"latency"`           // ms
	PacketLoss       float64 `json:"packet_loss"`      // %
	DeviceType       string `json:"device_type"`       // desktop, mobile, tablet
	OS               string `json:"os"`                // windows, macos, linux, android, ios
	Browser          string `json:"browser"`           // chrome, firefox, safari, edge
	HardwareAccel    bool   `json:"hardware_accel"`
	BatteryOptimized bool   `json:"battery_optimized"`
}

// NewCodecManager 创建编解码器管理器
func NewCodecManager() *CodecManager {
	cm := &CodecManager{
		supportedCodecs: make(map[string]*CodecInfo),
		preferences:     make(map[string][]string),
	}
	
	// 初始化支持的编解码器
	cm.initializeCodecs()
	
	// 初始化偏好设置
	cm.initializePreferences()
	
	return cm
}

// initializeCodecs 初始化支持的编解码器
func (cm *CodecManager) initializeCodecs() {
	// 音频编解码器
	cm.registerAudioCodec("opus", &CodecInfo{
		Name:            "OPUS",
		Type:            "audio",
		MimeType:        "audio/opus",
		ClockRate:       48000,
		Channels:        2,
		Bitrate:         128,
		Complexity:      5,
		Latency:         20,
		CPUUsage:        15.0,
		Bandwidth:       128,
		PacketLoss:      2.0,
		NetworkAdaptive: true,
		HardwareAccel:   false,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  false,
			"edge":    true,
		},
		MobileSupport: true,
		Description:   "高质量音频编解码器，低延迟，网络自适应",
	})

	cm.registerAudioCodec("g722", &CodecInfo{
		Name:            "G.722",
		Type:            "audio",
		MimeType:        "audio/G722",
		ClockRate:       8000,
		Channels:        1,
		Bitrate:         64,
		Complexity:      3,
		Latency:         30,
		CPUUsage:        10.0,
		Bandwidth:       64,
		PacketLoss:      3.0,
		NetworkAdaptive: false,
		HardwareAccel:   true,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  true,
			"edge":    true,
		},
		MobileSupport: true,
		Description:   "经典音频编解码器，兼容性好",
	})

	cm.registerAudioCodec("pcmu", &CodecInfo{
		Name:            "PCMU",
		Type:            "audio",
		MimeType:        "audio/PCMU",
		ClockRate:       8000,
		Channels:        1,
		Bitrate:         64,
		Complexity:      1,
		Latency:         20,
		CPUUsage:        5.0,
		Bandwidth:       64,
		PacketLoss:      5.0,
		NetworkAdaptive: false,
		HardwareAccel:   true,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  true,
			"edge":    true,
		},
		MobileSupport: true,
		Description:   "简单音频编解码器，低CPU使用率",
	})

	// 视频编解码器
	cm.registerVideoCodec("vp8", &CodecInfo{
		Name:            "VP8",
		Type:            "video",
		MimeType:        "video/VP8",
		ClockRate:       90000,
		Bitrate:         2000,
		Complexity:      4,
		Latency:         50,
		CPUUsage:        25.0,
		Bandwidth:       2000,
		PacketLoss:      3.0,
		NetworkAdaptive: true,
		HardwareAccel:   true,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  false,
			"edge":    true,
		},
		MobileSupport: true,
		Description:   "开源视频编解码器，网络自适应，移动设备友好",
	})

	cm.registerVideoCodec("vp9", &CodecInfo{
		Name:            "VP9",
		Type:            "video",
		MimeType:        "video/VP9",
		ClockRate:       90000,
		Bitrate:         3000,
		Complexity:      6,
		Latency:         60,
		CPUUsage:        35.0,
		Bandwidth:       3000,
		PacketLoss:      2.5,
		NetworkAdaptive: true,
		HardwareAccel:   true,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  false,
			"edge":    false,
		},
		MobileSupport: false,
		Description:   "新一代开源视频编解码器，高压缩率，需要更多CPU",
	})

	cm.registerVideoCodec("h264", &CodecInfo{
		Name:            "H.264",
		Type:            "video",
		MimeType:        "video/H264",
		ClockRate:       90000,
		Bitrate:         2500,
		Complexity:      5,
		Latency:         40,
		CPUUsage:        20.0,
		Bandwidth:       2500,
		PacketLoss:      2.0,
		NetworkAdaptive: false,
		HardwareAccel:   true,
		BrowserSupport: map[string]bool{
			"chrome":  true,
			"firefox": true,
			"safari":  true,
			"edge":    true,
		},
		MobileSupport: true,
		Description:   "广泛支持的视频编解码器，硬件加速支持好",
	})
}

// registerAudioCodec 注册音频编解码器
func (cm *CodecManager) registerAudioCodec(name string, info *CodecInfo) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.supportedCodecs[name] = info
}

// registerVideoCodec 注册视频编解码器
func (cm *CodecManager) registerVideoCodec(name string, info *CodecInfo) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.supportedCodecs[name] = info
}

// initializePreferences 初始化偏好设置
func (cm *CodecManager) initializePreferences() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 音频编解码器偏好（按优先级排序）
	cm.preferences["audio"] = []string{"opus", "g722", "pcmu"}
	
	// 视频编解码器偏好（按优先级排序）
	cm.preferences["video"] = []string{"vp8", "h264", "vp9"}
	
	// 网络类型偏好
	cm.preferences["wifi"] = []string{"vp8", "h264", "opus"}
	cm.preferences["4g"] = []string{"vp8", "h264", "opus"}
	cm.preferences["3g"] = []string{"vp8", "opus", "g722"}
	cm.preferences["2g"] = []string{"opus", "g722", "pcmu"}
	
	// 设备类型偏好
	cm.preferences["desktop"] = []string{"vp8", "h264", "opus"}
	cm.preferences["mobile"] = []string{"vp8", "opus", "g722"}
	cm.preferences["tablet"] = []string{"vp8", "h264", "opus"}
}

// GetOptimalCodecs 获取最优编解码器
func (cm *CodecManager) GetOptimalCodecs(prefs *CodecPreferences) ([]string, []string, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 根据偏好设置选择编解码器
	audioCodecs := cm.selectAudioCodecs(prefs)
	videoCodecs := cm.selectVideoCodecs(prefs)

	return audioCodecs, videoCodecs, nil
}

// selectAudioCodecs 选择音频编解码器
func (cm *CodecManager) selectAudioCodecs(prefs *CodecPreferences) []string {
	var candidates []string

	// 根据网络条件筛选
	for _, codecName := range cm.preferences["audio"] {
		codec := cm.supportedCodecs[codecName]
		if codec == nil {
			continue
		}

		// 检查带宽要求
		if codec.Bandwidth > prefs.Bandwidth {
			continue
		}

		// 检查延迟要求
		if codec.Latency > prefs.Latency {
			continue
		}

		// 检查丢包率适应性
		if codec.PacketLoss < prefs.PacketLoss && !codec.NetworkAdaptive {
			continue
		}

		// 检查浏览器支持
		if browserSupported, exists := codec.BrowserSupport[prefs.Browser]; exists && !browserSupported {
			continue
		}

		// 检查移动设备支持
		if prefs.DeviceType == "mobile" && !codec.MobileSupport {
			continue
		}

		candidates = append(candidates, codecName)
	}

	// 根据网络类型调整优先级
	if networkPrefs, exists := cm.preferences[prefs.NetworkType]; exists {
		candidates = cm.reorderByPreferences(candidates, networkPrefs)
	}

	return candidates
}

// selectVideoCodecs 选择视频编解码器
func (cm *CodecManager) selectVideoCodecs(prefs *CodecPreferences) []string {
	var candidates []string

	// 根据网络条件筛选
	for _, codecName := range cm.preferences["video"] {
		codec := cm.supportedCodecs[codecName]
		if codec == nil {
			continue
		}

		// 检查带宽要求
		if codec.Bandwidth > prefs.Bandwidth {
			continue
		}

		// 检查延迟要求
		if codec.Latency > prefs.Latency {
			continue
		}

		// 检查丢包率适应性
		if codec.PacketLoss < prefs.PacketLoss && !codec.NetworkAdaptive {
			continue
		}

		// 检查浏览器支持
		if browserSupported, exists := codec.BrowserSupport[prefs.Browser]; exists && !browserSupported {
			continue
		}

		// 检查移动设备支持
		if prefs.DeviceType == "mobile" && !codec.MobileSupport {
			continue
		}

		// 检查硬件加速支持
		if prefs.HardwareAccel && !codec.HardwareAccel {
			continue
		}

		candidates = append(candidates, codecName)
	}

	// 根据网络类型调整优先级
	if networkPrefs, exists := cm.preferences[prefs.NetworkType]; exists {
		candidates = cm.reorderByPreferences(candidates, networkPrefs)
	}

	return candidates
}

// reorderByPreferences 根据偏好重新排序
func (cm *CodecManager) reorderByPreferences(candidates []string, preferences []string) []string {
	var reordered []string
	var remaining []string

	// 按照偏好顺序添加
	for _, pref := range preferences {
		for i, candidate := range candidates {
			if candidate == pref {
				reordered = append(reordered, candidate)
				candidates = append(candidates[:i], candidates[i+1:]...)
				break
			}
		}
	}

	// 添加剩余的编解码器
	reordered = append(reordered, candidates...)

	return reordered
}

// GetCodecInfo 获取编解码器信息
func (cm *CodecManager) GetCodecInfo(codecName string) (*CodecInfo, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	codec, exists := cm.supportedCodecs[codecName]
	if !exists {
		return nil, fmt.Errorf("编解码器 %s 不存在", codecName)
	}

	return codec, nil
}

// GetSupportedCodecs 获取支持的编解码器列表
func (cm *CodecManager) GetSupportedCodecs(codecType string) []*CodecInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var codecs []*CodecInfo
	for _, codec := range cm.supportedCodecs {
		if codecType == "" || codec.Type == codecType {
			codecs = append(codecs, codec)
		}
	}

	// 按名称排序
	sort.Slice(codecs, func(i, j int) bool {
		return codecs[i].Name < codecs[j].Name
	})

	return codecs
}

// CompareCodecs 比较编解码器
func (cm *CodecManager) CompareCodecs(codec1, codec2 string) (*CodecComparison, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	info1, exists1 := cm.supportedCodecs[codec1]
	info2, exists2 := cm.supportedCodecs[codec2]

	if !exists1 {
		return nil, fmt.Errorf("编解码器 %s 不存在", codec1)
	}
	if !exists2 {
		return nil, fmt.Errorf("编解码器 %s 不存在", codec2)
	}

	comparison := &CodecComparison{
		Codec1: info1,
		Codec2: info2,
		Winner: "tie",
		Reasons: []string{},
	}

	// 比较各项指标
	score1 := 0
	score2 := 0

	// 比较延迟
	if info1.Latency < info2.Latency {
		score1++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s延迟更低", info1.Name))
	} else if info1.Latency > info2.Latency {
		score2++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s延迟更低", info2.Name))
	}

	// 比较CPU使用率
	if info1.CPUUsage < info2.CPUUsage {
		score1++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%sCPU使用率更低", info1.Name))
	} else if info1.CPUUsage > info2.CPUUsage {
		score2++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%sCPU使用率更低", info2.Name))
	}

	// 比较丢包率适应性
	if info1.PacketLoss < info2.PacketLoss {
		score1++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s丢包率适应性更好", info1.Name))
	} else if info1.PacketLoss > info2.PacketLoss {
		score2++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s丢包率适应性更好", info2.Name))
	}

	// 比较网络自适应
	if info1.NetworkAdaptive && !info2.NetworkAdaptive {
		score1++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s支持网络自适应", info1.Name))
	} else if !info1.NetworkAdaptive && info2.NetworkAdaptive {
		score2++
		comparison.Reasons = append(comparison.Reasons, fmt.Sprintf("%s支持网络自适应", info2.Name))
	}

	// 确定获胜者
	if score1 > score2 {
		comparison.Winner = codec1
	} else if score2 > score1 {
		comparison.Winner = codec2
	}

	return comparison, nil
}

// CodecComparison 编解码器比较结果
type CodecComparison struct {
	Codec1  *CodecInfo `json:"codec1"`
	Codec2  *CodecInfo `json:"codec2"`
	Winner  string     `json:"winner"`
	Reasons []string   `json:"reasons"`
}

// GetCodecStats 获取编解码器统计信息
func (cm *CodecManager) GetCodecStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_codecs":     len(cm.supportedCodecs),
		"audio_codecs":     0,
		"video_codecs":     0,
		"network_adaptive": 0,
		"hardware_accel":   0,
		"mobile_support":   0,
	}

	for _, codec := range cm.supportedCodecs {
		switch codec.Type {
		case "audio":
			stats["audio_codecs"] = stats["audio_codecs"].(int) + 1
		case "video":
			stats["video_codecs"] = stats["video_codecs"].(int) + 1
		}

		if codec.NetworkAdaptive {
			stats["network_adaptive"] = stats["network_adaptive"].(int) + 1
		}
		if codec.HardwareAccel {
			stats["hardware_accel"] = stats["hardware_accel"].(int) + 1
		}
		if codec.MobileSupport {
			stats["mobile_support"] = stats["mobile_support"].(int) + 1
		}
	}

	return stats
}
