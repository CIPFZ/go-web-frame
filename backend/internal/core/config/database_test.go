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

func TestLoadSQLite3DatabaseConfig(t *testing.T) {
	path := writeConfig(t, `
system:
  router_prefix: /api/v1
i18n:
  path: locales
database:
  driver: sqlite3
  sqlite:
    path: data/app.db
    wal: true
    busy_timeout_ms: 9000
    foreign_keys: true
`)

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.Driver != "sqlite3" {
		t.Fatalf("driver = %q, want sqlite3", cfg.Database.Driver)
	}
	wantPath := filepath.Join(filepath.Dir(path), "data", "app.db")
	if cfg.Database.SQLite.Path != wantPath {
		t.Fatalf("sqlite path = %q, want %q", cfg.Database.SQLite.Path, wantPath)
	}
	if !cfg.Database.SQLite.WAL {
		t.Fatal("sqlite wal = false, want true")
	}
	if cfg.Database.SQLite.BusyTimeoutMS != 9000 {
		t.Fatalf("busy_timeout_ms = %d, want 9000", cfg.Database.SQLite.BusyTimeoutMS)
	}
	if !cfg.Database.SQLite.ForeignKeys {
		t.Fatal("foreign_keys = false, want true")
	}
}

func TestLoadSQLitePathCanBeOverriddenByEnv(t *testing.T) {
	t.Setenv("SQLITE_PATH", filepath.Join(t.TempDir(), "env", "override.db"))
	path := writeConfig(t, `
system:
  router_prefix: /api/v1
i18n:
  path: locales
database:
  driver: sqlite3
  sqlite:
    path: data/app.db
`)

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.SQLite.Path != os.Getenv("SQLITE_PATH") {
		t.Fatalf("sqlite path = %q, want env override %q", cfg.Database.SQLite.Path, os.Getenv("SQLITE_PATH"))
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

func TestPostgresDsnUsesPgxKeywordFormat(t *testing.T) {
	postgres := Postgres{
		Host:     "postgres",
		Port:     "5432",
		Dbname:   "gva",
		Username: "gva",
		Password: "secret",
		SSLMode:  "disable",
		TimeZone: "Asia/Shanghai",
	}

	want := "host=postgres user=gva password=secret dbname=gva port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	if got := postgres.Dsn(); got != want {
		t.Fatalf("Dsn() = %q, want %q", got, want)
	}
}

func TestNormalizeDatabaseDriverSupportsSQLite3(t *testing.T) {
	if got := normalizeDatabaseDriver("sqlite3"); got != "sqlite3" {
		t.Fatalf("normalizeDatabaseDriver(sqlite3) = %q, want sqlite3", got)
	}
}
