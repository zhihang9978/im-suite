package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

// TestAuthService_GenerateToken 测试Token生成
func TestAuthService_GenerateToken(t *testing.T) {
	authService := service.NewAuthService()

	// 测试生成token
	token, err := authService.GenerateToken(1, "13800138000")

	assert.NoError(t, err, "生成token不应该出错")
	assert.NotEmpty(t, token, "token不应该为空")
	assert.Greater(t, len(token), 20, "token长度应该大于20")
}

// TestAuthService_ValidateToken 测试Token验证
func TestAuthService_ValidateToken(t *testing.T) {
	authService := service.NewAuthService()

	// 生成一个token
	token, err := authService.GenerateToken(1, "13800138000")
	assert.NoError(t, err)

	// 验证token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err, "验证token不应该出错")
	assert.NotNil(t, claims, "claims不应该为空")
	assert.Equal(t, uint(1), claims.UserID, "UserID应该匹配")
	assert.Equal(t, "13800138000", claims.Phone, "Phone应该匹配")
}

// TestAuthService_ValidateInvalidToken 测试无效Token
func TestAuthService_ValidateInvalidToken(t *testing.T) {
	authService := service.NewAuthService()

	// 测试无效token
	_, err := authService.ValidateToken("invalid-token")
	assert.Error(t, err, "无效token应该返回错误")
}

// TestAuthService_ValidateExpiredToken 测试过期Token
func TestAuthService_ValidateExpiredToken(t *testing.T) {
	// 这个测试需要mock时间，暂时跳过
	t.Skip("需要mock时间来测试过期token")
}

// BenchmarkAuthService_GenerateToken 性能测试：Token生成
func BenchmarkAuthService_GenerateToken(b *testing.B) {
	authService := service.NewAuthService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.GenerateToken(uint(i), "13800138000")
	}
}

// BenchmarkAuthService_ValidateToken 性能测试：Token验证
func BenchmarkAuthService_ValidateToken(b *testing.B) {
	authService := service.NewAuthService()
	token, _ := authService.GenerateToken(1, "13800138000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.ValidateToken(token)
	}
}

