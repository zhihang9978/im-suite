# 🚀 Devin 部署唯一入口

**这是唯一需要阅读的部署文档**

---

## ⚠️ 硬规则（第一优先级）

**在执行任何操作前，必须遵守以下规则**：

1. **唯一部署方式**: 仅使用 `docker-compose.production.yml` 进行生产部署
2. **网络/SSH异常处理**: 如遇网络异常或外部SSH失败，只输出将执行的脚本（print-only），**不要实际执行**
3. **严禁构建客户端**: 
   - ❌ 不要进入 `clients/telegram-web/` 目录
   - ❌ 不要进入 `clients/telegram-android/` 目录  
   - ❌ 不要构建/打包客户端（它们不属于生产部署）
   - ✅ 客户端已预先构建在Docker镜像中

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
