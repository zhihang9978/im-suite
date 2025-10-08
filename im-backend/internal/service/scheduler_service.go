package service

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	db *gorm.DB
	messageAdvancedService *MessageAdvancedService
	messageEncryptionService *MessageEncryptionService
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(db *gorm.DB, messageAdvancedService *MessageAdvancedService, messageEncryptionService *MessageEncryptionService) *SchedulerService {
	return &SchedulerService{
		db: db,
		messageAdvancedService: messageAdvancedService,
		messageEncryptionService: messageEncryptionService,
	}
}

// StartScheduler 启动定时任务调度器
func (s *SchedulerService) StartScheduler(ctx context.Context) {
	log.Println("启动定时任务调度器...")

	// 启动定时消息执行任务
	go s.scheduleMessageExecutor(ctx)
	
	// 启动自毁消息处理任务
	go s.selfDestructProcessor(ctx)
	
	// 启动清理任务
	go s.cleanupTask(ctx)
	
	log.Println("定时任务调度器启动完成")
}

// scheduleMessageExecutor 定时消息执行器
func (s *SchedulerService) scheduleMessageExecutor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("定时消息执行器停止")
			return
		case <-ticker.C:
			if err := s.messageAdvancedService.ExecuteScheduledMessages(ctx); err != nil {
				log.Printf("执行定时消息失败: %v", err)
			}
		}
	}
}

// selfDestructProcessor 自毁消息处理器
func (s *SchedulerService) selfDestructProcessor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("自毁消息处理器停止")
			return
		case <-ticker.C:
			if err := s.messageEncryptionService.ProcessSelfDestructMessages(ctx); err != nil {
				log.Printf("处理自毁消息失败: %v", err)
			}
		}
	}
}

// cleanupTask 清理任务
func (s *SchedulerService) cleanupTask(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // 每小时执行一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("清理任务停止")
			return
		case <-ticker.C:
			s.cleanupExpiredData(ctx)
		}
	}
}

// cleanupExpiredData 清理过期数据
func (s *SchedulerService) cleanupExpiredData(ctx context.Context) {
	log.Println("开始清理过期数据...")

	// 清理过期的消息编辑历史（保留最近30天）
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := s.db.WithContext(ctx).Where("edit_time < ?", thirtyDaysAgo).
		Delete(&model.MessageEdit{}).Error; err != nil {
		log.Printf("清理过期编辑历史失败: %v", err)
	}

	// 清理过期的消息撤回记录（保留最近90天）
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	if err := s.db.WithContext(ctx).Where("recall_time < ?", ninetyDaysAgo).
		Delete(&model.MessageRecall{}).Error; err != nil {
		log.Printf("清理过期撤回记录失败: %v", err)
	}

	// 清理过期的消息转发记录（保留最近30天）
	if err := s.db.WithContext(ctx).Where("forward_time < ?", thirtyDaysAgo).
		Delete(&model.MessageForward{}).Error; err != nil {
		log.Printf("清理过期转发记录失败: %v", err)
	}

	// 清理已执行的定时消息记录（保留最近7天）
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.WithContext(ctx).Where("is_executed = ? AND execute_time < ?", 
		true, sevenDaysAgo).Delete(&model.ScheduledMessage{}).Error; err != nil {
		log.Printf("清理过期定时消息记录失败: %v", err)
	}

	log.Println("过期数据清理完成")
}
