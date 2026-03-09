# Local Run Guide

## Start
```bash
docker compose up -d --build
```

## Default Access
- Frontend: `http://localhost`
- Backend health: `http://localhost:8080/health`
- MinIO console: `http://localhost:9001`
- Seed admin:
  - Username: `admin`
  - Password: `Admin@123456`

## Smoke Test
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke_test.ps1
```

## Full Template Health Check
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\template_health_check.ps1
```

## Frontend E2E (Playwright)
```bash
cd front-end
npm ci --legacy-peer-deps
npx playwright install chromium
npm run e2e
```
Includes:
- login -> workplace -> notice visible
- admin notice list visible -> user workplace visible

## Useful Commands
```bash
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
docker compose restart backend frontend
```

## Quality Docs
- Template quality baseline: `TEMPLATE_QUALITY.md`
- Environment profiles: `ENVIRONMENTS.md`
- CI pipeline: `.github/workflows/ci.yml`

## Seed Admin Environment Variables
Set on `backend` service in `docker-compose.yml`:
- `SEED_ADMIN_ENABLED`
- `SEED_ADMIN_USERNAME`
- `SEED_ADMIN_PASSWORD`
- `SEED_ADMIN_AUTHORITY_ID`
- `SEED_ADMIN_DEFAULT_ROUTER`

Default behavior:
- Enabled only for `dev/local/development` unless overridden by `SEED_ADMIN_ENABLED`.
- In `production`, default is disabled.
