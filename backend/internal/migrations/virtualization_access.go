package migrations

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func migrateVirtualizationAccess(db *gorm.DB, cfg *config.Config, logger *zap.Logger) error {
	if err := ensureVirtualizationAccess(db, cfg.System.RouterPrefix); err != nil {
		return err
	}
	if err := ensureVirtualizationStorageAccess(db, cfg.System.RouterPrefix); err != nil {
		return err
	}
	if logger != nil {
		logger.Info("virtualization access metadata ensured")
	}
	return nil
}

func ensureVirtualizationAccess(db *gorm.DB, prefix string) error {
	var menu sysModel.SysMenu
	err := db.Where("path = ?", "/virtualization/vms").First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = sysModel.SysMenu{
			Path: "/virtualization/vms", Name: "\u865a\u62df\u673a\u7ba1\u7406", NameEn: "Virtual Machines",
			Component: "virtualization/vms", Icon: "CloudServerOutlined",
			Locale: "menu.virtualization.vms", Sort: 30,
		}
		if err := db.Create(&menu).Error; err != nil {
			return fmt.Errorf("create virtualization menu: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("load virtualization menu: %w", err)
	}

	var agentMenu sysModel.SysMenu
	if err := db.Where("path = ?", "/virtualization/agent").First(&agentMenu).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		agentMenu = sysModel.SysMenu{ParentId: menu.ID, Path: "/virtualization/agent", Name: "Agent 联动", NameEn: "Agent Tasks", Component: "virtualization/agent", Icon: "ApiOutlined", Locale: "menu.virtualization.agent", Sort: 32}
		if err := db.Create(&agentMenu).Error; err != nil {
			return fmt.Errorf("create agent menu: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("load agent menu: %w", err)
	}
	apiSpecs := []sysModel.SysApi{
		{Path: prefix + "/virtualization/vms", Method: "GET", ApiGroup: "virtualization", Description: "List virtual machines"},
		{Path: prefix + "/virtualization/vms/:name", Method: "GET", ApiGroup: "virtualization", Description: "Get virtual machine"},
		{Path: prefix + "/virtualization/vms", Method: "POST", ApiGroup: "virtualization", Description: "Create virtual machine"},
		{Path: prefix + "/virtualization/vms/:name", Method: "PUT", ApiGroup: "virtualization", Description: "Update virtual machine"},
		{Path: prefix + "/virtualization/vms/:name/start", Method: "POST", ApiGroup: "virtualization", Description: "Start virtual machine"},
		{Path: prefix + "/virtualization/vms/:name/shutdown", Method: "POST", ApiGroup: "virtualization", Description: "Shutdown virtual machine"},
		{Path: prefix + "/virtualization/vms/:name/reboot", Method: "POST", ApiGroup: "virtualization", Description: "Reboot virtual machine"},
		{Path: prefix + "/virtualization/vms/:name/force-stop", Method: "POST", ApiGroup: "virtualization", Description: "Force stop virtual machine"},
		{Path: prefix + "/virtualization/vms/:name", Method: "DELETE", ApiGroup: "virtualization", Description: "Delete virtual machine"},
		{Path: prefix + "/virtualization/vms/:name/agent", Method: "GET", ApiGroup: "virtualization", Description: "Get Agent status"},
		{Path: prefix + "/virtualization/vms/:name/agent/health", Method: "GET", ApiGroup: "virtualization", Description: "Check Agent Gateway connectivity"},
		{Path: prefix + "/virtualization/vms/:name/agent/tasks", Method: "POST", ApiGroup: "virtualization", Description: "Create Agent task"},
		{Path: prefix + "/virtualization/agent/tasks/:taskID", Method: "GET", ApiGroup: "virtualization", Description: "Get Agent task"},
		{Path: prefix + "/virtualization/agent/tasks/:taskID/cancel", Method: "POST", ApiGroup: "virtualization", Description: "Cancel Agent task"},
	}
	apis := make([]sysModel.SysApi, 0, len(apiSpecs))
	for _, spec := range apiSpecs {
		var item sysModel.SysApi
		err := db.Where("path = ? AND method = ?", spec.Path, spec.Method).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = spec
			if err := db.Create(&item).Error; err != nil {
				return fmt.Errorf("create virtualization api: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("load virtualization api: %w", err)
		}
		apis = append(apis, item)
	}

	for _, authorityID := range []uint{1, 9528} {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityMenu{AuthorityId: authorityID, MenuId: menu.ID}).Error; err != nil {
			return fmt.Errorf("grant virtualization menu: %w", err)
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityMenu{AuthorityId: authorityID, MenuId: agentMenu.ID}).Error; err != nil {
			return fmt.Errorf("grant agent menu: %w", err)
		}
		for _, item := range apis {
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityApi{AuthorityId: authorityID, ApiId: item.ID}).Error; err != nil {
				return fmt.Errorf("grant virtualization api: %w", err)
			}
			var count int64
			if err := db.Model(&sysModel.SysCasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.FormatUint(uint64(authorityID), 10), item.Path, item.Method).Count(&count).Error; err != nil {
				return fmt.Errorf("check virtualization policy: %w", err)
			}
			if count == 0 {
				if err := db.Create(&sysModel.SysCasbinRule{Ptype: "p", V0: strconv.FormatUint(uint64(authorityID), 10), V1: item.Path, V2: item.Method}).Error; err != nil {
					return fmt.Errorf("create virtualization policy: %w", err)
				}
			}
		}
	}
	return nil
}

func ensureVirtualizationStorageAccess(db *gorm.DB, prefix string) error {
	var menu sysModel.SysMenu
	err := db.Where("path = ?", "/virtualization/storage").First(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = sysModel.SysMenu{
			Path: "/virtualization/storage", Name: "\u5b58\u50a8\u7ba1\u7406", NameEn: "Storage",
			Component: "virtualization/storage", Icon: "HddOutlined",
			Locale: "menu.virtualization.storage", Sort: 31,
		}
		if err := db.Create(&menu).Error; err != nil {
			return fmt.Errorf("create virtualization storage menu: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("load virtualization storage menu: %w", err)
	}
	apiSpecs := []sysModel.SysApi{
		{Path: prefix + "/virtualization/storage/pools", Method: "GET", ApiGroup: "virtualization", Description: "List storage pools"},
		{Path: prefix + "/virtualization/storage/volumes", Method: "GET", ApiGroup: "virtualization", Description: "List storage volumes"},
		{Path: prefix + "/virtualization/storage/volumes", Method: "POST", ApiGroup: "virtualization", Description: "Create storage volume"},
		{Path: prefix + "/virtualization/storage/volumes/:pool/:name", Method: "PUT", ApiGroup: "virtualization", Description: "Resize storage volume"},
		{Path: prefix + "/virtualization/storage/volumes/:pool/:name", Method: "DELETE", ApiGroup: "virtualization", Description: "Delete storage volume"},
	}
	apis := make([]sysModel.SysApi, 0, len(apiSpecs))
	for _, spec := range apiSpecs {
		var item sysModel.SysApi
		err := db.Where("path = ? AND method = ?", spec.Path, spec.Method).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = spec
			if err := db.Create(&item).Error; err != nil {
				return fmt.Errorf("create virtualization storage api: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("load virtualization storage api: %w", err)
		}
		apis = append(apis, item)
	}
	for _, authorityID := range []uint{1, 9528} {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityMenu{AuthorityId: authorityID, MenuId: menu.ID}).Error; err != nil {
			return fmt.Errorf("grant virtualization storage menu: %w", err)
		}
		for _, item := range apis {
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sysModel.SysAuthorityApi{AuthorityId: authorityID, ApiId: item.ID}).Error; err != nil {
				return fmt.Errorf("grant virtualization storage api: %w", err)
			}
			var count int64
			if err := db.Model(&sysModel.SysCasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.FormatUint(uint64(authorityID), 10), item.Path, item.Method).Count(&count).Error; err != nil {
				return fmt.Errorf("check virtualization storage policy: %w", err)
			}
			if count == 0 {
				if err := db.Create(&sysModel.SysCasbinRule{Ptype: "p", V0: strconv.FormatUint(uint64(authorityID), 10), V1: item.Path, V2: item.Method}).Error; err != nil {
					return fmt.Errorf("create virtualization storage policy: %w", err)
				}
			}
		}
	}
	return nil
}
