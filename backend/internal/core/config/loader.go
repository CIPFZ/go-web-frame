package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Load 加载配置
func Load(path string) (*Config, *viper.Viper, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(absPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
	})

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	normalizeConfig(&cfg, filepath.Dir(absPath))

	return &cfg, v, nil
}

func normalizeConfig(cfg *Config, configDir string) {
	cfg.System.RouterPrefix = normalizeRouterPrefix(cfg.System.RouterPrefix)
	cfg.I18n.Path = normalizeI18nPath(cfg.I18n.Path, configDir)
	normalizeDatabase(&cfg.Database, cfg.Mysql, configDir)
}

func normalizeDatabase(database *Database, legacy MySQL, configDir string) {
	database.Driver = normalizeDatabaseDriver(database.Driver)
	if database.Driver == "" {
		database.Driver = "mysql"
	}
	if isEmptyMySQL(database.MySQL) && !isEmptyMySQL(legacy) {
		database.MySQL = legacy
	}
	normalizeMySQLDefaults(&database.MySQL)
	normalizePostgresDefaults(&database.Postgres)
	normalizeSQLiteDefaults(&database.SQLite, configDir)
}

func normalizeDatabaseDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "mysql":
		return strings.ToLower(strings.TrimSpace(driver))
	case "postgres", "postgresql", "pgsql":
		return "postgres"
	case "sqlite3":
		return "sqlite3"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func isEmptyMySQL(mysql MySQL) bool {
	return strings.TrimSpace(mysql.Host) == "" &&
		strings.TrimSpace(mysql.Path) == "" &&
		strings.TrimSpace(mysql.Dbname) == "" &&
		strings.TrimSpace(mysql.Username) == ""
}

func normalizeMySQLDefaults(mysql *MySQL) {
	if strings.TrimSpace(mysql.Host) == "" {
		mysql.Host = strings.TrimSpace(mysql.Path)
	}
	if strings.TrimSpace(mysql.Path) == "" {
		mysql.Path = strings.TrimSpace(mysql.Host)
	}
	if strings.TrimSpace(mysql.Port) == "" {
		mysql.Port = "3306"
	}
	if strings.TrimSpace(mysql.Engine) == "" {
		mysql.Engine = "InnoDB"
	}
}

func normalizePostgresDefaults(postgres *Postgres) {
	if strings.TrimSpace(postgres.Port) == "" {
		postgres.Port = "5432"
	}
	if strings.TrimSpace(postgres.SSLMode) == "" {
		postgres.SSLMode = "disable"
	}
	if strings.TrimSpace(postgres.TimeZone) == "" {
		postgres.TimeZone = "Asia/Shanghai"
	}
}

func normalizeSQLiteDefaults(sqlite *SQLite, configDir string) {
	if envPath := strings.TrimSpace(os.Getenv("SQLITE_PATH")); envPath != "" {
		sqlite.Path = envPath
	}
	if strings.TrimSpace(sqlite.Path) == "" {
		sqlite.Path = filepath.Join(configDir, "data", "app.db")
	} else if !filepath.IsAbs(sqlite.Path) {
		sqlite.Path = filepath.Join(configDir, sqlite.Path)
	}
	if sqlite.BusyTimeoutMS <= 0 {
		sqlite.BusyTimeoutMS = 5000
	}
	if sqlite.MaxIdleConns <= 0 {
		sqlite.MaxIdleConns = 1
	}
	if sqlite.MaxOpenConns <= 0 {
		sqlite.MaxOpenConns = 1
	}
}

func normalizeRouterPrefix(prefix string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return "/api/v1"
	}
	return "/" + strings.Trim(trimmed, "/")
}

func normalizeI18nPath(path, configDir string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return filepath.Join(configDir, "locales")
	}
	if filepath.IsAbs(trimmed) {
		return trimmed
	}
	return filepath.Join(configDir, trimmed)
}
