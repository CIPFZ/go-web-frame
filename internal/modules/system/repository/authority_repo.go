package repository

import (
	"context"

	"github.com/CIPFZ/gowebframe/internal/modules/system/model"

	"gorm.io/gorm"
)

type IAuthorityRepository interface {
	// 查询
	GetAll(ctx context.Context) ([]model.SysAuthority, int64, error)
	FindById(ctx context.Context, authorityId uint) (*model.SysAuthority, error)
	CountByParentId(ctx context.Context, parentId uint) (int64, error)
	CountUserUsage(ctx context.Context, authorityId uint) (int64, error)

	// 写入
	Create(ctx context.Context, auth *model.SysAuthority) error
	Update(ctx context.Context, auth *model.SysAuthority, cols map[string]interface{}) error
	Delete(ctx context.Context, authorityId uint) error

	// 关联操作
	SetMenuAuthority(ctx context.Context, authorityId uint, menuIds []uint) error
}

type AuthorityRepository struct {
	db *gorm.DB
}

func NewAuthorityRepository(db *gorm.DB) IAuthorityRepository {
	return &AuthorityRepository{db: db}
}

// GetAll 获取所有角色 (用于构建树)
func (r *AuthorityRepository) GetAll(ctx context.Context) ([]model.SysAuthority, int64, error) {
	var list []model.SysAuthority
	var total int64

	db := r.db.WithContext(ctx).Model(&model.SysAuthority{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *AuthorityRepository) FindById(ctx context.Context, authorityId uint) (*model.SysAuthority, error) {
	var auth model.SysAuthority
	// 使用 authority_id 字段查询 (因为是自定义ID)
	err := r.db.WithContext(ctx).Where("authority_id = ?", authorityId).First(&auth).Error
	return &auth, err
}

func (r *AuthorityRepository) CountByParentId(ctx context.Context, parentId uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SysAuthority{}).Where("parent_id = ?", parentId).Count(&count).Error
	return count, err
}

// CountUserUsage 检查角色是否被用户使用
func (r *AuthorityRepository) CountUserUsage(ctx context.Context, authorityId uint) (int64, error) {
	var count int64
	// 查询 sys_user_authorities 表
	err := r.db.WithContext(ctx).Table("sys_user_authorities").Where("authority_id = ?", authorityId).Count(&count).Error
	return count, err
}

func (r *AuthorityRepository) Create(ctx context.Context, auth *model.SysAuthority) error {
	return r.db.WithContext(ctx).Create(auth).Error
}

func (r *AuthorityRepository) Update(ctx context.Context, auth *model.SysAuthority, cols map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(auth).Updates(cols).Error
}

func (r *AuthorityRepository) Delete(ctx context.Context, authorityId uint) error {
	// 这里的 Delete 会根据 GORM 配置执行软删除
	return r.db.WithContext(ctx).Where("authority_id = ?", authorityId).Delete(&model.SysAuthority{}).Error
}

// SetMenuAuthority 设置角色菜单权限 (事务)
func (r *AuthorityRepository) SetMenuAuthority(ctx context.Context, authorityId uint, menuIds []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 硬删除旧关联
		if err := tx.Table("sys_authority_menus").
			Where("authority_id = ?", authorityId).
			Delete(nil).Error; err != nil {
			return err
		}

		// 2. 批量插入新关联
		if len(menuIds) > 0 {
			var relations []model.SysAuthorityMenu
			for _, menuId := range menuIds {
				relations = append(relations, model.SysAuthorityMenu{
					AuthorityId: authorityId,
					MenuId:      menuId,
				})
			}
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
