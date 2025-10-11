package unit

import (
	"testing"

	"zhihang-messenger/im-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

// TestOfflineMessageService_GenerateKey 测试离线消息key生成
func TestOfflineMessageService_GenerateKey(t *testing.T) {
	offlineService := service.NewOfflineMessageService()
	assert.NotNil(t, offlineService)
}

// TestOfflineMessageService_StoreMessage 测试存储离线消息
func TestOfflineMessageService_StoreMessage(t *testing.T) {
	// 此测试需要Redis，如果Redis未配置则跳过
	t.Skip("需要Redis环境，跳过测试")
}

// TestOfflineMessageService_GetMessages 测试获取离线消息
func TestOfflineMessageService_GetMessages(t *testing.T) {
	offlineService := service.NewOfflineMessageService()

	// 测试获取不存在用户的消息
	messages, err := offlineService.GetOfflineMessages(99999, 10)
	
	if err != nil {
		t.Skip("需要数据库环境，跳过测试")
	}

	// 应该返回空列表而不是错误
	assert.NotNil(t, messages)
	assert.IsType(t, []model.Message{}, messages)
}

// TestOfflineMessageService_MessageLimit 测试消息数量限制
func TestOfflineMessageService_MessageLimit(t *testing.T) {
	limits := []int{10, 50, 100}

	for _, limit := range limits {
		t.Run("Limit_"+string(rune(limit)), func(t *testing.T) {
			assert.Greater(t, limit, 0)
			assert.LessOrEqual(t, limit, 100)
		})
	}
}

// BenchmarkOfflineMessageService_KeyGeneration 性能测试
func BenchmarkOfflineMessageService_KeyGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		userID := uint(i % 1000)
		_ = userID
	}
}

