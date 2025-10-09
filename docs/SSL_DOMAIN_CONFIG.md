# SSL证书和域名配置指南

## 📋 概述

本指南说明如何在测试环境（localhost）和生产环境（域名）之间切换SSL证书配置。

---

## 🧪 阶段1：测试环境（当前）- localhost

### ✅ 当前配置状态

**使用场景**：
- 本地开发和功能测试
- Devin测试验证
- 内网部署测试

**访问地址**：
```
HTTP（推荐用于测试）:
- 后端API: http://localhost:8080
- Web客户端: http://localhost:3002  
- 管理后台: http://localhost:3001

HTTPS（可选，需生成证书）:
- 后端API: https://localhost:443/api
- Web客户端: https://localhost:443
- 管理后台: https://localhost:443/admin
```

**证书配置**：
- 类型：自签名证书
- 域名：localhost, 127.0.0.1
- 浏览器警告：正常（可忽略）

### 生成自签名证书（可选）

```bash
# 使用项目脚本生成
chmod +x scripts/generate-self-signed-cert.sh
./scripts/generate-self-signed-cert.sh

# 证书位置
# ssl/cert.pem - 证书文件
# ssl/key.pem - 私钥文件
```

**注意**：
- ⚠️ 浏览器会显示"不安全"警告（正常现象）
- ⚠️ 仅用于测试，不要用于生产环境
- ✅ HTTP访问更简单，推荐测试时使用

---

## 🌐 阶段2：生产环境 - 使用域名

### 准备工作

#### 1. 域名准备
**需要准备**：
- 主域名：`yourdomain.com`
- API子域名：`api.yourdomain.com`（推荐）
- 管理后台：`admin.yourdomain.com`（推荐）

**DNS解析配置**：
```
类型  主机记录  记录值
A     @         您的服务器公网IP
A     api       您的服务器公网IP  
A     admin     您的服务器公网IP
A     im        您的服务器公网IP
```

#### 2. 服务器准备
- 开放80端口（HTTP）
- 开放443端口（HTTPS）
- 确保防火墙允许

---

### 方式1：宝塔面板（最简单，推荐）⭐

#### Step 1: 安装宝塔面板

```bash
# Ubuntu/Debian
wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh
sudo bash install.sh

# CentOS
yum install -y wget && wget -O install.sh https://download.bt.cn/install/install_6.0.sh
sudo sh install.sh
```

#### Step 2: 在宝塔中配置SSL

1. 登录宝塔面板（默认端口8888）
2. 网站 → 添加站点 → 输入域名
3. 网站设置 → SSL → Let's Encrypt
4. 勾选域名 → 申请
5. 自动续期已默认开启 ✅

#### Step 3: 导入nginx配置

1. 网站设置 → 配置文件
2. 复制项目的 `config/nginx/nginx.conf` 内容
3. 修改 `server_name` 为您的域名
4. 保存并重载

**完成！只需10分钟！** ⚡

---

### 方式2：Certbot手动配置

#### Step 1: 安装Certbot

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install certbot

# CentOS
sudo yum install certbot
```

#### Step 2: 获取证书

```bash
# 单域名
sudo certbot certonly --standalone -d yourdomain.com

# 多域名（推荐）
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d api.yourdomain.com \
  -d admin.yourdomain.com
```

#### Step 3: 配置自动续期

```bash
# 添加定时任务
sudo crontab -e

# 添加以下行（每天凌晨2点检查）
0 2 * * * certbot renew --quiet --post-hook "systemctl reload nginx"
```

#### Step 4: 更新配置文件

见下文"配置文件修改"部分

---

## 🔄 环境切换步骤

### 从测试环境切换到生产环境

#### Step 1: 修改nginx配置

**文件**: `config/nginx/nginx.conf`

```nginx
# 将第175行的 server_name 修改为您的域名
server_name yourdomain.com api.yourdomain.com admin.yourdomain.com;

# 将证书路径修改为Let's Encrypt路径
ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
```

#### Step 2: 修改Docker Compose

**文件**: `docker-compose.production.yml`

找到nginx服务，修改volumes:

```yaml
nginx:
  volumes:
    # 生产环境证书（启用）
    - /etc/letsencrypt:/etc/letsencrypt:ro
    
    # 测试环境证书（注释掉）
    # - ./ssl:/etc/nginx/ssl:ro
```

#### Step 3: 更新前端配置

**Web客户端**: `telegram-web/.env.production`
```bash
VITE_API_BASE_URL=https://api.yourdomain.com
VITE_WS_URL=wss://api.yourdomain.com/ws
```

**管理后台**: `im-admin/.env.production`
```bash
VITE_API_URL=https://api.yourdomain.com
```

#### Step 4: 重新部署

```bash
# 停止服务
docker-compose -f docker-compose.production.yml down

# 重新构建和启动
docker-compose -f docker-compose.production.yml up --build -d

# 验证
curl https://yourdomain.com/health
```

---

## 🔐 客户端SSL Pinning（生产环境必需）

### Android应用配置

创建 `telegram-android/TMessagesProj/src/main/res/xml/network_security_config.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <!-- 生产环境：启用SSL Pinning -->
    <domain-config cleartextTrafficPermitted="false">
        <domain includeSubdomains="true">yourdomain.com</domain>
        <pin-set expiration="2025-12-31">
            <!-- 主证书指纹 -->
            <pin digest="SHA-256">YOUR_CERT_FINGERPRINT_HERE</pin>
            <!-- 备用证书指纹 -->
            <pin digest="SHA-256">BACKUP_CERT_FINGERPRINT_HERE</pin>
        </pin-set>
    </domain-config>
    
    <!-- 测试环境：允许HTTP（开发时启用） -->
    <!--
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="true">localhost</domain>
        <domain includeSubdomains="true">127.0.0.1</domain>
        <domain includeSubdomains="true">10.0.2.2</domain>
    </domain-config>
    -->
</network-security-config>
```

### 获取证书指纹

```bash
# 方法1: 从证书文件获取
openssl x509 -in /etc/letsencrypt/live/yourdomain.com/fullchain.pem -pubkey -noout | \
  openssl rsa -pubin -outform der | \
  openssl dgst -sha256 -binary | \
  openssl enc -base64

# 方法2: 从服务器获取
echo | openssl s_client -connect yourdomain.com:443 2>/dev/null | \
  openssl x509 -pubkey -noout | \
  openssl rsa -pubin -outform der | \
  openssl dgst -sha256 -binary | \
  openssl enc -base64
```

---

## 📊 配置对照表

| 配置项 | 测试环境（现在） | 生产环境（上线后） |
|--------|----------------|------------------|
| **域名** | localhost | yourdomain.com |
| **SSL证书** | 自签名（可选） | Let's Encrypt / 商业证书 |
| **证书路径** | `./ssl/` | `/etc/letsencrypt/` |
| **HTTP** | ✅ 推荐 | ⚠️ 重定向到HTTPS |
| **HTTPS** | ✅ 可选 | ✅ 必需 |
| **SSL Pinning** | ❌ 禁用 | ✅ 启用（移动端） |
| **自动续期** | ❌ 不需要 | ✅ 必需 |
| **CDN** | ❌ 不需要 | ✅ 推荐 |

---

## ✅ 当前状态总结

### 测试阶段配置 ✅

**现在可以使用**：
- ✅ HTTP访问（推荐）：`http://localhost:8080`
- ✅ 无需SSL证书
- ✅ 无需域名
- ✅ Devin可以直接测试

**后续生产部署**：
- 📋 获取域名
- 📋 申请SSL证书（宝塔面板一键）
- 📋 修改配置文件（5个文件）
- 📋 重新部署（10分钟）

---

**建议**：
1. **现在**: 使用HTTP + localhost进行测试 ✅
2. **Devin测试通过后**: 准备域名和SSL证书
3. **正式上线**: 按照本文档切换到生产环境

**切换难度**: ⭐⭐ (简单，10-15分钟)  
**推荐工具**: 宝塔面板（最简单）

---

**最后更新**: 2024-12-19  
**适用版本**: v1.4.0+

