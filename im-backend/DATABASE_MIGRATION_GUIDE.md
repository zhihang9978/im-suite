# 数据库迁移安全机制

## 📋 概述

本项目采用优化的数据库迁移机制，确保在部署过程中不会因为表依赖顺序、外键约束、索引长度等问题导致失败。

## 🎯 核心特性

### 1. 智能依赖排序
- ✅ 自动分析表之间的外键依赖关系
- ✅ 按照依赖顺序创建表，被引用的表优先创建
- ✅ 支持复杂的多层依赖关系

### 2. 表存在检测
- ✅ 迁移前检查表是否已存在
- ✅ 已存在的表只执行结构更新
- ✅ 避免重复创建和循环依赖

### 3. 迁移后验证
- ✅ 自动验证所有关键表是否成功创建
- ✅ 打印所有已创建表的列表
- ✅ 缺失关键表时立即报错退出

### 4. 字段长度规范
- ✅ 所有 `uniqueIndex` 字段明确声明 `varchar` 长度
- ✅ Token、Phone、Username 等字段长度在 20~255 之间
- ✅ 避免 "Specified key was too long; max key length is 3072 bytes" 错误

### 5. 详细日志输出
- ✅ 打印完整的迁移顺序和依赖关系
- ✅ 每个表的迁移状态实时显示
- ✅ 成功/失败信息清晰标识

## 📊 表依赖关系

### 第一层：基础表（无外键依赖）
```
users, chats, themes
```

### 第二层：依赖基础表
```
sessions → users
contacts → users
chat_members → chats, users
user_theme_settings → users, themes
```

### 第三层：消息回复链
```
message_replies（被 messages 引用）
```

### 第四层：消息主表
```
messages → users, chats, message_replies
```

### 第五层：消息相关表
```
message_reads → messages, users
message_edits → messages
message_recalls → messages, users
message_forwards → messages, users
... 其他消息相关表
```

### 其他层：文件、群组、安全、机器人、屏幕共享等
```
files → users
file_chunks → files
bots → users
bot_api_logs → bots
screen_share_sessions → users
... 更多表
```

## 🔧 使用方法

### 自动迁移（推荐）

后端启动时会自动执行迁移：

```go
// im-backend/main.go
if err := config.AutoMigrate(); err != nil {
    log.Fatalf("数据库迁移失败: %v", err)
}
```

### 手动迁移

```go
import "zhihang-messenger/im-backend/config"

// 执行迁移
err := config.MigrateTables(db)
if err != nil {
    log.Fatalf("迁移失败: %v", err)
}
```

### 检查表是否存在

```go
exists := config.CheckTableExists(db, "users")
if exists {
    log.Println("users 表已存在")
}
```

### 获取所有表列表

```go
tables, err := config.GetTableList(db)
if err != nil {
    log.Fatalf("获取表列表失败: %v", err)
}
for _, table := range tables {
    fmt.Println(table)
}
```

## 🧪 单元测试

### 运行所有测试

```bash
cd im-backend/config
go test -v
```

### 运行特定测试

```bash
# 测试迁移顺序
go test -v -run TestMigrationOrder

# 测试表依赖关系
go test -v -run TestTableDependencies

# 测试表验证功能
go test -v -run TestVerifyTables
```

### 基准测试

```bash
go test -bench=BenchmarkMigration -benchmem
```

## 📝 字段长度规范

### uniqueIndex 字段必须明确长度

❌ **错误示例**：
```go
Phone string `json:"phone" gorm:"uniqueIndex;not null"`
Token string `json:"token" gorm:"uniqueIndex;not null"`
```

✅ **正确示例**：
```go
Phone string `json:"phone" gorm:"type:varchar(20);uniqueIndex;not null"`
Token string `json:"token" gorm:"type:varchar(255);uniqueIndex;not null"`
```

### 推荐字段长度

| 字段类型 | 推荐长度 | 说明 |
|---------|---------|------|
| Phone | varchar(20) | 手机号 |
| Username | varchar(50) | 用户名 |
| Token | varchar(255) | 令牌/密钥 |
| Email | varchar(100) | 邮箱地址 |
| Nickname | varchar(100) | 昵称 |
| URL | varchar(255) | URL地址 |
| IP | varchar(45) | IP地址（支持IPv6） |
| Bio/Description | varchar(500) | 简介/描述 |
| LongText | text | 长文本内容 |

## 🚨 常见问题

### 问题1：Error 1824 - Failed to open the referenced table

**原因**：被引用的表未创建就尝试创建引用表

**解决**：使用新的迁移机制，已自动处理依赖顺序

### 问题2：Specified key was too long; max key length is 3072 bytes

**原因**：uniqueIndex 字段长度过长或未指定长度

**解决**：明确指定 varchar 长度，建议不超过 255

### 问题3：表迁移失败但无明确错误

**原因**：可能是字符集、排序规则等数据库配置问题

**解决**：检查数据库配置，确保使用 utf8mb4 字符集

## 🔍 迁移日志示例

```
========================================
🚀 开始数据库表迁移...
========================================
📋 计划迁移 60 个表：
  1. users                            (无依赖)
  2. chats                            (无依赖)
  3. themes                           (无依赖)
  4. sessions                         (依赖: [users])
  5. contacts                         (依赖: [users])
  6. chat_members                     (依赖: [chats users])
  7. message_replies                  (无依赖)
  8. messages                         (依赖: [users chats message_replies])
  ... 更多表
----------------------------------------
⏳ [1/60] 迁移表: users
   ✨ 创建新表: users
   ✅ 迁移成功: users
⏳ [2/60] 迁移表: chats
   ✨ 创建新表: chats
   ✅ 迁移成功: chats
... 继续迁移
----------------------------------------
✅ 数据库迁移完成！成功迁移 60/60 个表
🔍 开始验证表结构...
✅ 数据库验证通过！当前共有 60 个表
📊 数据库表列表：
  users                         sessions                      contacts
  chats                         chat_members                  messages
  ... 更多表
========================================
```

## 📦 新增表时的注意事项

### 1. 添加新模型

```go
// im-backend/internal/model/your_model.go
type YourModel struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    Name      string    `gorm:"type:varchar(100);not null" json:"name"`
    CreatedAt time.Time `json:"created_at"`
    
    // 关联
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

### 2. 添加到迁移列表

在 `config/database_migration.go` 的 `GetMigrationOrder()` 函数中添加：

```go
// 根据依赖关系选择合适的位置
{Model: &model.YourModel{}, Name: "your_models", Deps: []string{"users"}},
```

### 3. 运行测试

```bash
cd im-backend/config
go test -v
```

### 4. 验证迁移

```bash
cd im-backend
go run main.go
# 查看日志确认表创建成功
```

## 🛠️ 维护指南

### 定期检查

1. **检查迁移顺序**：确保新表的依赖关系正确
2. **运行单元测试**：每次修改后运行测试
3. **检查字段长度**：新增字段时明确指定长度
4. **验证外键**：确保外键引用的表已在依赖列表中

### 性能优化

1. **索引优化**：合理添加索引，避免过多索引
2. **批量迁移**：使用事务批量创建表
3. **缓存机制**：缓存表存在检查结果

## 📚 相关文档

- [GORM 官方文档](https://gorm.io/docs/)
- [MySQL 外键约束](https://dev.mysql.com/doc/refman/8.0/en/create-table-foreign-keys.html)
- [数据库索引优化](https://dev.mysql.com/doc/refman/8.0/en/optimization-indexes.html)

## 💡 最佳实践

1. ✅ 总是明确指定 varchar 长度
2. ✅ 被引用的表必须先创建
3. ✅ 使用 uniqueIndex 时确保字段长度合理
4. ✅ 每次修改模型后运行单元测试
5. ✅ 部署前在测试环境验证迁移
6. ✅ 生产环境先备份数据库
7. ✅ 使用版本控制管理数据库变更

## 🔐 安全建议

1. **备份优先**：生产环境迁移前必须备份
2. **测试先行**：在测试环境充分测试
3. **日志监控**：关注迁移日志，及时发现问题
4. **回滚方案**：准备数据库回滚脚本
5. **分步迁移**：大规模变更分多次进行

---

**维护者**: 志航密信开发团队  
**最后更新**: 2025-10-09  
**版本**: v1.6.0

