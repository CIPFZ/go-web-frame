package system

import (
	"errors"
	"strconv"

	"github.com/CIPFZ/gowebframe/internal/model/common/request"
	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

type IMenuService interface {
	getMenuTreeMap(authorityId uint) (map[uint][]systemModel.SysBaseMenu, error)
	GetMenuTree(authorityId uint) ([]systemModel.SysBaseMenu, error)
	buildChildrenTree(menu *systemModel.SysBaseMenu, treeMap map[uint][]systemModel.SysBaseMenu) (err error)
	getChildrenList(menu *systemModel.SysMenu, treeMap map[uint][]systemModel.SysMenu) (err error)
	GetInfoList(authorityID uint) (list interface{}, err error)
	getBaseChildrenList(menu *systemModel.SysBaseMenu, treeMap map[uint][]systemModel.SysBaseMenu) (err error)
	AddBaseMenu(menu systemModel.SysBaseMenu) error
	getBaseMenuTreeMap(authorityID uint) (treeMap map[uint][]systemModel.SysBaseMenu, err error)
	GetBaseMenuTree(authorityID uint) (menus []systemModel.SysBaseMenu, err error)
	AddMenuAuthority(menus []systemModel.SysBaseMenu, adminAuthorityID, authorityId uint) (err error)
	GetMenuAuthority(info *request.GetAuthorityId) (menus []systemModel.SysMenu, err error)
	UserAuthorityDefaultRouter(c *gin.Context, user *systemModel.SysUser)
}

type MenuService struct {
	svcCtx *svc.ServiceContext
}

func NewMenuService(svcCtx *svc.ServiceContext) IMenuService {
	return &MenuService{svcCtx: svcCtx}
}

// GetMenuTree 获取动态菜单树 (给前端, 适配 Antd Pro)
// 返回 []systemModel.SysBaseMenu，其 Children 字段的 json tag 应为 "routes"
func (s *MenuService) GetMenuTree(authorityId uint) ([]systemModel.SysBaseMenu, error) {
	// 1. 获取该角色有权限的菜单 Map
	menuTreeMap, err := s.getMenuTreeMap(authorityId)
	if err != nil {
		return nil, err
	}

	// 2. 从根 (ParentId = 0) 开始构建树
	menus := menuTreeMap[0]
	for i := 0; i < len(menus); i++ {
		// 递归填充 Children
		err = s.buildChildrenTree(&menus[i], menuTreeMap)
		if err != nil {
			return nil, err
		}
	}
	return menus, nil
}

//// getMenuTreeMap 获取路由总树map
//func (s *MenuService) getMenuTreeMap(authorityId uint) (map[uint][]systemModel.SysMenu, error) {
//	var allMenus []systemModel.SysMenu
//	var baseMenu []systemModel.SysBaseMenu
//	var btns []systemModel.SysAuthorityBtn
//	var err error
//	treeMap := make(map[uint][]systemModel.SysMenu)
//
//	var SysAuthorityMenus []systemModel.SysAuthorityMenu
//	err = s.svcCtx.DB.Where("sys_authority_authority_id = ?", authorityId).Find(&SysAuthorityMenus).Error
//	if err != nil {
//		return nil, err
//	}
//
//	var MenuIds []string
//
//	for i := range SysAuthorityMenus {
//		MenuIds = append(MenuIds, SysAuthorityMenus[i].MenuId)
//	}
//
//	err = s.svcCtx.DB.Where("id in (?)", MenuIds).Order("sort").Preload("Parameters").Find(&baseMenu).Error
//	if err != nil {
//		return nil, err
//	}
//
//	for i := range baseMenu {
//		allMenus = append(allMenus, systemModel.SysMenu{
//			SysBaseMenu: baseMenu[i],
//			AuthorityId: authorityId,
//			MenuId:      baseMenu[i].ID,
//			Parameters:  nil,
//		})
//	}
//
//	err = s.svcCtx.DB.Where("authority_id = ?", authorityId).Preload("SysBaseMenuBtn").Find(&btns).Error
//	if err != nil {
//		return nil, err
//	}
//	var btnMap = make(map[uint]map[string]uint)
//	for _, v := range btns {
//		if btnMap[v.SysMenuID] == nil {
//			btnMap[v.SysMenuID] = make(map[string]uint)
//		}
//		btnMap[v.SysMenuID][v.SysBaseMenuBtn.Name] = authorityId
//	}
//	for _, v := range allMenus {
//		v.Btns = btnMap[v.SysBaseMenu.ID]
//		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
//	}
//	return treeMap, err
//}

// GetMenuTree 获取动态菜单树
//func (s *MenuService) GetMenuTree(authorityId uint) (menus []systemModel.SysMenu, err error) {
//	menuTree, err := s.getMenuTreeMap(authorityId)
//	menus = menuTree[0]
//	for i := 0; i < len(menus); i++ {
//		err = s.getChildrenList(&menus[i], menuTree)
//	}
//	return menus, err
//}

// getMenuTreeMap (内部) - 获取角色有权限的菜单并构建 ParentId -> []Menus 的 Map
func (s *MenuService) getMenuTreeMap(authorityId uint) (map[uint][]systemModel.SysBaseMenu, error) {
	var baseMenus []systemModel.SysBaseMenu
	treeMap := make(map[uint][]systemModel.SysBaseMenu)

	// 1. 仍然使用 GVA 的方式，通过联结表获取菜单 ID
	var SysAuthorityMenus []systemModel.SysAuthorityMenu
	err := s.svcCtx.DB.Where("sys_authority_authority_id = ?", authorityId).Find(&SysAuthorityMenus).Error
	if err != nil {
		return nil, err
	}

	var MenuIds []string
	for i := range SysAuthorityMenus {
		MenuIds = append(MenuIds, SysAuthorityMenus[i].MenuId)
	}

	// 2. 查询 SysBaseMenu
	// 优化：不再 Preload 已被移除的 GVA 字段 (Parameters, MenuBtn)
	err = s.svcCtx.DB.Where("id IN (?)", MenuIds).Order("sort").Find(&baseMenus).Error
	if err != nil {
		return nil, err
	}

	// 3. 构建 Map
	// 优化：不再需要 SysMenu 包装器，也不再需要处理按钮权限 (Btns)
	for _, v := range baseMenus {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, nil
}

// buildChildrenTree (内部) - 递归构建子菜单树
// (替换旧的 getChildrenList，逻辑相同，签名更新)
func (s *MenuService) buildChildrenTree(menu *systemModel.SysBaseMenu, treeMap map[uint][]systemModel.SysBaseMenu) (err error) {
	// 关键：使用 menu.ID (主键) 作为 key 来查找子菜单
	menu.Children = treeMap[menu.ID]
	for i := 0; i < len(menu.Children); i++ {
		err = s.buildChildrenTree(&menu.Children[i], treeMap)
	}
	return err
}

// getChildrenList 获取子菜单
func (s *MenuService) getChildrenList(menu *systemModel.SysMenu, treeMap map[uint][]systemModel.SysMenu) (err error) {
	menu.Children = treeMap[menu.MenuId]
	for i := 0; i < len(menu.Children); i++ {
		err = s.getChildrenList(&menu.Children[i], treeMap)
	}
	return err
}

// GetInfoList 获取路由分页
func (s *MenuService) GetInfoList(authorityID uint) (list interface{}, err error) {
	var menuList []systemModel.SysBaseMenu
	treeMap, err := s.getBaseMenuTreeMap(authorityID)
	menuList = treeMap[0]
	for i := 0; i < len(menuList); i++ {
		err = s.getBaseChildrenList(&menuList[i], treeMap)
	}
	return menuList, err
}

// getBaseChildrenList 获取菜单的子菜单
func (s *MenuService) getBaseChildrenList(menu *systemModel.SysBaseMenu, treeMap map[uint][]systemModel.SysBaseMenu) (err error) {
	menu.Children = treeMap[menu.ID]
	for i := 0; i < len(menu.Children); i++ {
		err = s.getBaseChildrenList(&menu.Children[i], treeMap)
	}
	return err
}

// AddBaseMenu 添加基础路由
func (s *MenuService) AddBaseMenu(menu systemModel.SysBaseMenu) error {
	return s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 检查name是否重复
		if !errors.Is(tx.Where("name = ?", menu.Name).First(&systemModel.SysBaseMenu{}).Error, gorm.ErrRecordNotFound) {
			return errors.New("存在重复name，请修改name")
		}

		if menu.ParentId != 0 {
			// 检查父菜单是否存在
			var parentMenu systemModel.SysBaseMenu
			if err := tx.First(&parentMenu, menu.ParentId).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("父菜单不存在")
				}
				return err
			}

			// 检查父菜单下现有子菜单数量
			var existingChildrenCount int64
			err := tx.Model(&systemModel.SysBaseMenu{}).Where("parent_id = ?", menu.ParentId).Count(&existingChildrenCount).Error
			if err != nil {
				return err
			}

			// 如果父菜单原本是叶子菜单（没有子菜单），现在要变成枝干菜单，需要清空其权限分配
			if existingChildrenCount == 0 {
				// 检查父菜单是否被其他角色设置为首页
				var defaultRouterCount int64
				err := tx.Model(&systemModel.SysAuthority{}).Where("default_router = ?", parentMenu.Name).Count(&defaultRouterCount).Error
				if err != nil {
					return err
				}
				if defaultRouterCount > 0 {
					return errors.New("父菜单已被其他角色的首页占用，请先释放父菜单的首页权限")
				}

				// 清空父菜单的所有权限分配
				err = tx.Where("sys_base_menu_id = ?", menu.ParentId).Delete(&systemModel.SysAuthorityMenu{}).Error
				if err != nil {
					return err
				}
			}
		}

		// 创建菜单
		return tx.Create(&menu).Error
	})
}

// getBaseMenuTreeMap 获取路由总树map
func (s *MenuService) getBaseMenuTreeMap(authorityID uint) (treeMap map[uint][]systemModel.SysBaseMenu, err error) {
	authorityService := NewAuthorityService(s.svcCtx)
	parentAuthorityID, err := authorityService.GetParentAuthorityID(authorityID)
	if err != nil {
		return nil, err
	}

	var allMenus []systemModel.SysBaseMenu
	treeMap = make(map[uint][]systemModel.SysBaseMenu)
	db := s.svcCtx.DB.Order("sort").Preload("MenuBtn").Preload("Parameters")

	// 当开启了严格的树角色并且父角色不为0时需要进行菜单筛选
	if s.svcCtx.Config.System.UseStrictAuth && parentAuthorityID != 0 {
		var authorityMenus []systemModel.SysAuthorityMenu
		err = s.svcCtx.DB.Where("sys_authority_authority_id = ?", authorityID).Find(&authorityMenus).Error
		if err != nil {
			return nil, err
		}
		var menuIds []string
		for i := range authorityMenus {
			menuIds = append(menuIds, authorityMenus[i].MenuId)
		}
		db = db.Where("id in (?)", menuIds)
	}

	err = db.Find(&allMenus).Error
	for _, v := range allMenus {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, err
}

// GetBaseMenuTree 获取基础路由树
func (s *MenuService) GetBaseMenuTree(authorityID uint) (menus []systemModel.SysBaseMenu, err error) {
	treeMap, err := s.getBaseMenuTreeMap(authorityID)
	menus = treeMap[0]
	for i := 0; i < len(menus); i++ {
		err = s.getBaseChildrenList(&menus[i], treeMap)
	}
	return menus, err
}

// AddMenuAuthority 为角色增加menu树
func (s *MenuService) AddMenuAuthority(menus []systemModel.SysBaseMenu, adminAuthorityID, authorityId uint) (err error) {
	var auth systemModel.SysAuthority
	auth.AuthorityId = authorityId
	auth.SysBaseMenus = menus

	authorityService := NewAuthorityService(s.svcCtx)
	err = authorityService.CheckAuthorityIDAuth(adminAuthorityID, authorityId)
	if err != nil {
		return err
	}

	var authority systemModel.SysAuthority
	_ = s.svcCtx.DB.First(&authority, "authority_id = ?", adminAuthorityID).Error
	var menuIds []string

	// 当开启了严格的树角色并且父角色不为0时需要进行菜单筛选
	if s.svcCtx.Config.System.UseStrictAuth && *authority.ParentId != 0 {
		var authorityMenus []systemModel.SysAuthorityMenu
		err = s.svcCtx.DB.Where("sys_authority_authority_id = ?", adminAuthorityID).Find(&authorityMenus).Error
		if err != nil {
			return err
		}
		for i := range authorityMenus {
			menuIds = append(menuIds, authorityMenus[i].MenuId)
		}

		for i := range menus {
			hasMenu := false
			for j := range menuIds {
				idStr := strconv.Itoa(int(menus[i].ID))
				if idStr == menuIds[j] {
					hasMenu = true
				}
			}
			if !hasMenu {
				return errors.New("添加失败,请勿跨级操作")
			}
		}
	}

	err = authorityService.SetMenuAuthority(&auth)
	return err
}

// GetMenuAuthority 查看当前角色树
func (s *MenuService) GetMenuAuthority(info *request.GetAuthorityId) (menus []systemModel.SysMenu, err error) {
	var baseMenu []systemModel.SysBaseMenu
	var SysAuthorityMenus []systemModel.SysAuthorityMenu
	err = s.svcCtx.DB.Where("sys_authority_authority_id = ?", info.AuthorityId).Find(&SysAuthorityMenus).Error
	if err != nil {
		return
	}

	var MenuIds []string

	for i := range SysAuthorityMenus {
		MenuIds = append(MenuIds, SysAuthorityMenus[i].MenuId)
	}

	err = s.svcCtx.DB.Where("id in (?) ", MenuIds).Order("sort").Find(&baseMenu).Error

	for i := range baseMenu {
		menus = append(menus, systemModel.SysMenu{
			SysBaseMenu: baseMenu[i],
			AuthorityId: info.AuthorityId,
			MenuId:      baseMenu[i].ID,
			Parameters:  nil,
		})
	}
	return menus, err
}

// UserAuthorityDefaultRouter 用户角色默认路由检查
func (s *MenuService) UserAuthorityDefaultRouter(c *gin.Context, user *systemModel.SysUser) {
	var menuIds []string
	err := s.svcCtx.DB.Model(&systemModel.SysAuthorityMenu{}).Where("sys_authority_authority_id = ?", user.AuthorityId).Pluck("sys_base_menu_id", &menuIds).Error
	if err != nil {
		return
	}
	var am systemModel.SysBaseMenu
	err = s.svcCtx.DB.First(&am, "name = ? and id in (?)", user.Authority.DefaultRouter, menuIds).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user.Authority.DefaultRouter = "404"
	}
}
