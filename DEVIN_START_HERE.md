# 🚀 Devin 部署唯一入口

**这是唯一需要阅读的部署文档**

---

## 🛡️ 三条硬约束（必须遵守）

### A) 只使用 docker-compose.production.yml

```bash
# ✅ 唯一正确命令
docker-compose -f docker-compose.production.yml up -d
```

### B) 网络异常时停止，输出错误

```bash
# 不要自动修复网络，输出错误并停止
```

### C) 不要构建客户端

```bash
# ❌ 不要进入 telegram-web 或 telegram-android 目录
# ✅ 客户端已预构建在Docker镜像中
```

---

## 📖 部署步骤

### 1. 环境检查
```bash
docker --version
docker-compose --version
test -f .env || echo "需要.env文件"
```

### 2. 启动服务
```bash
docker-compose -f docker-compose.production.yml up -d
sleep 120
```

### 3. 健康检查
```bash
docker-compose -f docker-compose.production.yml ps
docker logs im-backend-prod | grep "数据库迁移"
curl http://localhost:8080/health
```

---

## 📚 其他文档

- `ENV_TEMPLATE.md` - 环境变量配置
- `SERVER_DEPLOYMENT_INSTRUCTIONS.md` - 详细说明
- `NETWORK_TROUBLESHOOTING_GUIDE.md` - 故障排查

---

**预计时间**: 10分钟
