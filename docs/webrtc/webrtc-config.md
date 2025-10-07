# IM-Suite WebRTC 配置文档

## 概述

本文档描述了 IM-Suite 中 WebRTC 的配置选项、服务器设置和优化建议。

## 服务器配置

### 1. STUN 服务器

#### 公共 STUN 服务器
```javascript
const iceServers = [
  { urls: 'stun:stun.l.google.com:19302' },
  { urls: 'stun:stun1.l.google.com:19302' },
  { urls: 'stun:stun2.l.google.com:19302' },
  { urls: 'stun:stun3.l.google.com:19302' },
  { urls: 'stun:stun4.l.google.com:19302' },
  { urls: 'stun:stun.ekiga.net' },
  { urls: 'stun:stun.ideasip.com' },
  { urls: 'stun:stun.schlund.de' },
  { urls: 'stun:stun.stunprotocol.org:3478' },
  { urls: 'stun:stun.voiparound.com' },
  { urls: 'stun:stun.voipbuster.com' },
  { urls: 'stun:stun.voipstunt.com' },
  { urls: 'stun:stun.voxgratia.org' },
  { urls: 'stun:stun.xten.com' }
];
```

#### 自建 STUN 服务器
```bash
# 使用 coturn 搭建 STUN/TURN 服务器
sudo apt-get install coturn

# 配置文件 /etc/turnserver.conf
listening-port=3478
tls-listening-port=5349
listening-ip=0.0.0.0
external-ip=YOUR_PUBLIC_IP
realm=im-suite.com
server-name=im-suite.com
user=im-user:im-password
user=im-user2:im-password2
cert=/etc/ssl/certs/turn.crt
pkey=/etc/ssl/private/turn.key
log-file=/var/log/turnserver.log
verbose
```

### 2. TURN 服务器

#### 自建 TURN 服务器配置
```bash
# 安装 coturn
sudo apt-get update
sudo apt-get install coturn

# 启用 coturn 服务
sudo systemctl enable coturn
sudo systemctl start coturn

# 配置 TURN 服务器
sudo nano /etc/turnserver.conf
```

```ini
# TURN 服务器配置
listening-port=3478
tls-listening-port=5349
listening-ip=0.0.0.0
external-ip=YOUR_PUBLIC_IP
realm=im-suite.com
server-name=im-suite.com

# 用户认证
user=im-user:im-password
user=im-user2:im-password2

# SSL 证书
cert=/etc/ssl/certs/turn.crt
pkey=/etc/ssl/private/turn.key

# 日志配置
log-file=/var/log/turnserver.log
verbose

# 安全配置
no-multicast-peers
no-cli
no-tlsv1
no-tlsv1_1
no-tlsv1_2
no-tlsv1_3

# 带宽限制
max-bps=1000000
total-quota=1000000

# 会话超时
user-timeout=3600
session-timeout=3600
```

#### 客户端配置
```javascript
const iceServers = [
  { urls: 'stun:stun.l.google.com:19302' },
  { 
    urls: 'turn:your-turn-server.com:3478',
    username: 'im-user',
    credential: 'im-password'
  },
  { 
    urls: 'turns:your-turn-server.com:5349',
    username: 'im-user',
    credential: 'im-password'
  }
];
```

## 媒体配置

### 1. 音频配置

#### 音频编码器
```javascript
const audioConstraints = {
  audio: {
    echoCancellation: true,
    noiseSuppression: true,
    autoGainControl: true,
    sampleRate: 48000,
    sampleSize: 16,
    channels: 2
  }
};
```

#### 音频编解码器优先级
```javascript
const audioCodecs = [
  'opus',      // 最佳质量，低延迟
  'PCMU',      // 兼容性好
  'PCMA',      // 兼容性好
  'G722',      // 中等质量
  'G729'       // 低带宽
];
```

### 2. 视频配置

#### 视频分辨率配置
```javascript
const videoConstraints = {
  video: {
    width: { ideal: 1280, max: 1920 },
    height: { ideal: 720, max: 1080 },
    frameRate: { ideal: 30, max: 60 },
    aspectRatio: 16/9
  }
};
```

#### 视频编解码器优先级
```javascript
const videoCodecs = [
  'VP8',       // 开源，兼容性好
  'VP9',       // 高质量，低带宽
  'H264',      // 硬件加速支持好
  'H265'       // 最新标准，高压缩率
];
```

### 3. 带宽配置

#### 带宽限制
```javascript
const bandwidthConstraints = {
  // 音频带宽
  audio: {
    min: 64000,    // 64 kbps
    max: 128000    // 128 kbps
  },
  
  // 视频带宽
  video: {
    min: 200000,   // 200 kbps
    max: 2000000   // 2 Mbps
  },
  
  // 总带宽
  total: {
    min: 300000,   // 300 kbps
    max: 3000000   // 3 Mbps
  }
};
```

## 网络配置

### 1. ICE 配置

#### ICE 候选收集
```javascript
const iceConfig = {
  iceServers: iceServers,
  iceCandidatePoolSize: 10,
  bundlePolicy: 'max-bundle',
  rtcpMuxPolicy: 'require',
  iceTransportPolicy: 'all'
};
```

#### ICE 连接检查
```javascript
const iceConnectionStates = {
  'new': '新连接',
  'checking': '检查中',
  'connected': '已连接',
  'completed': '完成',
  'failed': '失败',
  'disconnected': '断开',
  'closed': '关闭'
};
```

### 2. 网络质量监控

#### 网络统计
```javascript
function getNetworkStats(peerConnection) {
  return peerConnection.getStats().then(stats => {
    const statsMap = new Map();
    stats.forEach(report => {
      statsMap.set(report.id, report);
    });
    
    return {
      audio: getAudioStats(statsMap),
      video: getVideoStats(statsMap),
      network: getNetworkStats(statsMap)
    };
  });
}

function getAudioStats(statsMap) {
  const audioStats = {};
  statsMap.forEach(report => {
    if (report.type === 'inbound-rtp' && report.mediaType === 'audio') {
      audioStats.packetsReceived = report.packetsReceived;
      audioStats.packetsLost = report.packetsLost;
      audioStats.bytesReceived = report.bytesReceived;
      audioStats.jitter = report.jitter;
    }
  });
  return audioStats;
}

function getVideoStats(statsMap) {
  const videoStats = {};
  statsMap.forEach(report => {
    if (report.type === 'inbound-rtp' && report.mediaType === 'video') {
      videoStats.packetsReceived = report.packetsReceived;
      videoStats.packetsLost = report.packetsLost;
      videoStats.bytesReceived = report.bytesReceived;
      videoStats.frameWidth = report.frameWidth;
      videoStats.frameHeight = report.frameHeight;
      videoStats.framesPerSecond = report.framesPerSecond;
    }
  });
  return videoStats;
}

function getNetworkStats(statsMap) {
  const networkStats = {};
  statsMap.forEach(report => {
    if (report.type === 'candidate-pair' && report.state === 'succeeded') {
      networkStats.rtt = report.currentRoundTripTime;
      networkStats.availableOutgoingBitrate = report.availableOutgoingBitrate;
    }
  });
  return networkStats;
}
```

## 安全配置

### 1. 传输安全

#### DTLS-SRTP 配置
```javascript
const securityConfig = {
  // 强制使用 DTLS
  dtls: {
    enabled: true,
    require: true
  },
  
  // SRTP 配置
  srtp: {
    enabled: true,
    require: true
  },
  
  // 证书配置
  certificates: {
    // 使用自签名证书
    selfSigned: true,
    // 或使用 CA 签发的证书
    caSigned: false
  }
};
```

### 2. 端到端加密

#### 媒体加密
```javascript
const encryptionConfig = {
  // 启用端到端加密
  e2ee: {
    enabled: true,
    algorithm: 'AES-256-GCM',
    keyExchange: 'X25519'
  },
  
  // 密钥管理
  keyManagement: {
    rotation: true,
    rotationInterval: 3600000, // 1小时
    keySize: 256
  }
};
```

## 性能优化

### 1. 硬件加速

#### 视频编码器选择
```javascript
const hardwareAcceleration = {
  // 优先使用硬件编码器
  video: {
    hardware: true,
    codecs: ['H264', 'VP8', 'VP9'],
    fallback: 'software'
  },
  
  // 音频处理
  audio: {
    hardware: true,
    processing: 'gpu',
    fallback: 'cpu'
  }
};
```

### 2. 自适应质量

#### 质量调整
```javascript
function adjustQuality(networkStats) {
  const { rtt, packetLoss, bandwidth } = networkStats;
  
  if (rtt > 200 || packetLoss > 0.05) {
    // 网络质量差，降低质量
    return {
      video: {
        width: 640,
        height: 480,
        frameRate: 15,
        bitrate: 500000
      },
      audio: {
        bitrate: 32000,
        sampleRate: 16000
      }
    };
  } else if (rtt < 50 && packetLoss < 0.01) {
    // 网络质量好，提高质量
    return {
      video: {
        width: 1280,
        height: 720,
        frameRate: 30,
        bitrate: 2000000
      },
      audio: {
        bitrate: 128000,
        sampleRate: 48000
      }
    };
  }
  
  // 默认质量
  return {
    video: {
      width: 854,
      height: 480,
      frameRate: 24,
      bitrate: 1000000
    },
    audio: {
      bitrate: 64000,
      sampleRate: 48000
    }
  };
}
```

### 3. 带宽管理

#### 动态带宽调整
```javascript
function manageBandwidth(peerConnection, networkStats) {
  const { availableBandwidth, currentUsage } = networkStats;
  
  if (availableBandwidth < currentUsage * 1.2) {
    // 带宽不足，降低质量
    const constraints = {
      video: {
        width: { max: 640 },
        height: { max: 480 },
        frameRate: { max: 15 }
      }
    };
    
    peerConnection.getSenders().forEach(sender => {
      if (sender.track && sender.track.kind === 'video') {
        sender.applyConstraints(constraints.video);
      }
    });
  }
}
```

## 错误处理

### 1. 连接错误

#### 错误分类
```javascript
const errorTypes = {
  'NETWORK_ERROR': '网络连接错误',
  'MEDIA_ERROR': '媒体设备错误',
  'SIGNALING_ERROR': '信令处理错误',
  'PERMISSION_ERROR': '权限错误',
  'TIMEOUT_ERROR': '超时错误'
};
```

#### 错误处理策略
```javascript
function handleWebRTCError(error) {
  switch (error.name) {
    case 'NotAllowedError':
      // 权限被拒绝
      showPermissionDialog();
      break;
      
    case 'NotFoundError':
      // 设备未找到
      showDeviceNotFoundDialog();
      break;
      
    case 'NotReadableError':
      // 设备被占用
      showDeviceBusyDialog();
      break;
      
    case 'OverconstrainedError':
      // 约束条件无法满足
      adjustConstraints();
      break;
      
    case 'SecurityError':
      // 安全错误
      showSecurityErrorDialog();
      break;
      
    default:
      // 未知错误
      showGenericErrorDialog(error.message);
  }
}
```

### 2. 重连机制

#### 自动重连
```javascript
class WebRTCReconnection {
  constructor(peerConnection, maxRetries = 3) {
    this.peerConnection = peerConnection;
    this.maxRetries = maxRetries;
    this.retryCount = 0;
    this.retryTimeout = null;
  }
  
  startReconnection() {
    if (this.retryCount >= this.maxRetries) {
      console.error('重连次数超限');
      return;
    }
    
    this.retryCount++;
    const delay = Math.pow(2, this.retryCount) * 1000; // 指数退避
    
    this.retryTimeout = setTimeout(() => {
      this.attemptReconnection();
    }, delay);
  }
  
  attemptReconnection() {
    // 重新创建 PeerConnection
    // 重新发送信令
    // 重新建立连接
  }
  
  reset() {
    this.retryCount = 0;
    if (this.retryTimeout) {
      clearTimeout(this.retryTimeout);
      this.retryTimeout = null;
    }
  }
}
```

## 测试配置

### 1. 测试环境

#### 本地测试
```bash
# 启动本地 TURN 服务器
sudo turnserver -v -r im-suite.com -a -b /etc/turnserver.conf

# 测试 STUN 服务器
stun-client stun.l.google.com 19302

# 测试 TURN 服务器
turn-client your-turn-server.com 3478 im-user im-password
```

#### 网络测试
```javascript
// 网络连接测试
function testNetworkConnectivity() {
  const testConfig = {
    iceServers: [
      { urls: 'stun:stun.l.google.com:19302' }
    ]
  };
  
  const pc = new RTCPeerConnection(testConfig);
  
  pc.onicecandidate = (event) => {
    if (event.candidate) {
      console.log('ICE 候选:', event.candidate);
    }
  };
  
  pc.onicegatheringstatechange = () => {
    console.log('ICE 收集状态:', pc.iceGatheringState);
  };
  
  // 创建数据通道触发 ICE 收集
  pc.createDataChannel('test');
  pc.createOffer().then(offer => {
    pc.setLocalDescription(offer);
  });
}
```

### 2. 性能测试

#### 延迟测试
```javascript
function measureLatency(peerConnection) {
  const startTime = Date.now();
  
  peerConnection.createDataChannel('latency-test');
  peerConnection.ondatachannel = (event) => {
    const channel = event.channel;
    channel.onmessage = (event) => {
      const latency = Date.now() - startTime;
      console.log('延迟:', latency, 'ms');
    };
  };
}
```

#### 带宽测试
```javascript
function measureBandwidth(peerConnection) {
  setInterval(() => {
    peerConnection.getStats().then(stats => {
      stats.forEach(report => {
        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          console.log('可用带宽:', report.availableOutgoingBitrate, 'bps');
        }
      });
    });
  }, 1000);
}
```

## 部署配置

### 1. Docker 配置

#### TURN 服务器 Dockerfile
```dockerfile
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y coturn

COPY turnserver.conf /etc/turnserver.conf
COPY ssl/ /etc/ssl/

EXPOSE 3478 5349

CMD ["turnserver", "-v", "-r", "im-suite.com", "-a", "-b", "/etc/turnserver.conf"]
```

#### Docker Compose 配置
```yaml
version: '3.8'

services:
  turn-server:
    build: ./turn-server
    ports:
      - "3478:3478"
      - "5349:5349"
    volumes:
      - ./ssl:/etc/ssl
    environment:
      - TURN_USER=im-user
      - TURN_PASSWORD=im-password
    networks:
      - im_net

  stun-server:
    image: coturn/coturn
    ports:
      - "3478:3478/udp"
    command: ["-n", "--log-file=stdout", "--external-ip=auto"]
    networks:
      - im_net

networks:
  im_net:
    driver: bridge
```

### 2. 负载均衡

#### Nginx 配置
```nginx
upstream turn_servers {
    server turn1.im-suite.com:3478;
    server turn2.im-suite.com:3478;
    server turn3.im-suite.com:3478;
}

server {
    listen 80;
    server_name turn.im-suite.com;
    
    location / {
        proxy_pass http://turn_servers;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. 监控配置

#### 健康检查
```bash
#!/bin/bash
# turn-health-check.sh

TURN_SERVER="localhost:3478"
TURN_USER="im-user"
TURN_PASSWORD="im-password"

# 检查 TURN 服务器状态
turn-client $TURN_SERVER $TURN_USER $TURN_PASSWORD

if [ $? -eq 0 ]; then
    echo "TURN 服务器正常"
    exit 0
else
    echo "TURN 服务器异常"
    exit 1
fi
```

#### 监控脚本
```bash
#!/bin/bash
# turn-monitor.sh

while true; do
    ./turn-health-check.sh
    if [ $? -ne 0 ]; then
        # 发送告警
        curl -X POST "https://api.im-suite.com/alerts" \
          -H "Content-Type: application/json" \
          -d '{"type": "turn_server_down", "message": "TURN 服务器异常"}'
    fi
    sleep 60
done
```

## 最佳实践

### 1. 开发建议
- 使用最新的 WebRTC API
- 实现完整的错误处理
- 添加网络质量监控
- 支持多种编解码器
- 实现自适应质量调整

### 2. 部署建议
- 使用多个 TURN 服务器
- 配置负载均衡
- 监控服务器状态
- 定期更新证书
- 备份配置文件

### 3. 安全建议
- 使用强密码
- 定期轮换密钥
- 启用日志记录
- 限制访问权限
- 使用 HTTPS/WSS
