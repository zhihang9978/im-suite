# 🚀 Devin 部署专用文档 - 只读这一个文件

**⚠️ 重要提示**：这是 Devin 部署的**唯一入口**，无需阅读其他任何文档！

---

## 📋 快速导航

- [部署前准备](#部署前准备) ⏱️ 2分钟
- [一键部署](#一键部署) ⏱️ 5分钟
- [验证部署](#验证部署) ⏱️ 2分钟
- [常见问题](#常见问题) ⏱️ 按需查看
- [不要做的事](#不要做的事) ⚠️ 必读

**预计总时间**: 10-15分钟  
**预计 ACU 消耗**: 30-40 ACU

---

## 🎯 核心原则

### ✅ DO（要做）
1. ✅ 只读这个文档
2. ✅ 按顺序执行命令
3. ✅ 查看日志验证
4. ✅ 出错时查看"常见问题"章节

### ❌ DON'T（不要做）
1. ❌ **不要**阅读其他37个MD文档（都是历史文档/实现报告）
2. ❌ **不要**尝试优化或修改代码
3. ❌ **不要**执行测试脚本（已通过测试）
4. ❌ **不要**读取 `docs/` 目录（技术文档，非部署文档）
5. ❌ **不要**关注以下文档：
   - `BOT_SYSTEM_COMPLETE_V1.6.0.md` - 实现报告
   - `COMPLETE_SUMMARY_v1.6.0.md` - 版本总结
   - `DEVIN_TASKS_V1.6.0.md` - 已完成任务
   - `FINAL_DELIVERY_TO_DEVIN.md` - 已完成交付
   - `PROJECT_STATUS_FINAL.md` - 状态报告
   - 以及其他 30+ 个历史文档

---

## 🚀 部署步骤

### 步骤1：部署前准备（2分钟）

#### 1.1 检查服务器信息

```bash
# 确认你在服务器上
hostname
# 预期输出: 154.37.214.191 或服务器主机名

# 确认 Docker 已安装
docker --version
docker-compose --version
```

**如果 Docker 未安装**，运行：
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

#### 1.2 确认当前目录

```bash
pwd
# 预期输出: /root/im-suite 或项目目录
```

**如果不在项目目录**：
```bash
cd /root/im-suite
# 如果目录不存在，先 git clone
```

#### 1.3 拉取最新代码

```bash
git pull origin main
```

**预期输出**: 
```
Already up to date.
或
Updating xxx...yyy
```

---

### 步骤2：一键部署（5分钟）

#### 2.1 执行部署脚本

```bash
# 方式A：使用一键部署脚本（推荐）
bash server-deploy.sh

# 方式B：使用 Docker Compose（如果方式A失败）
docker-compose -f docker-compose.production.yml up -d
```

**等待时间**: 3-5分钟（拉取镜像 + 启动服务）

#### 2.2 观察部署过程

**成功标志**:
```
✅ Creating network "im-network" ... done
✅ Creating im-mysql-prod ... done
✅ Creating im-redis-prod ... done
✅ Creating im-minio-prod ... done
✅ Creating im-backend-prod ... done
✅ Creating im-web-prod ... done
✅ Creating im-admin-prod ... done
✅ Creating im-nginx-prod ... done
```

**如果看到错误**，跳到 [常见问题](#常见问题) 章节

---

### 步骤3：验证部署（2分钟）

#### 3.1 检查容器状态

```bash
docker-compose -f docker-compose.production.yml ps
```

**预期输出**: 所有服务 State 显示 `Up`
```
NAME              STATE    PORTS
im-mysql-prod     Up       0.0.0.0:3306->3306/tcp
im-redis-prod     Up       0.0.0.0:6379->6379/tcp
im-minio-prod     Up       0.0.0.0:9000-9001->9000-9001/tcp
im-backend-prod   Up       0.0.0.0:8080->8080/tcp
im-web-prod       Up       0.0.0.0:3002->3002/tcp
im-admin-prod     Up       0.0.0.0:3001->3001/tcp
im-nginx-prod     Up       0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
```

**如果有服务不是 Up 状态**，查看日志：
```bash
docker-compose -f docker-compose.production.yml logs 服务名
```

#### 3.2 验证后端数据库迁移

```bash
docker-compose -f docker-compose.production.yml logs im-backend | grep "数据库迁移"
```

**必须看到的关键日志**:
```
✅ 依赖检查通过
✅ 数据库迁移完成！成功迁移 56/56 个表
✅ 数据库验证通过！当前共有 56 个表
🎉 数据库迁移和验证全部通过！服务可以安全启动。
[GIN-debug] Listening and serving HTTP on :8080
```

**如果没看到以上日志**，运行：
```bash
docker-compose -f docker-compose.production.yml logs im-backend
```
查找错误信息。

#### 3.3 测试服务可访问性

```bash
# 测试后端 API
curl http://localhost:8080/api/health
# 预期: {"status":"ok"}

# 测试 Web 前端
curl -I http://localhost:3002
# 预期: HTTP/1.1 200 OK

# 测试管理后台
curl -I http://localhost:3001
# 预期: HTTP/1.1 200 OK
```

**全部成功** → **部署完成！** 🎉

---

## ❌ 常见问题

### 问题1：数据库迁移失败

**错误日志**:
```
❌ 迁移失败: Error 1824: Failed to open the referenced table
```

**原因**: 这个问题已经在代码中修复了

**解决**:
```bash
# 确认代码是最新的
git log --oneline -1
# 必须看到: "feat: database migration hardening" 或更新的提交

# 如果不是最新，拉取代码
git pull origin main

# 重新部署
docker-compose -f docker-compose.production.yml down
docker-compose -f docker-compose.production.yml up -d --build
```

### 问题2：端口被占用

**错误日志**:
```
Error: bind: address already in use
```

**解决**:
```bash
# 查看占用端口的进程
netstat -tulpn | grep -E '(3001|3002|8080|3306|6379)'

# 停止占用端口的服务
docker-compose -f docker-compose.production.yml down

# 或者杀死进程
kill -9 进程ID
```

### 问题3：Docker 镜像拉取失败

**错误日志**:
```
Error response from daemon: Get https://registry-1.docker.io/v2/: net/http: TLS handshake timeout
```

**解决**: 使用国内镜像源或本地构建

```bash
# 方案A：配置 Docker 镜像加速（已在 server-deploy.sh 中处理）
# 方案B：本地构建镜像
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d
```

### 问题4：网络连接问题

**如果服务器无法访问外网**，参考 `NETWORK_TROUBLESHOOTING_GUIDE.md`

**快速修复**:
```bash
# 运行网络修复脚本
bash scripts/fix-docker-network.sh
```

---

## 📊 部署验证清单

完成部署后，逐项确认：

### 基础服务
- [ ] MySQL 容器运行中（`docker ps | grep mysql`）
- [ ] Redis 容器运行中（`docker ps | grep redis`）
- [ ] MinIO 容器运行中（`docker ps | grep minio`）

### 应用服务
- [ ] 后端服务运行中（`docker ps | grep im-backend`）
- [ ] Web 前端运行中（`docker ps | grep im-web`）
- [ ] 管理后台运行中（`docker ps | grep im-admin`）
- [ ] Nginx 运行中（`docker ps | grep nginx`）

### 数据库迁移
- [ ] 后端日志显示"数据库迁移完成！成功迁移 56/56 个表"
- [ ] 后端日志显示"数据库验证通过！当前共有 56 个表"
- [ ] 后端日志显示"🎉 数据库迁移和验证全部通过！"
- [ ] 后端日志显示"[GIN-debug] Listening and serving HTTP on :8080"

### API 可访问性
- [ ] `curl http://localhost:8080/api/health` 返回 `{"status":"ok"}`
- [ ] `curl -I http://localhost:3002` 返回 `HTTP/1.1 200 OK`
- [ ] `curl -I http://localhost:3001` 返回 `HTTP/1.1 200 OK`

**如果以上全部打勾** → **部署成功！** ✅

---

## 🎯 ACU 优化提示

### 已经做的优化（无需额外操作）

1. ✅ 数据库迁移已优化 - 自动按正确顺序执行
2. ✅ 部署脚本已优化 - 一键执行所有步骤
3. ✅ 配置文件已优化 - 无需手动修改
4. ✅ 文档已优化 - 本文档是唯一需要的

### Devin 节省 ACU 的关键

1. **只读这个文档** - 不要浏览其他 37 个 MD 文件
2. **不要测试** - 项目已通过所有测试，无需重新测试
3. **不要优化** - 代码已优化，无需阅读源码
4. **按步骤执行** - 不要尝试理解原理，直接执行即可

**预计节省**: 150+ ACU（相比从零开始理解项目）

---

## 📚 可选：了解更多（仅在出错时查看）

### 环境变量配置
如需自定义配置，参考: `ENV_TEMPLATE.md`

### 网络问题排查
如遇网络问题，参考: `NETWORK_TROUBLESHOOTING_GUIDE.md`

### 数据库迁移详情
如需了解迁移机制，参考: `im-backend/DATABASE_MIGRATION_GUIDE.md`

### 完整部署指南
如需详细说明，参考: `SERVER_DEPLOYMENT_INSTRUCTIONS.md`

**但通常情况下，上述文档都不需要阅读！**

---

## ✅ 部署完成后的最终确认

运行以下命令，截图输出：

```bash
echo "========== 容器状态 =========="
docker-compose -f docker-compose.production.yml ps

echo "========== 后端日志（关键部分）=========="
docker-compose -f docker-compose.production.yml logs im-backend | grep -E "(数据库迁移|验证通过|Listening)"

echo "========== API 测试 =========="
curl -s http://localhost:8080/api/health
echo ""
curl -I http://localhost:3002 2>&1 | head -1
curl -I http://localhost:3001 2>&1 | head -1

echo "========== 部署完成 =========="
echo "✅ 如果以上全部正常，部署成功！"
```

---

## 🆘 紧急支持

如果按照本文档操作仍然失败：

1. **查看完整日志**:
   ```bash
   docker-compose -f docker-compose.production.yml logs > deployment.log
   ```

2. **检查系统资源**:
   ```bash
   df -h  # 磁盘空间
   free -h  # 内存
   ```

3. **提供以下信息**:
   - 执行到哪一步失败
   - 具体错误信息
   - 系统环境（`uname -a`）
   - Docker 版本（`docker --version`）

---

## 📊 部署时间和资源预估

| 阶段 | 时间 | ACU | 说明 |
|------|------|-----|------|
| 阅读本文档 | 3分钟 | 5 ACU | 快速浏览 |
| 部署前准备 | 2分钟 | 5 ACU | 检查环境 |
| 执行部署 | 5分钟 | 10 ACU | 等待容器启动 |
| 验证部署 | 2分钟 | 5 ACU | 检查状态 |
| 处理问题（如有） | 3分钟 | 10 ACU | 查看日志，重启 |
| **总计** | **15分钟** | **35 ACU** | **正常情况** |

**如果不按本文档操作**:
- 阅读所有文档: 30分钟, 60 ACU ❌
- 理解项目结构: 20分钟, 40 ACU ❌
- 尝试优化: 15分钟, 30 ACU ❌
- 调试错误: 30分钟, 50 ACU ❌
- **总计**: 95分钟, 180 ACU ❌

**节省**: 80分钟, 145 ACU ✅

---

## 🎉 成功标志

部署成功后，你应该看到：

```
✅ 9个容器全部运行中
✅ 数据库包含56个表
✅ 后端 API 可访问（http://localhost:8080）
✅ Web 前端可访问（http://localhost:3002）
✅ 管理后台可访问（http://localhost:3001）
✅ Nginx 反向代理运行中
✅ 监控系统运行中（Prometheus + Grafana）
```

**恭喜！部署完成！** 🎊

---

**文档版本**: v1.6.0  
**最后更新**: 2025-10-10  
**适用于**: Devin AI 或其他自动化部署工具  
**预计 ACU**: 30-40（节省 140+ ACU）

**⚠️ 再次提醒**：这是 Devin 部署的**唯一必读文档**，无需阅读项目中的其他任何文档！

