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

## 跟随上游 QuantumNous/new-api 升级

由于本模块采用「模块直挂」架构，业务代码全部收敛在 `custom/ai-security/` 内，
因此可以按常规 Git 流程同步官方 upstream 更新。

### 1. 准备工作

确保已添加官方仓库为 remote（通常命名为 `upstream`）：

```bash
git remote add upstream https://github.com/QuantumNous/new-api.git
```

### 2. 拉取上游并合并

```bash
# 获取上游最新代码
git fetch upstream main

# 方式 A：merge（推荐，保留本地提交历史，最不容易出错）
git checkout feature/ai-content-security
git merge upstream/main

# 方式 B：rebase（线性历史，但会重写提交哈希；若已推送到远端，之后需要强推）
git rebase upstream/main feature/ai-content-security
```

### 3. 解决冲突

合并时**绝大多数冲突只可能出现在以下挂载点文件**，模块内部文件
（`custom/ai-security/**`、`web/default/src/routes/_authenticated/ai-security/**`）
通常不会与上游冲突。

| 文件 | 你的改动 | 处理原则 |
|---|---|---|
| `main.go` | `ai_security.Init()` 初始化调用 | 保留上游新增的初始化调用，同时保留 `ai_security.Init()` |
| `router/api-router.go` | `/api/ai-security/*` 路由注册 | 保留上游路由调整，同时保留 `ai_security.RegisterRoutes()` |
| `router/relay-router.go` | `CheckRequest()` / `CheckResponse()` 挂载 | 保持「官方 `SecurityCheck` 在前、ai-security 在后」的顺序 |
| `web/default/src/i18n/config.ts` | `keySeparator: false` | 保留该配置，它是模块扁平 i18n 键生效的前提 |
| `web/default/rsbuild.config.ts` | `@custom` alias | 保留别名定义 |
| `web/default/tsconfig.json` | `@custom/*` paths | 保留路径映射 |
| `web/default/tsconfig.app.json` | `@custom/*` paths + include | 保留路径映射与 include |
| `web/default/src/features/system-settings/security/section-registry.tsx` | AI Content Security 菜单入口 | 保留菜单项 |

### 4. 验证构建

冲突解决后，按顺序验证：

```bash
# Go 后端
go build ./...

# 前端（需已安装 Bun）
cd web/default
bun install
bun run typecheck
bun run build
```

### 5. 数据库兼容性说明

- 官方表结构变更由上游 GORM `AutoMigrate` 自动处理。
- `aisec_*` 表由 `custom/ai-security/migration/` 独立管理，互不影响。
- 升级前建议备份数据库，尤其是生产环境。

### 6. 推送

验证通过后推送到你的远端：

```bash
git push origin feature/ai-content-security
# 如需同步到 new-api
git push new-api feature/ai-content-security
```

如果使用了 rebase 且远端已存在旧提交，则可能需要 `--force-with-lease` 强推。

## Docker 部署

### 注意：docker-compose.yml 默认指向官方镜像

项目根目录的 `docker-compose.yml` 第 19 行默认使用：

```yaml
image: calciumion/new-api:latest
```

这是 QuantumNous/new-api 官方发布到 Docker Hub 的预编译镜像，**不包含你的 ai-security 模块**。要部署你自己的分支，需要改用本地构建。

### 方案 A：服务器本地构建（推荐）

在服务器上克隆你自己的仓库并切换到对应分支：

```bash
git clone -b feature/ai-content-security https://github.com/Channing-xiao/c-api.git
cd c-api
```

修改 `docker-compose.yml`，将 `new-api` 服务从拉取镜像改为本地构建：

```yaml
services:
  new-api:
    build:
      context: .
      dockerfile: Dockerfile
    image: new-api:local
    container_name: new-api
```

启动：

```bash
docker compose up -d --build
```

`Dockerfile` 会通过 `COPY . .` 把当前目录源码（含 `custom/ai-security/`）打包进镜像，并自动完成前端构建、Go 编译。

### 方案 B：先构建镜像再推送仓库

如果你有多台服务器，不想每台都重新构建：

```bash
git clone -b feature/ai-content-security https://github.com/Channing-xiao/c-api.git
cd c-api
docker build -t your-dockerhub-username/new-api:ai-security .
docker push your-dockerhub-username/new-api:ai-security
```

然后在服务器 `docker-compose.yml` 中改为：

```yaml
services:
  new-api:
    image: your-dockerhub-username/new-api:ai-security
```

### 不需要修改的地方

| 项目 | 是否需修改 | 原因 |
|---|---|---|
| `Dockerfile` 里的 `github.com/QuantumNous/new-api/common.Version` | 否 | 这是 Go 模块路径，用于编译时注入版本号，不是 Git 仓库地址 |
| `go.mod` 里的 `module github.com/QuantumNous/new-api` | 否 | 整个项目的 import 路径都基于该模块名 |
| `ENTRYPOINT ["/new-api"]` | 否 | 本模块未修改入口，且开发文档要求不动 Docker ENTRYPOINT |
| `Dockerfile.dev` / `docker-compose.dev.yml` | 可选 | 开发版默认就是从本地源码构建，可直接使用 |

### 数据库配置

`docker-compose.yml` 默认使用 PostgreSQL，并通过 `SQL_DSN` 连接。生产环境务必修改默认密码：

```yaml
environment:
  - SQL_DSN=postgresql://root:YOUR_PASSWORD@postgres:5432/new-api
  - REDIS_CONN_STRING=redis://:YOUR_PASSWORD@redis:6379
```

如果使用 MySQL，请按 `docker-compose.yml` 内注释切换对应的服务和 `SQL_DSN`。

## 开发说明

- 后端使用 Go + Gin + GORM。
- 前端使用 React 19 + TypeScript + TanStack Router + Base UI。
- 所有 JSON 序列化通过 `common` 包封装函数完成。
- 数据库需同时兼容 SQLite、MySQL、PostgreSQL。
