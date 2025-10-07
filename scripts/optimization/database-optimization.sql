-- 志航密信数据库性能优化脚本
-- 用于优化数据库查询性能和存储效率

-- ==============================================
-- 1. 索引优化
-- ==============================================

-- 用户表索引优化
-- 为常用查询字段添加索引
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_online_status ON users(online_status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_users_phone_username ON users(phone, username);
CREATE INDEX IF NOT EXISTS idx_users_status_last_seen ON users(online_status, last_seen);

-- 聊天表索引优化
CREATE INDEX IF NOT EXISTS idx_chats_type ON chats(type);
CREATE INDEX IF NOT EXISTS idx_chats_is_active ON chats(is_active);
CREATE INDEX IF NOT EXISTS idx_chats_updated_at ON chats(updated_at);
CREATE INDEX IF NOT EXISTS idx_chats_created_at ON chats(created_at);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_chats_type_active ON chats(type, is_active);
CREATE INDEX IF NOT EXISTS idx_chats_active_updated ON chats(is_active, updated_at);

-- 聊天成员表索引优化
CREATE INDEX IF NOT EXISTS idx_chat_members_chat_id ON chat_members(chat_id);
CREATE INDEX IF NOT EXISTS idx_chat_members_user_id ON chat_members(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_members_role ON chat_members(role);
CREATE INDEX IF NOT EXISTS idx_chat_members_joined_at ON chat_members(joined_at);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_chat_members_chat_user ON chat_members(chat_id, user_id);
CREATE INDEX IF NOT EXISTS idx_chat_members_user_chat ON chat_members(user_id, chat_id);

-- 消息表索引优化
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_type ON messages(type);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_is_read ON messages(is_read);
CREATE INDEX IF NOT EXISTS idx_messages_is_deleted ON messages(is_deleted);
CREATE INDEX IF NOT EXISTS idx_messages_is_pinned ON messages(is_pinned);

-- 复合索引优化（重要查询）
CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_chat_read ON messages(chat_id, is_read);
CREATE INDEX IF NOT EXISTS idx_messages_chat_deleted ON messages(chat_id, is_deleted);
CREATE INDEX IF NOT EXISTS idx_messages_sender_created ON messages(sender_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_type_created ON messages(type, created_at);

-- 消息已读表索引优化
CREATE INDEX IF NOT EXISTS idx_message_reads_message_id ON message_reads(message_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_user_id ON message_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_read_at ON message_reads(read_at);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_message_reads_message_user ON message_reads(message_id, user_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_user_read_at ON message_reads(user_id, read_at);

-- 联系人表索引优化
CREATE INDEX IF NOT EXISTS idx_contacts_user_id ON contacts(user_id);
CREATE INDEX IF NOT EXISTS idx_contacts_contact_id ON contacts(contact_id);
CREATE INDEX IF NOT EXISTS idx_contacts_is_blocked ON contacts(is_blocked);
CREATE INDEX IF NOT EXISTS idx_contacts_is_muted ON contacts(is_muted);
CREATE INDEX IF NOT EXISTS idx_contacts_created_at ON contacts(created_at);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_contacts_user_contact ON contacts(user_id, contact_id);
CREATE INDEX IF NOT EXISTS idx_contacts_user_blocked ON contacts(user_id, is_blocked);

-- ==============================================
-- 2. 查询优化
-- ==============================================

-- 优化用户查询
-- 使用覆盖索引优化用户信息查询
CREATE INDEX IF NOT EXISTS idx_users_cover_phone ON users(phone, id, username, nickname, avatar, online_status);

-- 优化聊天列表查询
-- 为聊天列表查询创建覆盖索引
CREATE INDEX IF NOT EXISTS idx_chats_cover_list ON chats(id, name, avatar, type, is_active, updated_at, members_count);

-- 优化消息列表查询
-- 为消息列表查询创建覆盖索引
CREATE INDEX IF NOT EXISTS idx_messages_cover_list ON messages(id, chat_id, sender_id, content, type, created_at, is_read, is_deleted);

-- ==============================================
-- 3. 分区优化
-- ==============================================

-- 消息表按月分区（适用于大量消息的场景）
-- 注意：MySQL 8.0+ 支持分区，需要根据实际需求调整

-- 创建消息表分区（示例）
-- ALTER TABLE messages PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
--     PARTITION p202401 VALUES LESS THAN (202402),
--     PARTITION p202402 VALUES LESS THAN (202403),
--     PARTITION p202403 VALUES LESS THAN (202404),
--     PARTITION p202404 VALUES LESS THAN (202405),
--     PARTITION p202405 VALUES LESS THAN (202406),
--     PARTITION p202406 VALUES LESS THAN (202407),
--     PARTITION p202407 VALUES LESS THAN (202408),
--     PARTITION p202408 VALUES LESS THAN (202409),
--     PARTITION p202409 VALUES LESS THAN (202410),
--     PARTITION p202410 VALUES LESS THAN (202411),
--     PARTITION p202411 VALUES LESS THAN (202412),
--     PARTITION p202412 VALUES LESS THAN (202501),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- ==============================================
-- 4. 存储优化
-- ==============================================

-- 优化表存储引擎
-- 确保使用 InnoDB 存储引擎（支持事务、行级锁定等）
ALTER TABLE users ENGINE=InnoDB;
ALTER TABLE chats ENGINE=InnoDB;
ALTER TABLE chat_members ENGINE=InnoDB;
ALTER TABLE messages ENGINE=InnoDB;
ALTER TABLE message_reads ENGINE=InnoDB;
ALTER TABLE contacts ENGINE=InnoDB;

-- 优化字符集和排序规则
-- 确保使用 utf8mb4 字符集支持完整的 Unicode
ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE chats CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE chat_members CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE messages CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE message_reads CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE contacts CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ==============================================
-- 5. 查询缓存优化
-- ==============================================

-- 启用查询缓存（MySQL 5.7 及以下版本）
-- 注意：MySQL 8.0 已移除查询缓存功能
-- SET GLOBAL query_cache_type = ON;
-- SET GLOBAL query_cache_size = 268435456; -- 256MB

-- ==============================================
-- 6. 连接池优化
-- ==============================================

-- 优化连接池设置
-- 这些设置需要在 MySQL 配置文件中设置
-- max_connections = 1000
-- max_connect_errors = 100000
-- connect_timeout = 10
-- wait_timeout = 28800
-- interactive_timeout = 28800
-- max_allowed_packet = 64M

-- ==============================================
-- 7. 慢查询优化
-- ==============================================

-- 启用慢查询日志
-- SET GLOBAL slow_query_log = ON;
-- SET GLOBAL long_query_time = 1;
-- SET GLOBAL log_queries_not_using_indexes = ON;

-- ==============================================
-- 8. 统计信息优化
-- ==============================================

-- 更新表统计信息
ANALYZE TABLE users;
ANALYZE TABLE chats;
ANALYZE TABLE chat_members;
ANALYZE TABLE messages;
ANALYZE TABLE message_reads;
ANALYZE TABLE contacts;

-- ==============================================
-- 9. 视图优化
-- ==============================================

-- 创建优化的视图
-- 用户在线状态视图
CREATE OR REPLACE VIEW v_user_online_status AS
SELECT 
    u.id,
    u.username,
    u.nickname,
    u.avatar,
    u.online_status,
    u.last_seen,
    CASE 
        WHEN u.online_status = 1 THEN '在线'
        WHEN u.last_seen > DATE_SUB(NOW(), INTERVAL 5 MINUTE) THEN '刚刚在线'
        WHEN u.last_seen > DATE_SUB(NOW(), INTERVAL 1 HOUR) THEN '1小时内在线'
        WHEN u.last_seen > DATE_SUB(NOW(), INTERVAL 1 DAY) THEN '1天内在线'
        ELSE '离线'
    END AS status_text
FROM users u
WHERE u.is_active = 1;

-- 聊天统计视图
CREATE OR REPLACE VIEW v_chat_stats AS
SELECT 
    c.id,
    c.name,
    c.type,
    c.members_count,
    COUNT(m.id) as message_count,
    MAX(m.created_at) as last_message_time,
    COUNT(CASE WHEN m.is_read = 0 THEN 1 END) as unread_count
FROM chats c
LEFT JOIN messages m ON c.id = m.chat_id AND m.is_deleted = 0
WHERE c.is_active = 1
GROUP BY c.id, c.name, c.type, c.members_count;

-- 用户活跃度视图
CREATE OR REPLACE VIEW v_user_activity AS
SELECT 
    u.id,
    u.username,
    u.nickname,
    COUNT(DISTINCT m.chat_id) as active_chats,
    COUNT(m.id) as message_count,
    MAX(m.created_at) as last_message_time,
    u.last_seen
FROM users u
LEFT JOIN messages m ON u.id = m.sender_id AND m.is_deleted = 0
WHERE u.is_active = 1
GROUP BY u.id, u.username, u.nickname, u.last_seen;

-- ==============================================
-- 10. 存储过程优化
-- ==============================================

-- 创建优化的存储过程
DELIMITER //

-- 获取用户聊天列表的存储过程
CREATE PROCEDURE GetUserChats(IN user_id INT)
BEGIN
    SELECT 
        c.id,
        c.name,
        c.avatar,
        c.type,
        c.members_count,
        c.updated_at,
        m.content as last_message,
        m.created_at as last_message_time,
        m.sender_id as last_message_sender,
        u.username as last_message_sender_name,
        COUNT(CASE WHEN mr.is_read = 0 THEN 1 END) as unread_count
    FROM chats c
    INNER JOIN chat_members cm ON c.id = cm.chat_id
    LEFT JOIN messages m ON c.id = m.chat_id AND m.is_deleted = 0
    LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = user_id
    LEFT JOIN users u ON m.sender_id = u.id
    WHERE cm.user_id = user_id AND c.is_active = 1
    GROUP BY c.id, c.name, c.avatar, c.type, c.members_count, c.updated_at, m.content, m.created_at, m.sender_id, u.username
    ORDER BY c.updated_at DESC;
END //

-- 获取聊天消息的存储过程
CREATE PROCEDURE GetChatMessages(
    IN chat_id INT,
    IN user_id INT,
    IN limit_count INT,
    IN offset_count INT
)
BEGIN
    SELECT 
        m.id,
        m.sender_id,
        m.content,
        m.type,
        m.created_at,
        m.is_read,
        m.is_edited,
        m.is_pinned,
        u.username as sender_name,
        u.nickname as sender_nickname,
        u.avatar as sender_avatar
    FROM messages m
    INNER JOIN users u ON m.sender_id = u.id
    WHERE m.chat_id = chat_id 
        AND m.is_deleted = 0
        AND (m.sender_id = user_id OR EXISTS (
            SELECT 1 FROM chat_members cm 
            WHERE cm.chat_id = chat_id AND cm.user_id = user_id
        ))
    ORDER BY m.created_at DESC
    LIMIT limit_count OFFSET offset_count;
END //

DELIMITER ;

-- ==============================================
-- 11. 触发器优化
-- ==============================================

-- 创建触发器自动更新聊天更新时间
DELIMITER //

CREATE TRIGGER tr_messages_update_chat_updated_at
AFTER INSERT ON messages
FOR EACH ROW
BEGIN
    UPDATE chats 
    SET updated_at = NEW.created_at 
    WHERE id = NEW.chat_id;
END //

CREATE TRIGGER tr_chat_members_update_members_count
AFTER INSERT ON chat_members
FOR EACH ROW
BEGIN
    UPDATE chats 
    SET members_count = (
        SELECT COUNT(*) FROM chat_members 
        WHERE chat_id = NEW.chat_id
    )
    WHERE id = NEW.chat_id;
END //

CREATE TRIGGER tr_chat_members_delete_update_members_count
AFTER DELETE ON chat_members
FOR EACH ROW
BEGIN
    UPDATE chats 
    SET members_count = (
        SELECT COUNT(*) FROM chat_members 
        WHERE chat_id = OLD.chat_id
    )
    WHERE id = OLD.chat_id;
END //

DELIMITER ;

-- ==============================================
-- 12. 清理和维护
-- ==============================================

-- 创建清理过期数据的存储过程
DELIMITER //

CREATE PROCEDURE CleanupExpiredData()
BEGIN
    -- 清理过期的已读记录（保留30天）
    DELETE FROM message_reads 
    WHERE read_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    -- 清理软删除的消息（保留90天）
    DELETE FROM messages 
    WHERE is_deleted = 1 AND updated_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
    
    -- 清理不活跃的用户（超过1年未登录）
    UPDATE users 
    SET is_active = 0 
    WHERE last_seen < DATE_SUB(NOW(), INTERVAL 1 YEAR) AND is_active = 1;
    
    -- 优化表
    OPTIMIZE TABLE users, chats, chat_members, messages, message_reads, contacts;
END //

DELIMITER ;

-- ==============================================
-- 13. 监控查询
-- ==============================================

-- 创建性能监控查询
-- 查看慢查询
-- SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10;

-- 查看表大小
SELECT 
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)',
    table_rows
FROM information_schema.tables 
WHERE table_schema = 'zhihang_messenger'
ORDER BY (data_length + index_length) DESC;

-- 查看索引使用情况
SELECT 
    table_name,
    index_name,
    cardinality,
    ROUND(((stat_value * @@innodb_page_size) / 1024 / 1024), 2) AS 'Index Size (MB)'
FROM information_schema.statistics 
WHERE table_schema = 'zhihang_messenger'
ORDER BY table_name, index_name;

-- 查看表锁等待情况
-- SELECT * FROM information_schema.innodb_lock_waits;

-- 查看连接数
-- SELECT * FROM information_schema.processlist WHERE command != 'Sleep';

-- ==============================================
-- 14. 备份优化
-- ==============================================

-- 创建备份脚本（需要在系统级别执行）
-- mysqldump --single-transaction --routines --triggers zhihang_messenger > backup_$(date +%Y%m%d_%H%M%S).sql

-- ==============================================
-- 执行完成
-- ==============================================

-- 显示优化完成信息
SELECT 'Database optimization completed successfully!' AS status;
