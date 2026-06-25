# Web Contracts: ai-security

## 页面路由

所有 ai-security 前端页面统一使用 `/ai-security/*` 路径。

| 路径 | 页面 | 说明 |
|---|---|---|
| `/ai-security` | Dashboard | 数据看板首页 |
| `/ai-security/dashboard` | Dashboard | 数据看板 |
| `/ai-security/groups` | GroupPage | 分组管理 |
| `/ai-security/rules` | RulePage | 规则管理 |
| `/ai-security/policies` | PolicyPage | 策略管理 |
| `/ai-security/logs` | LogPage | 命中日志 |

## 菜单入口

ai-security 不作为独立一级菜单，而是作为系统设置安全模块的高级功能入口：

```text
System Settings
  └── Security
       ├── Rate Limits
       ├── Sensitive Words
       └── AI Content Security   ← /ai-security
```

## 组件结构

前端代码全部位于 `custom/ai-security/web/`，建议结构：

```text
custom/ai-security/web/
├── api/
│   ├── query-keys.ts
│   └── ai-security.ts
├── components/
│   ├── ai-security-layout.tsx
│   ├── ai-security-tabs.tsx
│   ├── group-form-modal.tsx
│   ├── rule-form-modal.tsx
│   ├── rule-tester.tsx
│   ├── policy-form-modal.tsx
│   ├── log-detail-drawer.tsx
│   ├── dashboard-stat-card.tsx
│   ├── risk-distribution-chart.tsx
│   ├── top-users-table.tsx
│   └── top-models-table.tsx
├── hooks/
│   ├── use-log-filters.ts
│   ├── use-rule-filters.ts
│   ├── use-policy-filters.ts
│   └── use-url-filters.ts
├── pages/
│   ├── dashboard-page.tsx
│   ├── group-page.tsx
│   ├── rule-page.tsx
│   ├── policy-page.tsx
│   └── log-page.tsx
├── routes/
│   ├── index.tsx
│   ├── dashboard.tsx
│   ├── groups.tsx
│   ├── rules.tsx
│   ├── policies.tsx
│   └── logs.tsx
├── constants.ts
└── i18n/
    └── ai-security.json
```

## 主项目前端改动

主项目前端只允许两处改动：

1. **注册路由**：在路由配置中引入 `custom/ai-security/web/routes/*`。
2. **菜单入口**：在 `System Settings > Security` 的菜单/section 注册中增加 `/ai-security` 入口。

不允许将 ai-security 的页面、组件、状态管理、样式散落到官方前端目录。

## 状态管理

- 使用本地 React state 或轻量级 store 管理页面状态。
- 通过 TanStack Query（或项目现有数据获取方案）调用 `/api/ai-security/*` 接口。
- Query Key 统一使用 `['ai-security', ...]` 前缀，避免与官方缓存冲突。

## 权限控制

- ai-security 所有管理页面需要 Admin 权限。
- 未授权用户访问 `/ai-security/*` 时重定向到无权限页面或登录页。

## 国际化

- 初始支持 `zh` 和 `en`。
- 翻译文件位于 `custom/ai-security/web/i18n/ai-security.json`。
- 主项目 i18n 可选择性加载该文件，或模块内部自管理。

## 与官方 Sensitive Words 页面的关系

- 官方 `/system-settings/security/sensitive-words` 页面保持不变。
- 在该页面或同级的 Security Section 中增加"AI Content Security"入口，文案上体现"高级"或"增强"。
- 两个页面之间可互相跳转，但互不依赖内部实现。
