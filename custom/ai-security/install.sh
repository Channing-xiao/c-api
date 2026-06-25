#!/bin/bash
set -e

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../.." && pwd)"

VERSION=$(grep -o '"version": *"[^"]*"' "$MODULE_DIR/version.json" | head -1 | sed 's/.*: *"\([^"]*\)".*/\1/')

echo "[INFO] ai-security module directory: $MODULE_DIR"
echo "[INFO] ai-security version: $VERSION"

# 1. Check module directory
if [ ! -d "$MODULE_DIR/api" ] || [ ! -d "$MODULE_DIR/service" ] || [ ! -d "$MODULE_DIR/web" ]; then
    echo "[ERROR] ai-security module structure is incomplete"
    exit 1
fi
echo "[OK] ai-security module directory found"

# 2. Run database migrations (performed by go run migration or at app startup)
echo "[INFO] Database migrations will be applied at application startup via ai_security.Init()"

# 3. Seed default configs and rules
echo "[INFO] Default configs and rules will be seeded at application startup via ai_security.Init()"

# 4. Menu entry configuration is handled by frontend code and router registration
echo "[INFO] Menu entry configured by frontend route registration"

# 5. Plugin status is implicitly registered by module existence
echo "[OK] ai-security plugin registered"

echo "[OK] Install script completed. Please restart new-api to apply migrations and seed data."
