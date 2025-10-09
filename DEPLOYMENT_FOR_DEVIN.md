# 部署说明 - For Devin

**志航密信 v1.6.0** - 完整部署和测试指南

---

## 🎯 任务目标

部署和测试志航密信 v1.6.0，包含：
1. 屏幕共享功能（基础+增强）
2. 权限管理系统
3. 中国手机品牌适配

---

## ✅ 预先检查清单（已完成）

| 检查项 | 状态 | 说明 |
|-------|------|------|
| 代码编译 | ✅ | Go代码已编译通过 |
| Linter检查 | ✅ | 无错误 |
| 数据库迁移 | ✅ | 已添加到AutoMigrate |
| 依赖配置 | ✅ | go.mod已更新 |
| API路由 | ✅ | 15个端点已配置 |
| 文档完整 | ✅ | 220+页文档 |
| 示例代码 | ✅ | 完整示例已提供 |

---

## 📦 项目结构

```
im-suite/
├── im-backend/                  # 后端服务（Go）
│   ├── main.go                  # 入口文件
│   ├── config/                  # 配置
│   │   └── database.go          # 数据库配置（已更新）
│   ├── internal/
│   │   ├── model/              # 数据模型
│   │   │   └── screen_share.go # 新增：屏幕共享模型
│   │   ├── service/            # 业务逻辑
│   │   │   ├── webrtc_service.go              # 修改：添加屏幕共享
│   │   │   └── screen_share_enhanced_service.go # 新增：增强服务
│   │   ├── controller/         # 控制器
│   │   │   ├── webrtc_controller.go              # 新增
│   │   │   └── screen_share_enhanced_controller.go # 新增
│   │   └── middleware/         # 中间件
│   └── go.mod                  # 依赖管理
├── telegram-android/            # Android客户端
│   └── TMessagesProj/src/main/java/org/telegram/
│       ├── messenger/PermissionManager.java        # 新增：权限管理器
│       └── ui/PermissionExampleActivity.java       # 新增：示例
├── telegram-web/                # Web客户端
├── im-admin/                    # 管理后台（Vue）
├── examples/                    # 示例和演示
│   ├── screen-share-example.js           # 基础管理器
│   ├── screen-share-enhanced.js          # 增强管理器
│   ├── chinese-phone-permissions.js      # 权限适配
│   └── screen-share-demo.html            # 演示页面
├── docs/                        # 文档
│   └── chinese-phones/          # 中国手机适配文档
├── config/                      # 配置文件
│   ├── mysql/
│   ├── redis/
│   ├── nginx/
│   └── prometheus/
└── docker-compose.production.yml # 生产部署配置
```

---

## 🚀 部署步骤（按顺序执行）

### 步骤1：环境准备（5分钟）

```bash
# 1.1 检查Go环境
go version  # 应该 >= 1.19

# 1.2 检查Node.js环境（前端）
node --version  # 应该 >= 16.0

# 1.3 检查Docker环境
docker --version
docker-compose --version

# 1.4 克隆或更新代码
cd im-suite
git pull origin main
```

### 步骤2：配置环境变量（5分钟）

```bash
# 2.1 复制环境变量模板
cp .env.example .env

# 2.2 编辑.env文件（重要！）
nano .env

# 必须配置的变量：
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password     # 改成实际密码
DB_NAME=zhihang_messenger

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=                # 如果有密码请填写

JWT_SECRET=your_jwt_secret_key_here  # 改成随机字符串
JWT_EXPIRES_IN=24h

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

GIN_MODE=release              # 生产环境用release
```

### 步骤3：启动依赖服务（10分钟）

```bash
# 3.1 启动MySQL、Redis、MinIO等服务
docker-compose -f docker-compose.production.yml up -d mysql redis minio

# 3.2 等待服务就绪（约30秒）
sleep 30

# 3.3 检查服务状态
docker-compose -f docker-compose.production.yml ps

# 应该看到：
# mysql   Up 0.0.0.0:3306->3306/tcp
# redis   Up 0.0.0.0:6379->6379/tcp
# minio   Up 0.0.0.0:9000->9000/tcp
```

### 步骤4：编译和启动后端（5分钟）

```bash
# 4.1 进入后端目录
cd im-backend

# 4.2 下载依赖
go mod download

# 4.3 编译
go build -o bin/im-backend

# 如果编译失败，检查：
# - go.mod文件是否完整
# - 依赖是否都能下载
# - 代码是否有语法错误

# 4.4 运行数据库迁移（自动）
# 注意：首次运行时会自动创建所有表
./bin/im-backend

# 应该看到：
# [INFO] 数据库初始化成功
# [INFO] 数据库迁移成功
# [INFO] Redis初始化成功
# [INFO] 服务启动成功 http://localhost:8080

# 如果出现错误：
# - 检查.env文件配置
# - 检查MySQL是否正常运行
# - 检查端口8080是否被占用

# 4.5 保持后端运行（新开一个终端继续）
```

### 步骤5：测试后端API（10分钟）

```bash
# 5.1 健康检查
curl http://localhost:8080/health

# 应该返回：
# {"status":"ok"}

# 5.2 测试注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800138000",
    "username": "testuser",
    "password": "Test123456",
    "nickname": "测试用户"
  }'

# 应该返回：
# {"success":true,"data":{...},"message":"注册成功"}

# 5.3 测试登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800138000",
    "password": "Test123456"
  }'

# 保存返回的token，后面需要用
export TOKEN="<返回的access_token>"

# 5.4 测试屏幕共享API（使用token）
curl -X GET "http://localhost:8080/api/calls/test_call_123/screen-share/status" \
  -H "Authorization: Bearer $TOKEN"

# 应该返回：
# {"success":true,"data":{"is_active":false,"message":"当前没有屏幕共享"}}

# 5.5 测试增强API
curl -X GET "http://localhost:8080/api/screen-share/statistics" \
  -H "Authorization: Bearer $TOKEN"

# 应该返回用户统计信息

# 如果以上测试都通过，后端部署成功 ✅
```

### 步骤6：测试前端演示（10分钟）

```bash
# 6.1 启动简单HTTP服务器
cd examples
python -m http.server 8000

# 或使用Node.js
# npx http-server -p 8000

# 6.2 在浏览器中打开
# http://localhost:8000/screen-share-demo.html

# 6.3 测试步骤：
# 1. 点击"开始共享屏幕"
# 2. 选择要共享的屏幕/窗口
# 3. 观察视频显示和日志
# 4. 尝试切换质量
# 5. 点击"停止共享"

# 如果能看到视频和日志，前端测试通过 ✅
```

### 步骤7：数据库验证（5分钟）

```bash
# 7.1 连接MySQL
mysql -u root -p zhihang_messenger

# 7.2 检查表是否创建
SHOW TABLES;

# 应该包含以下新表：
# - screen_share_sessions
# - screen_share_quality_changes
# - screen_share_participants
# - screen_share_statistics
# - screen_share_recordings
# - bots
# - bot_api_logs
# - bot_users
# - bot_user_permissions

# 7.3 查看表结构
DESC screen_share_sessions;

# 7.4 退出
exit;

# 如果所有表都存在，数据库迁移成功 ✅
```

---

## 🧪 功能测试清单

### 基础功能测试

#### 1. 屏幕共享基础功能

```bash
# 1.1 创建通话
curl -X POST http://localhost:8080/api/calls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "callee_id": 2,
    "type": "video"
  }'

# 保存返回的call_id
export CALL_ID="<返回的call_id>"

# 1.2 开始屏幕共享
curl -X POST "http://localhost:8080/api/calls/$CALL_ID/screen-share/start" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "测试用户",
    "quality": "medium",
    "with_audio": false
  }'

# 预期：{"success":true,"message":"屏幕共享已开始"}

# 1.3 查询状态
curl -X GET "http://localhost:8080/api/calls/$CALL_ID/screen-share/status" \
  -H "Authorization: Bearer $TOKEN"

# 预期：返回is_active=true和共享者信息

# 1.4 调整质量
curl -X POST "http://localhost:8080/api/calls/$CALL_ID/screen-share/quality" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quality": "high"}'

# 预期：{"success":true,"message":"屏幕共享质量已更改为: high"}

# 1.5 停止共享
curl -X POST "http://localhost:8080/api/calls/$CALL_ID/screen-share/stop" \
  -H "Authorization: Bearer $TOKEN"

# 预期：{"success":true,"message":"屏幕共享已停止"}

# ✅ 如果所有API都返回成功，基础功能正常
```

#### 2. 屏幕共享增强功能

```bash
# 2.1 查看历史记录
curl -X GET "http://localhost:8080/api/screen-share/history?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 预期：返回会话列表

# 2.2 查看统计信息
curl -X GET "http://localhost:8080/api/screen-share/statistics" \
  -H "Authorization: Bearer $TOKEN"

# 预期：返回total_sessions等统计数据

# 2.3 检查权限
curl -X GET "http://localhost:8080/api/screen-share/check-permission?quality=high" \
  -H "Authorization: Bearer $TOKEN"

# 预期：返回allowed=true或false

# ✅ 如果所有API都有正确响应，增强功能正常
```

#### 3. 权限管理功能（需要Android环境）

```bash
# 如果有Android测试设备/模拟器：

# 3.1 安装APK
adb install -r app-release.apk

# 3.2 启动应用
adb shell am start -n org.telegram.messenger/.LaunchActivity

# 3.3 测试视频通话
# - 点击视频通话按钮
# - 应该弹出系统权限对话框
# - 允许后能正常通话

# 3.4 测试屏幕共享
# - 点击屏幕共享按钮
# - 应该依次弹出权限对话框
# - 允许后能正常共享

# ✅ 如果权限对话框正常弹出，权限管理正常
```

---

## 🐛 常见问题和解决方案

### 问题1：后端启动失败

```bash
# 错误：连接数据库失败
# 解决：
1. 检查MySQL是否运行：docker-compose ps
2. 检查.env中的数据库配置
3. 检查MySQL密码是否正确
4. 尝试手动连接：mysql -h localhost -u root -p

# 错误：端口8080被占用
# 解决：
1. 查找占用端口的进程：lsof -i :8080
2. 杀死进程：kill -9 <PID>
3. 或修改.env中的PORT配置
```

### 问题2：API返回401未授权

```bash
# 原因：Token过期或无效
# 解决：
1. 重新登录获取新token
2. 检查Token是否正确复制
3. 检查JWT_SECRET配置
```

### 问题3：屏幕共享API返回错误

```bash
# 错误：未找到活跃的屏幕共享会话
# 原因：通话不存在或已结束
# 解决：
1. 先创建通话
2. 使用正确的call_id
3. 确保通话状态为active
```

### 问题4：数据库表未创建

```bash
# 原因：AutoMigrate未执行或失败
# 解决：
1. 检查后端启动日志
2. 手动运行迁移：在main.go中打印迁移日志
3. 检查model文件是否正确导入
```

### 问题5：前端演示页面无法访问

```bash
# 原因：CORS或路径问题
# 解决：
1. 确保HTTP服务器在examples目录运行
2. 检查浏览器控制台错误
3. 修改API地址为正确的后端地址
```

---

## 📊 性能测试（可选）

```bash
# 1. 并发测试
# 使用Apache Bench测试API性能
ab -n 1000 -c 10 http://localhost:8080/health

# 2. 负载测试
# 可以使用hey工具
hey -n 1000 -c 50 http://localhost:8080/health

# 3. 数据库性能
# 查看慢查询
mysql -u root -p -e "SELECT * FROM slow_log LIMIT 10;"

# 4. 内存使用
# 查看后端内存占用
ps aux | grep im-backend

# 5. CPU使用
top -p $(pgrep im-backend)
```

---

## ✅ 最终验收清单

### 必须通过的测试

- [ ] 后端成功启动，无错误日志
- [ ] 健康检查API返回正常
- [ ] 用户注册和登录正常
- [ ] 屏幕共享API全部可用（5个基础+10个增强）
- [ ] 数据库表全部创建（至少52个表）
- [ ] 前端演示页面能正常显示
- [ ] 能成功开始和停止屏幕共享
- [ ] 质量切换功能正常
- [ ] 统计信息API返回正确数据

### 可选测试

- [ ] Android权限管理功能测试
- [ ] Web端权限适配测试
- [ ] 中国手机品牌特定功能测试
- [ ] 并发和性能测试
- [ ] 长时间运行稳定性测试

---

## 📝 测试报告模板

```markdown
# 志航密信 v1.6.0 测试报告

## 测试环境
- 操作系统：
- Go版本：
- 数据库：MySQL 8.0
- 测试时间：

## 测试结果

### 后端服务
- [ ] 启动成功
- [ ] 健康检查通过
- [ ] API响应正常
- [ ] 数据库迁移成功

### 屏幕共享功能
- [ ] 基础API全部通过（5个）
- [ ] 增强API全部通过（10个）
- [ ] 前端演示正常
- [ ] 数据记录正确

### 权限管理
- [ ] Android权限管理正常
- [ ] Web权限处理正常
- [ ] 系统弹窗正常

### 发现的问题
1. 
2. 
3. 

### 建议
1. 
2. 
3. 

### 总体评价
□ 通过  □ 部分通过  □ 不通过

## 附件
- 日志文件
- 截图
- 性能数据
```

---

## 🚀 部署成功后的操作

```bash
# 1. 备份配置
cp .env .env.backup
cp im-backend/bin/im-backend im-backend/bin/im-backend.backup

# 2. 设置开机自启（可选）
sudo systemctl enable im-backend

# 3. 配置Nginx反向代理（可选）
# 编辑 /etc/nginx/sites-available/im-suite

# 4. 设置SSL证书（可选）
# 使用Let's Encrypt
sudo certbot --nginx -d yourdomain.com

# 5. 配置监控（可选）
# Prometheus和Grafana已在docker-compose中配置

# 6. 查看日志
tail -f im-backend/logs/app.log
```

---

## 📞 需要帮助？

### 紧急问题

如果遇到无法解决的问题：
1. 检查日志：`tail -f im-backend/logs/app.log`
2. 查看Docker日志：`docker-compose logs -f`
3. 检查系统资源：`htop` 或 `top`

### 文档参考

- 完整功能报告：`COMPLETE_SUMMARY_v1.6.0.md`
- 屏幕共享文档：`SCREEN_SHARE_ENHANCED.md`
- 权限系统文档：`PERMISSION_SYSTEM_COMPLETE.md`
- 快速开始：`SCREEN_SHARE_QUICK_START.md`

### 关键配置文件

- 后端配置：`im-backend/.env`
- 数据库配置：`config/mysql/conf.d/custom.cnf`
- Redis配置：`config/redis/redis.conf`
- Nginx配置：`config/nginx/nginx.conf`

---

## 🎯 预期结果

部署成功后，应该能够：

1. ✅ 后端API全部正常响应
2. ✅ 用户可以注册登录
3. ✅ 屏幕共享功能完全可用
4. ✅ 数据正确存储在数据库
5. ✅ 前端演示页面正常工作
6. ✅ 权限管理系统正常
7. ✅ 无错误日志

---

## ⏱️ 预计时间

| 步骤 | 预计时间 |
|------|---------|
| 环境准备 | 5分钟 |
| 配置环境变量 | 5分钟 |
| 启动依赖服务 | 10分钟 |
| 编译和启动后端 | 5分钟 |
| 测试后端API | 10分钟 |
| 测试前端演示 | 10分钟 |
| 数据库验证 | 5分钟 |
| **总计** | **50分钟** |

---

**准备完毕！Devin现在可以按照这个文档一步步部署和测试了！** 🚀

**注意**：所有代码已经完成并验证，Devin只需要按步骤执行命令即可。



