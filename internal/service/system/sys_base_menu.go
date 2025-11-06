package system

import (
	"errors"

	systemModel "github.com/CIPFZ/gowebframe/internal/model/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

type IBaseMenuService interface {
	DeleteBaseMenu(c *gin.Context, id int) (err error)
	UpdateBaseMenu(c *gin.Context, menu systemModel.SysBaseMenu) (err error)
	GetBaseMenuById(c *gin.Context, id int) (menu systemModel.SysBaseMenu, err error)
}

type BaseMenuService struct {
	svcCtx *svc.ServiceContext
}

func NewBaseMenuService(svcCtx *svc.ServiceContext) IBaseMenuService {
	return &BaseMenuService{svcCtx: svcCtx}
}

// DeleteBaseMenu 删除基础路由
func (s *BaseMenuService) DeleteBaseMenu(c *gin.Context, id int) (err error) {
	err = s.svcCtx.DB.First(&systemModel.SysBaseMenu{}, "parent_id = ?", id).Error
	if err == nil {
		return errors.New("此菜单存在子菜单不可删除")
	}
	var menu systemModel.SysBaseMenu
	err = s.svcCtx.DB.First(&menu, id).Error
	if err != nil {
		return errors.New("记录不存在")
	}
	err = s.svcCtx.DB.First(&systemModel.SysAuthority{}, "default_router = ?", menu.Name).Error
	if err == nil {
		return errors.New("此菜单有角色正在作为首页，不可删除")
	}
	return s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {

		err = tx.Delete(&systemModel.SysBaseMenu{}, "id = ?", id).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&systemModel.SysBaseMenuParameter{}, "sys_base_menu_id = ?", id).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&systemModel.SysBaseMenuBtn{}, "sys_base_menu_id = ?", id).Error
		if err != nil {
			return err
		}
		err = tx.Delete(&systemModel.SysAuthorityBtn{}, "sys_menu_id = ?", id).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&systemModel.SysAuthorityMenu{}, "sys_base_menu_id = ?", id).Error
		if err != nil {
			return err
		}
		return nil
	})

}

// UpdateBaseMenu 更新路由
func (s *BaseMenuService) UpdateBaseMenu(c *gin.Context, menu systemModel.SysBaseMenu) (err error) {
	//var oldMenu systemModel.SysBaseMenu
	//upDateMap := make(map[string]interface{})
	//upDateMap["keep_alive"] = menu.KeepAlive
	//upDateMap["transition_type"] = menu.TransitionType
	//upDateMap["close_tab"] = menu.CloseTab
	//upDateMap["default_menu"] = menu.DefaultMenu
	//upDateMap["parent_id"] = menu.ParentId
	//upDateMap["path"] = menu.Path
	//upDateMap["name"] = menu.Name
	//upDateMap["hidden"] = menu.HideInMenu
	//upDateMap["component"] = menu.Component
	//upDateMap["title"] = menu.Title
	//upDateMap["active_name"] = menu.ActiveName
	//upDateMap["icon"] = menu.Icon
	//upDateMap["sort"] = menu.Sort
	//
	//log := logger.GetLogger(c)
	//
	//err = s.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
	//	tx.Where("id = ?", menu.ID).Find(&oldMenu)
	//	if oldMenu.Name != menu.Name {
	//		if !errors.Is(tx.Where("id <> ? AND name = ?", menu.ID, menu.Name).First(&systemModel.SysBaseMenu{}).Error, gorm.ErrRecordNotFound) {
	//			log.Debug("存在相同name修改失败")
	//			return errors.New("存在相同name修改失败")
	//		}
	//	}
	//	txErr := tx.Unscoped().Delete(&systemModel.SysBaseMenuParameter{}, "sys_base_menu_id = ?", menu.ID).Error
	//	if txErr != nil {
	//		log.Debug(txErr.Error())
	//		return txErr
	//	}
	//	txErr = tx.Unscoped().Delete(&systemModel.SysBaseMenuBtn{}, "sys_base_menu_id = ?", menu.ID).Error
	//	if txErr != nil {
	//		log.Debug(txErr.Error())
	//		return txErr
	//	}
	//	if len(menu.Parameters) > 0 {
	//		for k := range menu.Parameters {
	//			menu.Parameters[k].SysBaseMenuID = menu.ID
	//		}
	//		txErr = tx.Create(&menu.Parameters).Error
	//		if txErr != nil {
	//			log.Debug(txErr.Error())
	//			return txErr
	//		}
	//	}
	//
	//	if len(menu.MenuBtn) > 0 {
	//		for k := range menu.MenuBtn {
	//			menu.MenuBtn[k].SysBaseMenuID = menu.ID
	//		}
	//		txErr = tx.Create(&menu.MenuBtn).Error
	//		if txErr != nil {
	//			log.Debug(txErr.Error())
	//			return txErr
	//		}
	//	}
	//
	//	txErr = tx.Model(&oldMenu).Updates(upDateMap).Error
	//	if txErr != nil {
	//		log.Debug(txErr.Error())
	//		return txErr
	//	}
	//	return nil
	//})
	return err
}

// GetBaseMenuById 返回当前选中menu
func (s *BaseMenuService) GetBaseMenuById(c *gin.Context, id int) (menu systemModel.SysBaseMenu, err error) {
	err = s.svcCtx.DB.Preload("MenuBtn").Preload("Parameters").Where("id = ?", id).First(&menu).Error
	return
}
