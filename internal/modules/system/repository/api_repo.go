package repository

import (
	"context"
	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"gorm.io/gorm"
)

type IApiRepository interface {
	GetList(ctx context.Context, req dto.SearchApiReq) ([]model.SysApi, int64, error)
	FindById(ctx context.Context, id uint) (*model.SysApi, error)
	FindByPathMethod(ctx context.Context, path, method string) (*model.SysApi, error)
	Create(ctx context.Context, api *model.SysApi) error
	UpdateWithSyncCasbin(ctx context.Context, oldApi *model.SysApi, newApi model.SysApi) error
	DeleteWithSyncCasbin(ctx context.Context, ids []uint) error
}

type ApiRepository struct {
	db *gorm.DB
}

func NewApiRepository(db *gorm.DB) IApiRepository {
	return &ApiRepository{db: db}
}

func (r *ApiRepository) GetList(ctx context.Context, req dto.SearchApiReq) ([]model.SysApi, int64, error) {
	logger.GetLogger(ctx).Info("CCCCCCC ->")
	var list []model.SysApi
	var total int64
	db := r.db.WithContext(ctx).Model(&model.SysApi{})

	if req.Path != "" {
		db = db.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Description != "" {
		db = db.Where("description LIKE ?", "%"+req.Description+"%")
	}
	if req.Method != "" {
		db = db.Where("method = ?", req.Method)
	}
	if req.ApiGroup != "" {
		db = db.Where("api_group = ?", req.ApiGroup)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("id desc").
		Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).
		Find(&list).Error

	return list, total, err
}

func (r *ApiRepository) FindById(ctx context.Context, id uint) (*model.SysApi, error) {
	var api model.SysApi
	err := r.db.WithContext(ctx).First(&api, id).Error
	return &api, err
}

func (r *ApiRepository) FindByPathMethod(ctx context.Context, path, method string) (*model.SysApi, error) {
	var api model.SysApi
	err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&api).Error
	return &api, err
}

func (r *ApiRepository) Create(ctx context.Context, api *model.SysApi) error {
	return r.db.WithContext(ctx).Create(api).Error
}

// UpdateWithSyncCasbin 更新API并同步更新Casbin规则
func (r *ApiRepository) UpdateWithSyncCasbin(ctx context.Context, oldApi *model.SysApi, newApi model.SysApi) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 更新 API 表
		if err := tx.Model(oldApi).Updates(newApi).Error; err != nil {
			return err
		}

		// 2. 如果路径或方法变了，更新 Casbin 规则表
		if oldApi.Path != newApi.Path || oldApi.Method != newApi.Method {
			if err := tx.Table("sys_casbin_rules").
				Where("v1 = ? AND v2 = ?", oldApi.Path, oldApi.Method).
				Updates(map[string]interface{}{
					"v1": newApi.Path,
					"v2": newApi.Method,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteWithSyncCasbin 删除API并同步删除Casbin规则
func (r *ApiRepository) DeleteWithSyncCasbin(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查询要删除的 API (为了拿 Path/Method 去删 Casbin)
		var apis []model.SysApi
		if err := tx.Find(&apis, ids).Error; err != nil {
			return err
		}

		// 2. 删除 API
		if err := tx.Delete(&model.SysApi{}, ids).Error; err != nil {
			return err
		}

		// 3. 删除 Casbin 规则
		for _, api := range apis {
			if err := tx.Table("sys_casbin_rules").
				Where("v1 = ? AND v2 = ?", api.Path, api.Method).
				Delete(nil).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
