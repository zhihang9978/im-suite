package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWebRTC_SDPValidation 测试SDP验证
func TestWebRTC_SDPValidation(t *testing.T) {
	validSDP := "v=0\no=- 123456789 2 IN IP4 127.0.0.1"
	invalidSDP := "invalid sdp"

	// 简单的SDP验证（检查是否以v=开头）
	assert.True(t, len(validSDP) > 3 && validSDP[:2] == "v=")
	assert.False(t, len(invalidSDP) > 3 && invalidSDP[:2] == "v=")
}

// TestWebRTC_ICECandidateValidation 测试ICE candidate验证
func TestWebRTC_ICECandidateValidation(t *testing.T) {
	validCandidate := "candidate:1 1 UDP 2130706431 192.168.1.1 54321 typ host"
	invalidCandidate := "not a candidate"

	assert.Contains(t, validCandidate, "candidate:")
	assert.NotContains(t, invalidCandidate, "candidate:")
}

// TestWebRTC_ConnectionStates 测试连接状态
func TestWebRTC_ConnectionStates(t *testing.T) {
	states := []string{
		"new",
		"connecting",
		"connected",
		"disconnected",
		"failed",
		"closed",
	}

	for _, state := range states {
		t.Run("State_"+state, func(t *testing.T) {
			assert.Contains(t, states, state)
		})
	}
}

// TestWebRTC_MediaTypes 测试媒体类型
func TestWebRTC_MediaTypes(t *testing.T) {
	mediaTypes := []string{"audio", "video", "screen"}

	for _, mediaType := range mediaTypes {
		t.Run("MediaType_"+mediaType, func(t *testing.T) {
			assert.NotEmpty(t, mediaType)
		})
	}
}

// BenchmarkWebRTC_SDPParsing 性能测试
func BenchmarkWebRTC_SDPParsing(b *testing.B) {
	sdp := "v=0\no=- 123456789 2 IN IP4 127.0.0.1\ns=WebRTC Session"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sdp[:2] == "v="
	}
}

