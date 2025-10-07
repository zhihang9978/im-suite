package service

import (
	"testing"
	"zhihang-messenger/im-backend/internal/model"
	"zhihang-messenger/im-backend/config"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MessageServiceTestSuite 消息服务测试套件
type MessageServiceTestSuite struct {
	suite.Suite
	db             *gorm.DB
	messageService *MessageService
	authService    *AuthService
	userID         uint
	chatID         uint
}

// SetupSuite 测试套件初始化
func (suite *MessageServiceTestSuite) SetupSuite() {
	// 使用内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)
	
	// 设置全局数据库实例
	config.DB = db
	
	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.User{},
		&model.Contact{},
		&model.Chat{},
		&model.ChatMember{},
		&model.Message{},
		&model.MessageRead{},
	)
	assert.NoError(suite.T(), err)
	
	suite.db = db
	suite.messageService = NewMessageService()
	suite.authService = NewAuthService()
}

// TearDownSuite 测试套件清理
func (suite *MessageServiceTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// SetupTest 每个测试前的准备
func (suite *MessageServiceTestSuite) SetupTest() {
	// 清理数据库
	suite.db.Exec("DELETE FROM message_reads")
	suite.db.Exec("DELETE FROM messages")
	suite.db.Exec("DELETE FROM chat_members")
	suite.db.Exec("DELETE FROM chats")
	suite.db.Exec("DELETE FROM contacts")
	suite.db.Exec("DELETE FROM users")
	
	// 创建测试用户
	user := model.User{
		Phone:    "13800138000",
		Username: "testuser",
		Nickname: "测试用户",
		Password: "hashed_password",
		Salt:     "salt",
		IsActive: true,
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)
	suite.userID = user.ID
	
	// 创建测试聊天
	chat := model.Chat{
		Name:        "测试聊天",
		Description: "这是一个测试聊天",
		Type:        "group",
		IsActive:    true,
	}
	err = suite.db.Create(&chat).Error
	assert.NoError(suite.T(), err)
	suite.chatID = chat.ID
	
	// 创建聊天成员
	member := model.ChatMember{
		ChatID:   chat.ID,
		UserID:   user.ID,
		Role:     "owner",
		JoinedAt: time.Now(),
	}
	err = suite.db.Create(&member).Error
	assert.NoError(suite.T(), err)
}

// TestSendMessage 测试发送消息
func (suite *MessageServiceTestSuite) TestSendMessage() {
	tests := []struct {
		name    string
		req     SendMessageRequest
		wantErr bool
	}{
		{
			name: "发送文本消息",
			req: SendMessageRequest{
				ChatID:  suite.chatID,
				Content: "这是一条测试消息",
				Type:    "text",
			},
			wantErr: false,
		},
		{
			name: "发送图片消息",
			req: SendMessageRequest{
				ChatID:   suite.chatID,
				Content:  "",
				Type:     "image",
				FileName: "test.jpg",
				FileSize: 1024,
				FileURL:  "http://example.com/test.jpg",
			},
			wantErr: false,
		},
		{
			name: "聊天不存在",
			req: SendMessageRequest{
				ChatID:  999,
				Content: "这是一条测试消息",
				Type:    "text",
			},
			wantErr: true,
		},
		{
			name: "用户不是聊天成员",
			req: SendMessageRequest{
				ChatID:  suite.chatID,
				Content: "这是一条测试消息",
				Type:    "text",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// 如果不是"用户不是聊天成员"的测试，使用正常用户ID
			userID := suite.userID
			if tt.name == "用户不是聊天成员" {
				userID = 999 // 使用不存在的用户ID
			}
			
			message, err := suite.messageService.SendMessage(userID, tt.req)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), message)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), message)
				assert.Equal(suite.T(), tt.req.ChatID, message.ChatID)
				assert.Equal(suite.T(), userID, message.SenderID)
				assert.Equal(suite.T(), tt.req.Content, message.Content)
				assert.Equal(suite.T(), tt.req.Type, message.Type)
			}
		})
	}
}

// TestGetMessages 测试获取消息列表
func (suite *MessageServiceTestSuite) TestGetMessages() {
	// 先创建一些测试消息
	for i := 0; i < 5; i++ {
		message := model.Message{
			ChatID:   suite.chatID,
			SenderID: suite.userID,
			Content:  fmt.Sprintf("测试消息 %d", i+1),
			Type:     "text",
			IsRead:   false,
		}
		err := suite.db.Create(&message).Error
		assert.NoError(suite.T(), err)
	}
	
	tests := []struct {
		name    string
		req     GetMessagesRequest
		wantErr bool
	}{
		{
			name: "获取消息列表",
			req: GetMessagesRequest{
				ChatID: suite.chatID,
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "分页获取消息",
			req: GetMessagesRequest{
				ChatID: suite.chatID,
				Limit:  2,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "搜索消息",
			req: GetMessagesRequest{
				ChatID: suite.chatID,
				Search: "测试消息 1",
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "用户不是聊天成员",
			req: GetMessagesRequest{
				ChatID: suite.chatID,
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// 如果不是"用户不是聊天成员"的测试，使用正常用户ID
			userID := suite.userID
			if tt.name == "用户不是聊天成员" {
				userID = 999 // 使用不存在的用户ID
			}
			
			resp, err := suite.messageService.GetMessages(userID, tt.req)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), resp)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), resp)
				assert.GreaterOrEqual(suite.T(), len(resp.Messages), 0)
				assert.GreaterOrEqual(suite.T(), resp.Total, int64(0))
				
				// 验证分页
				if tt.req.Limit > 0 {
					assert.LessOrEqual(suite.T(), len(resp.Messages), tt.req.Limit)
				}
			}
		})
	}
}

// TestEditMessage 测试编辑消息
func (suite *MessageServiceTestSuite) TestEditMessage() {
	// 先创建一条测试消息
	message := model.Message{
		ChatID:   suite.chatID,
		SenderID: suite.userID,
		Content:  "原始消息内容",
		Type:     "text",
		IsRead:   false,
	}
	err := suite.db.Create(&message).Error
	assert.NoError(suite.T(), err)
	
	// 测试编辑消息
	editedMessage, err := suite.messageService.EditMessage(message.ID, suite.userID, "编辑后的消息内容")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), editedMessage)
	assert.Equal(suite.T(), "编辑后的消息内容", editedMessage.Content)
	assert.True(suite.T(), editedMessage.IsEdited)
}

// TestDeleteMessage 测试删除消息
func (suite *MessageServiceTestSuite) TestDeleteMessage() {
	// 先创建一条测试消息
	message := model.Message{
		ChatID:   suite.chatID,
		SenderID: suite.userID,
		Content:  "要删除的消息",
		Type:     "text",
		IsRead:   false,
	}
	err := suite.db.Create(&message).Error
	assert.NoError(suite.T(), err)
	
	// 测试删除消息
	err = suite.messageService.DeleteMessage(message.ID, suite.userID)
	
	assert.NoError(suite.T(), err)
	
	// 验证消息已被软删除
	var deletedMessage model.Message
	err = suite.db.First(&deletedMessage, message.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), deletedMessage.IsDeleted)
}

// TestMarkAsRead 测试标记消息为已读
func (suite *MessageServiceTestSuite) TestMarkAsRead() {
	// 先创建一条测试消息
	message := model.Message{
		ChatID:   suite.chatID,
		SenderID: suite.userID,
		Content:  "要标记为已读的消息",
		Type:     "text",
		IsRead:   false,
	}
	err := suite.db.Create(&message).Error
	assert.NoError(suite.T(), err)
	
	// 测试标记为已读
	err = suite.messageService.MarkAsRead(message.ID, suite.userID)
	
	assert.NoError(suite.T(), err)
	
	// 验证消息已标记为已读
	var readMessage model.Message
	err = suite.db.First(&readMessage, message.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), readMessage.IsRead)
	
	// 验证已读记录已创建
	var readRecord model.MessageRead
	err = suite.db.Where("message_id = ? AND user_id = ?", message.ID, suite.userID).First(&readRecord).Error
	assert.NoError(suite.T(), err)
}

// TestCreateChat 测试创建聊天
func (suite *MessageServiceTestSuite) TestCreateChat() {
	req := CreateChatRequest{
		Name:        "新测试聊天",
		Description: "这是一个新创建的测试聊天",
		Type:        "group",
		Members:     []uint{suite.userID},
	}
	
	chat, err := suite.messageService.CreateChat(suite.userID, req)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), chat)
	assert.Equal(suite.T(), req.Name, chat.Name)
	assert.Equal(suite.T(), req.Description, chat.Description)
	assert.Equal(suite.T(), req.Type, chat.Type)
	assert.Greater(suite.T(), len(chat.Members), 0)
}

// TestGetChats 测试获取聊天列表
func (suite *MessageServiceTestSuite) TestGetChats() {
	chats, err := suite.messageService.GetChats(suite.userID)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), chats)
	assert.GreaterOrEqual(suite.T(), len(chats), 0)
}

// TestMessageServiceSuite 运行测试套件
func TestMessageServiceSuite(t *testing.T) {
	suite.Run(t, new(MessageServiceTestSuite))
}
