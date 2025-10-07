package service

import (
	"testing"
	"time"
	"zhihang-messenger/im-backend/internal/model"
	"zhihang-messenger/im-backend/config"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AuthServiceTestSuite 认证服务测试套件
type AuthServiceTestSuite struct {
	suite.Suite
	db         *gorm.DB
	authService *AuthService
}

// SetupSuite 测试套件初始化
func (suite *AuthServiceTestSuite) SetupSuite() {
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
	suite.authService = NewAuthService()
}

// TearDownSuite 测试套件清理
func (suite *AuthServiceTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// SetupTest 每个测试前的准备
func (suite *AuthServiceTestSuite) SetupTest() {
	// 清理数据库
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM contacts")
	suite.db.Exec("DELETE FROM chats")
	suite.db.Exec("DELETE FROM chat_members")
	suite.db.Exec("DELETE FROM messages")
	suite.db.Exec("DELETE FROM message_reads")
}

// TestRegister 测试用户注册
func (suite *AuthServiceTestSuite) TestRegister() {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{
			name: "正常注册",
			req: RegisterRequest{
				Phone:    "13800138000",
				Username: "testuser",
				Password: "password123",
				Nickname: "测试用户",
			},
			wantErr: false,
		},
		{
			name: "手机号已存在",
			req: RegisterRequest{
				Phone:    "13800138000",
				Username: "testuser2",
				Password: "password123",
				Nickname: "测试用户2",
			},
			wantErr: true,
		},
		{
			name: "用户名已存在",
			req: RegisterRequest{
				Phone:    "13800138001",
				Username: "testuser",
				Password: "password123",
				Nickname: "测试用户3",
			},
			wantErr: true,
		},
		{
			name: "密码太短",
			req: RegisterRequest{
				Phone:    "13800138002",
				Username: "testuser3",
				Password: "123",
				Nickname: "测试用户4",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.authService.Register(tt.req)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), resp)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), resp)
				assert.NotNil(suite.T(), resp.User)
				assert.NotEmpty(suite.T(), resp.AccessToken)
				assert.NotEmpty(suite.T(), resp.RefreshToken)
				assert.Greater(suite.T(), resp.ExpiresIn, int64(0))
			}
		})
	}
}

// TestLogin 测试用户登录
func (suite *AuthServiceTestSuite) TestLogin() {
	// 先注册一个用户
	registerReq := RegisterRequest{
		Phone:    "13800138000",
		Username: "testuser",
		Password: "password123",
		Nickname: "测试用户",
	}
	_, err := suite.authService.Register(registerReq)
	assert.NoError(suite.T(), err)
	
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name: "密码登录成功",
			req: LoginRequest{
				Phone:    "13800138000",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "验证码登录成功",
			req: LoginRequest{
				Phone: "13800138000",
				Code:  "123456",
			},
			wantErr: false,
		},
		{
			name: "用户不存在",
			req: LoginRequest{
				Phone:    "13800138001",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "密码错误",
			req: LoginRequest{
				Phone:    "13800138000",
				Password: "wrongpassword",
			},
			wantErr: true,
		},
		{
			name: "验证码错误",
			req: LoginRequest{
				Phone: "13800138000",
				Code:  "654321",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.authService.Login(tt.req)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), resp)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), resp)
				assert.NotNil(suite.T(), resp.User)
				assert.NotEmpty(suite.T(), resp.AccessToken)
				assert.NotEmpty(suite.T(), resp.RefreshToken)
				assert.Greater(suite.T(), resp.ExpiresIn, int64(0))
			}
		})
	}
}

// TestRefreshToken 测试刷新令牌
func (suite *AuthServiceTestSuite) TestRefreshToken() {
	// 先注册并登录一个用户
	registerReq := RegisterRequest{
		Phone:    "13800138000",
		Username: "testuser",
		Password: "password123",
		Nickname: "测试用户",
	}
	registerResp, err := suite.authService.Register(registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试刷新令牌
	refreshReq := RefreshRequest{
		RefreshToken: registerResp.RefreshToken,
	}
	
	resp, err := suite.authService.RefreshToken(refreshReq)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.NotEmpty(suite.T(), resp.AccessToken)
	assert.NotEmpty(suite.T(), resp.RefreshToken)
	assert.Greater(suite.T(), resp.ExpiresIn, int64(0))
}

// TestValidateToken 测试令牌验证
func (suite *AuthServiceTestSuite) TestValidateToken() {
	// 先注册并登录一个用户
	registerReq := RegisterRequest{
		Phone:    "13800138000",
		Username: "testuser",
		Password: "password123",
		Nickname: "测试用户",
	}
	registerResp, err := suite.authService.Register(registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试令牌验证
	user, err := suite.authService.ValidateToken(registerResp.AccessToken)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "13800138000", user.Phone)
	assert.Equal(suite.T(), "testuser", user.Username)
	assert.Equal(suite.T(), "测试用户", user.Nickname)
}

// TestLogout 测试用户登出
func (suite *AuthServiceTestSuite) TestLogout() {
	// 先注册并登录一个用户
	registerReq := RegisterRequest{
		Phone:    "13800138000",
		Username: "testuser",
		Password: "password123",
		Nickname: "测试用户",
	}
	registerResp, err := suite.authService.Register(registerReq)
	assert.NoError(suite.T(), err)
	
	// 测试登出
	err = suite.authService.Logout(registerResp.AccessToken)
	
	assert.NoError(suite.T(), err)
}

// TestRegisterSuite 运行测试套件
func TestRegisterSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
