package system

import (
	"context"
	"errors"

	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"gorm.io/gorm"
)

type IAuthorityService interface {
	GetAuthorityList(ctx context.Context, pageInfo request.PageInfo) ([]systemModel.SysAuthority, int64, error)
	CreateAuthority(ctx context.Context, req systemReq.CreateAuthorityReq) error
	UpdateAuthority(ctx context.Context, req systemReq.UpdateAuthorityReq) error
	DeleteAuthority(ctx context.Context, authId uint) error
	SetAuthorityMenus(ctx context.Context, req systemReq.SetAuthorityMenusReq) error
}

type AuthorityService struct {
	svcCtx *svc.ServiceContext
}

func NewAuthorityService(svcCtx *svc.ServiceContext) IAuthorityService {
	return &AuthorityService{svcCtx: svcCtx}
}

// GetAuthorityList 获取角色列表 (全量树形结构)
func (s *AuthorityService) GetAuthorityList(ctx context.Context, pageInfo request.PageInfo) ([]systemModel.SysAuthority, int64, error) {
	var allList []systemModel.SysAuthority
	var total int64

	db := s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysAuthority{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 全量查询，不分页，以便构建树
	if err := db.Find(&allList).Error; err != nil {
		return nil, 0, err
	}

	// 构建树形结构
	return s.buildAuthorityTree(allList), total, nil
}

// CreateAuthority 创建角色
func (s *AuthorityService) CreateAuthority(ctx context.Context, req systemReq.CreateAuthorityReq) error {
	// 查重
	var existed systemModel.SysAuthority
	if !errors.Is(s.svcCtx.DB.WithContext(ctx).First(&existed, "authority_id = ?", req.AuthorityId).Error, gorm.ErrRecordNotFound) {
		return errors.New("角色ID已存在")
	}

	auth := systemModel.SysAuthority{
		AuthorityId:   req.AuthorityId,
		AuthorityName: req.AuthorityName,
		ParentId:      req.ParentId,
		DefaultRouter: req.DefaultRouter,
	}
	return s.svcCtx.DB.WithContext(ctx).Create(&auth).Error
}

// UpdateAuthority 更新角色
func (s *AuthorityService) UpdateAuthority(ctx context.Context, req systemReq.UpdateAuthorityReq) error {
	return s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysAuthority{}).
		Where("authority_id = ?", req.AuthorityId).
		Updates(map[string]interface{}{
			"authority_name": req.AuthorityName,
			"default_router": req.DefaultRouter,
		}).Error
}

// DeleteAuthority 删除角色
func (s *AuthorityService) DeleteAuthority(ctx context.Context, authId uint) error {
	// 1. 检查是否有子角色
	var childCount int64
	s.svcCtx.DB.WithContext(ctx).Model(&systemModel.SysAuthority{}).Where("parent_id = ?", authId).Count(&childCount)
	if childCount > 0 {
		return errors.New("此角色存在子角色，不可删除")
	}

	// 2. 检查是否有用户使用
	var userCount int64
	s.svcCtx.DB.WithContext(ctx).Table("sys_user_authorities").Where("authority_id = ?", authId).Count(&userCount)
	if userCount > 0 {
		return errors.New("此角色有用户正在使用，不可删除")
	}

	// 3. 删除角色 (软删除)
	// 关联表清理通常依赖 GORM 的级联设置，或者手动清理 sys_authority_menus 和 sys_authority_apis
	// 简单起见，这里先只删除角色本身
	return s.svcCtx.DB.WithContext(ctx).Where("authority_id = ?", authId).Delete(&systemModel.SysAuthority{}).Error
}

// SetAuthorityMenus 设置角色菜单权限 (事务)
func (s *AuthorityService) SetAuthorityMenus(ctx context.Context, req systemReq.SetAuthorityMenusReq) error {
	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 硬删除旧关联
		if err := tx.Table("sys_authority_menus").
			Where("authority_id = ?", req.AuthorityId).
			Delete(nil).Error; err != nil {
			return err
		}

		// 2. 批量插入新关联
		if len(req.MenuIds) > 0 {
			var relations []systemModel.SysAuthorityMenu
			for _, menuId := range req.MenuIds {
				relations = append(relations, systemModel.SysAuthorityMenu{
					AuthorityId: req.AuthorityId,
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

// buildAuthorityTree 内部辅助：构建树
func (s *AuthorityService) buildAuthorityTree(list []systemModel.SysAuthority) []systemModel.SysAuthority {
	treeMap := make(map[uint][]*systemModel.SysAuthority)
	for i := range list {
		treeMap[list[i].ParentId] = append(treeMap[list[i].ParentId], &list[i])
	}
	for i := range list {
		if children, ok := treeMap[list[i].AuthorityId]; ok {
			for _, child := range children {
				list[i].Children = append(list[i].Children, *child)
			}
		}
	}
	// 提取根节点 (ParentId=0)
	var roots []systemModel.SysAuthority
	if rootNodes, ok := treeMap[0]; ok {
		for _, root := range rootNodes {
			roots = append(roots, *root)
		}
	}
	return roots
}
