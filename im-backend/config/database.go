package config

import (
	"fmt"
	"os"
	"time"
	"zhihang-messenger/im-backend/internal/model"

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
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(3600 * time.Second) // 1小时

	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&model.User{},
		&model.Contact{},
		&model.Session{},
		&model.Chat{},
		&model.ChatMember{},
		&model.Message{},
		&model.MessageRead{},
		&model.MessageEdit{},
		&model.MessageRecall{},
		&model.MessageForward{},
		&model.ScheduledMessage{},
		&model.MessageSearchIndex{},
		&model.MessagePin{},
		&model.MessageMark{},
		&model.MessageStatus{},
		&model.MessageShare{},
		&model.MessageReply{},
		&model.File{},
		&model.FileChunk{},
		&model.FilePreview{},
		&model.FileAccess{},
		&model.ContentReport{},
		&model.ContentFilter{},
		&model.UserWarning{},
		&model.ModerationLog{},
		&model.ContentStatistics{},
		&model.Theme{},
		&model.UserThemeSetting{},
		&model.ThemeTemplate{},
		&model.GroupInvite{},
		&model.GroupInviteUsage{},
		&model.AdminRole{},
		&model.ChatAdmin{},
		&model.GroupJoinRequest{},
		&model.GroupAuditLog{},
		&model.GroupPermissionTemplate{},
		&model.Alert{},
		&model.AdminOperationLog{},
		&model.SystemConfig{},
		&model.IPBlacklist{},
		&model.UserBlacklist{},
		&model.LoginAttempt{},
		&model.SuspiciousActivity{},
		&model.TwoFactorAuth{},
		&model.TrustedDevice{},
		&model.DeviceSession{},
		&model.DeviceActivity{},
		&model.Bot{},
		&model.BotAPILog{},
	)

	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
