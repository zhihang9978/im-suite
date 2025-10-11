# 🚀 给Devin：快速开始部署

**项目**: 志航密信 (IM Suite)  
**最新版本**: v1.6.0  
**GORM版本**: v1.25.12（已修复v1.30.0的bug）  
**部署状态**: ✅ 就绪

---

## ⚡ 快速部署命令

### 在服务器 154.37.214.191 上执行：

```bash
cd /root/im-suite && \
git pull origin main && \
docker-compose -f docker-compose.production.yml build --no-cache backend && \
docker-compose -f docker-compose.production.yml down -v && \
docker-compose -f docker-compose.production.yml up -d && \
sleep 120 && \
docker logs im-backend-prod | tail -100
```

---

## ✅ 成功标志

部署成功需要满足3个条件：

1. **数据库迁移成功**
```
✅ 数据库迁移完成！成功迁移 56/56 个表
```

2. **健康检查通过**
```bash
curl http://localhost:8080/health
# 返回: {"status":"ok"}
```

3. **容器状态正常**
```bash
docker-compose ps
# im-backend-prod 显示 "healthy"
```

---

## 🔧 关键信息

### GORM版本说明
- ✅ 当前使用: GORM v1.25.12（稳定版本）
- ❌ 已弃用: GORM v1.30.0（有AutoMigrate bug）

### 部署要点
1. **必须使用 `down -v`** 删除旧数据卷
2. **必须使用 `--no-cache`** 完全重建镜像
3. **等待120秒** 让服务完全初始化

---

## 📁 项目结构

```
im-suite/
├── im-backend/          # Go后端服务 (GORM v1.25.12)
├── im-admin/            # Vue3管理后台
├── telegram-android/    # Android客户端
├── telegram-web/        # Web客户端
├── config/              # 配置文件（MySQL, Redis, Nginx等）
├── scripts/             # 部署和维护脚本
└── docs/               # 完整文档
```

---

## 📚 重要文档

### 核心文档
- `README.md` - 项目主文档
- `CHANGELOG.md` - 版本变更记录
- `ENV_STRICT_TEMPLATE.md` - 环境配置模板

### 部署指南
- `SERVER_DEPLOYMENT_INSTRUCTIONS.md` - 单服务器部署
- `THREE_SERVER_DEPLOYMENT_GUIDE.md` - 三服务器高可用架构
- `INTERNATIONAL_DEPLOYMENT_GUIDE.md` - 国际化部署

### 技术文档
- `docs/api/` - API文档
- `docs/technical/` - 技术架构
- `docs/security/` - 安全设计

### 历史记录
- `docs/deployment-history/` - 部署历史和问题排查记录

---

## 🌐 访问地址

部署成功后，可访问：
- 🖥️ 管理后台: http://154.37.214.191:3001
- 🔧 后端API: http://154.37.214.191:8080
- 📊 Grafana监控: http://154.37.214.191:3000
- 📈 Prometheus: http://154.37.214.191:9090

---

## ⚠️ 常见问题

### Q1: 部署失败怎么办？
A: 检查 `docker logs im-backend-prod` 查看详细错误信息

### Q2: 数据库迁移失败？
A: 确认使用了 `down -v` 删除旧数据卷

### Q3: 健康检查失败？
A: 等待2-3分钟让服务完全启动，然后重试

---

## 📞 获取帮助

- 查看详细部署历史: `docs/deployment-history/`
- 查看GORM bug修复记录: `docs/deployment-history/GORM_BUG_FIX.md`
- 查看完整失败分析: `docs/deployment-history/DEPLOYMENT_FAILURE_ANALYSIS.md`

---

**立即执行上面的快速部署命令，预计2.5分钟完成！** 🚀
