# 志航密信开发指南

## 📋 目录

- [开发环境搭建](#开发环境搭建)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [测试指南](#测试指南)
- [调试技巧](#调试技巧)
- [常见问题](#常见问题)

## 🛠️ 开发环境搭建

### 系统要求

- **操作系统**: Windows 10+, macOS 10.15+, Ubuntu 20.04+
- **内存**: 8GB 以上
- **存储**: 20GB 可用空间
- **网络**: 稳定的互联网连接

### 必需软件

#### 1. Go 开发环境

```bash
# 安装 Go 1.21+
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 2. Node.js 开发环境

```bash
# 使用 nvm 安装 Node.js
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc

# 安装 Node.js 18+
nvm install 18
nvm use 18

# 验证安装
node --version
npm --version
```

#### 3. Docker 环境

```bash
# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

#### 4. 数据库工具

```bash
# MySQL 客户端
sudo apt-get install mysql-client

# Redis 客户端
sudo apt-get install redis-tools
```

### 开发工具推荐

#### IDE 和编辑器

- **VS Code**: 推荐使用，支持 Go、TypeScript、Vue 等
- **GoLand**: JetBrains 出品的 Go IDE
- **IntelliJ IDEA**: 支持多种语言的 IDE

#### VS Code 插件推荐

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-typescript-next",
    "vue.volar",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "ms-azuretools.vscode-docker"
  ]
}
```

## 📁 项目结构

### 整体结构

```
im-suite/
├── im-backend/              # 后端服务
│   ├── cmd/                 # 应用入口
│   ├── internal/            # 内部包
│   │   ├── config/          # 配置管理
│   │   ├── controller/      # 控制器
│   │   ├── service/         # 业务逻辑
│   │   ├── model/           # 数据模型
│   │   ├── middleware/      # 中间件
│   │   └── utils/           # 工具函数
│   ├── pkg/                 # 公共包
│   ├── scripts/             # 脚本文件
│   ├── tests/               # 测试文件
│   ├── go.mod               # Go 模块文件
│   ├── go.sum               # Go 依赖校验
│   └── Dockerfile           # Docker 镜像
├── telegram-web/            # Web 客户端
│   ├── src/                 # 源代码
│   │   ├── im/              # IM 适配层
│   │   │   ├── adapter/     # 适配器
│   │   │   ├── debug/       # 调试工具
│   │   │   └── types/       # 类型定义
│   │   └── components/      # 组件
│   ├── public/              # 静态资源
│   ├── package.json         # 依赖配置
│   └── Dockerfile           # Docker 镜像
├── telegram-android/        # Android 客户端
│   └── TMessagesProj_App/   # Android 项目
│       ├── src/main/java/   # Java 源代码
│       │   └── org/telegram/im/adapter/  # IM 适配层
│       ├── src/main/assets/ # 资源文件
│       └── build.gradle     # 构建配置
├── im-admin/                # 管理后台
│   ├── src/                 # 源代码
│   ├── public/              # 静态资源
│   ├── package.json         # 依赖配置
│   └── Dockerfile           # Docker 镜像
├── scripts/                 # 脚本目录
│   ├── deploy/              # 部署脚本
│   ├── testing/             # 测试脚本
│   ├── ssl/                 # SSL 证书
│   └── nginx/               # Nginx 配置
├── docs/                    # 文档目录
│   ├── technical/           # 技术文档
│   ├── api/                 # API 文档
│   ├── deployment/          # 部署文档
│   └── user/                # 用户文档
├── k8s/                     # Kubernetes 配置
├── docker-compose.yml       # Docker Compose 配置
├── docker-compose.dev.yml   # 开发环境配置
├── docker-compose.prod.yml  # 生产环境配置
└── README.md                # 项目说明
```

### 后端结构详解

```
im-backend/
├── cmd/
│   └── main.go              # 应用入口点
├── internal/
│   ├── config/
│   │   ├── config.go        # 配置结构
│   │   ├── database.go      # 数据库配置
│   │   └── redis.go         # Redis 配置
│   ├── controller/
│   │   ├── auth.go          # 认证控制器
│   │   ├── user.go          # 用户控制器
│   │   ├── chat.go          # 聊天控制器
│   │   └── message.go       # 消息控制器
│   ├── service/
│   │   ├── auth_service.go  # 认证服务
│   │   ├── user_service.go  # 用户服务
│   │   ├── chat_service.go  # 聊天服务
│   │   └── message_service.go # 消息服务
│   ├── model/
│   │   ├── user.go          # 用户模型
│   │   ├── chat.go          # 聊天模型
│   │   └── message.go       # 消息模型
│   ├── middleware/
│   │   ├── auth.go          # 认证中间件
│   │   ├── cors.go          # CORS 中间件
│   │   ├── logger.go        # 日志中间件
│   │   └── recovery.go      # 恢复中间件
│   └── utils/
│       ├── jwt.go           # JWT 工具
│       ├── crypto.go        # 加密工具
│       └── validator.go     # 验证工具
├── pkg/
│   ├── database/
│   │   └── mysql.go         # MySQL 连接
│   ├── redis/
│   │   └── redis.go         # Redis 连接
│   └── minio/
│       └── minio.go         # MinIO 连接
└── tests/
    ├── unit/                # 单元测试
    ├── integration/         # 集成测试
    └── e2e/                 # 端到端测试
```

## 🔄 开发流程

### 1. 获取代码

```bash
# 克隆仓库
git clone https://github.com/your-org/zhihang-messenger.git
cd zhihang-messenger

# 初始化子模块
git submodule update --init --recursive
```

### 2. 启动开发环境

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yml ps

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f backend
```

### 3. 后端开发

```bash
# 进入后端目录
cd im-backend

# 安装依赖
go mod download

# 运行开发服务器
go run cmd/main.go

# 运行测试
go test ./...

# 运行特定测试
go test ./internal/service -v

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. 前端开发

```bash
# Web 客户端开发
cd telegram-web
npm install
npm run dev

# 管理后台开发
cd im-admin
npm install
npm run dev
```

### 5. Android 开发

```bash
# Android 开发
cd telegram-android/TMessagesProj_App
./gradlew assembleDebug
./gradlew installDebug
```

## 📝 代码规范

### Go 代码规范

#### 1. 命名规范

```go
// 包名：小写，简短，有意义
package user

// 接口名：以 -er 结尾
type Reader interface {
    Read([]byte) (int, error)
}

// 结构体名：大写开头，驼峰命名
type UserService struct {
    db *gorm.DB
}

// 方法名：大写开头，驼峰命名
func (s *UserService) GetUser(id uint) (*User, error) {
    // 实现
}

// 变量名：小写开头，驼峰命名
var userName string
var userCount int

// 常量名：大写，下划线分隔
const MAX_USER_COUNT = 1000
const DEFAULT_TIMEOUT = 30 * time.Second
```

#### 2. 注释规范

```go
// Package user 提供用户相关的业务逻辑
package user

// User 表示系统中的用户
type User struct {
    ID       uint   `json:"id"`       // 用户ID
    Username string `json:"username"` // 用户名
    Email    string `json:"email"`    // 邮箱
}

// GetUser 根据ID获取用户信息
// 如果用户不存在，返回错误
func (s *UserService) GetUser(id uint) (*User, error) {
    // 实现
}
```

#### 3. 错误处理

```go
// 使用 errors.New 创建简单错误
if user == nil {
    return nil, errors.New("用户不存在")
}

// 使用 fmt.Errorf 创建格式化错误
if err != nil {
    return fmt.Errorf("获取用户失败: %w", err)
}

// 自定义错误类型
type UserNotFoundError struct {
    UserID uint
}

func (e *UserNotFoundError) Error() string {
    return fmt.Sprintf("用户不存在: ID=%d", e.UserID)
}
```

#### 4. 日志规范

```go
import "github.com/sirupsen/logrus"

// 使用结构化日志
log.WithFields(logrus.Fields{
    "user_id": userID,
    "action":  "get_user",
}).Info("获取用户信息")

// 错误日志
log.WithFields(logrus.Fields{
    "error": err.Error(),
    "user_id": userID,
}).Error("获取用户失败")
```

### TypeScript 代码规范

#### 1. 命名规范

```typescript
// 接口名：大写开头，驼峰命名
interface UserInfo {
    id: number;
    username: string;
    email: string;
}

// 类名：大写开头，驼峰命名
class UserService {
    // 方法名：小写开头，驼峰命名
    async getUser(id: number): Promise<UserInfo> {
        // 实现
    }
}

// 变量名：小写开头，驼峰命名
const userName = 'testuser';
const userCount = 100;

// 常量名：大写，下划线分隔
const MAX_USER_COUNT = 1000;
const DEFAULT_TIMEOUT = 30000;
```

#### 2. 类型定义

```typescript
// 使用 interface 定义对象类型
interface LoginRequest {
    phone: string;
    password: string;
}

interface LoginResponse {
    user: UserInfo;
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
}

// 使用 type 定义联合类型
type MessageType = 'text' | 'image' | 'video' | 'audio' | 'file';

// 使用 enum 定义枚举
enum UserStatus {
    ACTIVE = 'active',
    INACTIVE = 'inactive',
    BANNED = 'banned'
}
```

#### 3. 错误处理

```typescript
// 使用 try-catch 处理异步错误
try {
    const user = await userService.getUser(id);
    console.log('用户信息:', user);
} catch (error) {
    console.error('获取用户失败:', error);
    // 处理错误
}

// 自定义错误类
class ApiError extends Error {
    constructor(
        public code: number,
        public message: string,
        public details?: any
    ) {
        super(message);
        this.name = 'ApiError';
    }
}
```

## 🧪 测试指南

### 单元测试

#### Go 单元测试

```go
// user_service_test.go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
    suite.Suite
    service *UserService
}

func (suite *UserServiceTestSuite) SetupTest() {
    // 测试前准备
    suite.service = NewUserService()
}

func (suite *UserServiceTestSuite) TestGetUser() {
    // 测试用例
    user, err := suite.service.GetUser(1)
    
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), user)
    assert.Equal(suite.T(), uint(1), user.ID)
}

func TestUserServiceSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}
```

#### TypeScript 单元测试

```typescript
// user.service.test.ts
import { UserService } from './user.service';
import { describe, it, expect, beforeEach } from '@jest/globals';

describe('UserService', () => {
    let userService: UserService;

    beforeEach(() => {
        userService = new UserService();
    });

    it('should get user by id', async () => {
        const user = await userService.getUser(1);
        
        expect(user).toBeDefined();
        expect(user.id).toBe(1);
    });
});
```

### 集成测试

```go
// integration_test.go
package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserLogin(t *testing.T) {
    // 启动测试服务器
    server := startTestServer()
    defer server.Close()

    // 发送登录请求
    resp, err := http.Post(server.URL+"/auth/login", "application/json", 
        strings.NewReader(`{"phone":"13800138000","password":"123456"}`))
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### 端到端测试

```typescript
// e2e/user.spec.ts
import { test, expect } from '@playwright/test';

test('用户登录流程', async ({ page }) => {
    // 访问登录页面
    await page.goto('/login');
    
    // 输入用户名和密码
    await page.fill('#username', 'testuser');
    await page.fill('#password', 'password123');
    
    // 点击登录按钮
    await page.click('#login-button');
    
    // 验证登录成功
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('.user-info')).toContainText('testuser');
});
```

## 🐛 调试技巧

### 后端调试

#### 1. 使用日志调试

```go
import "github.com/sirupsen/logrus"

func (s *UserService) GetUser(id uint) (*User, error) {
    log.WithField("user_id", id).Debug("开始获取用户信息")
    
    user, err := s.db.First(&User{}, id).Error
    if err != nil {
        log.WithFields(logrus.Fields{
            "user_id": id,
            "error": err.Error(),
        }).Error("获取用户失败")
        return nil, err
    }
    
    log.WithField("user_id", id).Info("用户信息获取成功")
    return user, nil
}
```

#### 2. 使用调试器

```bash
# 安装调试器
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug cmd/main.go

# 设置断点
(dlv) break internal/service/user_service.go:25

# 运行程序
(dlv) continue

# 查看变量
(dlv) print user
```

### 前端调试

#### 1. 使用浏览器开发者工具

```typescript
// 在代码中添加调试信息
console.log('用户信息:', user);
console.table(users); // 以表格形式显示数组
console.group('用户操作'); // 分组显示日志
console.log('登录成功');
console.log('获取用户信息');
console.groupEnd();
```

#### 2. 使用 VS Code 调试

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "type": "node",
            "request": "launch",
            "name": "调试前端",
            "program": "${workspaceFolder}/telegram-web/src/index.ts",
            "env": {
                "NODE_ENV": "development"
            }
        }
    ]
}
```

### 数据库调试

#### 1. 查看 SQL 查询

```go
// 启用 GORM 的 SQL 日志
db.Debug().First(&user, id)

// 或者设置全局日志级别
db.Logger = logger.Default.LogMode(logger.Info)
```

#### 2. 使用数据库客户端

```bash
# 连接 MySQL
mysql -h localhost -P 3306 -u zhihang_messenger -p zhihang_messenger

# 连接 Redis
redis-cli -h localhost -p 6379
```

## ❓ 常见问题

### 1. 依赖问题

**问题**: Go 模块下载失败

**解决方案**:
```bash
# 设置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct

# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download
```

**问题**: npm 安装失败

**解决方案**:
```bash
# 使用国内镜像
npm config set registry https://registry.npmmirror.com

# 清理缓存
npm cache clean --force

# 删除 node_modules 重新安装
rm -rf node_modules package-lock.json
npm install
```

### 2. 数据库连接问题

**问题**: MySQL 连接失败

**解决方案**:
```bash
# 检查 MySQL 服务状态
docker-compose ps mysql

# 查看 MySQL 日志
docker-compose logs mysql

# 重启 MySQL 服务
docker-compose restart mysql
```

**问题**: Redis 连接失败

**解决方案**:
```bash
# 检查 Redis 服务状态
docker-compose ps redis

# 测试 Redis 连接
docker-compose exec redis redis-cli ping
```

### 3. 端口冲突问题

**问题**: 端口被占用

**解决方案**:
```bash
# 查看端口占用
netstat -tulpn | grep :8080

# 杀死占用端口的进程
sudo kill -9 <PID>

# 或者修改配置文件中的端口
```

### 4. 权限问题

**问题**: Docker 权限不足

**解决方案**:
```bash
# 将用户添加到 docker 组
sudo usermod -aG docker $USER

# 重新登录或重启
sudo systemctl restart docker
```

### 5. 内存不足问题

**问题**: 开发环境内存不足

**解决方案**:
```bash
# 限制 Docker 容器内存使用
docker-compose -f docker-compose.dev.yml up -d --scale backend=1

# 或者修改 docker-compose.dev.yml 添加内存限制
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 512M
```

## 📚 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [React 官方文档](https://reactjs.org/docs/)
- [TypeScript 官方文档](https://www.typescriptlang.org/docs/)
- [Vue3 官方文档](https://vuejs.org/guide/)
- [Docker 官方文档](https://docs.docker.com/)
- [VS Code 官方文档](https://code.visualstudio.com/docs)
