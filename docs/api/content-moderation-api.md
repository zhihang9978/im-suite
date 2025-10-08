# 内容审核 API

## 概述

内容审核 API 提供了违规内容检测、举报、处理和统计功能。系统采用**检测上报机制**：自动检测到可疑内容后仅上报给管理员，**不会自动拦截或删除**，所有处理决定由管理员人工审核后执行。

## 核心特性

1. **自动检测但不拦截**: 系统检测到违规内容后仅上报，不影响用户正常使用
2. **人工审核决策**: 所有处理动作（警告、删除、封禁）由管理员审核后执行
3. **多维度检测**: 支持关键词、正则表达式、URL检测
4. **优先级分级**: 根据违规严重程度自动分配优先级
5. **完整审计日志**: 记录所有审核操作，可追溯

## 认证

所有内容审核 API 都需要通过 JWT Bearer Token 进行认证。在请求头中添加 `Authorization: Bearer <token>`。

## API 列表

### 1. 举报内容

- **URL**: `/api/moderation/report`
- **方法**: `POST`
- **描述**: 用户举报违规内容
- **请求体**:
  ```json
  {
    "content_type": "message",
    "content_id": 123,
    "content_user_id": 456,
    "reporter_id": 789,
    "report_reason": "spam",
    "report_detail": "这是一条垃圾广告信息",
    "report_evidence": "http://example.com/screenshot.jpg"
  }
  ```
- **支持的举报原因**:
  - `spam`: 垃圾信息
  - `porn`: 色情内容
  - `violence`: 暴力内容
  - `politics`: 政治敏感
  - `harassment`: 骚扰辱骂
  - `fraud`: 诈骗
  - `other`: 其他
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "举报已提交，我们会尽快处理",
    "report": {
      "id": 1,
      "content_type": "message",
      "content_id": 123,
      "status": "pending",
      "priority": "normal",
      "created_at": "2025-01-27T10:00:00Z"
    }
  }
  ```

### 2. 自动检测内容

- **URL**: `/api/moderation/check`
- **方法**: `POST`
- **描述**: 自动检测内容是否违规（仅上报，不拦截）
- **请求体**:
  ```json
  {
    "content_type": "message",
    "content_id": 123,
    "content_text": "要检测的内容文本",
    "user_id": 456
  }
  ```
- **成功响应 (200 OK)**:
  - 检测到违规:
    ```json
    {
      "detected": true,
      "message": "检测到可疑内容，已自动上报至管理员",
      "report": {
        "id": 2,
        "auto_detected": true,
        "detection_score": 0.85,
        "detection_keywords": "关键词1, 关键词2",
        "status": "pending"
      }
    }
    ```
  - 未检测到违规:
    ```json
    {
      "detected": false,
      "message": "内容检测通过"
    }
    ```

### 3. 获取待处理举报（管理员）

- **URL**: `/api/moderation/reports/pending`
- **方法**: `GET`
- **描述**: 获取待处理的举报列表
- **查询参数**:
  - `limit`: 每页数量 (default: 20)
  - `offset`: 偏移量 (default: 0)
  - `priority`: 优先级筛选 (urgent, high, normal, low)
- **成功响应 (200 OK)**:
  ```json
  {
    "reports": [
      {
        "id": 1,
        "content_type": "message",
        "content_id": 123,
        "content_text": "违规内容文本",
        "content_user": {
          "id": 456,
          "nickname": "用户昵称"
        },
        "reporter": {
          "id": 789,
          "nickname": "举报人昵称"
        },
        "report_reason": "spam",
        "auto_detected": false,
        "status": "pending",
        "priority": "normal",
        "created_at": "2025-01-27T10:00:00Z"
      }
    ],
    "total": 10,
    "limit": 20,
    "offset": 0
  }
  ```

### 4. 获取举报详情（管理员）

- **URL**: `/api/moderation/reports/{report_id}`
- **方法**: `GET`
- **描述**: 查看举报的详细信息
- **URL 参数**:
  - `report_id`: 举报ID
- **成功响应 (200 OK)**:
  ```json
  {
    "id": 1,
    "content_type": "message",
    "content_id": 123,
    "content_text": "完整的违规内容文本",
    "content_user": {
      "id": 456,
      "nickname": "涉事用户",
      "email": "user@example.com"
    },
    "reporter": {
      "id": 789,
      "nickname": "举报人"
    },
    "report_reason": "spam",
    "report_detail": "详细举报说明",
    "report_evidence": "证据链接",
    "auto_detected": true,
    "detection_type": "keyword",
    "detection_score": 0.85,
    "detection_keywords": "广告, 推广",
    "status": "pending",
    "priority": "high",
    "created_at": "2025-01-27T10:00:00Z"
  }
  ```

### 5. 处理举报（管理员）

- **URL**: `/api/moderation/reports/handle`
- **方法**: `POST`
- **描述**: 管理员处理举报内容
- **请求体**:
  ```json
  {
    "report_id": 1,
    "handler_id": 999,
    "handle_action": "warn",
    "handle_comment": "已向用户发出警告"
  }
  ```
- **支持的处理动作**:
  - `warn`: 发出警告（仅记录，不拦截）
  - `delete`: 标记删除（需管理员手动执行实际删除）
  - `ban`: 标记封禁（需管理员手动执行实际封禁）
  - `ignore`: 忽略举报
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "举报处理成功"
  }
  ```

### 6. 创建过滤规则（管理员）

- **URL**: `/api/moderation/filters`
- **方法**: `POST`
- **描述**: 创建内容自动过滤规则
- **请求体**:
  ```json
  {
    "name": "垃圾广告检测",
    "description": "检测常见垃圾广告关键词",
    "rule_type": "keyword",
    "rule_content": "广告,推广,加微信,免费领取",
    "category": "spam",
    "severity": "normal",
    "action": "report",
    "threshold": 0.8,
    "auto_report": true,
    "creator_id": 999
  }
  ```
- **规则类型**:
  - `keyword`: 关键词匹配（多个关键词用逗号分隔）
  - `regex`: 正则表达式匹配
  - `url`: URL检测
- **严重程度**:
  - `critical`: 严重
  - `high`: 高
  - `normal`: 普通
  - `low`: 低
- **成功响应 (200 OK)**:
  ```json
  {
    "message": "过滤规则创建成功",
    "filter": {
      "id": 1,
      "name": "垃圾广告检测",
      "rule_type": "keyword",
      "category": "spam",
      "is_enabled": true,
      "created_at": "2025-01-27T10:00:00Z"
    }
  }
  ```

### 7. 获取用户警告记录

- **URL**: `/api/moderation/users/{user_id}/warnings`
- **方法**: `GET`
- **描述**: 查看用户的警告记录
- **URL 参数**:
  - `user_id`: 用户ID
- **查询参数**:
  - `limit`: 每页数量 (default: 20)
  - `offset`: 偏移量 (default: 0)
- **成功响应 (200 OK)**:
  ```json
  {
    "warnings": [
      {
        "id": 1,
        "user_id": 456,
        "warning_type": "spam",
        "warning_level": "medium",
        "reason": "发送垃圾广告信息",
        "issued_by_user": {
          "id": 999,
          "nickname": "管理员"
        },
        "created_at": "2025-01-27T10:00:00Z",
        "is_acknowledged": false
      }
    ],
    "total": 1,
    "limit": 20,
    "offset": 0
  }
  ```

### 8. 获取审核统计

- **URL**: `/api/moderation/statistics`
- **方法**: `GET`
- **描述**: 获取内容审核的统计数据
- **查询参数**:
  - `start_date`: 开始日期 (format: 2006-01-02)
  - `end_date`: 结束日期 (format: 2006-01-02)
- **成功响应 (200 OK)**:
  ```json
  {
    "statistics": [
      {
        "date": "2025-01-27",
        "total_reports": 50,
        "pending_reports": 10,
        "resolved_reports": 35,
        "rejected_reports": 5,
        "auto_detected": 30,
        "manual_reported": 20,
        "warnings_issued": 15,
        "content_deleted": 5,
        "users_banned": 2,
        "spam_reports": 25,
        "porn_reports": 5,
        "violence_reports": 3,
        "politics_reports": 2,
        "other_reports": 15
      }
    ],
    "start_date": "2025-01-20",
    "end_date": "2025-01-27"
  }
  ```

## 工作流程

### 自动检测流程

1. 用户发送消息
2. 系统自动调用 `/api/moderation/check` 检测内容
3. 如果检测到违规：
   - 创建举报记录，标记为 `auto_detected`
   - 消息**仍然正常发送**，不被拦截
   - 管理员收到通知
4. 如果未检测到违规：
   - 消息正常发送
   - 不创建任何记录

### 人工举报流程

1. 用户看到违规内容，点击"举报"
2. 调用 `/api/moderation/report` 提交举报
3. 系统创建举报记录，状态为 `pending`
4. 管理员在审核面板查看举报
5. 管理员调用 `/api/moderation/reports/handle` 处理

### 管理员处理流程

1. 管理员查看待处理举报列表
2. 点击举报查看详情
3. 根据内容严重程度选择处理动作：
   - **发出警告**: 仅记录警告，不影响用户使用
   - **标记删除**: 标记需要删除的内容，管理员手动执行实际删除
   - **标记封禁**: 标记需要封禁的用户，管理员手动执行实际封禁
   - **忽略**: 误报或不构成违规
4. 填写处理备注，提交处理
5. 系统记录审核日志，更新统计数据

## 使用示例

### JavaScript/TypeScript 示例

```typescript
// 自动检测消息内容
const checkMessageContent = async (messageId: number, content: string, userId: number) => {
  const response = await fetch('/api/moderation/check', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      content_type: 'message',
      content_id: messageId,
      content_text: content,
      user_id: userId
    })
  });
  
  const result = await response.json();
  
  if (result.detected) {
    console.log('检测到可疑内容，已上报:', result.report);
    // 注意：消息仍然会正常发送，不会被拦截
  }
  
  return result;
};

// 用户举报内容
const reportContent = async (contentId: number, reason: string, detail: string) => {
  const response = await fetch('/api/moderation/report', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      content_type: 'message',
      content_id: contentId,
      content_user_id: targetUserId,
      report_reason: reason,
      report_detail: detail
    })
  });
  
  return response.json();
};

// 管理员处理举报
const handleReport = async (reportId: number, action: string, comment: string) => {
  const response = await fetch('/api/moderation/reports/handle', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${adminToken}`
    },
    body: JSON.stringify({
      report_id: reportId,
      handle_action: action,
      handle_comment: comment
    })
  });
  
  return response.json();
};
```

## 注意事项

1. **不自动拦截**: 系统检测到违规内容后仅上报，不会自动拦截或删除，确保不会误伤正常内容
2. **人工审核**: 所有处理决定由管理员人工审核后执行，避免自动化误判
3. **权限控制**: 只有管理员可以查看举报列表和处理举报
4. **隐私保护**: 举报人信息对被举报用户不可见
5. **审计日志**: 所有审核操作都会记录日志，可追溯查询
6. **统计分析**: 提供完整的统计数据，帮助了解系统使用情况

## 最佳实践

1. **及时处理**: 优先处理 `urgent` 和 `high` 优先级的举报
2. **详细记录**: 处理举报时填写详细的处理备注，便于后续追溯
3. **规则优化**: 根据统计数据定期优化过滤规则，提高检测准确性
4. **用户沟通**: 对警告用户及时沟通，说明违规原因
5. **误报处理**: 及时处理误报举报，避免影响用户体验
