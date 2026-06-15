# Data Model — AI Content Security Management v2

**Scope**: Existing data model for the AI Content Security module. The group-status bug fix does not add or modify columns; it only fixes the update path for the existing `status` column.

## Entities

### SecurityGroup

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 (PK) | Auto-increment primary key |
| `name` | string | Group name (unique) |
| `description` | string | Optional description |
| `status` | int | `1` = enabled, `0` = disabled |
| `parent_id` | int64 | Parent group ID (`0` = root) |
| `depth` | int | Nesting depth (max 5) |
| `path` | string | Materialized path (e.g., `/1/2`) |
| `sort_order` | int | Display order |
| `created_at` | int64 | Unix timestamp |
| `updated_at` | int64 | Unix timestamp |

### SecurityRule

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 (PK) | Primary key |
| `group_id` | int64 (FK) | Belongs to SecurityGroup |
| `name` | string | Rule name |
| `type` | int | Keyword, Regex, NER, AI Detection |
| `content` | string | Pattern / model identifier |
| `extra_config` | string | JSON-encoded extra options |
| `action` | int | Pass, Alert, Mask, Block, Review |
| `priority` | int | Evaluation order |
| `risk_score` | int | 0-100 risk score |
| `status` | int | `1` = enabled, `0` = disabled |
| `created_at` | int64 | Unix timestamp |
| `updated_at` | int64 | Unix timestamp |

### SecurityUserPolicy

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 (PK) | Primary key |
| `user_id` | int | Target user |
| `group_id` | int64 (FK) | Linked SecurityGroup |
| `scope` | int | Request / Response / Both |
| `default_action` | int | Action when rule matches |
| `custom_response` | string | Message for block action |
| `whitelist_ips` | string | Comma-separated IPs |
| `priority` | int | Lower = higher precedence |
| `status` | int | `1` = enabled, `0` = disabled |
| `created_at` | int64 | Unix timestamp |
| `updated_at` | int64 | Unix timestamp |

### SecurityHitLog

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 (PK) | Primary key |
| `request_id` | string | Trace ID |
| `user_id` | int | User who triggered detection |
| `channel_id` | int | Upstream channel |
| `model_name` | string | AI model name |
| `token_id` | int | API token used |
| `rule_id` | int64 (FK) | Matching rule |
| `group_id` | int64 (FK) | Matching group |
| `content_type` | int | Request / Response |
| `action` | int | Action taken |
| `risk_level` | int | Low / Medium / High / Critical |
| `risk_score` | int | 0-100 |
| `original_content_hash` | string | SHA256 of original content |
| `processed_content` | string | Masked/replaced content |
| `match_detail` | string | JSON match metadata |
| `ip` | string | Client IP |
| `created_at` | int64 | Unix timestamp |

## State Transitions

### SecurityGroup.status

- `1` (enabled) → `0` (disabled): Group and its rules no longer participate in detection.
- `0` (disabled) → `1` (enabled): Group and its enabled rules participate in detection again.

The bug fix ensures the `PUT /api/security/groups/:id` update path correctly persists this transition.

## Validation Rules

- `status` must be `0` or `1`.
- `name` is required and max 128 characters.
- `description` max 255 characters.
- `parent_id` must reference an existing group and not exceed max depth (5).
