package service

import (
	"context"
	"errors"

	logger "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IMenuService interface {
	GetUserMenuTree(ctx context.Context, authorityId uint) ([]model.SysMenu, error)
	GetMenuList(ctx context.Context) ([]model.SysMenu, error)
	AddBaseMenu(ctx context.Context, req dto.AddMenuReq) error
	UpdateBaseMenu(ctx context.Context, req dto.UpdateMenuReq) error
	DeleteBaseMenu(ctx context.Context, id uint) error
	GetMenuAuthority(ctx context.Context, authorityId uint) ([]model.SysMenu, error)
}

type MenuService struct {
	svcCtx   *svc.ServiceContext
	menuRepo repository.IMenuRepository // 依赖注入
}

func NewMenuService(svcCtx *svc.ServiceContext, menuRepo repository.IMenuRepository) IMenuService {
	return &MenuService{
		svcCtx:   svcCtx,
		menuRepo: menuRepo,
	}
}

// GetUserMenuTree 根据角色ID获取动态菜单树
func (s *MenuService) GetUserMenuTree(ctx context.Context, authorityId uint) ([]model.SysMenu, error) {
	// 1. 通过 Repo 查询
	allMenus, err := s.menuRepo.GetByAuthorityId(ctx, authorityId)
	if err != nil {
		s.svcCtx.Logger.Error("get_menu_tree_repo_error", zap.Error(err))
		return nil, errors.New("获取菜单数据失败")
	}

	// 2. 业务逻辑: 构建树
	return s.buildMenuTree(allMenus), nil
}

// GetMenuList 获取所有菜单列表
func (s *MenuService) GetMenuList(ctx context.Context) ([]model.SysMenu, error) {
	allMenus, err := s.menuRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(allMenus), nil
}

// AddBaseMenu 新增菜单
func (s *MenuService) AddBaseMenu(ctx context.Context, req dto.AddMenuReq) error {
	log := logger.GetLogger(ctx)

	// 1. 查重
	_, err := s.menuRepo.FindByPath(ctx, req.Path)
	if err == nil {
		log.Warn("添加菜单失败，Path已存在", zap.String("path", req.Path))
		return errors.New("路由Path已存在，请更换")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 2. 构建对象
	menu := model.SysMenu{
		ParentId:   req.ParentId,
		Path:       req.Path,
		Name:       req.Name,
		Component:  req.Component,
		Sort:       req.Sort,
		Icon:       req.Icon,
		HideInMenu: req.HideInMenu,
		Access:     req.Access,
		Target:     req.Target,
		Locale:     req.Locale,
	}

	// 3. 写入
	if err := s.menuRepo.Create(ctx, &menu); err != nil {
		log.Error("创建菜单失败", zap.Error(err))
		return err
	}
	return nil
}

// UpdateBaseMenu 更新菜单
func (s *MenuService) UpdateBaseMenu(ctx context.Context, req dto.UpdateMenuReq) error {
	log := logger.GetLogger(ctx)

	// 1. 查旧数据
	menu, err := s.menuRepo.FindById(ctx, req.ID)
	if err != nil {
		return errors.New("菜单不存在")
	}

	// 2. 查重 (如果修改了 Path)
	if req.Path != menu.Path {
		if _, err := s.menuRepo.FindByPathExcludeId(ctx, req.Path, req.ID); err == nil {
			return errors.New("路由Path已存在")
		}
	}

	// 3. 更新
	updMap := map[string]interface{}{
		"parent_id":    req.ParentId,
		"path":         req.Path,
		"name":         req.Name,
		"component":    req.Component,
		"sort":         req.Sort,
		"icon":         req.Icon,
		"hide_in_menu": req.HideInMenu,
		"access":       req.Access,
		"target":       req.Target,
		"locale":       req.Locale,
	}

	if err := s.menuRepo.Update(ctx, menu, updMap); err != nil {
		log.Error("更新菜单失败", zap.Error(err))
		return err
	}
	return nil
}

// DeleteBaseMenu 删除菜单
func (s *MenuService) DeleteBaseMenu(ctx context.Context, id uint) error {
	log := logger.GetLogger(ctx)

	// 1. 检查子菜单
	count, err := s.menuRepo.CountByParentId(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("此菜单存在子菜单，不可删除")
	}

	// 2. 级联删除
	if err := s.menuRepo.DeleteWithAssociations(ctx, id); err != nil {
		log.Error("删除菜单失败", zap.Error(err))
		return err
	}
	return nil
}

// GetMenuAuthority 获取指定角色的菜单 (用于回显)
func (s *MenuService) GetMenuAuthority(ctx context.Context, authorityId uint) ([]model.SysMenu, error) {
	// 直接调用 Repo
	return s.menuRepo.GetByAuthorityId(ctx, authorityId)
}

// buildMenuTree (纯内存逻辑，保持不变)
func (s *MenuService) buildMenuTree(menus []model.SysMenu) []model.SysMenu {
	// ... (代码逻辑保持不变)
	// 注意：确保 model 包名一致
	treeMap := make(map[uint][]*model.SysMenu)
	for i := range menus {
		treeMap[menus[i].ParentId] = append(treeMap[menus[i].ParentId], &menus[i])
	}
	for i := range menus {
		if children, ok := treeMap[menus[i].ID]; ok {
			for _, child := range children {
				menus[i].Children = append(menus[i].Children, *child)
			}
		}
	}
	var rootMenus []model.SysMenu
	if roots, ok := treeMap[0]; ok {
		for _, root := range roots {
			rootMenus = append(rootMenus, *root)
		}
	}
	return rootMenus
}
