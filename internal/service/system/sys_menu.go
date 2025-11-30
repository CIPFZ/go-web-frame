package system

import (
	"context"
	"errors"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	systemReq "github.com/CIPFZ/gowebframe/internal/model/system/request"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IMenuService interface {
	// GetUserMenuTree 根据角色ID获取动态菜单树
	GetUserMenuTree(ctx context.Context, authorityId uint) ([]systemModel.SysMenu, error)
	// GetMenuList 获取
	GetMenuList(ctx context.Context) ([]systemModel.SysMenu, error)
	AddBaseMenu(ctx context.Context, req systemReq.AddMenuReq) error
	UpdateBaseMenu(ctx context.Context, req systemReq.UpdateMenuReq) error
	DeleteBaseMenu(ctx context.Context, id uint) error
	GetMenuAuthority(ctx context.Context, authorityId uint) ([]systemModel.SysMenu, error)
}

type MenuService struct {
	svcCtx *svc.ServiceContext
}

func NewMenuService(svcCtx *svc.ServiceContext) IMenuService {
	return &MenuService{svcCtx: svcCtx}
}

// GetUserMenuTree 获取用户菜单树
func (s *MenuService) GetUserMenuTree(ctx context.Context, authorityId uint) ([]systemModel.SysMenu, error) {
	var allMenus []systemModel.SysMenu

	// 1. 查询数据库 (联表查询)
	// 利用 GORM 的 Join 查询，通过 sys_authority_menus 过滤出该角色拥有的菜单
	// 必须按 sort 排序，保证前端菜单顺序正确
	err := s.svcCtx.DB.WithContext(ctx).
		Table("sys_menus").
		Select("sys_menus.*").
		Joins("JOIN sys_authority_menus ON sys_menus.id = sys_authority_menus.menu_id").
		Where("sys_authority_menus.authority_id = ?", authorityId).
		Where("sys_menus.deleted_at IS NULL").
		Order("sys_menus.sort").
		Find(&allMenus).Error

	if err != nil {
		s.svcCtx.Logger.Error("get_menu_tree_db_error", zap.Error(err), zap.Uint("authority_id", authorityId))
		return nil, errors.New("获取菜单数据失败")
	}

	// 2. 列表转树 (List -> Tree)
	return s.buildMenuTree(allMenus), nil
}

// buildMenuTree 内存中构建树形结构 (O(n) 复杂度)
func (s *MenuService) buildMenuTree(menus []systemModel.SysMenu) []systemModel.SysMenu {
	// 使用 map 存储 parentId -> []children 的映射
	// 注意：这里存储的是切片索引或指针，为了方便组装
	treeMap := make(map[uint][]*systemModel.SysMenu)

	// 第一次遍历：将所有菜单按 ParentId 分组放入 Map
	for i := range menus {
		// 必须取地址，否则 range 中的 v 是拷贝，无法修改 Children
		treeMap[menus[i].ParentId] = append(treeMap[menus[i].ParentId], &menus[i])
	}

	// 第二次遍历：将子节点挂载到父节点上
	// 我们只需要处理那些“是父亲”的节点
	for i := range menus {
		// 尝试从 Map 中找到当前节点的子节点
		if children, ok := treeMap[menus[i].ID]; ok {
			// 因为 children 是 []*SysMenu，我们需要解引用放入切片
			// 这里稍微繁琐一点因为 Go 的切片类型匹配
			for _, child := range children {
				menus[i].Children = append(menus[i].Children, *child)
			}
		}
	}

	// 第三次遍历：提取根节点 (ParentId == 0)
	var rootMenus []systemModel.SysMenu
	// 这里直接从 treeMap[0] 取更高效
	if roots, ok := treeMap[0]; ok {
		for _, root := range roots {
			rootMenus = append(rootMenus, *root)
		}
	}

	return rootMenus
}

// GetMenuList 获取所有菜单列表 (管理后台用)
func (s *MenuService) GetMenuList(ctx context.Context) ([]systemModel.SysMenu, error) {
	var allMenus []systemModel.SysMenu
	// 查询所有未删除的菜单
	err := s.svcCtx.DB.WithContext(ctx).
		Order("sort").
		Find(&allMenus).Error

	if err != nil {
		return nil, err
	}

	// 复用之前的 buildMenuTree 方法生成树形结构
	return s.buildMenuTree(allMenus), nil
}

// AddBaseMenu 新增菜单
func (s *MenuService) AddBaseMenu(ctx context.Context, req systemReq.AddMenuReq) error {
	// 获取带 TraceID 的 Logger
	log := logger.GetLogger(ctx)

	// 1. ✨ 校验 Path 是否重复 (仅校验未删除的)
	var existed systemModel.SysMenu
	err := s.svcCtx.DB.WithContext(ctx).
		Where("path = ?", req.Path).
		First(&existed).Error

	if err == nil {
		log.Warn("添加菜单失败，Path已存在", zap.String("path", req.Path))
		return errors.New("路由Path已存在，请更换")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("检查Path重复时数据库出错", zap.Error(err))
		return err
	}

	// 2. 创建菜单
	menu := systemModel.SysMenu{
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

	if err := s.svcCtx.DB.WithContext(ctx).Create(&menu).Error; err != nil {
		log.Error("创建菜单失败", zap.Error(err))
		return err
	}
	return nil
}

// UpdateBaseMenu 更新菜单
func (s *MenuService) UpdateBaseMenu(ctx context.Context, req systemReq.UpdateMenuReq) error {
	log := logger.GetLogger(ctx)

	// 1. 检查是否存在
	var menu systemModel.SysMenu
	if err := s.svcCtx.DB.WithContext(ctx).First(&menu, req.ID).Error; err != nil {
		return errors.New("菜单不存在")
	}

	// 2. ✨ 校验 Path 是否与其他菜单重复 (排除自己)
	if req.Path != menu.Path {
		var duplicate systemModel.SysMenu
		if err := s.svcCtx.DB.WithContext(ctx).Where("path = ? AND id != ?", req.Path, req.ID).First(&duplicate).Error; err == nil {
			return errors.New("路由Path已存在")
		}
	}

	// 3. 更新字段 (使用 Map 以支持零值更新)
	updMap := map[string]interface{}{
		"parent_id":    req.ParentId,
		"path":         req.Path,
		"name":         req.Name,
		"component":    req.Component,
		"sort":         req.Sort,
		"icon":         req.Icon,
		"hide_in_menu": req.HideInMenu,
		"access":       req.Access,
		// ✨ 新增字段
		"target": req.Target,
		"locale": req.Locale,
	}

	if err := s.svcCtx.DB.WithContext(ctx).Model(&menu).Updates(updMap).Error; err != nil {
		log.Error("更新菜单失败", zap.Error(err))
		return err
	}
	return nil
}

// DeleteBaseMenu 删除菜单
func (s *MenuService) DeleteBaseMenu(ctx context.Context, id uint) error {
	log := logger.GetLogger(ctx)

	return s.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查子菜单
		var count int64
		if err := tx.Model(&systemModel.SysMenu{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("此菜单存在子菜单，不可删除")
		}

		// 2. ✨ 级联硬删除：sys_authority_menus (角色-菜单关联)
		// 这里必须使用 Unscoped() 进行硬删除，否则会有脏数据残留
		if err := tx.Table("sys_authority_menus").
			Where("menu_id = ?", id).
			Delete(nil).Error; err != nil {
			log.Error("删除菜单关联失败", zap.Error(err))
			return err
		}

		// 3. 删除菜单自身 (软删除)
		if err := tx.Delete(&systemModel.SysMenu{}, id).Error; err != nil {
			log.Error("删除菜单失败", zap.Error(err))
			return err
		}

		return nil
	})
}

// GetMenuAuthority 查看指定角色拥有的菜单 (用于权限分配回显)
func (s *MenuService) GetMenuAuthority(ctx context.Context, authorityId uint) ([]systemModel.SysMenu, error) {
	var menus []systemModel.SysMenu

	// 使用 INNER JOIN 关联查询
	// 只需要查出关联表中存在的菜单即可
	err := s.svcCtx.DB.WithContext(ctx).
		Model(&systemModel.SysMenu{}).
		Joins("INNER JOIN sys_authority_menus ON sys_menus.id = sys_authority_menus.menu_id").
		Where("sys_authority_menus.authority_id = ?", authorityId).
		Order("sys_menus.sort").
		Find(&menus).Error

	if err != nil {
		return nil, err
	}

	return menus, nil
}
