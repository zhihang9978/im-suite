# 🚀 给Devin：立即部署命令

**紧急修复**: GORM bug已修复  
**最新提交**: fac8ff7  
**状态**: ✅ 可以继续部署

---

## ⚡ 立即执行（复制粘贴）

```bash
# 在服务器 154.37.214.191 上执行

# 1. 进入项目目录
cd /root/im-suite

# 2. 拉取最新修复（包含GORM bug修复）
git pull origin main

# 3. 重建后端镜像（必须！）
docker-compose -f docker-compose.production.yml build backend

# 4. 停止旧服务
docker-compose -f docker-compose.production.yml down

# 5. 启动所有服务
docker-compose -f docker-compose.production.yml up -d

# 6. 等待服务启动（重要）
sleep 120

# 7. 验证后端迁移成功
docker logs im-backend-prod | grep "数据库迁移"

# 8. 验证健康检查
curl http://localhost:8080/health

# 9. 检查所有服务状态
docker-compose -f docker-compose.production.yml ps
```

---

## ✅ 预期结果

### 1. git pull 输出

```
Updating ab81f81..fac8ff7
Fast-forward
 im-backend/internal/model/user.go | ...
 im-backend/internal/model/bot.go | ...
 ... (7个model文件更新)
```

### 2. docker logs 输出

```
========================================
🚀 开始数据库表迁移...
========================================

✅ 依赖检查通过

⏳ [1/56] 迁移表: User
   ✅ 迁移成功: User
⏳ [2/56] 迁移表: Session
   ✅ 迁移成功: Session
... (所有表成功)

✅ 数据库迁移完成！成功迁移 56/56 个表

========================================
🎉 数据库迁移和验证全部通过！服务可以安全启动。
========================================
```

### 3. curl 输出

```json
{
  "status": "ok",
  "timestamp": 1728586800,
  "service": "zhihang-messenger-backend",
  "version": "1.4.0"
}
```

### 4. docker-compose ps 输出

```
NAME                STATUS              HEALTH
im-mysql-prod       running             healthy
im-redis-prod       running             healthy
im-minio-prod       running             healthy
im-backend-prod     running             healthy  ✅
im-admin-prod       running             healthy
```

---

## 🔧 修复内容

**GORM v1.30.0 Bug修复**:
- ✅ 修复了7个模型文件
- ✅ 修复了11个uniqueIndex字段
- ✅ 将`uniqueIndex`改为`index:idx_xxx,unique`
- ✅ 避开GORM bug，功能完全相同

**清理内容**:
- ✅ 归档6个老旧修复文档
- ✅ 删除1个重复环境模板
- ✅ 根目录更整洁

---

## 🎯 成功标志

**部署成功的3个标志**:

1. ✅ 数据库迁移日志显示 "56/56 个表" 成功
2. ✅ curl /health 返回 HTTP 200
3. ✅ 所有容器状态为 "healthy"

**如果看到这3个标志，说明部署完全成功！**

---

## ⚠️ 如果遇到问题

### 问题1: git pull失败

```bash
# 解决: 重置本地更改
git fetch origin
git reset --hard origin/main
```

### 问题2: 容器构建失败

```bash
# 解决: 清理缓存重建
docker system prune -f
docker-compose -f docker-compose.production.yml build --no-cache backend
```

### 问题3: 数据库连接失败

```bash
# 解决: 检查MySQL容器
docker logs im-mysql-prod
docker restart im-mysql-prod
```

---

## 📞 部署后验证

```bash
# 完整验证命令
echo "=== 1. 容器状态 ==="
docker-compose ps

echo "=== 2. 后端健康 ==="
curl http://localhost:8080/health

echo "=== 3. 管理后台 ==="
curl -I http://localhost:3001

echo "=== 4. 数据库迁移日志 ==="
docker logs im-backend-prod | grep "数据库迁移完成"

# 全部通过说明部署成功！
```

---

## 🎊 可以开始部署了！

**GORM bug已修复，老旧文件已清理，远程仓库100%最新！**

**执行上面的命令即可！** 🚀

