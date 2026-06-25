# Data Model: ai-security 模块

## 概述

所有 ai-security 数据表使用 `aisec_` 前缀，与官方表隔离。数据存储在同一个数据库实例中，但由模块内部独立迁移和管理。

---

## 实体关系图

```text
+----------------+       +----------------+       +----------------+
|  aisec_configs |       |  aisec_groups  |       |  aisec_rules   |
+----------------+       +----------------+       +----------------+
| id (PK)        |       | id (PK)        |<-----| group_id (FK)  |
| key            |       | name           |       | id (PK)        |
| value          |       | description    |       | name           |
| updated_at     |       | parent_id (FK) |       | type           |
+----------------+       | depth          |       | content        |
                         | path           |       | extra_config   |
+----------------+       | status         |       | action         |
|  aisec_policies|       | sort_order     |       | priority       |
+----------------+       +----------------+       | risk_score     |
| id (PK)        |              ^                | status         |
| user_id (FK)   |              |                | created_at     |
| group_id (FK)  |              |                | updated_at     |
| scope          |       +----------------+       +----------------+
| default_action |       |  aisec_words   |
| custom_response|       +----------------+
| whitelist_ips  |       | id (PK)        |
| priority       |       | group_id (FK)  |
| status         |       | word           |
+----------------+       | type           |
                         +----------------+
+----------------+       +----------------+       +----------------+
|aisec_hit_logs  |       |aisec_daily_stats|      |aisec_sync_state|
+----------------+       +----------------+       +----------------+
| id (PK)        |       | id (PK)        |       | id (PK)        |
| request_id     |       | date           |       | source         |
| user_id        |       | total_detected |       | last_sync_at   |
| token_id       |       | total_blocked  |       | sync_count     |
| model_name     |       | total_masked   |       +----------------+
| channel_id     |       | total_alerted  |
| direction      |       | total_reviewed |
| rule_id (FK)   |       | top_category   |
| group_id (FK)  |       | top_user       |
| risk_level     |       | top_model      |
| action         |       | created_at     |
| matched_text   |       | updated_at     |
| hit_reason     |       +----------------+
| ip             |
| created_at     |       +----------------+
+----------------+       |aisec_migrations|       +----------------+
                         +----------------+       |aisec_audit_logs|
                         | id (PK)        |       +----------------+
                         | version        |       | id (PK)        |
                         | applied_at     |       | user_id        |
                         +----------------+       | action_type    |
                                                 | target_type    |
                                                 | target_id      |
                                                 | old_value      |
                                                 | new_value      |
                                                 | created_at     |
                                                 +----------------+
```

---

## 表结构详细设计

### aisec_configs

模块全局配置。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| config_key | VARCHAR(128) UNIQUE | 配置键 |
| config_value | TEXT | 配置值（JSON 字符串） |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

**默认配置项：**
- `ai_security_enabled`: `true`
- `ai_model_name`: `gpt-4o-mini`
- `ai_timeout_seconds`: `3`
- `log_retention_days`: `30`
- `audit_log_retention_days`: `90`
- `default_risk_score`: `50`
- `max_group_depth`: `5`

---

### aisec_groups

规则分组，支持嵌套。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| name | VARCHAR(128) | 分组名称 |
| description | VARCHAR(255) | 说明 |
| parent_id | BIGINT FK | 父分组 ID，0 表示根分组 |
| depth | INT | 当前深度，根为 0 |
| path | VARCHAR(500) | Materialized Path，如 `/1/2/3` |
| status | INT | 0 停用，1 启用 |
| sort_order | INT | 排序 |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

**约束：**
- `depth <= 4`（根为 0，最大支持 5 层）
- `name` 在同级唯一

---

### aisec_rules

检测规则。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| code | VARCHAR(128) UNIQUE | 唯一规则编码，用于 seed 幂等 |
| group_id | BIGINT FK | 归属分组 |
| name | VARCHAR(128) | 规则名称 |
| type | INT | 1 Keyword, 2 Regex, 3 NER, 4 AI |
| content | TEXT | 规则内容 |
| extra_config | TEXT | 额外配置（JSON），如脱敏配置 |
| action | INT | 1 Allow, 2 Alert, 3 Mask, 4 Block, 5 Review |
| priority | INT | 优先级，数值越大优先级越高 |
| risk_score | INT | 风险分数 0-100 |
| status | INT | 0 停用，1 启用 |
| is_seed | BOOLEAN | 是否来自默认种子 |
| seed_modified_at | BIGINT | 用户修改种子规则的时间 |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

**约束：**
- `risk_score` 范围 0-100
- `type` 范围 1-4
- `action` 范围 1-5

---

### aisec_words

敏感词表（Keyword 规则的展开存储，可选）。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| group_id | BIGINT FK | 归属分组 |
| word | VARCHAR(255) | 敏感词 |
| type | INT | 1 普通词，2 正则模式 |
| status | INT | 0 停用，1 启用 |
| created_at | BIGINT | 创建时间戳 |

**说明：**
- 用于快速导入官方 sensitive-words 或批量管理关键词。
- Keyword 规则可引用分组下的 words，也可直接在 `content` 中存储逗号分隔的关键词。

---

### aisec_policies

用户策略，绑定用户与分组。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| user_id | INT FK | 用户 ID |
| group_id | BIGINT FK | 分组 ID |
| scope | INT | 1 仅请求，2 仅响应，3 双向 |
| default_action | INT | 默认动作 |
| custom_response | TEXT | 自定义拦截提示 |
| whitelist_ips | TEXT | IP 白名单（JSON 数组） |
| priority | INT | 优先级 |
| status | INT | 0 停用，1 启用 |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

**约束：**
- 同一用户对同一分组只能有一条启用策略。

---

### aisec_hit_logs

命中日志。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| request_id | VARCHAR(64) | 请求 ID |
| user_id | INT | 用户 ID |
| token_id | INT | Token ID |
| model_name | VARCHAR(128) | 模型名称 |
| channel_id | INT | 渠道 ID |
| direction | INT | 1 请求，2 响应 |
| rule_id | BIGINT FK | 命中规则 ID |
| group_id | BIGINT FK | 命中分组 ID |
| risk_level | INT | 1 Low, 2 Medium, 3 High, 4 Critical |
| action | INT | 执行动作 |
| matched_text | VARCHAR(500) | 命中片段 |
| hit_reason | VARCHAR(255) | 命中原因 |
| original_content_hash | VARCHAR(64) | 原始内容 SHA256 Hash |
| processed_content | TEXT | 脱敏后的内容（可选） |
| ip | VARCHAR(64) | 客户端 IP |
| created_at | BIGINT | 创建时间戳 |

**约束：**
- 不默认保存完整原始 prompt，只保存 hash 和命中片段。

---

### aisec_daily_stats

每日统计，用于 Dashboard 历史趋势。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| date | DATE | 日期 |
| total_detected | INT | 累计检测命中数 |
| total_blocked | INT | 拦截数 |
| total_masked | INT | 脱敏数 |
| total_alerted | INT | 告警数 |
| total_reviewed | INT | 审核数 |
| top_category | TEXT | 热门分类 JSON |
| top_user | TEXT | 热门用户 JSON |
| top_model | TEXT | 热门模型 JSON |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

---

### aisec_sync_state

与官方 sensitive-words 的同步状态。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| source | VARCHAR(64) | 来源标识，如 `official_sensitive_words` |
| last_sync_at | BIGINT | 上次同步时间 |
| sync_count | INT | 同步数量 |
| created_at | BIGINT | 创建时间戳 |
| updated_at | BIGINT | 更新时间戳 |

---

### aisec_migrations

模块迁移记录。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| version | VARCHAR(64) UNIQUE | 迁移版本 |
| applied_at | BIGINT | 应用时间 |

---

### aisec_audit_logs

管理员操作审计日志。

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT PK | 主键 |
| user_id | INT | 操作者用户 ID |
| action_type | VARCHAR(32) | 操作类型：create/update/delete/copy/enable/disable |
| target_type | VARCHAR(32) | 对象类型：group/rule/policy/config |
| target_id | BIGINT | 对象 ID |
| old_value | TEXT | 修改前内容 JSON |
| new_value | TEXT | 修改后内容 JSON |
| created_at | BIGINT | 创建时间戳 |

---

## 枚举定义

### Rule Type

| 值 | 含义 |
|---|---|
| 1 | Keyword |
| 2 | Regex |
| 3 | NER |
| 4 | AI |

### Action

| 值 | 含义 | 优先级 |
|---|---|---|
| 1 | Allow（放行） | 1 |
| 2 | Alert（告警） | 2 |
| 3 | Mask（脱敏） | 3 |
| 4 | Block（拦截） | 4 |
| 5 | Review（审核） | 5 |

### Risk Level

| 值 | 含义 | 分数范围 |
|---|---|---|
| 1 | Low | 0-25 |
| 2 | Medium | 26-50 |
| 3 | High | 51-75 |
| 4 | Critical | 76-100 |

### Scope

| 值 | 含义 |
|---|---|
| 1 | Request Only |
| 2 | Response Only |
| 3 | Both |

### Direction

| 值 | 含义 |
|---|---|
| 1 | Request |
| 2 | Response |

---

## 关键关系说明

1. **Group 自引用**：`aisec_groups.parent_id` 指向同一表，`path` 字段用于快速查询子孙分组。
2. **Rule 归属 Group**：规则必须属于一个分组；检测时取用户策略绑定的分组及其所有子孙分组的规则。
3. **Policy 绑定 User + Group**：一个用户可绑定多个分组策略，按优先级生效。
4. **HitLog 关联 Rule + Group**：记录命中时的规则 ID 和分组 ID，便于追溯。
5. **SyncState 记录官方敏感词同步**：用于幂等导入和增量同步。
