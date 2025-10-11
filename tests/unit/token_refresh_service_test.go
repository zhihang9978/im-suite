package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

// TestTokenRefreshService_GenerateRefreshToken 测试Refresh Token生成
func TestTokenRefreshService_GenerateRefreshToken(t *testing.T) {
	tokenService := service.NewTokenRefreshService()

	// 生成refresh token
	token, err := tokenService.GenerateRefreshToken(1, "13800138000")

	if err != nil {
		t.Skip("需要Redis支持，跳过测试")
	}

	assert.NotEmpty(t, token, "refresh token不应该为空")
	assert.Greater(t, len(token), 20, "refresh token长度应该大于20")
}

// TestTokenRefreshService_ValidateRefreshToken 测试Refresh Token验证
func TestTokenRefreshService_ValidateRefreshToken(t *testing.T) {
	tokenService := service.NewTokenRefreshService()

	// 生成refresh token
	token, err := tokenService.GenerateRefreshToken(1, "13800138000")
	if err != nil {
		t.Skip("需要Redis支持，跳过测试")
	}

	// 验证refresh token
	claims, err := tokenService.ValidateRefreshToken(token)
	assert.NoError(t, err, "验证refresh token不应该出错")
	assert.NotNil(t, claims, "claims不应该为空")
	assert.Equal(t, uint(1), claims.UserID, "UserID应该匹配")
	assert.Equal(t, "13800138000", claims.Phone, "Phone应该匹配")
}

// TestTokenRefreshService_RevokeRefreshToken 测试Token撤销
func TestTokenRefreshService_RevokeRefreshToken(t *testing.T) {
	tokenService := service.NewTokenRefreshService()

	phone := "13800138000"

	// 生成refresh token
	token, err := tokenService.GenerateRefreshToken(1, phone)
	if err != nil {
		t.Skip("需要Redis支持，跳过测试")
	}

	// 撤销token
	err = tokenService.RevokeRefreshToken(phone)
	assert.NoError(t, err, "撤销token不应该出错")

	// 验证token应该失败
	_, err = tokenService.ValidateRefreshToken(token)
	assert.Error(t, err, "撤销后的token应该无效")
}

// BenchmarkTokenRefreshService_GenerateRefreshToken 性能测试
func BenchmarkTokenRefreshService_GenerateRefreshToken(b *testing.B) {
	tokenService := service.NewTokenRefreshService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenService.GenerateRefreshToken(uint(i), "13800138000")
	}
}

