# IM-Suite 数据库设计文档

## 概述

IM-Suite 使用 MySQL 8.0 作为主数据库，Redis 作为缓存，MinIO 作为文件存储。本文档描述了完整的数据库表结构和关系。

## 数据库配置

### MySQL 配置
```sql
-- 创建数据库
CREATE DATABASE im_suite CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'im_user'@'%' IDENTIFIED BY 'im_password';
GRANT ALL PRIVILEGES ON im_suite.* TO 'im_user'@'%';
FLUSH PRIVILEGES;
```

### Redis 配置
```bash
# 缓存配置
redis-cli CONFIG SET maxmemory 256mb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

## 表结构设计

### 1. 用户表 (users)

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    phone VARCHAR(20) NOT NULL UNIQUE COMMENT '手机号',
    username VARCHAR(50) UNIQUE COMMENT '用户名',
    nickname VARCHAR(100) NOT NULL COMMENT '昵称',
    avatar_url VARCHAR(500) COMMENT '头像URL',
    bio TEXT COMMENT '个人简介',
    is_online BOOLEAN DEFAULT FALSE COMMENT '是否在线',
    last_seen TIMESTAMP NULL COMMENT '最后上线时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_phone (phone),
    INDEX idx_username (username),
    INDEX idx_online (is_online),
    INDEX idx_last_seen (last_seen)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

### 2. 联系人表 (contacts)

```sql
CREATE TABLE contacts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    contact_id BIGINT NOT NULL COMMENT '联系人用户ID',
    nickname VARCHAR(100) COMMENT '联系人昵称',
    remark VARCHAR(200) COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
    
    UNIQUE KEY uk_user_contact (user_id, contact_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contact_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_contact_id (contact_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='联系人表';
```

### 3. 聊天表 (chats)

```sql
CREATE TABLE chats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    type ENUM('private', 'group', 'channel') NOT NULL DEFAULT 'private' COMMENT '聊天类型',
    title VARCHAR(200) NOT NULL COMMENT '聊天标题',
    description TEXT COMMENT '聊天描述',
    avatar_url VARCHAR(500) COMMENT '聊天头像',
    created_by BIGINT NOT NULL COMMENT '创建者用户ID',
    last_message_id BIGINT NULL COMMENT '最后一条消息ID',
    unread_count INT DEFAULT 0 COMMENT '未读消息数',
    is_pinned BOOLEAN DEFAULT FALSE COMMENT '是否置顶',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_created_by (created_by),
    INDEX idx_type (type),
    INDEX idx_pinned (is_pinned),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天表';
```

### 4. 聊天成员表 (chat_members)

```sql
CREATE TABLE chat_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    chat_id BIGINT NOT NULL COMMENT '聊天ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    role ENUM('admin', 'member') DEFAULT 'member' COMMENT '角色',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    
    UNIQUE KEY uk_chat_user (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_chat_id (chat_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天成员表';
```

### 5. 消息表 (messages)

```sql
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    chat_id BIGINT NOT NULL COMMENT '聊天ID',
    sender_id BIGINT NOT NULL COMMENT '发送者用户ID',
    content TEXT NOT NULL COMMENT '消息内容',
    message_type ENUM('text', 'image', 'video', 'audio', 'voice', 'file', 'sticker', 'gif') DEFAULT 'text' COMMENT '消息类型',
    file_url VARCHAR(500) COMMENT '文件URL',
    file_name VARCHAR(200) COMMENT '文件名',
    file_size BIGINT COMMENT '文件大小（字节）',
    mime_type VARCHAR(100) COMMENT 'MIME类型',
    reply_to_id BIGINT NULL COMMENT '回复的消息ID',
    is_edited BOOLEAN DEFAULT FALSE COMMENT '是否已编辑',
    is_deleted BOOLEAN DEFAULT FALSE COMMENT '是否已删除',
    ttl_seconds INT DEFAULT 0 COMMENT '阅后即焚时间（秒）',
    send_at TIMESTAMP NULL COMMENT '发送时间',
    is_silent BOOLEAN DEFAULT FALSE COMMENT '是否静默发送',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL,
    INDEX idx_chat_id (chat_id),
    INDEX idx_sender_id (sender_id),
    INDEX idx_created_at (created_at),
    INDEX idx_message_type (message_type),
    INDEX idx_ttl_seconds (ttl_seconds)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';
```

### 6. 消息已读表 (message_reads)

```sql
CREATE TABLE message_reads (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id BIGINT NOT NULL COMMENT '消息ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '已读时间',
    
    UNIQUE KEY uk_message_user (message_id, user_id),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_message_id (message_id),
    INDEX idx_user_id (user_id),
    INDEX idx_read_at (read_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息已读表';
```

### 7. 通话记录表 (calls)

```sql
CREATE TABLE calls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    call_id VARCHAR(100) NOT NULL UNIQUE COMMENT '通话ID',
    from_user BIGINT NOT NULL COMMENT '发起者用户ID',
    to_user BIGINT NOT NULL COMMENT '接收者用户ID',
    call_type ENUM('audio', 'video') NOT NULL COMMENT '通话类型',
    status ENUM('pending', 'answered', 'rejected', 'ended', 'missed') DEFAULT 'pending' COMMENT '通话状态',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    ended_at TIMESTAMP NULL COMMENT '结束时间',
    duration INT DEFAULT 0 COMMENT '通话时长（秒）',
    reason VARCHAR(100) COMMENT '结束原因',
    
    FOREIGN KEY (from_user) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_call_id (call_id),
    INDEX idx_from_user (from_user),
    INDEX idx_to_user (to_user),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通话记录表';
```

### 8. 文件表 (files)

```sql
CREATE TABLE files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    file_name VARCHAR(200) NOT NULL COMMENT '文件名',
    file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
    file_url VARCHAR(500) NOT NULL COMMENT '文件URL',
    file_size BIGINT NOT NULL COMMENT '文件大小（字节）',
    mime_type VARCHAR(100) NOT NULL COMMENT 'MIME类型',
    file_hash VARCHAR(64) COMMENT '文件哈希值',
    uploaded_by BIGINT NOT NULL COMMENT '上传者用户ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
    
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_file_hash (file_hash),
    INDEX idx_uploaded_by (uploaded_by),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';
```

### 9. 系统配置表 (system_configs)

```sql
CREATE TABLE system_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(100) NOT NULL UNIQUE COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    description VARCHAR(200) COMMENT '配置描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';
```

### 10. 操作日志表 (operation_logs)

```sql
CREATE TABLE operation_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT COMMENT '用户ID',
    operation VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) COMMENT '资源类型',
    resource_id BIGINT COMMENT '资源ID',
    description TEXT COMMENT '操作描述',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent VARCHAR(500) COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_operation (operation),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
```

## 索引优化

### 1. 复合索引
```sql
-- 聊天消息查询优化
CREATE INDEX idx_chat_created_at ON messages(chat_id, created_at DESC);

-- 用户联系人查询优化
CREATE INDEX idx_user_contact_created ON contacts(user_id, created_at DESC);

-- 消息已读查询优化
CREATE INDEX idx_message_user_read ON message_reads(message_id, user_id, read_at);
```

### 2. 分区表
```sql
-- 按时间分区消息表（按月分区）
ALTER TABLE messages PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
    PARTITION p202510 VALUES LESS THAN (202511),
    PARTITION p202511 VALUES LESS THAN (202512),
    PARTITION p202512 VALUES LESS THAN (202601),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
```

## Redis 缓存设计

### 1. 用户会话缓存
```
# 用户在线状态
user:online:{user_id} -> "true" (TTL: 300s)

# 用户最后上线时间
user:last_seen:{user_id} -> timestamp (TTL: 3600s)

# 用户信息缓存
user:info:{user_id} -> JSON (TTL: 1800s)
```

### 2. 聊天缓存
```
# 聊天信息缓存
chat:info:{chat_id} -> JSON (TTL: 1800s)

# 聊天成员缓存
chat:members:{chat_id} -> Set (TTL: 1800s)

# 聊天未读消息数
chat:unread:{chat_id}:{user_id} -> count (TTL: 3600s)
```

### 3. 消息缓存
```
# 最新消息缓存
chat:messages:{chat_id}:latest -> List (TTL: 3600s)

# 消息搜索缓存
search:messages:{query_hash} -> List (TTL: 1800s)
```

### 4. 实时通讯缓存
```
# WebSocket 连接映射
ws:connection:{user_id} -> connection_id (TTL: 3600s)

# 正在输入状态
typing:{chat_id}:{user_id} -> timestamp (TTL: 30s)

# 通话状态
call:status:{call_id} -> JSON (TTL: 3600s)
```

## 数据迁移脚本

### 1. 初始化数据
```sql
-- 插入系统配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
('app_name', 'IM-Suite', '应用名称'),
('app_version', '1.0.0', '应用版本'),
('max_file_size', '104857600', '最大文件大小（字节）'),
('max_message_length', '4096', '最大消息长度'),
('message_ttl_max', '86400', '最大阅后即焚时间（秒）'),
('call_timeout', '30', '通话超时时间（秒）');

-- 插入测试用户
INSERT INTO users (phone, username, nickname, bio) VALUES
('+8613800000000', 'admin', '管理员', '系统管理员'),
('+8613800000001', 'user1', '用户1', '测试用户1'),
('+8613800000002', 'user2', '用户2', '测试用户2');
```

### 2. 数据清理脚本
```sql
-- 清理过期的阅后即焚消息
DELETE FROM messages 
WHERE ttl_seconds > 0 
AND created_at < DATE_SUB(NOW(), INTERVAL ttl_seconds SECOND);

-- 清理过期的操作日志（保留30天）
DELETE FROM operation_logs 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 清理已删除的消息
DELETE FROM messages 
WHERE is_deleted = TRUE 
AND updated_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
```

## 性能优化建议

### 1. 查询优化
- 使用适当的索引
- 避免 SELECT * 查询
- 使用 LIMIT 限制结果集
- 使用 EXPLAIN 分析查询计划

### 2. 缓存策略
- 热点数据缓存
- 使用 Redis 集群
- 设置合理的 TTL
- 实现缓存预热

### 3. 分库分表
- 按用户 ID 分片
- 按时间分片
- 使用中间件管理分片

### 4. 监控指标
- 查询响应时间
- 连接池使用率
- 缓存命中率
- 磁盘 I/O 使用率

## 备份策略

### 1. 全量备份
```bash
# 每日全量备份
mysqldump -u im_user -p im_suite > backup_$(date +%Y%m%d).sql
```

### 2. 增量备份
```bash
# 使用 binlog 进行增量备份
mysqlbinlog --start-datetime="2025-10-07 00:00:00" mysql-bin.000001 > incremental_backup.sql
```

### 3. 恢复策略
```bash
# 恢复全量备份
mysql -u im_user -p im_suite < backup_20251007.sql

# 恢复增量备份
mysql -u im_user -p im_suite < incremental_backup.sql
```

## 安全考虑

### 1. 数据加密
- 敏感字段加密存储
- 使用 AES-256 加密
- 密钥轮换策略

### 2. 访问控制
- 最小权限原则
- 定期审计权限
- 使用连接池

### 3. 数据脱敏
- 生产数据脱敏
- 测试环境数据隔离
- 日志脱敏处理
