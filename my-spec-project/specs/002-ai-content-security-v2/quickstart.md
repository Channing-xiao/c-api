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

## Bug Fix Validation Scenarios

### BF-1: Mask Action Replaces Keywords with `***`

**Goal**: Verify Mask action works for both single-message and multi-message requests.

**Prerequisites**:
- A Keyword rule exists with keyword `测试` and Action = Mask.
- A valid API token is available.

**Steps**:
1. Send a single-message request:
   ```bash
   curl -X POST http://45.251.106.61:3000/v1/chat/completions \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"这是一个测试"}]}'
   ```
2. Verify the response content contains `***` instead of `测试`.
3. Send a multi-message request:
   ```bash
   curl -X POST http://45.251.106.61:3000/v1/chat/completions \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"你好"},{"role":"user","content":"这是一个测试"}]}'
   ```
4. Verify all occurrences of `测试` are replaced with `***`.
5. Go to `/security/logs` and verify the latest log's `processed_content` contains `***`.

**Expected**: Masking applies correctly in both single and multi-message scenarios, and audit logs reflect the masked content.

---

### BF-2: Form and List Labels Are Localized

**Goal**: Verify Type, Action, Scope, Status, and Risk Level labels switch with the UI language.

**Steps**:
1. Open `/security/rules` with UI language set to Chinese.
2. Click "新建规则" and verify Type/Action/Status dropdowns show Chinese labels (e.g., "关键词匹配", "脱敏", "启用").
3. Save the rule and verify the list page shows Chinese badges.
4. Switch UI language to English.
5. Reopen `/security/rules` and verify labels are now in English.
6. Repeat steps 1-5 for `/security/policies`, `/security/groups`, and `/security/logs`.

**Expected**: All labels are localized and consistent across forms and lists.

---

### BF-3: Block Action Does Not False-Positive

**Goal**: Verify a Block rule only blocks requests containing the keyword.

**Prerequisites**:
- A Keyword rule exists with keyword `敏感词` and Action = Block.
- Structured debug logging is enabled for the security module.

**Steps**:
1. Send request A containing `敏感词`:
   ```bash
   curl -X POST http://45.251.106.61:3000/v1/chat/completions \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"这句话包含敏感词"}]}'
   ```
2. Verify response is blocked (HTTP 403 or success=false with block message).
3. Send request B **without** `敏感词`:
   ```bash
   curl -X POST http://45.251.106.61:3000/v1/chat/completions \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"这句话很正常"}]}'
   ```
4. Verify response is **not** blocked and proceeds normally.
5. Send request C containing `敏感词` again and verify it is blocked.
6. Check server logs for each request; confirm request B has `hits=0` and `action=pass`.

**Expected**: Only requests containing the keyword are blocked; logs confirm zero hits for clean requests.

---

### BF-4: Group Status Can Be Disabled and Re-Enabled

**Goal**: Verify that changing a security group's status to "Disabled" persists and stops detection for rules in that group.

**Prerequisites**:
- A security group exists with at least one enabled Keyword rule that matches a known keyword.

**Steps**:
1. Go to `/security/groups`.
2. Find the group containing the test rule and click "Edit".
3. Change **Status** from "Enabled" to "Disabled" and save.
4. Verify the table row now shows "Disabled".
5. Refresh the page and verify the group still shows "Disabled".
6. Submit a chat request containing the keyword:
   ```bash
   curl -X POST http://45.251.106.61:3000/v1/chat/completions \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"这句话包含测试关键词"}]}'
   ```
7. Verify the request is **not** intercepted (the disabled group's rules are skipped).
8. Re-open the group, change Status back to "Enabled", and save.
9. Submit the same request again and verify it **is** intercepted.

**Expected**: Group status changes persist across refresh and correctly control whether the group's rules participate in detection.

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Security pages show blank/403 | User is not admin | Log in with an admin account |
| Filters lost on refresh | URL sync not implemented | Verify `useUrlFilters` hook is applied |
| Logs export ignores filters | Export missing filter params | Verify `start_time`/`end_time`/`model_name` passed to export endpoint |
| Batch operations fail | Backend batch API missing | Verify `POST /api/security/rules/batch-delete` and `/batch-status` exist |
| Policy priority not sorting | Database missing `priority` column | Run Migration 1 from `contracts/api.md` |
| Mask does nothing on multi-message requests | Middleware replaces joined `\n` string that does not exist in raw JSON | Verify per-message replacement is implemented |
| Block false-positives after first hit | Cache has no TTL; disabled rule still cached | Verify cache TTL implementation and check debug logs |
| Form labels stay in English | i18n keys missing or hardcoded labels | Verify `constants.ts` labels use `t()` and keys exist in locale files |
| Group status reverts after save | `SecurityGroupRequest` missing `Status` field or `UpdateSecurityGroup` not updating status | Verify `dto/security.go` has `Status` and `service/security/group.go` includes it in updates |
| Disabled group's rules still intercept | Rule cache not invalidated on group status change | Verify `InvalidateRuleCache()` is called after group update/toggle |
