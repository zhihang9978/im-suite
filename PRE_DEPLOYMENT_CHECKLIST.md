# 🚀 生产环境部署前检查清单

## 📋 部署前必检项目

### ✅ 环境配置检查

#### 1. 环境变量配置
- [ ] 已创建 `.env` 文件
- [ ] 所有必需环境变量已配置
- [ ] 数据库密码（至少16位强密码）
- [ ] Redis密码（至少16位强密码）
- [ ] MinIO密码（至少16位强密码）
- [ ] JWT_SECRET（至少32位随机字符）
- [ ] `.env` 文件未被提交到Git

#### 2. 数据库检查
- [ ] MySQL服务正常运行
- [ ] 数据库连接测试通过
- [ ] 数据库用户权限正确
- [ ] 数据库字符集为 utf8mb4
- [ ] 数据库时区配置正确（+08:00）

#### 3. Redis检查
- [ ] Redis服务正常运行
- [ ] Redis密码已配置
- [ ] Redis持久化已启用（AOF）
- [ ] Redis连接测试通过

#### 4. MinIO检查
- [ ] MinIO服务正常运行
- [ ] MinIO访问密钥已配置
- [ ] MinIO存储桶已创建
- [ ] MinIO连接测试通过

---

### 🔐 安全性检查

#### 5. 密码安全
- [ ] 所有默认密码已修改
- [ ] 密码复杂度符合要求（大小写+数字+特殊字符）
- [ ] 密码已安全存储（密码管理器）
- [ ] 生产环境密码与开发环境不同

#### 6. 网络安全
- [ ] 防火墙规则已配置
- [ ] 只开放必要端口（80, 443, 可选SSH）
- [ ] 内部服务端口未对外暴露
- [ ] SSL/TLS证书已配置（如使用HTTPS）

#### 7. 访问控制
- [ ] 超级管理员账号已创建
- [ ] 管理员账号启用了2FA
- [ ] IP黑名单功能已启用
- [ ] 限流规则已配置

---

### 🏗️ 代码和配置检查

#### 8. 代码质量
- [ ] 所有Critical和High优先级问题已修复
- [ ] 代码已通过CI检查
- [ ] 无明显的内存泄漏
- [ ] Goroutine管理正确

#### 9. Docker配置
- [ ] Docker Compose文件已检查
- [ ] 所有服务健康检查已配置
- [ ] 数据卷挂载正确
- [ ] 网络配置正确
- [ ] 重启策略已设置（unless-stopped）

#### 10. 日志配置
- [ ] 日志级别设置正确（生产环境：info或warn）
- [ ] 日志输出目录已配置
- [ ] 日志轮转已配置
- [ ] 敏感信息不输出到日志

---

### 📊 性能和监控

#### 11. 性能优化
- [ ] 数据库连接池已优化
- [ ] Redis缓存已启用
- [ ] 静态资源已优化
- [ ] CDN已配置（可选）

#### 12. 监控配置
- [ ] 系统监控已启用
- [ ] 告警规则已配置
- [ ] 日志监控已配置
- [ ] 性能指标收集已启用

#### 13. 备份配置
- [ ] 数据库自动备份已配置
- [ ] 备份策略已测试
- [ ] 备份恢复流程已验证
- [ ] 备份文件存储位置已确认

---

### 🧪 功能测试

#### 14. 基础功能测试
- [ ] 用户注册功能正常
- [ ] 用户登录功能正常
- [ ] JWT认证功能正常
- [ ] 2FA功能正常（如启用）

#### 15. 核心功能测试
- [ ] 消息发送接收正常
- [ ] 文件上传下载正常
- [ ] 群组创建管理正常
- [ ] WebRTC功能正常（如启用）

#### 16. 管理功能测试
- [ ] 超级管理后台可访问
- [ ] 用户管理功能正常
- [ ] 系统统计功能正常
- [ ] 内容审核功能正常

---

### 🚨 错误处理

#### 17. 错误处理测试
- [ ] 数据库连接失败处理
- [ ] Redis连接失败处理
- [ ] MinIO连接失败处理
- [ ] 网络异常处理

#### 18. 限流测试
- [ ] API限流功能正常
- [ ] 限流错误提示友好
- [ ] 限流阈值合理

---

### 📖 文档准备

#### 19. 文档完整性
- [ ] README.md已更新
- [ ] API文档已准备
- [ ] 部署文档已准备
- [ ] 用户手册已准备

#### 20. 运维文档
- [ ] 故障排查指南已准备
- [ ] 备份恢复指南已准备
- [ ] 扩容指南已准备
- [ ] 监控指南已准备

---

### 🔄 回滚准备

#### 21. 回滚计划
- [ ] 回滚方案已制定
- [ ] 回滚脚本已准备
- [ ] 回滚测试已完成
- [ ] 数据库迁移可回滚

---

## 📝 部署流程

### 步骤1：环境准备
```bash
# 1. 检查系统资源
df -h
free -h
docker info

# 2. 检查Docker和Docker Compose版本
docker --version
docker-compose --version

# 3. 拉取最新代码
git pull origin main
```

### 步骤2：配置环境变量
```bash
# 1. 复制环境变量模板
cp ENV_EXAMPLE.md .env

# 2. 编辑环境变量（填写真实值）
nano .env

# 3. 验证环境变量
cat .env | grep -v "^#" | grep "="
```

### 步骤3：构建和启动服务
```bash
# 1. 清理旧容器和数据（首次部署）
docker-compose -f docker-compose.production.yml down -v

# 2. 构建镜像
docker-compose -f docker-compose.production.yml build

# 3. 启动服务
docker-compose -f docker-compose.production.yml up -d

# 4. 查看服务状态
docker-compose -f docker-compose.production.yml ps
```

### 步骤4：健康检查
```bash
# 1. 检查所有容器状态
docker-compose -f docker-compose.production.yml ps

# 2. 检查后端健康
curl http://localhost:8080/health

# 3. 检查日志
docker-compose -f docker-compose.production.yml logs -f --tail=100
```

### 步骤5：创建超级管理员
```bash
# 参考 scripts/create-super-admin.md
```

### 步骤6：功能验证
```bash
# 1. 访问前端
# http://your-domain/

# 2. 测试登录

# 3. 访问超级管理后台
# http://your-domain/super-admin
```

---

## ⚠️ 常见问题排查

### 问题1：服务启动失败
```bash
# 查看详细日志
docker-compose -f docker-compose.production.yml logs [service-name]

# 检查端口占用
netstat -tulpn | grep [port]

# 检查磁盘空间
df -h
```

### 问题2：数据库连接失败
```bash
# 进入MySQL容器
docker exec -it im-mysql-prod mysql -uroot -p

# 检查用户权限
SHOW GRANTS FOR 'zhihang_messenger'@'%';

# 检查数据库
SHOW DATABASES;
USE zhihang_messenger;
SHOW TABLES;
```

### 问题3：Redis连接失败
```bash
# 进入Redis容器
docker exec -it im-redis-prod redis-cli

# 使用密码认证
AUTH your_redis_password

# 测试连接
PING
```

### 问题4：MinIO连接失败
```bash
# 检查MinIO状态
docker logs im-minio-prod

# 访问MinIO控制台
http://your-domain:9001
```

---

## 📊 性能基准测试

### 预期性能指标
- **API响应时间**: < 100ms (P95)
- **并发用户数**: > 1000
- **消息延迟**: < 500ms
- **CPU使用率**: < 70%
- **内存使用率**: < 80%
- **数据库连接数**: < 80

### 压力测试命令
```bash
# 使用 Apache Bench
ab -n 1000 -c 100 http://localhost:8080/health

# 使用 wrk
wrk -t4 -c100 -d30s http://localhost:8080/health
```

---

## ✅ 部署完成确认

部署完成后，请确认以下所有项目：

- [ ] 所有服务状态为 healthy
- [ ] 前端页面可正常访问
- [ ] 用户可正常注册登录
- [ ] 消息功能正常
- [ ] 管理后台可访问
- [ ] 监控数据正常收集
- [ ] 日志正常输出
- [ ] 备份任务正常运行

---

## 📞 紧急联系

如遇到紧急问题，请：

1. 立即检查服务日志
2. 查看监控告警
3. 如需回滚，执行回滚脚本
4. 记录问题详情
5. 联系技术支持

---

**检查清单版本**: v1.0  
**最后更新**: 2025-10-11  
**适用版本**: im-suite v1.4.0+

