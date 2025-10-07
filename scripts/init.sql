-- 志航密信数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS zhihang_messenger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE zhihang_messenger;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    -- 基本信息
    phone VARCHAR(20) NOT NULL UNIQUE,
    username VARCHAR(50) UNIQUE,
    nickname VARCHAR(100),
    bio TEXT,
    avatar VARCHAR(500),
    
    -- 认证信息
    password VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,
    
    -- 状态信息
    is_active BOOLEAN DEFAULT TRUE,
    last_seen DATETIME(3),
    online BOOLEAN DEFAULT FALSE,
    
    -- 设置信息
    language VARCHAR(10) DEFAULT 'zh-CN',
    theme VARCHAR(20) DEFAULT 'auto',
    
    -- 索引
    INDEX idx_users_phone (phone),
    INDEX idx_users_username (username),
    INDEX idx_users_online (online),
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建联系人表
CREATE TABLE IF NOT EXISTS contacts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    user_id BIGINT UNSIGNED NOT NULL,
    contact_id BIGINT UNSIGNED NOT NULL,
    nickname VARCHAR(100),
    is_blocked BOOLEAN DEFAULT FALSE,
    is_muted BOOLEAN DEFAULT FALSE,
    
    -- 索引
    INDEX idx_contacts_user_id (user_id),
    INDEX idx_contacts_contact_id (contact_id),
    INDEX idx_contacts_deleted_at (deleted_at),
    UNIQUE KEY uk_contacts_user_contact (user_id, contact_id),
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contact_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建聊天表
CREATE TABLE IF NOT EXISTS chats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    -- 基本信息
    name VARCHAR(200),
    description TEXT,
    avatar VARCHAR(500),
    type ENUM('private', 'group', 'channel') DEFAULT 'private',
    
    -- 设置信息
    is_active BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_muted BOOLEAN DEFAULT FALSE,
    
    -- 索引
    INDEX idx_chats_type (type),
    INDEX idx_chats_is_active (is_active),
    INDEX idx_chats_updated_at (updated_at),
    INDEX idx_chats_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建聊天成员表
CREATE TABLE IF NOT EXISTS chat_members (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    chat_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    role ENUM('owner', 'admin', 'member') DEFAULT 'member',
    joined_at DATETIME(3) NOT NULL,
    
    -- 索引
    INDEX idx_chat_members_chat_id (chat_id),
    INDEX idx_chat_members_user_id (user_id),
    INDEX idx_chat_members_role (role),
    INDEX idx_chat_members_deleted_at (deleted_at),
    UNIQUE KEY uk_chat_members_chat_user (chat_id, user_id),
    
    -- 外键约束
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建消息表
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    -- 基本信息
    chat_id BIGINT UNSIGNED NOT NULL,
    sender_id BIGINT UNSIGNED NOT NULL,
    content TEXT,
    type ENUM('text', 'image', 'video', 'file', 'audio', 'voice', 'sticker', 'gif') DEFAULT 'text',
    
    -- 文件信息
    file_name VARCHAR(255),
    file_size BIGINT,
    file_url VARCHAR(500),
    thumbnail VARCHAR(500),
    
    -- 消息状态
    is_read BOOLEAN DEFAULT FALSE,
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    is_pinned BOOLEAN DEFAULT FALSE,
    
    -- 特殊功能
    reply_to_id BIGINT UNSIGNED NULL,
    forward_from BIGINT UNSIGNED NULL,
    ttl INT DEFAULT 0,
    send_at DATETIME(3) NULL,
    is_silent BOOLEAN DEFAULT FALSE,
    
    -- 索引
    INDEX idx_messages_chat_id (chat_id),
    INDEX idx_messages_sender_id (sender_id),
    INDEX idx_messages_type (type),
    INDEX idx_messages_created_at (created_at),
    INDEX idx_messages_is_deleted (is_deleted),
    INDEX idx_messages_is_pinned (is_pinned),
    INDEX idx_messages_deleted_at (deleted_at),
    
    -- 外键约束
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL,
    FOREIGN KEY (forward_from) REFERENCES messages(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建消息已读表
CREATE TABLE IF NOT EXISTS message_reads (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    
    message_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    read_at DATETIME(3) NOT NULL,
    
    -- 索引
    INDEX idx_message_reads_message_id (message_id),
    INDEX idx_message_reads_user_id (user_id),
    INDEX idx_message_reads_read_at (read_at),
    INDEX idx_message_reads_deleted_at (deleted_at),
    UNIQUE KEY uk_message_reads_message_user (message_id, user_id),
    
    -- 外键约束
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认管理员用户
INSERT INTO users (created_at, updated_at, phone, username, nickname, password, salt, is_active, last_seen, online, language, theme) 
VALUES (
    NOW(), 
    NOW(), 
    '13800138000', 
    'admin', 
    '系统管理员', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: password
    'default_salt', 
    TRUE, 
    NOW(), 
    FALSE, 
    'zh-CN', 
    'auto'
) ON DUPLICATE KEY UPDATE updated_at = NOW();

-- 创建默认聊天室
INSERT INTO chats (created_at, updated_at, name, description, type, is_active, is_pinned, is_muted)
VALUES (
    NOW(),
    NOW(),
    '欢迎使用志航密信',
    '这是系统默认聊天室，欢迎您的加入！',
    'group',
    TRUE,
    TRUE,
    FALSE
) ON DUPLICATE KEY UPDATE updated_at = NOW();

-- 将管理员添加到默认聊天室
INSERT INTO chat_members (created_at, updated_at, chat_id, user_id, role, joined_at)
SELECT 
    NOW(),
    NOW(),
    c.id,
    u.id,
    'owner',
    NOW()
FROM chats c, users u 
WHERE c.name = '欢迎使用志航密信' 
  AND u.username = 'admin'
  AND NOT EXISTS (
    SELECT 1 FROM chat_members cm 
    WHERE cm.chat_id = c.id AND cm.user_id = u.id
  );

-- 创建欢迎消息
INSERT INTO messages (created_at, updated_at, chat_id, sender_id, content, type, is_read, is_edited, is_deleted, is_pinned, is_silent)
SELECT 
    NOW(),
    NOW(),
    c.id,
    u.id,
    '欢迎使用志航密信！这是一个安全、快速的即时通讯系统。',
    'text',
    FALSE,
    FALSE,
    FALSE,
    TRUE,
    FALSE
FROM chats c, users u 
WHERE c.name = '欢迎使用志航密信' 
  AND u.username = 'admin'
  AND NOT EXISTS (
    SELECT 1 FROM messages m 
    WHERE m.chat_id = c.id AND m.sender_id = u.id AND m.is_pinned = TRUE
  );

-- 创建数据库用户（如果不存在）
CREATE USER IF NOT EXISTS 'zhihang_messenger'@'%' IDENTIFIED BY 'zhihang_messenger_pass';
GRANT ALL PRIVILEGES ON zhihang_messenger.* TO 'zhihang_messenger'@'%';
FLUSH PRIVILEGES;

-- 显示创建结果
SELECT 'Database initialization completed successfully!' as status;
SELECT COUNT(*) as user_count FROM users;
SELECT COUNT(*) as chat_count FROM chats;
SELECT COUNT(*) as message_count FROM messages;