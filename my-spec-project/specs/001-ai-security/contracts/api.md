# API Contracts: ai-security

## 基础约定

- 基础路径：`/api/ai-security`
- 响应格式统一为：
  ```json
  {
    "success": true,
    "message": "",
    "data": {}
  }
  ```
- 认证：管理接口需要 Admin 权限；检测接口在 relay 链路中内部调用。

---

## 配置管理

### GET /api/ai-security/configs

获取模块全局配置。

**Response:**
```json
{
  "success": true,
  "data": {
    "ai_security_enabled": true,
    "ai_model_name": "gpt-4o-mini",
    "ai_timeout_seconds": 3,
    "log_retention_days": 30,
    "audit_log_retention_days": 90,
    "default_risk_score": 50,
    "max_group_depth": 5
  }
}
```

### PUT /api/ai-security/configs

更新模块全局配置。

**Request:**
```json
{
  "ai_security_enabled": true,
  "ai_model_name": "gpt-4o-mini",
  "ai_timeout_seconds": 3,
  "log_retention_days": 30
}
```

**Response:**
```json
{
  "success": true,
  "message": "配置更新成功"
}
```

---

## 分组管理

### GET /api/ai-security/groups

获取分组列表。

**Query Parameters:**
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20
- `status`: 状态筛选，-1 表示全部
- `parent_id`: 父分组 ID，-1 表示全部
- `name`: 名称模糊搜索

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "基础安全策略",
        "description": "",
        "parent_id": 0,
        "depth": 0,
        "path": "/1",
        "status": 1,
        "sort_order": 0,
        "created_at": 1719200000,
        "updated_at": 1719200000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### POST /api/ai-security/groups

创建分组。

**Request:**
```json
{
  "name": "个人隐私信息",
  "description": "身份证、手机号等",
  "parent_id": 0,
  "sort_order": 1,
  "status": 1
}
```

### PUT /api/ai-security/groups/:id

更新分组。

### PATCH /api/ai-security/groups/:id/status

更新分组状态。

**Request:**
```json
{
  "status": 0
}
```

### DELETE /api/ai-security/groups/:id

删除分组及其子分组、规则。

### POST /api/ai-security/groups/:id/copy

复制分组及其规则。

---

## 规则管理

### GET /api/ai-security/rules

获取规则列表。

**Query Parameters:**
- `page`: 页码
- `page_size`: 每页数量
- `group_id`: 分组筛选
- `type`: 类型筛选
- `status`: 状态筛选

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "code": "basic-political",
        "group_id": 1,
        "group_name": "基础安全策略",
        "name": "政治敏感",
        "type": 1,
        "content": "词1,词2,词3",
        "extra_config": "{\"mask_type\":\"full\"}",
        "action": 4,
        "priority": 10,
        "risk_score": 80,
        "status": 1,
        "is_seed": true,
        "created_at": 1719200000,
        "updated_at": 1719200000
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### POST /api/ai-security/rules

创建规则。

**Request:**
```json
{
  "group_id": 1,
  "name": "手机号检测",
  "type": 2,
  "content": "\\b1[3-9]\\d{9}\\b",
  "extra_config": "{\"mask_type\":\"preserve\",\"preserve_start\":3,\"preserve_end\":4}",
  "action": 3,
  "priority": 5,
  "risk_score": 60
}
```

### PUT /api/ai-security/rules/:id

更新规则。

### DELETE /api/ai-security/rules/:id

删除规则。

### POST /api/ai-security/rules/:id/test

测试单条规则。

**Request:**
```json
{
  "content": "测试文本包含敏感词"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "detected": true,
    "action": 4,
    "action_name": "block",
    "risk_score": 80,
    "risk_level": 4,
    "processed_content": "",
    "matches": [
      {
        "rule_id": 1,
        "group_id": 1,
        "type": 1,
        "matched_text": "敏感词",
        "position": [8, 11]
      }
    ]
  }
}
```

### POST /api/ai-security/rules/batch-delete

批量删除规则。

**Request:**
```json
{
  "ids": [1, 2, 3]
}
```

### POST /api/ai-security/rules/batch-status

批量更新规则状态。

**Request:**
```json
{
  "ids": [1, 2, 3],
  "status": 0
}
```

---

## 策略管理

### GET /api/ai-security/policies

获取策略列表。

**Query Parameters:**
- `page`: 页码
- `page_size`: 每页数量
- `user_id`: 用户筛选
- `status`: 状态筛选

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "user_name": "admin",
        "group_id": 1,
        "group_name": "基础安全策略",
        "scope": 3,
        "default_action": 4,
        "custom_response": "请求包含敏感内容，已被拦截。",
        "whitelist_ips": "[]",
        "priority": 0,
        "status": 1,
        "created_at": 1719200000,
        "updated_at": 1719200000
      }
    ],
    "total": 1
  }
}
```

### POST /api/ai-security/policies

创建策略。

**Request:**
```json
{
  "user_id": 1,
  "group_id": 1,
  "scope": 3,
  "default_action": 4,
  "custom_response": "",
  "whitelist_ips": "[]",
  "priority": 0
}
```

### PUT /api/ai-security/policies/:id

更新策略。

### DELETE /api/ai-security/policies/:id

删除策略。

---

## 命中日志

### GET /api/ai-security/logs

获取命中日志列表。

**Query Parameters:**
- `page`: 页码
- `page_size`: 每页数量
- `user_id`: 用户筛选
- `action`: 动作筛选
- `risk_level`: 风险等级筛选
- `direction`: 方向筛选
- `start_time`: 开始时间戳
- `end_time`: 结束时间戳
- `model_name`: 模型名称筛选
- `group_id`: 分组筛选
- `rule_id`: 规则筛选

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "request_id": "req-xxx",
        "user_id": 1,
        "user_name": "admin",
        "token_id": 1,
        "model_name": "gpt-4o",
        "channel_id": 1,
        "direction": 1,
        "rule_id": 1,
        "rule_name": "政治敏感",
        "group_id": 1,
        "group_name": "基础安全策略",
        "risk_level": 4,
        "action": 4,
        "matched_text": "敏感词",
        "hit_reason": "命中关键词规则",
        "ip": "127.0.0.1",
        "created_at": 1719200000
      }
    ],
    "total": 1
  }
}
```

### GET /api/ai-security/logs/export

导出命中日志（CSV/Excel）。

**Query Parameters:**
- `format`: `csv` 或 `excel`
- 其他同 `/logs`

---

## Dashboard

### GET /api/ai-security/dashboard

获取看板统计数据。

**Query Parameters:**
- `start_time`: 开始时间戳
- `end_time`: 结束时间戳
- `user_id`: 用户筛选（可选）
- `group_id`: 分组筛选（可选）
- `rule_id`: 规则筛选（可选）

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_detections": 1000,
      "total_interceptions": 50,
      "total_alerts": 200,
      "today_detections": 30,
      "today_interceptions": 2
    },
    "risk_distribution": {
      "low": 500,
      "medium": 300,
      "high": 150,
      "critical": 50
    },
    "top_categories": [
      {"category": "基础安全策略", "count": 400}
    ],
    "top_users": [
      {"user_id": 1, "user_name": "admin", "count": 100}
    ],
    "top_models": [
      {"model_name": "gpt-4o", "count": 600}
    ]
  }
}
```

---

## 检测接口

### POST /api/ai-security/check/request

请求内容检测（管理测试用）。

**Request:**
```json
{
  "user_id": 1,
  "content": "测试文本",
  "model_name": "gpt-4o"
}
```

### POST /api/ai-security/check/response

响应内容检测（管理测试用）。

---

## 同步官方敏感词

### POST /api/ai-security/sync/official-sensitive-words

将官方 sensitive-words 单向导入到 ai-security 规则库。

**Request:**
```json
{
  "target_group_id": 1,
  "action": 4,
  "risk_score": 50
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "sync_count": 100,
    "last_sync_at": 1719200000
  }
}
```

---

## 状态与安装

### GET /api/ai-security/status

获取模块状态。

**Response:**
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "rule_count": 50,
    "group_count": 10,
    "policy_count": 5,
    "cache_enabled": true,
    "version": "1.0.0"
  }
}
```

### POST /api/ai-security/install

执行安装/初始化（可选 API，install.sh 也会调用）。

**Response:**
```json
{
  "success": true,
  "message": "安装完成",
  "data": {
    "migrated": true,
    "seed_count": 20,
    "version": "1.0.0"
  }
}
```
