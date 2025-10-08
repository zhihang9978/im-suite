# SSL 证书目录

此目录包含志航密信项目的 SSL 证书相关文件。

## 📁 文件说明

- `generate-ssl.sh` - SSL 证书生成脚本（开发环境）
- `README.md` - 本说明文件

## 🔐 证书文件

运行生成脚本后，会在此目录下创建以下文件：

- `zhihang-messenger.key` - 私钥文件
- `zhihang-messenger.crt` - 证书文件

## ⚠️ 安全提醒

1. **私钥文件**（.key）包含敏感信息，请妥善保管
2. **不要将私钥文件提交到版本控制系统**
3. **生产环境请使用正式的 SSL 证书**
4. **定期更新证书以确保安全**

## 🚀 使用方法

### 开发环境

```bash
# 生成自签名证书
./scripts/ssl/generate-ssl.sh

# 启动服务
docker-compose up -d
```

### 生产环境

请参考 [SSL 证书配置指南](../docs/deployment/ssl-certificates.md) 中的宝塔面板配置部分。

## 📚 相关文档

- [SSL 证书配置指南](../docs/deployment/ssl-certificates.md)
- [部署文档](../docs/deployment/README.md)


