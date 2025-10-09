# 超级管理员功能清单

**角色**: super_admin（超级管理员）  
**权限级别**: 最高  
**版本**: v1.4.0

---

## 🎯 超级管理员拥有的所有功能

### 📊 1. 系统统计和监控

#### 系统统计仪表板
**端点**: `GET /api/super-admin/stats`

**功能**：
- ✅ 总用户数
- ✅ 在线用户数
- ✅ 总聊天数
- ✅ 总消息数
- ✅ 今日消息数
- ✅ 今日新用户数
- ✅ 今日活跃用户数
- ✅ 总存储空间
- ✅ 服务器运行时间

#### 系统性能指标
**端点**: `GET /api/super-admin/stats/system`

**功能**：
- ✅ CPU使用率
- ✅ 内存使用率
- ✅ 磁盘使用率
- ✅ 网络流量统计
- ✅ 数据库连接数
- ✅ Redis内存使用
- ✅ 缓存命中率
- ✅ API响应时间

---

### 👥 2. 用户全面管理

#### 在线用户监控
**端点**: `GET /api/super-admin/users/online`

**功能**：
- ✅ 查看所有在线用户
- ✅ 用户详细信息（ID、用户名、昵称、头像）
- ✅ 在线状态
- ✅ IP地址
- ✅ 设备信息
- ✅ 登录时间
- ✅ 最后活动时间
- ✅ 会话数量

#### 用户行为分析
**端点**: `GET /api/super-admin/users/:id/analysis`

**功能**：
- ✅ 用户基本信息
- ✅ 消息发送总数
- ✅ 参与聊天数
- ✅ 最后活跃时间
- ✅ 风险评分
- ✅ 违规次数
- ✅ 被举报次数
- ✅ 封禁状态

#### 强制用户下线
**端点**: `POST /api/super-admin/users/:id/logout`

**功能**：
- ✅ 强制指定用户下线
- ✅ 清除用户会话
- ✅ 更新在线状态
- ✅ 记录管理员操作日志
- ✅ 从Redis删除会话

#### 封禁用户
**端点**: `POST /api/super-admin/users/:id/ban`

**参数**：
```json
{
  "duration": 86400,  // 封禁时长（秒）
  "reason": "违规原因"
}
```

**功能**：
- ✅ 临时或永久封禁用户
- ✅ 自定义封禁时长
- ✅ 记录封禁原因
- ✅ 自动计算解封时间
- ✅ 记录管理员操作

#### 解封用户
**端点**: `POST /api/super-admin/users/:id/unban`

**功能**：
- ✅ 解除用户封禁
- ✅ 清除封禁记录
- ✅ 恢复用户权限
- ✅ 记录管理员操作

#### 删除用户
**端点**: `DELETE /api/super-admin/users/:id`

**功能**：
- ✅ 永久删除用户账号
- ✅ 删除用户所有数据
- ✅ 删除用户消息（可选）
- ✅ 删除用户文件（可选）
- ✅ 记录管理员操作
- ⚠️ 不可恢复操作

---

### 📢 3. 系统管理功能

#### 系统告警管理
**端点**: `GET /api/super-admin/alerts`

**功能**：
- ✅ 查看所有活跃告警
- ✅ CPU告警
- ✅ 内存告警
- ✅ 磁盘告警
- ✅ 数据库告警
- ✅ Redis告警
- ✅ 告警级别分类
- ✅ 告警时间记录

#### 管理员操作日志
**端点**: `GET /api/super-admin/logs`

**参数**: `?limit=100`

**功能**：
- ✅ 查看管理员操作历史
- ✅ 操作类型（下线、封禁、删除等）
- ✅ 操作对象
- ✅ 操作时间
- ✅ 操作详情
- ✅ 管理员信息
- ✅ 完整审计追踪

#### 系统广播消息
**端点**: `POST /api/super-admin/broadcast`

**参数**：
```json
{
  "message": "系统维护通知"
}
```

**功能**：
- ✅ 向所有用户发送系统消息
- ✅ 向指定群组广播
- ✅ 紧急通知推送
- ✅ 系统公告发布
- ✅ 记录广播日志

---

### 🛡️ 4. 内容审核管理（Admin权限也可访问）

#### 待审核内容队列
**端点**: `GET /api/moderation/reports/pending`

**功能**：
- ✅ 查看所有待审核举报
- ✅ 举报优先级排序
- ✅ 举报类型分类
- ✅ 举报时间
- ✅ 举报用户信息

#### 查看举报详情
**端点**: `GET /api/moderation/reports/:id`

**功能**：
- ✅ 举报完整详情
- ✅ 被举报内容
- ✅ 举报原因
- ✅ 举报用户
- ✅ 被举报用户
- ✅ 相关证据

#### 处理举报
**端点**: `POST /api/moderation/reports/:id/handle`

**功能**：
- ✅ 批准/拒绝举报
- ✅ 删除违规内容
- ✅ 警告用户
- ✅ 封禁用户（如严重）
- ✅ 记录处理结果
- ✅ 通知举报者

#### 创建过滤规则
**端点**: `POST /api/moderation/filters`

**功能**：
- ✅ 添加敏感词过滤
- ✅ 正则表达式过滤
- ✅ URL黑名单
- ✅ 自动检测规则
- ✅ 过滤级别设置

#### 用户警告历史
**端点**: `GET /api/moderation/users/:id/warnings`

**功能**：
- ✅ 查看用户所有警告
- ✅ 警告时间
- ✅ 警告原因
- ✅ 警告级别
- ✅ 处理管理员

#### 审核统计
**端点**: `GET /api/moderation/statistics`

**功能**：
- ✅ 总举报数
- ✅ 待处理举报数
- ✅ 已处理举报数
- ✅ 违规用户统计
- ✅ 处理效率分析
- ✅ 内容类型分布

#### 内容检测
**端点**: `POST /api/moderation/content/check`

**功能**：
- ✅ 手动检测内容
- ✅ 敏感词检测
- ✅ URL检测
- ✅ 风险评分
- ✅ 检测报告

---

### 👤 5. 用户管理（User Management）

**注意**: 这些功能通过 `/api/users/` 路径，需要管理员权限

#### 黑名单管理
- ✅ 添加用户到黑名单: `POST /api/users/:id/blacklist`
- ✅ 移除黑名单: `DELETE /api/users/:id/blacklist/:blacklist_id`
- ✅ 查看黑名单列表: `GET /api/users/:id/blacklist`

#### 用户活动追踪
- ✅ 查看用户活动: `GET /api/users/:id/activity`
- ✅ 用户统计信息: `GET /api/users/:id/stats`

#### 用户限制管理
- ✅ 设置用户限制: `POST /api/users/:id/restrictions`
- ✅ 查看限制: `GET /api/users/:id/restrictions`
- ✅ 检查限制: `GET /api/users/:id/restrictions/check`

#### 用户封禁（User Management版本）
- ✅ 封禁用户: `POST /api/users/:id/ban`
- ✅ 解封用户: `POST /api/users/:id/unban`

#### 可疑用户检测
- ✅ 查看可疑用户: `GET /api/users/suspicious`
- ✅ 自动风险评分
- ✅ 异常行为检测

---

### 💬 6. 群组和聊天管理

#### 群组管理
- ✅ 查看所有群组
- ✅ 群组权限管理
- ✅ 群组公告管理
- ✅ 群组统计分析
- ✅ 群组备份恢复

#### 群组审核
- ✅ 查看入群申请
- ✅ 审批/拒绝申请
- ✅ 查看审计日志

---

### 📁 7. 文件管理

**通过普通用户API，但可以管理所有文件**：
- ✅ 查看所有上传文件
- ✅ 删除违规文件
- ✅ 文件统计分析
- ✅ 存储空间管理

---

### 🎨 8. 主题和系统配置

#### 主题管理
- ✅ 创建系统主题
- ✅ 初始化内置主题
- ✅ 管理主题模板

---

### 🔐 9. 安全功能（v1.4.0新增）

#### 所有用户的2FA管理（未来可扩展）
- 📋 查看用户2FA状态
- 📋 强制重置用户2FA
- 📋 解锁2FA失败锁定

#### 所有用户的设备管理（未来可扩展）
- 📋 查看所有用户设备
- 📋 强制撤销可疑设备
- 📋 设备风险分析

---

## 📊 功能分类统计

### 超级管理员专属功能（仅super_admin）

| 分类 | 功能数 | 说明 |
|------|--------|------|
| 系统统计 | 2个 | 实时系统数据 |
| 用户管理 | 4个 | 强制下线、封禁、删除、分析 |
| 系统管理 | 3个 | 告警、日志、广播 |
| **小计** | **9个** | 核心管理功能 |

### 管理员级别功能（admin和super_admin都可访问）

| 分类 | 功能数 | 说明 |
|------|--------|------|
| 内容审核 | 7个 | 举报处理、过滤规则 |
| 用户限制 | 9个 | 黑名单、限制管理 |
| 群组管理 | 15个 | 权限、公告、统计 |
| **小计** | **31个** | 日常管理功能 |

### 普通功能（所有角色都可访问）

| 分类 | 功能数 | 说明 |
|------|--------|------|
| 消息管理 | 12个 | 发送、接收、撤回等 |
| 个人2FA | 9个 | 自己的2FA设置 |
| 个人设备 | 9个 | 自己的设备管理 |
| 文件管理 | 8个 | 上传、下载、预览 |
| 其他 | 50个 | 主题、加密等 |
| **小计** | **88个** | 基础通讯功能 |

**总API端点**: 128个

---

## 🔐 权限对照表

| 功能 | user | admin | super_admin |
|------|------|-------|-------------|
| **系统统计** | ❌ | ❌ | ✅ |
| **系统监控** | ❌ | ❌ | ✅ |
| **强制下线** | ❌ | ❌ | ✅ |
| **封禁用户** | ❌ | ❌ | ✅ |
| **删除用户** | ❌ | ❌ | ✅ |
| **用户分析** | ❌ | ❌ | ✅ |
| **系统告警** | ❌ | ❌ | ✅ |
| **操作日志** | ❌ | ❌ | ✅ |
| **系统广播** | ❌ | ❌ | ✅ |
| **内容审核** | ❌ | ✅ | ✅ |
| **用户管理** | ❌ | ✅ | ✅ |
| **群组管理** | ❌ | ✅ | ✅ |
| **访问管理后台** | ❌ | ✅ | ✅ |
| **个人2FA** | ✅ | ✅ | ✅ |
| **个人设备** | ✅ | ✅ | ✅ |
| **发送消息** | ✅ | ✅ | ✅ |
| **文件上传** | ✅ | ✅ | ✅ |

---

## 💻 管理后台界面功能

### 超级管理员可以访问的所有页面

#### 1. 仪表盘（Dashboard）
- ✅ 系统概览
- ✅ 实时统计
- ✅ 快速操作

#### 2. 用户管理（Users）
- ✅ 用户列表
- ✅ 在线用户
- ✅ 用户详情
- ✅ 用户搜索
- ✅ 批量操作

#### 3. 聊天管理（Chats）
- ✅ 群组列表
- ✅ 群组详情
- ✅ 成员管理
- ✅ 权限配置

#### 4. 消息管理（Messages）
- ✅ 消息列表
- ✅ 消息搜索
- ✅ 消息删除
- ✅ 违规消息处理

#### 5. 系统管理（System）
- ✅ 系统配置
- ✅ 性能监控
- ✅ 告警管理
- ✅ 服务器状态

#### 6. 日志管理（Logs）
- ✅ 操作日志
- ✅ 审计日志
- ✅ 错误日志
- ✅ 访问日志

#### 7. 插件管理（Plugins）
- ✅ 插件列表
- ✅ 插件配置
- ✅ 插件启用/禁用

#### 8. 安全设置（Security）✨ v1.4.0新增
- ✅ 2FA设置（`/security/2fa`）
- ✅ 受信任设备管理
- ✅ 会话管理
- ✅ 权限配置

---

## 🎯 超级管理员完整API列表

### 系统管理（9个端点）
```
GET    /api/super-admin/stats                    # 系统统计
GET    /api/super-admin/stats/system             # 系统指标
GET    /api/super-admin/users/online             # 在线用户
POST   /api/super-admin/users/:id/logout         # 强制下线
POST   /api/super-admin/users/:id/ban            # 封禁用户
POST   /api/super-admin/users/:id/unban          # 解封用户
DELETE /api/super-admin/users/:id                # 删除用户
GET    /api/super-admin/users/:id/analysis       # 用户分析
GET    /api/super-admin/alerts                   # 系统告警
GET    /api/super-admin/logs                     # 操作日志
POST   /api/super-admin/broadcast                # 系统广播
```

### 内容审核（7个端点）
```
GET    /api/moderation/reports/pending           # 待审核列表
GET    /api/moderation/reports/:id               # 举报详情
POST   /api/moderation/reports/:id/handle        # 处理举报
POST   /api/moderation/filters                   # 创建过滤规则
GET    /api/moderation/users/:id/warnings        # 用户警告
GET    /api/moderation/statistics                # 审核统计
POST   /api/moderation/content/check             # 内容检测
```

### 用户管理（9个端点）
```
POST   /api/users/:id/blacklist                  # 添加黑名单
DELETE /api/users/:id/blacklist/:id              # 移除黑名单
GET    /api/users/:id/blacklist                  # 查看黑名单
GET    /api/users/:id/activity                   # 用户活动
POST   /api/users/:id/restrictions               # 设置限制
GET    /api/users/:id/restrictions               # 查看限制
POST   /api/users/:id/ban                        # 封禁
POST   /api/users/:id/unban                      # 解封
GET    /api/users/suspicious                     # 可疑用户
```

**超级管理员专属API**: 11个  
**管理员级API**: 16个  
**总管理API**: 27个

---

## 🛡️ 超级管理员的安全权限

### 数据访问权限
- ✅ 查看所有用户数据
- ✅ 查看所有消息记录
- ✅ 查看所有群组信息
- ✅ 查看所有文件记录
- ✅ 查看系统日志
- ✅ 查看审计日志

### 操作权限
- ✅ 强制用户下线
- ✅ 封禁/解封用户
- ✅ 删除用户账号
- ✅ 删除消息
- ✅ 删除群组
- ✅ 删除文件
- ✅ 修改用户权限
- ✅ 系统配置修改
- ✅ 广播系统消息

### 审计权限
- ✅ 查看所有操作日志
- ✅ 导出审计数据
- ✅ 追踪用户行为
- ✅ 风险评分分析

---

## 📋 v1.4.0 新增的超级管理员功能

### 可以管理所有用户的2FA和设备（未来扩展）

**建议在v1.4.1中添加**：

```go
// 超级管理员专用API
GET    /api/super-admin/users/:id/2fa/status      # 查看用户2FA状态
POST   /api/super-admin/users/:id/2fa/reset       # 强制重置2FA
GET    /api/super-admin/users/:id/devices         # 查看用户所有设备
POST   /api/super-admin/users/:id/devices/revoke  # 强制撤销设备
GET    /api/super-admin/security/overview         # 安全总览
```

---

## 🎯 使用示例

### 示例1：强制用户下线
```bash
curl -X POST http://localhost:8080/api/super-admin/users/123/logout \
  -H "Authorization: Bearer {super_admin_token}"

# 响应
{
  "success": true,
  "message": "用户已强制下线"
}
```

### 示例2：封禁用户24小时
```bash
curl -X POST http://localhost:8080/api/super-admin/users/123/ban \
  -H "Authorization: Bearer {super_admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "duration": 86400,
    "reason": "发送违规内容"
  }'
```

### 示例3：系统广播
```bash
curl -X POST http://localhost:8080/api/super-admin/broadcast \
  -H "Authorization: Bearer {super_admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "系统将于今晚23:00进行维护，预计1小时"
  }'
```

---

## 🔒 安全限制

### 超级管理员也受到的限制

1. **不能删除自己的账号**
2. **操作都会被记录到审计日志**
3. **不能修改系统核心配置**（需要重启服务）
4. **敏感操作需要二次确认**（前端实现）

---

## 📝 总结

### 超级管理员权限范围

**拥有**：
- ✅ 所有普通用户功能
- ✅ 所有管理员功能
- ✅ 系统管理功能（独有）
- ✅ 强制操作功能（独有）
- ✅ 全局数据访问

**数量统计**：
- 专属功能：11个API端点
- 管理员功能：16个API端点
- 普通功能：88个API端点
- **总计可用**：115个API端点

**权限级别**：**最高**  
**访问范围**：**全部**  
**限制**：**仅审计日志追踪**

---

**最后更新**: 2024-12-19  
**版本**: v1.4.0  
**文档状态**: ✅ 完整

