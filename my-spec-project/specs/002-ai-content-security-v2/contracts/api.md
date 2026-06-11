# API Contracts — AI Content Security Management v2

> **Scope**: New and modified endpoints for the v2 enhancement.  
> **Base Path**: `/api/security`  
> **Auth**: All admin endpoints require `middleware.AdminAuth()`.

---

## Modified Endpoints

### `GET /api/security/groups`

**Changes**: Adds `name` query parameter for fuzzy search.

| Query Param | Type | Required | Description |
|-------------|------|----------|-------------|
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20) |
| `status` | int | No | Filter by status: `-1`=all, `0`=disabled, `1`=enabled |
| `parent_id` | int64 | No | Filter by parent group ID; `-1`=all |
| `name` | string | No | **NEW** Fuzzy search by group name (`LIKE %name%`) |

**Response** (unchanged shape):
```json
{
  "success": true,
  "data": {
    "items": [ /* SecurityGroup[] */ ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

### `GET /api/security/logs`

**Changes**: Adds `start_time`, `end_time`, and `model_name` query parameters.

| Query Param | Type | Required | Description |
|-------------|------|----------|-------------|
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20) |
| `user_id` | int | No | Filter by user ID; `0`=all |
| `action` | int | No | Filter by action; `0`=all |
| `risk_level` | int | No | Filter by risk level; `0`=all |
| `content_type` | int | No | Filter by content type; `0`=all |
| `start_time` | int64 | No | **NEW** Unix timestamp — filter logs created at or after |
| `end_time` | int64 | No | **NEW** Unix timestamp — filter logs created at or before |
| `model_name` | string | No | **NEW** Fuzzy search by model name (`LIKE %model_name%`) |

**Response** (unchanged shape):
```json
{
  "success": true,
  "data": {
    "items": [ /* SecurityHitLogWithDetails[] */ ],
    "total": 1000,
    "page": 1,
    "page_size": 20
  }
}
```

---

### `GET /api/security/logs/export`

**Changes**: Adds `start_time`, `end_time`, and `model_name` query parameters (same as above). These are forwarded to the export query so exported data matches the filtered view.

---

### `GET /api/security/dashboard`

**Changes**: Response `summary` object now includes `today_interceptions`.

**Response**:
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_detections": 15000,
      "total_interceptions": 1200,
      "total_alerts": 800,
      "today_detections": 320,
      "today_interceptions": 45
    },
    "top_categories": [ /* ... */ ],
    "top_users": [ /* ... */ ],
    "top_models": [ /* ... */ ],
    "risk_distribution": { "low": 100, "medium": 50, "high": 30, "critical": 10 }
  }
}
```

---

## New Endpoints

### `POST /api/security/rules/:id/test`

Test a single rule against provided content.

**Path Params**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | int64 | Rule ID |

**Request Body**:
```json
{
  "content": "This is a test message containing 13800138000"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "detected": true,
    "action": 4,
    "action_name": "block",
    "risk_score": 80,
    "risk_level": 3,
    "processed_content": "This is a test message containing 138****8000",
    "matches": [
      {
        "rule_id": 1,
        "group_id": 2,
        "type": 2,
        "matched_text": "13800138000",
        "position": [42, 53]
      }
    ]
  }
}
```

---

### `PATCH /api/security/rules/:id/status`

Toggle a single rule's status (enable/disable) without sending the full rule body.

**Path Params**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | int64 | Rule ID |

**Request Body**:
```json
{
  "status": 0
}
```
- `status`: `0` = disabled, `1` = enabled

**Response**:
```json
{
  "success": true,
  "message": "状态更新成功"
}
```

---

### `POST /api/security/rules/batch-delete`

Delete multiple rules in a single operation.

**Request Body**:
```json
{
  "ids": [1, 2, 3, 5]
}
```

**Response**:
```json
{
  "success": true,
  "message": "批量删除成功"
}
```

---

### `POST /api/security/rules/batch-status`

Enable or disable multiple rules in a single operation.

**Request Body**:
```json
{
  "ids": [1, 2, 3, 5],
  "status": 1
}
```
- `status`: `0` = disabled, `1` = enabled

**Response**:
```json
{
  "success": true,
  "message": "批量状态更新成功"
}
```

---

### `GET /api/security/migration-status`

Check whether legacy Sensitive Words data has been migrated to the new system.

**Response**:
```json
{
  "success": true,
  "data": {
    "migrated": true,
    "migrated_count": 15,
    "migrated_at": 1718083200,
    "source_group_id": 42
  }
}
```

---

## DTO Changes

### `SecurityPolicyRequest` / `SecurityPolicyResponse`

**Added field**:
```go
Priority int `json:"priority"`  // default: 0, lower = higher precedence
```

### `SecurityDashboardResponse.Summary`

**Added field**:
```go
TodayInterceptions int `json:"today_interceptions"`
```

---

## Database Migrations

### Migration 1: Add `priority` to `security_user_policies`

```sql
-- MySQL
ALTER TABLE security_user_policies ADD COLUMN priority INT NOT NULL DEFAULT 0 AFTER whitelist_ips;
CREATE INDEX idx_security_policy_priority ON security_user_policies(priority);

-- PostgreSQL
ALTER TABLE security_user_policies ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_security_policy_priority ON security_user_policies(priority);

-- SQLite
ALTER TABLE security_user_policies ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
```

### Migration 2: Add index on `security_hit_logs.created_at`

```sql
-- MySQL / PostgreSQL / SQLite
CREATE INDEX idx_security_hit_created_at ON security_hit_logs(created_at);
```
