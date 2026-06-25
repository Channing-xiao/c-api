# Implementation Plan: ai-security 模块重构

## 计划概述

本计划基于 `spec.md`、`research.md`、`data-model.md` 和接口契约，将当前分散在官方目录的 ai-security 代码重构为独立的 `custom/ai-security/` 模块，并最小化主项目改动。

---

## Phase 0: 准备与基础设施

### 任务 P0-1: 初始化 custom/ai-security 目录结构

**目标**：在 `custom/ai-security/` 下建立模块目录骨架。

**输出目录**：

```text
custom/ai-security/
├── api/
├── service/
├── engine/
├── model/
├── migration/
├── seed/
├── web/
│   ├── api/
│   ├── components/
│   ├── hooks/
│   ├── pages/
│   ├── routes/
│   ├── constants.ts
│   └── i18n/
├── install.sh
├── version.json
└── README.md
```

**验收标准**：
- 目录结构与规格一致。
- 每个目录下创建 `.gitkeep` 或占位文件。

---

### 任务 P0-2: 设计模块初始化入口

**目标**：设计 `custom/ai-security` 对外暴露的最小初始化接口。

**关键接口**：

```go
package ai_security

func Init() error
func RegisterRoutes(router *gin.RouterGroup)
func RegisterRelayMiddleware(router *gin.RouterGroup)
func CheckRequest() gin.HandlerFunc
func CheckResponse() gin.HandlerFunc
```

**验收标准**：
- 主项目只需调用 `ai_security.Init()` 和注册路由/中间件。
- 模块内部完成数据库迁移、配置加载、缓存初始化。

---

### 任务 P0-3: 实现数据库迁移脚本

**目标**：创建 `aisec_*` 表的迁移脚本，不修改官方迁移逻辑。

**输出**：
- `custom/ai-security/migration/001_init_tables.go`
- 支持 SQLite、MySQL、PostgreSQL

**验收标准**：
- `Init()` 自动执行未应用的迁移。
- `aisec_migrations` 表记录已应用版本。

---

## Phase 1: 后端核心

### 任务 P1-1: 实现配置管理

**目标**：实现 `aisec_configs` 的读写和默认配置初始化。

**输出**：
- `custom/ai-security/service/config.go`
- `custom/ai-security/api/config.go`

**验收标准**：
- `GET /api/ai-security/configs` 返回默认配置。
- `PUT /api/ai-security/configs` 可更新配置并失效缓存。

---

### 任务 P1-2: 实现分组管理

**目标**：实现分组 CRUD、嵌套、复制和状态管理。

**输出**：
- `custom/ai-security/service/group.go`
- `custom/ai-security/api/group.go`
- `custom/ai-security/model/group.go`

**验收标准**：
- 支持最多 5 层嵌套。
- 删除父分组时级联删除子分组和规则。
- 复制分组时同时复制规则。

---

### 任务 P1-3: 实现规则管理

**目标**：实现规则 CRUD、测试、批量操作。

**输出**：
- `custom/ai-security/service/rule.go`
- `custom/ai-security/api/rule.go`
- `custom/ai-security/model/rule.go`

**验收标准**：
- 支持 4 种规则类型。
- 支持测试接口返回检测结果。
- 规则变更后自动失效缓存。

---

### 任务 P1-4: 实现策略管理

**目标**：实现用户策略 CRUD。

**输出**：
- `custom/ai-security/service/policy.go`
- `custom/ai-security/api/policy.go`
- `custom/ai-security/model/policy.go`

**验收标准**：
- 同一用户对同一分组只能有一条启用策略。
- 策略变更后自动失效缓存。

---

### 任务 P1-5: 实现检测引擎

**目标**：实现 Keyword、Regex、NER、AI 四种检测引擎。

**输出**：
- `custom/ai-security/engine/base.go`
- `custom/ai-security/engine/keyword.go`
- `custom/ai-security/engine/regex.go`
- `custom/ai-security/engine/ner.go`
- `custom/ai-security/engine/ai.go`
- `custom/ai-security/service/detector.go`

**验收标准**：
- 各引擎返回统一结构。
- AI 引擎超时 3 秒降级。
- 多引擎并行执行，结果合并。

---

### 任务 P1-6: 实现命中日志

**目标**：实现异步批量写入命中日志。

**输出**：
- `custom/ai-security/service/hitlog.go`
- `custom/ai-security/model/hitlog.go`
- `custom/ai-security/api/log.go`

**验收标准**：
- 检测命中后异步记录日志。
- 不保存完整 prompt，只保存 hash 和命中片段。
- 服务关闭时刷新剩余日志。

---

### 任务 P1-7: 实现 Dashboard 统计

**目标**：实现 Dashboard 后端统计接口。

**输出**：
- `custom/ai-security/service/dashboard.go`
- `custom/ai-security/api/dashboard.go`
- `custom/ai-security/model/dailystats.go`

**验收标准**：
- 返回 summary、risk_distribution、top_categories、top_users、top_models。
- 支持时间范围过滤。

---

### 任务 P1-8: 实现请求检测中间件

**目标**：实现 `CheckRequest` 中间件。

**输出**：
- `custom/ai-security/middleware/check_request.go`

**验收标准**：
- 只检测聊天补全类接口。
- 根据策略动作执行放行、告警、脱敏、拦截、审核。
- 脱敏时同步更新请求体。

---

### 任务 P1-9: 实现响应检测中间件（第二阶段）

**目标**：实现 `CheckResponse` 中间件。

**输出**：
- `custom/ai-security/middleware/check_response.go`

**验收标准**：
- 支持非流式和流式响应检测。
- 命中规则时按动作处理。

---

### 任务 P1-10: 实现 install.sh

**目标**：实现可重复执行的安装脚本。

**输出**：
- `custom/ai-security/install.sh`
- `custom/ai-security/version.json`

**验收标准**：
- 执行迁移。
- 初始化默认配置。
- 写入默认规则种子。
- 重复执行不覆盖用户数据。

---

### 任务 P1-11: 实现官方敏感词同步

**目标**：实现从官方 sensitive-words 单向导入。

**输出**：
- `custom/ai-security/service/sync.go`
- `custom/ai-security/api/sync.go`

**验收标准**：
- 提供 `/api/ai-security/sync/official-sensitive-words` 接口。
- 导入后官方 sensitive-words 保持不变。
- 记录同步状态。

---

## Phase 2: 前端

### 任务 P2-1: 搭建前端页面骨架

**目标**：在 `custom/ai-security/web/` 下创建前端页面和路由。

**输出**：
- `custom/ai-security/web/routes/*.tsx`
- `custom/ai-security/web/pages/*.tsx`
- `custom/ai-security/web/components/ai-security-layout.tsx`
- `custom/ai-security/web/components/ai-security-tabs.tsx`

**验收标准**：
- `/ai-security` 及子路径可访问。
- 页面布局包含 Tab 导航。

---

### 任务 P2-2: 实现 Dashboard 页面

**目标**：实现数据看板页面。

**输出**：
- `custom/ai-security/web/pages/dashboard-page.tsx`
- `custom/ai-security/web/components/*-chart.tsx`

**验收标准**：
- 展示 summary 卡片。
- 展示风险分布、热门用户、热门模型。
- 支持时间范围过滤。

---

### 任务 P2-3: 实现 Groups 页面

**目标**：实现分组管理页面。

**输出**：
- `custom/ai-security/web/pages/group-page.tsx`
- `custom/ai-security/web/components/group-form-modal.tsx`

**验收标准**：
- 展示分组列表（树形或表格）。
- 支持创建、编辑、删除、复制。

---

### 任务 P2-4: 实现 Rules 页面

**目标**：实现规则管理页面。

**输出**：
- `custom/ai-security/web/pages/rule-page.tsx`
- `custom/ai-security/web/components/rule-form-modal.tsx`
- `custom/ai-security/web/components/rule-tester.tsx`

**验收标准**：
- 展示规则列表。
- 支持创建、编辑、复制、删除、测试。
- 支持批量操作。

---

### 任务 P2-5: 实现 Policies 页面

**目标**：实现策略管理页面。

**输出**：
- `custom/ai-security/web/pages/policy-page.tsx`
- `custom/ai-security/web/components/policy-form-modal.tsx`

**验收标准**：
- 展示策略列表。
- 支持创建、编辑、删除。

---

### 任务 P2-6: 实现 Logs 页面

**目标**：实现命中日志页面。

**输出**：
- `custom/ai-security/web/pages/log-page.tsx`
- `custom/ai-security/web/components/log-detail-drawer.tsx`

**验收标准**：
- 展示命中日志列表。
- 支持多维过滤。
- 支持导出。

---

### 任务 P2-7: 实现 i18n

**目标**：实现 zh/en 翻译。

**输出**：
- `custom/ai-security/web/i18n/ai-security.json`

**验收标准**：
- 所有用户可见文本支持翻译。

---

## Phase 3: 主项目挂载

### 任务 P3-1: 注册后端路由

**目标**：在 `router/api-router.go` 中注册 ai-security 路由。

**改动**：

```go
import "github.com/QuantumNous/new-api/custom/ai-security"

// 在 SetApiRouter 中
ai_security.RegisterRoutes(apiRouter)
```

**验收标准**：
- `/api/ai-security/*` 可访问。

---

### 任务 P3-2: 注册初始化

**目标**：在 `main.go` 的 `InitResources()` 中调用 ai-security 初始化。

**改动**：

```go
// 在 InitResources 中
if err := ai_security.Init(); err != nil {
    common.SysError("ai-security init failed: " + err.Error())
}
```

**验收标准**：
- 启动时自动执行迁移和默认数据初始化。

---

### 任务 P3-3: 注册 Relay 检测中间件

**目标**：在 `router/relay-router.go` 中挂载请求/响应检测。

**改动**：

```go
httpRouter.Use(ai_security.CheckRequest())
httpRouter.Use(ai_security.CheckResponse())
```

**验收标准**：
- 请求和响应经过 ai-security 检测。

---

### 任务 P3-4: 注册前端路由和菜单

**目标**：在主项目中注册 `/ai-security` 路由和菜单入口。

**改动**：
- 路由注册：引入 `custom/ai-security/web/routes/*`。
- 菜单：在 `System Settings > Security` 下增加入口。

**验收标准**：
- 前端可访问 `/ai-security`。
- 菜单位置符合规格。

---

## Phase 4: 测试与验收

### 任务 P4-1: 后端单元测试

**目标**：为检测引擎、服务层编写单元测试。

**输出**：
- `custom/ai-security/engine/*_test.go`
- `custom/ai-security/service/*_test.go`

**验收标准**：
- 核心引擎覆盖主要分支。
- 规则解析和动作计算正确。

---

### 任务 P4-2: 集成测试

**目标**：验证完整检测链路。

**测试场景**：
- 安装后默认规则存在。
- 敏感请求被拦截。
- 敏感响应被脱敏。
- 关闭 ai-security 后官方 sensitive-words 仍工作。
- 重复 install 不覆盖用户数据。

**验收标准**：
- 所有场景通过。

---

### 任务 P4-3: 升级兼容性测试

**目标**：验证同步官方 new-api 时代码冲突范围。

**验收标准**：
- 冲突只出现在 router/api-router.go、router/relay-router.go、main.go、前端路由/菜单文件。
- 移除 `custom/ai-security/` 后项目可编译运行。

---

## 依赖关系

```text
P0-1 → P0-2 → P0-3
P0-3 → P1-1, P1-2, P1-3, P1-4, P1-6
P1-2, P1-3 → P1-5
P1-5 → P1-8, P1-9
P1-6, P1-7 → P2-2, P2-6
P1-2 → P2-3
P1-3 → P2-4
P1-4 → P2-5
P1-1, P1-10, P1-11 → P3-2
P2-1 → P2-2, P2-3, P2-4, P2-5, P2-6
P1-8, P1-9 → P3-3
P2-7 → P3-4
P3-1, P3-2, P3-3, P3-4 → P4-2, P4-3
```

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| 主项目挂载点冲突 | 高 | 保持挂载代码最小化，使用独立函数调用 |
| AI 检测超时影响实时性 | 中 | 设置 3 秒超时，超时降级本地规则 |
| 默认规则误报 | 中 | 默认规则使用较低风险分数，管理员可调整 |
| 迁移脚本跨数据库兼容性 | 中 | 使用 GORM AutoMigrate 并测试 SQLite/MySQL/PostgreSQL |
| 前端路由与官方路由冲突 | 低 | 使用 `/ai-security` 独立前缀 |

## 完成标准

- [ ] `custom/ai-security/` 目录包含全部业务代码。
- [ ] 主项目改动不超过 4 个后端挂载点 + 2 个前端挂载点。
- [ ] 所有 `aisec_*` 表创建成功。
- [ ] `install.sh` 可重复执行且不覆盖用户数据。
- [ ] `/api/ai-security/*` 接口正常工作。
- [ ] `/ai-security` 前端页面可访问且菜单位置正确。
- [ ] 请求/响应检测链路工作正常。
- [ ] 与官方 sensitive-words 完全解耦。
- [ ] 集成测试和升级兼容性测试通过。
