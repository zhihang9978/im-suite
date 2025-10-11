# 硬编码配置审计报告

**审计时间**: 2025-10-11 19:00  
**审计范围**: im-backend + im-admin  
**审计目标**: 识别并移除所有硬编码配置

---

## 🔴 严重问题（立即修复）

### 1. JWT密钥硬编码 🔴
**文件**: `im-backend/internal/service/auth_service.go:332`

**问题代码**:
```go
secretKey := []byte("zhihang_messenger_secret_key_2024")
```

**风险等级**: 🔴 **CRITICAL**  
**影响**: JWT密钥泄露，所有token可被伪造

**修复方案**:
```go
secretKey := []byte(os.Getenv("JWT_SECRET"))
if len(secretKey) == 0 {
    return nil, fmt.Errorf("JWT_SECRET环境变量未设置")
}
```

**状态**: ✅ 已修复

---

### 2. generateTokens方法中可能的硬编码
**文件**: `im-backend/internal/service/auth_service.go`

**需检查**: generateTokens方法是否也有硬编码

**状态**: ⏸️ 待检查

---

## 🟡 中等问题

### 3. 默认值配置
**文件**: `im-backend/config/database.go:18-22`

**代码**:
```go
host := getEnv("DB_HOST", "localhost")
port := getEnv("DB_PORT", "3306")
username := getEnv("DB_USER", "root")
password := getEnv("DB_PASSWORD", "")
database := getEnv("DB_NAME", "zhihang_messenger")
```

**评估**: 🟡 **中等风险**  
**说明**: 有默认值是好的，但生产环境应强制要求配置

**建议**: 生产模式下检查必需变量

---

## ✅ 已正确处理

### 4. 前端API基础路径
**文件**: `im-admin/src/api/request.js`

**代码**:
```javascript
const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})
```

**评估**: ✅ **正确**  
**说明**: 使用相对路径，由Nginx代理转发

---

## 📋 审计检查清单

### 后端检查
- [x] JWT密钥 - 🔴 发现硬编码，已修复
- [x] 数据库配置 - ✅ 使用环境变量
- [x] Redis配置 - ✅ 使用环境变量
- [ ] MinIO配置 - 待检查
- [ ] SMTP配置 - 待检查
- [ ] 第三方API密钥 - 待检查

### 前端检查
- [x] API基础URL - ✅ 使用相对路径
- [ ] WebSocket URL - 待检查
- [ ] 上传URL - 待检查
- [ ] WebRTC配置 - 待检查

---

## 🎯 修复计划

### 第1轮：严重问题（立即）
1. ✅ 修复JWT密钥硬编码
2. ⏸️ 检查其他密钥
3. ⏸️ 检查generateTokens方法

### 第2轮：环境变量验证（1天）
1. ⏸️ 添加启动时环境变量检查
2. ⏸️ 生产模式强制要求所有必需变量
3. ⏸️ 添加配置验证脚本

### 第3轮：配置文档化（0.5天）
1. ⏸️ 更新.env.example
2. ⏸️ 更新环境变量文档
3. ⏸️ 添加配置验证命令

---

**审计人**: AI Assistant  
**审计时间**: 2025-10-11 19:00

