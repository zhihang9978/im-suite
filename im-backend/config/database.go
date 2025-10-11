package config

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	// 从环境变量获取数据库配置
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	username := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	database := getEnv("DB_NAME", "zhihang_messenger")

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, database)

	// 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                          // 空闲连接数
	sqlDB.SetMaxOpenConns(100)                         // 最大连接数
	sqlDB.SetConnMaxLifetime(30 * time.Minute)         // 连接最大生命周期：30分钟
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)         // 空闲连接超时：10分钟

	return nil
}

// AutoMigrate 自动迁移数据库表
// 使用优化的迁移逻辑，确保依赖关系正确
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 使用新的迁移模块
	return MigrateTables(DB)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
