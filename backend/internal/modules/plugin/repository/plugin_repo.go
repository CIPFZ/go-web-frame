package repository

import (
	"context"

	"github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"gorm.io/gorm"
)

type IPluginRepository interface {
	DB() *gorm.DB
	CreatePlugin(ctx context.Context, item *model.Plugin) error
	UpdatePlugin(ctx context.Context, item *model.Plugin, updates map[string]interface{}) error
	FindPluginByID(ctx context.Context, id uint) (*model.Plugin, error)
	ListPlugins(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.Plugin, int64, error)
	CreateRelease(ctx context.Context, item *model.PluginRelease, compatibles []model.PluginCompatibleProduct) error
	UpdateRelease(ctx context.Context, item *model.PluginRelease, updates map[string]interface{}, compatibles []model.PluginCompatibleProduct) error
	FindReleaseByID(ctx context.Context, id uint) (*model.PluginRelease, error)
	FindLatestPublishedReleaseByPluginID(ctx context.Context, pluginID uint) (*model.PluginRelease, error)
	ListPublishedReleasesByPluginID(ctx context.Context, pluginID uint) ([]model.PluginRelease, error)
	ListReleasesByPluginID(ctx context.Context, pluginID uint) ([]model.PluginRelease, error)
	ListWorkOrders(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.PluginRelease, int64, error)
	ListPublishedPlugins(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.Plugin, int64, error)
	FindPublishedPluginByID(ctx context.Context, id uint) (*model.Plugin, *model.PluginRelease, error)
	CreateEvent(ctx context.Context, item *model.PluginReleaseEvent) error
	ListEventsByReleaseID(ctx context.Context, releaseID uint) ([]model.PluginReleaseEvent, error)
	ClaimRelease(ctx context.Context, id, claimerID uint) (bool, error)
	ResetClaim(ctx context.Context, id uint) error
	ListProducts(ctx context.Context, page, pageSize int) ([]model.PluginProduct, int64, error)
	CreateProduct(ctx context.Context, item *model.PluginProduct) error
	UpdateProduct(ctx context.Context, item *model.PluginProduct, updates map[string]interface{}) error
	FindProductByID(ctx context.Context, id uint) (*model.PluginProduct, error)
	ListDepartments(ctx context.Context, page, pageSize int) ([]model.PluginDepartment, int64, error)
}

type PluginRepository struct {
	db *gorm.DB
}

func NewPluginRepository(db *gorm.DB) IPluginRepository {
	return &PluginRepository{db: db}
}

func (r *PluginRepository) DB() *gorm.DB { return r.db }

func (r *PluginRepository) CreatePlugin(ctx context.Context, item *model.Plugin) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PluginRepository) UpdatePlugin(ctx context.Context, item *model.Plugin, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(item).Updates(updates).Error
}

func (r *PluginRepository) FindPluginByID(ctx context.Context, id uint) (*model.Plugin, error) {
	var item model.Plugin
	err := r.db.WithContext(ctx).Preload("Department").First(&item, id).Error
	return &item, err
}

func (r *PluginRepository) ListPlugins(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.Plugin, int64, error) {
	var items []model.Plugin
	var total int64
	if err := query.WithContext(ctx).Model(&model.Plugin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	err := query.WithContext(ctx).Preload("Department").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *PluginRepository) CreateRelease(ctx context.Context, item *model.PluginRelease, compatibles []model.PluginCompatibleProduct) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		for i := range compatibles {
			compatibles[i].ReleaseID = item.ID
		}
		if len(compatibles) > 0 {
			if err := tx.Create(&compatibles).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PluginRepository) UpdateRelease(ctx context.Context, item *model.PluginRelease, updates map[string]interface{}, compatibles []model.PluginCompatibleProduct) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(item).Updates(updates).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("release_id = ?", item.ID).Delete(&model.PluginCompatibleProduct{}).Error; err != nil {
			return err
		}
		for i := range compatibles {
			compatibles[i].BaseModel = model.ToBaseModel(0)
			compatibles[i].ReleaseID = item.ID
		}
		if len(compatibles) > 0 {
			if err := tx.Create(&compatibles).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PluginRepository) FindReleaseByID(ctx context.Context, id uint) (*model.PluginRelease, error) {
	var item model.PluginRelease
	err := r.db.WithContext(ctx).
		Preload("Plugin").
		Preload("Plugin.Department").
		Preload("CompatibleItems").
		Preload("CompatibleItems.Product").
		First(&item, id).Error
	return &item, err
}

func (r *PluginRepository) FindLatestPublishedReleaseByPluginID(ctx context.Context, pluginID uint) (*model.PluginRelease, error) {
	var item model.PluginRelease
	err := r.db.WithContext(ctx).
		Where("plugin_id = ? AND status = ?", pluginID, model.ReleaseStatusReleased).
		Preload("CompatibleItems").
		Preload("CompatibleItems.Product").
		Order("released_at desc, id desc").
		First(&item).Error
	return &item, err
}

func (r *PluginRepository) ListPublishedReleasesByPluginID(ctx context.Context, pluginID uint) ([]model.PluginRelease, error) {
	var items []model.PluginRelease
	err := r.db.WithContext(ctx).
		Where("plugin_id = ? AND status = ?", pluginID, model.ReleaseStatusReleased).
		Preload("CompatibleItems").
		Preload("CompatibleItems.Product").
		Order("released_at desc, id desc").
		Find(&items).Error
	return items, err
}

func (r *PluginRepository) ListReleasesByPluginID(ctx context.Context, pluginID uint) ([]model.PluginRelease, error) {
	var items []model.PluginRelease
	err := r.db.WithContext(ctx).
		Where("plugin_id = ?", pluginID).
		Preload("CompatibleItems").
		Preload("CompatibleItems.Product").
		Order("id desc").
		Find(&items).Error
	return items, err
}

func (r *PluginRepository) ListWorkOrders(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.PluginRelease, int64, error) {
	var items []model.PluginRelease
	var total int64
	if err := query.WithContext(ctx).Model(&model.PluginRelease{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	err := query.WithContext(ctx).
		Preload("Plugin").
		Preload("CompatibleItems").
		Preload("CompatibleItems.Product").
		Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

func (r *PluginRepository) ListPublishedPlugins(ctx context.Context, query *gorm.DB, page, pageSize int) ([]model.Plugin, int64, error) {
	var items []model.Plugin
	var total int64
	if err := query.WithContext(ctx).Model(&model.Plugin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	err := query.WithContext(ctx).Preload("Department").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *PluginRepository) FindPublishedPluginByID(ctx context.Context, id uint) (*model.Plugin, *model.PluginRelease, error) {
	plugin, err := r.FindPluginByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	release, err := r.FindLatestPublishedReleaseByPluginID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return plugin, release, nil
}

func (r *PluginRepository) CreateEvent(ctx context.Context, item *model.PluginReleaseEvent) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PluginRepository) ListEventsByReleaseID(ctx context.Context, releaseID uint) ([]model.PluginReleaseEvent, error) {
	var items []model.PluginReleaseEvent
	err := r.db.WithContext(ctx).Where("release_id = ?", releaseID).Order("id desc").Find(&items).Error
	return items, err
}

func (r *PluginRepository) ClaimRelease(ctx context.Context, id, claimerID uint) (bool, error) {
	now := gorm.Expr("CURRENT_TIMESTAMP")
	res := r.db.WithContext(ctx).Model(&model.PluginRelease{}).
		Where("id = ? AND status = ? AND process_status = ?", id, model.ReleaseStatusPendingReview, model.ReleaseProcessStatusPending).
		Updates(map[string]interface{}{
			"claimer_id":     claimerID,
			"claimed_at":     now,
			"process_status": model.ReleaseProcessStatusProcessing,
		})
	return res.RowsAffected == 1, res.Error
}

func (r *PluginRepository) ResetClaim(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.PluginRelease{}).Where("id = ?", id).Updates(map[string]interface{}{
		"claimer_id":     nil,
		"claimed_at":     nil,
		"process_status": model.ReleaseProcessStatusPending,
	}).Error
}

func (r *PluginRepository) ListProducts(ctx context.Context, page, pageSize int) ([]model.PluginProduct, int64, error) {
	var items []model.PluginProduct
	var total int64
	query := r.db.WithContext(ctx).Model(&model.PluginProduct{}).Where("status = ?", true)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 999
	}
	err := query.Order("sort asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *PluginRepository) CreateProduct(ctx context.Context, item *model.PluginProduct) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PluginRepository) UpdateProduct(ctx context.Context, item *model.PluginProduct, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(item).Updates(updates).Error
}

func (r *PluginRepository) FindProductByID(ctx context.Context, id uint) (*model.PluginProduct, error) {
	var item model.PluginProduct
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}

func (r *PluginRepository) ListDepartments(ctx context.Context, page, pageSize int) ([]model.PluginDepartment, int64, error) {
	var items []model.PluginDepartment
	var total int64
	query := r.db.WithContext(ctx).Model(&model.PluginDepartment{}).Where("status = ?", true)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 999
	}
	err := query.Order("sort asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}
