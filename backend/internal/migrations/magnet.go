package migrations

import (
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

// Install the catalog and grant only the built-in administrator. Other roles
// can be assigned these entries through the existing permission UI.
func ensureMagnetModule(tx *gorm.DB, prefix string) error {
	menu := model.SysMenu{Path: "/magnet/preview", Name: "磁力预览", NameEn: "Magnet preview",
		Component: "magnet/preview", Locale: "menu.magnetPreview", Icon: "LinkOutlined", Sort: 15}
	if err := tx.Where("path = ?", menu.Path).FirstOrCreate(&menu).Error; err != nil {
		return err
	}
	var adminCount int64
	if err := tx.Model(&model.SysAuthority{}).Where("authority_id = ?", 1).Count(&adminCount).Error; err != nil {
		return err
	}
	if adminCount > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SysAuthorityMenu{MenuId: menu.ID, AuthorityId: 1}).Error; err != nil {
			return err
		}
	}
	for suffix, method := range map[string]string{"/magnet/preview": "POST", "/magnet/cover/:hash": "GET"} {
		api := model.SysApi{Path: strings.TrimRight(prefix, "/") + suffix, Method: method, ApiGroup: "magnet", Description: "Magnet resource preview"}
		if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
			return err
		}
		if adminCount == 0 {
			continue
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SysAuthorityApi{AuthorityId: 1, ApiId: api.ID}).Error; err != nil {
			return err
		}
		rule := model.SysCasbinRule{Ptype: "p", V0: strconv.Itoa(1), V1: api.Path, V2: method}
		if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).FirstOrCreate(&rule).Error; err != nil {
			return err
		}
	}
	return nil
}
