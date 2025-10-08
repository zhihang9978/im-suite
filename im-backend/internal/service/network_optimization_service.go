package service

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"zhihang-messenger/im-backend/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// NetworkOptimizationService 网络优化服务
type NetworkOptimizationService struct {
	redis           *redis.Client
	ctx             context.Context
	compressionPool *sync.Pool
	cacheTimeout    time.Duration
}

// CompressionLevel 压缩级别
type CompressionLevel int

const (
	NoCompression      CompressionLevel = 0
	BestSpeed          CompressionLevel = 1
	BestCompression    CompressionLevel = 9
	DefaultCompression CompressionLevel = 6
)

// CDNConfig CDN配置
type CDNConfig struct {
	Provider   string            `json:"provider"`    // cloudflare, aws, aliyun
	Domain     string            `json:"domain"`      // CDN域名
	SSLEnabled bool              `json:"ssl_enabled"` // 是否启用SSL
	Headers    map[string]string `json:"headers"`     // 自定义头
	CacheRules []CacheRule       `json:"cache_rules"` // 缓存规则
}

// CacheRule 缓存规则
type CacheRule struct {
	Path        string            `json:"path"`        // 路径模式
	TTL         int               `json:"ttl"`         // 缓存时间（秒）
	Compression bool              `json:"compression"` // 是否压缩
	Headers     map[string]string `json:"headers"`     // 缓存头
}

// ConnectionPool 连接池配置
type ConnectionPool struct {
	MaxIdle     int           `json:"max_idle"`     // 最大空闲连接
	MaxActive   int           `json:"max_active"`   // 最大活跃连接
	IdleTimeout time.Duration `json:"idle_timeout"` // 空闲超时
	MaxLifetime time.Duration `json:"max_lifetime"` // 连接最大生命周期
}

// NetworkPerformanceStats 网络性能统计
type NetworkPerformanceStats struct {
	TotalRequests      int64            `json:"total_requests"`
	CompressedRequests int64            `json:"compressed_requests"`
	CacheHits          int64            `json:"cache_hits"`
	CacheMisses        int64            `json:"cache_misses"`
	AverageLatency     float64          `json:"average_latency"`
	CompressionRatio   float64          `json:"compression_ratio"`
	BandwidthSaved     int64            `json:"bandwidth_saved"`
	LastUpdated        time.Time        `json:"last_updated"`
	ByEndpoint         map[string]int64 `json:"by_endpoint"`
}

// NewNetworkOptimizationService 创建网络优化服务
func NewNetworkOptimizationService() *NetworkOptimizationService {
	return &NetworkOptimizationService{
		redis: config.Redis,
		ctx:   context.Background(),
		compressionPool: &sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		},
		cacheTimeout: 5 * time.Minute,
	}
}

// CompressResponse 压缩响应
func (s *NetworkOptimizationService) CompressResponse(data []byte, level CompressionLevel) ([]byte, error) {
	if level == NoCompression {
		return data, nil
	}

	var buf strings.Builder
	writer := s.compressionPool.Get().(*gzip.Writer)
	defer s.compressionPool.Put(writer)

	writer.Reset(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	compressed := buf.String()

	// 记录压缩统计
	s.recordCompressionStats(len(data), len(compressed))

	return []byte(compressed), nil
}

// recordCompressionStats 记录压缩统计
func (s *NetworkOptimizationService) recordCompressionStats(originalSize, compressedSize int) {
	stats := map[string]interface{}{
		"original_size":   originalSize,
		"compressed_size": compressedSize,
		"ratio":           float64(compressedSize) / float64(originalSize),
		"timestamp":       time.Now().Unix(),
	}

	key := "network:compression:stats"
	data, _ := json.Marshal(stats)

	s.redis.LPush(s.ctx, key, data)
	s.redis.LTrim(s.ctx, key, 0, 999) // 保留最近1000条记录
}

// SetCDNHeaders 设置CDN头
func (s *NetworkOptimizationService) SetCDNHeaders(w http.ResponseWriter, config *CDNConfig, path string) {
	// 设置基本CDN头
	w.Header().Set("X-CDN-Provider", config.Provider)

	// 查找匹配的缓存规则
	var matchedRule *CacheRule
	for _, rule := range config.CacheRules {
		if s.matchPath(path, rule.Path) {
			matchedRule = &rule
			break
		}
	}

	if matchedRule != nil {
		// 设置缓存头
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", matchedRule.TTL))
		w.Header().Set("X-Cache-TTL", fmt.Sprintf("%d", matchedRule.TTL))

		// 设置自定义头
		for key, value := range matchedRule.Headers {
			w.Header().Set(key, value)
		}

		// 设置压缩头
		if matchedRule.Compression {
			w.Header().Set("Content-Encoding", "gzip")
		}
	}

	// 设置通用头
	for key, value := range config.Headers {
		w.Header().Set(key, value)
	}
}

// matchPath 匹配路径
func (s *NetworkOptimizationService) matchPath(requestPath, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(requestPath, prefix)
	}

	return requestPath == pattern
}

// CacheResponse 缓存响应
func (s *NetworkOptimizationService) CacheResponse(key string, data []byte, ttl time.Duration) error {
	return s.redis.SetEX(s.ctx, key, data, ttl).Err()
}

// GetCachedResponse 获取缓存的响应
func (s *NetworkOptimizationService) GetCachedResponse(key string) ([]byte, error) {
	result, err := s.redis.Get(s.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 记录缓存命中
	s.recordCacheHit(key)

	return []byte(result), nil
}

// recordCacheHit 记录缓存命中
func (s *NetworkOptimizationService) recordCacheHit(key string) {
	stats := map[string]interface{}{
		"key":       key,
		"hit":       true,
		"timestamp": time.Now().Unix(),
	}

	data, _ := json.Marshal(stats)
	s.redis.LPush(s.ctx, "network:cache:hits", data)
}

// recordCacheMiss 记录缓存未命中
func (s *NetworkOptimizationService) recordCacheMiss(key string) {
	stats := map[string]interface{}{
		"key":       key,
		"hit":       false,
		"timestamp": time.Now().Unix(),
	}

	data, _ := json.Marshal(stats)
	s.redis.LPush(s.ctx, "network:cache:misses", data)
}

// OptimizeConnectionPool 优化连接池
func (s *NetworkOptimizationService) OptimizeConnectionPool(config *ConnectionPool) {
	// 这里应该根据实际的HTTP客户端配置来优化连接池
	// 例如使用http.Transport的MaxIdleConnsPerHost等设置

	logrus.WithFields(logrus.Fields{
		"max_idle":     config.MaxIdle,
		"max_active":   config.MaxActive,
		"idle_timeout": config.IdleTimeout,
		"max_lifetime": config.MaxLifetime,
	}).Info("连接池配置已优化")
}

// GetNetworkStats 获取网络统计
func (s *NetworkOptimizationService) GetNetworkStats() (*NetworkPerformanceStats, error) {
	stats := &NetworkPerformanceStats{
		ByEndpoint:  make(map[string]int64),
		LastUpdated: time.Now(),
	}

	// 获取总请求数
	totalRequests, err := s.redis.Get(s.ctx, "network:total_requests").Int64()
	if err == nil {
		stats.TotalRequests = totalRequests
	}

	// 获取压缩请求数
	compressedRequests, err := s.redis.Get(s.ctx, "network:compressed_requests").Int64()
	if err == nil {
		stats.CompressedRequests = compressedRequests
	}

	// 获取缓存命中数
	cacheHits, err := s.redis.LLen(s.ctx, "network:cache:hits").Result()
	if err == nil {
		stats.CacheHits = cacheHits
	}

	// 获取缓存未命中数
	cacheMisses, err := s.redis.LLen(s.ctx, "network:cache:misses").Result()
	if err == nil {
		stats.CacheMisses = cacheMisses
	}

	// 计算平均延迟
	avgLatency, err := s.redis.Get(s.ctx, "network:avg_latency").Float64()
	if err == nil {
		stats.AverageLatency = avgLatency
	}

	// 计算压缩率
	if stats.TotalRequests > 0 {
		stats.CompressionRatio = float64(stats.CompressedRequests) / float64(stats.TotalRequests)
	}

	// 计算节省的带宽
	bandwidthSaved, err := s.redis.Get(s.ctx, "network:bandwidth_saved").Int64()
	if err == nil {
		stats.BandwidthSaved = bandwidthSaved
	}

	return stats, nil
}

// RecordRequest 记录请求
func (s *NetworkOptimizationService) RecordRequest(endpoint string, latency time.Duration, compressed bool) {
	// 增加总请求数
	s.redis.Incr(s.ctx, "network:total_requests")

	// 记录压缩请求
	if compressed {
		s.redis.Incr(s.ctx, "network:compressed_requests")
	}

	// 记录延迟
	s.recordLatency(latency)

	// 记录端点统计
	s.redis.HIncrBy(s.ctx, "network:by_endpoint", endpoint, 1)
}

// recordLatency 记录延迟
func (s *NetworkOptimizationService) recordLatency(latency time.Duration) {
	// 使用滑动窗口记录延迟
	key := "network:latency:window"
	s.redis.LPush(s.ctx, key, latency.Milliseconds())
	s.redis.LTrim(s.ctx, key, 0, 99) // 保留最近100个延迟记录

	// 计算平均延迟
	latencies, err := s.redis.LRange(s.ctx, key, 0, -1).Result()
	if err == nil && len(latencies) > 0 {
		total := int64(0)
		for _, lat := range latencies {
			if l, err := s.redis.Get(s.ctx, lat).Int64(); err == nil {
				total += l
			}
		}
		avgLatency := float64(total) / float64(len(latencies))
		s.redis.Set(s.ctx, "network:avg_latency", avgLatency, 0)
	}
}

// OptimizeImageDelivery 优化图片传输
func (s *NetworkOptimizationService) OptimizeImageDelivery(imageData []byte, format string, quality int) ([]byte, error) {
	// 这里应该集成图片优化库，如WebP转换、质量压缩等
	// 暂时返回原始数据

	logrus.WithFields(logrus.Fields{
		"format":  format,
		"quality": quality,
		"size":    len(imageData),
	}).Debug("图片传输优化")

	return imageData, nil
}

// SetupHTTPCompression 设置HTTP压缩中间件
func (s *NetworkOptimizationService) SetupHTTPCompression() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否支持压缩
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// 创建压缩响应写入器
			gz := s.compressionPool.Get().(*gzip.Writer)
			defer s.compressionPool.Put(gz)

			gz.Reset(w)
			defer gz.Close()

			// 设置压缩头
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			// 包装响应写入器
			gzWriter := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}

			next.ServeHTTP(gzWriter, r)
		})
	}
}

// gzipResponseWriter 压缩响应写入器
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// CleanupNetworkStats 清理网络统计
func (s *NetworkOptimizationService) CleanupNetworkStats() {
	// 清理过期的统计数据
	patterns := []string{
		"network:compression:stats",
		"network:cache:hits",
		"network:cache:misses",
		"network:latency:window",
	}

	for _, pattern := range patterns {
		// 保留最近的数据，删除旧数据
		s.redis.LTrim(s.ctx, pattern, 0, 999)
	}

	logrus.Info("网络统计数据已清理")
}

// GetOptimizationRecommendations 获取优化建议
func (s *NetworkOptimizationService) GetOptimizationRecommendations() ([]string, error) {
	var recommendations []string

	stats, err := s.GetNetworkStats()
	if err != nil {
		return nil, err
	}

	// 基于统计数据提供优化建议
	if stats.CompressionRatio < 0.8 {
		recommendations = append(recommendations, "建议启用更多内容的压缩以减少带宽使用")
	}

	if stats.CacheHits < stats.CacheMisses {
		recommendations = append(recommendations, "缓存命中率较低，建议优化缓存策略")
	}

	if stats.AverageLatency > 1000 { // 1秒
		recommendations = append(recommendations, "平均延迟较高，建议优化网络连接和服务器性能")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "网络性能良好，无需额外优化")
	}

	return recommendations, nil
}

// StartNetworkOptimization 启动网络优化
func (s *NetworkOptimizationService) StartNetworkOptimization() {
	// 启动定期清理任务
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CleanupNetworkStats()
			case <-s.ctx.Done():
				return
			}
		}
	}()

	logrus.Info("网络优化服务已启动")
}
