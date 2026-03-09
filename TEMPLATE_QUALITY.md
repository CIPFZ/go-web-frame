# Template Quality Guide

## Positioning
This repository is now a **production-oriented starter template** for Go + Gin + Umi.
It includes backend/frontend scaffolding, RBAC, menu system, operation logs, and directed notices.

## What Is Enforced
- Backend compile + tests: `go test ./...`
- Frontend build: `npm run build`
- Frontend UI E2E smoke: `npm run e2e` (Playwright)
- CI checks on push and PR (`.github/workflows/ci.yml`)
- Health smoke for local runtime (`scripts/template_health_check.ps1`)

## Local Quality Gate
1. Start services:
   - `docker compose up -d --build`
2. Run quick smoke:
   - `powershell -ExecutionPolicy Bypass -File .\scripts\smoke_test.ps1`
3. Run full template health check (includes notice flow):
   - `powershell -ExecutionPolicy Bypass -File .\scripts\template_health_check.ps1`
4. Run UI E2E:
   - `cd front-end && npm ci --legacy-peer-deps && npx playwright install chromium && npm run e2e`

## Encoding/Charset Baseline
- Database and tables use `utf8mb4`.
- Backend MySQL DSN now enforces defaults if missing:
  - `charset=utf8mb4`
  - `parseTime=True`
  - `loc=Local`
- Avoid using non-UTF8 terminals when validating Chinese content.

## Recommended Next Hardening
- Add API integration tests for critical endpoints (login/menu/state/notice).
- Add DB backup/restore scripts to `scripts/`.
- Add secret rotation runbook and rollback SOP.
