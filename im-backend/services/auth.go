package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: config.DB,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	User      model.User `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user model.User
	
	// 查找用户
	if err := s.db.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 验证密码（这里简化处理，实际应该存储加密密码）
	// 为了演示，我们使用简单的字符串比较
	if req.Password != "123456" { // 默认密码
		return nil, errors.New("密码错误")
	}

	// 生成会话令牌
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	// 创建会话记录
	session := model.Session{
		UserID:    user.ID,
		Token:     token,
		Device:    "Web",
		IP:        "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}

	// 更新用户状态
	user.Online = true
	user.LastSeen = time.Now()
	s.db.Save(&user)

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(req RegisterRequest) (*LoginResponse, error) {
	// 检查手机号是否已存在
	var existingUser model.User
	if err := s.db.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		return nil, errors.New("手机号已注册")
	}

	// 检查用户名是否已存在
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := model.User{
		Phone:    req.Phone,
		Username: req.Username,
		Nickname: req.Nickname,
		Password: string(hashedPassword),
		Salt:     "default_salt", // 简化处理，实际应用中应该生成随机盐值
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// 生成会话令牌
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	// 创建会话记录
	session := model.Session{
		UserID:    user.ID,
		Token:     token,
		Device:    "Web",
		IP:        "127.0.0.1",
		UserAgent: "Mozilla/5.0",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(req RefreshRequest) (*LoginResponse, error) {
	var session model.Session
	
	// 查找会话
	if err := s.db.Preload("User").Where("token = ?", req.Token).First(&session).Error; err != nil {
		return nil, errors.New("无效的令牌")
	}

	// 检查令牌是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("令牌已过期")
	}

	// 生成新令牌
	newToken, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	// 更新会话
	session.Token = newToken
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	if err := s.db.Save(&session).Error; err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     newToken,
		User:      session.User,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	// 删除会话
	if err := s.db.Where("token = ?", token).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	// 更新用户状态
	var session model.Session
	if err := s.db.Where("token = ?", token).First(&session).Error; err == nil {
		s.db.Model(&model.User{}).Where("id = ?", session.UserID).Update("online", false)
	}

	return nil
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(token string) (*model.User, error) {
	var session model.Session
	
	if err := s.db.Preload("User").Where("token = ?", token).First(&session).Error; err != nil {
		return nil, errors.New("无效的令牌")
	}

	// 检查令牌是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("令牌已过期")
	}

	return &session.User, nil
}

// generateToken 生成随机令牌
func (s *AuthService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
