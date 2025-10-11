package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"zhihang-messenger/im-backend/config"
)

// TokenRefreshService Token刷新服务
type TokenRefreshService struct {
	jwtSecret []byte
}

// RefreshTokenClaims Refresh Token的Claims
type RefreshTokenClaims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

// NewTokenRefreshService 创建Token刷新服务
func NewTokenRefreshService() *TokenRefreshService {
	return &TokenRefreshService{
		jwtSecret: []byte(getEnv("JWT_SECRET", "default-secret-key")),
	}
}

// GenerateRefreshToken 生成Refresh Token
func (s *TokenRefreshService) GenerateRefreshToken(userID uint, phone string) (string, error) {
	// Refresh Token有效期7天
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	
	claims := &RefreshTokenClaims{
		UserID: userID,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "zhihang-messenger",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	// 将Refresh Token存储到Redis（支持撤销）
	if config.Redis != nil {
		key := "refresh_token:" + phone
		err = config.Redis.Set(config.Redis.Context(), key, tokenString, 7*24*time.Hour).Err()
		if err != nil {
			return "", err
		}
	}

	return tokenString, nil
}

// ValidateRefreshToken 验证Refresh Token
func (s *TokenRefreshService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	// 检查Redis中是否存在（未被撤销）
	if config.Redis != nil {
		key := "refresh_token:" + claims.Phone
		storedToken, err := config.Redis.Get(config.Redis.Context(), key).Result()
		if err != nil || storedToken != tokenString {
			return nil, errors.New("refresh token has been revoked")
		}
	}

	return claims, nil
}

// RevokeRefreshToken 撤销Refresh Token
func (s *TokenRefreshService) RevokeRefreshToken(phone string) error {
	if config.Redis != nil {
		key := "refresh_token:" + phone
		return config.Redis.Del(config.Redis.Context(), key).Err()
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	// 简化版本，实际应该从config包获取
	return defaultValue
}

