# Tasks: AI Content Security Management Enhancement

**Input**: Design documents from `/specs/002-ai-content-security-v2/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), contracts/api.md, quickstart.md

**Tests**: Test tasks are OPTIONAL — validation is primarily via manual testing per quickstart.md scenarios.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new shared components and hooks that multiple user stories will depend on.

- [x] T001 [P] Create `web/default/src/components/empty-state.tsx` — reusable EmptyState component with types: initial, empty, filtered, forbidden
- [x] T002 [P] Create `web/default/src/components/confirm-dialog.tsx` — reusable ConfirmDialog component for destructive actions
- [x] T003 [P] Create `web/default/src/features/security/components/admin-guard.tsx` — AdminGuard wrapper that shows/hides children based on user.role === 'admin'
- [x] T004 [P] Create `web/default/src/features/security/components/security-tabs.tsx` — Tab navigation component with 5 tabs (Dashboard, Groups, Rules, Policies, Audit Logs), click triggers navigate()
- [x] T005 [P] Create `web/default/src/features/security/components/security-page-layout.tsx` — SecurityPageLayout wrapper with title + tabs + children slot
- [x] T006 [P] Create `web/default/src/features/security/api/query-keys.ts` — TanStack Query key definitions for all security entities
- [x] T007 [P] Create `web/default/src/features/security/hooks/use-url-filters.ts` — Generic useUrlFilters hook using TanStack Router useSearch + Zod validation
- [x] T008 Add security i18n keys to `web/default/src/i18n/locales/en.json` and `zh.json`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Backend database changes and API enhancements that MUST be complete before frontend user story implementation.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T009 [P] Add `priority` column to `model/security.go` SecurityUserPolicy struct and run database migration
- [x] T010 [P] Add `Priority` field to `dto/security.go` SecurityPolicyRequest and SecurityPolicyResponse
- [x] T011 [P] Add `TodayInterceptions` to `dto/security.go` SecurityDashboardResponse Summary struct
- [x] T012 Add `name` fuzzy search parameter to `service/security/group.go` GetSecurityGroups function
- [x] T013 Change copy suffix to Chinese in `service/security/group.go` CopySecurityGroup (`(副本)` instead of `_copy`)
- [x] T014 Add `start_time`, `end_time`, `model_name` parameters to `service/security/hitlog.go` SecurityLogQueryParams and GetSecurityLogs
- [x] T015 Add `start_time`, `end_time`, `model_name` parameters to `service/security/hitlog.go` ExportSecurityLogParams and GetSecurityLogsForExport
- [x] T016 Add `TodayInterceptions` calculation to `service/security/dashboard.go` GetSecurityDashboard
- [x] T017 Add `Priority` handling to `service/security/policy.go` CreateSecurityPolicy and UpdateSecurityPolicy; sort by priority ASC
- [x] T018 Add UpdateSecurityRuleStatus, TestSecurityRule, BatchDeleteSecurityRules, BatchUpdateSecurityRuleStatus to `service/security/rule.go`
- [x] T019 Add DetectWithRule method to `service/security/detector.go`
- [x] T020 Update `controller/security.go` — add `name` param to GetSecurityGroups; add `start_time`/`end_time`/`model_name` to GetSecurityLogs and ExportSecurityLogs
- [x] T021 Add new controllers to `controller/security.go`: TestSecurityRule, UpdateSecurityRuleStatus, BatchDeleteSecurityRules, BatchUpdateSecurityRuleStatus, GetSecurityMigrationStatus
- [x] T022 Register new routes in `router/api-router.go`: POST `/rules/:id/test`, PATCH `/rules/:id/status`, POST `/rules/batch-delete`, POST `/rules/batch-status`, GET `/migration-status`

**Checkpoint**: Foundation ready — all new backend APIs are available and returning correct data

---

## Phase 3: User Story 1 — Unified Security Navigation (Priority: P1) 🎯 MVP

**Goal**: Administrators can navigate between all Security sub-pages using clear in-page Tab navigation, with the sidebar highlighting correctly for all `/security/*` routes.

**Independent Test**: Open `/security`, click each tab (Groups, Rules, Policies, Logs, Dashboard), verify URL changes and correct page loads without full refresh. Verify sidebar Security item stays highlighted on all sub-pages.

### Implementation for User Story 1

- [x] T023 [US1] Modify `web/default/src/hooks/use-sidebar-data.ts` — update Security entry `activeUrls` to match `/security/*` prefix
- [x] T024 [P] [US1] Wrap each security page (`dashboard-page.tsx`, `group-page.tsx`, `rule-page.tsx`, `policy-page.tsx`, `log-page.tsx`) with SecurityPageLayout and SecurityTabs
- [x] T025 [US1] Remove legacy Sensitive Words section from `web/default/src/features/system-settings/security/section-registry.tsx`
- [x] T026 [US1] Verify all `/security/*` routes render correctly with Tab navigation and no console errors

**Checkpoint**: User Story 1 is fully functional — navigation works across all 5 pages, old system entry removed

---

## Phase 4: User Story 2 — Consolidated Content Security System (Priority: P1)

**Goal**: The legacy Sensitive Words system is fully replaced by the new module; global toggle on Dashboard controls all detection uniformly.

**Independent Test**: Disable global toggle on Dashboard, submit sensitive content via chat, verify no detection occurs. Enable toggle, verify detection resumes. Verify legacy Sensitive Words panel is absent from System Settings.

### Implementation for User Story 2

- [x] T027 [US2] Add global security status toggle to `web/default/src/features/security/pages/dashboard-page.tsx` — calls `/api/security/status`, shows enabled/disabled state
- [x] T028 [US2] Add migration status banner to `web/default/src/features/security/pages/group-page.tsx` — calls `/api/security/migration-status`, shows banner if `migrated: true`, dismissible with localStorage
- [x] T029 [US2] Verify middleware `middleware/security.go` already checks both `SECURITY_ENABLED` and `CheckSensitiveEnabled` (FIX-1 from backend checklist)

**Checkpoint**: User Story 2 is fully functional — global toggle works, legacy system removed, migration banner shows

---

## Phase 5: User Story 3 — Advanced Audit Log Filtering (Priority: P1)

**Goal**: Administrators can filter audit logs by time range and model name with URL-synced filters that survive refresh.

**Independent Test**: Go to Logs page, select date range June 1-10, enter model name "gpt-4", verify filtered results. Copy URL, open in new tab, verify same filters applied.

### Implementation for User Story 5

- [x] T030 [P] [US3] Create `web/default/src/features/security/hooks/use-log-filters.ts` — Zod schema for log filters (start_time, end_time, model_name, user_id, action, risk_level, content_type) with useUrlFilters
- [x] T031 [P] [US3] Enhance `web/default/src/features/security/pages/log-page.tsx` — add filter bar with date range picker, model name input, user selector, action/risk_level/content_type dropdowns; sync all to URL
- [x] T032 [US3] Update `web/default/src/features/security/api/security.ts` — add `start_time`, `end_time`, `model_name` to getLogs and exportLogs API methods
- [x] T033 [US3] Verify export functionality passes current filters to backend

**Checkpoint**: User Story 3 is fully functional — all log filters work, URL sync verified, export respects filters

---

## Phase 6: User Story 4 — Rule Testing and Batch Management (Priority: P2)

**Goal**: Administrators can test individual rules and perform batch enable/disable/delete on multiple rules.

**Independent Test**: Open Rules page, click "Test" on a rule, enter content, verify detection result. Select 3 rules, click "Batch Disable", confirm, verify all 3 show disabled.

### Implementation for User Story 4

- [x] T034 [P] [US4] Create `web/default/src/features/security/hooks/use-rule-filters.ts` — Zod schema for rule filters (group_id, type, status) with useUrlFilters
- [x] T035 [P] [US4] Enhance `web/default/src/features/security/pages/rule-page.tsx` — add filter bar (group dropdown, type dropdown, status dropdown), URL sync; add row selection checkboxes, batch action toolbar
- [x] T036 [US4] Create rule test modal component `web/default/src/features/security/components/rule-tester.tsx` — input content, call POST `/api/security/rules/:id/test`, show detection result
- [x] T037 [US4] Update `web/default/src/features/security/api/security.ts` — add testRule, updateRuleStatus, batchDeleteRules, batchUpdateRuleStatus methods
- [x] T038 [US4] Add "Test" button to rule row actions and integrate RuleTester modal
- [x] T039 [US4] Implement batch enable/disable/delete with ConfirmDialog confirmation

**Checkpoint**: User Story 4 is fully functional — rule testing works, batch operations work with confirmation

---

## Phase 7: User Story 6 — Policy Priority Management (Priority: P2)

**Goal**: Administrators can assign numeric priorities to policies and see them sorted by priority.

**Independent Test**: Create two policies with priorities 10 and 5. Verify priority 5 appears first. Edit priority 10 to 1, verify it moves to top.

### Implementation for User Story 7

- [x] T040 [P] [US6] Create `web/default/src/features/security/hooks/use-policy-filters.ts` — Zod schema for policy filters with useUrlFilters
- [x] T041 [P] [US6] Enhance `web/default/src/features/security/pages/policy-page.tsx` — add filter bar (user selector with debounce search, status dropdown), URL sync; add Priority column and sort indicator
- [x] T042 [US6] Enhance `web/default/src/features/security/components/policy-form-modal.tsx` — add Priority number input (1-100); replace user_id numeric input with user search selector (debounce 300ms); replace whitelist_ips text input with Tag input (comma/ newline separated)
- [x] T043 [US6] Update `web/default/src/features/security/api/security.ts` — ensure priority field is included in createPolicy and updatePolicy payloads

**Checkpoint**: User Story 6 is fully functional — policies sort by priority, form has user selector and IP tags

---

## Phase 8: User Story 5 — Real-Time Audit Log Monitoring (Priority: P2)

**Goal**: Administrators can enable auto-refresh on Audit Logs with notification banners for new data, and polling pauses in inactive tabs.

**Independent Test**: Enable auto-refresh on Logs page, trigger a detection event, verify notification banner appears (not auto-replace). Switch tabs, trigger another event, switch back, verify no new banner (polling paused).

### Implementation for User Story 5

- [x] T044 [US5] Create `web/default/src/features/security/api/use-security-logs.ts` — TanStack Query hook with refetchInterval: 5000, refetchIntervalInBackground: false
- [x] T045 [US5] Enhance `web/default/src/features/security/pages/log-page.tsx` — add auto-refresh toggle switch; when new data arrives, show notification banner "N new logs available" instead of replacing table; click banner to refresh
- [x] T046 [US5] Implement tab visibility detection — when document.hidden, disable polling; when visible again, resume polling
- [x] T047 [US5] Add maximum polling duration limit (30 minutes) — auto-disable after timeout with toast notification

**Checkpoint**: User Story 5 is fully functional — auto-refresh works, banner notifications work, tab pause works

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements that affect multiple user stories

- [x] T048 [P] Enhance `web/default/src/features/security/pages/dashboard-page.tsx` — add time range filter (today/week/month/custom) synced to URL; add refresh button with auto-refresh toggle (30s); wrap charts in React.lazy for code splitting
- [x] T049 [P] Enhance `web/default/src/features/security/pages/group-page.tsx` — add URL filters (status, parent_id, name), parent group column in table, copy button calling existing copy API; add "系统迁移" info banner
- [x] T050 [P] Wrap Dashboard chart components (`risk-distribution-chart.tsx`, `top-categories-chart.tsx`) in `React.lazy()` with Suspense fallback
- [x] T051 Add skeleton loading states to all Security pages for initial data fetch
- [x] T052 Add toast feedback for all mutations (create/update/delete/copy/batch) using sonner
- [x] T053 Add loading states to all form submit buttons (disabled + spinner during mutation)
- [x] T054 Add content masking (手机号 → 138****8000, 身份证 → 110101********xxxx) to `web/default/src/features/security/components/log-detail-drawer.tsx`
- [x] T055 Verify all user input in log details renders as plain text (no dangerouslySetInnerHTML)
- [ ] T056 Run quickstart.md validation scenarios 1-6 and fix any issues
- [x] T057 Run `bun run typecheck` in `web/default/` and fix TypeScript errors
- [ ] T058 Run `bun run i18n:sync` in `web/default/` to sync translations

---

## Phase 10: Bug Fixes (Post-Implementation)

**Purpose**: Fix the bugs documented in `my-spec-project/BUGFIX-PLAN.md` before final release.

### Bug Fix 1: Mask Action Does Not Replace Keywords with `***` (P1)

- [ ] T059 [P] Update `service/security/detector.go::maskText()` to return `"***"`
- [ ] T060 [P] Fix `service/security/detector.go::applyMasking()` to sort matches by `Position[0]` descending
- [ ] T061 Refactor `middleware/security.go` content extraction/replacement to operate per-message instead of joining with `\n`
- [ ] T062 Add fallback warning log in `middleware/security.go` when Mask replacement does not modify the body
- [ ] T063 Update `service/security/detector_test.go` expectations for `TestMaskText` and `TestApplyMasking`
- [ ] T064 Add middleware-level test or manual validation for multi-message Mask scenario

**Checkpoint**: Single-message and multi-message Mask requests both produce `***`; audit logs show `processed_content` with `***`.

### Bug Fix 2: Form and List Labels Are Not Localized (P2)

- [ ] T065 [P] Create `web/default/src/features/security/constants.ts` with `RULE_TYPES`, `ACTIONS`, `SCOPES`, `STATUSES`, `RISK_LEVELS`
- [ ] T066 [P] Replace hardcoded option arrays in `rule-form-modal.tsx`, `policy-form-modal.tsx`, `group-form-modal.tsx` with imports from `constants.ts`
- [ ] T067 [P] Replace hardcoded mapping objects in `rule-page.tsx`, `policy-page.tsx`, `log-page.tsx` with imports from `constants.ts`
- [ ] T068 Wrap all security labels with `t()` and add keys to `web/default/src/i18n/locales/en.json` and `zh.json`
- [ ] T069 Add safe fallback for `SelectValue` when value has no matching option
- [ ] T070 Run `bun run typecheck` and `bun run i18n:sync`

**Checkpoint**: All security forms and lists display localized labels in both Chinese and English.

### Bug Fix 3: Block Action False-Positives After First Hit (P1/P2)

- [x] T071 Add structured debug logging to `service/security/detector.go::Detect()`
- [x] T072 Add structured debug logging to `service/security/engine_keyword.go::Detect()`
- [x] T073 Add structured debug logging to `middleware/security.go::SecurityCheck()` and `SecurityCheckResponse()`
- [ ] T074 Reproduce false-positive scenario and collect logs to confirm root cause
- [ ] T075 Implement cache TTL in `service/security/cache.go` (use `SecurityCacheExpiration` or Redis TTL)
- [ ] T076 Unify policy caching strategy: use `GetCachedUserPolicies()` in `detector.go` or remove the unused function
- [ ] T077 Add regression test in `service/security/engine_keyword_test.go` for consecutive detections (hit then miss then hit)
- [ ] T078 Verify Bug Fix 1 per-message replacement also resolves any middleware-side false positives

**Checkpoint**: Clean requests are no longer blocked; logs confirm zero hits for non-matching content; cache expires correctly.

### Bug Fix 4: Group Status Cannot Be Changed to Disabled (P1)

- [x] T079 [P] Add `Status int` field to `dto/security.go` `SecurityGroupRequest` with `binding:"oneof=0 1"`
- [x] T080 Update `service/security/group.go::UpdateSecurityGroup()` to include `"status": req.Status` in updates map and call `InvalidateRuleCache()`
- [x] T081 [P] (Recommended UX) Add `UpdateSecurityGroupStatus` to `service/security/group.go`
- [x] T082 [P] (Recommended UX) Add `PATCH /api/security/groups/:id/status` controller in `controller/security.go`
- [x] T083 (Recommended UX) Register `PATCH /api/security/groups/:id/status` route in `router/api-router.go`
- [x] T084 [P] (Recommended UX) Add `updateGroupStatus` to `web/default/src/features/security/api/security.ts`
- [x] T085 (Recommended UX) Add row-level Switch toggle to `web/default/src/features/security/pages/group-page.tsx` and refresh list on success
- [x] T086 Verify edit-form status change persists after refresh and disabled groups stop participating in detection

**Checkpoint**: Group status can be disabled via form or toggle, persists across refresh, and correctly gates detection.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion conceptually, but can actually start in parallel since backend and frontend are separate. BLOCKS all user story frontend work.
- **User Stories (Phase 3+)**: All depend on Foundational phase completion for backend APIs
  - Phase 3 (US1) should be done first as it establishes the navigation framework
  - Phase 4-7 can proceed in parallel after Phase 3 (if staffed), or sequentially
- **Polish (Phase 9)**: Depends on all desired user stories being complete
- **Bug Fixes (Phase 10)**: Depends on the feature being functionally complete; can start after Phase 9 or in parallel with final polish tasks T056-T058. Bug Fix 1 should be done first because it may affect Bug Fix 3.

### User Story Dependencies

- **User Story 1 (P1)**: No dependencies on other stories — MUST be first as it provides the navigation framework all other stories use
- **User Story 2 (P1)**: Depends on US1 navigation being in place; independent otherwise
- **User Story 3 (P1)**: Depends on US1; independent of US2
- **User Story 4 (P2)**: Depends on US1; independent of US2/US3
- **User Story 5 (P2)**: Depends on US1 and US3 (shares Logs page); independent otherwise
- **User Story 6 (P2)**: Depends on US1; independent of others

### Within Each User Story

- URL filter hooks before page enhancements
- API client updates before page integrations
- Components before page wiring
- Story complete before moving to next

### Parallel Opportunities

- Phase 1 (Setup) tasks T001-T008 can all run in parallel
- Phase 2 (Foundational) backend tasks T009-T022 can run in parallel (most touch different files)
- After Phase 3 (US1 Navigation), Phase 4-7 can be worked on in parallel by different developers
- Phase 9 polish tasks T048-T058 can run in parallel
- Bug Fix 4 tasks T079-T086 can run in parallel and are independent of Bug Fixes 1-3

---

## Parallel Example: Phase 2 Backend Enhancement

```bash
# All backend service enhancements can run in parallel:
Task: "Add name fuzzy search to service/security/group.go"
Task: "Add log filtering to service/security/hitlog.go"
Task: "Add priority to service/security/policy.go"
Task: "Add batch ops to service/security/rule.go"
Task: "Add today interceptions to service/security/dashboard.go"
Task: "Add DetectWithRule to service/security/detector.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + Foundational)

1. Complete Phase 1: Setup (shared components)
2. Complete Phase 2: Foundational (backend API enhancements)
3. Complete Phase 3: User Story 1 (Navigation)
4. **STOP and VALIDATE**: Test navigation, verify old system removed
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (Navigation) → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 (Consolidation) + User Story 3 (Log Filtering) → Test → Deploy
4. Add User Story 4 (Rule Testing/Batch) + User Story 6 (Policy Priority) → Test → Deploy
5. Add User Story 5 (Real-time Monitoring) → Test → Deploy
6. Polish phase → Final validation → Deploy
7. Bug Fixes (Phase 10): Bug Fix 1 first (may affect Bug Fix 3), then Bug Fix 2 and Bug Fix 4 can proceed in parallel

### Parallel Team Strategy

With multiple developers:

1. Developer A: Phase 1 (Setup components) + Phase 3 (US1 Navigation)
2. Developer B: Phase 2 (Backend APIs T009-T022)
3. Once Phase 2 + Phase 3 complete:
   - Developer A: Phase 4 (US2 Consolidation) + Phase 6 (US5 Real-time)
   - Developer B: Phase 5 (US4 Rule Testing/Batch) + Phase 7 (US6 Policy Priority)
   - Developer C: Phase 8 (US3 Log Filtering)
4. All merge and run Phase 9 (Polish) together

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Backend tasks (Phase 2) can start immediately in parallel with Phase 1
