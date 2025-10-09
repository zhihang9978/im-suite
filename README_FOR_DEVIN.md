# 给 Devin 的部署说明

**志航密信 v1.6.0** - 快速部署指南

---

## 👋 你好，Devin！

我已经为你准备好了所有代码和配置，你只需要按照下面的步骤执行即可完成部署和测试。

**预计时间：约1小时**（大部分是自动化的）

---

## 🚀 快速开始（3个命令）

```bash
# 1. 检查项目完整性（2分钟）
bash scripts/check-project-integrity.sh

# 2. 一键部署（50分钟，大部分是等待时间）
bash scripts/auto-deploy.sh

# 3. 自动测试（10分钟）
bash scripts/auto-test.sh
```

**就这么简单！** ✅

---

## 📋 详细说明

### 第一步：检查完整性

```bash
bash scripts/check-project-integrity.sh
```

这个脚本会检查：
- ✅ 所有必需的文件是否存在
- ✅ Go代码能否编译
- ✅ 配置文件是否齐全
- ✅ 目录结构是否正确

**预期输出：**
```
✅ 项目完整，可以开始部署！
完整性: 100%
```

如果有文件缺失，会明确告诉你哪些文件不存在。

---

### 第二步：一键部署

```bash
bash scripts/auto-deploy.sh
```

这个脚本会自动：
1. ✅ 检查Go、Docker等环境
2. ✅ 配置环境变量（.env）
3. ✅ 启动MySQL、Redis、MinIO
4. ✅ 编译Go后端
5. ✅ 运行数据库迁移
6. ✅ 启动后端服务
7. ✅ 执行健康检查
8. ✅ 验证数据库表

**预期输出：**
```
🎉 部署成功！

服务信息：
  后端API:    http://localhost:8080
  健康检查:   http://localhost:8080/health
  前端演示:   http://localhost:8000/examples/screen-share-demo.html
```

**如果部署失败**，脚本会显示具体的错误信息和建议的解决方案。

---

### 第三步：自动测试

```bash
bash scripts/auto-test.sh
```

这个脚本会自动测试：
1. ✅ 健康检查API
2. ✅ 用户注册和登录
3. ✅ 屏幕共享基础API（5个端点）
4. ✅ 屏幕共享增强API（10个端点）
5. ✅ WebRTC通话API
6. ✅ API响应时间

**预期输出：**
```
✅ 所有测试通过！

总测试数:  15
通过:      15
失败:      0
通过率:    100%
```

测试报告会自动生成在 `logs/test-report-YYYYMMDD_HHMMSS.txt`

---

## 📊 你需要做的事情

### 必须做（核心功能）

1. **运行3个脚本** ✅ （如上所述）
2. **检查测试报告** - 确保所有测试通过
3. **测试前端页面** - 打开 `examples/screen-share-demo.html`
4. **记录问题** - 如果有任何失败，记录详细信息

### 可选做（深度测试）

1. **Android权限测试** - 如果有Android环境
2. **压力测试** - 测试并发和性能
3. **长时间运行测试** - 测试稳定性
4. **浏览器兼容性测试** - 测试不同浏览器

---

## 🎯 预期结果

部署和测试完成后，你应该能：

1. ✅ 访问 http://localhost:8080/health 看到 `{"status":"ok"}`
2. ✅ 用户能注册和登录
3. ✅ 所有15个屏幕共享API正常工作
4. ✅ 前端演示页面能正常显示
5. ✅ 数据库有50+个表
6. ✅ 无错误日志

---

## 📁 重要文件位置

### 部署相关
- **部署脚本**: `scripts/auto-deploy.sh`
- **测试脚本**: `scripts/auto-test.sh`
- **检查脚本**: `scripts/check-project-integrity.sh`
- **部署文档**: `DEPLOYMENT_FOR_DEVIN.md`（详细版）

### 功能文档
- **总体报告**: `COMPLETE_SUMMARY_v1.6.0.md`（必读！）
- **屏幕共享**: `SCREEN_SHARE_ENHANCED.md`
- **权限系统**: `PERMISSION_SYSTEM_COMPLETE.md`
- **快速开始**: `SCREEN_SHARE_QUICK_START.md`

### 配置文件
- **环境变量**: `.env`（部署时自动创建）
- **数据库配置**: `config/mysql/`
- **Redis配置**: `config/redis/`

### 日志文件
- **后端日志**: `logs/backend.log`
- **测试报告**: `logs/test-report-*.txt`

---

## ⚠️ 可能遇到的问题

### 问题1：脚本没有执行权限

```bash
# 解决方案
chmod +x scripts/*.sh
```

### 问题2：Docker服务启动失败

```bash
# 查看具体错误
docker-compose -f docker-compose.production.yml logs

# 常见原因：端口被占用
# 解决：修改docker-compose.production.yml中的端口
```

### 问题3：Go编译失败

```bash
# 查看具体错误
cd im-backend
go build -v main.go

# 常见原因：依赖下载失败
# 解决：配置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
```

### 问题4：测试失败

```bash
# 查看详细日志
tail -f logs/backend.log

# 查看测试报告
cat logs/test-report-*.txt
```

---

## 📞 需要帮助？

### 遇到问题时的检查顺序

1. **查看脚本输出** - 错误信息通常很明确
2. **查看日志文件** - `logs/backend.log`
3. **查看详细文档** - `DEPLOYMENT_FOR_DEVIN.md`
4. **检查配置** - `.env` 文件
5. **检查服务状态** - `docker-compose ps`

### 关键命令

```bash
# 查看后端日志
tail -f logs/backend.log

# 查看Docker服务状态
docker-compose -f docker-compose.production.yml ps

# 查看Docker日志
docker-compose -f docker-compose.production.yml logs -f

# 重启服务
docker-compose -f docker-compose.production.yml restart

# 停止所有服务
docker-compose -f docker-compose.production.yml down
kill $(cat logs/backend.pid)
```

---

## ✅ 验收标准

部署和测试全部完成后，请确认：

- [ ] `check-project-integrity.sh` 显示 100% 完整性
- [ ] `auto-deploy.sh` 显示部署成功
- [ ] `auto-test.sh` 显示所有测试通过
- [ ] 能访问 http://localhost:8080/health
- [ ] 前端演示页面正常工作
- [ ] 无错误日志

**如果以上全部打勾，部署成功！** 🎉

---

## 💡 提示

### 节省时间的技巧

1. **并行操作** - Docker服务启动时，可以同时查看文档
2. **使用日志** - 出错时先看日志，不要盲目重试
3. **保存Token** - 登录后保存token，后续测试直接用

### 记录信息

测试过程中，请记录：
1. ✅ 通过的测试数量
2. ❌ 失败的测试和错误信息
3. ⚠️ 遇到的问题和解决方案
4. 📊 性能数据（如果有）
5. 💡 改进建议

---

## 🎉 完成后

部署和测试全部通过后，可以：

1. 📝 生成测试报告（自动生成在logs/目录）
2. 📸 截图记录成功界面
3. 📊 记录性能数据
4. 🎯 规划下一步工作

---

## 📌 快速参考

### 一行命令完整部署

```bash
bash scripts/check-project-integrity.sh && bash scripts/auto-deploy.sh && bash scripts/auto-test.sh
```

### 查看所有服务状态

```bash
# 后端
ps aux | grep im-backend

# Docker服务
docker-compose -f docker-compose.production.yml ps

# 端口占用
netstat -tulnp | grep -E "8080|3306|6379|9000"
```

### 快速重启

```bash
# 重启后端
kill $(cat logs/backend.pid)
cd im-backend && nohup ./bin/im-backend > ../logs/backend.log 2>&1 &
echo $! > ../logs/backend.pid

# 重启Docker服务
docker-compose -f docker-compose.production.yml restart
```

---

**准备好了吗？让我们开始吧！** 🚀

**第一步命令：**
```bash
bash scripts/check-project-integrity.sh
```

祝你部署顺利！ 💪



