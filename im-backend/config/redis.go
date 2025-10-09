package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() error {
	// 从环境变量获取Redis配置
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0 // Redis数据库编号，默认0

	// 创建Redis客户端
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if err := Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %v", err)
	}

	return nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return Redis
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if Redis != nil {
		return Redis.Close()
	}
	return nil
}
