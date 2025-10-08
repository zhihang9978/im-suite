package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"gorm.io/gorm"
	"zhihang-messenger/im-backend/internal/model"
)

// MessageEncryptionService 消息加密服务
type MessageEncryptionService struct {
	db *gorm.DB
}

// NewMessageEncryptionService 创建消息加密服务
func NewMessageEncryptionService(db *gorm.DB) *MessageEncryptionService {
	return &MessageEncryptionService{
		db: db,
	}
}

// EncryptMessageRequest 加密消息请求
type EncryptMessageRequest struct {
	MessageID        uint   `json:"message_id" binding:"required"`
	EncryptionType   string `json:"encryption_type" binding:"required"` // "simple", "e2e"
	Password         string `json:"password,omitempty"`
	SelfDestructTime *int   `json:"self_destruct_time,omitempty"` // 秒数
}

// DecryptMessageRequest 解密消息请求
type DecryptMessageRequest struct {
	MessageID uint   `json:"message_id" binding:"required"`
	Password  string `json:"password,omitempty"`
}

// MessageKey 消息密钥
type MessageKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	KeyHash   string    `gorm:"type:varchar(255);not null" json:"key_hash"`
	KeyData   string    `gorm:"type:text" json:"key_data"` // 加密后的密钥数据
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Message model.Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// SelfDestructLog 自毁日志
type SelfDestructLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	DestructTime time.Time `json:"destruct_time"`
	Reason    string    `gorm:"type:varchar(255)" json:"reason"`

	// 关联
	Message model.Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

// EncryptMessage 加密消息
func (s *MessageEncryptionService) EncryptMessage(ctx context.Context, userID uint, req *EncryptMessageRequest) error {
	var message model.Message
	
	// 查找消息
	if err := s.db.WithContext(ctx).First(&message, req.MessageID).Error; err != nil {
		return fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限：只有发送者可以加密消息
	if message.SenderID != userID {
		return fmt.Errorf("没有权限加密此消息")
	}

	// 检查消息是否已被撤回
	if message.IsRecalled {
		return fmt.Errorf("已撤回的消息不能加密")
	}

	// 检查消息是否已加密
	if message.IsEncrypted {
		return fmt.Errorf("消息已加密")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成加密密钥
	var encryptionKey []byte
	var keyHash string
	
	if req.EncryptionType == "simple" {
		// 简单加密：使用用户密码生成密钥
		if req.Password == "" {
			tx.Rollback()
			return fmt.Errorf("简单加密需要提供密码")
		}
		hash := sha256.Sum256([]byte(req.Password + fmt.Sprintf("%d", message.ID)))
		encryptionKey = hash[:]
		keyHash = fmt.Sprintf("%x", hash)
	} else if req.EncryptionType == "e2e" {
		// 端到端加密：生成随机密钥
		encryptionKey = make([]byte, 32)
		if _, err := rand.Read(encryptionKey); err != nil {
			tx.Rollback()
			return fmt.Errorf("生成加密密钥失败: %w", err)
		}
		keyHash = fmt.Sprintf("%x", sha256.Sum256(encryptionKey))
	} else {
		tx.Rollback()
		return fmt.Errorf("不支持的加密类型")
	}

	// 加密消息内容
	encryptedContent, err := s.encryptContent(message.Content, encryptionKey)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("加密消息内容失败: %w", err)
	}

	// 更新消息
	updates := map[string]interface{}{
		"content":      encryptedContent,
		"is_encrypted": true,
		"updated_at":   time.Now(),
	}

	// 设置自毁时间
	if req.SelfDestructTime != nil {
		selfDestructTime := time.Now().Add(time.Duration(*req.SelfDestructTime) * time.Second)
		updates["is_self_destruct"] = true
		updates["self_destruct_time"] = &selfDestructTime
	}

	if err := tx.Model(&message).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新消息失败: %w", err)
	}

	// 保存密钥信息
	messageKey := &MessageKey{
		MessageID: message.ID,
		KeyHash:   keyHash,
		KeyData:   base64.StdEncoding.EncodeToString(encryptionKey),
	}

	if err := tx.Create(messageKey).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存密钥信息失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DecryptMessage 解密消息
func (s *MessageEncryptionService) DecryptMessage(ctx context.Context, userID uint, req *DecryptMessageRequest) (string, error) {
	var message model.Message
	var messageKey MessageKey
	
	// 查找消息
	if err := s.db.WithContext(ctx).Preload("Sender").First(&message, req.MessageID).Error; err != nil {
		return "", fmt.Errorf("消息不存在: %w", err)
	}

	// 检查消息是否已加密
	if !message.IsEncrypted {
		return message.Content, nil // 未加密，直接返回原内容
	}

	// 检查权限：只有发送者和接收者可以解密
	if message.SenderID != userID && (message.ReceiverID == nil || *message.ReceiverID != userID) {
		// 检查是否为群成员
		if message.ChatID != nil {
			var chatMember model.ChatMember
			if err := s.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", 
				*message.ChatID, userID).First(&chatMember).Error; err != nil {
				return "", fmt.Errorf("没有权限解密此消息")
			}
		} else {
			return "", fmt.Errorf("没有权限解密此消息")
		}
	}

	// 查找密钥信息
	if err := s.db.WithContext(ctx).Where("message_id = ?", message.ID).First(&messageKey).Error; err != nil {
		return "", fmt.Errorf("找不到消息密钥: %w", err)
	}

	// 解密密钥
	encryptionKey, err := base64.StdEncoding.DecodeString(messageKey.KeyData)
	if err != nil {
		return "", fmt.Errorf("解码密钥失败: %w", err)
	}

	// 验证密钥（对于简单加密）
	if req.Password != "" {
		expectedHash := sha256.Sum256([]byte(req.Password + fmt.Sprintf("%d", message.ID)))
		if fmt.Sprintf("%x", expectedHash) != messageKey.KeyHash {
			return "", fmt.Errorf("密码错误")
		}
	}

	// 解密消息内容
	decryptedContent, err := s.decryptContent(message.Content, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("解密消息内容失败: %w", err)
	}

	return decryptedContent, nil
}

// ProcessSelfDestructMessages 处理自毁消息（定时任务调用）
func (s *MessageEncryptionService) ProcessSelfDestructMessages(ctx context.Context) error {
	now := time.Now()
	var messages []model.Message

	// 查找到期的自毁消息
	if err := s.db.WithContext(ctx).Where("is_self_destruct = ? AND self_destruct_time <= ?", 
		true, now).Find(&messages).Error; err != nil {
		return fmt.Errorf("查找自毁消息失败: %w", err)
	}

	for _, message := range messages {
		// 开始事务
		tx := s.db.WithContext(ctx).Begin()

		// 删除消息密钥
		if err := tx.Where("message_id = ?", message.ID).Delete(&MessageKey{}).Error; err != nil {
			tx.Rollback()
			continue
		}

		// 记录自毁日志
		selfDestructLog := &SelfDestructLog{
			MessageID:    message.ID,
			DestructTime: now,
			Reason:       "自动自毁",
		}

		if err := tx.Create(selfDestructLog).Error; err != nil {
			tx.Rollback()
			continue
		}

		// 更新消息状态
		if err := tx.Model(&message).Updates(map[string]interface{}{
			"content":             "[消息已自毁]",
			"is_self_destruct":    false,
			"self_destruct_time":  nil,
			"status":              "destroyed",
		}).Error; err != nil {
			tx.Rollback()
			continue
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			continue
		}
	}

	return nil
}

// GetEncryptedMessageInfo 获取加密消息信息
func (s *MessageEncryptionService) GetEncryptedMessageInfo(ctx context.Context, messageID uint, userID uint) (map[string]interface{}, error) {
	var message model.Message
	var messageKey MessageKey
	
	// 查找消息
	if err := s.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		return nil, fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限
	if message.SenderID != userID && (message.ReceiverID == nil || *message.ReceiverID != userID) {
		if message.ChatID != nil {
			var chatMember model.ChatMember
			if err := s.db.WithContext(ctx).Where("chat_id = ? AND user_id = ?", 
				*message.ChatID, userID).First(&chatMember).Error; err != nil {
				return nil, fmt.Errorf("没有权限访问此消息")
			}
		} else {
			return nil, fmt.Errorf("没有权限访问此消息")
		}
	}

	info := map[string]interface{}{
		"is_encrypted":      message.IsEncrypted,
		"is_self_destruct":  message.IsSelfDestruct,
		"message_type":      message.MessageType,
		"created_at":        message.CreatedAt,
	}

	if message.IsEncrypted {
		if err := s.db.WithContext(ctx).Where("message_id = ?", message.ID).First(&messageKey).Error; err == nil {
			info["key_hash"] = messageKey.KeyHash
		}
	}

	if message.IsSelfDestruct && message.SelfDestructTime != nil {
		info["self_destruct_time"] = message.SelfDestructTime
		info["time_remaining"] = message.SelfDestructTime.Sub(time.Now()).Seconds()
	}

	return info, nil
}

// encryptContent 加密内容
func (s *MessageEncryptionService) encryptContent(content string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(content), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptContent 解密内容
func (s *MessageEncryptionService) decryptContent(encryptedContent string, key []byte) (string, error) {
	// 解码base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedContent)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 提取nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("密文长度不足")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// SetMessageSelfDestruct 设置消息自毁时间
func (s *MessageEncryptionService) SetMessageSelfDestruct(ctx context.Context, messageID uint, userID uint, destructTime int) error {
	var message model.Message
	
	// 查找消息
	if err := s.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		return fmt.Errorf("消息不存在: %w", err)
	}

	// 检查权限：只有发送者可以设置自毁时间
	if message.SenderID != userID {
		return fmt.Errorf("没有权限设置自毁时间")
	}

	// 检查消息是否已被撤回
	if message.IsRecalled {
		return fmt.Errorf("已撤回的消息不能设置自毁时间")
	}

	// 计算自毁时间
	selfDestructTime := time.Now().Add(time.Duration(destructTime) * time.Second)

	// 更新消息
	if err := s.db.WithContext(ctx).Model(&message).Updates(map[string]interface{}{
		"is_self_destruct":   true,
		"self_destruct_time": &selfDestructTime,
		"updated_at":         time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("更新消息自毁时间失败: %w", err)
	}

	return nil
}
