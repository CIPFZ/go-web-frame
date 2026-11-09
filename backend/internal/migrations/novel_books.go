package migrations

import (
	"errors"
	"fmt"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	novelModel "github.com/CIPFZ/gowebframe/internal/modules/novel/model"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const novelSourceSchema = "ebook_treasure_chest"

// migrateNovelBooks creates the CMS-local catalog and takes an idempotent
// snapshot from the existing cloud ebook catalog. The source is never changed.
func migrateNovelBooks(db *gorm.DB, cfg *config.Config, logger *zap.Logger) error {
	if err := db.AutoMigrate(&novelModel.NovelBook{}); err != nil {
		return fmt.Errorf("create novel_books table: %w", err)
	}

	// SQLite and other test dialects do not expose MySQL's information_schema.
	// They still get the local table and permissions, but no cross-database copy.
	if db.Dialector.Name() != "mysql" {
		return ensureNovelAccess(db, cfg.System.RouterPrefix)
	}

	var sourceExists int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", novelSourceSchema, "books").Scan(&sourceExists).Error; err != nil {
		return fmt.Errorf("check novel source table: %w", err)
	}
	if sourceExists > 0 {
		const copySQL = `
INSERT INTO novel_books
	(id, title, author, category, source_url, source_hash, record_hash, language, formats, created_at, updated_at)
SELECT id, title, author, category, source_url, source_hash, record_hash, language, formats, created_at, updated_at
FROM ebook_treasure_chest.books
ON DUPLICATE KEY UPDATE
	title = VALUES(title),
	author = VALUES(author),
	category = VALUES(category),
	source_url = VALUES(source_url),
	source_hash = VALUES(source_hash),
	language = VALUES(language),
	formats = VALUES(formats),
	created_at = VALUES(created_at),
	updated_at = VALUES(updated_at)`
		result := db.Exec(copySQL)
		if result.Error != nil {
			return fmt.Errorf("copy novel catalog: %w", result.Error)
		}
		if logger != nil {
			logger.Info("novel catalog imported into CMS database", zap.Int64("affected_rows", result.RowsAffected))
		}
	} else if logger != nil {
		logger.Info("novel source table is absent; created empty CMS catalog", zap.String("table", novelModel.NovelBook{}.TableName()))
	}
	return ensureNovelAccess(db, cfg.System.RouterPrefix)
}

func dropNovelBookLevel(db *gorm.DB) error {
	if db.Dialector.Name() == "mysql" && db.Migrator().HasColumn(&novelModel.NovelBook{}, "level") {
		if err := db.Migrator().DropColumn(&novelModel.NovelBook{}, "level"); err != nil {
			return fmt.Errorf("drop novel_books.level: %w", err)
		}
	}
	return nil
}

func dropNovelBookDescription(db *gorm.DB) error {
	if db.Dialector.Name() == "mysql" && db.Migrator().HasColumn(&novelModel.NovelBook{}, "description") {
		if err := db.Migrator().DropColumn(&novelModel.NovelBook{}, "description"); err != nil {
			return fmt.Errorf("drop novel_books.description: %w", err)
		}
	}
	return nil
}

func ensureNovelAccess(db *gorm.DB, prefix string) error {
	var menu sysModel.SysMenu
	err := db.Where("path = ?", "/novel/books").First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = sysModel.SysMenu{Path: "/novel/books", Name: "小说库", NameEn: "Novel Library", Component: "novel/books", Icon: "ReadOutlined", Locale: "menu.novel.books", Sort: 80}
		if err := db.Create(&menu).Error; err != nil {
			return fmt.Errorf("create novel menu: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("load novel menu: %w", err)
	}

	apiSpecs := []sysModel.SysApi{
		{Path: prefix + "/novel/book/list", Method: "POST", ApiGroup: "novel-book", Description: "List novel books"},
		{Path: prefix + "/novel/book/detail", Method: "GET", ApiGroup: "novel-book", Description: "Get novel book detail"},
	}
	return grantAccess(db, []uint{1, 9528}, []uint{menu.ID}, apiSpecs)
}
