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

## Complexity Tracking

> No constitution violations identified. All changes are incremental enhancements to an existing well-structured module.

| Decision | Why Needed | Simpler Alternative Rejected Because |
|----------|------------|-------------------------------------|
| Page-level Tab navigation instead of sidebar sub-menus | Sidebar already has many entries; Tabs keep URL shareable and match admin mental model | Sidebar sub-menus would add hierarchy depth and conflict with existing flat navigation style |
| URL-synced filters with Zod validation | Refresh persistence and link sharing are explicit requirements; Zod provides type safety | `useState` + `useEffect` manual sync is error-prone and lacks validation |
| TanStack Query layer over existing axios API | Caching, optimistic updates, and polling control are required; existing raw axios lacks these | Continuing raw axios would require reimplementing caching and state management manually |
| Code-split charts with React.lazy | Dashboard charts are heavy; other pages should not load chart libraries | Eager loading would increase bundle size for all Security pages |
