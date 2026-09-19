package migrations

import (
	"strconv"
	"strings"

	proxyModel "github.com/CIPFZ/gowebframe/internal/modules/proxy"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ensureProxyModule installs the proxy manager menu, API catalog and administrator grants.
func ensureProxyModule(tx *gorm.DB, prefix string) error {
	menu := model.SysMenu{
		Path: "/proxy/manager", Name: "代理管理", NameEn: "Proxy management",
		Component: "proxy/manager", Locale: "menu.proxyManager", Icon: "ApiOutlined", Sort: 16,
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
		{"/proxy/instances", "GET", "List proxy instances"},
		{"/proxy/instances", "POST", "Create proxy instance"},
		{"/proxy/instances/:id", "GET", "Get proxy instance"},
		{"/proxy/instances/:id", "PUT", "Update proxy instance"},
		{"/proxy/instances/:id/action", "POST", "Control proxy service"},
		{"/proxy/instances/:id/config", "GET", "Read proxy configuration"},
		{"/proxy/instances/:id/config/validate", "POST", "Validate proxy configuration"},
		{"/proxy/instances/:id/config", "PUT", "Save proxy configuration"},
		{"/proxy/instances/:id/config/rollback", "POST", "Rollback proxy configuration"},
		{"/proxy/instances/:id/metrics", "GET", "Read proxy metrics"},
	}
	for _, endpoint := range endpoints {
		api := model.SysApi{
			Path:   strings.TrimRight(prefix, "/") + endpoint.suffix,
			Method: endpoint.method, ApiGroup: "proxy", Description: endpoint.desc,
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

	if err := tx.AutoMigrate(&proxyModel.Instance{}); err != nil {
		return err
	}
	// The existing Xray unit is managed by claude's lingering user systemd manager.
	if err := tx.Model(&proxyModel.Instance{}).Where("name = ? AND engine = ? AND scope = ?", "Xray", proxyModel.EngineXray, proxyModel.ScopeSystem).Update("scope", proxyModel.ScopeUser).Error; err != nil {
		return err
	}
	return nil
}
