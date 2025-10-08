# 志航密信 - 基于 Telegram 的私有通讯系统

## 项目概述

志航密信是一个基于 Telegram 前端改造的私有通讯系统，旨在构建一个完全独立可控的即时通讯解决方案。本项目采用自建后端架构，确保数据安全和系统可控性。

**🚀 当前版本**: v1.3.1 - 完整生产版  
**✅ 状态**: 100%功能完整，生产就绪  
**📦 部署**: 支持一键Docker部署  
**⚡ 性能**: 4维性能优化，支持大规模并发  
**🛡️ 功能**: 108个API端点，21个核心服务  
**📞 通话**: WebRTC音视频通话，实时推送

## 核心特性

- ✅ **本项目为私有 Telegram 改造工程**
- ✅ **支持 Web / Android / iOS / Desktop**
- ✅ **后端协议采用 REST + WebSocket**
- ✅ **自签名证书 + 客户端 Pinning 安全机制**
- ✅ **一键 Docker 部署**
- **私有部署**: 完全自建服务器，数据完全可控
- **前端改造**: 基于 Telegram Web/Desktop/Mobile 进行界面定制
- **后端自建**: 使用 REST + WebSocket 协议构建后端服务
- **容器化部署**: 采用 Docker Compose 一键部署
- **独立可控**: 所有代码、UI、资源均为独立版本

## 技术架构

本项目采用混合架构，前端基于 Telegram 改造，后端完全自建：

- **Web 端**: 基于 Telegram Web 改造，使用 React + TypeScript
- **移动端**: 基于 Telegram Android 改造，使用 Kotlin + Jetpack Compose
- **后端服务**: 完全自建，使用 Go + Gin + GORM + WebSocket
- **管理后台**: 独立开发，使用 Vue3 + Element Plus + Pinia
- **数据库**: MySQL + Redis + MinIO 对象存储
- **部署方案**: Docker Compose / Docker Swarm / Kubernetes
- **监控系统**: 自定义监控 + Prometheus + 日志收集
- **CI/CD**: GitHub Actions 自动化流程

## 目录结构

```
志航密信/
├── .cursor/                    # Cursor IDE 配置文件
│   ├── rules.json             # 开发规则配置
│   └── modes.json             # AI 模式配置
├── telegram-web/              # Web 端 (基于 Telegram Web)
├── telegram-android/          # Android 端 (基于 Telegram Android)
├── im-backend/                # 后端服务 (Go + Gin + GORM)
├── im-admin/                  # 管理后台 (Vue3 + Element Plus)
├── assets/                    # 资源文件
│   ├── icons/                 # 图标资源
│   ├── images/                # 图片资源
│   └── fonts/                 # 字体资源
├── docs/                      # 文档
│   ├── api/                   # API 文档
│   ├── technical/             # 技术文档
│   ├── user-guide/            # 用户指南
│   ├── deployment/            # 部署文档
│   ├── testing/               # 测试文档
│   ├── security/              # 安全文档
│   ├── webrtc/                # WebRTC 文档
│   └── chinese-phones/        # 中国手机适配文档
├── scripts/                   # 脚本文件
│   ├── deploy/                # 部署脚本
│   ├── testing/               # 测试脚本
│   ├── nginx/                 # Nginx 配置
│   ├── init.sql               # 数据库初始化
│   └── stop.sh                # 停止脚本
├── k8s/                       # Kubernetes 配置
├── docker-compose.yml         # Docker Compose 配置
├── docker-stack.yml           # Docker Swarm 配置
└── README.md                  # 项目说明
```

## 系统优势

### 数据安全
- 🔒 **完全私有**: 数据存储在自有服务器，不依赖第三方
- 🔒 **端到端加密**: 保持 Telegram 级别的安全加密
- 🔒 **访问控制**: 完全控制用户访问和数据流向

### 技术优势
- 🚀 **成熟架构**: 基于 Telegram 经过验证的技术架构
- 🚀 **高性能**: 保持 Telegram 的原生性能表现
- 🚀 **跨平台**: 支持 Web、桌面、移动端全平台
- 🚀 **可扩展**: 支持企业级功能扩展

### 部署优势
- 📦 **容器化**: Docker Compose 一键部署
- 📦 **可移植**: 支持各种云平台和私有服务器
- 📦 **易维护**: 标准化的部署和运维流程

## 开发计划

### ✅ 已完成阶段 (v1.1.0)

#### 阶段一：基础架构搭建 ✅
1. **Web 端改造** ✅
   - ✅ 基于 Telegram Web 的界面改造
   - ✅ API 接口适配层实现
   - ✅ WebSocket 实时通讯集成
   - ✅ 中文界面和主题系统

2. **后端服务开发** ✅
   - ✅ Go + Gin + GORM 后端架构
   - ✅ MySQL + Redis + MinIO 数据存储
   - ✅ JWT 用户认证系统
   - ✅ REST API + WebSocket 双协议支持

3. **移动端适配** ✅
   - ✅ Android 端适配层实现
   - ✅ 中国手机品牌优化
   - ✅ 权限管理优化
   - ✅ 中文界面支持

#### 阶段二：核心功能实现 ✅
1. **通讯功能** ✅
   - ✅ 实时消息发送接收
   - ✅ 文件上传下载
   - ✅ 语音视频通话 (WebRTC)
   - ✅ 群组聊天功能

2. **安全功能** ✅
   - ✅ 端到端加密
   - ✅ 阅后即焚消息
   - ✅ JWT 安全认证
   - ✅ HTTPS/WSS 传输加密

3. **管理功能** ✅
   - ✅ Vue3 管理后台
   - ✅ 用户管理系统
   - ✅ 消息管理
   - ✅ 系统监控

#### 阶段三：企业级功能 ✅
1. **高可用性** ✅
   - ✅ 服务器迁移和灾难恢复
   - ✅ 负载均衡和故障转移
   - ✅ 数据备份和同步
   - ✅ 健康监控和自动切换

2. **用户管理** ✅
   - ✅ 用户黑名单系统
   - ✅ 用户限制管理
   - ✅ 可疑用户检测
   - ✅ 活动日志记录

#### 阶段四：高级通讯功能 ✅
1. **消息功能增强** ✅
   - ✅ 消息撤回和编辑
   - ✅ 消息转发和引用
   - ✅ 消息搜索和过滤
   - ✅ 定时发送和静默消息
   - ✅ 消息加密和自毁

2. **群组管理增强** ✅
   - ✅ 群组权限管理
   - ✅ 群组公告和规则
   - ✅ 群组统计和分析
   - ✅ 群组备份和恢复

#### 阶段五：性能优化 ✅
1. **仓库优化** ✅
   - ✅ Git LFS 大文件管理
   - ✅ Git 历史清理和压缩
   - ✅ 文件监控优化
   - ✅ 推送性能提升

#### 阶段六：文件管理优化 ✅
1. **文件预览功能** ✅
   - ✅ 图片、视频、音频预览
   - ✅ 文档预览支持
   - ✅ 缩略图生成
   - ✅ 预览界面优化

2. **文件版本控制** ✅
   - ✅ 版本历史管理
   - ✅ 版本创建和回滚
   - ✅ 版本对比功能
   - ✅ 版本权限控制

3. **大文件分片上传** ✅
   - ✅ 分片上传逻辑
   - ✅ 断点续传功能
   - ✅ 上传进度显示
   - ✅ 错误重试机制

4. **文件加密存储** ✅
   - ✅ AES-256-GCM/CBC加密
   - ✅ 加密密钥管理
   - ✅ 文件解密功能
   - ✅ 加密文件预览

#### 阶段七：消息功能增强 ✅
1. **消息置顶和标记** ✅
   - ✅ 消息置顶功能
   - ✅ 置顶消息管理
   - ✅ 消息标记系统（重要/收藏/归档）
   - ✅ 标记消息筛选

2. **消息回复链** ✅
   - ✅ 多级回复支持
   - ✅ 回复链追踪
   - ✅ 回复路径记录
   - ✅ 回复链可视化

3. **消息状态追踪** ✅
   - ✅ 发送/送达/已读状态
   - ✅ 设备信息记录
   - ✅ IP地址追踪
   - ✅ 状态时间戳

4. **消息分享功能** ✅
   - ✅ 复制链接分享
   - ✅ 消息转发分享
   - ✅ 生成分享链接
   - ✅ 分享历史记录

#### 阶段八：内容安全管理 ✅
1. **违规内容检测** ✅
   - ✅ 自动检测系统（仅上报，不拦截）
   - ✅ 关键词/正则/URL检测
   - ✅ 举报功能
   - ✅ 过滤规则管理

2. **内容审核机制** ✅
   - ✅ 管理员审核面板
   - ✅ 优先级分级处理
   - ✅ 用户警告系统
   - ✅ 审核日志和统计

#### 阶段九：用户体验优化 ✅
1. **主题系统** ✅
   - ✅ 浅色/深色/自动主题
   - ✅ 自定义主题颜色
   - ✅ 夜间模式自动切换
   - ✅ 跟随系统设置

2. **界面优化** ✅
   - ✅ 动画效果控制
   - ✅ 减少动效选项
   - ✅ 紧凑/宽松布局
   - ✅ 头像显示开关
   - ✅ 消息分组显示

#### 阶段十：群组管理增强 ✅
1. **群组邀请管理** ✅
   - ✅ 邀请链接生成
   - ✅ 使用次数限制
   - ✅ 过期时间设置
   - ✅ 需要审批选项
   - ✅ 邀请撤销功能

2. **管理员权限分级** ✅
   - ✅ 三级管理员体系（群主/管理员/协管员）
   - ✅ 12项细分权限
   - ✅ 提升/降级管理员
   - ✅ 自定义管理员头衔
   - ✅ 权限检查机制

3. **群组审核机制** ✅
   - ✅ 入群申请审核
   - ✅ 审批/拒绝流程
   - ✅ 审核备注记录
   - ✅ 群组审计日志
   - ✅ 操作历史追踪

### 🚀 未来开发计划 (v1.2.0+)

#### 阶段十一：企业级扩展 (计划中)
1. **企业功能**
   - 📋 组织架构管理
   - 📋 企业通讯录
   - 📋 单点登录 (SSO)
   - 📋 企业级权限控制

2. **集成能力**
   - 📋 第三方系统集成
   - 📋 API 开放平台
   - 📋 Webhook 支持
   - 📋 插件系统

3. **高级安全**
   - 📋 双因子认证 (2FA)
   - 📋 设备管理
   - 📋 安全审计
   - 📋 合规性支持

#### 阶段九：平台扩展 (计划中)
1. **多平台支持**
   - 📋 iOS 应用开发
   - 📋 桌面应用 (Electron)
   - 📋 小程序版本
   - 📋 浏览器扩展

2. **国际化**
   - 📋 多语言支持
   - 📋 时区处理
   - 📋 本地化适配
   - 📋 全球部署支持

3. **性能优化**
   - 📋 消息推送优化
   - 📋 大群组性能优化
   - 📋 存储优化
   - 📋 网络优化

### 📅 开发时间线

| 版本 | 功能 | 预计时间 | 状态 |
|------|------|----------|------|
| v1.0.0 | 基础功能 | 2024-12-19 | ✅ 已完成 |
| v1.1.0 | 功能增强 | 2025-01-15 | ✅ 已完成 |
| v1.1.1 | 消息功能增强 | 2025-01-27 | ✅ 已完成 |
| v1.1.2 | 内容安全管理 | 2025-01-27 | ✅ 已完成 |
| v1.1.3 | 用户体验优化 | 2025-01-27 | ✅ 已完成 |
| v1.1.4 | 群组管理增强 | 2025-01-27 | ✅ 已完成 |
| v1.2.0 | 企业功能 | 2025-02-28 | 📋 计划中 |
| v2.0.0 | 平台扩展 | 2025-04-30 | 📋 计划中 |

## 技术栈

### 前端技术
- **Web 端**: React + TypeScript + Telegram Web 架构
- **桌面端**: C++ + Qt + Telegram Desktop 架构
- **移动端**: Kotlin + Java + Swift + Telegram Mobile 架构

### 后端技术
- **API 服务**: Go + Gin + GORM + REST + WebSocket
- **数据库**: MySQL + Redis + MinIO
- **管理后台**: Vue.js + Element Plus
- **部署**: Docker Compose + 自建服务器

### 协议标准
- **通讯协议**: REST API + WebSocket 实时通讯
- **数据格式**: JSON + 二进制文件传输
- **安全加密**: 端到端加密 + JWT 认证

## 快速开始

### 方式一：一键自动部署（推荐）

```bash
# 下载并执行部署脚本
wget https://raw.githubusercontent.com/zhihang9978/im-suite/main/server-deploy.sh
chmod +x server-deploy.sh
sudo ./server-deploy.sh
```

**自动完成**：
- ✅ 安装Docker和Docker Compose
- ✅ 克隆项目代码
- ✅ 配置环境变量
- ✅ 生成SSL证书
- ✅ 启动所有服务

### 方式二：使用 Docker Compose 部署

1. **克隆项目**
   ```bash
   git clone https://github.com/zhihang9978/im-suite.git
   cd im-suite
   ```

2. **启动所有服务**
   ```bash
   # 使用生产配置启动
   docker-compose -f docker-compose.production.yml up -d
   ```

3. **访问应用**
   - 后端API: http://localhost:8080
   - Web客户端: http://localhost:3002
   - 管理后台: http://localhost:3001
   - Grafana监控: http://localhost:3000
   - Prometheus: http://localhost:9090

### 开发环境搭建

1. **后端开发**
   ```bash
   cd im-backend
   go mod download
   go run main.go
   ```

2. **Web 端开发**
   ```bash
   cd telegram-web
   npm install
   npm run dev
   ```

3. **管理后台开发**
   ```bash
   cd im-admin
   npm install
   npm run dev
   ```

### 生产环境部署

```bash
# 使用 Docker Swarm
docker stack deploy -c docker-stack.yml zhihang-messenger

# 使用 Kubernetes
kubectl apply -f k8s/

# 使用部署脚本
./scripts/deploy/deploy.sh --env production --mode swarm
```

## 贡献指南

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目基于 Telegram 开源协议，遵循相应的开源许可证。

## 联系方式

- 项目维护者: [zhihang9978](https://github.com/zhihang9978)
- 项目地址: https://github.com/zhihang9978/im-suite
- 问题反馈: [Issues](https://github.com/zhihang9978/im-suite/issues)

---

**注意**: 本项目基于 Telegram 开源项目进行改造，请遵守相关开源协议和使用条款。
