package config

import (
	"fmt"
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
