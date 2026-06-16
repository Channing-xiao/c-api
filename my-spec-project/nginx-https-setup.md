# New-API Nginx 反向代理 + HTTPS 配置指南

服务器：`45.251.106.61`  
域名：`ai-api.cncarecc.com`  
邮箱：`channing@goipgroup.com`  
后端服务：`http://127.0.0.1:3000`

---

## 1. 安装 Nginx 与 Certbot

```bash
# 更新软件源
sudo apt update

# 安装 Nginx
sudo apt install -y nginx

# 安装 Certbot 及 Nginx 插件
sudo apt install -y certbot python3-certbot-nginx

# 验证安装
nginx -v
certbot --version
```

---

## 2. 编写 Nginx 反向代理配置

文件路径：`/etc/nginx/sites-available/new-api`

```nginx
# 上游 New-API 服务（Docker 127.0.0.1:3000）
upstream new_api_backend {
    server 127.0.0.1:3000;
    keepalive 32;  # 长连接复用
}

# HTTP 80 端口：强制跳转 HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ai-api.cncarecc.com;

    # Certbot 验证路径（申请/续期证书时使用）
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # 其他请求全部 301 跳转到 HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

# HTTPS 443 端口：反向代理到 New-API
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ai-api.cncarecc.com;

    # SSL 证书路径（由 Certbot 自动管理）
    ssl_certificate /etc/letsencrypt/live/ai-api.cncarecc.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/ai-api.cncarecc.com/privkey.pem;

    # 推荐 SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;

    # 安全响应头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # 客户端请求体大小限制（根据 New-API 需求调整）
    client_max_body_size 50m;

    # 代理到 New-API
    location / {
        proxy_pass http://new_api_backend;
        proxy_http_version 1.1;

        # 关键代理头
        proxy_set_header Host $host;                    # 保留原始 Host 头
        proxy_set_header X-Real-IP $remote_addr;        # 客户端真实 IP
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;  # 转发链
        proxy_set_header X-Forwarded-Proto $scheme;     # 原始协议（https）
        proxy_set_header X-Forwarded-Host $host;        # 原始主机名
        proxy_set_header X-Forwarded-Port $server_port; # 原始端口

        # WebSocket 支持（如 New-API 后续使用流式 WS）
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

启用站点配置：

```bash
# 创建符号链接启用站点
sudo ln -sf /etc/nginx/sites-available/new-api /etc/nginx/sites-enabled/new-api

# 删除默认站点（避免 80 端口冲突）
sudo rm -f /etc/nginx/sites-enabled/default

# 创建 Certbot 验证目录
sudo mkdir -p /var/www/certbot

# 检查 Nginx 语法
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

---

## 3. 申请并配置 Let's Encrypt SSL 证书

### 方式 A：Certbot Nginx 插件自动配置（推荐）

```bash
# 使用 nginx 插件自动申请并配置证书
sudo certbot --nginx -d ai-api.cncarecc.com --non-interactive --agree-tos -m channing@goipgroup.com

# 测试自动续期
sudo certbot renew --dry-run
```

### 方式 B：手动 standalone 申请（如 Nginx 插件不可用）

```bash
# 临时停止 Nginx，使用 standalone 模式申请
sudo systemctl stop nginx
sudo certbot certonly --standalone -d ai-api.cncarecc.com --non-interactive --agree-tos -m channing@goipgroup.com
sudo systemctl start nginx
```

> 申请成功后，确保证书路径与 `/etc/nginx/sites-available/new-api` 中配置的一致。

---

## 4. 防火墙与安全组配置

### 服务器本地防火墙（UFW）

```bash
# 默认拒绝所有入站
sudo ufw default deny incoming
sudo ufw default allow outgoing

# 仅开放 80、443、SSH
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP（用于证书验证和跳转）
sudo ufw allow 443/tcp   # HTTPS

# 启用防火墙
sudo ufw enable

# 查看状态
sudo ufw status verbose
```

### 云服务器安全组建议

| 协议 | 端口 | 来源 | 说明 |
|---|---|---|---|
| TCP | 22 | 你的办公 IP / VPN | SSH 管理 |
| TCP | 80 | 0.0.0.0/0 | HTTP 跳转 + 证书验证 |
| TCP | 443 | 0.0.0.0/0 | HTTPS 访问 |

> **注意**：`3000` 端口不建议对外开放，New-API 仅通过 Nginx 127.0.0.1:3000 反向代理访问。

---

## 5. 验证方法

### 5.1 域名解析检查

```bash
# 确认域名解析到 45.251.106.61
nslookup ai-api.cncarecc.com
dig ai-api.cncarecc.com +short
```

### 5.2 端口监听检查

```bash
# 检查 Nginx 是否在监听 80 和 443
sudo ss -tlnp | grep -E '(:80|:443)'

# 检查 New-API 是否在监听 3000
sudo ss -tlnp | grep :3000
```

### 5.3 Nginx 语法与状态检查

```bash
# 语法检查
sudo nginx -t

# 服务状态
sudo systemctl status nginx

# 查看错误日志
sudo tail -f /var/log/nginx/error.log
```

### 5.4 HTTP 跳转 HTTPS 测试

```bash
# 应返回 301 Moved Permanently
curl -I http://ai-api.cncarecc.com
```

### 5.5 HTTPS 反向代理测试

```bash
# 测试 HTTPS 访问，应返回 New-API 响应
 curl -I https://ai-api.cncarecc.com

# 测试具体 API 路径
 curl https://ai-api.cncarecc.com/api/security/status

# 检查响应头中是否包含 X-Forwarded-Proto
 curl -I -H "X-Forwarded-Proto: https" https://ai-api.cncarecc.com
```

### 5.6 SSL 证书检查

```bash
# 查看证书有效期与颁发者
 echo | openssl s_client -servername ai-api.cncarecc.com -connect ai-api.cncarecc.com:443 2>/dev/null | openssl x509 -noout -issuer -subject -dates

# SSL Labs 在线检测
# https://www.ssllabs.com/ssltest/analyze.html?d=ai-api.cncarecc.com
```

---

## 6. 快速故障排查清单

| 现象 | 排查命令 / 步骤 |
|---|---|
| 域名无法解析 | `dig ai-api.cncarecc.com +short` 确认 A 记录指向 `45.251.106.61` |
| 80 端口无法访问 | `sudo ss -tlnp \| grep :80`、`sudo ufw status`、`curl -I http://ai-api.cncarecc.com` |
| 443 端口无法访问 | 检查安全组、UFW、Nginx `listen 443 ssl` 配置 |
| Nginx 启动失败 | `sudo nginx -t`、`sudo systemctl status nginx`、`sudo tail /var/log/nginx/error.log` |
| HTTPS 证书错误 | `sudo certbot certificates`、`openssl x509 -in /etc/letsencrypt/live/ai-api.cncarecc.com/cert.pem -noout -dates` |
| 反向代理 502/503 | 确认 Docker 中 New-API 监听 `127.0.0.1:3000`：`curl http://127.0.0.1:3000/api/security/status` |
| 真实 IP 获取异常 | 检查 `proxy_set_header X-Forwarded-For` 是否配置，New-API 是否读取该头 |
| 证书续期失败 | `sudo certbot renew --dry-run`、`sudo systemctl status certbot.timer` |

---

## 7. 可选：自动续期定时任务

Certbot 安装后通常已自带 `certbot.timer`，可通过以下命令确认：

```bash
sudo systemctl status certbot.timer
sudo systemctl enable certbot.timer
```

如未启用，可手动添加 cron：

```bash
# 每天凌晨 3 点尝试续期，成功后重载 Nginx
sudo crontab -l | { cat; echo "0 3 * * * /usr/bin/certbot renew --quiet --deploy-hook 'systemctl reload nginx'"; } | sudo crontab -
```
