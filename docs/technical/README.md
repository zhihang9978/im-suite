# 志航密信 - 技术文档

## 📖 文档概述

本文档提供志航密信项目的详细技术文档，包括架构设计、API文档、部署指南、开发指南等。

## 🏗️ 系统架构

### 整体架构图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   管理后台      │    │   移动端应用    │
│  (telegram-web) │    │   (im-admin)    │    │ (telegram-android) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   后端服务      │
                    │  (im-backend)   │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MySQL 数据库  │    │   Redis 缓存    │    │   MinIO 存储    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 技术栈

#### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin (Web框架)
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **存储**: MinIO (对象存储)
- **WebSocket**: Gorilla WebSocket
- **ORM**: GORM
- **认证**: JWT
- **日志**: Logrus

#### 前端技术栈
- **Web端**: React + Vite + TypeScript
- **管理后台**: Vue3 + Element Plus + Pinia
- **移动端**: Kotlin + Jetpack Compose
- **UI框架**: Ant Design (Web), Material Design (Android)
- **状态管理**: Pinia (Vue), 自定义 (Android)
- **HTTP客户端**: Axios (Web), Retrofit2 (Android)

#### 基础设施
- **容器化**: Docker + Docker Compose
- **反向代理**: Nginx
- **监控**: 自定义监控系统
- **CI/CD**: GitHub Actions
- **部署**: Docker Swarm / Kubernetes

## 🔧 开发环境搭建

### 环境要求
- Go 1.21+
- Node.js 18+
- Docker 20+
- MySQL 8.0
- Redis 7.0

### 快速开始
```bash
# 1. 克隆项目
git clone https://github.com/your-org/zhihang-messenger.git
cd zhihang-messenger

# 2. 启动基础设施
docker-compose up -d mysql redis minio

# 3. 启动后端服务
cd im-backend
go mod download
go run main.go

# 4. 启动前端服务
cd ../telegram-web
npm install
npm run dev

# 5. 启动管理后台
cd ../im-admin
npm install
npm run dev
```

## 📡 API 文档

### 认证 API
```http
POST /api/auth/login
Content-Type: application/json

{
  "phone": "13800138000",
  "password": "password123"
}
```

### 用户 API
```http
GET /api/users/me
Authorization: Bearer <token>
```

### 消息 API
```http
POST /api/messages/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "chat_id": "chat_001",
  "content": "Hello World",
  "type": "text"
}
```

### WebSocket API
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};
```

## 🗄️ 数据库设计

### 用户表 (users)
```sql
CREATE TABLE users (
  id VARCHAR(36) PRIMARY KEY,
  phone VARCHAR(20) UNIQUE NOT NULL,
  username VARCHAR(50),
  nickname VARCHAR(100),
  avatar VARCHAR(255),
  bio TEXT,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 聊天表 (chats)
```sql
CREATE TABLE chats (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(100),
  type ENUM('private', 'group', 'channel') NOT NULL,
  created_by VARCHAR(36),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (created_by) REFERENCES users(id)
);
```

### 消息表 (messages)
```sql
CREATE TABLE messages (
  id VARCHAR(36) PRIMARY KEY,
  chat_id VARCHAR(36) NOT NULL,
  sender_id VARCHAR(36) NOT NULL,
  content TEXT NOT NULL,
  type ENUM('text', 'image', 'video', 'audio', 'file') NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (chat_id) REFERENCES chats(id),
  FOREIGN KEY (sender_id) REFERENCES users(id)
);
```

## 🔐 安全设计

### 认证与授权
- JWT Token 认证
- Token 刷新机制
- 权限控制中间件
- 会话管理

### 数据加密
- 密码哈希 (bcrypt)
- 传输加密 (HTTPS/WSS)
- 端到端加密 (可选)
- 敏感数据加密存储

### 安全措施
- 输入验证和过滤
- SQL 注入防护
- XSS 防护
- CSRF 防护
- 速率限制

## 🚀 部署指南

### Docker 部署
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 生产环境部署
```bash
# 使用 Docker Swarm
docker stack deploy -c docker-stack.yml zhihang-messenger

# 使用 Kubernetes
kubectl apply -f k8s/
```

## 📊 监控与日志

### 性能监控
- API 响应时间监控
- 数据库性能监控
- 内存和CPU使用监控
- 网络流量监控

### 错误追踪
- 应用错误日志
- 系统错误日志
- 用户行为日志
- 安全事件日志

### 日志格式
```json
{
  "timestamp": "2024-12-19T10:30:00Z",
  "level": "INFO",
  "service": "im-backend",
  "message": "用户登录成功",
  "user_id": "user_001",
  "ip": "192.168.1.100",
  "request_id": "req_001"
}
```

## 🔄 CI/CD 流程

### 持续集成
1. 代码提交触发构建
2. 运行单元测试
3. 运行集成测试
4. 代码质量检查
5. 安全扫描

### 持续部署
1. 构建 Docker 镜像
2. 推送到镜像仓库
3. 部署到测试环境
4. 运行自动化测试
5. 部署到生产环境

## 🧪 测试策略

### 单元测试
- 后端 Go 单元测试
- 前端 React 组件测试
- 移动端 Kotlin 单元测试

### 集成测试
- API 集成测试
- 数据库集成测试
- WebSocket 集成测试

### 端到端测试
- 用户流程测试
- 跨平台测试
- 性能测试

## 📈 性能优化

### 后端优化
- 数据库查询优化
- 缓存策略优化
- 并发处理优化
- 内存使用优化

### 前端优化
- 代码分割
- 懒加载
- 图片优化
- 缓存策略

### 移动端优化
- 内存管理
- 电池优化
- 网络优化
- 启动优化

## 🔧 故障排除

### 常见问题
1. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接配置
   - 检查网络连接

2. **WebSocket 连接失败**
   - 检查防火墙设置
   - 验证 WebSocket 配置
   - 检查代理设置

3. **文件上传失败**
   - 检查 MinIO 服务状态
   - 验证存储配置
   - 检查文件大小限制

### 日志分析
```bash
# 查看应用日志
docker-compose logs im-backend

# 查看数据库日志
docker-compose logs mysql

# 查看 Redis 日志
docker-compose logs redis
```

## 📞 技术支持

### 联系方式
- 邮箱: support@zhihang-messenger.com
- 文档: https://docs.zhihang-messenger.com
- 社区: https://community.zhihang-messenger.com

### 贡献指南
1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

---

**最后更新**: 2024年12月19日
**文档版本**: v1.0.0
