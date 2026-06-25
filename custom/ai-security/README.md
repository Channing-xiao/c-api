# AI Content Security Module

AI 内容安全高级检测模块，以「模块直挂（工程化版）」方式集成到 new-api。

## 设计原则

- **产品体验**：作为官方 `/system-settings/security/sensitive-words` 的高级功能入口。
- **代码架构**：完全独立，所有业务代码位于 `custom/ai-security/`，最小化主项目挂载点。

## 目录结构

```text
custom/ai-security/
├── api/              # Gin handlers
├── service/          # Business logic
├── engine/           # Detection engines (keyword, regex, NER, AI)
├── model/            # GORM models for aisec_* tables
├── migration/        # Module migrations
├── seed/             # Default rules seed
├── middleware/       # Request/response detection middleware
├── web/              # Frontend pages, components, API clients
│   ├── api/
│   ├── components/
│   ├── pages/
│   ├── routes/       # TanStack Router route files
│   └── i18n/
├── install.sh        # Idempotent install script
├── version.json      # Module version metadata
└── README.md         # This file
```

## 安装

```bash
bash custom/ai-security/install.sh
```

实际迁移与种子数据在 `ai_security.Init()` 中完成，`install.sh` 仅做校验与版本输出。

## 主项目挂载点

后端：
- `router/api-router.go`：`ai_security.RegisterRoutes(apiRouter)`
- `main.go`：`ai_security.Init()`
- `router/relay-router.go`：`ai_security.CheckRequest()` / `CheckResponse()`

前端：
- `web/default/src/routes/_authenticated/ai-security/*`：路由文件
- `web/default/src/features/system-settings/security/section-registry.tsx`：菜单入口

## 数据表

所有表使用 `aisec_` 前缀，独立存储：

- `aisec_configs`
- `aisec_groups`
- `aisec_rules`
- `aisec_words`
- `aisec_policies`
- `aisec_hit_logs`
- `aisec_daily_stats`
- `aisec_sync_state`
- `aisec_audit_logs`
- `aisec_migrations`

## API 概览

- `GET/PUT /api/ai-security/configs`
- `GET /api/ai-security/status`
- `POST /api/ai-security/install`
- `GET/POST/PUT/PATCH/DELETE /api/ai-security/groups/*`
- `GET/POST/PUT/DELETE /api/ai-security/rules/*`
- `POST /api/ai-security/rules/:id/test`
- `GET/POST/PUT/DELETE /api/ai-security/policies/*`
- `GET /api/ai-security/logs`
- `GET /api/ai-security/logs/export`
- `GET /api/ai-security/dashboard`
- `POST /api/ai-security/sync/official-sensitive-words`

## 升级兼容性

- 模块升级不影响官方 new-api 升级。
- 主项目冲突仅可能出现在 4 个后端挂载点 + 2 个前端挂载点。
- 删除 `custom/ai-security/` 后项目仍可编译运行。

## 开发说明

- 后端使用 Go + Gin + GORM。
- 前端使用 React 19 + TypeScript + TanStack Router + Base UI。
- 所有 JSON 序列化通过 `common` 包封装函数完成。
- 数据库需同时兼容 SQLite、MySQL、PostgreSQL。
