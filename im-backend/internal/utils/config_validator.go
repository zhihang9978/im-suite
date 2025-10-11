package utils

import (
	"fmt"
	"os"
	"strings"
)

// RequiredEnvVars 必需的环境变量列表
var RequiredEnvVars = []string{
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASSWORD",
	"DB_NAME",
	"REDIS_HOST",
	"REDIS_PORT",
	"REDIS_PASSWORD",
	"JWT_SECRET",
}

// ValidateRequiredEnv 验证必需的环境变量
func ValidateRequiredEnv() error {
	var missing []string

	for _, key := range RequiredEnvVars {
		value := os.Getenv(key)
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("缺少必需的环境变量: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateJWTSecret 验证JWT密钥长度
func ValidateJWTSecret() error {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET长度不足32字符，当前长度: %d", len(secret))
	}
	return nil
}

// ValidateProduction 生产环境验证
func ValidateProduction() error {
	// 检查必需环境变量
	if err := ValidateRequiredEnv(); err != nil {
		return err
	}

	// 检查JWT密钥长度
	if err := ValidateJWTSecret(); err != nil {
		return err
	}

	// 检查DEBUG模式
	debug := os.Getenv("DEBUG")
	if debug == "true" {
		return fmt.Errorf("生产环境不应启用DEBUG模式")
	}

	// 检查GIN模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "release" {
		return fmt.Errorf("生产环境GIN_MODE应为release，当前: %s", ginMode)
	}

	return nil
}

// GetEnv 获取环境变量，支持默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvRequired 获取必需的环境变量
func GetEnvRequired(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("环境变量 %s 未设置", key)
	}
	return value, nil
}

