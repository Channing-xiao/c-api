# Implementation Plan: AI Content Security Management Enhancement

**Branch**: `002-ai-content-security-v2` | **Date**: 2026-06-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-ai-content-security-v2/spec.md`

## Summary

Enhance the existing AI Content Security Management module (v1) by unifying navigation, consolidating the legacy Sensitive Words system, enriching backend APIs with advanced filtering and batch operations, and improving the frontend data layer with URL-synced filters, optimistic updates, and real-time monitoring.

The v1 implementation established core backend APIs (Gin/GORM) and basic frontend pages (React/TypeScript). This v2 enhancement addresses usability gaps identified in the frontend design document and API gaps identified in the backend supplement checklist, without changing the underlying technology stack.

## Technical Context

**Language/Version**: Go 1.22+ (backend), TypeScript/React 19 (frontend)

**Primary Dependencies**:
- Backend: Gin web framework, GORM v2 ORM, go-redis
- Frontend: React 19, TanStack Router, TanStack Query, TanStack Table, Base UI, Tailwind CSS, Recharts, React Hook Form, Zod

**Storage**: SQLite / MySQL 5.7.8+ / PostgreSQL 9.6+ (all three supported via GORM)

**Testing**: Go test (backend), TypeScript type checking + manual validation (frontend)

**Target Platform**: Web application (desktop-primary admin dashboard)

**Performance Goals**:
- Audit log filtered queries: < 500ms for 1M records
- Rule test API: < 1s response time
- Batch operations (50 rules): < 2s
- Dashboard data cache: 5 minutes TTL

**Constraints**:
- Must maintain backward compatibility with existing security middleware
- Must support all three databases (SQLite, MySQL, PostgreSQL) without raw SQL divergence
- Must keep sidebar navigation flat (no nested sub-menus)
- Frontend bundle size increase for charts: < 50KB gzip

**Scale/Scope**:
- Admin-only interface (~10-100 admin users)
- Rules: up to 10,000 per deployment
- Audit logs: up to 1M records with pagination

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Specification-Driven Development | ✅ Pass | spec.md completed with user stories, requirements, success criteria |
| II. Test-First Quality Assurance | ✅ Pass | Acceptance scenarios defined per user story; implementation will include validation tests |
| III. Modular & Decoupled Architecture | ✅ Pass | Enhancement builds on existing modular backend (controller/service/model) and frontend feature-based structure |
| IV. Documentation as Code | ✅ Pass | All design documents and API contracts will be version-controlled in `specs/002-ai-content-security-v2/` |
| V. Continuous Validation & Review | ✅ Pass | Changes limited to existing security module; no direct commits to main branch planned |

**Re-check after Phase 1**: All design artifacts (data model, contracts, quickstart) will be committed alongside implementation.

## Project Structure

### Documentation (this feature)

```text
specs/002-ai-content-security-v2/
├── plan.md              # This file (/speckit-plan command output)
├── spec.md              # Feature specification (/speckit-specify command output)
├── checklists/
│   └── requirements.md  # Spec quality checklist
├── research.md          # (Skipped - tech stack already established in v1)
├── data-model.md        # (Skipped - data model exists in v1, incremental changes documented in spec)
├── quickstart.md        # Phase 1 output - validation guide
├── contracts/
│   └── api.md           # Phase 1 output - API contracts for new endpoints
└── tasks.md             # Phase 2 output (/speckit-tasks command)
```

### Source Code (repository root)

**Backend (Go)**:

```text
c-api/
├── controller/
│   └── security.go          # Security API controllers (enhance existing)
├── service/
│   └── security/
│       ├── detector.go      # Detection engine (add single-rule test)
│       ├── group.go         # Group service (add name filter, copy suffix)
│       ├── hitlog.go        # Log service (add time/model filtering)
│       ├── policy.go        # Policy service (add priority field)
│       ├── rule.go          # Rule service (add batch ops, status toggle, test)
│       └── dashboard.go     # Dashboard service (add today interceptions)
├── middleware/
│   └── security.go          # Security middleware (already unified with legacy)
├── model/
│   └── security.go          # Data models (add Priority to policy)
├── dto/
│   └── security.go          # DTOs (add Priority, TodayInterceptions)
├── router/
│   └── api-router.go        # Route definitions (add new endpoints)
└── constant/
    └── security.go          # Security constants (unchanged)
```

**Frontend (React/TypeScript)**:

```text
web/default/src/
├── features/security/
│   ├── api/
│   │   ├── security.ts          # API client (add new endpoints)
│   │   ├── query-keys.ts        # NEW: TanStack Query key definitions
│   │   └── use-security-*.ts    # NEW: TanStack Query hooks
│   ├── components/
│   │   ├── security-tabs.tsx    # NEW: Tab navigation component
│   │   ├── security-page-layout.tsx  # NEW: Unified page layout
│   │   ├── admin-guard.tsx      # NEW: Admin role guard
│   │   ├── empty-state.tsx      # NEW: Reusable empty state
│   │   ├── confirm-dialog.tsx   # NEW: Reusable confirm dialog
│   │   ├── group-form-modal.tsx     # Existing - enhance
│   │   ├── rule-form-modal.tsx      # Existing - enhance
│   │   ├── policy-form-modal.tsx    # Existing - enhance
│   │   ├── log-detail-drawer.tsx    # Existing - enhance
│   │   ├── risk-distribution-chart.tsx  # Existing - wrap in lazy
│   │   ├── top-categories-chart.tsx     # Existing - wrap in lazy
│   │   └── top-users-table.tsx          # Existing
│   ├── hooks/
│   │   └── use-url-filters.ts   # NEW: URL query param sync hooks
│   └── pages/
│       ├── dashboard-page.tsx   # Enhance: global toggle, time filter
│       ├── group-page.tsx       # Enhance: URL filters, copy, parent col
│       ├── rule-page.tsx        # Enhance: URL filters, batch ops, test
│       ├── policy-page.tsx      # Enhance: user selector, IP tags, priority
│       └── log-page.tsx         # Enhance: URL filters, real-time refresh
├── hooks/
│   └── use-sidebar-data.ts      # Modify: activeUrls for /security/*
├── features/system-settings/security/
│   └── section-registry.tsx     # Modify: remove sensitive-words section
└── i18n/locales/
    ├── en.json                  # Add security translations
    └── zh.json                  # Add security translations
```

**Structure Decision**: Single full-stack project with Go backend and React frontend. Enhancement is localized to the existing `features/security/` module on the frontend and `controller/service/model/dto/security*` on the backend. No new project structure needed.

## Post-Implementation Bug Fixes

After the v2 feature implementation, the following bugs were identified during end-to-end validation and must be fixed before the feature is considered complete. These fixes are tracked in `my-spec-project/BUGFIX-PLAN.md` and are summarized below for traceability within this implementation plan.

### Bug Fix 1: Mask Action Does Not Replace Keywords with `***`

**Priority**: P1

**Symptom**: When a rule with Action = Mask matches, the sensitive keyword is not replaced with `***`.

**Root Causes**:
1. `service/security/detector.go::maskText()` currently preserves first/last characters (e.g., `password` → `p******d`) instead of returning a fixed `***` mask.
2. `service/security/detector.go::applyMasking()` sorts matches by simply reversing the slice, which only works if matches are already ordered. It should sort by `Position[0]` descending.
3. `middleware/security.go::extractContentFromRequest()` joins multiple user messages with `\n`, but `replaceContentInRequest()` then tries to replace the joined string inside the original JSON. When there are multiple messages, the joined string does not exist in the raw request body, so masking silently fails. The same issue exists for responses.
4. `middleware/security.go::replaceContentInRequest()` and `replaceContentInResponse()` use `strings.Replace(..., -1)`, which can unintentionally replace occurrences in unrelated JSON fields.

**Fix**:
- Change `maskText()` to always return `"***"`.
- Fix `applyMasking()` to explicitly sort matches by `Position[0]` descending before replacing.
- Refactor request/response content extraction and replacement to operate per-message instead of joining with `\n`. Replace each individual message content in the JSON body rather than the concatenated string.
- Add a fallback log warning when a Mask action produces a `ProcessedContent` but the middleware fails to perform any replacement.
- Update `service/security/detector_test.go` expectations (`TestMaskText`, `TestApplyMasking`).

**Affected Files**:
- `service/security/detector.go`
- `middleware/security.go`
- `service/security/detector_test.go`

**Validation**:
- Create a Keyword rule with Action = Mask and keyword `测试`.
- Send a single-message chat request containing `测试`; verify response shows `***`.
- Send a multi-message chat request where one message contains `测试`; verify all occurrences are masked.
- Verify `processed_content` in the audit log contains `***`.

---

### Bug Fix 2: Form and List Labels Are Not Localized

**Priority**: P2

**Symptom**: Security forms and list pages display hardcoded English labels (`Keyword`, `Block`, `Low`, etc.) regardless of the selected UI language.

**Root Causes**:
1. `rule-form-modal.tsx`, `policy-form-modal.tsx`, `group-form-modal.tsx`, and list pages (`rule-page.tsx`, `policy-page.tsx`, `log-page.tsx`) define label mappings with hardcoded English strings.
2. There is no shared frontend constant file; the same mappings are duplicated across components.
3. Backend enum values (`constant/security.go`) and frontend numeric mappings are not synchronized through a single source of truth.

**Fix**:
- Create `web/default/src/features/security/constants.ts` with shared, localizable option definitions:
  - `RULE_TYPES`, `ACTIONS`, `SCOPES`, `STATUSES`, `RISK_LEVELS`
- Replace all hardcoded option arrays and mapping objects (`ruleTypeMap`, `actionMap`, `scopeMap`, `riskLevelMap`) across form modals and list pages with imports from `constants.ts`.
- Wrap all labels with `t(labelKey)` and add the corresponding keys to `en.json`, `zh.json`, and other supported locales.
- Add a safe fallback (e.g., `t('Unknown')`) for `SelectValue` when the current value has no matching option.

**Affected Files**:
- `web/default/src/features/security/constants.ts` (new)
- `web/default/src/features/security/components/rule-form-modal.tsx`
- `web/default/src/features/security/components/policy-form-modal.tsx`
- `web/default/src/features/security/components/group-form-modal.tsx`
- `web/default/src/features/security/pages/rule-page.tsx`
- `web/default/src/features/security/pages/policy-page.tsx`
- `web/default/src/features/security/pages/log-page.tsx`
- `web/default/src/i18n/locales/en.json`
- `web/default/src/i18n/locales/zh.json`

**Validation**:
- Open `/security/rules` in Chinese; verify Type/Action/Status dropdowns and list badges show Chinese labels.
- Switch UI to English; verify labels switch to English.
- Repeat for `/security/policies`, `/security/groups`, and `/security/logs`.

---

### Bug Fix 3: Block Action Continues to Block Non-Matching Content

**Priority**: P1/P2

**Symptom**: After a Keyword rule with Action = Block matches once, subsequent requests without the keyword are still blocked.

**Root Causes** (to be confirmed with logs):
1. **Most likely**: `service/security/cache.go` defines `SecurityCacheExpiration = 5 * time.Minute` but never uses it. `securityRuleCache` and `securityPolicyCache` are plain maps with no TTL, so cached rules never expire automatically. A disabled/deleted rule can remain active indefinitely unless `InvalidateRuleCache()` is explicitly called.
2. `service/security/detector.go` calls `GetUserPolicies(userID)` directly instead of the existing `GetCachedUserPolicies()`, indicating the caching layer is partially unused and inconsistent.
3. The middleware content-joining issue described in Bug Fix 1 may also cause unexpected matching behavior on multi-message requests.

**Fix**:
1. Add structured debug logging to `DetectionEngine.Detect()`, `KeywordDetector.Detect()`, and both middleware `SecurityCheck*` functions to capture per-request policy/rule counts, hit counts, and final actions.
2. Reproduce the false-positive scenario once with logging enabled, then analyze the logs to confirm the root cause.
3. Implement cache TTL:
   - Either wrap cached entries with a timestamp and expire them on read based on `SecurityCacheExpiration`;
   - Or move cache to Redis with a TTL.
4. Decide on a consistent policy caching strategy: either use `GetCachedUserPolicies()` in `detector.go` or remove the unused function.
5. Fix the multi-message content extraction issue from Bug Fix 1 to rule out middleware-side false positives.

**Affected Files**:
- `service/security/cache.go`
- `service/security/detector.go`
- `service/security/engine_keyword.go`
- `middleware/security.go`
- `service/security/engine_keyword_test.go` (add regression test)

**Validation**:
- Create a Keyword rule with keyword `敏感词` and Action = Block.
- Send request 1 with `敏感词` → expect block.
- Send request 2 without `敏感词` → expect normal response.
- Send request 3 with `敏感词` → expect block again.
- Verify server logs show the correct match counts for each request.

---

### Bug Fix Rollback & Security Notes

- **Rollback**: Use `git revert HEAD` (safer) instead of `git reset --hard HEAD~1` + `git push --force`.
- **Password safety**: Do not commit the test server password (`%XSui9i81fm1!ZC8`) to the repository. Remove it from `BUGFIX-PLAN.md` and rotate it if it has already been committed.

## Complexity Tracking

> No constitution violations identified. All changes are incremental enhancements to an existing well-structured module. Bug fixes are localized to the security module and its tests.

| Decision | Why Needed | Simpler Alternative Rejected Because |
|----------|------------|-------------------------------------|
| Page-level Tab navigation instead of sidebar sub-menus | Sidebar already has many entries; Tabs keep URL shareable and match admin mental model | Sidebar sub-menus would add hierarchy depth and conflict with existing flat navigation style |
| URL-synced filters with Zod validation | Refresh persistence and link sharing are explicit requirements; Zod provides type safety | `useState` + `useEffect` manual sync is error-prone and lacks validation |
| TanStack Query layer over existing axios API | Caching, optimistic updates, and polling control are required; existing raw axios lacks these | Continuing raw axios would require reimplementing caching and state management manually |
| Code-split charts with React.lazy | Dashboard charts are heavy; other pages should not load chart libraries | Eager loading would increase bundle size for all Security pages |
