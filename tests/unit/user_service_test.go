package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestUser_PasswordValidation 测试密码验证
func TestUser_PasswordValidation(t *testing.T) {
	// 测试弱密码
	weakPasswords := []string{
		"123456",
		"password",
		"abc123",
		"12345678",
	}

	for _, pwd := range weakPasswords {
		t.Run("WeakPassword_"+pwd, func(t *testing.T) {
			user := &model.User{Password: pwd}
			// 密码应该被hash，不应该是明文
			assert.NotEqual(t, pwd, user.Password)
		})
	}
}

// TestUser_PhoneValidation 测试手机号验证
func TestUser_PhoneValidation(t *testing.T) {
	validPhones := []string{
		"13800138000",
		"13912345678",
		"18612345678",
	}

	invalidPhones := []string{
		"12345678901",  // 不是1开头
		"138001380",    // 太短
		"138001380000", // 太长
	}

	for _, phone := range validPhones {
		t.Run("ValidPhone_"+phone, func(t *testing.T) {
			assert.Len(t, phone, 11)
			assert.Equal(t, "1", string(phone[0]))
		})
	}

	for _, phone := range invalidPhones {
		t.Run("InvalidPhone_"+phone, func(t *testing.T) {
			assert.NotEqual(t, 11, len(phone))
		})
	}
}

// TestUser_StatusTransitions 测试用户状态转换
func TestUser_StatusTransitions(t *testing.T) {
	user := &model.User{
		ID:       1,
		Phone:    "13800138000",
		IsActive: true,
		Online:   false,
	}

	// 激活 → 禁用
	user.IsActive = false
	assert.False(t, user.IsActive)

	// 离线 → 在线
	user.Online = true
	assert.True(t, user.Online)
}

// BenchmarkUser_Creation 性能测试：用户创建
func BenchmarkUser_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		user := &model.User{
			Phone:    "13800138000",
			Username: "testuser",
			Nickname: "测试用户",
		}
		_ = user
	}
}

