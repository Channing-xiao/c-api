# Security 模块后端 API 补充清单

> **版本**：v1.1
> **日期**：2026-06-11
> **关联文档**：AI内容安全前端设计方案.md（v2.2）
> **说明**：本清单基于前端 v2.2 方案与当前后端代码的逐项对比审计结果编制。**绿色章节为已修复项**。

---

## ✅ 已修复项（架构层）

### FIX-1. 新旧系统开关统一（方案 A）

**问题**：旧系统的 `CheckSensitiveEnabled` 开关和新模块的 `SECURITY_ENABLED` 环境变量是两套独立开关，互不感知。管理员在旧页面关闭后，新模块仍在运行。

**修复方式**：修改 `middleware/security.go`，让 `SecurityCheck()` 和 `SecurityCheckResponse()` 同时检查两个开关：

```go
// middleware/security.go
import "github.com/QuantumNous/new-api/setting"

func SecurityCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !security.IsSecurityEnabled() || !setting.CheckSensitiveEnabled {
            c.Next()
            return
        }
        // ...
    }
}

func SecurityCheckResponse() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !security.IsSecurityEnabled() || !setting.CheckSensitiveEnabled {
            c.Next()
            return
        }
        // ...
    }
}
```

**结果**：任一开关关闭，内容安全检测即停止。Dashboard 页面的全局开关与旧系统 `System Settings → Sensitive Words` 页面实现联动。

**验证方式**：
- 关闭旧系统页面的「Enable filtering」开关 → 新模块检测同步停止
- Dashboard 开关状态与 `/api/security/status` 返回的 `enabled` 一致

---

## 清单总览

| 优先级 | 数量 | 类别 |
|--------|------|------|
| P0（阻塞上线） | 2 | 缺失 API、字段不匹配 |
| P1（影响核心体验） | 3 | 核心功能 API 缺失 |
| P2（性能优化） | 2 | 批量操作、单字段更新 |
| P3（体验增强） | 3 | 搜索、趋势数据、Excel 导出 |

---

## P0 阻塞项（不解决则前端功能无法运行）

### P0-1. 日志查询接口缺少时间范围和模型名称筛选

**现状**：
```go
// controller/security.go:251-266
userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
action, _ := strconv.Atoi(c.DefaultQuery("action", "0"))
riskLevel, _ := strconv.Atoi(c.DefaultQuery("risk_level", "0"))
contentType, _ := strconv.Atoi(c.DefaultQuery("content_type", "0"))
```

**缺失参数**：
- `start_time` / `end_time`：时间范围筛选（Unix 时间戳）
- `model_name`：模型名称模糊搜索

**注意**：`dto.SecurityHitLogQuery` 中已定义 `StartTime` 和 `EndTime` 字段，但 Controller 未读取。

**前端影响**：Logs 页面的时间范围选择器和模型搜索框无法工作。

**建议修改**：
```go
// controller/security.go GetSecurityLogs
startTime, _ := strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64)
endTime, _ := strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64)
modelName := c.DefaultQuery("model_name", "")

// 传入 service
logs, count, err := security.GetSecurityLogs(security.SecurityLogQueryParams{
    Page:        page,
    PageSize:    pageSize,
    UserID:      userID,
    Action:      action,
    RiskLevel:   riskLevel,
    ContentType: contentType,
    StartTime:   startTime,
    EndTime:     endTime,
    ModelName:   modelName,
})
```

**对应 Service 修改**（`service/security/hitlog.go`）：
```go
type SecurityLogQueryParams struct {
    Page        int
    PageSize    int
    UserID      int
    Action      int
    RiskLevel   int
    ContentType int
    StartTime   int64  // 新增
    EndTime     int64  // 新增
    ModelName   string // 新增
}
```

在 `GetSecurityLogs` 中增加：
```go
if params.StartTime > 0 {
    db = db.Where("security_hit_logs.created_at >= ?", params.StartTime)
}
if params.EndTime > 0 {
    db = db.Where("security_hit_logs.created_at <= ?", params.EndTime)
}
if params.ModelName != "" {
    db = db.Where("security_hit_logs.model_name LIKE ?", "%"+params.ModelName+"%")
}
```

导出接口 `GetSecurityLogsForExport` / `ExportSecurityLogParams` 同样需要同步增加这三个字段。

---

### P0-2. 策略表缺少 `priority` 字段

**现状**：
```go
// model/security.go SecurityUserPolicy
ID, UserID, GroupID, Scope, DefaultAction, CustomResponse, WhitelistIPs, Status, CreatedAt, UpdatedAt
```

**缺失**：无 `priority` 字段。

**前端影响**：Policies 页面的优先级调整（1-100，数字越小优先级越高）功能完全无法实现。表格也无法按优先级排序。

**建议修改**：

1. **Model 增加字段**：
```go
// model/security.go
type SecurityUserPolicy struct {
    // ... 现有字段
    Priority    int    `json:"priority" gorm:"column:priority;type:int;default:0"`
}
```

2. **DTO 增加字段**：
```go
// dto/security.go SecurityPolicyRequest / SecurityPolicyResponse
type SecurityPolicyRequest struct {
    // ... 现有字段
    Priority    int    `json:"priority"`
}
```

3. **Service 排序**：`GetSecurityPolicies` 按 `priority ASC` 排序。

4. **数据库迁移**：新增 `priority` 列，默认值为 0。

**替代方案**：如果近期不增加字段，前端文档将移除策略优先级功能，待后续迭代。

---

## P1 核心功能缺失（影响主要用户体验）

### P1-1. 规则测试 API

**现状**：前端设计 `POST /api/security/rules/:id/test`，用于测试单条规则是否命中给定文本。后端无此 API。

**建议新增**：

```go
// controller/security.go
func TestSecurityRule(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    var req struct {
        Content string `json:"content" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }

    rule, err := security.GetSecurityRuleById(id)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": "规则不存在"})
        return
    }

    // 调用检测引擎的单规则检测（或构造临时检测请求）
    result, err := security.TestRule(rule, req.Content)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

**返回结构建议**：
```json
{
  "detected": true,
  "action": 4,
  "risk_score": 80,
  "risk_level": 3,
  "matched_text": "13800138000",
  "position": [0, 11]
}
```

**路由注册**：
```go
securityRuleRoute.POST("/:id/test", controller.TestSecurityRule)
```

**替代方案**：前端可临时使用 `POST /api/security/check/request` 测试，但该接口需要 user_id 且检测的是全量规则，不是单条规则。

---

### P1-2. 规则状态单独切换 API

**现状**：前端需要实现规则的启用/停用开关（乐观更新）。但后端无单独的状态切换 API，必须调用完整 `PUT /api/security/rules/:id`。

**问题**：
- 频繁开关时，需要传输完整的规则字段
- 并发修改可能覆盖其他字段
- 批量启用/停用需要多次完整更新

**建议新增**：

```go
// controller/security.go
func UpdateSecurityRuleStatus(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    var req struct {
        Status int `json:"status" binding:"oneof=0 1"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }

    if err := security.UpdateSecurityRuleStatus(id, req.Status); err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "message": "状态更新成功"})
}
```

**Service 层**：
```go
func UpdateSecurityRuleStatus(id int64, status int) error {
    return model.DB.Model(&model.SecurityRule{}).Where("id = ?", id).Update("status", status).Error
}
```

**路由注册**：
```go
securityRuleRoute.PATCH("/:id/status", controller.UpdateSecurityRuleStatus)
```

---

### P1-3. 旧系统数据迁移状态查询 API

**现状**：前端设计展示旧系统 `SensitiveWords` 是否已迁移的 Banner。后端无此查询接口，也无迁移逻辑。

**建议新增**：

```go
// controller/security.go
func GetSecurityMigrationStatus(c *gin.Context) {
    status, err := security.GetMigrationStatus()
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
}
```

**返回结构**：
```go
type MigrationStatusResponse struct {
    Migrated      bool   `json:"migrated"`       // 是否已完成迁移
    MigratedCount int    `json:"migrated_count"` // 迁移的规则数量
    MigratedAt    int64  `json:"migrated_at"`    // 迁移时间戳
    SourceGroupID int64  `json:"source_group_id"` // 「系统迁移」分组 ID
}
```

**路由注册**：
```go
securityRoute.GET("/migration-status", middleware.AdminAuth(), controller.GetSecurityMigrationStatus)
```

**说明**：此 API 可与后端的数据迁移脚本配合使用。迁移脚本在初始化时检查 `setting.SensitiveWords`，非空则创建「系统迁移」分组并转换规则。

---

## P2 性能优化项

### P2-1. 批量操作 API（规则）

**现状**：前端设计批量删除/启用/停用规则。后端只有单个操作 API。

**建议新增**：

```go
// 批量删除
POST /api/security/rules/batch-delete
Body: { "ids": [1, 2, 3] }

// 批量更新状态
PATCH /api/security/rules/batch-status
Body: { "ids": [1, 2, 3], "status": 1 }
```

**Service 层**：
```go
func BatchDeleteSecurityRules(ids []int64) error
func BatchUpdateSecurityRuleStatus(ids []int64, status int) error
```

---

### P2-2. Dashboard 增加「今日拦截数」和「安全趋势」数据

**现状**：Dashboard 返回 `today_detections`（今日检测数），但无「今日拦截数」和「近 7/30 天趋势」数据。

**建议增强 `GetSecurityDashboard`**：

1. 增加今日拦截数：
```go
var todayInterceptions int64
model.DB.Model(&model.SecurityHitLog{}).
    Where("created_at >= ? AND action = ?", todayStart, constant.SecurityActionBlock).
    Count(&todayInterceptions)
response.Summary.TodayInterceptions = int(todayInterceptions)
```

2. 增加趋势数据（按天聚合）：
```go
type TrendData struct {
    Date          string `json:"date"`
    Detections    int    `json:"detections"`
    Interceptions int    `json:"interceptions"`
    Alerts        int    `json:"alerts"`
}
```

**或替代方案**：前端自行用 `start_time`/`end_time` 请求多天的 Dashboard 数据并聚合（不推荐，效率低）。

---

## P3 体验增强项

### P3-1. 分组名称模糊搜索

**现状**：`GetSecurityGroups` 无 `name` 搜索参数。

**建议修改**：
```go
name := c.DefaultQuery("name", "")
// 传入 service
if name != "" {
    db = db.Where("name LIKE ?", "%"+name+"%")
}
```

---

### P3-2. 分组复制后名称统一

**现状**：后端 `CopySecurityGroup` 生成的名称是 `"_copy"` 后缀。

**建议修改**：`group.go:160` 改为中文后缀：
```go
Name: srcGroup.Name + "(副本)",
```

---

### P3-3. 真正的 Excel 导出

**现状**：`format=excel` 实际返回的是 CSV + `application/vnd.ms-excel` MIME 类型。

**建议**：引入 `github.com/xuri/excelize/v2` 实现真正的 `.xlsx` 导出，或前端将按钮文案改为「导出 CSV」。

---

## 数据库迁移脚本

### Migration 1: 策略表增加 priority 字段

```sql
-- MySQL
ALTER TABLE security_user_policies ADD COLUMN priority INT NOT NULL DEFAULT 0 AFTER whitelist_ips;
CREATE INDEX idx_security_policy_priority ON security_user_policies(priority);

-- PostgreSQL
ALTER TABLE security_user_policies ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_security_policy_priority ON security_user_policies(priority);

-- SQLite
ALTER TABLE security_user_policies ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
```

### Migration 2: 命中日志表索引优化（新增时间范围查询）

```sql
-- MySQL / PostgreSQL
CREATE INDEX idx_security_hit_created_at ON security_hit_logs(created_at);

-- SQLite
CREATE INDEX idx_security_hit_created_at ON security_hit_logs(created_at);
```

---

## 实施建议

### 最短路径（让前端方案可运行）

只需完成 **P0-1**（日志筛选增强）+ **P1-2**（规则状态切换）+ **P3-2**（复制名称中文），前端 v2.1 方案的核心功能即可全部运行。

### 完整路径（达到前端设计的完整体验）

按优先级顺序：P0 → P1 → P2 → P3，预计后端开发工作量 **3-5 天**。

---

## 前后端接口对照速查表

| 前端设计 | 后端现状 | 状态 |
|----------|----------|------|
| `GET /api/security/groups?status=&parent_id=&name=` | 有 `status`、`parent_id`，缺 `name` | ⚠️ 需增强 |
| `POST /api/security/groups/:id/copy` | ✅ 已实现 | ✅ |
| `GET /api/security/rules?group_id=&type=&status=` | ✅ 已实现 | ✅ |
| `POST /api/security/rules/:id/test` | ❌ 未实现 | 🔴 需新增 |
| `PATCH /api/security/rules/:id/status` | ❌ 未实现 | 🟡 建议新增 |
| `GET /api/security/policies` | ✅ 已实现 | ✅ |
| `GET /api/security/logs?user_id=&action=&risk_level=&content_type=` | ✅ 已实现 | ✅ |
| `GET /api/security/logs?start_time=&end_time=&model_name=` | ❌ 未实现 | 🔴 需增强 |
| `GET /api/security/dashboard?start_time=&end_time=` | ✅ 已实现 | ⚠️ 字段需对齐 |
| `GET /api/security/migration-status` | ❌ 未实现 | 🟡 建议新增 |
