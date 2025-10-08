# 志航密信技术架构文档

## 📋 目录

- [系统概述](#系统概述)
- [架构设计](#架构设计)
- [技术栈](#技术栈)
- [核心组件](#核心组件)
- [数据流](#数据流)
- [安全架构](#安全架构)
- [部署架构](#部署架构)
- [监控架构](#监控架构)
- [扩展性设计](#扩展性设计)

## 🏗️ 系统概述

志航密信是一个基于 Telegram 官方前端改造的即时通讯系统，采用现代化的微服务架构，支持多平台客户端，提供安全、快速、可靠的通讯服务。

### 核心特性

- **多平台支持**: Web、Android 全平台覆盖
- **实时通讯**: WebSocket 实时消息推送
- **端到端加密**: 消息传输加密保护
- **文件传输**: 支持多媒体文件传输
- **音视频通话**: WebRTC 音视频通话
- **群组管理**: 完整的群组聊天功能
- **管理后台**: 系统管理和监控

## 🏛️ 架构设计

### 整体架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │ Android Client  │    │  Admin Panel    │
│  (Telegram Web) │    │ (Telegram And.) │    │   (Vue3)        │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      Load Balancer        │
                    │        (Nginx)            │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │     API Gateway          │
                    │    (Go + Gin)            │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
    ┌─────┴─────┐         ┌─────┴─────┐         ┌─────┴─────┐
    │  Auth     │         │  Message  │         │  File     │
    │ Service   │         │ Service   │         │ Service   │
    └─────┬─────┘         └─────┬─────┘         └─────┬─────┘
          │                     │                     │
          └─────────────────────┼─────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    │    Data Layer         │
                    │  MySQL + Redis + MinIO│
                    └───────────────────────┘
```

### 分层架构

#### 1. 表示层 (Presentation Layer)
- **Web 客户端**: 基于 Telegram Web 改造的 React 应用
- **Android 客户端**: 基于 Telegram Android 改造的 Kotlin 应用
- **管理后台**: Vue3 + Element Plus 管理界面

#### 2. 网关层 (Gateway Layer)
- **Nginx 反向代理**: 负载均衡和 SSL 终止
- **API 网关**: 统一的 API 入口和路由

#### 3. 应用层 (Application Layer)
- **认证服务**: 用户认证和授权
- **消息服务**: 消息处理和分发
- **文件服务**: 文件上传和下载
- **通知服务**: 推送通知管理

#### 4. 数据层 (Data Layer)
- **MySQL**: 主数据库，存储用户、消息等结构化数据
- **Redis**: 缓存和会话存储
- **MinIO**: 对象存储，存储文件和媒体

## 🛠️ 技术栈

### 后端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 主要编程语言 |
| Gin | 1.9+ | Web 框架 |
| GORM | 1.25+ | ORM 框架 |
| Gorilla WebSocket | 1.5+ | WebSocket 支持 |
| JWT | 5.2+ | 身份认证 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存和会话 |
| MinIO | Latest | 对象存储 |

### 前端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| React | 18+ | Web 客户端框架 |
| TypeScript | 5.0+ | 类型安全的 JavaScript |
| Telegram Web | Latest | Web UI 组件库 |
| Kotlin | 1.9+ | Android 开发语言 |
| Jetpack Compose | 1.5+ | Android UI 框架 |
| Telegram Android | Latest | Android UI 组件库 |
| Vue3 | 3.3+ | 管理后台框架 |
| Element Plus | 2.3+ | UI 组件库 |

### 基础设施

| 技术 | 版本 | 用途 |
|------|------|------|
| Docker | 24.0+ | 容器化 |
| Docker Compose | 2.0+ | 容器编排 |
| Kubernetes | 1.28+ | 容器编排 |
| Nginx | 1.24+ | 反向代理 |
| Prometheus | 2.45+ | 监控 |
| Grafana | 10.0+ | 可视化 |

## 🔧 核心组件

### 1. 认证服务 (Auth Service)

**职责**:
- 用户注册和登录
- JWT 令牌管理
- 权限验证
- 会话管理

**核心功能**:
```go
type AuthService struct {
    db *gorm.DB
}

// 用户登录
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error)

// 用户注册
func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error)

// 令牌验证
func (s *AuthService) ValidateToken(token string) (*User, error)

// 令牌刷新
func (s *AuthService) RefreshToken(req RefreshRequest) (*RefreshResponse, error)
```

### 2. 消息服务 (Message Service)

**职责**:
- 消息发送和接收
- 消息存储和检索
- 消息状态管理
- 消息加密和解密

**核心功能**:
```go
type MessageService struct {
    db *gorm.DB
}

// 发送消息
func (s *MessageService) SendMessage(senderID uint, req SendMessageRequest) (*Message, error)

// 获取消息列表
func (s *MessageService) GetMessages(userID uint, req GetMessagesRequest) (*GetMessagesResponse, error)

// 编辑消息
func (s *MessageService) EditMessage(messageID uint, userID uint, content string) (*Message, error)

// 删除消息
func (s *MessageService) DeleteMessage(messageID uint, userID uint) error
```

### 3. WebSocket 服务 (WebSocket Service)

**职责**:
- 实时消息推送
- 连接管理
- 事件分发
- 心跳检测

**核心功能**:
```go
type Hub struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []byte
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
}

// 连接管理
func (h *Hub) run()

// 消息广播
func (h *Hub) broadcastMessage(message []byte)

// 连接注册
func (h *Hub) registerClient(conn *websocket.Conn)

// 连接注销
func (h *Hub) unregisterClient(conn *websocket.Conn)
```

### 4. 文件服务 (File Service)

**职责**:
- 文件上传和下载
- 文件类型验证
- 文件存储管理
- 缩略图生成

**核心功能**:
```go
type FileService struct {
    minioClient *minio.Client
}

// 文件上传
func (s *FileService) UploadFile(file *multipart.FileHeader) (*FileInfo, error)

// 文件下载
func (s *FileService) DownloadFile(fileID string) (*FileData, error)

// 文件删除
func (s *FileService) DeleteFile(fileID string) error

// 生成缩略图
func (s *FileService) GenerateThumbnail(fileID string) (*FileInfo, error)
```

## 📊 数据流

### 1. 消息发送流程

```
用户输入消息
    ↓
客户端验证
    ↓
发送到后端 API
    ↓
消息服务处理
    ↓
存储到数据库
    ↓
通过 WebSocket 推送给接收者
    ↓
接收者客户端显示消息
```

### 2. 文件传输流程

```
用户选择文件
    ↓
客户端文件验证
    ↓
上传到 MinIO
    ↓
返回文件 URL
    ↓
发送消息（包含文件 URL）
    ↓
存储消息到数据库
    ↓
推送给接收者
    ↓
接收者下载文件
```

### 3. 用户认证流程

```
用户输入凭据
    ↓
发送到认证服务
    ↓
验证用户信息
    ↓
生成 JWT 令牌
    ↓
返回令牌给客户端
    ↓
客户端存储令牌
    ↓
后续请求携带令牌
    ↓
服务端验证令牌
```

## 🔒 安全架构

### 1. 传输层安全

- **HTTPS/WSS**: 所有通信使用加密传输
- **SSL/TLS**: 使用现代加密协议
- **证书管理**: 支持 Let's Encrypt 自动证书

### 2. 身份认证

- **JWT 令牌**: 无状态的用户认证
- **令牌刷新**: 自动刷新过期令牌
- **会话管理**: Redis 存储会话信息

### 3. 数据加密

- **端到端加密**: 消息内容端到端加密
- **数据库加密**: 敏感数据加密存储
- **文件加密**: 上传文件加密存储

### 4. 访问控制

- **权限验证**: 细粒度的权限控制
- **API 限流**: 防止恶意请求
- **IP 白名单**: 限制访问来源

## 🚀 部署架构

### 1. 开发环境

```yaml
# docker-compose.dev.yml
services:
  mysql:      # 开发数据库
  redis:      # 开发缓存
  minio:      # 开发文件存储
  backend:    # 后端服务
  web:        # Web 客户端
  admin:      # 管理后台
  nginx:      # 反向代理
```

### 2. 生产环境

```yaml
# docker-compose.prod.yml
services:
  mysql:      # 生产数据库（主从复制）
  redis:      # 生产缓存（集群）
  minio:      # 生产文件存储（分布式）
  backend:    # 后端服务（多实例）
  web:        # Web 客户端（CDN）
  admin:      # 管理后台
  nginx:      # 反向代理（负载均衡）
  prometheus: # 监控
  grafana:    # 可视化
```

### 3. Kubernetes 部署

```yaml
# k8s/
├── namespace.yaml           # 命名空间
├── backend-deployment.yaml  # 后端部署
├── web-deployment.yaml      # Web 部署
├── admin-deployment.yaml    # 管理后台部署
├── mysql-deployment.yaml    # 数据库部署
├── redis-deployment.yaml    # 缓存部署
├── minio-deployment.yaml    # 文件存储部署
└── ingress.yaml            # 入口配置
```

## 📈 监控架构

### 1. 指标收集

- **应用指标**: 请求量、响应时间、错误率
- **系统指标**: CPU、内存、磁盘、网络
- **业务指标**: 用户数、消息数、文件数

### 2. 日志管理

- **应用日志**: 结构化日志输出
- **访问日志**: Nginx 访问日志
- **错误日志**: 错误和异常日志
- **审计日志**: 用户操作审计

### 3. 告警机制

- **阈值告警**: 基于指标的阈值告警
- **异常告警**: 异常模式检测告警
- **业务告警**: 业务异常告警
- **通知渠道**: 邮件、短信、钉钉等

## 🔄 扩展性设计

### 1. 水平扩展

- **无状态服务**: 后端服务无状态设计
- **负载均衡**: 支持多实例负载均衡
- **数据库分片**: 支持数据库水平分片
- **缓存集群**: Redis 集群支持

### 2. 垂直扩展

- **资源动态调整**: 支持动态资源调整
- **性能优化**: 持续性能优化
- **容量规划**: 基于监控的容量规划

### 3. 微服务化

- **服务拆分**: 按业务领域拆分服务
- **服务发现**: 自动服务发现和注册
- **配置管理**: 集中化配置管理
- **API 网关**: 统一的 API 网关

## 📚 相关文档

- [API 文档](../api/README.md)
- [部署文档](../deployment/README.md)
- [安全文档](../security/README.md)
- [监控文档](../monitoring/README.md)
- [开发文档](../development/README.md)

## 🔗 外部链接

- [Go 官方文档](https://golang.org/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [React 官方文档](https://reactjs.org/docs/)
- [Vue3 官方文档](https://vuejs.org/guide/)
- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)


