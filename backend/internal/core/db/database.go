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
