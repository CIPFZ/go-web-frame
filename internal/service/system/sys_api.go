package system

import (
	"context"
	"errors"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"gorm.io/gorm"
)

type IApiService interface {
	GetApiList(ctx context.Context, req systemReq.SearchApiReq) ([]systemModel.SysApi, int64, error)
	CreateApi(ctx context.Context, req systemReq.CreateApiReq) error
	UpdateApi(ctx context.Context, req systemReq.UpdateApiReq) error
	DeleteApi(ctx context.Context, req systemReq.DeleteApiReq) error
}

type ApiService struct {
	svcCtx *svc.ServiceContext
}

func NewApiService(svcCtx *svc.ServiceContext) IApiService {
	return &ApiService{svcCtx: svcCtx}
}

// GetApiList 分页获取
func (s *ApiService) GetApiList(ctx context.Context, req systemReq.SearchApiReq) ([]systemModel.SysApi, int64, error) {
	var list []systemModel.SysApi
	var total int64

	db := s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysApi{})

	// 动态查询条件
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

	// 默认按 ID 倒序
	err := db.Order("id desc").
		Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).
		Find(&list).Error

	return list, total, err
}

// CreateApi 新增
func (s *ApiService) CreateApi(ctx context.Context, req systemReq.CreateApiReq) error {
	// 1. 检查是否存在 (Path + Method 必须唯一)
	var existed systemModel.SysApi
	if !errors.Is(s.svcCtx.DB.WithContext(ctx).Where("path = ? AND method = ?", req.Path, req.Method).First(&existed).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同路径和方法的API")
	}

	api := systemModel.SysApi{
		Path:        req.Path,
		Description: req.Description,
		ApiGroup:    req.ApiGroup,
		Method:      req.Method,
	}

	return s.svcCtx.DB.WithContext(ctx).Create(&api).Error
}

// UpdateApi 更新 (重点：同步 Casbin)
func (s *ApiService) UpdateApi(ctx context.Context, req systemReq.UpdateApiReq) error {
	var oldApi systemModel.SysApi
	err := s.svcCtx.DB.WithContext(ctx).First(&oldApi, req.ID).Error
	if err != nil {
		return errors.New("API不存在")
	}

	// 1. 检查是否更改了关键字段 (Path 或 Method)
	if oldApi.Path != req.Path || oldApi.Method != req.Method {
		// 检查新值是否冲突
		var existed systemModel.SysApi
		if !errors.Is(s.svcCtx.DB.WithContext(ctx).Where("path = ? AND method = ? AND id != ?", req.Path, req.Method, req.ID).First(&existed).Error, gorm.ErrRecordNotFound) {
			return errors.New("修改后的API路径和方法已存在")
		}
	}

	// 2. 开启事务
	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// a. 更新 API 表
		api := systemModel.SysApi{
			Path:        req.Path,
			Description: req.Description,
			ApiGroup:    req.ApiGroup,
			Method:      req.Method,
		}
		// 这里手动设置 ID 以便 GORM 知道更新哪条
		api.ID = req.ID

		if err := tx.Model(&oldApi).Updates(api).Error; err != nil {
			return err
		}

		// b. ✨ 关键：如果 Path 或 Method 变了，同步更新 Casbin 规则
		if oldApi.Path != req.Path || oldApi.Method != req.Method {
			// Casbin UpdatePolicy(oldRule, newRule)
			// 参数顺序依赖于你的 Model 定义，通常是: sub, obj, act
			// 这里的 obj=Path, act=Method
			// 我们需要更新所有角色 (sub) 的这条规则，这比较麻烦
			// 简单的做法是：UpdateFilteredPolicy 也就是直接用 SQL 更新 sys_casbin_rules 表更直接

			// GVA 的做法是直接操作 Casbin API：
			// UpdatePolicy(oldParams, newParams)
			// 但 UpdatePolicy 只能更新单条特定的规则。

			// 最稳妥的“批量更新”做法是：
			// 在 sys_casbin_rules 表中，将所有 v1=oldPath AND v2=oldMethod 的记录
			// 更新为 v1=newPath AND v2=newMethod

			if err := tx.Table("sys_casbin_rules").
				Where("v1 = ? AND v2 = ?", oldApi.Path, oldApi.Method).
				Updates(map[string]interface{}{
					"v1": req.Path,
					"v2": req.Method,
				}).Error; err != nil {
				return err
			}

			// 更新完数据库后，记得在 Controller 层或这里 reload Casbin 策略
			// 稍后在 API 层调用 LoadPolicy
		}

		return nil
	})
}

// DeleteApi 删除/批量删除 (重点：同步 Casbin)
func (s *ApiService) DeleteApi(ctx context.Context, req systemReq.DeleteApiReq) error {
	var ids []uint
	if req.ID != 0 {
		ids = append(ids, req.ID)
	}
	if len(req.IDs) > 0 {
		ids = append(ids, req.IDs...)
	}
	if len(ids) == 0 {
		return errors.New("未选择数据")
	}

	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先查出要删除的 API 详情 (为了拿 Path 和 Method 去清理 Casbin)
		var apis []systemModel.SysApi
		if err := tx.Find(&apis, ids).Error; err != nil {
			return err
		}

		// 2. 删除 API 表记录 (软删除或硬删除，GVA API通常是硬删除或软删除均可，这里用软删除)
		if err := tx.Delete(&systemModel.SysApi{}, ids).Error; err != nil {
			return err
		}

		// 3. ✨ 清理 Casbin 规则
		// 必须把关联的权限策略也删掉，否则会残留垃圾数据
		for _, api := range apis {
			// 删除所有包含此 (Path, Method) 的规则
			// 对应 Casbin 字段: v1=path, v2=method
			if err := tx.Table("sys_casbin_rules").
				Where("v1 = ? AND v2 = ?", api.Path, api.Method).
				Delete(nil).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
