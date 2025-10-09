# 🔧 服务器网络问题排查和修复指南

**服务器IP**: 154.37.214.191  
**问题**: Docker镜像拉取失败  
**状态**: 🔴 需要修复  
**更新日期**: 2024-12-19

---

## 🚨 问题症状

### 观察到的问题

1. ❌ **DNS解析失败**
   - `docker.mirrors.ustc.edu.cn` 无法解析
   - 域名解析超时

2. ❌ **Docker Hub连接失败**
   - `registry-1.docker.io` 100%丢包
   - TLS握手超时

3. ❌ **镜像拉取失败**
   - 所有Docker镜像拉取失败
   - 无法部署任何服务

---

## 🔍 问题诊断

### 步骤1: 检查网络连接

```bash
# 测试DNS解析
ping -c 4 8.8.8.8
ping -c 4 google.com

# 测试HTTPS连接
curl -I https://www.google.com
curl -I https://hub.docker.com

# 检查DNS配置
cat /etc/resolv.conf

# 检查路由
ip route show
```

### 步骤2: 检查防火墙

```bash
# 检查防火墙状态
sudo ufw status
sudo iptables -L

# 检查SELinux（如果是CentOS/RHEL）
sestatus
```

### 步骤3: 检查Docker配置

```bash
# 检查Docker守护进程
sudo systemctl status docker

# 查看Docker配置
sudo cat /etc/docker/daemon.json

# 检查Docker网络
docker network ls
```

---

## 🛠️ 修复方案

### 方案1: 修复DNS配置（推荐优先尝试）

#### 1.1 更换DNS服务器

```bash
# 备份原配置
sudo cp /etc/resolv.conf /etc/resolv.conf.backup

# 修改DNS为可靠的公共DNS
sudo tee /etc/resolv.conf > /dev/null <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 114.114.114.114
nameserver 223.5.5.5
EOF

# 测试DNS解析
ping -c 4 google.com
nslookup docker.io
```

#### 1.2 配置永久DNS（防止重启后丢失）

**Ubuntu/Debian系统**:
```bash
# 安装resolvconf
sudo apt-get update
sudo apt-get install -y resolvconf

# 配置DNS
sudo tee /etc/resolvconf/resolv.conf.d/head > /dev/null <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 114.114.114.114
EOF

# 重启服务
sudo systemctl restart resolvconf
sudo systemctl restart systemd-resolved
```

**CentOS/RHEL系统**:
```bash
# 修改网络配置
sudo vi /etc/sysconfig/network-scripts/ifcfg-eth0
# 添加:
# DNS1=8.8.8.8
# DNS2=114.114.114.114

# 重启网络
sudo systemctl restart network
```

---

### 方案2: 配置Docker国内镜像源（推荐）

#### 2.1 创建Docker配置文件

```bash
# 创建配置目录
sudo mkdir -p /etc/docker

# 配置国内镜像源
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com"
  ],
  "dns": ["8.8.8.8", "114.114.114.114"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "insecure-registries": []
}
EOF

# 重启Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证配置
sudo docker info | grep -A 10 "Registry Mirrors"
```

#### 2.2 测试镜像拉取

```bash
# 测试拉取小镜像
docker pull alpine:latest

# 如果成功，继续拉取项目镜像
docker pull mysql:8.0
docker pull redis:7-alpine
docker pull nginx:alpine
```

---

### 方案3: 使用阿里云镜像加速（如果方案2失败）

#### 3.1 获取阿里云镜像加速器地址

访问: https://cr.console.aliyun.com/cn-hangzhou/instances/mirrors

登录后获取专属加速器地址，格式如: `https://xxxxx.mirror.aliyuncs.com`

#### 3.2 配置阿里云镜像

```bash
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://xxxxx.mirror.aliyuncs.com",
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ],
  "dns": ["223.5.5.5", "8.8.8.8"]
}
EOF

sudo systemctl daemon-reload
sudo systemctl restart docker
```

---

### 方案4: 检查云服务器安全组（重要！）

#### 4.1 阿里云

```
1. 登录阿里云控制台
2. 进入ECS管理
3. 找到实例 154.37.214.191
4. 点击"安全组配置"
5. 确保有以下规则：
   - 出方向: HTTPS(443) → 0.0.0.0/0 允许
   - 出方向: HTTP(80) → 0.0.0.0/0 允许
   - 出方向: DNS(53) → 0.0.0.0/0 允许
```

#### 4.2 腾讯云

```
1. 登录腾讯云控制台
2. 进入云服务器CVM
3. 找到实例 154.37.214.191
4. 点击"安全组"
5. 检查出站规则，确保允许：
   - 协议: TCP, 端口: 443, 目标: 0.0.0.0/0
   - 协议: TCP, 端口: 80, 目标: 0.0.0.0/0
   - 协议: UDP, 端口: 53, 目标: 0.0.0.0/0
```

#### 4.3 AWS

```
1. 登录AWS控制台
2. 进入EC2
3. 找到实例 154.37.214.191
4. 检查Security Groups
5. 确保Outbound Rules包含：
   - Type: HTTPS, Protocol: TCP, Port: 443, Destination: 0.0.0.0/0
   - Type: HTTP, Protocol: TCP, Port: 80, Destination: 0.0.0.0/0
   - Type: DNS(UDP), Protocol: UDP, Port: 53, Destination: 0.0.0.0/0
```

---

### 方案5: 手动上传Docker镜像（备用方案）

如果以上方案都失败，可以手动上传镜像：

#### 5.1 在本地构建和保存镜像

```bash
# 在本地机器上（有网络的环境）
cd im-suite

# 拉取基础镜像
docker pull mysql:8.0
docker pull redis:7-alpine
docker pull nginx:alpine
docker pull golang:1.21-alpine
docker pull node:18-alpine

# 保存镜像为tar文件
docker save mysql:8.0 -o mysql-8.0.tar
docker save redis:7-alpine -o redis-7-alpine.tar
docker save nginx:alpine -o nginx-alpine.tar
docker save golang:1.21-alpine -o golang-1.21-alpine.tar
docker save node:18-alpine -o node-18-alpine.tar

# 打包所有镜像
tar czf docker-images.tar.gz *.tar
```

#### 5.2 上传到服务器

```bash
# 使用scp上传
scp docker-images.tar.gz root@154.37.214.191:/tmp/

# 或使用rsync
rsync -avz docker-images.tar.gz root@154.37.214.191:/tmp/
```

#### 5.3 在服务器上加载镜像

```bash
# SSH连接到服务器
ssh root@154.37.214.191

# 解压
cd /tmp
tar xzf docker-images.tar.gz

# 加载镜像
docker load -i mysql-8.0.tar
docker load -i redis-7-alpine.tar
docker load -i nginx-alpine.tar
docker load -i golang-1.21-alpine.tar
docker load -i node-18-alpine.tar

# 验证
docker images

# 清理
rm -f /tmp/*.tar /tmp/docker-images.tar.gz
```

---

### 方案6: 使用国内VPS中转（高级方案）

如果服务器在国外无法访问国内镜像，可以使用代理：

```bash
# 配置HTTP代理
sudo mkdir -p /etc/systemd/system/docker.service.d

sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf > /dev/null <<EOF
[Service]
Environment="HTTP_PROXY=http://your-proxy:port"
Environment="HTTPS_PROXY=http://your-proxy:port"
Environment="NO_PROXY=localhost,127.0.0.1"
EOF

sudo systemctl daemon-reload
sudo systemctl restart docker
```

---

## 🎯 推荐执行顺序

### 快速修复流程（15分钟）

```bash
# 1. 修复DNS（2分钟）
sudo tee /etc/resolv.conf > /dev/null <<EOF
nameserver 8.8.8.8
nameserver 114.114.114.114
nameserver 223.5.5.5
EOF

# 2. 配置Docker镜像源（3分钟）
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com"
  ],
  "dns": ["8.8.8.8", "114.114.114.114"]
}
EOF

# 3. 重启Docker（1分钟）
sudo systemctl daemon-reload
sudo systemctl restart docker

# 4. 测试拉取（5分钟）
docker pull alpine:latest

# 5. 检查云服务器安全组（5分钟）
# 在云服务器控制台操作
```

---

## 📋 诊断命令清单

### 网络诊断

```bash
# 1. 测试基础连接
ping -c 4 8.8.8.8                    # 测试Google DNS
ping -c 4 114.114.114.114            # 测试国内DNS
ping -c 4 baidu.com                  # 测试国内网站

# 2. 测试DNS解析
nslookup docker.io
nslookup hub.docker.com
nslookup docker.mirrors.ustc.edu.cn

# 3. 测试HTTPS连接
curl -v https://hub.docker.com
curl -v https://registry-1.docker.io/v2/

# 4. 测试端口
telnet registry-1.docker.io 443
nc -zv registry-1.docker.io 443

# 5. 查看路由
traceroute registry-1.docker.io
mtr -n -c 10 registry-1.docker.io
```

### Docker诊断

```bash
# 1. 检查Docker状态
sudo systemctl status docker
sudo docker info

# 2. 查看Docker日志
sudo journalctl -u docker -n 50

# 3. 测试Docker网络
sudo docker run --rm alpine ping -c 4 8.8.8.8
sudo docker run --rm alpine nslookup google.com

# 4. 查看镜像源配置
sudo docker info | grep -i mirror
cat /etc/docker/daemon.json
```

---

## 🔧 完整修复脚本

### 一键修复脚本

将以下内容保存为 `fix-docker-network.sh`:

```bash
#!/bin/bash

echo "========================================="
echo "Docker网络问题一键修复脚本"
echo "========================================="

# 1. 备份配置
echo "1. 备份原配置..."
sudo cp /etc/resolv.conf /etc/resolv.conf.backup.$(date +%Y%m%d_%H%M%S)
if [ -f /etc/docker/daemon.json ]; then
    sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d_%H%M%S)
fi

# 2. 修复DNS
echo "2. 修复DNS配置..."
sudo tee /etc/resolv.conf > /dev/null <<EOF
# Google DNS
nameserver 8.8.8.8
nameserver 8.8.4.4
# 阿里DNS
nameserver 223.5.5.5
nameserver 223.6.6.6
# 国内DNS
nameserver 114.114.114.114
nameserver 119.29.29.29
EOF

echo "   DNS配置已更新"

# 3. 配置Docker镜像源
echo "3. 配置Docker镜像源..."
sudo mkdir -p /etc/docker

sudo tee /etc/docker/daemon.json > /dev/null <<'EOF'
{
  "registry-mirrors": [
    "https://hub-mirror.c.163.com",
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com",
    "https://docker.mirrors.ustc.edu.cn"
  ],
  "dns": ["8.8.8.8", "114.114.114.114", "223.5.5.5"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
EOF

echo "   Docker配置已更新"

# 4. 重启Docker
echo "4. 重启Docker服务..."
sudo systemctl daemon-reload
sudo systemctl restart docker
sleep 3

# 5. 验证
echo "5. 验证修复结果..."
echo "   测试DNS解析..."
if ping -c 2 8.8.8.8 > /dev/null 2>&1; then
    echo "   ✅ DNS连接正常"
else
    echo "   ❌ DNS连接失败"
fi

if nslookup docker.io > /dev/null 2>&1; then
    echo "   ✅ DNS解析正常"
else
    echo "   ❌ DNS解析失败"
fi

echo "   测试Docker镜像拉取..."
if timeout 60 docker pull alpine:latest > /dev/null 2>&1; then
    echo "   ✅ Docker镜像拉取成功"
    docker rmi alpine:latest
else
    echo "   ❌ Docker镜像拉取失败"
fi

echo ""
echo "========================================="
echo "修复完成！"
echo "========================================="
echo ""
echo "下一步:"
echo "1. 如果仍然失败，请检查云服务器安全组"
echo "2. 确保允许HTTPS(443)和DNS(53)出站连接"
echo "3. 或使用方案5手动上传镜像"
```

#### 运行修复脚本

```bash
# 上传到服务器
scp fix-docker-network.sh root@154.37.214.191:/tmp/

# SSH连接并执行
ssh root@154.37.214.191
cd /tmp
chmod +x fix-docker-network.sh
sudo ./fix-docker-network.sh
```

---

## 🌐 云服务器安全组配置

### 必需的出站规则

| 协议 | 端口 | 目标 | 说明 |
|------|------|------|------|
| TCP | 443 | 0.0.0.0/0 | HTTPS（Docker Hub） |
| TCP | 80 | 0.0.0.0/0 | HTTP |
| UDP | 53 | 0.0.0.0/0 | DNS查询 |
| ICMP | - | 0.0.0.0/0 | Ping测试 |

### 阿里云配置步骤

```
1. 登录阿里云控制台: https://ecs.console.aliyun.com
2. 左侧菜单 → 网络与安全 → 安全组
3. 找到实例绑定的安全组
4. 点击"配置规则"
5. 切换到"出方向"标签
6. 添加规则:
   - 授权策略: 允许
   - 协议类型: 全部
   - 端口范围: -1/-1
   - 授权对象: 0.0.0.0/0
   或者具体添加：
   - TCP 443, 80
   - UDP 53
7. 点击"保存"
```

### 腾讯云配置步骤

```
1. 登录腾讯云控制台: https://console.cloud.tencent.com/cvm
2. 找到实例 154.37.214.191
3. 点击实例ID进入详情
4. 点击"安全组"标签
5. 点击对应安全组ID
6. 切换到"出站规则"
7. 添加规则:
   - 类型: 全部流量
   - 来源: 0.0.0.0/0
   - 策略: 允许
8. 点击"完成"
```

---

## 🔄 备用部署方案

### 方案A: 使用预构建镜像（推荐）

```bash
# 1. 在本地机器构建项目镜像
cd im-suite

# 构建后端镜像
cd im-backend
docker build -t zhihang-backend:v1.6.0 -f Dockerfile.production .

# 构建管理后台镜像
cd ../im-admin
docker build -t zhihang-admin:v1.6.0 -f Dockerfile.production .

# 构建Web端镜像
cd ../telegram-web
docker build -t zhihang-web:v1.6.0 -f Dockerfile.production .

# 2. 保存镜像
cd ..
docker save zhihang-backend:v1.6.0 -o backend.tar
docker save zhihang-admin:v1.6.0 -o admin.tar
docker save zhihang-web:v1.6.0 -o web.tar

# 保存基础镜像
docker save mysql:8.0 -o mysql.tar
docker save redis:7-alpine -o redis.tar
docker save nginx:alpine -o nginx.tar
docker save minio/minio:latest -o minio.tar

# 3. 打包所有镜像
tar czf docker-images-v1.6.0.tar.gz *.tar

# 4. 上传到服务器
scp docker-images-v1.6.0.tar.gz root@154.37.214.191:/tmp/

# 5. 在服务器加载
ssh root@154.37.214.191
cd /tmp
tar xzf docker-images-v1.6.0.tar.gz
docker load -i mysql.tar
docker load -i redis.tar
docker load -i nginx.tar
docker load -i minio.tar
docker load -i backend.tar
docker load -i admin.tar
docker load -i web.tar

# 6. 验证
docker images

# 7. 启动服务
cd /path/to/im-suite
docker-compose up -d
```

### 方案B: 二进制部署（无Docker）

如果Docker问题无法解决，可以使用二进制部署：

详见: `docs/deployment/BINARY_DEPLOYMENT.md`（需要创建）

---

## 📊 问题优先级

### 高优先级（立即解决）

1. ⚠️ **检查云服务器安全组** - 最可能的原因
2. ⚠️ **修复DNS配置** - 基础网络问题
3. ⚠️ **配置Docker镜像源** - 加速访问

### 中优先级（如果上述失败）

4. 📋 使用阿里云镜像加速
5. 📋 配置HTTP代理

### 低优先级（最后手段）

6. 📋 手动上传镜像
7. 📋 二进制部署

---

## ✅ 验证修复结果

### 验证清单

```bash
# 1. DNS验证
ping -c 4 8.8.8.8                    # ✅ 应该成功
ping -c 4 docker.io                  # ✅ 应该成功
nslookup hub.docker.com              # ✅ 应该返回IP

# 2. 网络验证
curl -I https://hub.docker.com       # ✅ 应该返回200
curl -I https://registry-1.docker.io # ✅ 应该返回200/401

# 3. Docker验证
docker pull alpine:latest            # ✅ 应该成功下载
docker run --rm alpine ping -c 2 8.8.8.8  # ✅ 应该成功

# 4. 镜像源验证
docker info | grep -A 5 "Registry Mirrors"  # ✅ 应该显示镜像源

# 5. 完整测试
docker pull mysql:8.0                # ✅ 应该成功
docker pull redis:7-alpine           # ✅ 应该成功
```

### 成功标准

✅ DNS解析正常  
✅ HTTPS连接正常  
✅ Docker镜像拉取成功  
✅ 所有基础镜像下载完成  

---

## 📞 如果仍然失败

### 联系云服务商技术支持

**阿里云**:
- 电话: 95187
- 工单系统: https://workorder.console.aliyun.com

**腾讯云**:
- 电话: 4009100100
- 工单系统: https://console.cloud.tencent.com/workorder

**AWS**:
- Support Center: https://console.aws.amazon.com/support

### 提供的信息

```
问题描述: Docker无法拉取镜像
服务器IP: 154.37.214.191
错误信息: TLS handshake timeout, DNS resolution failed
已尝试: [列出已尝试的方案]
```

---

## 📝 记录和报告

### 执行日志模板

```markdown
## 网络修复执行记录

**执行日期**: 2024-12-19
**执行人**: Devin
**服务器**: 154.37.214.191

### 执行的步骤:
1. [ ] 检查DNS配置
2. [ ] 修复DNS服务器
3. [ ] 配置Docker镜像源
4. [ ] 重启Docker服务
5. [ ] 检查安全组规则
6. [ ] 测试镜像拉取

### 遇到的问题:
- 

### 解决方案:
- 

### 最终结果:
- [ ] 成功
- [ ] 失败（原因：）

### 下一步:
- 
```

---

## 🎯 总结

### 最可能的原因

根据症状分析，最可能的原因是：

1. **云服务器安全组限制** (80%可能性)
   - 出站规则未开放HTTPS(443)
   - 出站规则未开放DNS(53)

2. **DNS配置问题** (15%可能性)
   - DNS服务器不可用
   - 配置文件被覆盖

3. **网络路由问题** (5%可能性)
   - ISP限制
   - 防火墙拦截

### 推荐的修复顺序

```
1. 检查云服务器安全组（5分钟）⭐⭐⭐⭐⭐
   ↓ 如果失败
2. 修复DNS配置（2分钟）⭐⭐⭐⭐
   ↓ 如果失败
3. 配置Docker镜像源（3分钟）⭐⭐⭐
   ↓ 如果失败
4. 手动上传镜像（30分钟）⭐⭐
   ↓ 如果失败
5. 联系云服务商支持 ⭐
```

### 预计解决时间

- 最快: 5分钟（安全组配置）
- 一般: 15分钟（DNS + 镜像源）
- 最慢: 60分钟（手动上传镜像）

---

**重要**: 建议Devin首先检查云服务器控制台的**安全组出站规则**，这是最常见的原因！

---

**最后更新**: 2024-12-19  
**适用版本**: 所有版本  
**优先级**: 🔴 高

