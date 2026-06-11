# AI 内容安全管理模块 — 前端设计方案

> **版本**：v2.2
> **日期**：2026-06-11
> **状态**：设计评审（已适配当前后端实现）
> **关联文档**：New-API-AI内容安全管理模块.md、后端API补充清单.md

---

## 一、现状诊断

### 1.1 已完成后端

后端已完成完整的 AI 内容安全管理模块：

| 模块 | 状态 | API 路径 |
|------|------|----------|
| 敏感词分组管理 | ✅ | `GET/POST/PUT/DELETE /api/security/groups` |
| 敏感词规则管理 | ✅ | `GET/POST/PUT/DELETE /api/security/rules` |
| 用户策略管理 | ✅ | `GET/POST/PUT/DELETE /api/security/policies` |
| 审计日志 | ✅ | `GET /api/security/logs`, `GET /api/security/logs/export` |
| 统计看板 | ✅ | `GET /api/security/dashboard` |
| 规则测试 | ✅ | `POST /api/security/rules/:id/test` |
| 请求/响应检测 | ✅ | 中间件 `SecurityCheck()` / `SecurityCheckResponse()` |

### 1.2 前端现状

**第一套系统（新 AI 内容安全模块）**：

代码已存在于 `web/default/src/features/security/`，包含：

- `pages/dashboard-page.tsx` — 统计看板
- `pages/group-page.tsx` — 敏感词分组管理
- `pages/rule-page.tsx` — 检测规则管理
- `pages/policy-page.tsx` — 用户策略管理
- `pages/log-page.tsx` — 审计日志
- `api/security.ts` — API 客户端

路由已注册（TanStack Router）：
- `/security` → Dashboard
- `/security/groups` → Groups
- `/security/rules` → Rules
- `/security/policies` → Policies
- `/security/logs` → Logs

**第二套系统（旧版敏感词）**：

位于 `system-settings/security/sensitive-words`，是一个简单的开关 + 文本框：
- `CheckSensitiveEnabled` — 是否启用
- `CheckSensitiveOnPromptEnabled` — 是否检测 Prompt
- `SensitiveWords` — 每行一个关键词的文本框

### 1.3 核心问题

| 编号 | 问题 | 影响 |
|------|------|------|
| P1 | `/security` 页面**没有子导航** | 用户进入 Dashboard 后，无法发现 Groups/Rules/Policies/Logs 页面 |
| P2 | 侧边栏 **Security 只有单个入口**，无子菜单 | 用户需要手动输入 URL 才能访问 groups/rules 等页面 |
| P3 | **新旧两套系统割裂** | system-settings 里的 Sensitive Words 是旧系统，和新模块毫无关联，用户困惑 |
| P4 | 旧系统与新模块**数据不互通** | 旧系统的 `SensitiveWords` 文本框和新模块的 `security_rules` 表是两套独立数据 |
| P5 | 页面内**缺少上下文导航** | 用户不知道自己身在何处，如何返回 |
| P6 | **缺少统一的数据获取层设计** | 各页面各自管理请求状态，无缓存、无乐观更新 |
| P7 | **筛选状态无法分享/持久化** | 刷新页面后筛选条件丢失，无法通过 URL 分享特定视图 |

---

## 二、设计目标

1. **统一入口**：将 AI 内容安全模块作为独立的管理模块，与 System Settings 平级
2. **清晰导航**：在 `/security` 模块内提供 Tab 导航，让用户可以在 5 个页面间自由切换
3. **整合旧系统**：废弃或迁移旧版 `SensitiveWords` 文本框，统一使用新模块的 `security_groups` + `security_rules`
4. **权限控制**：基于权限（`security:read`, `security:write`）控制访问与操作，而非简单的角色比较
5. **用户体验**：提供 Tab 导航、操作反馈、空状态、骨架屏、加载状态等完整体验
6. **状态可分享**：所有筛选条件同步到 URL 查询参数，支持刷新保留、链接分享
7. **性能优化**：大数据表格使用分页 + 虚拟滚动，实时轮询在后台标签页自动暂停
8. **安全合规**：用户输入内容纯文本渲染，敏感命中信息自动脱敏展示

---

## 三、方案总览

### 3.1 架构调整

```
Before（现状）:
┌──────────────────────────────────────────────────────────┐
│  Sidebar                                                 │
│  ├─ ...                                                  │
│  ├─ Security ──→ /security (Dashboard，无子导航)          │
│  └─ System Settings ──→ /system-settings/security        │
│         ├─ Rate Limiting                                 │
│         ├─ Sensitive Words ← 旧系统（文本框）              │
│         └─ SSRF Protection                               │
└──────────────────────────────────────────────────────────┘

After（目标）:
┌──────────────────────────────────────────────────────────┐
│  Sidebar                                                 │
│  ├─ ...                                                  │
│  ├─ Security ──→ /security                               │
│  │      ↑ Tab 导航切换页面                                │
│  │      [Dashboard][Groups][Rules][Policies][Audit Logs] │
│  │                                                       │
│  └─ System Settings ──→ /system-settings/security        │
│         ├─ Rate Limiting                                 │
│         └─ SSRF Protection                               │
│         (Removed: Sensitive Words — 由新模块统一替代)      │
└──────────────────────────────────────────────────────────┘
```

**设计决策**：
- 侧边栏 Security 保持**单入口**（不展开子菜单），由页面内 Tab 负责子页面切换
- **Tab 点击触发路由跳转**（`navigate()`），非视图切换。每个 Tab 对应独立路由（`/security`、`/security/groups` 等），确保 URL 可分享、刷新可恢复
- 原因：Security 内部 5 个页面数据流相似（管理后台风格），Tab 切换在视觉上更符合用户心智模型；同时保持 URL 的语义化和可分享性；侧边栏保持扁平，减少层级
- **响应式适配**：Tab 导航在容器宽度不足时支持横向滚动（`overflow-x-auto`），小屏幕下保持可用性；筛选区在小屏幕下垂直堆叠

### 3.2 页面布局统一设计

每个 Security 子页面采用统一的布局框架：

```
┌─────────────────────────────────────────────────────────┐
│  AI 内容安全                           [Create Group]   │  ← 标题 + 主操作
├─────────────────────────────────────────────────────────┤
│  [Dashboard] [Groups] [Rules] [Policies] [Audit Logs]   │  ← Tab 子导航（路由跳转）
├─────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────┐   │
│  │  Filters: [Status ▼] [Search 🔍]                │   │  ← 筛选区（URL 同步）
│  ├─────────────────────────────────────────────────┤   │
│  │  Name      │ Description │ Status │ Actions      │   │  ← 数据表格
│  │  ─────────────────────────────────────────────  │   │
│  │  个人隐私   │ 保护个人隐私  │ 启用   │ Edit Delete│   │
│  └─────────────────────────────────────────────────┘   │
│                              < 1 2 3 ... >              │  ← 分页（默认 20/页）
└─────────────────────────────────────────────────────────┘
```

**响应式要点**：
- 桌面端（≥1024px）：标题 + 操作按钮左右排列，筛选区横向排列
- 平板端（768px–1023px）：筛选区可折叠为「高级筛选」按钮
- 移动端（<768px）：Tab 导航横向滚动，筛选区垂直堆叠，表格横向滚动

---

## 四、路由与导航设计

### 4.1 路由结构调整

| 路径 | 页面组件 | 说明 |
|------|----------|------|
| `/security` | `SecurityDashboardPage` | 统计看板（默认页） |
| `/security/groups` | `SecurityGroupPage` | 敏感词分组管理 |
| `/security/rules` | `SecurityRulePage` | 检测规则管理 |
| `/security/policies` | `SecurityPolicyPage` | 用户策略管理 |
| `/security/logs` | `SecurityLogPage` | 审计日志 |

**路由守卫**：所有页面统一要求 `hasPermission('security:read')`。ADMIN 角色默认拥有此权限。无权限时渲染 `<EmptyState type="forbidden" />`，而非空白页或控制台错误。

**URL 查询参数同步**：
所有筛选状态同步到 URL，确保刷新保留、可分享链接。

```typescript
// Groups 页面示例
/security/groups?status=active&search=隐私&page=1&pageSize=20

// Logs 页面示例
/security/logs?start=2026-06-01&end=2026-06-11&actions=block,alert&level=high&page=1
```

### 4.2 侧边栏导航调整

修改 `useSidebarData.ts`，Security 保持为单入口链接：

```typescript
// 目标：Security 为单入口，子页面通过 Tab 切换
{
  title: t('Security'),
  url: '/security',
  activeUrls: ['/security'],     // 所有 /security/* 都高亮
  icon: ShieldAlert,
}
```

**注意**：`activeUrls` 使用 `/security` 前缀匹配，所有子页面侧边栏 Security 项保持高亮。

### 4.3 页面内 Tab 导航

在 `/security/*` 的所有页面顶部，增加统一的 Tab 导航组件：

```typescript
// features/security/components/security-tabs.tsx
const tabs = [
  { label: 'Dashboard', path: '/security', permission: 'security:read' },
  { label: 'Groups', path: '/security/groups', permission: 'security:read' },
  { label: 'Rules', path: '/security/rules', permission: 'security:read' },
  { label: 'Policies', path: '/security/policies', permission: 'security:read' },
  { label: 'Audit Logs', path: '/security/logs', permission: 'security:read' },
]

// Tab 点击触发路由跳转，而非视图切换
function handleTabChange(path: string) {
  navigate({ to: path, search: (prev) => prev }) // 保留当前 URL 其他参数
}
```

使用 Base UI 的 `Tabs` 组件或自定义 NavTab 组件，当前路由高亮显示。

**优势**：
- 即使用户通过直接 URL 访问某个子页面，也能立即看到全部导航选项
- 与侧边栏单入口设计配合，避免导航层级过深
- 减少用户迷失，切换仅需一次点击
- 每个子页面都有独立的、可分享的 URL

### 4.4 URL 状态同步技术方案

**问题**：各页面的筛选状态（如 `status`、`search`、`page`）需要与 URL 查询参数双向同步，确保刷新保留、链接可分享。

**方案**：封装通用 Hook `useUrlFilters`，基于 TanStack Router 的 `useSearch` + Zod 校验。

```typescript
// features/security/hooks/use-url-filters.ts
import { useSearch, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'

const groupFilterSchema = z.object({
  status: z.enum(['all', 'active', 'inactive']).catch('all'),
  search: z.string().catch(''),
  page: z.coerce.number().catch(1),
  pageSize: z.coerce.number().catch(20),
})

export type GroupFilters = z.infer<typeof groupFilterSchema>

export function useGroupFilters() {
  const navigate = useNavigate({ from: '/security/groups' })
  const search = useSearch({ from: '/security/groups', strict: false })
  const filters = groupFilterSchema.parse(search)

  const setFilters = (updater: Partial<GroupFilters>) => {
    navigate({
      search: (prev) => ({ ...prev, ...updater, page: updater.page ?? 1 }),
      replace: true, // 不留下历史记录
    })
  }

  return [filters, setFilters] as const
}
```

**设计要点**：
- 使用 `zod.catch()` 提供默认值，避免非法参数导致页面崩溃
- 筛选条件变更时重置 `page` 到第 1 页（避免筛选后出现在空页）
- 使用 `replace: true` 避免筛选操作污染浏览器历史栈
- 每个子页面独立定义自己的 Filter Schema，保持类型安全

### 4.5 权限控制设计

**现状说明**：当前后端所有 Security API 均使用 `middleware.AdminAuth()` 进行保护（`router/api-router.go`），即只有 ADMIN 角色可访问和管理。后端尚未实现基于 `security:read` / `security:write` 的细粒度权限码。

**前端适配方案**：

| 检查维度 | 实现方式 | 说明 |
|----------|----------|------|
| 页面访问 | `user.role === 'admin'` | 非 ADMIN 用户显示 `<EmptyState type="forbidden" />` |
| 管理操作 | `user.role === 'admin'` | 所有「创建/编辑/删除/复制/批量」按钮仅对 ADMIN 显示 |
| 危险操作 | 二次确认弹窗 | 删除分组/规则/策略时强制确认 |

**实现方式**：
- 前端通过 `useAuth()` 获取当前用户角色
- 提供组件 `<AdminGuard fallback={null}>{children}</AdminGuard>`

```typescript
// features/security/components/admin-guard.tsx
export function AdminGuard({
  fallback = null,
  children,
}: {
  fallback?: React.ReactNode
  children: React.ReactNode
}) {
  const { user } = useAuth()
  return user?.role === 'admin' ? children : fallback
}

// 使用方式
<AdminGuard>
  <Button onClick={handleCreate}>Create Group</Button>
</AdminGuard>
```

**后续迭代**：后端引入 RBAC 权限码体系后，可将 `AdminGuard` 升级为 `PermissionGuard`，支持 `security:read` / `security:write` 细粒度控制。

---

## 五、数据层设计

### 5.1 技术选型

使用 **TanStack Query (React Query)** 作为数据获取层：
- 缓存管理（Dashboard 数据缓存 5 分钟，Logs 不缓存）
- 错误重试（失败时自动重试 2 次）
- 乐观更新（启用/停用开关）
- 轮询控制（Logs 实时刷新）

### 5.2 Query Key 设计规范

```typescript
// features/security/api/query-keys.ts
export const securityKeys = {
  dashboard: (range: TimeRange) => ['security', 'dashboard', range] as const,
  groups: (params: GroupQueryParams) => ['security', 'groups', params] as const,
  groupTree: () => ['security', 'groups', 'tree'] as const,
  rules: (params: RuleQueryParams) => ['security', 'rules', params] as const,
  ruleTest: (ruleId: string) => ['security', 'rules', ruleId, 'test'] as const, // Phase 2: 后端暂无测试 API
  policies: (params: PolicyQueryParams) => ['security', 'policies', params] as const,
  logs: (params: LogQueryParams) => ['security', 'logs', params] as const,
  logDetail: (id: string) => ['security', 'logs', id] as const,
}
```

### 5.3 API Hook 设计

```typescript
// features/security/api/use-security-groups.ts
export function useSecurityGroups(params: GroupQueryParams) {
  return useQuery({
    queryKey: securityKeys.groups(params),
    queryFn: () => fetchSecurityGroups(params),
    staleTime: 30_000, // 30 秒内不重复请求
  })
}

// features/security/api/use-toggle-rule.ts
export function useToggleRule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      updateRuleStatus(id, enabled),
    // 乐观更新
    onMutate: async ({ id, enabled }) => {
      await queryClient.cancelQueries({ queryKey: ['security', 'rules'] })
      const previousRules = queryClient.getQueryData<RulesResponse>(
        securityKeys.rules({ page: 1, pageSize: 20 })
      )
      queryClient.setQueryData<RulesResponse>(
        securityKeys.rules({ page: 1, pageSize: 20 }),
        (old) => {
          if (!old) return old
          return {
            ...old,
            data: old.data.map((rule) =>
              rule.id === id ? { ...rule, enabled } : rule
            ),
          }
        }
      )
      return { previousRules }
    },
    onError: (err, variables, context) => {
      if (context?.previousRules) {
        queryClient.setQueryData(
          securityKeys.rules({ page: 1, pageSize: 20 }),
          context.previousRules
        )
      }
      toast.error(t('Update failed'))
    },
    onSettled: () => {
      // 乐观更新后刷新规则列表和看板，确保数据最终一致
      queryClient.invalidateQueries({ queryKey: ['security', 'rules'] })
      queryClient.invalidateQueries({ queryKey: ['security', 'dashboard'] })
    },
  })
}
```

### 5.4 实时轮询策略（Logs 页面）

```typescript
// features/security/api/use-security-logs.ts
export function useSecurityLogs(params: LogQueryParams, options: { enabled: boolean }) {
  return useQuery({
    queryKey: securityKeys.logs(params),
    queryFn: () => fetchSecurityLogs(params),
    refetchInterval: options.enabled ? 5000 : false,
    refetchIntervalInBackground: false, // 后台标签页暂停轮询
    placeholderData: (previousData) => previousData, // 保持旧数据直到新数据到达
  })
}
```

---

## 六、页面详细设计

### 6.1 Dashboard（统计看板）

**现状**：已实现基础卡片 + 图表，基本可用。

**核心指标定义**：

| 指标 | 说明 | 数据来源 |
|------|------|----------|
| 总检测数 | 时间范围内经过安全检测的总次数 | `/api/security/dashboard?start_time=&end_time=` |
| 总拦截数 | 时间范围内被 Block 的总次数 | `/api/security/dashboard?start_time=&end_time=` |
| 总告警数 | 时间范围内被 Alert 的总次数 | `/api/security/dashboard?start_time=&end_time=` |
| 今日检测数 | 当日的总检测次数 | `/api/security/dashboard`（后端默认今日） |
| 拦截率 | 总拦截数 / 总检测数 × 100% | 前端计算 |
| 风险分布 | 各风险等级（Low/Medium/High/Critical）占比 | 饼图/环形图 |
| 分类 TOP10 | 命中次数最多的分组分类排行 | 横向柱状图 |
| 用户 TOP10 | 命中次数最多的用户排行 | 横向柱状图 |
| 模型 TOP10 | 命中次数最多的模型排行 | 横向柱状图 |
| ~~安全趋势~~ | ~~近 7/30 天拦截/放行趋势~~ | ~~折线图（后端暂无时间序列数据，待后续补充）~~ |

**优化点**：
1. 时间范围筛选同步到 URL：`?range=today|week|month|custom`
   - **注意**：后端 `GET /api/security/dashboard` 接收 `start_time`/`end_time`（Unix 时间戳），前端需将 `range` 转换为对应的时间戳
2. 自定义时间范围时显示日期选择器
3. 增加刷新按钮（手动刷新 + 自动刷新开关，间隔 30 秒）
4. 图表使用骨架屏加载，空状态显示引导文案
5. 数字卡片增加环比变化指示（↑ ↓ —）
6. **图表库选型**：使用 **Recharts**（如果项目已引入）或轻量级的 **Chart.js + react-chartjs-2**。优先复用项目中已有的图表库，避免重复引入。Dashboard 图表组件需做 Code Splitting（`React.lazy`），避免影响其他 Security 页面首屏加载。

**指标与后端数据对照**：

| 前端展示指标 | 后端返回字段 | 说明 |
|-------------|-------------|------|
| 总检测数 | `summary.total_detections` | 时间范围内的总检测次数 |
| 总拦截数 | `summary.total_interceptions` | 时间范围内被 Block 的次数 |
| 总告警数 | `summary.total_alerts` | 时间范围内被 Alert 的次数 |
| 今日检测数 | `summary.today_detections` | 今日的总检测次数 |
| 今日拦截数 | ⚠️ 后端暂未返回 | 前端用 `start_time=todayStart` 单独请求 Dashboard 计算，或等待后端补充 `today_interceptions` |
| 风险分布 | `risk_distribution` | 低/中/高/严重 分布数据 |
| 分类 TOP10 | `top_categories` | 命中次数最多的分组分类 |
| 用户 TOP10 | `top_users` | 命中次数最多的用户 |
| 模型 TOP10 | `top_models` | 命中次数最多的模型 |

**全局安全总开关设计**：

Dashboard 页面顶部增加全局安全检测开关：
- **开关状态来源**：调用 `GET /api/security/status` 获取 `enabled` 字段
- **开关控制逻辑**：
  - 该开关联动控制旧系统（`System Settings → Sensitive Words`）和新模块（`features/security`）的检测行为
  - 后端 `middleware/security.go` 已统一：`SecurityCheck()` 和 `SecurityCheckResponse()` 同时检查 `SECURITY_ENABLED` 环境变量 **和** `setting.CheckSensitiveEnabled`
  - **任一开关关闭，检测即停止**
- **UI 展示**：开关旁显示状态文字「安全检测已启用 / 已停用」
- **操作权限**：仅 ADMIN 可操作

> 为什么需要统一开关？旧系统的 `CheckSensitiveEnabled`（System Settings → Sensitive Words 页面）和新模块的 `SECURITY_ENABLED`（环境变量）曾是两套独立开关，管理员在旧页面关闭后新模块仍在运行，造成困惑。通过后端中间件联动检查，实现「一处关闭，全局生效」。

**⚠️ 与后端文档的差异**：
- 后端文档提到「命中规则 TOP5」，但后端实际返回的是「分类 TOP10」「用户 TOP10」「模型 TOP10」，无「规则 TOP5」数据
- 后端暂无「近 7/30 天安全趋势」时间序列数据，折线图需后续后端补充或前端移除
- 「拦截率」由前端自行计算：`total_interceptions / total_detections × 100%`

### 6.2 Groups（敏感词分组管理）

**现状**：已实现表格 + 增删改 Modal。

**分阶段实现方案**：

**第一阶段（当前迭代）**：平铺表格 + 父分组信息
- 表格增加「父分组」列，展示层级关系
- 按父分组筛选（下拉选择），同步到 URL `?parentId=`
- ~~搜索过滤：按名称搜索，同步到 URL `?search=`~~ ⚠️ 后端暂不支持 `name` 模糊搜索，待后端增强后启用
- 状态筛选：全部 / 启用 / 停用，同步到 URL `?status=`
- 复制功能：后端已实现 `POST /api/security/groups/:id/copy`，前端需增加「复制」按钮；复制后后端自动在名称后加 `_copy`

**第二阶段（后续优化）**：树形视图
- 当分组层级较深（>2 层）时，提供树形表格或层级缩进视图切换
- 树形视图支持展开/折叠，但不支持批量操作（批量操作保留在平铺视图）

```
第一阶段平铺表格：
名称          │ 父分组        │ 描述          │ 状态   │ 操作
─────────────────────────────────────────────────────────────────
个人隐私信息    │ 基础安全策略    │ 保护个人隐私    │ 启用   │ Edit Delete Copy
手机号检测      │ 个人隐私信息    │ 手机号正则     │ 启用   │ Edit Delete Copy
身份证号检测    │ 个人隐私信息    │ 身份证正则     │ 启用   │ Edit Delete Copy
企业机密        │ 基础安全策略    │ 保护企业核心   │ 启用   │ Edit Delete Copy
```

**分组详情抽屉**：点击分组名称可展开查看该分组下的规则数量、命中次数。

**数据迁移提示**：如果后端返回存在已迁移的旧系统数据，在页面顶部显示 Info Banner：
> 「系统迁移」分组包含从旧版敏感词系统迁移的规则，请检查规则配置是否符合预期。[查看详情]

### 6.3 Rules（检测规则管理）

**现状**：已实现表格 + 增删改 Modal。

**优化点**：
1. **按分组筛选**：下拉选择分组，只显示该分组下的规则，同步到 URL `?groupId=`
2. **规则类型筛选**：关键词 / 正则 / NER / AI，同步到 URL `?type=`
3. **状态筛选**：启用 / 停用，同步到 URL `?status=`
4. **批量操作**：批量启用 / 停用 / 删除（危险操作需二次确认）。**批量操作需要 ADMIN 角色。**
   - ⚠️ 后端暂无批量操作 API，前端需循环调用单条 API，或等待后端补充 `POST /api/security/rules/batch-delete` 和 `PATCH /api/security/rules/batch-status`
5. **规则测试**：~~增加「测试」按钮，点击后调用 `POST /api/security/rules/:id/test`~~ ⚠️ 后端暂无单条规则测试 API，此功能列为**后续迭代（Phase 2）**
   - **临时替代方案**：可引导用户使用 `POST /api/security/check/request` 进行通用检测（需传入 user_id 和 content，检测的是全量规则而非单条规则）
6. **内容预览**：规则内容过长时截断显示，hover tooltip 显示完整内容
7. **分页**：默认 20 条/页，选项 20/50/100

**规则状态切换说明**：
- 规则启用/停用需要调用 `PUT /api/security/rules/:id` 传入完整规则数据
- ⚠️ 后端暂无 `PATCH /api/security/rules/:id/status` 单字段更新 API，乐观更新时需注意并发覆盖风险
- 建议后端优先补充 `PATCH /api/security/rules/:id/status`（见《后端API补充清单.md》P1-2）

### 6.4 Policies（用户策略管理）

**现状**：已实现表格 + 增删改 Modal。

**优化点**：
1. **用户选择器**：当前 `user_id` 是数字输入框，应改为用户搜索选择器（支持按用户名搜索），使用 debounce 搜索（300ms）
2. **策略生效范围可视化**：用 Tag/Chip 显示 Request / Response / Both
3. **白名单 IP 输入优化**：支持输入多个 IP，用 Tag 形式展示，支持粘贴逗号/换行分隔的多个 IP
4. ~~**策略优先级**：支持数字优先级调整（1-100），数字越小优先级越高；表格按优先级排序~~ ⚠️ 后端 `security_user_policies` 表暂无 `priority` 字段，此功能列为**后续迭代（Phase 2）**
5. **权限控制**：创建/编辑/删除策略需要 **ADMIN 角色**

**⚠️ 后端限制**：
- `whitelist_ips` 后端存储为 `TEXT` 类型，前端提交 JSON 字符串或逗号分隔字符串均可，需与后端确认存储格式
- 策略表缺少 `priority` 字段，暂不支持优先级排序（默认按 `id DESC` 排序）

### 6.5 Audit Logs（审计日志）

**现状**：已实现表格 + 导出 + 详情 Drawer。

**优化点**：
1. **高级筛选栏**（全部同步到 URL）：
   - ~~时间范围选择器（默认今日）~~ ⚠️ 后端 `GET /api/security/logs` 暂无 `start_time`/`end_time` 参数支持，需后端增强（见《后端API补充清单.md》P0-1）
   - 用户下拉选择（支持搜索）
   - ~~模型名称搜索~~ ⚠️ 后端暂无 `model_name` 筛选参数，需后端增强
   - 处理动作单选（Pass/Alert/Mask/Block/Review）⚠️ 后端只支持单值 `action`，不支持多选
   - 风险等级单选（Low/Medium/High/Critical）⚠️ 后端只支持单值 `risk_level`，不支持多选
   - 内容类型（Request/Response）
2. **实时刷新**：开关控制是否自动刷新（默认关闭，开启后 5 秒间隔），后台标签页自动暂停
3. **数据更新策略（避免自动替换打断用户浏览）**：
   - 轮询获取新数据时，**不自动替换当前表格**
   - 如果检测到新日志，在表格顶部显示提示条：「有 N 条新日志，点击刷新查看」
   - 用户点击后才刷新表格，避免打断正在查看的日志详情
   - 最大保留条数：单页 100 条，超出时提示用户缩小筛选范围
4. **详情优化**：Drawer 中显示：
   - 请求摘要（模型、Token 数、时间）
   - 命中规则详情（规则名、风险等级、命中片段）
   - 处理前后对比（如果发生 Mask/替换）
   - 原始内容哈希值（用于校验）
5. **批量导出**：支持按当前筛选条件导出 CSV/Excel
   - ⚠️ 后端 `format=excel` 实际返回 CSV 数据 + Excel MIME 类型，非真正 `.xlsx`。前端按钮文案保持「导出 CSV/Excel」，或等待后端引入真正的 Excel 库
6. **敏感信息脱敏**：日志详情中的命中片段（手机号、身份证号等）展示为 `138****8000` 格式

---

## 七、新旧系统整合策略

### 7.1 问题分析

旧系统（`system-settings/security/sensitive-words`）与新模块的关系：

| 维度 | 旧系统 | 新模块 |
|------|--------|--------|
| 存储 | `SensitiveWords` 选项（文本） | `security_rules` 表（结构化） |
| 匹配 | 简单字符串包含 | AC 自动机 + 正则 + NER + AI |
| 分组 | 无 | 树形分组 |
| 策略 | 无 | 用户级策略绑定 |
| 日志 | 无 | 完整审计日志 |
| 动作 | 仅拦截 | Pass/Alert/Mask/Block/Review |

### 7.2 整合方案

**方案：废弃旧系统，由新模块完全替代**

理由：
1. 新模块已完全覆盖旧系统功能（关键词检测 + 拦截）
2. 两套系统并存会导致数据不一致和维护困难
3. 旧系统的 `SensitiveWords` 可以迁移到 `security_rules` 表中

**前端职责**：

1. **前端移除旧入口**：
   - 从 `system-settings/security/section-registry.tsx` 中移除 `sensitive-words` section
   - 保留 `rate-limit` 和 `ssrf`

2. **迁移状态提示（待后端提供 API 后实现）**：
   - ⚠️ 后端暂无 `GET /api/security/migration-status` 接口，此功能列为**后续迭代**
   - 待后端提供接口后，前端在 Groups 页面挂载时调用，展示 Info Banner：「系统迁移」分组包含从旧版敏感词系统迁移的规则，请检查规则配置是否符合预期
   - Banner 可手动关闭，关闭状态存入 `localStorage`

**后端职责（不在本文档范围内，仅列出品约）**：

1. **数据迁移**：
   - 在系统升级/初始化时，如果 `SensitiveWords` 不为空，将其内容逐行转换为 `security_rules` 记录
   - 目标分组：创建一个名为「系统迁移」的分组
   - 规则类型：关键词（type=1）
   - 动作：拦截（action=4）
   - **迁移前备份**：将原始 `SensitiveWords` 内容写入备份字段/表，便于回滚
   - 处理重复内容和空行（去重、过滤空行）

2. **统一开关（已实施后端正向兼容）**：
   - `middleware/security.go` 已修改：`SecurityCheck()` 和 `SecurityCheckResponse()` 同时检查 `SECURITY_ENABLED` 环境变量 **和** `setting.CheckSensitiveEnabled`
   - **任一开关关闭，检测即停止**，实现新旧系统开关联动
   - Dashboard 页面提供全局开关 UI，调用 `GET /api/security/status` 查看当前状态
   - 旧系统 `System Settings → Sensitive Words` 页面的 `CheckSensitiveEnabled` 开关仍然有效，关闭后新模块也同步停止检测

3. **回滚方案**：
   - 如果新模块出现问题，可通过备份恢复旧系统配置
   - 紧急关闭：通过 Dashboard 总开关或旧系统页面关闭 `CheckSensitiveEnabled`

4. **后端清理（后续）**：
   - 保留 `CheckSensitiveEnabled` 作为全局安全模块总开关（已与新模块联动）
   - 废弃 `SensitiveWords` 和 `CheckSensitiveOnPromptEnabled` 选项

---

## 八、组件设计

### 8.1 新增组件清单

| 组件 | 位置 | 说明 |
|------|------|------|
| `SecurityTabs` | `features/security/components/security-tabs.tsx` | 统一子导航 Tab（Base UI），点击触发路由跳转 |
| `SecurityPageLayout` | `features/security/components/security-page-layout.tsx` | 统一页面布局（标题 + Tab + 内容区） |
| `GroupTreeTable` | `features/security/components/group-tree-table.tsx` | 树形分组表格（第二阶段实现） |
| `RuleTester` | `features/security/components/rule-tester.tsx` | 规则测试器（**Phase 2**，后端暂无单条规则测试 API） |
| `UserSelector` | `features/security/components/user-selector.tsx` | 用户搜索选择器（debounce 300ms） |
| `LogFilters` | `features/security/components/log-filters.tsx` | 日志高级筛选（URL 同步） |
| `GroupCopyButton` | `features/security/components/group-copy-button.tsx` | 分组复制按钮（复制后后端自动加 `_copy`） |
| `AdminGuard` | `features/security/components/admin-guard.tsx` | 管理员角色守卫组件，控制功能显隐（当前后端使用 AdminAuth） |
| `EmptyState` | `components/empty-state.tsx` | 统一空状态（首次使用/无数据/筛选为空/无权限） |
| `ConfirmDialog` | `components/confirm-dialog.tsx` | 统一确认对话框 |
| `DataTable` | `components/data-table.tsx` | 通用数据表格（基于 `@tanstack/react-table`，支持分页、排序、选择） |

**DataTable 技术选型说明**：
- 基于 `@tanstack/react-table`（headless UI）封装，不重复造轮子
- 与 TanStack Query 配合是社区标准实践
- 项目中如果已有表格组件，优先评估复用可能性，再决定是否新建

### 8.2 SecurityPageLayout 设计

```tsx
interface SecurityPageLayoutProps {
  title: string
  description?: string
  actions?: React.ReactNode
  children: React.ReactNode
}

// 使用方式：
<SecurityPageLayout
  title="Sensitive Word Groups"
  description="Manage detection rule groups"
  actions={
    <AdminGuard>
      <Button>Create Group</Button>
    </AdminGuard>
  }
>
  <GroupTable ... />
</SecurityPageLayout>
```

渲染结构：
```
┌──────────────────────────────────────────────────┐
│  AI 内容安全                           [Create]  │
│  Manage detection rule groups                    │
├──────────────────────────────────────────────────┤
│  [Dashboard] [Groups] [Rules] [Policies] [Logs]  │
├──────────────────────────────────────────────────┤
│                                                  │
│  {children}                                      │
│                                                  │
└──────────────────────────────────────────────────┘
```

### 8.3 EmptyState 设计

```tsx
type EmptyStateType = 'initial' | 'empty' | 'filtered' | 'forbidden'

interface EmptyStateProps {
  type: EmptyStateType
  title?: string
  description?: string
  action?: React.ReactNode
}
```

| 类型 | 场景 | 默认文案 |
|------|------|----------|
| `initial` | 首次使用，无数据 | "暂无分组，创建第一个检测分组开始保护内容安全" |
| `empty` | 数据为空 | "暂无数据" |
| `filtered` | 筛选无结果 | "未找到匹配结果，请调整筛选条件" |
| `forbidden` | 无权限 | "您没有权限查看此页面" |

### 8.4 表单设计规范

**技术选型**：使用 **React Hook Form** + **Zod** 进行表单状态管理和校验。

**原因**：
- React Hook Form 性能优秀（非受控组件），适合管理后台频繁的 Modal 表单
- Zod 与 TypeScript 集成良好，与 URL Filter 的校验层技术栈统一

**示例：Group 创建表单**

```typescript
// features/security/components/group-form-modal.tsx
const groupSchema = z.object({
  name: z.string().min(1, t('Name is required')).max(100),
  description: z.string().max(500).optional(),
  parentId: z.string().optional(),
  enabled: z.boolean().default(true),
})

type GroupFormData = z.infer<typeof groupSchema>

export function GroupFormModal({ open, onClose, initialData }: GroupFormModalProps) {
  const form = useForm<GroupFormData>({
    resolver: zodResolver(groupSchema),
    defaultValues: initialData ?? { enabled: true },
  })

  const mutation = useCreateGroup()

  const onSubmit = (data: GroupFormData) => {
    mutation.mutate(data, {
      onSuccess: () => {
        toast.success(t('Create success'))
        onClose()
      },
      onError: (error) => {
        toast.error(error.message || t('Create failed'))
      },
    })
  }

  return (
    <Dialog open={open} onClose={onClose}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <Input {...form.register('name')} error={form.formState.errors.name?.message} />
        <Textarea {...form.register('description')} />
        <Switch {...form.register('enabled')} />
        <Button type="submit" loading={mutation.isPending}>
          {t('Submit')}
        </Button>
      </form>
    </Dialog>
  )
}
```

**表单规范**：
1. 所有表单字段必须有 `label` 和 `placeholder`
2. 提交按钮在请求中显示 `loading` 状态并禁用表单
3. 服务端校验错误（如名称重复）回显到对应字段
4. 创建成功后重置表单，编辑成功后保持当前值
5. 关闭 Modal 时如果有未保存变更，提示确认

---

## 九、安全设计

### 9.1 XSS 防护

所有展示用户输入内容的区域必须使用纯文本渲染：
- 审计日志中的 Prompt/Response 内容使用 `textContent` 或 HTML escape 后渲染
- **禁止**使用 `dangerouslySetInnerHTML` 渲染日志内容
- 规则内容、分组描述等管理员输入字段同样纯文本渲染

### 9.2 敏感信息脱敏

审计日志详情中，命中规则的高亮片段自动脱敏：
- 手机号：`13800138000` → `138****8000`
- 身份证号：`11010119900101xxxx` → `110101********xxxx`
- 银行卡号：保留后 4 位
- 邮箱：`user@example.com` → `u***@example.com`

脱敏逻辑在前端渲染层实现，后端返回原始命中片段用于精确匹配展示。

### 9.3 操作审计

管理操作（创建/编辑/删除分组、规则、策略）需要记录操作日志：
- 操作人、操作时间、操作类型、操作对象、变更前后对比
- 操作日志存储于系统操作日志表（复用现有操作日志机制）
- 前端在关键操作成功后发送操作审计请求

### 9.4 错误处理与降级策略

**API 错误分类处理**：

| 错误类型 | HTTP 状态 | 前端处理 |
|----------|-----------|----------|
| 网络错误 | 无 / `Network Error` | Toast 提示「网络异常，请检查连接」，自动重试 2 次 |
| 服务器错误 | 500 | Toast 提示「服务器繁忙，请稍后重试」 |
| 权限变更 | 403 | 刷新页面，若仍无权限则显示 `EmptyState type="forbidden"` |
| 数据不存在 | 404 | 如果是编辑/删除操作，Toast 提示「数据已被删除或不存在」并刷新列表 |
| 业务校验 | 400 / 422 | 将错误信息映射到表单字段或 Toast 展示 |

**Error Boundary**：
- Security 模块内包裹 `<SecurityErrorBoundary>`，捕获渲染错误
- 错误边界内显示友好的错误页面（「页面出错了，点击刷新」），而非白屏
- 错误信息上报到监控系统（如 Sentry，如果项目已接入）

**TanStack Query 全局配置**：
```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1, // 仅网络错误时重试
    },
  },
})
```

---

## 十、交互与反馈设计

### 10.1 全局反馈规范

| 场景 | 反馈方式 | 持续时间 |
|------|----------|----------|
| 操作成功 | Toast（绿色） | 3 秒 |
| 操作失败 | Toast（红色）+ 错误详情 | 5 秒 |
| 表单校验错误 | 字段级红色提示 + 滚动到首个错误 | 持久 |
| 加载中 | 骨架屏 / 按钮 Loading 状态 | - |
| 危险操作 | ConfirmDialog 二次确认 | 需手动确认 |

### 10.2 加载状态规范

| 场景 | 加载方式 |
|------|----------|
| 页面首次加载 | 骨架屏（Skeleton） |
| 表格数据刷新 | 表格内 Loading 遮罩，保留表头 |
| 操作提交 | 按钮 Loading 状态，禁用表单 |
| 下拉选择数据 | 下拉框内 Spinner |

### 10.3 危险操作确认

以下操作必须弹出确认对话框：
- 删除分组（提示影响子分组和规则数量）
- 删除规则（提示该规则在策略中的引用）
- 批量删除规则
- 删除策略

```
┌─────────────────────────────────────────┐
│  ⚠️  确认删除                           │
│                                         │
│  确定要删除分组「个人隐私信息」吗？       │
│  该分组下有 5 条规则，将一并删除。       │
│  此操作不可撤销。                        │
│                                         │
│  [取消]              [确认删除]          │
└─────────────────────────────────────────┘
```

---

## 十一、国际化（i18n）

新增翻译键值（以 `en.json` 为基准）：

```json
{
  "AI Content Security": "AI Content Security",
  "Security Dashboard": "Security Dashboard",
  "Sensitive Word Groups": "Sensitive Word Groups",
  "Detection Rules": "Detection Rules",
  "Security Policies": "Security Policies",
  "Audit Logs": "Audit Logs",
  "Create Group": "Create Group",
  "Create Rule": "Create Rule",
  "Create Policy": "Create Policy",
  "Edit Group": "Edit Group",
  "Edit Rule": "Edit Rule",
  "Edit Policy": "Edit Policy",
  "Copy Group": "Copy Group",
  "Test Rule": "Test Rule",
  "Parent Group": "Parent Group",
  "Risk Score": "Risk Score",
  "Content Type": "Content Type",
  "Match Detail": "Match Detail",
  "Export CSV": "Export CSV",
  "Export Excel": "Export Excel",
  "No data": "No data",
  "No matching results": "No matching results. Try adjusting your filters.",
  "Loading": "Loading...",
  "Refresh": "Refresh",
  "Auto refresh": "Auto refresh",
  "Real-time": "Real-time",
  "Confirm Delete": "Confirm Delete",
  "Confirm delete description": "Are you sure you want to delete {{name}}? This action cannot be undone.",
  "Delete success": "Deleted successfully",
  "Create success": "Created successfully",
  "Update success": "Updated successfully",
  "Copy success": "Copied successfully",
  "Test success": "Test completed",
  "Match found": "Match found: {{content}}",
  "No match": "No match",
  "Rule syntax error": "Rule syntax error",
  "Today": "Today",
  "This week": "This week",
  "This month": "This month",
  "Custom range": "Custom range",
  "Total requests": "Total requests",
  "Blocked requests": "Blocked requests",
  "Block rate": "Block rate",
  "Risk distribution": "Risk distribution",
  "Top rules": "Top triggered rules",
  "Security trend": "Security trend",
  "Sensitive content masked": "Sensitive content has been masked for display",
  "No permission": "You don't have permission to access this page",
  "Migrated from old system": "Migrated from old system",
  "Migrated group tip": "This group contains rules migrated from the legacy sensitive words system.",
  "Update failed": "Update failed",
  "Create failed": "Create failed",
  "Name is required": "Name is required",
  "New logs available": "{{count}} new logs available. Click to refresh.",
  "New logs available_plural": "{{count}} new logs available. Click to refresh.",
  "Rule count": "{{count}} rule",
  "Rule count_plural": "{{count}} rules"
}
```

**插值与复数**：
- 使用 i18next 的 `t('key', { name: 'xxx' })` 进行插值
- 复数形式使用 `count` 参数，i18next 自动选择 `_zero`、`_one`、`_other` 后缀（根据语言配置）
- 示例：`t('Rule count', { count: 5 })` → "5 rules"

---

## 十二、实施计划

### 依赖关系说明

```
T7/T8 (EmptyState/ConfirmDialog)
  ↓
T2/T3 (SecurityTabs/SecurityPageLayout)
  ↓
T4 (页面套用布局)
  ↓
T5 (Query Key 规范)
  ↓
T9 (TanStack Query Hooks) → T10-T14/T16-T20 (各页面增强)
  ↓
T15 (DataTable) → T10/T12/T16/T18 (表格依赖)
  ↓
T22 (i18n)
  ↓
T23 (测试)
```

### Milestone 1：导航框架 + 全局组件 + 旧系统移除（2 天）

| 任务 | 文件 | 说明 | 依赖 |
|------|------|------|------|
| T0 | `components/data-table.tsx` | 调研并评估项目中是否已有表格组件。若无，基于 `@tanstack/react-table` 封装通用 DataTable | 无 |
| T7 | `components/empty-state.tsx` | 新建统一空状态组件 | 无 |
| T8 | `components/confirm-dialog.tsx` | 新建统一确认对话框 | 无 |
| T1 | `useSidebarData.ts` | Security 保持单入口，`activeUrls` 匹配 `/security/*` | 无 |
| T2 | `features/security/components/security-tabs.tsx` | 新建子导航 Tab 组件（Base UI），点击触发 `navigate()` | T7 |
| T3 | `features/security/components/security-page-layout.tsx` | 新建统一页面布局 | T2 |
| T4 | `features/security/pages/*-page.tsx` | 所有页面套用 SecurityPageLayout | T3 |
| T5 | `features/security/api/query-keys.ts` | 建立 Query Key 规范 | 无 |
| T6 | `system-settings/security/section-registry.tsx` | 移除 sensitive-words section | 无 |
| T8b | `features/security/components/admin-guard.tsx` | 新建管理员角色守卫组件（适配后端 AdminAuth） | 无 |

### Milestone 2：数据层 + URL 同步 + Groups/Rules 增强（3 天）

| 任务 | 文件 | 说明 | 依赖 |
|------|------|------|------|
| T9 | `features/security/api/*.ts` | 建立 TanStack Query Hooks（含乐观更新） | T5 |
| T9b | `features/security/hooks/use-url-filters.ts` | 封装 URL 同步 Hook（Groups/Rules/Logs/Policies 各一个） | T9 |
| T10 | `group-page.tsx` | 集成平铺表格、父分组列、URL 筛选同步、复制按钮 | T0, T4, T9b |
| T11 | `group-page.tsx` | ~~增加迁移状态 Banner~~ ⚠️ 后端暂无 `migration-status` API，此任务移至 Phase 2 | T10 |
| T12 | `rule-page.tsx` | 增加分组/类型/状态筛选、URL 同步 | T0, T4, T9b |
| T13 | `rule-page.tsx` | ~~集成 RuleTester 组件~~ ⚠️ 后端暂无规则测试 API，此任务移至 Phase 2 | T12 |
| T14 | `rule-page.tsx` | 批量操作（循环调用单条 API）+ 二次确认 | T12 |

### Milestone 3：Policies/Logs 增强 + 国际化 + 测试（3 天）

| 任务 | 文件 | 说明 | 依赖 |
|------|------|------|------|
| T16 | `policy-page.tsx` | 用户选择器替换数字输入（debounce 搜索） | T0, T4, T9b |
| T17 | `policy-form-modal.tsx` | 白名单 IP Tag 输入 | T16 |
| T17b | `policy-form-modal.tsx` | ~~优先级调整~~ ⚠️ 后端暂无 `priority` 字段，移至 Phase 2 | T16 |
| T18 | `log-page.tsx` | 高级筛选栏（URL 同步所有条件） | T0, T4, T9b |
| T19 | `log-page.tsx` | 实时刷新开关 + 新数据提示条（后台暂停轮询） | T18 |
| T20 | `log-detail-drawer.tsx` | 详情信息丰富化、脱敏展示 | T18 |
| T22 | `i18n/locales/*.json` | 补充所有新增翻译键值 | T4, T10-T20 |
| T23 | `__tests__/security/` | 单元测试（LogFilters、RuleTester、useUrlFilters） | T9b, T13, T18 |
| T24 | E2E Tests | 导航切换、创建规则、日志筛选、权限测试 | T22 |

---

## 十三、验证清单

### 13.1 功能验证

- [ ] 侧边栏 Security 单入口高亮，所有 `/security/*` 页面均高亮
- [ ] 每个 `/security/*` 页面顶部有 Tab 导航，当前页高亮
- [ ] Tab 点击切换页面无整页刷新（`navigate()` 行为）
- [ ] Groups 页面筛选条件同步到 URL，刷新后保留
- [ ] ~~Rules 页面支持规则测试（调用 API）~~ Phase 2 实现（后端暂无测试 API）
- [ ] Rules 页面批量删除有二次确认
- [ ] Policies 页面用户选择器可搜索（debounce）
- [ ] Logs 页面基础筛选同步到 URL（user_id/action/risk_level/content_type）
- [ ] ~~Logs 页面时间范围/模型名称筛选~~ Phase 2 实现（后端暂无对应参数）
- [ ] Logs 页面实时刷新开关工作正常，后台标签页暂停轮询，新日志以提示条形式通知
- [ ] System Settings 中不再显示旧 Sensitive Words
- [ ] 所有页面仅 **ADMIN 角色**用户可见（当前后端使用 AdminAuth）
- [ ] 管理操作（创建/编辑/删除）仅 **ADMIN 角色**可操作
- [ ] 管理操作后显示 Toast 反馈
- [ ] 表单提交时按钮显示 loading 状态，禁用重复提交

### 13.2 性能验证

- [ ] Dashboard 数据缓存有效，5 分钟内重复访问不发请求
- [ ] Dashboard 图表组件已做 Code Splitting，不影响其他页面首屏
- [ ] 表格分页正常（默认 20/页，可选 50/100）
- [ ] Logs 轮询在后台标签页自动暂停
- [ ] 页面切换无内存泄漏（检查 setInterval 清理）
- [ ] DataTable 基于 `@tanstack/react-table`，虚拟滚动正常工作（大数据量下）

### 13.3 安全验证

- [ ] 日志内容纯文本渲染，无 XSS 漏洞
- [ ] 日志详情中手机号、身份证号等自动脱敏
- [ ] 删除操作必须二次确认
- [ ] 路由守卫基于 ADMIN 角色（当前后端使用 AdminAuth）
- [ ] 非 ADMIN 用户访问 `/security/*` 显示无权限提示
- [ ] 权限不足时操作按钮隐藏（而非仅禁用）

### 13.4 国际化验证

- [ ] 所有新增文案有 zh/en 翻译
- [ ] 切换语言后所有页面正常显示
- [ ] 插值翻译正常工作（如 `{{name}}`、`{{count}}`）
- [ ] 复数形式正确（1 rule / 5 rules）

### 13.5 E2E 测试场景

- [ ] **导航测试**：从 Dashboard 依次点击 Tab 切换 Groups → Rules → Policies → Logs → Dashboard
- [ ] **URL 同步测试**：在 Groups 页面设置筛选条件，刷新页面后筛选保留
- [ ] **创建规则**：在 Rules 页面创建关键词规则，验证列表刷新
- [ ] **规则测试**：对规则进行测试，验证命中/未命中/错误结果展示
- [ ] **日志筛选**：在 Logs 页面选择时间范围 + 风险等级，验证筛选结果
- [ ] **全局开关测试**：Dashboard 总开关关闭后，`SecurityCheck()` 中间件停止检测；旧系统 `CheckSensitiveEnabled` 关闭后，新模块同步停止检测
- [ ] **权限测试**：非 ADMIN 用户访问 `/security` 被重定向或无权限提示
- [ ] **实时日志**：开启自动刷新，模拟新日志到达，验证提示条出现

---

## 十四、附录

### 14.1 相关文件清单

**前端核心文件**：
- `web/default/src/features/security/pages/dashboard-page.tsx`
- `web/default/src/features/security/pages/group-page.tsx`
- `web/default/src/features/security/pages/rule-page.tsx`
- `web/default/src/features/security/pages/policy-page.tsx`
- `web/default/src/features/security/pages/log-page.tsx`
- `web/default/src/features/security/api/security.ts`
- `web/default/src/hooks/use-sidebar-data.ts`
- `web/default/src/hooks/use-sidebar-config.ts`
- `web/default/src/features/system-settings/security/section-registry.tsx`

**后端核心文件**：
- `controller/security.go`
- `service/security/*.go`
- `middleware/security.go`
- `router/relay-router.go`

### 14.2 风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| 树形表格组件性能差 | Groups 页面卡顿 | 第一阶段不做树形，使用平铺表格；第二阶段若需树形，使用虚拟滚动或「平铺/树形」视图切换 |
| Logs 轮询导致服务器压力 | 后端负载增加 | 后台标签页自动暂停；最大轮询时长限制（30 分钟后自动关闭）；用户切走标签页时暂停 |
| 旧系统数据迁移失败 | 数据丢失 | 迁移完全由后端负责，保留备份；前端仅展示状态 Banner，不介入迁移逻辑 |
| 实时日志数据量大 | 前端内存泄漏 | 限制单页最大 100 条；轮询时以提示条通知新数据，不自动追加；用户手动刷新后才替换 |
| 图表库增加包体积 | Dashboard 加载慢 | 图表组件 Code Splitting（React.lazy）；优先复用项目已有图表库 |
| 前后端 API 不一致 | 前端功能无法运行 | 实施前对照《后端API补充清单.md》逐项确认；前端设计以当前后端实现为准，超出现有能力的特性列为 Phase 2 |

### 14.3 后续优化方向（非本次范围）

1. **操作审计页面**：独立页面查看管理员的操作记录
2. **规则模板市场**：提供预设规则模板（如金融、医疗、教育行业模板）
3. **告警通知**：规则命中率达到阈值时发送邮件/Webhook 通知
4. **多租户隔离**：不同用户组只能看到自己创建的分组和规则
5. **Groups 树形视图**：当分组层级超过 2 层时，提供树形表格浏览模式

### 14.4 现有代码评估

| 文件 | 当前状态 | 复用建议 |
|------|----------|----------|
| `features/security/pages/dashboard-page.tsx` | 基础卡片 + 图表已实现 | ✅ 可直接复用，套用 `SecurityPageLayout` |
| `features/security/pages/group-page.tsx` | 表格 + Modal 已实现 | 🔄 需重构：替换为 `DataTable`，集成 URL 筛选、复制按钮 |
| `features/security/pages/rule-page.tsx` | 表格 + Modal 已实现 | 🔄 需重构：增加筛选栏、批量操作；RuleTester 待后端 API 补充后集成 |
| `features/security/pages/policy-page.tsx` | 表格 + Modal 已实现 | 🔄 需重构：用户选择器、IP Tag 输入；优先级调整待后端字段补充后集成 |
| `features/security/pages/log-page.tsx` | 表格 + 导出 + Drawer 已实现 | 🔄 需重构：高级筛选栏、实时刷新提示条、脱敏展示 |
| `features/security/api/security.ts` | 原始 fetch 封装 | ⚠️ 评估：若未使用 TanStack Query，需新建 hooks 层并迁移调用点 |
| `hooks/use-sidebar-data.ts` | 已有 sidebar 配置 | ✅ 修改 `activeUrls` 即可 |

**说明**：
- `api/security.ts` 如果使用的是原始 `axios/fetch` 封装，保留作为底层 HTTP 客户端，在其之上新建 `features/security/api/*.ts` 的 TanStack Query Hooks 层。逐步迁移，而非一次性替换。
- 各页面的 Modal 表单如果使用的是手动 `useState` 管理，建议迁移到 React Hook Form + Zod。

### 14.5 性能预算

| 指标 | 目标值 | 验证方式 |
|------|--------|----------|
| Security 模块首屏加载（除 Dashboard） | < 500ms（本地开发） | Lighthouse / 浏览器 Performance Tab |
| Dashboard 图表包体积增量 | < 50KB gzip | Bundle Analyzer |
| Dashboard 图表首屏渲染 | < 1s（含数据加载） | 手动计时 |
| 表格分页切换 | < 200ms | 手动计时 |
| Logs 轮询请求 | 5s 间隔，单次响应 < 100ms | 浏览器 Network Tab |
| 内存占用（Logs 页面停留 30 分钟） | 不显著增长 | Chrome DevTools Memory Tab |

### 14.6 技术决策记录（ADR）

**ADR-1：Tab 导航 vs 侧边栏子菜单**
- **决策**：使用页面内 Tab 导航，侧边栏保持单入口
- **理由**：Security 内部 5 个页面属于同一管理模块，Tab 切换更符合用户心智模型；保持 URL 可分享；侧边栏层级保持扁平（项目 sidebar 已有较多入口）
- **替代方案**：侧边栏展开 5 个子菜单项（ rejected：增加 sidebar 层级，与其他模块风格不一致）

**ADR-2：URL 筛选同步方案**
- **决策**：使用 TanStack Router 的 `useSearch` + Zod 校验，封装 `useUrlFilters`
- **理由**：与项目路由方案一致；Zod 提供类型安全和默认值；`replace: true` 避免污染历史栈
- **替代方案**：自定义 `useState` + `useEffect` 同步 URL（rejected：易出错，无类型安全）

**ADR-3：树形表格分阶段实现**
- **决策**：第一阶段使用平铺表格 + 父分组列，第二阶段再引入树形视图
- **理由**：降低 Milestone 2 复杂度；Base UI 无现成 TreeTable，从零实现风险高；平铺表格已能满足 80% 使用场景
- **替代方案**：第一阶段直接实现树形表格（rejected：工期不可控，性能风险）

**ADR-4：数据迁移职责划分**
- **决策**：数据迁移完全由后端负责，前端仅展示迁移状态 Banner
- **理由**：前端设计方案不应涉及数据库字段设计；后端更了解数据一致性约束；前端通过 API 获取状态即可
- **替代方案**：前端负责触发迁移流程（rejected：引入不必要的前后端耦合）

**ADR-5：实时日志刷新策略**
- **决策**：轮询检测新数据，以提示条通知用户，不自动替换表格
- **理由**：自动替换会打断正在浏览/对比日志的用户；提示条模式让用户掌控刷新时机
- **替代方案**：自动追加新数据到表格顶部（rejected：内存无限增长，且会改变用户当前滚动位置）

**ADR-6：权限模型适配（v2.2 新增）**
- **决策**：前端使用 ADMIN 角色检查（`user.role === 'admin'`），而非 `security:read/write` 权限码
- **理由**：当前后端所有 Security API 均使用 `middleware.AdminAuth()`，基于角色而非权限码；后端 `controller/user.go` 的权限计算中无 `security:read/write` 定义；前端适配后端现有实现可减少后端改造成本，加速上线
- **替代方案**：后端引入完整 RBAC 权限码体系（rejected：改动面大，涉及所有 API 的权限中间件改造，不适合当前迭代）
- **后续迭代**：后端引入 `security:read` / `security:write` 权限码后，将 `AdminGuard` 升级为 `PermissionGuard`
