# OpenWan API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **API Version**: v1
- **认证方式**: Session Cookie / JWT Token
- **Content-Type**: `application/json`

## 认证接口

### 登录
```http
POST /v1/auth/login
```

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "group_id": 1,
      "level_id": 1
    },
    "permissions": [
      {
        "namespace": "admin",
        "controller": "users",
        "action": "index"
      }
    ],
    "roles": ["admin"]
  }
}
```

### 登出
```http
POST /v1/auth/logout
```

### 获取当前用户信息
```http
GET /v1/auth/me
```

## 文件管理

### 获取文件列表
```http
GET /v1/files?page=1&page_size=20&type=1&status=2&category_id=1
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `type`: 文件类型 (1:视频 2:音频 3:图片 4:富媒体)
- `status`: 文件状态 (0:新上传 1:待审核 2:已发布 3:已拒绝 4:已删除)
- `category_id`: 分类ID
- `keyword`: 搜索关键词

**响应**:
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "id": 1,
        "title": "新闻联播",
        "type": 1,
        "status": 2,
        "size": 524288000,
        "ext": "mp4",
        "upload_username": "admin",
        "upload_at": 1706745600,
        "category_id": 1,
        "category_name": "新闻"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### 获取文件详情
```http
GET /v1/files/{id}
```

### 上传文件
```http
POST /v1/files/upload
Content-Type: multipart/form-data
```

**表单数据**:
- `file`: 文件对象 (required)
- `title`: 文件标题 (required)
- `type`: 文件类型 (required)
- `category_id`: 分类ID (required)
- `description`: 文件描述

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 123,
    "title": "新上传的文件",
    "path": "/data1/abc123/def456.mp4",
    "size": 1048576,
    "md5": "abc123def456..."
  }
}
```

### 更新文件
```http
PUT /v1/files/{id}
```

### 删除文件
```http
DELETE /v1/files/{id}
```

### 提交审核
```http
POST /v1/files/{id}/submit
```

### 发布文件
```http
POST /v1/files/{id}/publish
```

### 拒绝文件
```http
POST /v1/files/{id}/reject
```

**请求体**:
```json
{
  "reason": "不符合发布要求"
}
```

### 下载文件
```http
GET /v1/files/{id}/download
```

### 获取预览文件
```http
GET /v1/files/{id}/preview
```

## 搜索接口

### 全文搜索
```http
POST /v1/search
```

**请求体**:
```json
{
  "q": "新闻",
  "types": [1, 2],
  "statuses": [2],
  "category_id": 1,
  "date_from": "2024-01-01",
  "date_to": "2024-12-31",
  "uploader": "admin",
  "sort_by": "relevance",
  "page": 1,
  "page_size": 20
}
```

## 分类管理

### 获取分类树
```http
GET /v1/categories/tree
```

### 获取分类列表
```http
GET /v1/categories?page=1&page_size=20
```

### 创建分类
```http
POST /v1/categories
```

**请求体**:
```json
{
  "parent_id": 1,
  "name": "国际新闻",
  "description": "国际新闻分类",
  "weight": 10,
  "level": 1,
  "group_ids": [1, 2],
  "status": 1
}
```

### 更新分类
```http
PUT /v1/categories/{id}
```

### 删除分类
```http
DELETE /v1/categories/{id}
```

## 目录配置

### 获取目录配置
```http
GET /v1/catalog/tree?type=1
```

**查询参数**:
- `type`: 文件类型 (1:视频 2:音频 3:图片 4:富媒体)

### 创建目录字段
```http
POST /v1/catalog
```

**请求体**:
```json
{
  "parent_id": null,
  "file_type": 1,
  "name": "director",
  "label": "导演",
  "type": "text",
  "options": "",
  "default_value": "",
  "placeholder": "请输入导演姓名",
  "rules": ["required"],
  "weight": 10,
  "enabled": 1
}
```

**字段类型**:
- `text`: 文本输入
- `textarea`: 多行文本
- `number`: 数字输入
- `date`: 日期选择
- `select`: 下拉选择
- `radio`: 单选框
- `checkbox`: 复选框
- `group`: 分组

## 用户管理

### 获取用户列表
```http
GET /v1/admin/users?page=1&page_size=20
```

### 创建用户
```http
POST /v1/admin/users
```

**请求体**:
```json
{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com",
  "group_id": 2,
  "level_id": 2,
  "status": 1
}
```

### 更新用户
```http
PUT /v1/admin/users/{id}
```

### 删除用户
```http
DELETE /v1/admin/users/{id}
```

### 重置密码
```http
POST /v1/admin/users/{id}/reset-password
```

## 组管理

### 获取组列表
```http
GET /v1/admin/groups
```

### 创建组
```http
POST /v1/admin/groups
```

### 分配分类
```http
POST /v1/admin/groups/{id}/categories
```

**请求体**:
```json
{
  "category_ids": [1, 2, 3]
}
```

### 分配角色
```http
POST /v1/admin/groups/{id}/roles
```

**请求体**:
```json
{
  "role_ids": [1, 2]
}
```

## 角色管理

### 获取角色列表
```http
GET /v1/admin/roles
```

### 创建角色
```http
POST /v1/admin/roles
```

### 分配权限
```http
POST /v1/admin/roles/{id}/permissions
```

**请求体**:
```json
{
  "permission_ids": [1, 2, 3, 4, 5]
}
```

### 获取所有权限
```http
GET /v1/admin/permissions
```

## 健康检查

### 健康检查
```http
GET /health
```

**响应**:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": 1706745600,
  "checks": {
    "database": "ok",
    "redis": "ok",
    "rabbitmq": "ok",
    "storage": "ok"
  }
}
```

### 就绪检查
```http
GET /ready
```

### 存活检查
```http
GET /alive
```

## 错误响应

所有错误响应遵循统一格式：

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "details": {
      "field": "username",
      "issue": "用户名已存在"
    }
  }
}
```

**错误码**:
- `VALIDATION_ERROR`: 参数验证错误
- `AUTHENTICATION_ERROR`: 认证失败
- `AUTHORIZATION_ERROR`: 权限不足
- `NOT_FOUND`: 资源不存在
- `CONFLICT`: 资源冲突
- `INTERNAL_ERROR`: 服务器内部错误

## 状态码

- `200`: 成功
- `201`: 创建成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 资源冲突
- `500`: 服务器错误
- `503`: 服务不可用

## 速率限制

- **匿名用户**: 10 请求/分钟
- **普通用户**: 100 请求/分钟
- **管理员**: 1000 请求/分钟

超出限制返回 `429 Too Many Requests`

## 分页

所有列表接口支持分页，使用以下参数：

- `page`: 页码 (从1开始)
- `page_size`: 每页数量 (最大100)

响应头包含：
- `X-Total-Count`: 总记录数
- `X-Page`: 当前页
- `X-Page-Size`: 每页数量
- `X-Total-Pages`: 总页数
