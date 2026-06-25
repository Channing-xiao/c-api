# Quickstart: ai-security 模块验证指南

本指南用于在开发/测试环境中验证 ai-security 模块是否按规格正确工作。

---

## 前置条件

1. 已有一个可运行的 new-api 开发环境（Go + 前端 + 数据库）。
2. 已切换到包含 ai-security 模块的分支。
3. 数据库连接正常。

---

## 1. 安装模块

在项目根目录执行：

```bash
bash custom/ai-security/install.sh
```

预期输出：

```text
[OK] ai-security module directory found
[OK] Database migrated
[OK] Default configs initialized
[OK] Default rules seeded: 20
[OK] Menu entry configured
[OK] Plugin registered
[OK] ai-security version: 1.0.0
```

重复执行应输出：

```text
[OK] ai-security already installed, skipping destructive operations
[OK] New seed rules added: 0
[OK] ai-security version: 1.0.0
```

---

## 2. 验证数据库表

检查数据库中是否存在以下表：

```sql
SHOW TABLES LIKE 'aisec_%';
```

预期结果包含：

```text
aisec_configs
aisec_groups
aisec_rules
aisec_words
aisec_policies
aisec_hit_logs
aisec_daily_stats
aisec_sync_state
aisec_migrations
aisec_audit_logs
```

---

## 3. 验证默认规则

启动后端服务后，使用管理员账号请求：

```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:3000/api/ai-security/rules
```

预期响应：

```json
{
  "success": true,
  "data": {
    "items": [
      { "name": "政治敏感", "group_name": "基础安全策略", "type": 1, ... },
      { "name": "手机号检测", "group_name": "个人隐私信息", "type": 2, ... }
    ],
    "total": >= 1
  }
}
```

---

## 4. 验证请求检测

### 4.1 测试拦截

```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "包含政治敏感词的测试内容"}]
  }'
```

预期结果：
- 如果命中拦截规则，返回 403 或类似拦截响应。
- `/api/ai-security/logs` 中出现一条 action=4（block）的命中日志。

### 4.2 测试脱敏

使用命中脱敏规则的请求内容，验证下游接收到的请求已被脱敏。

---

## 5. 验证独立开关

### 5.1 关闭 ai-security

```bash
curl -X PUT http://localhost:3000/api/ai-security/configs \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ai_security_enabled": false}'
```

发送敏感请求，验证官方 sensitive-words 仍可正常工作。

### 5.2 开启 ai-security 并关闭官方 sensitive-words

在系统设置中关闭官方 sensitive-words，重新开启 ai-security，验证 ai-security 仍可独立检测。

---

## 6. 验证前端页面

启动前端开发服务器：

```bash
cd web/default
bun run dev
```

访问：

```text
http://localhost:5173/ai-security
```

验证：
- 页面可正常加载。
- 菜单路径为 `System Settings > Security > AI Content Security`。
- Dashboard、Groups、Rules、Policies、Logs 五个页面可切换。

---

## 7. 验证重复安装不覆盖数据

1. 在 `/ai-security/rules` 修改一条默认规则的 action。
2. 重新执行 `bash custom/ai-security/install.sh`。
3. 回到 `/ai-security/rules`，验证修改后的规则保持不变。

---

## 8. 验证官方代码升级兼容性

1. 记录主项目中被 ai-security 修改的文件清单（应为 4 个后端挂载点 + 2 个前端挂载点）。
2. 从官方 new-api 仓库同步最新代码。
3. 检查冲突范围，确认冲突只出现在挂载点文件。

---

## 预期结果总结

| 验证项 | 预期结果 |
|---|---|
| install.sh 首次执行 | 成功初始化表、配置、默认规则 |
| install.sh 重复执行 | 不破坏已有数据 |
| 数据库表 | 全部以 `aisec_` 前缀存在 |
| 默认规则 | 首次安装后自动出现 |
| 请求检测 | 命中规则后执行对应动作 |
| 独立开关 | 两个功能可独立启停 |
| 前端入口 | 属于 System Settings > Security 子项 |
| 升级兼容 | 冲突仅出现在必要挂载点 |

---

## 故障排查

- **install.sh 提示模块目录不存在**：检查 `custom/ai-security/` 是否在项目根目录。
- **默认规则未出现**：检查 `aisec_migrations` 表中是否已记录迁移版本。
- **请求未被检测**：检查 `ai_security_enabled` 配置是否为 `true`，以及用户是否绑定了策略。
- **日志未记录**：检查 `aisec_hit_logs` 表是否存在，以及日志 worker 是否正常运行。
