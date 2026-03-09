# Environment Profiles

## Target
Use these profiles to keep behavior consistent across `dev`, `stage`, and `production`.

## Profile Matrix
- dev:
  - `system.environment=dev`
  - seed admin enabled by default
  - permissive CORS (`allow-all`)
  - verbose logs (`debug/info`)
- stage:
  - `system.environment=staging`
  - seed admin disabled (explicitly)
  - whitelist CORS to stage domain
  - logs at `info`
- production:
  - `system.environment=production`
  - seed admin disabled
  - strict CORS whitelist
  - strong secrets for jwt/db/redis/minio/email

## Backend Config Files
- Dev baseline: `backend/configs/config.yaml`
- Production template: `backend/configs/config.prod.yaml`
- Stage/prod env examples:
  - `.env.stage.example`
  - `.env.prod.example`

## Required Secret Replacements (Before Prod)
- `jwt.signing_key`
- `mysql.password`
- `redis.password`
- `file.minio.access_key`
- `file.minio.secret_key`
- `email.secret`

## Seed Admin Controls
Configure in `docker-compose.yml` (backend service env):
- `SEED_ADMIN_ENABLED`
- `SEED_ADMIN_USERNAME`
- `SEED_ADMIN_PASSWORD`
- `SEED_ADMIN_AUTHORITY_ID`
- `SEED_ADMIN_DEFAULT_ROUTER`

Recommended:
- dev/local: enable seed
- stage/prod: disable seed

## Suggested Env Workflow
1. Copy target example:
   - `cp .env.stage.example .env.stage` or `cp .env.prod.example .env.prod`
2. Replace all `ChangeMe`/`ReplaceWithStrong...` placeholders.
3. Pass env file when launching:
   - `docker compose --env-file .env.stage up -d --build`

## Frontend Runtime Targets
- Local dev/proxy in Umi config points to `http://127.0.0.1:8080`
- Docker Nginx proxy maps `/api/*` to `backend:8080`
- For production deployment, keep frontend domain and backend API origin consistent with CORS whitelist.

## Pre-release Checklist
1. Run backend tests: `go test ./...` in `backend`
2. Build frontend: `npm ci --legacy-peer-deps && npm run build` in `front-end`
3. Run smoke: `powershell -ExecutionPolicy Bypass -File .\scripts\smoke_test.ps1`
4. Run health check: `powershell -ExecutionPolicy Bypass -File .\scripts\template_health_check.ps1`
5. Run UI E2E: `npm run e2e` in `front-end`
