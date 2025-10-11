# 创建超级管理员账号指南

## 📋 部署后首次创建超级管理员

### 方法1：使用注册API（推荐）

**步骤1：注册普通用户**
```bash
curl -X POST http://your-domain/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "您的手机号",
    "username": "youradmin",
    "password": "您的强密码",
    "nickname": "系统管理员"
  }'
```

**步骤2：在MySQL中提升为超级管理员**
```bash
# 进入MySQL容器
docker exec -it im-mysql-prod mysql -uroot -p

# 在MySQL中执行
USE zhihang_messenger;
UPDATE users SET role = 'super_admin' WHERE username = 'youradmin';
SELECT id, username, role FROM users WHERE username = 'youradmin';
exit;
```

---

### 方法2：直接在数据库中创建

**注意**：需要先使用bcrypt生成密码哈希

#### 步骤1：生成bcrypt密码哈希

**使用在线工具**:
- https://bcrypt-generator.com/
- 输入您的密码
- 选择 rounds: 10
- 复制生成的哈希值

**或使用命令行**（需要安装bcrypt工具）:
```bash
# Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'你的密码', bcrypt.gensalt()).decode())"

# Node.js
node -e "const bcrypt = require('bcrypt'); console.log(bcrypt.hashSync('你的密码', 10));"
```

#### 步骤2：在MySQL中创建用户

```bash
# 进入MySQL容器
docker exec -it im-mysql-prod mysql -uroot -p

# 在MySQL中执行
USE zhihang_messenger;

INSERT INTO users (
    created_at, updated_at,
    phone, username, nickname,
    password, salt,
    role, is_active, online,
    language, theme
) VALUES (
    NOW(), NOW(),
    '您的手机号',                              -- 修改：例如 13800138000
    '您的管理员用户名',                         -- 修改：例如 admin
    '您的昵称',                                -- 修改：例如 系统管理员
    '步骤1生成的bcrypt哈希',                   -- 修改：粘贴bcrypt哈希
    '',                                       -- 可以留空，bcrypt已包含salt
    'super_admin',                            -- 角色：超级管理员
    TRUE,                                     -- 激活状态
    FALSE,                                    -- 在线状态
    'zh-CN',                                  -- 语言
    'auto'                                    -- 主题
);

-- 验证创建成功
SELECT id, phone, username, nickname, role, is_active, created_at 
FROM users 
WHERE role = 'super_admin';

exit;
```

---

### 方法3：使用准备好的SQL脚本

**创建文件**: `scripts/create-admin.sql`

```sql
-- 创建超级管理员
-- ⚠️ 请先修改以下信息

SET @admin_phone = '13800138000';              -- 修改：您的手机号
SET @admin_username = 'admin';                 -- 修改：您的用户名
SET @admin_nickname = '系统管理员';             -- 修改：您的昵称
SET @admin_password = '$2a$10$...';            -- 修改：bcrypt密码哈希

INSERT INTO users (
    created_at, updated_at,
    phone, username, nickname,
    password, salt,
    role, is_active, online,
    language, theme
) VALUES (
    NOW(), NOW(),
    @admin_phone,
    @admin_username,
    @admin_nickname,
    @admin_password,
    '',
    'super_admin',
    TRUE,
    FALSE,
    'zh-CN',
    'auto'
)
ON DUPLICATE KEY UPDATE 
    updated_at = NOW(),
    role = 'super_admin';

-- 显示结果
SELECT 
    id, 
    phone, 
    username, 
    nickname, 
    role, 
    is_active, 
    created_at 
FROM users 
WHERE username = @admin_username;
```

**执行脚本**:
```bash
# 修改scripts/create-admin.sql后执行
docker exec -i im-mysql-prod mysql -uroot -p${MYSQL_ROOT_PASSWORD} zhihang_messenger < scripts/create-admin.sql
```

---

## 🔐 密码安全建议

### 强密码要求
- ✅ 至少12个字符
- ✅ 包含大小写字母
- ✅ 包含数字
- ✅ 包含特殊字符
- ✅ 不使用常见密码

### 示例强密码
```
Admin@Zh2024!Secure
SuperAdmin#2024$Strong
ZhMessenger@2024!Admin
```

---

## 📊 验证超级管理员

### 检查用户列表
```sql
USE zhihang_messenger;

SELECT 
    id,
    phone,
    username,
    nickname,
    role,
    is_active,
    created_at
FROM users
ORDER BY id;
```

### 检查超级管理员权限
```sql
SELECT COUNT(*) as super_admin_count 
FROM users 
WHERE role = 'super_admin' AND is_active = TRUE;
```

---

## 🚀 登录超级管理后台

### 1. 访问管理后台
```
http://your-domain/
或
http://your-ip:8080/
```

### 2. 使用创建的账号登录
- 手机号：您设置的手机号
- 密码：您设置的密码

### 3. 访问超级管理功能
- 登录后会自动跳转到超级管理后台
- 路径：`/super-admin`

---

## ⚠️ 安全注意事项

### 1. 首次登录后立即操作
- [ ] 修改默认密码（如果使用了示例密码）
- [ ] 启用两步验证
- [ ] 检查登录日志

### 2. 账号管理
- ✅ 定期更换密码
- ✅ 不共享超级管理员账号
- ✅ 为不同管理员创建独立账号
- ✅ 定期审查管理员权限

### 3. 备份账号
- 建议创建2-3个超级管理员账号
- 分别保管在不同位置
- 避免单点故障

---

## 🔄 重置超级管理员密码

### 如果忘记密码

**步骤1：生成新的bcrypt哈希**
```bash
python3 -c "import bcrypt; print(bcrypt.hashpw(b'新密码', bcrypt.gensalt()).decode())"
```

**步骤2：在MySQL中更新**
```sql
USE zhihang_messenger;

UPDATE users 
SET password = '新的bcrypt哈希',
    updated_at = NOW()
WHERE username = 'youradmin';

-- 验证更新
SELECT username, updated_at FROM users WHERE username = 'youradmin';
```

---

## 📝 常见问题

### Q: 为什么不在初始化脚本中创建管理员？
**A**: 为了安全性和灵活性
- 避免默认密码泄露
- 每个部署环境使用不同的管理员信息
- 符合生产环境最佳实践

### Q: 可以创建多个超级管理员吗？
**A**: 可以
- 重复上述步骤，使用不同的用户名和手机号
- 每个超级管理员拥有相同的权限

### Q: 普通用户可以升级为超级管理员吗？
**A**: 可以
```sql
UPDATE users SET role = 'super_admin' WHERE username = '用户名';
```

---

## ✅ 检查清单

部署后首次配置：
- [ ] 创建超级管理员账号
- [ ] 使用强密码
- [ ] 测试登录
- [ ] 访问超级管理后台（/super-admin）
- [ ] 修改初始密码（如果使用了示例）
- [ ] 启用两步验证
- [ ] 创建备份管理员账号
- [ ] 记录账号信息到安全位置

---

**重要提醒**：
🔐 超级管理员拥有最高权限，务必保管好账号信息！
📝 建议将账号信息记录在安全的密码管理器中。

