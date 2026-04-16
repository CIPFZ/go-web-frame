# Database Driver Support Design

## Goal

Support either MySQL or PostgreSQL as the primary relational database for the backend. The application should select one database driver at startup through configuration, while keeping repository and service code on the existing `*gorm.DB` abstraction.

## Recommended Approach

Use a unified `database` configuration block:

```yaml
database:
  driver: mysql
  mysql:
    host: mysql
    port: "3306"
    db_name: gva
    username: gva
    password: Gva_Pass!2025
    config: charset=utf8mb4&parseTime=True&loc=Local
    max_idle_conns: 10
    max_open_conns: 100
    log_mode: info
    log_zap: false
  postgres:
    host: postgres
    port: "5432"
    db_name: gva
    username: gva
    password: Gva_Pass!2025
    ssl_mode: disable
    timezone: Asia/Shanghai
    max_idle_conns: 10
    max_open_conns: 100
    log_mode: info
    log_zap: false
```

`database.driver` accepts `mysql`, `postgres`, and `pgsql` as aliases, with `mysql` as the default when omitted.

## Compatibility

Keep short-term compatibility with the existing top-level `mysql:` block. If `database:` is missing or empty but `mysql:` is present, the loader should populate `database.driver=mysql` and reuse the old MySQL settings. This avoids breaking existing local, Docker, and production configs during the transition.

The old `system.db_type` field is not used today. It should not become the new source of truth. It can remain for now if removing it would cause unrelated churn, but database selection should come from `database.driver`.

## Backend Changes

Add configuration types:

- `Database`
- `MySQL` under `Database`
- `Postgres`

Add database initialization:

- `db.InitDatabase(config.Database, *zap.Logger) (*gorm.DB, error)`
- existing MySQL initialization continues to use the MySQL GORM driver
- new PostgreSQL initialization uses `gorm.io/driver/postgres`

`ServiceContext.DB` remains unchanged. Module repositories and services should not branch on database type unless a real dialect-specific problem appears.

## Config And Deployment

Update:

- `backend/configs/config.yaml`
- `backend/configs/config.prod.yaml`
- `deploy/k3s/base/backend-config.yaml`
- environment documentation where database configuration is described

Docker Compose can continue to start MySQL by default. PostgreSQL support should be documented as an alternate deployment choice rather than replacing MySQL in the default local stack.

## Risks

GORM hides most dialect differences, but these areas need verification:

- auto migration column types and indexes
- seed data idempotency
- JSON or text columns
- timestamp defaults
- table options such as MySQL `ENGINE=...`, which should not be applied to PostgreSQL

## Verification

Minimum verification for this branch:

- `docker compose config --quiet`
- `cd backend && go test ./...` when local disk/network state allows dependency compilation
- targeted build for `./cmd/server` and `./cmd/migrate`
- config load tests for new `database` config and legacy `mysql` fallback

