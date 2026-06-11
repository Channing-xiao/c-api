# Feature Specification: AI Content Security Management Enhancement

**Feature Branch**: `002-ai-content-security-v2`

**Created**: 2026-06-11

**Status**: Draft

**Input**: User description: "按照AI内容安全前端设计方案.md和后端API补充清单.md重新开发AI内容安全管理模块，保持第一次开发时的技术和习惯，对前后端进行增强和改进"

---

## Overview

The AI Content Security Management module provides administrators with a unified interface to manage sensitive word detection rules, user policies, and audit logs for AI-generated content. The initial implementation (v1) established core backend APIs and basic frontend pages. This enhancement (v2) addresses navigation gaps, consolidates legacy systems, improves data layer reliability, and enriches backend APIs to support advanced filtering and batch operations.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Security Navigation (Priority: P1)

As an administrator, I want to navigate between Dashboard, Groups, Rules, Policies, and Audit Logs within the Security module using clear in-page navigation, so that I can manage all security settings without manually typing URLs or getting lost.

**Why this priority**: Without sub-navigation, administrators entering the Security Dashboard cannot discover the Groups, Rules, Policies, or Logs pages. This is a fundamental usability blocker that prevents users from accessing 80% of the module's functionality.

**Independent Test**: Can be fully tested by opening `/security` and verifying that Tab navigation allows switching to `/security/groups`, `/security/rules`, `/security/policies`, and `/security/logs` without page refresh or manual URL entry.

**Acceptance Scenarios**:

1. **Given** the administrator is on the Security Dashboard, **When** they click the "Groups" tab, **Then** they are navigated to the Groups management page and the "Groups" tab is highlighted.
2. **Given** the administrator is on any Security sub-page, **When** they look at the sidebar, **Then** the "Security" menu item remains highlighted for all `/security/*` routes.
3. **Given** the administrator visits `/security/rules` directly via URL, **When** the page loads, **Then** the Tab navigation is visible with "Rules" highlighted.

---

### User Story 2 - Consolidated Content Security System (Priority: P1)

As an administrator, I want a single, unified content security system instead of two disconnected systems (legacy Sensitive Words and the new AI Security module), so that I don't experience confusion from conflicting controls and divergent data.

**Why this priority**: The legacy Sensitive Words system (a simple text box in System Settings) and the new AI Security module operate on completely separate data stores. Administrators disabling one system expect content detection to stop, but the other system continues running, creating a false sense of security.

**Independent Test**: Can be fully tested by verifying that disabling content security in any admin interface stops all detection, and that the legacy Sensitive Words entry is no longer visible in System Settings.

**Acceptance Scenarios**:

1. **Given** the administrator disables the global security toggle, **When** a user submits content to the AI model, **Then** no security detection is performed (neither old nor new system).
2. **Given** the administrator opens System Settings, **When** they navigate to the Security section, **Then** the legacy "Sensitive Words" configuration panel is no longer present.
3. **Given** the system was previously configured with legacy Sensitive Words, **When** the administrator opens the Groups page, **Then** they see a notification banner indicating that legacy rules have been migrated to the new system.

---

### User Story 3 - Advanced Audit Log Filtering (Priority: P1)

As an administrator, I want to filter audit logs by time range, model name, user, action type, and risk level, so that I can quickly investigate security incidents and compliance issues.

**Why this priority**: Security investigations require precise filtering. Without time range and model name filtering, administrators must manually scan through hundreds or thousands of log entries, making incident response slow and error-prone.

**Independent Test**: Can be fully tested by opening the Audit Logs page, applying a date range filter and a model name filter, and verifying that only matching logs are displayed.

**Acceptance Scenarios**:

1. **Given** the administrator is on the Audit Logs page, **When** they select a date range of June 1-10 and enter "gpt-4" as the model name, **Then** only logs from that period matching that model are displayed.
2. **Given** the administrator has applied filters, **When** they copy the URL and open it in a new tab, **Then** the same filters are automatically applied.
3. **Given** the administrator has filtered the logs, **When** they click "Export", **Then** the exported file contains only the filtered results.

---

### User Story 4 - Rule Testing and Batch Management (Priority: P2)

As an administrator, I want to test individual rules against sample content and perform batch operations (enable/disable/delete) on multiple rules, so that I can efficiently manage large rule sets without repetitive single-item actions.

**Why this priority**: Managing dozens of rules one-by-one is tedious and time-consuming. Testing rules before deployment prevents false positives and ensures rules behave as expected.

**Independent Test**: Can be fully tested by selecting multiple rules, performing a batch enable operation, and verifying all selected rules change status simultaneously.

**Acceptance Scenarios**:

1. **Given** the administrator is on the Rules page, **When** they select 5 rules and click "Batch Enable", **Then** all 5 rules are enabled with a single operation.
2. **Given** the administrator views a specific rule, **When** they enter test content and click "Test", **Then** they see whether the rule matches the content and what action would be taken.
3. **Given** the administrator toggles a rule's status, **Then** the UI updates immediately (optimistic update) and reflects the actual server state after confirmation.

---

### User Story 5 - Real-Time Audit Log Monitoring (Priority: P2)

As an administrator, I want to optionally enable auto-refresh on the Audit Logs page with a notification when new logs arrive, so that I can monitor security events in real-time without the display jumping or disrupting my review.

**Why this priority**: During security incidents, administrators need to see new events as they happen. Auto-replacing the table interrupts reading and comparison of log entries.

**Independent Test**: Can be fully tested by enabling auto-refresh, triggering a security event, and verifying that a notification appears rather than the table automatically updating.

**Acceptance Scenarios**:

1. **Given** auto-refresh is enabled, **When** new audit logs arrive, **Then** a banner appears saying "N new logs available" rather than automatically replacing the current table.
2. **Given** the administrator switches to another browser tab, **When** auto-refresh is enabled, **Then** polling pauses to reduce unnecessary server load.
3. **Given** the administrator clicks the "Refresh" button on the notification banner, **Then** the table updates with the new logs.

---

### User Story 6 - Policy Priority Management (Priority: P2)

As an administrator, I want to assign numeric priorities to user policies and have them sorted accordingly, so that I can control which policies take precedence when multiple policies apply to the same user.

**Why this priority**: When a user has multiple policies, the order of evaluation matters. Without explicit priority control, policies are evaluated in arbitrary order, leading to unpredictable security outcomes.

**Independent Test**: Can be fully tested by creating two policies for the same user with different priorities and verifying they are displayed in priority order.

**Acceptance Scenarios**:

1. **Given** the administrator creates a policy with priority 1 and another with priority 5, **When** they view the Policies list, **Then** the priority 1 policy appears before the priority 5 policy.
2. **Given** the administrator edits a policy, **When** they change its priority, **Then** the list reorders to reflect the new priority.

---

### Edge Cases

- What happens when an administrator tries to delete a group that contains active rules? The system should warn about cascading deletions.
- How does the system handle concurrent rule status updates by multiple administrators? Optimistic updates should gracefully handle conflicts.
- What happens when no audit logs exist for the selected filters? The system should show an appropriate empty state.
- How does the system behave when the real-time log refresh runs for extended periods? Polling should automatically pause after a maximum duration to prevent resource exhaustion.
- What happens when an administrator without permission accesses Security pages? The system should display a forbidden state rather than erroring.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide unified Tab navigation across all Security sub-pages (Dashboard, Groups, Rules, Policies, Audit Logs).
- **FR-002**: The system MUST remove the legacy Sensitive Words configuration panel from System Settings.
- **FR-003**: The system MUST synchronize all filter states to URL query parameters so that refreshing the page or sharing the URL preserves the current view.
- **FR-004**: The system MUST support filtering audit logs by time range (start and end timestamps).
- **FR-005**: The system MUST support filtering audit logs by model name using partial text matching.
- **FR-006**: The system MUST support testing individual rules against user-provided sample content and returning the detection result.
- **FR-007**: The system MUST support batch enabling, disabling, and deleting of rules via single operations.
- **FR-008**: The system MUST support toggling individual rule status without requiring a full rule update.
- **FR-009**: The system MUST support assigning numeric priorities to user policies and sorting policies by priority.
- **FR-010**: The system MUST support filtering security groups by name using partial text matching.
- **FR-011**: The system MUST provide a global security status toggle on the Dashboard that controls whether content detection is active.
- **FR-012**: The system MUST display a migration status banner when legacy Sensitive Words data has been migrated to the new system.
- **FR-013**: The system MUST provide real-time audit log monitoring with an auto-refresh toggle that displays a notification banner when new logs arrive rather than auto-replacing the table.
- **FR-014**: The system MUST pause real-time polling when the browser tab is not active.
- **FR-015**: The system MUST display the total number of today's interceptions on the Dashboard.
- **FR-016**: The system MUST show appropriate empty states for initial use, no data, filtered empty results, and forbidden access.
- **FR-017**: The system MUST display a confirmation dialog before destructive operations (delete group, delete rule, delete policy, batch delete).
- **FR-018**: The system MUST mask sensitive information (phone numbers, ID numbers, emails) when displaying audit log match details.
- **FR-019**: The system MUST render all user-generated content in audit logs as plain text without HTML interpretation to prevent XSS.
- **FR-020**: The system MUST restrict all Security management operations to administrators.

### Key Entities *(include if feature involves data)*

- **Security Group**: A logical collection of detection rules organized hierarchically. Groups can be nested up to 5 levels deep.
- **Security Rule**: A detection pattern (keyword, regex, NER, or AI-based) belonging to a group. Rules have a type, content, action, priority, risk score, and status.
- **User Policy**: A binding between a user and one or more security groups, defining scope (request/response/both), default action, custom response message, whitelist IPs, and priority.
- **Audit Log**: A record of content detection events including request metadata, matched rules, action taken, risk score/level, and match details.
- **Migration Status**: Tracks whether legacy Sensitive Words have been migrated to the new group/rule structure.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can navigate from the Security Dashboard to any sub-page (Groups, Rules, Policies, Logs) in a single click.
- **SC-002**: Filtered views in Groups, Rules, Policies, and Logs pages are fully reproducible via URL sharing and survive page refresh.
- **SC-003**: Audit log queries with time range and model name filters return results in under 500 milliseconds for datasets up to 1 million records.
- **SC-004**: Administrators can test any rule against sample content and see results in under 1 second.
- **SC-005**: Batch operations on up to 50 rules complete in under 2 seconds.
- **SC-006**: 100% of audit log views correctly mask sensitive data (phone numbers, ID numbers, bank cards, emails) before display.
- **SC-007**: Real-time log monitoring does not cause server load increases when the administrator's browser tab is inactive.
- **SC-008**: The legacy Sensitive Words system is no longer accessible or functional after migration; all content detection is controlled through the unified Security module.

---

## Assumptions

- The target users are system administrators with elevated privileges. Regular users do not access the Security management interface.
- The existing backend APIs for Groups, Rules, Policies, Logs, and Dashboard provide a stable foundation for enhancement.
- The legacy Sensitive Words data can be programmatically migrated to the new group/rule structure during system initialization or upgrade.
- Browser support targets modern evergreen browsers (Chrome, Firefox, Safari, Edge) with ES2022+ compatibility.
- Mobile responsiveness is a secondary concern; the primary use case is desktop administration.
- All time-based filtering uses Unix timestamps for backend compatibility.
