# Research — AI Content Security Group Status Bug Fix

**Date**: 2026-06-12
**Feature**: AI Content Security Management Enhancement (002-ai-content-security-v2)
**Scope**: Bug fix for `/security/groups` status change not persisting

## Unknowns / Clarifications

None. The v1/v2 technology stack and data model are already established in the project.

## Root Cause Analysis

The bug was identified by inspecting the code path for updating a security group:

1. Frontend (`group-form-modal.tsx`) sends `status` in the `PUT /api/security/groups/:id` payload.
2. Backend DTO (`dto/security.go::SecurityGroupRequest`) does **not** declare a `Status` field, so Gin's `ShouldBindJSON` ignores the value.
3. Backend service (`service/security/group.go::UpdateSecurityGroup`) builds an `updates` map that writes `name`, `description`, `sort_order`, and `updated_at`, but never `status`.

As a result, selecting "Disabled" in the edit form has no effect on the database row.

## Decision

- Add `Status int` to `SecurityGroupRequest` with `binding:"oneof=0 1"`.
- Include `"status": req.Status` in the `updates` map inside `UpdateSecurityGroup`.
- Call `InvalidateRuleCache()` after the update so disabled groups are immediately excluded from detection.
- (Recommended) Add a dedicated `PATCH /api/security/groups/:id/status` endpoint and row-level Switch toggle for UX consistency with rules.

## Rationale

This is the smallest surgical fix that resolves the reported symptom. It aligns with the existing rule status toggle pattern and avoids changing the data model or detection engine behavior.

## Alternatives Considered

- **Only fix the form persistence**: Sufficient for the reported bug but leaves group status management less convenient than rule status management.
- **Add batch status endpoint**: Out of scope for the reported bug; can be added later if administrators request bulk group operations.
