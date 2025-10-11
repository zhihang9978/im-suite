package config

import (
	"testing"

	"zhihang-messenger/im-backend/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestForeignKeyDependencies 测试所有表的外键依赖顺序
func TestForeignKeyDependencies(t *testing.T) {
	// 使用内存数据库测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 获取迁移顺序
	migrations := GetMigrationOrder()

	// 验证每个表的依赖都在它之前创建
	createdTables := make(map[string]bool)

	for i, m := range migrations {
		t.Logf("检查表 [%d/%d]: %s", i+1, len(migrations), m.Name)

		// 检查依赖
		for _, dep := range m.Deps {
			if !createdTables[dep] {
				t.Errorf("❌ 表 '%s' 依赖 '%s'，但 '%s' 还未创建（顺序错误）", m.Name, dep, dep)
			}
		}

		// 尝试创建表
		if err := db.Migrator().CreateTable(m.Model); err != nil {
			t.Errorf("❌ 创建表 '%s' 失败: %v", m.Name, err)
		} else {
			t.Logf("   ✅ 创建成功: %s", m.Name)
			createdTables[m.Name] = true
		}
	}

	// 验证所有表都创建成功
	if len(createdTables) != len(migrations) {
		t.Errorf("❌ 预期创建 %d 个表，实际创建 %d 个", len(migrations), len(createdTables))
	} else {
		t.Logf("✅ 所有 %d 个表创建成功", len(createdTables))
	}
}

// TestMessageReplyDependency 专门测试MessageReply的依赖
func TestMessageReplyDependency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 先创建依赖表
	tables := []interface{}{
		&model.User{},
		&model.Chat{},
		&model.Message{},
	}

	for _, table := range tables {
		if err := db.Migrator().CreateTable(table); err != nil {
			t.Fatalf("创建依赖表失败: %v", err)
		}
	}

	// 创建MessageReply（应该成功，因为依赖的Message表已存在）
	if err := db.Migrator().CreateTable(&model.MessageReply{}); err != nil {
		t.Errorf("❌ 创建MessageReply表失败: %v", err)
	} else {
		t.Log("✅ MessageReply表创建成功")
	}

	// 验证外键约束
	if db.Migrator().HasConstraint(&model.MessageReply{}, "fk_message_replies_message") ||
		db.Migrator().HasConstraint(&model.MessageReply{}, "Message") {
		t.Log("✅ MessageReply的外键约束已创建")
	}
}

// TestCircularDependencyDetection 测试循环依赖检测
func TestCircularDependencyDetection(t *testing.T) {
	migrations := GetMigrationOrder()

	// 构建依赖图
	depGraph := make(map[string][]string)
	for _, m := range migrations {
		depGraph[m.Name] = m.Deps
	}

	// 检测循环依赖
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range depGraph[node] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				t.Errorf("❌ 检测到循环依赖: %s -> %s", node, dep)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for tableName := range depGraph {
		if !visited[tableName] {
			if hasCycle(tableName) {
				t.Fatal("❌ 存在循环依赖，测试失败")
			}
		}
	}

	t.Log("✅ 未检测到循环依赖")
}

// TestAllForeignKeysHaveDependencies 测试所有外键都在依赖列表中
func TestAllForeignKeysHaveDependencies(t *testing.T) {
	// 定义已知的外键关系
	foreignKeys := map[string][]string{
		"sessions":           {"users"},
		"contacts":           {"users"},
		"chats":              {},
		"chat_members":       {"chats", "users"},
		"messages":           {"users", "chats"},
		"message_replies":    {"messages"},
		"message_reads":      {"messages", "users"},
		"message_edits":      {"messages"},
		"message_recalls":    {"messages", "users"},
		"message_forwards":   {"messages", "users"},
		"scheduled_messages": {"messages", "users"},
		"message_reactions":  {"messages", "users"},
		"message_shares":     {"messages", "users", "chats"},
		"files":              {"users"},
		"bots":               {"users"},
		"bot_users":          {"bots"},
		"bot_permissions":    {"bots"},
		// ... 其他表
	}

	migrations := GetMigrationOrder()

	for _, m := range migrations {
		expectedDeps, exists := foreignKeys[m.Name]
		if !exists {
			// 跳过未定义的表
			continue
		}

		// 检查所有预期的依赖是否都在Deps中
		for _, expectedDep := range expectedDeps {
			found := false
			for _, actualDep := range m.Deps {
				if actualDep == expectedDep {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("❌ 表 '%s' 缺少依赖 '%s'（外键关系未声明）", m.Name, expectedDep)
			}
		}
	}

	t.Log("✅ 外键依赖检查完成")
}

// BenchmarkMigrationOrder 性能测试：迁移顺序计算
func BenchmarkMigrationOrder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetMigrationOrder()
	}
}
