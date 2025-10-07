# IM-Suite 传输安全配置

## 概述

本文档描述了 IM-Suite 的传输安全配置，包括 HTTPS/WSS 配置、证书管理、安全头设置等。

## HTTPS 配置

### 1. 服务器端配置

#### Nginx HTTPS 配置
```nginx
# 主配置文件
server {
    listen 80;
    server_name im-suite.com www.im-suite.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name im-suite.com;
    
    # SSL 证书配置
    ssl_certificate /etc/ssl/certs/im-suite.crt;
    ssl_certificate_key /etc/ssl/private/im-suite.key;
    ssl_certificate_chain /etc/ssl/certs/im-suite-chain.crt;
    
    # SSL 协议配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA;
    ssl_prefer_server_ciphers on;
    
    # SSL 会话配置
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/ssl/certs/ca-certificates.crt;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
    
    # HSTS 配置
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    
    # 安全头配置
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: https:; frame-ancestors 'none';" always;
    
    # API 路由
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 安全配置
        proxy_hide_header X-Powered-By;
        proxy_hide_header Server;
    }
    
    # WebSocket 路由
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 安全配置
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }
}
```

#### Apache HTTPS 配置
```apache
# Apache 虚拟主机配置
<VirtualHost *:443>
    ServerName im-suite.com
    DocumentRoot /var/www/im-suite
    
    # SSL 配置
    SSLEngine on
    SSLCertificateFile /etc/ssl/certs/im-suite.crt
    SSLCertificateKeyFile /etc/ssl/private/im-suite.key
    SSLCertificateChainFile /etc/ssl/certs/im-suite-chain.crt
    
    # SSL 协议配置
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256
    SSLHonorCipherOrder on
    
    # HSTS 配置
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
    
    # 安全头配置
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Referrer-Policy "strict-origin-when-cross-origin"
    Header always set Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: https:; frame-ancestors 'none';"
    
    # 隐藏服务器信息
    ServerTokens Prod
    Header unset Server
    Header unset X-Powered-By
</VirtualHost>
```

### 2. 客户端配置

#### Web 端 HTTPS 配置
```javascript
// HTTPS 配置
const httpsConfig = {
  // 强制 HTTPS
  forceHttps: true,
  
  // 证书固定
  certificatePinning: {
    'api.im-suite.com': [
      'sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      'sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB='
    ],
    'ws.im-suite.com': [
      'sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=',
      'sha256/DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD='
    ]
  },
  
  // 最小 TLS 版本
  minTlsVersion: 'TLSv1.2',
  
  // 支持的密码套件
  ciphers: [
    'ECDHE-RSA-AES256-GCM-SHA384',
    'ECDHE-RSA-AES128-GCM-SHA256',
    'ECDHE-RSA-AES256-SHA384',
    'ECDHE-RSA-AES128-SHA256'
  ],
  
  // 安全头检查
  securityHeaders: {
    'Strict-Transport-Security': 'max-age=31536000; includeSubDomains; preload',
    'X-Frame-Options': 'DENY',
    'X-Content-Type-Options': 'nosniff',
    'X-XSS-Protection': '1; mode=block',
    'Referrer-Policy': 'strict-origin-when-cross-origin'
  }
};

// 证书固定实现
class CertificatePinning {
  constructor() {
    this.pinnedCertificates = new Map();
  }
  
  // 添加固定证书
  addPinnedCertificate(hostname, certificates) {
    this.pinnedCertificates.set(hostname, certificates);
  }
  
  // 验证证书
  async verifyCertificate(hostname, certificate) {
    const pinnedCerts = this.pinnedCertificates.get(hostname);
    if (!pinnedCerts) {
      return true; // 没有固定证书，允许通过
    }
    
    // 计算证书哈希
    const certHash = await this.calculateCertificateHash(certificate);
    
    // 检查是否匹配
    return pinnedCerts.includes(certHash);
  }
  
  // 计算证书哈希
  async calculateCertificateHash(certificate) {
    const certBuffer = await certificate.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest('SHA-256', certBuffer);
    const hashArray = new Uint8Array(hashBuffer);
    const hashBase64 = btoa(String.fromCharCode(...hashArray));
    return `sha256/${hashBase64}`;
  }
}
```

#### Android 端 HTTPS 配置
```kotlin
// Android 网络安全配置
class NetworkSecurityConfig {
    
    // 证书固定配置
    fun createCertificatePinning(): NetworkSecurityConfig {
        return NetworkSecurityConfig.Builder()
            .addCertificatePinner(
                "api.im-suite.com",
                "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
                "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="
            )
            .addCertificatePinner(
                "ws.im-suite.com",
                "sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=",
                "sha256/DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD="
            )
            .build()
    }
    
    // OkHttp 客户端配置
    fun createSecureOkHttpClient(): OkHttpClient {
        val certificatePinner = CertificatePinner.Builder()
            .add("api.im-suite.com", "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
            .add("api.im-suite.com", "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=")
            .add("ws.im-suite.com", "sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=")
            .add("ws.im-suite.com", "sha256/DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD=")
            .build()
        
        return OkHttpClient.Builder()
            .certificatePinner(certificatePinner)
            .connectionSpecs(listOf(ConnectionSpec.MODERN_TLS))
            .build()
    }
}
```

## WSS 配置

### 1. WebSocket 安全配置

#### 服务器端 WSS 配置
```javascript
// WebSocket 服务器配置
const WebSocket = require('ws');
const https = require('https');
const fs = require('fs');

// 读取 SSL 证书
const options = {
  cert: fs.readFileSync('/etc/ssl/certs/im-suite.crt'),
  key: fs.readFileSync('/etc/ssl/private/im-suite.key'),
  ca: fs.readFileSync('/etc/ssl/certs/im-suite-chain.crt')
};

// 创建 HTTPS 服务器
const server = https.createServer(options);

// 创建 WebSocket 服务器
const wss = new WebSocket.Server({
  server: server,
  verifyClient: (info) => {
    // 验证客户端连接
    const origin = info.origin;
    const allowedOrigins = ['https://im-suite.com', 'https://www.im-suite.com'];
    return allowedOrigins.includes(origin);
  }
});

// 连接处理
wss.on('connection', (ws, req) => {
  // 设置安全头
  ws.on('headers', (headers) => {
    headers['X-Frame-Options'] = 'DENY';
    headers['X-Content-Type-Options'] = 'nosniff';
    headers['X-XSS-Protection'] = '1; mode=block';
  });
  
  // 消息处理
  ws.on('message', (message) => {
    // 验证消息格式
    try {
      const data = JSON.parse(message);
      if (this.validateMessage(data)) {
        this.handleMessage(ws, data);
      }
    } catch (error) {
      ws.close(1008, 'Invalid message format');
    }
  });
  
  // 错误处理
  ws.on('error', (error) => {
    console.error('WebSocket error:', error);
    ws.close(1011, 'Internal error');
  });
  
  // 连接关闭
  ws.on('close', (code, reason) => {
    console.log('WebSocket closed:', code, reason);
  });
});

// 启动服务器
server.listen(443, () => {
  console.log('WSS server running on port 443');
});
```

#### 客户端 WSS 配置
```javascript
// WebSocket 客户端配置
class SecureWebSocket {
  constructor(url, options = {}) {
    this.url = url;
    this.options = {
      // 安全配置
      secure: true,
      rejectUnauthorized: true,
      
      // 证书固定
      certificatePinning: options.certificatePinning || {},
      
      // 超时配置
      handshakeTimeout: 10000,
      pingTimeout: 30000,
      pongTimeout: 5000,
      
      // 重连配置
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      
      ...options
    };
    
    this.ws = null;
    this.reconnectAttempts = 0;
    this.isConnected = false;
  }
  
  // 连接 WebSocket
  connect() {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url, [], {
          // 安全配置
          rejectUnauthorized: this.options.rejectUnauthorized,
          checkServerIdentity: this.checkServerIdentity.bind(this)
        });
        
        this.ws.onopen = () => {
          this.isConnected = true;
          this.reconnectAttempts = 0;
          resolve();
        };
        
        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };
        
        this.ws.onclose = (event) => {
          this.isConnected = false;
          this.handleClose(event);
        };
        
        this.ws.onerror = (error) => {
          this.handleError(error);
          reject(error);
        };
        
      } catch (error) {
        reject(error);
      }
    });
  }
  
  // 检查服务器身份
  checkServerIdentity(hostname, cert) {
    const pinnedCerts = this.options.certificatePinning[hostname];
    if (!pinnedCerts) {
      return undefined; // 没有固定证书，使用默认验证
    }
    
    // 计算证书哈希
    const certHash = this.calculateCertificateHash(cert);
    
    // 检查是否匹配
    if (!pinnedCerts.includes(certHash)) {
      throw new Error('Certificate pinning failed');
    }
    
    return undefined;
  }
  
  // 计算证书哈希
  calculateCertificateHash(cert) {
    const crypto = require('crypto');
    const hash = crypto.createHash('sha256').update(cert).digest('base64');
    return `sha256/${hash}`;
  }
  
  // 处理消息
  handleMessage(data) {
    try {
      const message = JSON.parse(data);
      this.onMessage(message);
    } catch (error) {
      console.error('Invalid message format:', error);
    }
  }
  
  // 处理连接关闭
  handleClose(event) {
    if (event.code !== 1000 && this.reconnectAttempts < this.options.maxReconnectAttempts) {
      // 非正常关闭，尝试重连
      setTimeout(() => {
        this.reconnectAttempts++;
        this.connect();
      }, this.options.reconnectInterval);
    }
  }
  
  // 处理错误
  handleError(error) {
    console.error('WebSocket error:', error);
  }
  
  // 发送消息
  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      throw new Error('WebSocket not connected');
    }
  }
  
  // 关闭连接
  close() {
    if (this.ws) {
      this.ws.close(1000, 'Normal closure');
    }
  }
}
```

## 证书管理

### 1. 证书生成

#### 自签名证书生成
```bash
#!/bin/bash
# 生成自签名证书

# 创建证书目录
mkdir -p /etc/ssl/certs /etc/ssl/private

# 生成私钥
openssl genrsa -out /etc/ssl/private/im-suite.key 4096

# 生成证书签名请求
openssl req -new -key /etc/ssl/private/im-suite.key -out /etc/ssl/certs/im-suite.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=IM-Suite/OU=IT/CN=im-suite.com"

# 生成自签名证书
openssl x509 -req -days 365 -in /etc/ssl/certs/im-suite.csr -signkey /etc/ssl/private/im-suite.key -out /etc/ssl/certs/im-suite.crt

# 设置权限
chmod 600 /etc/ssl/private/im-suite.key
chmod 644 /etc/ssl/certs/im-suite.crt

# 生成证书链
cat /etc/ssl/certs/im-suite.crt > /etc/ssl/certs/im-suite-chain.crt
```

#### Let's Encrypt 证书生成
```bash
#!/bin/bash
# 使用 Let's Encrypt 生成证书

# 安装 certbot
apt-get update
apt-get install -y certbot python3-certbot-nginx

# 生成证书
certbot --nginx -d im-suite.com -d www.im-suite.com

# 设置自动续期
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### 2. 证书验证

#### 证书验证脚本
```bash
#!/bin/bash
# 证书验证脚本

check_certificate() {
    local domain=$1
    local port=${2:-443}
    
    echo "检查证书: $domain:$port"
    
    # 检查证书有效期
    local expiry=$(echo | openssl s_client -servername $domain -connect $domain:$port 2>/dev/null | openssl x509 -noout -dates | grep notAfter | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry" +%s)
    local current_epoch=$(date +%s)
    local days_left=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    echo "证书有效期: $expiry"
    echo "剩余天数: $days_left"
    
    if [ $days_left -lt 30 ]; then
        echo "警告: 证书将在 30 天内过期"
        return 1
    fi
    
    # 检查证书链
    local chain_length=$(echo | openssl s_client -servername $domain -connect $domain:$port 2>/dev/null | openssl x509 -noout -text | grep -c "Issuer:")
    
    echo "证书链长度: $chain_length"
    
    # 检查证书指纹
    local fingerprint=$(echo | openssl s_client -servername $domain -connect $domain:$port 2>/dev/null | openssl x509 -noout -fingerprint -sha256 | cut -d= -f2)
    echo "证书指纹: $fingerprint"
    
    return 0
}

# 检查主域名
check_certificate "im-suite.com"

# 检查子域名
check_certificate "api.im-suite.com"
check_certificate "ws.im-suite.com"
```

### 3. 证书监控

#### 证书监控脚本
```bash
#!/bin/bash
# 证书监控脚本

MONITOR_DOMAINS=("im-suite.com" "api.im-suite.com" "ws.im-suite.com")
ALERT_DAYS=30
WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

check_certificate_expiry() {
    local domain=$1
    local port=${2:-443}
    
    # 获取证书过期时间
    local expiry=$(echo | openssl s_client -servername $domain -connect $domain:$port 2>/dev/null | openssl x509 -noout -dates | grep notAfter | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry" +%s)
    local current_epoch=$(date +%s)
    local days_left=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    if [ $days_left -lt $ALERT_DAYS ]; then
        send_alert "$domain" "$days_left" "$expiry"
    fi
}

send_alert() {
    local domain=$1
    local days_left=$2
    local expiry=$3
    
    local message="SSL证书即将过期\n域名: $domain\n剩余天数: $days_left\n过期时间: $expiry"
    
    # 发送到 Slack
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"$message\"}" \
        $WEBHOOK_URL
    
    # 发送邮件
    echo "$message" | mail -s "SSL证书即将过期: $domain" admin@im-suite.com
}

# 检查所有域名
for domain in "${MONITOR_DOMAINS[@]}"; do
    check_certificate_expiry "$domain"
done
```

## 安全头配置

### 1. 安全头实现

#### 服务器端安全头
```javascript
// Express.js 安全头中间件
const helmet = require('helmet');
const express = require('express');

const app = express();

// 使用 helmet 设置安全头
app.use(helmet({
  // HSTS 配置
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true
  },
  
  // 内容安全策略
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "'unsafe-inline'", "'unsafe-eval'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
      fontSrc: ["'self'", "data:"],
      connectSrc: ["'self'", "wss:", "https:"],
      frameAncestors: ["'none'"],
      objectSrc: ["'none'"],
      upgradeInsecureRequests: []
    }
  },
  
  // 其他安全头
  frameguard: { action: 'deny' },
  noSniff: true,
  xssFilter: true,
  referrerPolicy: { policy: 'strict-origin-when-cross-origin' }
}));

// 自定义安全头
app.use((req, res, next) => {
  // 隐藏服务器信息
  res.removeHeader('X-Powered-By');
  res.setHeader('Server', 'IM-Suite');
  
  // 防止点击劫持
  res.setHeader('X-Frame-Options', 'DENY');
  
  // 防止 MIME 类型嗅探
  res.setHeader('X-Content-Type-Options', 'nosniff');
  
  // XSS 保护
  res.setHeader('X-XSS-Protection', '1; mode=block');
  
  // 引用策略
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
  
  next();
});
```

#### 客户端安全头检查
```javascript
// 客户端安全头检查
class SecurityHeaderChecker {
  constructor() {
    this.requiredHeaders = {
      'Strict-Transport-Security': 'max-age=31536000; includeSubDomains; preload',
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'X-XSS-Protection': '1; mode=block',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Content-Security-Policy': /default-src 'self'/
    };
  }
  
  // 检查安全头
  async checkSecurityHeaders(url) {
    try {
      const response = await fetch(url, { method: 'HEAD' });
      const headers = response.headers;
      
      const results = {};
      
      for (const [header, expected] of Object.entries(this.requiredHeaders)) {
        const actual = headers.get(header);
        
        if (typeof expected === 'string') {
          results[header] = actual === expected;
        } else if (expected instanceof RegExp) {
          results[header] = expected.test(actual);
        }
      }
      
      return results;
    } catch (error) {
      console.error('检查安全头失败:', error);
      return null;
    }
  }
  
  // 验证所有安全头
  async validateAllHeaders(url) {
    const results = await this.checkSecurityHeaders(url);
    if (!results) {
      return false;
    }
    
    const allValid = Object.values(results).every(valid => valid);
    
    if (!allValid) {
      console.warn('部分安全头配置不正确:', results);
    }
    
    return allValid;
  }
}
```

## 网络安全

### 1. 防火墙配置

#### iptables 配置
```bash
#!/bin/bash
# iptables 防火墙配置

# 清除现有规则
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# 设置默认策略
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# 允许本地回环
iptables -A INPUT -i lo -j ACCEPT

# 允许已建立的连接
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# 允许 SSH (端口 22)
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# 允许 HTTP (端口 80)
iptables -A INPUT -p tcp --dport 80 -j ACCEPT

# 允许 HTTPS (端口 443)
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# 允许 WebSocket (端口 8080)
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# 限制连接数
iptables -A INPUT -p tcp --dport 80 -m connlimit --connlimit-above 20 -j DROP
iptables -A INPUT -p tcp --dport 443 -m connlimit --connlimit-above 20 -j DROP

# 防止 DDoS 攻击
iptables -A INPUT -p tcp --dport 80 -m limit --limit 25/minute --limit-burst 100 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -m limit --limit 25/minute --limit-burst 100 -j ACCEPT

# 保存规则
iptables-save > /etc/iptables/rules.v4
```

#### UFW 配置
```bash
#!/bin/bash
# UFW 防火墙配置

# 重置 UFW
ufw --force reset

# 设置默认策略
ufw default deny incoming
ufw default allow outgoing

# 允许 SSH
ufw allow 22/tcp

# 允许 HTTP
ufw allow 80/tcp

# 允许 HTTPS
ufw allow 443/tcp

# 允许 WebSocket
ufw allow 8080/tcp

# 启用 UFW
ufw --force enable

# 查看状态
ufw status verbose
```

### 2. 入侵检测

#### Fail2ban 配置
```ini
# /etc/fail2ban/jail.local
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3
```

#### 监控脚本
```bash
#!/bin/bash
# 安全监控脚本

# 监控异常登录
monitor_failed_logins() {
    tail -f /var/log/auth.log | grep "Failed password" | while read line; do
        echo "异常登录尝试: $line"
        # 发送告警
        send_alert "异常登录尝试" "$line"
    done
}

# 监控异常请求
monitor_suspicious_requests() {
    tail -f /var/log/nginx/access.log | grep -E "(sqlmap|nikto|nmap|masscan)" | while read line; do
        echo "可疑请求: $line"
        # 发送告警
        send_alert "可疑请求" "$line"
    done
}

# 监控资源使用
monitor_resource_usage() {
    while true; do
        # 检查 CPU 使用率
        cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
        if (( $(echo "$cpu_usage > 80" | bc -l) )); then
            send_alert "高CPU使用率" "CPU使用率: $cpu_usage%"
        fi
        
        # 检查内存使用率
        memory_usage=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
        if (( $(echo "$memory_usage > 80" | bc -l) )); then
            send_alert "高内存使用率" "内存使用率: $memory_usage%"
        fi
        
        sleep 60
    done
}

# 发送告警
send_alert() {
    local title=$1
    local message=$2
    
    # 发送到 Slack
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"$title: $message\"}" \
        $WEBHOOK_URL
    
    # 发送邮件
    echo "$message" | mail -s "$title" admin@im-suite.com
}

# 启动监控
monitor_failed_logins &
monitor_suspicious_requests &
monitor_resource_usage &
```

## 最佳实践

### 1. 安全配置建议
- 使用强密码和证书
- 定期更新安全补丁
- 监控安全事件
- 实施访问控制
- 备份安全配置

### 2. 监控建议
- 实时监控安全事件
- 设置告警机制
- 定期安全审计
- 分析安全日志
- 更新安全策略

### 3. 维护建议
- 定期更新证书
- 监控安全漏洞
- 测试安全配置
- 培训安全团队
- 建立响应流程
