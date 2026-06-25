# Feature: AI 内容安全高级检测模块（ai-security）

## Problem Statement

`new-api` 已经内置基础的 `sensitive-words` 功能，但只能做简单的关键词拦截，缺少分层管理、规则类型、风险评分、动作策略、命中日志和可视化看板等高级能力。管理员需要在不破坏官方功能的前提下，获得一套可独立升级、独立配置、独立数据的高级 AI 内容安全中心，从而更精细地控制请求和响应中的敏感内容。

## User Scenarios & Testing

### Scenario 1：管理员首次安装并查看默认规则

**As a** 系统管理员  
**I want to** 首次部署模块后自动看到预置的默认安全规则  
**So that** 无需手动逐条创建规则即可开启基础防护

**Acceptance Criteria:**
- 执行初始化后，`/ai-security/rules` 页面展示默认规则列表
- 默认规则具有唯一标识，重复初始化不会覆盖已有规则
- 用户修改过的默认规则在重复初始化后保留用户修改

### Scenario 2：管理员配置分组与规则

**As a** 系统管理员  
**I want to** 创建支持嵌套的分组，并在分组下创建关键词、正则、NER、AI 识别等规则  
**So that** 可以按业务维度组织安全策略

**Acceptance Criteria:**
- 分组支持最多 5 层嵌套，子分组规则可叠加到父分组
- 规则支持类型、动作、风险评分、优先级等属性
- 规则支持测试、编辑、复制、删除操作

### Scenario 3：管理员为用户绑定安全策略

**As a** 系统管理员  
**I want to** 为指定用户绑定分组策略并设置作用域和默认动作  
**So that** 不同用户可应用不同的安全标准

**Acceptance Criteria:**
- 策略可设置作用域：仅请求、仅响应、双向
- 策略可设置默认动作：放行、告警、脱敏、拦截、审核
- 策略支持优先级排序

### Scenario 4：API 请求经过高级安全检测

**As a** 终端用户  
**I want to** 我发起的对话请求在转发给模型前经过高级安全检测  
**So that** 敏感内容可以被拦截、告警或脱敏

**Acceptance Criteria:**
- 开启 ai-security 后，请求按策略执行检测
- 命中规则时根据动作返回拦截、脱敏后的内容或继续流转
- 命中日志记录请求 ID、用户、模型、动作、规则、风险等级等

### Scenario 5：管理员查看安全看板

**As a** 系统管理员  
**I want to** 在安全看板查看累计检测、拦截、告警、今日检测、今日拦截等指标  
**So that** 快速了解平台安全态势

**Acceptance Criteria:**
- 看板支持今天、本周、本月、自定义时间范围
- 展示风险分布、热门用户、热门模型等排行
- 支持按用户、分组、规则过滤

### Scenario 6：与官方 sensitive-words 解耦运行

**As a** 系统管理员  
**I want to** 单独开启或关闭 ai-security，不影响官方 sensitive-words  
**So that** 两个功能可以独立启用或灰度验证

**Acceptance Criteria:**
- 关闭 ai-security 时，官方 sensitive-words 正常工作
- 关闭官方 sensitive-words 时，ai-security 仍可独立工作
- ai-security 拥有独立的启用开关

## Functional Requirements

1. **FR-001 模块目录隔离**：所有 ai-security 业务代码必须集中在 `custom/ai-security/` 目录下，不允许分散到官方 controller / service / model / middleware / 前端 features 目录。
   - **Acceptance Criteria：** 官方目录中除必要的挂载点外，不存在 ai-security 业务逻辑文件。

2. **FR-002 后端路由隔离**：ai-security 后端接口统一使用 `/api/ai-security/*` 前缀。
   - **Acceptance Criteria：** 所有管理接口、检测接口、同步接口均位于 `/api/ai-security/` 下。

3. **FR-003 前端路径隔离**：前端管理页面统一使用 `/ai-security/*` 路径。
   - **Acceptance Criteria：** 浏览器访问 `/ai-security` 及其子路径可进入高级安全中心。

4. **FR-004 菜单入口从属关系**：`/ai-security` 入口在菜单层级上属于官方 `/system-settings/security/sensitive-words` 的高级功能，不作为独立一级菜单。
   - **Acceptance Criteria：** 管理员从系统设置的安全模块可导航到 ai-security 高级功能。

5. **FR-005 数据表前缀**：所有 ai-security 数据表必须使用 `aisec_` 前缀。
   - **Acceptance Criteria：** 数据库中存在的 ai-security 相关表名均以 `aisec_` 开头。

6. **FR-006 数据独立保存**：ai-security 的配置、规则、策略、命中日志、审计日志、每日统计、同步状态、迁移记录使用独立的数据结构保存，不修改官方 options 表结构。
   - **Acceptance Criteria：** 官方 `options` 表结构与官方版本保持一致；ai-security 新增表不影响官方表。

7. **FR-007 独立开关**：ai-security 拥有独立的启用开关，开关状态不影响官方 sensitive-words 的启停。
   - **Acceptance Criteria：** 在 ai-security 关闭时官方 sensitive-words 仍可拦截；在官方 sensitive-words 关闭时 ai-security 仍可检测。

8. **FR-008 默认规则初始化**：模块提供 install 初始化能力，首次安装或执行初始化脚本后，自动写入默认规则和默认配置。
   - **Acceptance Criteria：** 全新部署后，`/ai-security/rules` 展示默认规则；重复初始化不覆盖用户修改。

9. **FR-009 规则管理**：支持创建、编辑、复制、删除、测试规则；规则支持名称、归属分组、类型、动作、风险评分、优先级等属性。
   - **Acceptance Criteria：** 管理员可在 `/ai-security/rules` 完成上述操作并立即生效。

10. **FR-010 分组管理**：支持创建、编辑、删除分组；分组支持父组、说明、状态；最多支持 5 层嵌套，子分组规则叠加生效。
    - **Acceptance Criteria：** 创建 5 层嵌套分组后，底层分组继承上层分组规则；删除父分组时处理子分组和规则。

11. **FR-011 策略管理**：支持按用户绑定分组策略，设置作用域（仅请求、仅响应、双向）和默认动作（放行、告警、脱敏、拦截、审核）及优先级。
    - **Acceptance Criteria：** 不同用户请求时按绑定策略生效；策略优先级影响最终动作。

12. **FR-012 请求检测**：在请求转发给模型前执行 ai-security 检测，根据策略动作进行放行、告警、脱敏或拦截。
    - **Acceptance Criteria：** 命中拦截规则时返回拦截提示；命中脱敏规则时下游接收到脱敏后的请求内容。

13. **FR-013 响应检测（第二阶段）**：在模型响应返回用户前执行 ai-security 检测。
    - **Acceptance Criteria：** 响应内容命中规则时按策略动作处理。

14. **FR-014 命中日志**：记录每次命中的请求 ID、用户、Token、模型、渠道、方向、规则、分组、风险等级、动作、命中片段、原因、时间。
    - **Acceptance Criteria：** `/ai-security/logs` 可查看并过滤命中日志；日志不默认保存完整用户 prompt。

15. **FR-015 Dashboard 看板**：提供累计检测、拦截、告警、今日检测、今日拦截等统计，以及风险分布、热门用户、热门模型排行。
    - **Acceptance Criteria：** `/ai-security/dashboard` 展示上述指标并支持时间范围过滤。

16. **FR-016 官方敏感词兼容**：ai-security 可读取官方 sensitive-words 作为兼容来源，并可单向导入到 ai-security 规则库，不允许直接覆盖官方 sensitive-words 逻辑。
    - **Acceptance Criteria：** 提供同步入口将官方敏感词导入为 ai-security 规则；导入后官方 sensitive-words 保持不变。

17. **FR-017 升级兼容性**：不修改官方 sensitive-words 核心逻辑、Docker ENTRYPOINT、main.go 启动入口（除非增加一个极小的 Init 调用）、官方 options 表结构。
    - **Acceptance Criteria：** 合并官方 new-api 更新时，冲突点仅出现在路由注册、菜单挂载、relay 检测挂载点。

## Non-Functional Requirements

1. **NFR-001 可重复初始化**：`install.sh` 可重复执行，不会破坏或覆盖已有用户数据和用户修改过的规则。
   - **Acceptance Criteria：** 重复执行 `install.sh` 后，已有配置、规则、日志保持不变，新增默认规则自动补充。

2. **NFR-002 模块可卸载**：移除 `custom/ai-security/` 目录并还原挂载点后，官方 new-api 仍可正常编译运行。
   - **Acceptance Criteria：** 删除模块目录并移除挂载代码后，项目编译通过且官方功能正常。

3. **NFR-003 检测性能**：请求检测应在合理时间内完成，不显著影响 API 响应延迟。
   - **Acceptance Criteria：** 非 AI 规则检测在常规请求下延迟增加可忽略；AI 检测设置超时降级机制。

4. **NFR-004 数据安全**：命中日志不默认保存完整用户 prompt，优先保存命中片段、脱敏内容或 hash。
   - **Acceptance Criteria：** 日志表中不存在完整原始 prompt 字段或默认不写入。

## Success Criteria

- 管理员可以在 5 分钟内通过 install 初始化完成 ai-security 部署并看到默认规则。
- 90% 以上的命中请求能够在日志中完整记录规则、分组、动作、风险等级和命中片段。
- 看板指标在用户选择时间范围后 3 秒内完成加载。
- 合并官方 new-api 更新时，非挂载点文件冲突不超过 5 处。
- 在 ai-security 关闭状态下，官方 sensitive-words 拦截率与未安装模块前保持一致。
- 重复执行 install 后，用户已修改的规则 100% 保留。

## Key Entities

- **Config（配置）**：模块全局配置，如启用开关、AI 检测超时、日志保留天数等。
- **Group（规则分组）**：安全规则的分类容器，支持嵌套和状态控制。
- **Rule（规则）**：具体的检测规则，包含类型、内容、动作、风险评分、优先级等。
- **Policy（用户策略）**：用户与分组的绑定关系，定义作用域、默认动作、优先级。
- **HitLog（命中日志）**：每次检测命中的记录，用于审计和看板分析。
- **AuditLog（操作日志）**：管理员对配置、规则、分组、策略的变更记录。
- **DailyStats（每日统计）**：按天聚合的检测、拦截、告警等统计指标。
- **SyncState（同步状态）**：记录与官方 sensitive-words 的同步状态。

## Assumptions

- 目标用户为 new-api 的系统管理员，具备管理官方 sensitive-words 的经验。
- 官方 new-api 的版本支持模块化挂载（router、init、relay 有可扩展的接入点）。
- AI 检测引擎调用外部模型或本地模型，超时或失败时降级为本地规则检测。
- 默认规则种子数据由中国法律法规、常见企业合规场景和平台安全实践构成。

## Out of Scope

- 修改官方 sensitive-words 的检测逻辑或数据结构。
- 替代 new-api 原有的鉴权、额度、渠道分发逻辑。
- 提供非中文/英文的多语言前端界面（zh/en 优先，其他语言后续补充）。
- 实时告警通知（如邮件、Webhook）作为第一阶段功能。

## Open Questions

- AI 检测引擎默认使用哪个模型接口？是否需要单独配置渠道和 API Key？
- 默认规则种子数据的具体内容是否需要按行业定制？
- 命中日志的默认保留时长是多少天？
