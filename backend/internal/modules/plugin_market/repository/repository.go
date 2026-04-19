package repository

import (
	"context"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(&model.MarketPlugin{}, &model.MarketPluginVersion{}, &model.MarketPluginCompatibility{})
}

func (r *Repository) CountPlugins(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.MarketPlugin{}).Count(&total).Error
	return total, err
}

func (r *Repository) UpsertPlugin(ctx context.Context, item *model.MarketPlugin) error {
	var existing model.MarketPlugin
	err := r.db.WithContext(ctx).Where("plugin_id = ?", item.PluginID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.WithContext(ctx).Create(item).Error
		}
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"code":           item.Code,
		"name_zh":        item.NameZh,
		"name_en":        item.NameEn,
		"description_zh": item.DescriptionZh,
		"description_en": item.DescriptionEn,
		"capability_zh":  item.CapabilityZh,
		"capability_en":  item.CapabilityEn,
		"owner_name":     item.OwnerName,
		"status":         item.Status,
		"latest_version": item.LatestVersion,
	}).Error
}

func (r *Repository) FindPluginByPluginID(ctx context.Context, pluginID uint) (*model.MarketPlugin, error) {
	var item model.MarketPlugin
	err := r.db.WithContext(ctx).Where("plugin_id = ?", pluginID).First(&item).Error
	return &item, err
}

func (r *Repository) UpsertVersion(ctx context.Context, version *model.MarketPluginVersion, compat []model.MarketPluginCompatibility) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.MarketPluginVersion
		err := tx.Where("release_id = ?", version.ReleaseID).First(&existing).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(version).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"plugin_ref_id":      version.PluginRefID,
				"plugin_id":          version.PluginID,
				"version":            version.Version,
				"changelog_zh":       version.ChangelogZh,
				"changelog_en":       version.ChangelogEn,
				"test_report_url":    version.TestReportURL,
				"package_x86_url":    version.PackageX86URL,
				"package_arm_url":    version.PackageARMURL,
				"released_at":        version.ReleasedAt,
				"version_constraint": version.VersionConstraint,
				"publisher":          version.Publisher,
				"status":             version.Status,
				"plugin_name_zh":     version.PluginNameZh,
				"plugin_name_en":     version.PluginNameEn,
				"description_zh":     version.DescriptionZh,
				"description_en":     version.DescriptionEn,
				"capability_zh":      version.CapabilityZh,
				"capability_en":      version.CapabilityEn,
				"owner_name":         version.OwnerName,
			}).Error; err != nil {
				return err
			}
			version.ID = existing.ID
		}

		if version.ID == 0 {
			if err := tx.Where("release_id = ?", version.ReleaseID).First(version).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("release_id = ?", version.ReleaseID).Delete(&model.MarketPluginCompatibility{}).Error; err != nil {
			return err
		}
		for i := range compat {
			compat[i].VersionRefID = version.ID
			compat[i].ReleaseID = version.ReleaseID
		}
		if len(compat) > 0 {
			if err := tx.Create(&compat).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) OfflineVersion(ctx context.Context, releaseID uint) error {
	return r.db.WithContext(ctx).Model(&model.MarketPluginVersion{}).Where("release_id = ?", releaseID).Update("status", "offlined").Error
}

func (r *Repository) DeletePlugin(ctx context.Context, pluginID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var plugin model.MarketPlugin
		if err := tx.Where("plugin_id = ?", pluginID).First(&plugin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}
		if err := tx.Where("release_id IN (?)",
			tx.Model(&model.MarketPluginVersion{}).Select("release_id").Where("plugin_ref_id = ?", plugin.ID),
		).Delete(&model.MarketPluginCompatibility{}).Error; err != nil {
			return err
		}
		if err := tx.Where("plugin_ref_id = ?", plugin.ID).Delete(&model.MarketPluginVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&plugin).Error
	})
}

func (r *Repository) ListPublishedPlugins(ctx context.Context, keyword string, page, pageSize int) ([]model.MarketPlugin, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	publishedVersionSubQuery := r.db.WithContext(ctx).
		Model(&model.MarketPluginVersion{}).
		Select("1").
		Where("market_plugin_versions.plugin_ref_id = market_plugins.id").
		Where("market_plugin_versions.status = ?", "published")

	query := r.db.WithContext(ctx).
		Model(&model.MarketPlugin{}).
		Where("status = ?", "published").
		Where("EXISTS (?)", publishedVersionSubQuery)
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		query = query.Where("code LIKE ? OR name_zh LIKE ? OR name_en LIKE ? OR description_zh LIKE ? OR description_en LIKE ?", like, like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.MarketPlugin
	err := query.Order("updated_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *Repository) FindPublishedDetail(ctx context.Context, pluginID uint) (*model.MarketPlugin, []model.MarketPluginVersion, error) {
	var plugin model.MarketPlugin
	if err := r.db.WithContext(ctx).Where("plugin_id = ? AND status = ?", pluginID, "published").First(&plugin).Error; err != nil {
		return nil, nil, err
	}
	var versions []model.MarketPluginVersion
	if err := r.db.WithContext(ctx).
		Where("plugin_ref_id = ? AND status = ?", plugin.ID, "published").
		Preload("CompatibleItems").
		Order("released_at desc, id desc").
		Find(&versions).Error; err != nil {
		return nil, nil, err
	}
	return &plugin, versions, nil
}

func (r *Repository) FindLatestPublishedVersion(ctx context.Context, pluginRefID uint) (*model.MarketPluginVersion, error) {
	var version model.MarketPluginVersion
	err := r.db.WithContext(ctx).
		Where("plugin_ref_id = ? AND status = ?", pluginRefID, "published").
		Preload("CompatibleItems").
		Order("released_at desc, id desc").
		First(&version).Error
	return &version, err
}
