# 数据库字段长度规范清单

## 📋 概述

本文档列出所有数据库字段的长度规范，特别是带 `uniqueIndex` 的字段，确保索引长度不超过 MySQL 限制（3072 bytes）。

## ✅ 唯一索引字段清单

### 用户相关 (user.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| Phone | varchar | 20 | uniqueIndex, not null | 手机号（国际格式） |
| Username | varchar | 50 | uniqueIndex | 用户名 |

### 会话相关 (user.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| Token | varchar | 255 | uniqueIndex, not null | 会话令牌 |

### 机器人相关 (bot.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| Name | varchar | 100 | uniqueIndex, not null | 机器人名称 |
| APIKey | varchar | 255 | uniqueIndex, not null | API密钥 |
| UserID | uint | - | uniqueIndex, not null | 系统用户ID（整型无需长度） |

### 文件相关 (file.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| FileHash | varchar | 64 | uniqueIndex, not null | 文件哈希值（SHA256） |

### 系统配置 (system.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| Key | varchar | 100 | uniqueIndex, not null | 配置键 |
| IPAddress | varchar | 45 | uniqueIndex, not null | IP地址（支持IPv6） |

### 群组管理 (group_management.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| InviteCode | varchar | 50 | uniqueIndex | 邀请码 |

### 主题设置 (theme.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| UserID | uint | - | uniqueIndex, not null | 用户ID（整型无需长度） |

### 屏幕共享 (screen_share.go)

| 字段 | 类型 | 长度 | 约束 | 说明 |
|------|------|------|------|------|
| UserID | uint | - | uniqueIndex, not null | 用户ID（整型无需长度） |

## 📏 通用字段长度标准

### 字符串字段长度规范

| 字段用途 | 推荐长度 | 示例字段 |
|---------|---------|---------|
| 手机号 | varchar(20) | Phone |
| 用户名 | varchar(50) | Username |
| 短标识符 | varchar(50) | InviteCode, Type |
| 昵称/名称 | varchar(100) | Nickname, Name |
| 邮箱 | varchar(100) | Email |
| MIME类型 | varchar(100) | MimeType |
| 设备信息 | varchar(100) | Device |
| 文件名 | varchar(255) | FileName |
| URL/路径 | varchar(500) | StoragePath, StorageURL |
| 令牌/密钥 | varchar(255) | Token, APIKey, AccessToken |
| IP地址 | varchar(45) | IP, IPAddress（支持IPv6） |
| 哈希值(MD5) | varchar(32) | - |
| 哈希值(SHA1) | varchar(40) | - |
| 哈希值(SHA256) | varchar(64) | FileHash |
| 描述/简介 | varchar(500) | Bio, Description |
| 用户代理 | varchar(500) | UserAgent |
| 长文本 | text | Content, Permissions |

## 🔍 索引长度计算

### MySQL 索引长度限制

- **InnoDB 引擎**: 索引最大长度 3072 bytes
- **utf8mb4 字符集**: 每个字符占 4 bytes
- **安全上限**: varchar(768) 在 utf8mb4 下

### 示例计算

```
varchar(20)  * 4 bytes = 80 bytes   ✅ 安全
varchar(50)  * 4 bytes = 200 bytes  ✅ 安全
varchar(100) * 4 bytes = 400 bytes  ✅ 安全
varchar(255) * 4 bytes = 1020 bytes ✅ 安全
varchar(500) * 4 bytes = 2000 bytes ✅ 安全
varchar(768) * 4 bytes = 3072 bytes ⚠️ 极限
varchar(1000)* 4 bytes = 4000 bytes ❌ 超限
```

## ✅ 验证规则

### 1. 唯一索引字段必须明确长度

❌ **错误**:
```go
Phone string `gorm:"uniqueIndex;not null"`
```

✅ **正确**:
```go
Phone string `gorm:"type:varchar(20);uniqueIndex;not null"`
```

### 2. 整型字段不需要 varchar

✅ **正确**:
```go
UserID uint `gorm:"uniqueIndex;not null"`  // uint/int 不需要声明 varchar
```

### 3. 路径和URL字段

✅ **正确**:
```go
StoragePath string `gorm:"type:varchar(500);not null"`
StorageURL  string `gorm:"type:varchar(500);not null"`
```

### 4. 文本字段使用 text 类型

✅ **正确**:
```go
Content     string `gorm:"type:text"`  // 长文本内容
Permissions string `gorm:"type:text"`  // JSON 数据
```

## 🧪 自动化检查

### 检查命令

```bash
# 查找所有 uniqueIndex 字段
cd im-backend/internal/model
grep -r "uniqueIndex" *.go

# 查找未声明长度的 string 字段
grep -r "string.*uniqueIndex" *.go | grep -v "varchar"
```

### 预期结果

所有带 `uniqueIndex` 的 `string` 类型字段都应该有 `type:varchar(n)` 声明。

## 📊 字段长度统计

### 已规范化字段统计

- **唯一索引字段**: 9 个（7个varchar + 2个uint）
- **varchar 字段总数**: 50+ 个
- **text 字段**: 15+ 个
- **最大 varchar 长度**: varchar(500)
- **平均 varchar 长度**: ~200

### 字段分布

```
varchar(20):   1 个  (Phone)
varchar(45):   1 个  (IPAddress)
varchar(50):   5 个  (Username, InviteCode, Type...)
varchar(64):   1 个  (FileHash)
varchar(100):  6 个  (Name, MimeType, Device...)
varchar(255): 10 个  (Token, APIKey, FileName, URL...)
varchar(500):  8 个  (StoragePath, Description, UserAgent...)
text:         15+ 个 (Content, Permissions...)
```

## 🔧 维护指南

### 添加新字段时

1. **确定字段用途**
   - 查看上面的"通用字段长度标准"表
   - 选择合适的长度

2. **添加字段定义**
   ```go
   FieldName string `gorm:"type:varchar(length);其他约束" json:"field_name"`
   ```

3. **如果是唯一索引**
   - 必须声明 varchar 长度
   - 长度建议不超过 255
   - 整型字段除外

4. **运行测试**
   ```bash
   cd im-backend/config
   go test -v -run TestMigrationOrder
   ```

5. **验证迁移**
   ```bash
   cd im-backend
   go run main.go
   # 查看迁移日志
   ```

### 修改现有字段时

1. **评估影响**
   - 是否有数据需要迁移
   - 索引是否需要重建

2. **测试环境验证**
   - 先在测试环境执行
   - 确认无错误后再生产部署

3. **生产环境部署**
   - 备份数据库
   - 执行迁移
   - 验证数据完整性

## 📋 检查清单

### 代码审查时检查

- [ ] 所有 string 字段都有明确的 type 声明
- [ ] uniqueIndex 字段的 varchar 长度 <= 255
- [ ] 路径和 URL 字段使用 varchar(500)
- [ ] 长文本使用 text 类型
- [ ] 整型字段不使用 varchar
- [ ] 字段名称和长度符合规范

### 部署前检查

- [ ] 运行单元测试通过
- [ ] 本地迁移测试通过
- [ ] 测试环境部署成功
- [ ] 迁移日志无错误
- [ ] 所有关键表已创建

## 🚨 常见错误

### 错误1：未声明 varchar 长度

```go
// ❌ 错误
Name string `gorm:"uniqueIndex;not null"`

// ✅ 正确
Name string `gorm:"type:varchar(100);uniqueIndex;not null"`
```

### 错误2：varchar 长度过长

```go
// ❌ 错误（索引可能失败）
LongField string `gorm:"type:varchar(1000);uniqueIndex"`

// ✅ 正确（使用合理长度或改用非唯一索引）
LongField string `gorm:"type:varchar(255);index"`
```

### 错误3：整型使用 varchar

```go
// ❌ 错误
UserID uint `gorm:"type:varchar(20);uniqueIndex"`

// ✅ 正确
UserID uint `gorm:"uniqueIndex"`
```

## 📚 参考资料

- [MySQL InnoDB Limits](https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html)
- [GORM Data Types](https://gorm.io/docs/models.html#Fields-Tags)
- [utf8mb4 Character Set](https://dev.mysql.com/doc/refman/8.0/en/charset-unicode-utf8mb4.html)

---

**维护者**: 志航密信开发团队  
**最后更新**: 2025-10-09  
**版本**: v1.6.0  
**状态**: ✅ 已规范化完成

