# Quickstart — AI Content Security Management v2

## Prerequisites

- Backend running with latest migrations applied (see `contracts/api.md` for migration scripts)
- Administrator account with role `admin`
- Frontend dev server running (`bun run dev` in `web/default/`)

## Validation Scenarios

### Scenario 1: Unified Navigation

**Goal**: Verify Tab navigation works across all Security sub-pages.

**Steps**:
1. Log in as admin and open `/security`
2. Verify Dashboard loads with metric cards
3. Click "Groups" tab → verify URL changes to `/security/groups` and Groups table loads
4. Click "Rules" tab → verify URL changes to `/security/rules`
5. Click "Policies" tab → verify URL changes to `/security/policies`
6. Click "Audit Logs" tab → verify URL changes to `/security/logs`
7. Click "Dashboard" tab → verify return to `/security`
8. Refresh on any sub-page → verify Tab highlights correctly and content loads

**Expected**: All transitions are smooth, no full-page reloads, sidebar "Security" item stays highlighted.

---

### Scenario 2: Legacy System Consolidation

**Goal**: Verify old Sensitive Words is gone and global toggle works.

**Steps**:
1. Navigate to `/system-settings/security`
2. Verify "Sensitive Words" section is **not** present
3. Go to `/security` (Dashboard)
4. Toggle the global security switch to OFF
5. Submit a chat request containing a sensitive keyword
6. Verify the request is **not** intercepted
7. Toggle the global security switch back to ON
8. Submit the same chat request
9. Verify the request **is** intercepted

**Expected**: Old system fully removed; global toggle controls all detection.

---

### Scenario 3: Advanced Log Filtering

**Goal**: Verify time range and model name filtering work.

**Steps**:
1. Go to `/security/logs`
2. Select a date range (e.g., last 7 days)
3. Enter a model name like "gpt-4" in the model filter
4. Click Apply
5. Verify only logs matching both criteria are shown
6. Copy the URL
7. Open the URL in a new tab
8. Verify the same filters are applied automatically

**Expected**: Filters persist in URL; results match criteria; exported CSV contains only filtered rows.

---

### Scenario 4: Rule Testing and Batch Operations

**Goal**: Verify rule testing and batch enable/disable work.

**Steps**:
1. Go to `/security/rules`
2. Find a keyword rule (e.g., "手机号检测")
3. Click "Test" and enter `"我的电话是13800138000"`
4. Verify the test shows "detected: true" with matched text
5. Select 3 rules via checkboxes
6. Click "Batch Disable"
7. Confirm in the dialog
8. Verify all 3 rules show "Disabled" status
9. Select the same 3 rules
10. Click "Batch Enable"
11. Verify all 3 rules show "Enabled" status

**Expected**: Rule test returns accurate detection results; batch operations apply to all selected rules.

---

### Scenario 5: Real-Time Log Monitoring

**Goal**: Verify auto-refresh behavior and new-log notification.

**Steps**:
1. Go to `/security/logs`
2. Enable "Auto Refresh" toggle
3. Trigger a content detection event (submit a blocked request via chat)
4. Wait up to 5 seconds
5. Verify a notification banner appears: "1 new log available. Click to refresh."
6. Do **not** click refresh yet — verify the existing table does **not** change
7. Click the refresh button on the banner
8. Verify the new log appears in the table
9. Switch to another browser tab
10. Trigger another detection event
11. Switch back to the logs tab
12. Verify no new banner appeared (polling was paused)
13. Toggle auto-refresh off and on again
14. Verify polling resumes

**Expected**: New logs are announced via banner, not auto-inserted; polling pauses in inactive tabs.

---

### Scenario 6: Policy Priority

**Goal**: Verify policy priority field and sorting.

**Steps**:
1. Go to `/security/policies`
2. Create a policy with priority `10`
3. Create another policy for the same user with priority `5`
4. Verify the priority `5` policy appears **before** the priority `10` policy
5. Edit the priority `10` policy and change it to `1`
6. Verify it moves to the top of the list

**Expected**: Policies sort by priority ascending (lower number = higher precedence).

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Security pages show blank/403 | User is not admin | Log in with an admin account |
| Filters lost on refresh | URL sync not implemented | Verify `useUrlFilters` hook is applied |
| Logs export ignores filters | Export missing filter params | Verify `start_time`/`end_time`/`model_name` passed to export endpoint |
| Batch operations fail | Backend batch API missing | Verify `POST /api/security/rules/batch-delete` and `/batch-status` exist |
| Policy priority not sorting | Database missing `priority` column | Run Migration 1 from `contracts/api.md` |
