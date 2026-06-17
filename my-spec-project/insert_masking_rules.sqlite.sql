INSERT INTO security_rules (group_id, name, type, content, extra_config, action, priority, risk_score, status, created_at, updated_at) VALUES
(1, '手机号脱敏', 2, '\b(1[3-9]\d)(\d{4})(\d{4})\b', '{"mask_type":"preserve","preserve_start":3,"preserve_end":4,"mask_char":"*"}', 3, 0, 70, 1, strftime('%s', 'now'), strftime('%s', 'now')),
(1, '银行卡号脱敏', 2, '\b(\d{4})(\d{8,11})(\d{4})\b', '{"mask_type":"preserve","preserve_start":4,"preserve_end":4,"mask_char":"*"}', 3, 0, 80, 1, strftime('%s', 'now'), strftime('%s', 'now')),
(1, '固定电话脱敏', 2, '\b(0\d{2,3}-?)(\d{4})(\d{4})\b', '{"mask_type":"preserve","preserve_start":3,"preserve_end":4,"mask_char":"*"}', 3, 0, 60, 1, strftime('%s', 'now'), strftime('%s', 'now'));
