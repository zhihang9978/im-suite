package config

import (
	"fmt"
	"log"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/gorm"
)

// MigrationInfo 迁移信息
type MigrationInfo struct {
	Model interface{}
	Name  string
	Deps  []string // 依赖的表名
}

// GetMigrationOrder 获取正确的迁移顺序
// 按照外键依赖关系排序，确保被引用的表先创建
func GetMigrationOrder() []MigrationInfo {
	return []MigrationInfo{
		// =======================================
		// 第一层：基础表（无外键依赖）
		// =======================================
		{Model: &model.User{}, Name: "users", Deps: []string{}},
		{Model: &model.Chat{}, Name: "chats", Deps: []string{}},
		{Model: &model.Theme{}, Name: "themes", Deps: []string{}},

		// =======================================
		// 第二层：依赖基础表
		// =======================================
		{Model: &model.Session{}, Name: "sessions", Deps: []string{"users"}},
		{Model: &model.Contact{}, Name: "contacts", Deps: []string{"users"}},
		{Model: &model.ChatMember{}, Name: "chat_members", Deps: []string{"chats", "users"}},
		{Model: &model.UserThemeSetting{}, Name: "user_theme_settings", Deps: []string{"users", "themes"}},
		{Model: &model.ThemeTemplate{}, Name: "theme_templates", Deps: []string{"themes"}},

	// =======================================
	// 第三层：消息主表（有自引用，可以创建）
	// =======================================
	{Model: &model.Message{}, Name: "messages", Deps: []string{"users", "chats"}},

	// =======================================
	// 第四层：消息回复链（依赖 Message）
	// =======================================
	{Model: &model.MessageReply{}, Name: "message_replies", Deps: []string{"messages"}},

		// =======================================
		// 第五层：消息相关表（依赖 Message）
		// =======================================
		{Model: &model.MessageRead{}, Name: "message_reads", Deps: []string{"messages", "users"}},
		{Model: &model.MessageEdit{}, Name: "message_edits", Deps: []string{"messages"}},
		{Model: &model.MessageRecall{}, Name: "message_recalls", Deps: []string{"messages", "users"}},
		{Model: &model.MessageForward{}, Name: "message_forwards", Deps: []string{"messages", "users"}},
		{Model: &model.ScheduledMessage{}, Name: "scheduled_messages", Deps: []string{"messages", "users"}},
		{Model: &model.MessageSearchIndex{}, Name: "message_search_indices", Deps: []string{"messages"}},
		{Model: &model.MessagePin{}, Name: "message_pins", Deps: []string{"messages", "users"}},
		{Model: &model.MessageMark{}, Name: "message_marks", Deps: []string{"messages", "users"}},
		{Model: &model.MessageStatus{}, Name: "message_statuses", Deps: []string{"messages", "users"}},
		{Model: &model.MessageShare{}, Name: "message_shares", Deps: []string{"messages", "users", "chats"}},

		// =======================================
		// 文件管理
		// =======================================
		{Model: &model.File{}, Name: "files", Deps: []string{"users"}},
		{Model: &model.FileChunk{}, Name: "file_chunks", Deps: []string{"files"}},
		{Model: &model.FilePreview{}, Name: "file_previews", Deps: []string{"files"}},
		{Model: &model.FileAccess{}, Name: "file_accesses", Deps: []string{"files", "users"}},

		// =======================================
		// 内容审核
		// =======================================
		{Model: &model.ContentReport{}, Name: "content_reports", Deps: []string{"users"}},
		{Model: &model.ContentFilter{}, Name: "content_filters", Deps: []string{}},
		{Model: &model.UserWarning{}, Name: "user_warnings", Deps: []string{"users"}},
		{Model: &model.ModerationLog{}, Name: "moderation_logs", Deps: []string{"users"}},
		{Model: &model.ContentStatistics{}, Name: "content_statistics", Deps: []string{"users"}},

		// =======================================
		// 群组管理
		// =======================================
		{Model: &model.GroupInvite{}, Name: "group_invites", Deps: []string{"chats", "users"}},
		{Model: &model.GroupInviteUsage{}, Name: "group_invite_usages", Deps: []string{"group_invites", "users"}},
		{Model: &model.AdminRole{}, Name: "admin_roles", Deps: []string{}},
		{Model: &model.ChatAdmin{}, Name: "chat_admins", Deps: []string{"chats", "users", "admin_roles"}},
		{Model: &model.GroupJoinRequest{}, Name: "group_join_requests", Deps: []string{"chats", "users"}},
		{Model: &model.GroupAuditLog{}, Name: "group_audit_logs", Deps: []string{"chats", "users"}},
		{Model: &model.GroupPermissionTemplate{}, Name: "group_permission_templates", Deps: []string{"chats"}},

		// =======================================
		// 系统管理
		// =======================================
		{Model: &model.Alert{}, Name: "alerts", Deps: []string{}},
		{Model: &model.AdminOperationLog{}, Name: "admin_operation_logs", Deps: []string{"users"}},
		{Model: &model.SystemConfig{}, Name: "system_configs", Deps: []string{}},

		// =======================================
		// 安全认证
		// =======================================
		{Model: &model.IPBlacklist{}, Name: "ip_blacklists", Deps: []string{}},
		{Model: &model.UserBlacklist{}, Name: "user_blacklists", Deps: []string{"users"}},
		{Model: &model.LoginAttempt{}, Name: "login_attempts", Deps: []string{"users"}},
		{Model: &model.SuspiciousActivity{}, Name: "suspicious_activities", Deps: []string{"users"}},
		{Model: &model.TwoFactorAuth{}, Name: "two_factor_auths", Deps: []string{"users"}},
		{Model: &model.TrustedDevice{}, Name: "trusted_devices", Deps: []string{"users"}},
		{Model: &model.DeviceSession{}, Name: "device_sessions", Deps: []string{"users"}},
		{Model: &model.DeviceActivity{}, Name: "device_activities", Deps: []string{"users"}},

		// =======================================
		// 机器人系统
		// =======================================
		{Model: &model.Bot{}, Name: "bots", Deps: []string{"users"}},
		{Model: &model.BotAPILog{}, Name: "bot_api_logs", Deps: []string{"bots"}},
		{Model: &model.BotUser{}, Name: "bot_users", Deps: []string{"bots", "users"}},
		{Model: &model.BotUserPermission{}, Name: "bot_user_permissions", Deps: []string{"bot_users"}},

		// =======================================
		// 屏幕共享
		// =======================================
		{Model: &model.ScreenShareSession{}, Name: "screen_share_sessions", Deps: []string{"users"}},
		{Model: &model.ScreenShareQualityChange{}, Name: "screen_share_quality_changes", Deps: []string{"screen_share_sessions"}},
		{Model: &model.ScreenShareParticipant{}, Name: "screen_share_participants", Deps: []string{"screen_share_sessions", "users"}},
		{Model: &model.ScreenShareStatistics{}, Name: "screen_share_statistics", Deps: []string{"users"}},
		{Model: &model.ScreenShareRecording{}, Name: "screen_share_recordings", Deps: []string{"screen_share_sessions", "users"}},
	}
}

// MigrateTables 执行数据库表迁移
func MigrateTables(db *gorm.DB) error {
	log.Println("========================================")
	log.Println("🚀 开始数据库表迁移...")
	log.Println("========================================")

	migrations := GetMigrationOrder()

	// 打印迁移顺序
	log.Printf("📋 计划迁移 %d 个表：\n", len(migrations))
	for i, m := range migrations {
		depsStr := "无依赖"
		if len(m.Deps) > 0 {
			depsStr = fmt.Sprintf("依赖: %v", m.Deps)
		}
		log.Printf("  %d. %-35s (%s)\n", i+1, m.Name, depsStr)
	}
	log.Println("----------------------------------------")

	// 第一阶段：检查依赖表是否存在
	log.Println("🔍 第一阶段：检查依赖表...")
	for i, m := range migrations {
		if len(m.Deps) > 0 {
			for _, dep := range m.Deps {
				// 检查该依赖表是否在之前的迁移列表中
				found := false
				for j := 0; j < i; j++ {
					if migrations[j].Name == dep {
						found = true
						break
					}
				}
				if !found {
					log.Printf("❌ 错误：表 %s 依赖 %s，但 %s 不在之前的迁移列表中", m.Name, dep, dep)
					log.Println("========================================")
					log.Println("🚨 依赖检查失败！服务将不会启动。")
					log.Println("========================================")
					return fmt.Errorf("依赖检查失败：表 %s 依赖不存在或顺序错误的表 %s (Fail Fast)", m.Name, dep)
				}
			}
		}
	}
	log.Println("✅ 依赖检查通过")
	log.Println("----------------------------------------")

	// 第二阶段：执行迁移
	log.Println("⚙️  第二阶段：执行表迁移...")
	successCount := 0
	for i, m := range migrations {
		log.Printf("⏳ [%d/%d] 迁移表: %s", i+1, len(migrations), m.Name)

	// 检查表是否已存在
	tableExists := db.Migrator().HasTable(m.Model)
	if tableExists {
		log.Printf("   ℹ️  表 %s 已存在，跳过创建（避免AutoMigrate bug）", m.Name)
		log.Printf("   ✅ 迁移成功: %s（表已存在）", m.Name)
		successCount++
		continue
	}

	log.Printf("   ✨ 创建新表: %s", m.Name)

	// 使用CreateTable而不是AutoMigrate - 避免GORM的AutoMigrate bug
	// AutoMigrate会错误识别UNIQUE INDEX为FOREIGN KEY
	if err := db.Migrator().CreateTable(m.Model); err != nil {
		log.Printf("   ❌ 迁移失败: %v", err)
		log.Println("========================================")
		log.Println("🚨 数据库迁移失败！服务将不会启动。")
		log.Println("========================================")
		return fmt.Errorf("迁移表 %s 失败: %v (Fail Fast - 服务停止启动)", m.Name, err)
	}

		// 验证表确实创建成功
		if !db.Migrator().HasTable(m.Model) {
			log.Printf("   ❌ 验证失败：表 %s 迁移后仍不存在", m.Name)
			log.Println("========================================")
			log.Println("🚨 数据库迁移验证失败！服务将不会启动。")
			log.Println("========================================")
			return fmt.Errorf("表 %s 创建失败验证 (Fail Fast - 服务停止启动)", m.Name)
		}

		log.Printf("   ✅ 迁移成功: %s", m.Name)
		successCount++
	}

	log.Println("----------------------------------------")
	log.Printf("✅ 数据库迁移完成！成功迁移 %d/%d 个表\n", successCount, len(migrations))

	// 第三阶段：迁移后完整性验证
	log.Println("🔍 第三阶段：验证表完整性...")
	if err := VerifyTables(db); err != nil {
		log.Println("========================================")
		log.Println("🚨 数据库验证失败！服务将不会启动。")
		log.Println("========================================")
		return fmt.Errorf("表完整性验证失败 (Fail Fast - 服务停止启动): %v", err)
	}

	log.Println("========================================")
	log.Println("🎉 数据库迁移和验证全部通过！服务可以安全启动。")
	log.Println("========================================")
	return nil
}

// VerifyTables 验证所有关键表是否存在
func VerifyTables(db *gorm.DB) error {
	log.Println("🔍 开始验证表结构...")

	// 关键表列表
	criticalTables := []string{
		"users", "sessions", "contacts",
		"chats", "chat_members",
		"message_replies", "messages",
		"files", "bots",
	}

	missingTables := []string{}
	for _, table := range criticalTables {
		if !db.Migrator().HasTable(table) {
			missingTables = append(missingTables, table)
		}
	}

	if len(missingTables) > 0 {
		log.Printf("❌ 缺失关键表: %v\n", missingTables)
		return fmt.Errorf("数据库迁移不完整，缺失关键表: %v", missingTables)
	}

	// 打印所有表
	var tables []string
	db.Raw("SHOW TABLES").Scan(&tables)
	log.Printf("✅ 数据库验证通过！当前共有 %d 个表\n", len(tables))

	if len(tables) > 0 {
		log.Println("📊 数据库表列表：")
		for i, table := range tables {
			if (i+1)%3 == 0 {
				log.Printf("  %s\n", table)
			} else {
				log.Printf("  %-30s", table)
			}
		}
		log.Println()
	}

	return nil
}

// CheckTableExists 检查表是否存在
func CheckTableExists(db *gorm.DB, tableName string) bool {
	return db.Migrator().HasTable(tableName)
}

// GetTableList 获取数据库中所有表的列表
func GetTableList(db *gorm.DB) ([]string, error) {
	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		return nil, fmt.Errorf("获取表列表失败: %v", err)
	}
	return tables, nil
}
