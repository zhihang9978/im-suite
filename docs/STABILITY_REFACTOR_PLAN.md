# 志航密信 - 全面稳定性整改计划

**开始时间**: 2025-10-11 19:00  
**预计完成**: 2025-10-18（7天）  
**整改目标**: 全功能通过、全风险修复、CI+测试全绿

---

## 📋 整改任务清单

| # | 任务 | 优先级 | 预计时间 | 状态 |
|---|------|--------|---------|------|
| 1 | ~~Android客户端~~ | ~~P1~~ | - | ❌ **跳过**（不可触碰） |
| 2 | 补充自动化测试 | P0 | 3天 | 🟡 进行中 |
| 3 | 环境变量与配置隔离 | P0 | 1天 | ⏸️ 待开始 |
| 4 | 增强部署脚本健壮性 | P0 | 1天 | ⏸️ 待开始 |
| 5 | 文档代码对齐检查 | P1 | 0.5天 | ⏸️ 待开始 |
| 6 | 权限与安全控制校验 | P0 | 1天 | ⏸️ 待开始 |
| 7 | 增加监控/可观测性 | P1 | 1天 | ⏸️ 待开始 |
| 8 | 错误处理全线打磨 | P0 | 1天 | ⏸️ 待开始 |
| 9 | 数据库迁移安全机制 | P1 | 0.5天 | ⏸️ 待开始 |
| 10 | 容错/灾备/多节点切换 | P2 | 1天 | ⏸️ 待开始 |

**总计**: 9项任务（跳过Android），预计10天

---

## ⚠️ 关于Android客户端的说明

**telegram-android/** 和 **telegram-web/** 这两个目录在 `.gitignore` 中被忽略。

**原因**:
1. 这是外部Telegram客户端的fork
2. 体积巨大（数万个文件）
3. **触碰这些文件会导致网络错误**

**处理方式**: ❌ **完全跳过，不做任何处理**

---

## 📊 当前状态基线

### 已完成的基础设施
- ✅ 10个运维脚本（2,400行）
- ✅ 18份文档（12,000行）
- ✅ 监控配置（Prometheus + Grafana）
- ✅ CI/CD工作流（3个）
- ✅ 7项核心功能（Token刷新、ACK、重连等）
- ✅ 错误处理框架（Recovery、ErrorBoundary）

### 当前问题
- 🔴 测试覆盖率: 0%（目标≥60%）
- 🔴 硬编码配置: 待清理
- 🔴 部署脚本: 待增强
- 🔴 权限校验: 待验证
- 🔴 监控指标: 待实现
- 🔴 错误处理: 待完善
- 🔴 数据库迁移: 无回滚脚本

---

## 🎯 第1阶段：补充自动化测试（3天）

### Day 1: 单元测试
**目标**: 覆盖率达到60%

**需编写的测试**:
```
tests/unit/
├── auth_service_test.go ✅ (已有)
├── token_refresh_service_test.go ✅ (已有)
├── message_ack_service_test.go ✅ (已有)
├── user_service_test.go (新增)
├── message_service_test.go (新增)
├── file_service_test.go (新增)
├── group_service_test.go (新增)
├── webrtc_service_test.go (新增)
├── content_moderation_service_test.go (新增)
└── ... (15-20个文件)
```

**验收标准**:
```bash
cd im-backend
go test ./tests/unit/... -v -cover -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# 预期: total coverage: 60.0%
```

---

### Day 2: 集成测试
**目标**: 核心业务流程全绿

**测试场景**:
```
tests/integration/
├── auth_integration_test.go
│   └── 测试：注册 → 登录 → Token验证 → 登出
├── message_integration_test.go
│   └── 测试：发送 → ACK → 接收 → 已读
├── file_integration_test.go
│   └── 测试：上传 → 下载 → 删除
└── webrtc_integration_test.go
    └── 测试：信令交换 → ICE candidate
```

**环境要求**: MySQL + Redis + MinIO

---

### Day 3: E2E测试
**目标**: 完整用户流程通过

**测试流程**:
1. 用户注册
2. 用户登录
3. 搜索并添加好友
4. 发送文本消息
5. 发送图片
6. 发送文件
7. 发起语音通话
8. 发起视频通话
9. 接收通知
10. 用户注销

**工具**: Postman/Bruno集合 或 自动化脚本

---

## 🎯 第2阶段：配置与安全（2天）

### Day 4: 环境变量隔离
**扫描范围**:
- 所有`.go`文件
- 所有`.js/.vue`文件
- 所有配置文件

**检查项**:
```bash
# 查找硬编码
grep -r "localhost" im-backend/internal/
grep -r "127.0.0.1" im-backend/internal/
grep -r "password.*=" im-backend/internal/
grep -r "secret.*=" im-backend/internal/
```

**修复方案**:
```go
// 修复前
db, err := gorm.Open("mysql://root:password@localhost:3306/db")

// 修复后
host := os.Getenv("DB_HOST")
pass := os.Getenv("DB_PASSWORD")
db, err := gorm.Open(fmt.Sprintf("mysql://root:%s@%s:3306/db", pass, host))
```

---

### Day 5: 权限校验
**检查文件**: `PERMISSION_SYSTEM_COMPLETE.md`

**验证项**:
- [ ] 所有管理员API有role检查
- [ ] 审计日志完整记录
- [ ] 日志不可删除
- [ ] 前端绕过防护

**实现**:
```go
// 在middleware中统一检查
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if userRole != role {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用
adminAPI.Use(middleware.RequireRole("admin"))
```

---

## 🎯 第3阶段：监控与容错（2天）

### Day 6: 监控指标
**实现Prometheus metrics**:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// 添加metrics端点
r.GET("/metrics", gin.WrapH(promhttp.Handler()))

// 自定义指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(...)
    httpRequestDuration = prometheus.NewHistogramVec(...)
    activeUsers = prometheus.NewGauge(...)
    webrtcConnections = prometheus.NewGauge(...)
)
```

**Grafana面板**: 已有，需集成到代码

---

### Day 7: 错误处理
**全面检查**:
```bash
# 查找未处理的error
grep -r "err :=" im-backend/internal/ | grep -v "if err"
grep -r "panic\|Fatal" im-backend/internal/

# 查找未捕获的Promise
grep -r "async\|await" im-admin/src/ | grep -v "try"
```

**修复模式**:
```go
// 统一错误包装
if err != nil {
    return fmt.Errorf("操作失败 %s: %w", context, err)
}

// 添加重试
err := retry.Do(
    func() error { return operation() },
    retry.Attempts(3),
    retry.Delay(time.Second),
)
```

---

## 🎯 第4阶段：部署与灾备（2天）

### Day 8-9: 增强部署脚本
**增强项**:
1. ✅ 部署前检查（DB、Redis连通性）
2. ✅ 健康检查失败退出
3. ✅ 自动回滚
4. ✅ 重试机制
5. ✅ 超时控制
6. ⏸️ 灰度发布支持
7. ⏸️ 蓝绿部署

**已有**: `ops/deploy.sh`已实现大部分  
**需增强**: 灰度发布、蓝绿部署

---

### Day 10: 容错与灾备
**设计内容**:
1. 备份节点配置
2. DNS切换脚本
3. 负载均衡配置
4. 故障切换流程
5. 切换演练文档

---

## 📋 PR流程规范

每个PR必须包含：

### PR模板
```markdown
## 改动摘要
<!-- 简要说明改了什么 -->

## 风险说明
<!-- 可能的风险和影响范围 -->

## 回滚方案
<!-- 如何快速回滚此改动 -->

## 测试证据
<!-- 单测截图、日志、覆盖率报告 -->
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 本地验证通过
- [ ] 代码Review通过

## Checklist
- [ ] 代码已通过lint
- [ ] 测试覆盖率≥60%
- [ ] 文档已更新
- [ ] 无硬编码配置
- [ ] 错误处理完整
```

---

## 📊 执行计划（10天）

### Week 1: 测试与配置
- Day 1-3: 补充测试（覆盖率0% → 60%+）
- Day 4: 环境变量隔离
- Day 5: 权限校验

### Week 2: 监控与容错
- Day 6: 监控指标实现
- Day 7: 错误处理完善
- Day 8-9: 部署脚本增强
- Day 10: 容错灾备设计

---

## ✅ 最终验收标准

### 必须达成
- [ ] CI全绿（lint + build + test）
- [ ] 单元测试覆盖率≥60%
- [ ] 集成测试全通过
- [ ] E2E测试全通过
- [ ] 无硬编码配置
- [ ] 所有API有权限校验
- [ ] Prometheus metrics可用
- [ ] 部署脚本支持回滚
- [ ] 数据库迁移有回滚脚本
- [ ] 容错方案文档完整

---

**计划制定人**: AI Assistant  
**计划时间**: 2025-10-11 19:00

