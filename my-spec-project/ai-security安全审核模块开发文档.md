#ai-security安全审核模块开发文档

## 项目目标

在 `new-api` 开源项目中，以“模块直挂（工程化版）”方式新增 `ai-security` 模块。

核心原则：

> 让 `/ai-security` 在产品体验上属于官方 `/system-settings/security/sensitive-words` 的高级功能，但在代码架构上保持完全独立。

## 设计目标

1. 不影响官方 new-api 后续升级。
2. `/ai-security` 数据独立保存。
3. 提供 `install` 初始化能力。
4. 扩展官方 sensitive-words，提供更丰富的敏感词高级功能。
5. 整个二开功能只保留一个模块：`custom/ai-security`。

## 请求检测链路

保持官方 sensitive-words 功能完全不变，在其后增加 ai-security 高级检测。

```text
用户请求
  ↓
new-api 原有鉴权 / 额度 / 渠道逻辑
  ↓
官方 sensitive-words 基础检测
  ↓
ai-security 高级检测
  ↓
模型请求
  ↓
ai-security 响应检测
  ↓
返回用户
```

`ai-security` 必须拥有独立开关，关闭后不影响官方 sensitive-words。同样的关闭官方 sensitive-words也不影响ai-security功能

## 目录结构

```text
new-api/
├── 官方模块
└── custom/
    └── ai-security/
        ├── api/          # 后端接口
        ├── service/      # 后端业务
        ├── engine/       # 检测引擎：AI / NER / Regex / Keyword
        ├── web/          # ai-security 前端页面、组件、接口请求
        ├── migration/    # 数据库迁移：建表 / 升级
        ├── seed/         # 默认规则 / 默认配置
        ├── install.sh    # 初始化脚本
        └── version.json  # 模块版本信息
```

## 前端改动范围

主项目前端只允许做两件事：

1. 注册 `/ai-security` 页面路由。
2. 在菜单 / 导航栏增加 `/ai-security` 入口。

`/ai-security` 页面、组件、状态管理、接口请求、样式全部放在：

```text
custom/ai-security/web/
```

不要把 ai-security 的业务页面散落到官方前端目录中。

## 后端改动范围

主项目后端只允许做四件事：

1. 在 router 中注册 `/api/ai-security/*`。
2. 在 Init 阶段初始化 ai-security。
3. 在 relay 请求转发前执行 `CheckRequest`。
4. 在 relay 响应返回前执行 `CheckResponse`。


## 后端路由

新增后端接口统一使用：

```text
/api/ai-security/*
```

建议接口：

```text
GET    /api/ai-security/configs
POST   /api/ai-security/configs
GET    /api/ai-security/rules
POST   /api/ai-security/rules
PUT    /api/ai-security/rules/:id
DELETE /api/ai-security/rules/:id
GET    /api/ai-security/logs
GET    /api/ai-security/dashboard
POST   /api/ai-security/sync/official-sensitive-words
```

## 数据独立设计

所有 ai-security 数据必须使用独立表前缀：

```text
aisec_
```

推荐表：

```text
aisec_configs
aisec_rules
aisec_word_groups
aisec_words
aisec_policies
aisec_hit_logs
aisec_daily_stats
aisec_sync_state
aisec_migrations
```

不要修改官方 options 表结构。

如需读取官方 sensitive-words，只能作为兼容来源读取或单向导入，不允许直接覆盖官方敏感词逻辑。

## 默认数据初始化

默认规则、默认配置必须放在：

```text
custom/ai-security/seed/
```

默认规则应在首次启动或执行 `install.sh` 时写入数据库。

要求：

1. 使用唯一 `code` 字段识别默认规则。
2. 已存在的规则不能被覆盖。
3. 新增默认规则时，下次初始化自动补充。
4. 用户修改过的规则不能被 seed 重置。

## install.sh 职责

`custom/ai-security/install.sh` 负责：

1. 检查 ai-security 模块目录。
2. 执行数据库迁移。
3. 初始化默认配置。
4. 初始化默认规则。
5. 创建菜单入口所需配置。
6. 注册插件状态。
7. 输出当前 ai-security 版本。

`install.sh` 必须可以重复执行，并且重复执行不能破坏已有数据。

## 检测引擎设计

检测引擎放在：

```text
custom/ai-security/engine/
```

至少预留以下能力：

```text
Keyword  # 关键词 / 敏感词
Regex    # 正则规则
NER      # 实体识别
AI       # AI 语义检测
```

统一返回结构：

```text
blocked: 是否拦截
risk_level: low / medium / high
action: allow / warn / replace / block / review
reason: 命中原因
rule_id: 命中规则
matched_text: 命中片段
```

## 日志设计

命中日志存入：

```text
aisec_hit_logs
```

日志必须支持 dashboard 使用。

日志字段至少包括：

```text
request_id
user_id
token_id
model_name
channel_id
direction
rule_id
risk_level
action
matched_text
hit_reason
created_at
```

不要默认保存完整用户 prompt，应优先保存命中片段、脱敏内容或 hash。

## Dashboard 设计

Dashboard 不要全部依赖前端计算。

推荐：

1. 今日实时数据从 `aisec_hit_logs` 聚合。
2. 历史趋势从 `aisec_daily_stats` 读取。
3. 高频敏感词、高风险用户、模型风险排行从日志聚合。

## 官方 sensitive-words 兼容关系

官方功能保留：

```text
/system-settings/security/sensitive-words
```

ai-security 作为高级功能：

```text
/ai-security
```

产品体验：

```text
官方 sensitive-words = 基础敏感词设置
ai-security = 高级安全中心
```

架构关系：

```text
官方 sensitive-words 不改核心逻辑
ai-security 可读取官方敏感词
ai-security 可单向导入官方敏感词
ai-security 不直接覆盖官方 sensitive-words
```

## 升级兼容要求

开发时必须遵守：

1. 不修改官方 sensitive-words 核心逻辑。
2. 不修改 Docker ENTRYPOINT。
3. 不修改 main.go 启动入口，除非只增加一个非常小的 Init 调用。
4. 不修改官方 options 表结构。
5. 不把 ai-security 业务逻辑写进官方 controller / service / model。
6. 主项目只保留必要挂载点。

## 最小主项目改动清单

前端：

```text
1. 注册 /ai-security 路由
2. 菜单增加 /ai-security 入口
```

后端：

```text
1. router 注册 /api/ai-security/*
2. Init 阶段初始化 ai-security
3. relay 请求前执行 ai-security.CheckRequest
4. relay 响应前执行 ai-security.CheckResponse
```

第一阶段可以先实现：

```text
1. /api/ai-security/* 后端接口
2. /ai-security 前端页面
3. 默认规则 seed
4. 请求前 CheckRequest
5. 命中日志
6. Dashboard 基础统计
```

响应检测 `CheckResponse` 可作为第二阶段实现。

## 验收标准

完成后必须满足：

1. 关闭 ai-security 后，官方 sensitive-words 正常工作。
2. 开启 ai-security 后，请求会经过高级检测。
3. 默认规则在首次安装后自动出现在 `/ai-security/rule`。
4. 重新编译后，已有配置、规则、日志不丢失。
5. 同步官方 new-api 代码时，冲突点尽量只出现在路由、菜单、relay 挂载点。
6. `install.sh` 可重复执行且不会覆盖用户数据。
7. 所有 ai-security 表均使用 `aisec_` 前缀。
8. `/ai-security` 页面不依赖官方 sensitive-words 页面内部实现。


### ai-security 页面实现

https://ai-api.cncarecc.com/ai-security/logs 显示 相关匹配日志
1 显示时间、用户、模型、动作、规则、分组、Risk、score
2 具有过滤功能，可以基于模型、用户、规则、分组过滤、规则过滤等等

https://ai-api.cncarecc.com/ai-security/rules 显示匹配规则
1 规则具有名称、归属分组、类型（关键字、正则、NER、AI识别）、动作（放行、告警、脱敏、拦截、审核） （risk score属性）
2 规则具有测试、编辑、复制、删除这几个操作

https://ai-api.cncarecc.com/ai-security/groups 显示分组信息，规则必须属于某个分组
1 分组有名称、父组、说明信息、状态、操作（编辑、删除）这几个属性
2 当分组有父组时，分组的规则需要叠加并包含
3 最多允许5层分组，也即是分组5-父组分组4-父组分组3-父组分组2-父组分组1-父组分组0

https://ai-api.cncarecc.com/ai-security/policies
1 策略具有用户、分组、作用域（仅请求、仅响应、双向）、默认动作（放行、告警、脱敏、拦截、审核）、优先级等属性

https://ai-api.cncarecc.com/ai-security/Dashboard 显示数据看板
1 支持基于时间过滤，可查看今天、本周、本月、自定义
2 具有以下分类 累计检测、拦截、告警、今日检测、今日拦截
3 有热门分类 风险分布、热门用户、热门模型 等top展示
4 具有过滤功能，可以根据用户过滤、分组过滤、规则过滤等等
5 可以查看不同用户的数据看板