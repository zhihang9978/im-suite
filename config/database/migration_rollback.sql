-- ========================================
-- 志航密信 - 数据库迁移回滚脚本
-- ========================================
-- 使用方法：
-- 1. 备份当前数据库
-- 2. 执行此脚本回滚到初始状态
-- 3. 验证数据完整性
-- ========================================

-- 按照依赖关系逆序删除表

-- 1. 删除依赖表（外键关联）
DROP TABLE IF EXISTS `message_replies`;
DROP TABLE IF EXISTS `message_reactions`;
DROP TABLE IF EXISTS `message_read_receipts`;
DROP TABLE IF EXISTS `chat_members`;
DROP TABLE IF EXISTS `chat_permissions`;
DROP TABLE IF EXISTS `chat_announcements`;
DROP TABLE IF EXISTS `chat_invitations`;
DROP TABLE IF EXISTS `group_invitations`;
DROP TABLE IF EXISTS `bot_users`;
DROP TABLE IF EXISTS `bot_activity_logs`;
DROP TABLE IF EXISTS `user_devices`;
DROP TABLE IF EXISTS `user_sessions`;
DROP TABLE IF EXISTS `two_factor_settings`;
DROP TABLE IF EXISTS `trusted_devices`;
DROP TABLE IF EXISTS `user_theme_settings`;
DROP TABLE IF EXISTS `file_access_logs`;
DROP TABLE IF EXISTS `shared_files`;
DROP TABLE IF EXISTS `message_keys`;
DROP TABLE IF EXISTS `screen_share_sessions`;
DROP TABLE IF EXISTS `screen_share_stats`;
DROP TABLE IF EXISTS `notification_settings`;
DROP TABLE IF EXISTS `user_notifications`;
DROP TABLE IF EXISTS `admin_audit_logs`;
DROP TABLE IF EXISTS `content_reports`;
DROP TABLE IF EXISTS `moderation_actions`;
DROP TABLE IF EXISTS `banned_users`;
DROP TABLE IF EXISTS `ip_blacklist`;
DROP TABLE IF EXISTS `chat_backup_configs`;
DROP TABLE IF EXISTS `chat_backup_files`;
DROP TABLE IF EXISTS `chat_statistics`;

-- 2. 删除主表
DROP TABLE IF EXISTS `messages`;
DROP TABLE IF EXISTS `files`;
DROP TABLE IF EXISTS `chats`;
DROP TABLE IF EXISTS `friendships`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `bots`;
DROP TABLE IF EXISTS `system_config`;

-- 3. 删除其他辅助表
DROP TABLE IF EXISTS `user_presence`;
DROP TABLE IF EXISTS `typing_indicators`;
DROP TABLE IF EXISTS `message_queue`;

-- ========================================
-- 验证回滚
-- ========================================

SELECT '✅ 所有表已删除' AS status;

SHOW TABLES;

-- 应该返回空结果或只有系统表

-- ========================================
-- 重新迁移（如果需要）
-- ========================================

-- 回滚后如需重新迁移，执行：
-- 1. 重启backend服务
-- 2. GORM AutoMigrate会自动创建所有表
-- 或者执行：
-- docker exec im-backend-prod go run main.go migrate

