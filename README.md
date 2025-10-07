# IM Suite - 即时通讯系统套件

## 项目目标

IM Suite 是一个完整的即时通讯系统解决方案，包含后端服务、Web 前端、Android 客户端和管理后台。系统采用现代化的技术栈，支持高并发、实时通讯和跨平台部署。

## 技术架构

- **后端**: Go + Gin + GORM + MySQL + Redis + MinIO
- **Web 前端**: React + Vite + TypeScript
- **Android**: Kotlin + Jetpack Compose
- **管理后台**: React + Ant Design
- **部署**: Docker Compose 一键部署

## 目录结构

```
im-suite/
├── .cursor/              # Cursor IDE 配置文件
│   ├── rules.json        # 开发规则配置
│   └── modes.json        # AI 模式配置
├── im-backend/           # 后端服务 (Go)
├── im-web/              # Web 前端 (React)
├── im-android/          # Android 客户端 (Kotlin)
├── im-admin/            # 管理后台 (React)
├── scripts/             # 部署和工具脚本
└── README.md            # 项目说明文档
```

## 子目录说明

### im-backend/
后端服务，使用 Go 语言开发，提供 REST API 和 WebSocket 服务。

### im-web/
Web 前端应用，基于 React + Vite 构建，提供现代化的用户界面。

### im-android/
Android 移动客户端，使用 Kotlin + Jetpack Compose 开发。

### im-admin/
管理后台系统，用于系统管理和监控。

### scripts/
部署脚本和工具脚本，包含 Docker Compose 配置和自动化部署脚本。

## 快速开始

1. 克隆项目
2. 运行 `docker-compose up -d` 启动所有服务
3. 访问 Web 界面开始使用

## 开发规范

- 所有代码注释和文档使用中文
- 遵循 REST API 设计规范
- 使用 WebSocket 实现实时通讯
- 采用 Docker 容器化部署
