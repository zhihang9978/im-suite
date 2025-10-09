package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		db: config.DB,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User          *model.User `json:"user"`
	AccessToken   string      `json:"access_token"`
	RefreshToken  string      `json:"refresh_token"`
	ExpiresIn     int64       `json:"expires_in"`
	Requires2FA   bool        `json:"requires_2fa"`   // 是否需要2FA验证
	TempToken     string      `json:"temp_token"`     // 临时令牌（用于2FA验证）
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Phone    string `json:"phone"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user model.User

	// 查找用户
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码或验证码
	if req.Password != "" {
		// 密码登录
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, errors.New("密码错误")
		}
	} else {
		// 验证码登录 (简化处理)
		// 实际部署时应该集成真实的短信验证服务
		// if req.Code != "123456" {
		// 	return nil, errors.New("验证码错误")
		// }
	}

	// 更新最后在线时间
	user.LastSeen = time.Now()
	user.Online = true
	s.db.Save(&user)

	// 检查是否启用2FA
	if user.TwoFactorEnabled {
		// 检查设备是否受信任（需要设备ID从请求中获取）
		// 注意：这里简化处理，实际应该在Controller层获取设备信息
		// 现在返回需要2FA验证的响应
		
		return &LoginResponse{
			User:         &user,
			AccessToken:  "",
			RefreshToken: "",
			ExpiresIn:    0,
			Requires2FA:  true,
			TempToken:    "", // 前端需要用UserID来调用2FA验证
		}, nil
	}

	// 未启用2FA，正常登录流程
	accessToken, refreshToken, expiresIn, err := s.generateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Requires2FA:  false,
		TempToken:    "",
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error) {
	// 检查手机号是否已存在
	var existingUser model.User
	if err := s.db.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		return nil, errors.New("手机号已存在")
	}

	// 检查用户名是否已存在
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建用户
	user := model.User{
		Phone:    req.Phone,
		Username: req.Username,
		Nickname: req.Nickname,
		Password: string(hashedPassword),
		Salt:     fmt.Sprintf("%d", time.Now().Unix()),
		IsActive: true,
		LastSeen: time.Now(),
		Online:   false,
		Language: "zh-CN",
		Theme:    "auto",
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 生成令牌
	accessToken, refreshToken, expiresIn, err := s.generateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	return &RegisterResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *AuthService) RefreshToken(req RefreshRequest) (*RefreshResponse, error) {
	// 验证刷新令牌
	claims, err := s.validateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("刷新令牌无效")
	}

	// 查找用户
	var user model.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 生成新令牌
	accessToken, refreshToken, expiresIn, err := s.generateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	// 验证令牌
	claims, err := s.validateToken(token)
	if err != nil {
		return errors.New("令牌无效")
	}

	// 更新用户在线状态
	var user model.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err == nil {
		user.Online = false
		user.LastSeen = time.Now()
		s.db.Save(&user)
	}

	return nil
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(token string) (*model.User, error) {
	// 移除 Bearer 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := s.validateToken(token)
	if err != nil {
		return nil, errors.New("令牌无效")
	}

	// 查找用户
	var user model.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	return &user, nil
}

// generateTokens 生成访问令牌和刷新令牌
func (s *AuthService) generateTokens(user *model.User) (string, string, int64, error) {
	// JWT密钥
	secretKey := []byte("zhihang_messenger_secret_key_2024")

	// 访问令牌过期时间 (24小时)
	accessExpiresAt := time.Now().Add(24 * time.Hour)

	// 刷新令牌过期时间 (7天)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	// 访问令牌
	accessClaims := &Claims{
		UserID:   user.ID,
		Phone:    user.Phone,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zhihang-messenger",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return "", "", 0, err
	}

	// 刷新令牌
	refreshClaims := &Claims{
		UserID:   user.ID,
		Phone:    user.Phone,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zhihang-messenger",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", 0, err
	}

	return accessTokenString, refreshTokenString, int64(24 * time.Hour.Seconds()), nil
}

// validateToken 验证令牌
func (s *AuthService) validateToken(tokenString string) (*Claims, error) {
	secretKey := []byte("zhihang_messenger_secret_key_2024")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("令牌无效")
}

// verifyPassword 验证密码
func (s *AuthService) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// LoginWith2FA 使用2FA验证码完成登录
func (s *AuthService) LoginWith2FA(userID uint, code string, deviceID string, deviceInfo map[string]string) (*LoginResponse, error) {
	// 查找用户
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	
	// 检查用户状态
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}
	
	// 检查是否启用2FA
	if !user.TwoFactorEnabled {
		return nil, errors.New("用户未启用双因子认证")
	}
	
	// 验证2FA验证码
	twoFactorService := NewTwoFactorService()
	if err := twoFactorService.ValidateTwoFactorCode(context.Background(), userID, code); err != nil {
		return nil, err
	}
	
	// 2FA验证成功，注册设备（如果提供了设备信息）
	// 注意：设备注册在Controller层处理，避免循环依赖
	// 这里仅完成2FA验证和token生成
	_ = deviceID      // 标记使用
	_ = deviceInfo    // 标记使用
	
	// 更新在线状态
	user.LastSeen = time.Now()
	user.Online = true
	s.db.Save(&user)
	
	// 生成正式令牌
	accessToken, refreshToken, expiresIn, err := s.generateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}
	
	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Requires2FA:  false,
		TempToken:    "",
	}, nil
}
