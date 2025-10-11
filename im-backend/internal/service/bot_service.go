package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"zhihang-messenger/im-backend/config"
	"zhihang-messenger/im-backend/internal/model"
)

// BotModel 机器人模型包装（用于中间件传递）
type BotModel struct {
	Bot *model.Bot
}

// BotService 机器人服务
type BotService struct {
	db *gorm.DB
}

// NewBotService 创建机器人服务实例
func NewBotService() *BotService {
	return &BotService{
		db: config.DB,
	}
}

// CreateBotRequest 创建机器人请求
type CreateBotRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Type        string   `json:"type" binding:"required"` // internal, webhook, plugin
	Permissions []string `json:"permissions" binding:"required"`
}

// CreateBotResponse 创建机器人响应
type CreateBotResponse struct {
	Bot       *model.Bot `json:"bot"`
	APIKey    string     `json:"api_key"`    // 只在创建时返回一次
	APISecret string     `json:"api_secret"` // 只在创建时返回一次
}

// BotCreateUserRequest 机器人创建用户请求
type BotCreateUserRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"` // user, admin (不能创建super_admin)
}

// BotBanUserRequest 机器人封禁用户请求
type BotBanUserRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Duration int64  `json:"duration"` // 秒，0表示永久
	Reason   string `json:"reason" binding:"required"`
}

// BotDeleteUserRequest 机器人删除用户请求
type BotDeleteUserRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// CreateBot 创建机器人
func (s *BotService) CreateBot(ctx context.Context, adminID uint, req *CreateBotRequest) (*CreateBotResponse, error) {
	// 验证创建者是超级管理员
	var admin model.User
	if err := s.db.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return nil, errors.New("管理员不存在")
	}

	if admin.Role != "super_admin" {
		return nil, errors.New("只有超级管理员可以创建机器人")
	}

	// 检查机器人名称是否已存在
	var existing model.Bot
	if err := s.db.WithContext(ctx).Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, errors.New("机器人名称已存在")
	}

	// 生成API密钥
	apiKey, apiSecret, err := s.generateAPIKeys()
	if err != nil {
		return nil, fmt.Errorf("生成API密钥失败: %w", err)
	}

	// 加密API Secret
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("加密API密钥失败: %w", err)
	}

	// 权限转JSON
	permissionsJSON, _ := json.Marshal(req.Permissions)

	// 创建机器人
	bot := model.Bot{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		APIKey:      apiKey,
		APISecret:   string(hashedSecret),
		Permissions: string(permissionsJSON),
		IsActive:    true,
		CreatedBy:   adminID,
		RateLimit:   100,
		DailyLimit:  10000,
	}

	if err := s.db.WithContext(ctx).Create(&bot).Error; err != nil {
		return nil, fmt.Errorf("创建机器人失败: %w", err)
	}

	return &CreateBotResponse{
		Bot:       &bot,
		APIKey:    apiKey,
		APISecret: apiSecret, // 明文返回，只显示一次
	}, nil
}

// ValidateBotAPIKey 验证机器人API密钥
func (s *BotService) ValidateBotAPIKey(ctx context.Context, apiKey, apiSecret string) (*model.Bot, error) {
	var bot model.Bot
	if err := s.db.WithContext(ctx).Where("api_key = ? AND is_active = ?", apiKey, true).First(&bot).Error; err != nil {
		return nil, errors.New("无效的API密钥")
	}

	// 验证API Secret
	if err := bcrypt.CompareHashAndPassword([]byte(bot.APISecret), []byte(apiSecret)); err != nil {
		return nil, errors.New("API密钥验证失败")
	}

	// 更新最后使用时间
	bot.LastUsedAt = time.Now()
	s.db.WithContext(ctx).Save(&bot)

	return &bot, nil
}

// CheckBotPermission 检查机器人权限
func (s *BotService) CheckBotPermission(bot *model.Bot, permission model.BotPermission) bool {
	var permissions []string
	if err := json.Unmarshal([]byte(bot.Permissions), &permissions); err != nil {
		return false
	}

	for _, p := range permissions {
		if p == string(permission) {
			return true
		}
	}

	return false
}

// BotCreateUser 机器人创建用户（仅限普通用户）
func (s *BotService) BotCreateUser(ctx context.Context, bot *model.Bot, req *BotCreateUserRequest) (*model.User, error) {
	// 检查权限
	if !s.CheckBotPermission(bot, model.PermissionCreateUser) {
		return nil, errors.New("机器人没有创建用户的权限")
	}

	// 检查手机号是否已存在
	var existingUser model.User
	if err := s.db.WithContext(ctx).Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		return nil, errors.New("手机号已存在")
	}

	// 检查用户名是否已存在
	if err := s.db.WithContext(ctx).Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 机器人只能创建普通用户，忽略请求中的role字段
	// 创建用户
	user := model.User{
		Phone:          req.Phone,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Password:       string(hashedPassword),
		Salt:           fmt.Sprintf("%d", time.Now().Unix()),
		IsActive:       true,
		Role:           "user", // 固定为普通用户
		LastSeen:       time.Now(),
		Online:         false,
		Language:       "zh-CN",
		Theme:          "auto",
		CreatedByBotID: &bot.ID, // 标记创建者机器人
		BotManageable:  true,    // 允许机器人管理
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 记录机器人操作
	s.logBotOperation(ctx, bot.ID, "create_user", fmt.Sprintf("创建用户: %s (ID:%d)", user.Username, user.ID), true, "")

	return &user, nil
}

// BotBanUser 机器人封禁用户
func (s *BotService) BotBanUser(ctx context.Context, bot *model.Bot, req *BotBanUserRequest) error {
	// 检查权限
	if !s.CheckBotPermission(bot, model.PermissionBanUser) {
		return errors.New("机器人没有封禁用户的权限")
	}

	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.UserID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 不能封禁超级管理员
	if user.Role == "super_admin" {
		return errors.New("不能封禁超级管理员")
	}

	// 计算封禁到期时间
	var banUntil *time.Time
	if req.Duration > 0 {
		t := time.Now().Add(time.Duration(req.Duration) * time.Second)
		banUntil = &t
	}

	// 封禁用户
	updates := map[string]interface{}{
		"is_banned":  true,
		"ban_reason": req.Reason,
		"online":     false,
	}
	if banUntil != nil {
		updates["ban_until"] = banUntil
	}

	if err := s.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("封禁失败: %w", err)
	}

	// 记录机器人操作
	details := fmt.Sprintf("封禁用户: %s, 原因: %s", user.Username, req.Reason)
	s.logBotOperation(ctx, bot.ID, "ban_user", details, true, "")

	return nil
}

// BotUnbanUser 机器人解封用户
func (s *BotService) BotUnbanUser(ctx context.Context, bot *model.Bot, userID uint) error {
	// 检查权限
	if !s.CheckBotPermission(bot, model.PermissionUnbanUser) {
		return errors.New("机器人没有解封用户的权限")
	}

	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 解封
	updates := map[string]interface{}{
		"is_banned":  false,
		"ban_until":  nil,
		"ban_reason": "",
	}

	if err := s.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("解封失败: %w", err)
	}

	// 记录操作
	s.logBotOperation(ctx, bot.ID, "unban_user", fmt.Sprintf("解封用户: %s", user.Username), true, "")

	return nil
}

// BotDeleteUser 机器人删除用户（仅限自己创建的用户）
func (s *BotService) BotDeleteUser(ctx context.Context, bot *model.Bot, req *BotDeleteUserRequest) error {
	// 检查权限
	if !s.CheckBotPermission(bot, model.PermissionDeleteUser) {
		return errors.New("机器人没有删除用户的权限")
	}

	// 查找用户
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.UserID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 只能删除被标记为可被机器人管理的用户
	if !user.BotManageable {
		return errors.New("该用户不允许被机器人管理")
	}

	// 只能删除自己创建的用户
	if user.CreatedByBotID == nil || *user.CreatedByBotID != bot.ID {
		return errors.New("只能删除本机器人创建的用户")
	}

	// 软删除用户
	if err := s.db.WithContext(ctx).Delete(&user).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 记录操作
	details := fmt.Sprintf("删除用户: %s (ID:%d), 原因: %s", user.Username, user.ID, req.Reason)
	s.logBotOperation(ctx, bot.ID, "delete_user", details, true, "")

	return nil
}

// GetBotList 获取机器人列表
func (s *BotService) GetBotList(ctx context.Context) ([]model.Bot, error) {
	var bots []model.Bot
	err := s.db.WithContext(ctx).Preload("Creator").Order("created_at DESC").Find(&bots).Error
	return bots, err
}

// GetBotByID 获取机器人详情
func (s *BotService) GetBotByID(ctx context.Context, botID uint) (*model.Bot, error) {
	var bot model.Bot
	err := s.db.WithContext(ctx).Preload("Creator").First(&bot, botID).Error
	if err != nil {
		return nil, errors.New("机器人不存在")
	}
	return &bot, nil
}

// UpdateBotPermissions 更新机器人权限
func (s *BotService) UpdateBotPermissions(ctx context.Context, botID uint, permissions []string) error {
	permissionsJSON, _ := json.Marshal(permissions)

	return s.db.WithContext(ctx).Model(&model.Bot{}).
		Where("id = ?", botID).
		Update("permissions", string(permissionsJSON)).Error
}

// ToggleBotStatus 切换机器人状态
func (s *BotService) ToggleBotStatus(ctx context.Context, botID uint, isActive bool) error {
	return s.db.WithContext(ctx).Model(&model.Bot{}).
		Where("id = ?", botID).
		Update("is_active", isActive).Error
}

// DeleteBot 删除机器人
func (s *BotService) DeleteBot(ctx context.Context, botID uint) error {
	return s.db.WithContext(ctx).Delete(&model.Bot{}, botID).Error
}

// GetBotLogs 获取机器人调用日志
func (s *BotService) GetBotLogs(ctx context.Context, botID uint, limit int) ([]model.BotAPILog, error) {
	var logs []model.BotAPILog
	query := s.db.WithContext(ctx).Where("bot_id = ?", botID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetBotStats 获取机器人统计
func (s *BotService) GetBotStats(ctx context.Context, botID uint) (map[string]interface{}, error) {
	var bot model.Bot
	if err := s.db.WithContext(ctx).First(&bot, botID).Error; err != nil {
		return nil, errors.New("机器人不存在")
	}

	// 今日调用次数
	today := time.Now().Truncate(24 * time.Hour)
	var todayCalls int64
	s.db.WithContext(ctx).Model(&model.BotAPILog{}).
		Where("bot_id = ? AND created_at >= ?", botID, today).
		Count(&todayCalls)

	// 成功率
	successRate := 0.0
	if bot.TotalCalls > 0 {
		successRate = float64(bot.SuccessCalls) / float64(bot.TotalCalls) * 100
	}

	stats := map[string]interface{}{
		"total_calls":   bot.TotalCalls,
		"success_calls": bot.SuccessCalls,
		"failed_calls":  bot.FailedCalls,
		"success_rate":  successRate,
		"today_calls":   todayCalls,
		"last_used_at":  bot.LastUsedAt,
		"is_active":     bot.IsActive,
	}

	return stats, nil
}

// generateAPIKeys 生成API密钥对
func (s *BotService) generateAPIKeys() (string, string, error) {
	// 生成API Key (32字节)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", err
	}
	apiKey := "bot_" + hex.EncodeToString(keyBytes)

	// 生成API Secret (32字节，避免bcrypt 72字节限制)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	apiSecret := hex.EncodeToString(secretBytes)

	return apiKey, apiSecret, nil
}

// logBotOperation 记录机器人操作
func (s *BotService) logBotOperation(ctx context.Context, botID uint, operation, details string, success bool, errorMsg string) {
	log := model.BotAPILog{
		BotID:    botID,
		Endpoint: operation,
		Method:   "BOT_OPERATION",
		StatusCode: func() int {
			if success {
				return 200
			}
			return 500
		}(),
		Duration: 0,
		Error:    errorMsg,
	}

	s.db.WithContext(ctx).Create(&log)

	// 更新机器人统计
	if success {
		s.db.WithContext(ctx).Model(&model.Bot{}).Where("id = ?", botID).
			UpdateColumn("success_calls", gorm.Expr("success_calls + ?", 1)).
			UpdateColumn("total_calls", gorm.Expr("total_calls + ?", 1))
	} else {
		s.db.WithContext(ctx).Model(&model.Bot{}).Where("id = ?", botID).
			UpdateColumn("failed_calls", gorm.Expr("failed_calls + ?", 1)).
			UpdateColumn("total_calls", gorm.Expr("total_calls + ?", 1))
	}
}

// RecordBotAPICall 记录Bot API调用
func (s *BotService) RecordBotAPICall(ctx context.Context, botID uint, endpoint, method string, statusCode int, duration int64, reqBody, respBody, errorMsg string) {
	log := model.BotAPILog{
		BotID:        botID,
		Endpoint:     endpoint,
		Method:       method,
		StatusCode:   statusCode,
		Duration:     duration,
		RequestBody:  reqBody,
		ResponseBody: respBody,
		Error:        errorMsg,
	}

	s.db.WithContext(ctx).Create(&log)
}

// CheckRateLimit 检查速率限制
func (s *BotService) CheckRateLimit(ctx context.Context, bot *model.Bot) error {
	// 使用Redis检查速率限制
	redis := config.GetRedis()
	if redis == nil {
		return nil // Redis不可用时跳过限制
	}

	// 分钟级限制
	minuteKey := fmt.Sprintf("bot:ratelimit:minute:%d", bot.ID)
	count, _ := redis.Get(ctx, minuteKey).Int()

	if count >= bot.RateLimit {
		return fmt.Errorf("速率限制：超过%d次/分钟", bot.RateLimit)
	}

	// 增加计数
	redis.Incr(ctx, minuteKey)
	redis.Expire(ctx, minuteKey, time.Minute)

	// 每日限制
	dayKey := fmt.Sprintf("bot:ratelimit:day:%d", bot.ID)
	dayCount, _ := redis.Get(ctx, dayKey).Int()

	if dayCount >= bot.DailyLimit {
		return fmt.Errorf("每日限制：超过%d次/天", bot.DailyLimit)
	}

	redis.Incr(ctx, dayKey)
	redis.Expire(ctx, dayKey, 24*time.Hour)

	return nil
}

// RegenerateAPISecret 重新生成API密钥
func (s *BotService) RegenerateAPISecret(ctx context.Context, botID uint) (string, error) {
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}
	apiSecret := hex.EncodeToString(secretBytes)

	// 加密
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// 更新
	if err := s.db.WithContext(ctx).Model(&model.Bot{}).
		Where("id = ?", botID).
		Update("api_secret", string(hashedSecret)).Error; err != nil {
		return "", err
	}

	return apiSecret, nil
}
