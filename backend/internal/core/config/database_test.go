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
