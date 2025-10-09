# 双因子认证 API 文档

## 概述

双因子认证（2FA）为用户账户提供额外的安全保护层。本文档描述了所有2FA相关的API接口。

## 基础URL

```
https://api.zhihang-messenger.com/api
```

## 认证

除了登录验证接口外，所有其他接口都需要在请求头中包含JWT令牌：

```
Authorization: Bearer {access_token}
```

---

## API接口

### 1. 启用双因子认证

**接口**: `POST /2fa/enable`

**描述**: 为用户账户启用双因子认证，生成TOTP密钥和备用码

**请求头**:
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**请求体**:
```json
{
  "password": "user_password"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "secret": "JBSWY3DPEHPK3PXP",
    "qr_code": "otpauth://totp/志航密信:username?secret=JBSWY3DPEHPK3PXP&issuer=志航密信",
    "backup_codes": [
      "ABCD1234",
      "EFGH5678",
      "IJKL9012",
      "MNOP3456",
      "QRST7890",
      "UVWX1234",
      "YZAB5678",
      "CDEF9012",
      "GHIJ3456",
      "KLMN7890"
    ]
  },
  "message": "请使用验证器APP扫描二维码，并输入验证码完成启用"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "密码错误"
}
```

---

### 2. 验证并启用2FA

**接口**: `POST /2fa/verify`

**描述**: 验证TOTP验证码并完成2FA启用

**请求头**:
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**请求体**:
```json
{
  "code": "123456"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "双因子认证已成功启用"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "验证码错误"
}
```

---

### 3. 禁用双因子认证

**接口**: `POST /2fa/disable`

**描述**: 禁用用户账户的双因子认证

**请求头**:
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**请求体**:
```json
{
  "password": "user_password",
  "code": "123456"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "双因子认证已禁用"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "密码错误"
}
```

---

### 4. 获取2FA状态

**接口**: `GET /2fa/status`

**描述**: 获取用户的双因子认证状态和统计信息

**请求头**:
```
Authorization: Bearer {access_token}
```

**响应** (200 OK):
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "trusted_devices_count": 3,
    "backup_codes_remaining": 8,
    "recent_attempts": [
      {
        "id": 1,
        "user_id": 123,
        "method": "totp",
        "status": "success",
        "ip": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "device": "iPhone",
        "created_at": "2024-12-19T10:30:00Z"
      }
    ]
  }
}
```

---

### 5. 重新生成备用码

**接口**: `POST /2fa/backup-codes/regenerate`

**描述**: 重新生成2FA备用码（旧备用码将失效）

**请求头**:
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**请求体**:
```json
{
  "password": "user_password",
  "code": "123456"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "backup_codes": [
    "ABCD1234",
    "EFGH5678",
    "IJKL9012",
    "MNOP3456",
    "QRST7890",
    "UVWX1234",
    "YZAB5678",
    "CDEF9012",
    "GHIJ3456",
    "KLMN7890"
  ],
  "message": "备用码已重新生成，请妥善保管"
}
```

---

### 6. 获取受信任设备列表

**接口**: `GET /2fa/trusted-devices`

**描述**: 获取用户的所有受信任设备

**请求头**:
```
Authorization: Bearer {access_token}
```

**响应** (200 OK):
```json
{
  "success": true,
  "devices": [
    {
      "id": 1,
      "user_id": 123,
      "device_id": "device-uuid-1234",
      "device_name": "iPhone 13 Pro",
      "device_type": "mobile",
      "ip": "192.168.1.1",
      "last_used_at": "2024-12-19T10:30:00Z",
      "trust_expires_at": "2025-01-18T10:30:00Z",
      "is_active": true,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ]
}
```

---

### 7. 移除受信任设备

**接口**: `DELETE /2fa/trusted-devices/:device_id`

**描述**: 移除指定的受信任设备

**请求头**:
```
Authorization: Bearer {access_token}
```

**URL参数**:
- `device_id`: 设备ID

**响应** (200 OK):
```json
{
  "success": true,
  "message": "设备已移除"
}
```

---

### 8. 验证2FA验证码（登录时使用）

**接口**: `POST /auth/2fa/validate`

**描述**: 在登录时验证用户的2FA验证码

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "user_id": 123,
  "code": "123456"
}
```

**响应** (200 OK):
```json
{
  "success": true,
  "message": "验证成功"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "验证码错误"
}
```

---

## 使用流程

### 启用2FA流程

1. 用户调用 `POST /2fa/enable` 接口，提供密码
2. 系统返回TOTP密钥和二维码URL
3. 用户使用验证器APP扫描二维码
4. 用户调用 `POST /2fa/verify` 接口，提供验证码
5. 系统验证成功后启用2FA

### 登录流程（已启用2FA）

1. 用户正常登录（用户名+密码）
2. 系统检测到用户已启用2FA
3. 检查设备是否受信任：
   - 如果设备受信任且未过期，直接登录成功
   - 如果设备不受信任，要求输入2FA验证码
4. 用户输入验证码（TOTP或备用码）
5. 调用 `POST /auth/2fa/validate` 验证
6. 验证成功后登录

### 设备信任机制

- 成功验证2FA后，可以将设备标记为受信任
- 受信任设备默认30天有效期
- 可以随时移除受信任设备

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（令牌无效或已过期） |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 安全建议

1. **妥善保管备用码**: 备用码应该打印或保存在安全的地方
2. **定期检查受信任设备**: 移除不再使用的设备
3. **使用官方验证器**: 推荐使用 Google Authenticator、Microsoft Authenticator 或 Authy
4. **不要共享密钥**: TOTP密钥应该保密，不要分享给任何人
5. **立即报告异常**: 如果发现未授权的验证尝试，请立即联系管理员

---

## 常见问题

### Q: 忘记验证器设备怎么办？
A: 可以使用备用码登录

### Q: 备用码用完了怎么办？
A: 登录后可以重新生成备用码

### Q: 可以在多个设备上使用同一个TOTP密钥吗？
A: 可以，但不推荐。建议每个设备使用不同的账户或密钥

### Q: 更换手机后如何迁移2FA？
A: 有两种方式：
1. 使用备用码登录，然后重新设置2FA
2. 在更换前，在新手机上也扫描相同的二维码

---

**版本**: v1.4.0  
**最后更新**: 2024-12-19  
**维护者**: 志航密信技术团队

