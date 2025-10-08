# 文件管理 API 文档

## 概述

文件管理 API 提供了完整的文件上传、下载、预览、版本控制等功能。支持单文件上传、分片上传、文件加密、文件预览等功能。

## 认证

所有 API 请求都需要在请求头中包含有效的访问令牌：

```
Authorization: Bearer <access_token>
```

## 文件上传

### 单文件上传

**端点**: `POST /api/files/upload`

**描述**: 上传单个文件，支持小文件直接上传

**请求类型**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file | File | 是 | 要上传的文件 |
| is_encrypted | boolean | 否 | 是否加密文件，默认 false |

**响应示例**:

```json
{
  "file_id": 123,
  "file_name": "example.jpg",
  "file_url": "/api/files/123/download",
  "thumbnail_url": "/api/files/123/thumbnail",
  "preview_url": "/api/files/123/preview",
  "upload_id": "abc123",
  "chunk_index": 0,
  "is_complete": true
}
```

### 分片上传

**端点**: `POST /api/files/upload-chunk`

**描述**: 上传文件分片，用于大文件上传

**请求类型**: `multipart/form-data`

**请求参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| chunk | File | 是 | 文件分片 |
| upload_id | string | 是 | 上传ID，用于标识同一文件的不同分片 |
| chunk_index | integer | 是 | 分片索引，从0开始 |
| total_chunks | integer | 是 | 总分片数 |
| file_name | string | 是 | 原始文件名 |
| file_size | integer | 是 | 文件总大小（字节） |
| is_encrypted | boolean | 否 | 是否加密文件 |

**响应示例**:

```json
{
  "upload_id": "abc123",
  "chunk_index": 0,
  "is_complete": false
}
```

**完成响应**:

```json
{
  "file_id": 123,
  "file_name": "large_file.zip",
  "file_url": "/api/files/123/download",
  "thumbnail_url": "/api/files/123/thumbnail",
  "preview_url": "/api/files/123/preview",
  "upload_id": "abc123",
  "chunk_index": 5,
  "is_complete": true
}
```

## 文件操作

### 获取文件信息

**端点**: `GET /api/files/{file_id}`

**描述**: 获取指定文件的详细信息

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**响应示例**:

```json
{
  "id": 123,
  "file_name": "example.jpg",
  "file_size": 1048576,
  "file_type": "image",
  "mime_type": "image/jpeg",
  "file_hash": "abc123...",
  "storage_path": "uploads/2024/01/15/abc123_example.jpg",
  "storage_url": "/api/files/123/download",
  "thumbnail": "/api/files/123/thumbnail",
  "preview": "/api/files/123/preview",
  "is_encrypted": false,
  "version": 1,
  "is_latest": true,
  "owner_id": 456,
  "is_public": false,
  "download_count": 5,
  "view_count": 12,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 下载文件

**端点**: `GET /api/files/{file_id}/download`

**描述**: 下载指定文件

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**响应**: 返回文件二进制内容，Content-Type 为文件的 MIME 类型

**响应头**:

```
Content-Disposition: attachment; filename="example.jpg"
Content-Type: image/jpeg
```

### 删除文件

**端点**: `DELETE /api/files/{file_id}`

**描述**: 删除指定文件（软删除）

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**响应示例**:

```json
{
  "message": "文件删除成功"
}
```

## 文件预览

### 获取文件预览

**端点**: `GET /api/files/{file_id}/preview`

**描述**: 获取文件的预览信息

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**查询参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| preview_type | string | 否 | 预览类型：image/video/audio/document |
| width | integer | 否 | 预览宽度，默认 800 |
| height | integer | 否 | 预览高度，默认 600 |
| quality | string | 否 | 预览质量：low/medium/high，默认 medium |

**响应示例**:

```json
{
  "id": 456,
  "file_id": 123,
  "preview_type": "image",
  "preview_url": "/api/files/123/preview?type=image&w=800&h=600&q=high",
  "thumbnail_url": "/api/files/123/thumbnail",
  "width": 800,
  "height": 600,
  "quality": "high"
}
```

## 文件版本管理

### 获取文件版本列表

**端点**: `GET /api/files/{file_id}/versions`

**描述**: 获取指定文件的所有版本

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**响应示例**:

```json
[
  {
    "id": 123,
    "file_name": "document.pdf",
    "version": 2,
    "is_latest": true,
    "created_at": "2024-01-15T15:30:00Z",
    "updated_at": "2024-01-15T15:30:00Z"
  },
  {
    "id": 122,
    "file_name": "document.pdf",
    "version": 1,
    "is_latest": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

### 创建文件版本

**端点**: `POST /api/files/{file_id}/versions`

**描述**: 为指定文件创建新版本

**路径参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| file_id | integer | 是 | 文件ID |

**请求体**:

```json
{
  "new_file_name": "document_v2.pdf"
}
```

**请求参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| new_file_name | string | 否 | 新版本的文件名 |

**响应示例**:

```json
{
  "id": 124,
  "file_name": "document_v2.pdf",
  "file_size": 2097152,
  "file_type": "document",
  "mime_type": "application/pdf",
  "file_hash": "def456...",
  "storage_path": "uploads/2024/01/15/def456_document_v2.pdf",
  "storage_url": "/api/files/124/download",
  "is_encrypted": false,
  "version": 2,
  "is_latest": true,
  "parent_id": 123,
  "owner_id": 456,
  "is_public": false,
  "download_count": 0,
  "view_count": 0,
  "created_at": "2024-01-15T16:00:00Z",
  "updated_at": "2024-01-15T16:00:00Z"
}
```

## 错误响应

所有 API 在出错时都会返回以下格式的错误响应：

```json
{
  "error": "错误描述"
}
```

### 常见错误码

| HTTP状态码 | 错误描述 |
|------------|----------|
| 400 | 请求参数错误 |
| 401 | 未授权访问 |
| 403 | 权限不足 |
| 404 | 文件不存在 |
| 413 | 文件过大 |
| 415 | 不支持的文件类型 |
| 500 | 服务器内部错误 |

## 使用示例

### JavaScript 上传文件

```javascript
// 单文件上传
async function uploadFile(file, isEncrypted = false) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('is_encrypted', isEncrypted.toString());

  const response = await fetch('/api/files/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    },
    body: formData
  });

  if (!response.ok) {
    throw new Error('上传失败');
  }

  return await response.json();
}

// 分片上传
async function uploadFileInChunks(file, chunkSize = 5 * 1024 * 1024) {
  const totalChunks = Math.ceil(file.size / chunkSize);
  const uploadId = generateUploadId();

  for (let i = 0; i < totalChunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('upload_id', uploadId);
    formData.append('chunk_index', i.toString());
    formData.append('total_chunks', totalChunks.toString());
    formData.append('file_name', file.name);
    formData.append('file_size', file.size.toString());

    const response = await fetch('/api/files/upload-chunk', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      },
      body: formData
    });

    if (!response.ok) {
      throw new Error(`分片 ${i} 上传失败`);
    }

    const result = await response.json();
    if (result.is_complete) {
      return result;
    }
  }
}

function generateUploadId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}
```

### 下载文件

```javascript
async function downloadFile(fileId) {
  const response = await fetch(`/api/files/${fileId}/download`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });

  if (!response.ok) {
    throw new Error('下载失败');
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'download';
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}
```

## 配置说明

### 文件上传限制

- 默认最大文件大小：100MB
- 默认分片大小：5MB
- 支持的文件类型：图片、视频、音频、文档等

### 存储配置

- 文件存储路径：`uploads/{year}/{month}/{day}/`
- 分片临时路径：`chunks/{upload_id}/`
- 支持本地存储和对象存储（MinIO）

### 加密配置

- 支持 AES-256-GCM 和 AES-256-CBC 加密算法
- 自动生成加密密钥
- 支持密码派生密钥
