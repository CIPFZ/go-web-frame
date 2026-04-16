# Database Driver Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add configurable MySQL/PostgreSQL primary database support while keeping MySQL as the default and preserving legacy `mysql:` config compatibility.

**Architecture:** Introduce a unified `database` config object and a single `db.InitDatabase` entry point. Existing service and repository code continues to depend on `*gorm.DB`; only config loading, DB initialization, config files, and deployment docs change.

**Tech Stack:** Go, Viper, GORM, `gorm.io/driver/mysql`, `gorm.io/driver/postgres`, Docker Compose.

---

## File Structure

- Modify `backend/internal/core/config/config.go`: add `Database`, keep legacy `Mysql` for fallback.
- Modify `backend/internal/core/config/mysql.go`: move common log-level behavior behind a small interface and align field names with the new `database.mysql.host` config while preserving legacy `path`.
- Create `backend/internal/core/config/postgres.go`: PostgreSQL config and DSN generation.
- Modify `backend/internal/core/config/loader.go`: normalize database defaults and legacy fallback after Viper unmarshal.
- Modify `backend/internal/core/db/gorm_logger.go`: accept a log-mode provider instead of only `config.MySQL`.
- Modify `backend/internal/core/db/mysql.go`: use new config shape and keep MySQL-specific table options.
- Create `backend/internal/core/db/postgres.go`: PostgreSQL GORM initialization.
- Create `backend/internal/core/db/database.go`: driver switch for `mysql`, `postgres`, `pgsql`.
- Modify `backend/cmd/server/main.go`: call `db.InitDatabase`.
- Modify `backend/cmd/migrate/main.go`: call `db.InitDatabase`.
- Add tests under `backend/internal/core/config` and `backend/internal/core/db`.
- Modify `backend/configs/config.yaml`, `backend/configs/config.prod.yaml`, and `deploy/k3s/base/backend-config.yaml`.
- Add optional PostgreSQL Compose override `docker-compose.pgsql.yml` for verification.
- Update `ENVIRONMENTS.md` and `LOCAL_RUN.md` with database selection notes.

---

### Task 1: Config Types And Normalization

**Files:**
- Modify: `backend/internal/core/config/config.go`
- Modify: `backend/internal/core/config/mysql.go`
- Create: `backend/internal/core/config/postgres.go`
- Modify: `backend/internal/core/config/loader.go`
- Test: `backend/internal/core/config/database_test.go`

- [ ] **Step 1: Write config normalization tests**

Create `backend/internal/core/config/database_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadDatabaseConfig(t *testing.T) {
	path := writeConfig(t, `
system:
  router_prefix: /api/v1
i18n:
  path: locales
database:
  driver: postgres
  postgres:
    host: pgsql
    port: "5432"
    db_name: gva
    username: gva
    password: secret
    ssl_mode: disable
    timezone: Asia/Shanghai
    log_mode: warn
`)

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.Driver != "postgres" {
		t.Fatalf("driver = %q, want postgres", cfg.Database.Driver)
	}
	if cfg.Database.Postgres.Host != "pgsql" {
		t.Fatalf("postgres host = %q, want pgsql", cfg.Database.Postgres.Host)
	}
	if cfg.Database.Postgres.LogMode != "warn" {
		t.Fatalf("postgres log_mode = %q, want warn", cfg.Database.Postgres.LogMode)
	}
}

func TestLoadLegacyMysqlFallback(t *testing.T) {
	path := writeConfig(t, `
system:
  router_prefix: /api/v1
i18n:
  path: locales
mysql:
  path: mysql
  port: "3306"
  db_name: gva
  username: gva
  password: secret
  log_mode: info
`)

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.Driver != "mysql" {
		t.Fatalf("driver = %q, want mysql", cfg.Database.Driver)
	}
	if cfg.Database.MySQL.Host != "mysql" {
		t.Fatalf("mysql host = %q, want mysql", cfg.Database.MySQL.Host)
	}
	if cfg.Database.MySQL.Dbname != "gva" {
		t.Fatalf("mysql db name = %q, want gva", cfg.Database.MySQL.Dbname)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail before implementation**

Run:

```powershell
cd backend
go test ./internal/core/config
```

Expected: FAIL because `Config.Database` and `Postgres` do not exist yet.

- [ ] **Step 3: Add config types and fallback normalization**

Implement:

```go
type Config struct {
	System     System          `mapstructure:"system" json:"system" yaml:"system"`
	Logger     Logger          `mapstructure:"logger" json:"logger" yaml:"logger"`
	I18n       I18n            `mapstructure:"i18n" json:"i18n" yaml:"i18n"`
	JWT        JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Database   Database        `mapstructure:"database" json:"database" yaml:"database"`
	Mysql      MySQL           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mongo      Mongo           `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Redis      Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	File       FileConfig      `mapstructure:"file" json:"file" yaml:"file"`
	Email      Email           `mapstructure:"email" json:"email" yaml:"email"`
	Captcha    Captcha         `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	Cors       CORS            `mapstructure:"cors" json:"cors" yaml:"cors"`
	Observable Observability   `mapstructure:"observable" json:"observable" yaml:"observable"`
	RateLimit  RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
}

type Database struct {
	Driver   string   `mapstructure:"driver" json:"driver" yaml:"driver"`
	MySQL    MySQL    `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Postgres Postgres `mapstructure:"postgres" json:"postgres" yaml:"postgres"`
}
```

Add `backend/internal/core/config/postgres.go`:

```go
package config

import (
	"net/url"
	"strings"

	"gorm.io/gorm/logger"
)

type Postgres struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Dbname       string `mapstructure:"db_name" json:"db_name" yaml:"db_name"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	SSLMode      string `mapstructure:"ssl_mode" json:"ssl_mode" yaml:"ssl_mode"`
	TimeZone     string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	LogMode      string `mapstructure:"log_mode" json:"log_mode" yaml:"log_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	LogZap       bool   `mapstructure:"log_zap" json:"log_zap" yaml:"log_zap"`
}

func (p Postgres) LogLevel() logger.LogLevel {
	switch strings.ToLower(p.LogMode) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

func (p Postgres) Dsn() string {
	values := url.Values{}
	values.Set("host", p.Host)
	values.Set("user", p.Username)
	values.Set("password", p.Password)
	values.Set("dbname", p.Dbname)
	values.Set("port", p.Port)
	values.Set("sslmode", p.SSLMode)
	values.Set("TimeZone", p.TimeZone)
	return values.Encode()
}
```

Update `normalizeConfig`:

```go
func normalizeConfig(cfg *Config, configDir string) {
	cfg.System.RouterPrefix = normalizeRouterPrefix(cfg.System.RouterPrefix)
	cfg.I18n.Path = normalizeI18nPath(cfg.I18n.Path, configDir)
	normalizeDatabase(&cfg.Database, cfg.Mysql)
}
```

Add:

```go
func normalizeDatabase(database *Database, legacy MySQL) {
	database.Driver = normalizeDatabaseDriver(database.Driver)
	if database.Driver == "" {
		database.Driver = "mysql"
	}
	if isEmptyMySQL(database.MySQL) && !isEmptyMySQL(legacy) {
		database.MySQL = legacy
	}
	normalizeMySQLDefaults(&database.MySQL)
	normalizePostgresDefaults(&database.Postgres)
}
```

- [ ] **Step 4: Run config tests**

Run:

```powershell
cd backend
go test ./internal/core/config
```

Expected: PASS.

---

### Task 2: Unified DB Initializer

**Files:**
- Modify: `backend/internal/core/db/gorm_logger.go`
- Modify: `backend/internal/core/db/mysql.go`
- Create: `backend/internal/core/db/postgres.go`
- Create: `backend/internal/core/db/database.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/cmd/migrate/main.go`
- Test: `backend/internal/core/db/database_test.go`

- [ ] **Step 1: Write driver selection tests**

Create `backend/internal/core/db/database_test.go`:

```go
package db

import (
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.uber.org/zap"
)

func TestInitDatabaseRejectsUnsupportedDriver(t *testing.T) {
	_, err := InitDatabase(config.Database{Driver: "oracle"}, zap.NewNop())
	if err == nil {
		t.Fatal("InitDatabase() error = nil, want unsupported driver error")
	}
}

func TestNormalizeDriverAliases(t *testing.T) {
	tests := map[string]string{
		"":         "mysql",
		"mysql":    "mysql",
		"postgres": "postgres",
		"pgsql":    "postgres",
	}
	for in, want := range tests {
		if got := normalizeDriver(in); got != want {
			t.Fatalf("normalizeDriver(%q) = %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run tests and verify they fail before implementation**

Run:

```powershell
cd backend
go test ./internal/core/db
```

Expected: FAIL because `InitDatabase` and `normalizeDriver` do not exist.

- [ ] **Step 3: Implement generic logger config input**

Change `NewZapGormLogger` to accept an interface:

```go
type gormLogConfig interface {
	LogLevel() gormlogger.LogLevel
}

func NewZapGormLogger(zapLogger *zap.Logger, cfg gormLogConfig) *ZapGormLogger {
	return &ZapGormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  cfg.LogLevel(),
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}
```

- [ ] **Step 4: Implement MySQL and PostgreSQL initializers**

Create `backend/internal/core/db/database.go`:

```go
package db

import (
	"fmt"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func InitDatabase(c config.Database, logger *zap.Logger) (*gorm.DB, error) {
	switch normalizeDriver(c.Driver) {
	case "mysql":
		return InitMysql(c.MySQL, logger)
	case "postgres":
		return InitPostgres(c.Postgres, logger)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", c.Driver)
	}
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "mysql":
		return "mysql"
	case "postgres", "postgresql", "pgsql":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}
```

Create `backend/internal/core/db/postgres.go`:

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitPostgres(p config.Postgres, logger *zap.Logger) (*gorm.DB, error) {
	if p.Dbname == "" {
		return nil, fmt.Errorf("postgres dbname is empty")
	}

	gormDB, err := gorm.Open(postgres.Open(p.Dsn()), &gorm.Config{
		Logger: NewZapGormLogger(logger, p),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: p.Singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err = gormDB.Use(otelgorm.NewPlugin(otelgorm.WithDBName(p.Dbname))); err != nil {
		return nil, fmt.Errorf("failed to use otelgorm plugin: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(p.MaxIdleConns)
	sqlDB.SetMaxOpenConns(p.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return gormDB, nil
}
```

- [ ] **Step 5: Wire server and migrate to `InitDatabase`**

Replace:

```go
serviceCtx.DB, err = db.InitMysql(cfg.Mysql, serviceCtx.Logger)
```

with:

```go
serviceCtx.DB, err = db.InitDatabase(cfg.Database, serviceCtx.Logger)
```

Replace in migrate:

```go
gormDB, err := db.InitMysql(cfg.Mysql, logger)
```

with:

```go
gormDB, err := db.InitDatabase(cfg.Database, logger)
```

- [ ] **Step 6: Run DB package tests**

Run:

```powershell
cd backend
go test ./internal/core/db
```

Expected: PASS.

---

### Task 3: Config Files And Deployment

**Files:**
- Modify: `backend/configs/config.yaml`
- Modify: `backend/configs/config.prod.yaml`
- Modify: `deploy/k3s/base/backend-config.yaml`
- Create: `docker-compose.pgsql.yml`
- Modify: `LOCAL_RUN.md`
- Modify: `ENVIRONMENTS.md`

- [ ] **Step 1: Update default backend configs**

Replace top-level `mysql:` with:

```yaml
database:
  driver: mysql
  mysql:
    host: mysql
    port: "3306"
    config: charset=utf8mb4&parseTime=True&loc=Local
    db_name: gva
    username: gva
    password: Gva_Pass!2025
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

Use production secrets placeholders in `config.prod.yaml`.

- [ ] **Step 2: Add PostgreSQL compose override**

Create `docker-compose.pgsql.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_DB: gva
      POSTGRES_USER: gva
      POSTGRES_PASSWORD: "Gva_Pass!2025"
    volumes:
      - pgsql_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gva -d gva"]
      interval: 10s
      timeout: 5s
      retries: 10

  backend:
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  pgsql_data:
```

- [ ] **Step 3: Document MySQL and PostgreSQL run modes**

Add commands:

```powershell
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build postgres backend
```

State that PostgreSQL mode requires setting `database.driver: postgres` in `backend/configs/config.yaml`.

---

### Task 4: Verification

**Files:**
- No code changes unless verification exposes a bug.

- [ ] **Step 1: Run config and DB unit tests**

Run:

```powershell
cd backend
go test ./internal/core/config ./internal/core/db
```

Expected: PASS.

- [ ] **Step 2: Run backend build checks**

Run:

```powershell
cd backend
go build ./cmd/server
go build ./cmd/migrate
```

Expected: both exit 0.

- [ ] **Step 3: Validate Compose syntax**

Run:

```powershell
docker compose config --quiet
docker compose -f docker-compose.yml -f docker-compose.pgsql.yml config --quiet
```

Expected: both exit 0.

- [ ] **Step 4: Verify MySQL default path**

Run:

```powershell
docker compose up -d --build mysql redis minio minio-init backend
docker compose ps
Invoke-RestMethod -Method Get -Uri http://127.0.0.1:8080/health
```

Expected: backend container healthy enough to return `status: ok`.

- [ ] **Step 5: Verify PostgreSQL path**

Change `backend/configs/config.yaml` temporarily to:

```yaml
database:
  driver: postgres
```

Run:

```powershell
docker compose -f docker-compose.yml -f docker-compose.pgsql.yml up -d --build postgres redis minio minio-init backend
docker compose -f docker-compose.yml -f docker-compose.pgsql.yml logs backend --tail=200
Invoke-RestMethod -Method Get -Uri http://127.0.0.1:8080/health
```

Expected: backend initializes PostgreSQL, migration/seed path succeeds, and health returns `status: ok`.

- [ ] **Step 6: Restore default config and commit**

Restore `database.driver: mysql`, then:

```powershell
git status --short
git add backend docker-compose.pgsql.yml LOCAL_RUN.md ENVIRONMENTS.md deploy/k3s/base/backend-config.yaml
git commit -m "feat: support postgres database driver"
```

Expected: only intended implementation files are committed.

