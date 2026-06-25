# Tasks: AI 内容安全高级检测模块（ai-security）

## Overview

本任务清单将当前分散在官方目录的 ai-security 代码重构为独立的 `custom/ai-security/` 模块。任务按用户故事组织，每个阶段都是可独立测试的增量。

## Dependencies

```text
Phase 1 (Setup) → Phase 2 (Foundational) → US1 (Install + Default Rules)
                                      ↓
                              US2 (Groups + Rules)
                                      ↓
                              US3 (Policies)
                                      ↓
                              US4 (Request Detection)
                                      ↓
                              US5 (Dashboard + Logs)
                                      ↓
                              US6 (Decoupling + Mounting)
                                      ↓
                              Final Phase (Polish + Tests)
```

## Phase 1: Setup

- [X] T001 Create `custom/ai-security/` directory structure per implementation plan
- [X] T002 Create `custom/ai-security/version.json` and initial `custom/ai-security/install.sh`
- [X] T003 Implement module initialization entry `custom/ai-security/init.go` exposing `Init()`, `RegisterRoutes()`, `RegisterRelayMiddleware()`
- [X] T004 Implement database migration framework `custom/ai-security/migration/migration.go` with version tracking in `aisec_migrations`
- [X] T005 Implement config data model `custom/ai-security/model/config.go` and default config seeding

## Phase 2: Foundational

- [X] T006 Define module constants and enums `custom/ai-security/constant/constant.go` (rule types, actions, risk levels, scopes)
- [X] T007 Define common DTOs `custom/ai-security/dto/common.go` and response wrappers
- [X] T008 Implement caching layer `custom/ai-security/service/cache.go` for rules, policies, and configs with `aisec:` prefix
- [X] T009 Implement audit logging service `custom/ai-security/service/audit.go` for tracking admin changes
- [X] T010 Implement seed loader `custom/ai-security/seed/seed.go` that loads default rules from `custom/ai-security/seed/default_rules.go`

## Phase 3: US1 - Admin installs module and sees default rules

**Story Goal**: 管理员执行 install 后，无需手动创建即可在 `/ai-security/rules` 看到默认规则。

**Independent Test Criteria**:
- `bash custom/ai-security/install.sh` runs successfully and creates `aisec_*` tables.
- `GET /api/ai-security/rules` returns at least one default rule after install.
- Re-running install does not overwrite user-modified rules.

- [X] T011 [US1] Implement config management API `custom/ai-security/api/config.go` with `GET /api/ai-security/configs` and `PUT /api/ai-security/configs`
- [X] T012 [US1] Implement default rule seed data `custom/ai-security/seed/default_rules.go` with unique `code` fields
- [X] T013 [US1] Complete `custom/ai-security/install.sh` to run migrations, seed configs, seed rules, and output version
- [X] T014 [US1] Implement status API `custom/ai-security/api/status.go` returning enabled state, rule/group/policy counts, and version
- [X] T015 [US1] [P] Implement basic Dashboard page `custom/ai-security/web/pages/dashboard-page.tsx` showing module status card

## Phase 4: US2 - Admin configures groups and rules

**Story Goal**: 管理员可以创建嵌套分组和多种类型的检测规则，并测试规则效果。

**Independent Test Criteria**:
- Admin can create a 5-level nested group hierarchy.
- Admin can create keyword/regex/NER/AI rules under groups.
- Rule test API returns detected/masked/blocked results for sample content.

- [X] T016 [US2] Implement group data model `custom/ai-security/model/group.go` with parent/depth/path fields
- [X] T017 [US2] Implement group service `custom/ai-security/service/group.go` with CRUD, copy, and cascading delete
- [X] T018 [US2] Implement group API `custom/ai-security/api/group.go` exposing `GET/POST/PUT/PATCH/DELETE /api/ai-security/groups/*`
- [X] T019 [US2] Implement rule data model `custom/ai-security/model/rule.go` with type, action, priority, risk_score
- [X] T020 [US2] Implement rule service `custom/ai-security/service/rule.go` with CRUD, batch operations, and cache invalidation
- [X] T021 [US2] Implement rule API `custom/ai-security/api/rule.go` exposing `GET/POST/PUT/DELETE /api/ai-security/rules/*`
- [X] T022 [US2] Implement rule test API `custom/ai-security/api/rule_test.go` (handler) exposing `POST /api/ai-security/rules/:id/test`
- [X] T023 [US2] [P] Implement Groups management page `custom/ai-security/web/pages/group-page.tsx` with form modal
- [X] T024 [US2] [P] Implement Rules management page `custom/ai-security/web/pages/rule-page.tsx` with form modal and rule tester

## Phase 5: US3 - Admin binds security policies to users

**Story Goal**: 管理员可以按用户绑定分组策略，设置作用域、默认动作和优先级。

**Independent Test Criteria**:
- Admin can create a policy binding user X to group Y with scope and default action.
- Duplicate active policy for same user+group is rejected.
- Policy changes invalidate cache and take effect on next request.

- [X] T025 [US3] Implement policy data model `custom/ai-security/model/policy.go` with scope, default_action, whitelist_ips
- [X] T026 [US3] Implement policy service `custom/ai-security/service/policy.go` with CRUD and uniqueness validation
- [X] T027 [US3] Implement policy API `custom/ai-security/api/policy.go` exposing `GET/POST/PUT/DELETE /api/ai-security/policies/*`
- [X] T028 [US3] [P] Implement Policies management page `custom/ai-security/web/pages/policy-page.tsx` with form modal

## Phase 6: US4 - API requests pass through advanced security detection

**Story Goal**: 用户请求在转发给模型前经过 ai-security 检测，根据策略动作放行/告警/脱敏/拦截/审核。

**Independent Test Criteria**:
- Request containing blocked content returns interception response and creates a hit log.
- Request containing masked content reaches downstream with masked text.
- Detection works when official sensitive-words is disabled.

- [X] T029 [US4] Implement Keyword detection engine `custom/ai-security/engine/keyword.go`
- [X] T030 [US4] Implement Regex detection engine `custom/ai-security/engine/regex.go`
- [X] T031 [US4] Implement NER detection engine `custom/ai-security/engine/ner.go`
- [X] T032 [US4] Implement AI detection engine `custom/ai-security/engine/ai.go` with 3-second timeout and fallback
- [X] T033 [US4] Implement detection orchestrator `custom/ai-security/service/detector.go` merging multi-engine results and resolving final action
- [X] T034 [US4] Implement request detection middleware `custom/ai-security/middleware/check_request.go` integrated into relay request chain
- [X] T035 [US4] Implement hit logging service `custom/ai-security/service/hitlog.go` with async batch write
- [X] T036 [US4] Implement masking utility `custom/ai-security/service/mask.go` supporting full/preserve/replace strategies

## Phase 7: US5 - Admin views security dashboard

**Story Goal**: 管理员在安全看板查看累计检测、拦截、告警等指标，以及风险分布和热门排行。

**Independent Test Criteria**:
- Dashboard API returns summary, risk distribution, top categories/users/models.
- Logs page supports filtering by user, model, action, risk level, time range.
- Hit logs do not contain full raw prompts.

- [X] T037 [US5] Implement hit log data model `custom/ai-security/model/hitlog.go` storing hash + matched fragment instead of full prompt
- [X] T038 [US5] Implement log query API `custom/ai-security/api/log.go` exposing `GET /api/ai-security/logs` with filters
- [X] T039 [US5] Implement log export API `custom/ai-security/api/log_export.go` supporting CSV and Excel
- [X] T040 [US5] Implement daily stats aggregation service `custom/ai-security/service/dailystats.go` for historical trends
- [X] T041 [US5] Implement Dashboard stats API `custom/ai-security/api/dashboard.go` exposing `GET /api/ai-security/dashboard`
- [X] T042 [US5] [P] Implement Logs page `custom/ai-security/web/pages/log-page.tsx` with filters and detail drawer
- [X] T043 [US5] [P] Implement Dashboard chart components `custom/ai-security/web/components/dashboard-charts.tsx` and stat cards

## Phase 8: US6 - Decouple from official sensitive-words

**Story Goal**: ai-security 和官方 sensitive-words 可独立开关、独立运行，互不影响。

**Independent Test Criteria**:
- Closing ai-security does not affect official sensitive-words interception.
- Closing official sensitive-words does not affect ai-security detection.
- Official sensitive-words can be one-way imported into ai-security rules.

- [X] T044 [US6] Implement independent enable switch reading from `aisec_configs` instead of `setting.CheckSensitiveEnabled`
- [X] T045 [US6] Implement official sensitive-words reader `custom/ai-security/service/official_sync.go` as read-only source
- [X] T046 [US6] Implement sync API `custom/ai-security/api/sync.go` exposing `POST /api/ai-security/sync/official-sensitive-words`
- [X] T047 [US6] Register ai-security backend routes in `router/api-router.go` via `ai_security.RegisterRoutes(apiRouter)`
- [X] T048 [US6] Register ai-security initialization in `main.go` `InitResources()` via `ai_security.Init()`
- [X] T049 [US6] Register request/response detection middleware in `router/relay-router.go` via `ai_security.CheckRequest()` and `ai_security.CheckResponse()`
- [X] T050 [US6] Register frontend route `/ai-security/*` and menu entry under `System Settings > Security`

## Final Phase: Polish & Cross-Cutting Concerns

- [X] T051 Implement frontend i18n `custom/ai-security/web/i18n/ai-security.json` with zh/en translations
- [X] T052 Write backend unit tests for detection engines and core services
- [X] T053 Write integration tests covering install, detection, masking, blocking, and log creation
- [X] T054 Perform upgrade compatibility test: sync latest official new-api and verify conflicts only at mount points
- [X] T055 Write module README `custom/ai-security/README.md` documenting install, architecture, and upgrade notes

## Parallel Opportunities

- **Frontend pages**: T015, T023, T024, T028, T042, T043 can be developed in parallel once foundational APIs are ready.
- **Detection engines**: T029, T030, T031, T032 can be implemented in parallel before T033 orchestrator.
- **Data models**: T016, T019, T025, T037 can be created in parallel during early phases.
- **APIs**: T011, T018, T021, T027, T038, T041 are parallel once corresponding services exist.

## Suggested MVP Scope

**MVP = US1 + US2 + US4 (请求检测)**

最小可运行版本包含：
1. 模块安装与默认规则（US1）
2. 分组和规则管理（US2）
3. 请求检测与命中日志（US4 核心链路）

响应检测（US4 后半）、Dashboard 统计（US5）、策略精细化管理（US3）、官方敏感词同步（US6 部分）可作为后续迭代。

## Task Count Summary

| Phase | Task Count |
|---|---|
| Phase 1: Setup | 5 |
| Phase 2: Foundational | 5 |
| Phase 3: US1 | 5 |
| Phase 4: US2 | 9 |
| Phase 5: US3 | 4 |
| Phase 6: US4 | 8 |
| Phase 7: US5 | 7 |
| Phase 8: US6 | 7 |
| Final Phase | 5 |
| **Total** | **55** |
