# Local Run Guide

## Start
```bash
docker compose up -d --build
```

## PostgreSQL Mode
Set `backend/configs/config.yaml` to `database.driver: postgres`, then start the PostgreSQL override:

```bash
docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build postgres redis minio minio-init backend
```

Restore `database.driver: mysql` to use the default MySQL stack again.

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
