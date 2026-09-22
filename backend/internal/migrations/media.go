package migrations

import (
	"strconv"
	"strings"

	"github.com/CIPFZ/gowebframe/internal/modules/media"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ensureMediaModule(tx *gorm.DB, prefix string) error {
	menu := model.SysMenu{
		Path: "/media/sources", Name: "影视采集", NameEn: "Media collection",
		Component: "media/sources", Locale: "menu.mediaSources", Icon: "VideoCameraOutlined", Sort: 17,
	}
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
	endpoints := []struct {
		suffix string
		method string
		desc   string
	}{
		{"/media/sources", "GET", "List media sources"},
		{"/media/sources", "POST", "Create media source"},
		{"/media/sources/:id", "PUT", "Update media source"},
		{"/media/sources/:id/health", "POST", "Check media source health"},
		{"/media/sources/:id/collect", "POST", "Collect media source records"},
		{"/media/items", "GET", "List collected media items"},
		{"/media/items/:id", "GET", "Get collected media item"},
	}
	for _, endpoint := range endpoints {
		api := model.SysApi{
			Path:   strings.TrimRight(prefix, "/") + endpoint.suffix,
			Method: endpoint.method, ApiGroup: "media", Description: endpoint.desc,
		}
		if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
			return err
		}
		if adminCount == 0 {
			continue
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SysAuthorityApi{AuthorityId: 1, ApiId: api.ID}).Error; err != nil {
			return err
		}
		rule := model.SysCasbinRule{Ptype: "p", V0: strconv.Itoa(1), V1: api.Path, V2: endpoint.method}
		if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).FirstOrCreate(&rule).Error; err != nil {
			return err
		}
	}
	return tx.AutoMigrate(
		&media.Source{}, &media.Item{}, &media.SourceItem{}, &media.PlayLine{},
		&media.PlayEpisode{}, &media.CollectionRun{}, &media.HealthCheck{},
	)
}
