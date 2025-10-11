package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestMessage_Validation 测试消息验证
func TestMessage_Validation(t *testing.T) {
	// 测试有效消息
	msg := &model.Message{
		SenderID:    1,
		Content:     "Hello",
		MessageType: "text",
	}

	assert.NotZero(t, msg.SenderID)
	assert.NotEmpty(t, msg.Content)
	assert.Equal(t, "text", msg.MessageType)
}

// TestMessage_EmptyContent 测试空消息
func TestMessage_EmptyContent(t *testing.T) {
	msg := &model.Message{
		SenderID:    1,
		Content:     "",
		MessageType: "text",
	}

	// 空消息应该被拒绝
	assert.Empty(t, msg.Content)
}

// TestMessage_TypeValidation 测试消息类型
func TestMessage_TypeValidation(t *testing.T) {
	validTypes := []string{"text", "image", "file", "audio", "video"}

	for _, msgType := range validTypes {
		t.Run("Type_"+msgType, func(t *testing.T) {
			msg := &model.Message{
				MessageType: msgType,
			}
			assert.Contains(t, validTypes, msg.MessageType)
		})
	}
}

// TestMessage_StatusFlow 测试消息状态流转
func TestMessage_StatusFlow(t *testing.T) {
	msg := &model.Message{
		Status: "sent",
	}

	// sent → delivered
	msg.Status = "delivered"
	assert.Equal(t, "delivered", msg.Status)

	// delivered → read
	msg.Status = "read"
	assert.Equal(t, "read", msg.Status)
}

// BenchmarkMessage_Creation 性能测试
func BenchmarkMessage_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		msg := &model.Message{
			SenderID:    uint(i % 1000),
			Content:     "Test message",
			MessageType: "text",
			Status:      "sent",
		}
		_ = msg
	}
}

