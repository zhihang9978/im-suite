package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

// TestMessageACKService_GenerateMessageID 测试消息ID生成
func TestMessageACKService_GenerateMessageID(t *testing.T) {
	ackService := service.NewMessageACKService()

	// 生成多个消息ID
	id1 := ackService.GenerateMessageID(1)
	id2 := ackService.GenerateMessageID(1)
	id3 := ackService.GenerateMessageID(2)

	// 验证ID不为空
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEmpty(t, id3)

	// 验证ID唯一性
	assert.NotEqual(t, id1, id2, "同一用户生成的ID应该不同")
	assert.NotEqual(t, id1, id3, "不同用户生成的ID应该不同")
}

// TestMessageACKService_CheckDuplicate 测试消息去重
func TestMessageACKService_CheckDuplicate(t *testing.T) {
	ackService := service.NewMessageACKService()

	// 生成一个消息ID
	messageID := ackService.GenerateMessageID(1)

	// 第一次检查，应该不重复
	isDuplicate := ackService.CheckDuplicate(messageID)
	assert.False(t, isDuplicate, "第一次检查应该不重复")

	// 标记消息已发送
	err := ackService.MarkMessageSent(messageID)
	if err != nil {
		t.Skip("需要Redis支持，跳过测试")
	}

	// 第二次检查，应该重复
	isDuplicate = ackService.CheckDuplicate(messageID)
	assert.True(t, isDuplicate, "第二次检查应该重复")
}

// TestMessageACKService_MarkMessageSent 测试标记消息已发送
func TestMessageACKService_MarkMessageSent(t *testing.T) {
	ackService := service.NewMessageACKService()

	messageID := ackService.GenerateMessageID(1)

	err := ackService.MarkMessageSent(messageID)
	if err != nil {
		t.Skip("需要Redis支持，跳过测试")
	}

	// 验证消息已标记
	isDuplicate := ackService.CheckDuplicate(messageID)
	assert.True(t, isDuplicate)
}

// BenchmarkMessageACKService_GenerateMessageID 性能测试：ID生成
func BenchmarkMessageACKService_GenerateMessageID(b *testing.B) {
	ackService := service.NewMessageACKService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ackService.GenerateMessageID(uint(i % 1000))
	}
}

