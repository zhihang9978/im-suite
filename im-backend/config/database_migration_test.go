package config

import (
	"testing"
	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestMigrationOrder 测试迁移顺序正确性
func TestMigrationOrder(t *testing.T) {
	// 使用内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("跳过测试: SQLite需要CGO支持 (CGO_ENABLED=0): %v", err)
		return
	}

	t.Log("测试数据库迁移顺序...")

	// 执行迁移
	err = MigrateTables(db)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	t.Log("✅ 数据库迁移成功")

	// 验证所有关键表是否存在
	criticalModels := []interface{}{
		&model.User{},
		&model.Session{},
		&model.Contact{},
		&model.Chat{},
		&model.ChatMember{},
		&model.MessageReply{},
		&model.Message{},
		&model.File{},
		&model.Bot{},
		&model.ScreenShareSession{},
	}

	for _, m := range criticalModels {
		if !db.Migrator().HasTable(m) {
			t.Errorf("❌ 关键表未创建: %T", m)
		} else {
			t.Logf("✅ 表已创建: %T", m)
		}
	}
}

// TestTableDependencies 测试表依赖关系
func TestTableDependencies(t *testing.T) {
	t.Log("测试表依赖关系...")

	migrations := GetMigrationOrder()

	// 检查 MessageReply 是否在 Message 之前
	messageReplyIndex := -1
	messageIndex := -1

	for i, m := range migrations {
		if m.Name == "message_replies" {
			messageReplyIndex = i
		}
		if m.Name == "messages" {
			messageIndex = i
		}
	}

	if messageReplyIndex == -1 {
		t.Error("❌ 未找到 message_replies 表")
	}
	if messageIndex == -1 {
		t.Error("❌ 未找到 messages 表")
	}

	if messageReplyIndex >= messageIndex {
		t.Errorf("❌ 依赖顺序错误: message_replies (索引:%d) 应该在 messages (索引:%d) 之前", messageReplyIndex, messageIndex)
	} else {
		t.Logf("✅ 依赖顺序正确: message_replies (索引:%d) 在 messages (索引:%d) 之前", messageReplyIndex, messageIndex)
	}

	// 检查 User 是否在所有依赖它的表之前
	userIndex := -1
	for i, m := range migrations {
		if m.Name == "users" {
			userIndex = i
			break
		}
	}

	if userIndex == -1 {
		t.Error("❌ 未找到 users 表")
	}

	userDependentTables := []string{"sessions", "contacts", "messages", "bots"}
	for _, depTable := range userDependentTables {
		for i, m := range migrations {
			if m.Name == depTable {
				if userIndex >= i {
					t.Errorf("❌ 依赖顺序错误: users (索引:%d) 应该在 %s (索引:%d) 之前", userIndex, depTable, i)
				} else {
					t.Logf("✅ 依赖顺序正确: users (索引:%d) 在 %s (索引:%d) 之前", userIndex, depTable, i)
				}
				break
			}
		}
	}
}

// TestVerifyTables 测试表验证功能
func TestVerifyTables(t *testing.T) {
	t.Log("测试表验证功能...")

	// 使用内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 执行迁移
	err = MigrateTables(db)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	// 验证表
	err = VerifyTables(db)
	if err != nil {
		t.Errorf("❌ 表验证失败: %v", err)
	} else {
		t.Log("✅ 表验证成功")
	}
}

// TestCheckTableExists 测试表存在检查
func TestCheckTableExists(t *testing.T) {
	t.Log("测试表存在检查...")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 创建一个表
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("迁移 User 表失败: %v", err)
	}

	// 检查表是否存在
	if !CheckTableExists(db, "users") {
		t.Error("❌ users 表应该存在")
	} else {
		t.Log("✅ users 表存在检查通过")
	}

	// 检查不存在的表
	if CheckTableExists(db, "non_existent_table") {
		t.Error("❌ non_existent_table 不应该存在")
	} else {
		t.Log("✅ 不存在的表检查通过")
	}
}

// TestMigrationCount 测试迁移表数量
func TestMigrationCount(t *testing.T) {
	t.Log("测试迁移表数量...")

	migrations := GetMigrationOrder()

	// 预期的最小表数量（根据实际模型数量调整）
	expectedMinTables := 50

	if len(migrations) < expectedMinTables {
		t.Errorf("❌ 迁移表数量不足: 期望至少 %d 个，实际 %d 个", expectedMinTables, len(migrations))
	} else {
		t.Logf("✅ 迁移表数量正常: %d 个表", len(migrations))
	}

	// 检查是否有重复的表名
	tableNames := make(map[string]bool)
	for _, m := range migrations {
		if tableNames[m.Name] {
			t.Errorf("❌ 发现重复的表名: %s", m.Name)
		}
		tableNames[m.Name] = true
	}

	if len(tableNames) == len(migrations) {
		t.Log("✅ 无重复表名")
	}
}

// BenchmarkMigration 基准测试迁移性能
func BenchmarkMigration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			b.Fatalf("创建测试数据库失败: %v", err)
		}

		err = MigrateTables(db)
		if err != nil {
			b.Fatalf("数据库迁移失败: %v", err)
		}
	}
}
