-- 志航密信 - MySQL初始化脚本
-- 创建数据库和基础配置

-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS zhihang_messenger 
  CHARACTER SET utf8mb4 
  COLLATE utf8mb4_unicode_ci;

USE zhihang_messenger;

-- 设置时区
SET time_zone = '+08:00';

-- 性能优化设置
SET GLOBAL max_connections = 1000;
SET GLOBAL connect_timeout = 60;
SET GLOBAL wait_timeout = 28800;
SET GLOBAL interactive_timeout = 28800;

-- 日志设置
SET GLOBAL slow_query_log = 1;
SET GLOBAL long_query_time = 2;
SET GLOBAL log_queries_not_using_indexes = 1;

-- 提示信息
SELECT 'MySQL初始化完成' AS status, 
       DATABASE() AS current_database,
       @@character_set_database AS charset,
       @@collation_database AS collation,
       @@version AS version;



