# 贡献指南

感谢您对志航密信项目的关注和贡献！我们欢迎各种形式的贡献，包括但不限于代码、文档、测试、反馈等。

## 📋 贡献方式

### 🐛 报告问题
如果您发现了 bug 或有功能建议，请通过以下方式报告：

1. **GitHub Issues**: 在项目 Issues 页面创建新的 issue
2. **问题分类**: 请选择合适的标签（bug、enhancement、question 等）
3. **详细描述**: 提供详细的问题描述、复现步骤和环境信息

### 🔧 代码贡献
如果您想贡献代码，请遵循以下流程：

1. **Fork 项目**: 点击 GitHub 上的 Fork 按钮
2. **创建分支**: 创建新的功能分支
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **提交代码**: 提交您的更改
   ```bash
   git commit -m "feat: 添加新功能描述"
   ```
4. **推送分支**: 推送到您的 fork
   ```bash
   git push origin feature/your-feature-name
   ```
5. **创建 PR**: 创建 Pull Request

### 📚 文档贡献
- 修正文档中的错误
- 添加新的文档内容
- 改进文档结构和可读性
- 翻译文档到其他语言

### 🧪 测试贡献
- 编写单元测试
- 编写集成测试
- 进行手动测试
- 报告测试结果

## 🛠️ 开发环境设置

### 环境要求
- Go 1.21+
- Node.js 18+
- Docker 20+
- MySQL 8.0
- Redis 7.0

### 快速开始
```bash
# 1. 克隆项目
git clone https://github.com/your-username/zhihang-messenger.git
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

## 📝 代码规范

### Go 代码规范
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

### JavaScript/TypeScript 代码规范
- 使用 ESLint 进行代码检查
- 遵循 Airbnb JavaScript 规范
- 使用 Prettier 格式化代码
- 添加类型注解（TypeScript）

```bash
# 代码检查
npm run lint

# 代码格式化
npm run format
```

### Vue.js 代码规范
- 遵循 Vue.js 官方风格指南
- 使用 Composition API
- 组件命名使用 PascalCase
- 文件命名使用 kebab-case

## 🧪 测试规范

### 单元测试
- 为所有公共函数编写单元测试
- 测试覆盖率应达到 80% 以上
- 使用有意义的测试名称

```bash
# 运行 Go 单元测试
cd im-backend
go test -v ./...

# 运行前端单元测试
cd telegram-web
npm run test
```

### 集成测试
- 测试 API 接口
- 测试数据库操作
- 测试 WebSocket 连接

```bash
# 运行集成测试
cd scripts/testing
python3 run_all_tests.py
```

## 📋 提交规范

### 提交消息格式
使用 Conventional Commits 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 类型说明
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 示例
```
feat(auth): 添加用户登录功能

- 实现 JWT 认证
- 添加密码加密
- 支持记住登录状态

Closes #123
```

## 🔍 代码审查

### Pull Request 要求
1. **功能完整**: 确保功能完整且经过测试
2. **代码质量**: 代码符合项目规范
3. **文档更新**: 相关文档已更新
4. **测试覆盖**: 添加必要的测试
5. **描述清晰**: PR 描述清晰，说明变更内容

### 审查要点
- 代码逻辑正确性
- 性能影响
- 安全性考虑
- 向后兼容性
- 文档完整性

## 🏷️ 版本发布

### 版本号规范
使用语义化版本号（Semantic Versioning）：
- `MAJOR`: 不兼容的 API 修改
- `MINOR`: 向下兼容的功能性新增
- `PATCH`: 向下兼容的问题修正

### 发布流程
1. 更新版本号
2. 更新 CHANGELOG.md
3. 创建 Git 标签
4. 发布到 GitHub Releases

## 📞 获取帮助

### 联系方式
- **GitHub Issues**: 技术问题和讨论
- **GitHub Discussions**: 一般性讨论
- **邮箱**: support@zhihang-messenger.com

### 社区资源
- **文档**: [项目文档](docs/)
- **API 文档**: [API 文档](docs/api/)
- **用户指南**: [用户指南](docs/user-guide/)

## 🙏 致谢

感谢所有为志航密信项目做出贡献的开发者！

### 贡献者名单
- [zhihang9978](https://github.com/zhihang9978) - 项目维护者

## 📄 许可证

本项目基于 MIT 许可证开源。贡献代码即表示您同意将代码以相同的许可证发布。

---

**再次感谢您的贡献！** 🎉
