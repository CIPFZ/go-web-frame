package migrations

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func helperTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := cleanupTestDB(t)
	require.NoError(t, db.AutoMigrate(&schemaMigration{}, &claims.PolicyRevision{}))
	require.NoError(t, db.Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error)
	require.NoError(t, db.Exec("INSERT INTO sys_authorities (authority_id, authority_name, parent_id) VALUES (1, 'admin', 0), (888, 'user', 0)").Error)
	return db
}

func policyVersion(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	var rev claims.PolicyRevision
	require.NoError(t, db.First(&rev, 1).Error)
	return rev.Version
}

func TestRunStepsRunsEachStepOnceAndResumesAfterFailure(t *testing.T) {
	db := helperTestDB(t)
	cfg := &config.Config{}
	calls := map[string]int{}
	fail := true
	list := []step{
		{Version: "v1", Run: func(*gorm.DB, *config.Config, *zap.Logger) error { calls["v1"]++; return nil }},
		{Version: "v2", Run: func(*gorm.DB, *config.Config, *zap.Logger) error {
			calls["v2"]++
			if fail {
				return errors.New("boom")
			}
			return nil
		}},
		{Version: "v3", Run: func(*gorm.DB, *config.Config, *zap.Logger) error { calls["v3"]++; return nil }},
	}

	require.ErrorContains(t, runSteps(db, cfg, zap.NewNop(), list), "migration v2")
	done, err := applied(db, "v2")
	require.NoError(t, err)
	require.False(t, done, "a failed step is not recorded")

	fail = false
	require.NoError(t, runSteps(db, cfg, zap.NewNop(), list))
	require.NoError(t, runSteps(db, cfg, zap.NewNop(), list))
	require.Equal(t, map[string]int{"v1": 1, "v2": 2, "v3": 1}, calls)
	require.Equal(t, "v3", latestVersion(list))
	require.Equal(t, legacyVersion, latestVersion(nil))
}

func TestEnsureMenuKeepsAdminEditsAndGrantSkipsMissingRoles(t *testing.T) {
	db := helperTestDB(t)
	menu, created, err := ensureMenu(db, model.SysMenu{Path: "/orders", Name: "Orders", Component: "orders"})
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, db.Model(menu).Update("name", "My orders").Error)

	again, created, err := ensureMenu(db, model.SysMenu{Path: "/orders", Name: "Orders", Component: "orders"})
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, "My orders", again.Name)

	require.NoError(t, grantMenu(db, menu.ID, 1, 888, 4040))
	require.NoError(t, grantMenu(db, menu.ID, 1))
	var grants []uint
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("menu_id = ?", menu.ID).Order("authority_id").Pluck("authority_id", &grants).Error)
	require.Equal(t, []uint{1, 888}, grants)

	require.NoError(t, removeMenu(db, "/orders"))
	var n int64
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("menu_id = ?", menu.ID).Count(&n).Error)
	require.Zero(t, n)
}

func TestSyncAPICatalogGrantsNewKeepsRevocationsAndRemovesStale(t *testing.T) {
	db := helperTestDB(t)
	other := model.SysApi{Path: "/api/v1/other/list", Method: "GET", ApiGroup: "other"}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, claims.PolicyTransaction(db.Statement.Context, db, func(tx *gorm.DB) error { return grantAPI(tx, 888, other) }))
	// Token 白名单先以 external 分组登记的接口
	require.NoError(t, db.Create(&model.SysApi{Path: "/api/v1/orders/export", Method: "POST", ApiGroup: tokenSeedGroup}).Error)

	catalog := []catalogEntry{
		{Path: "/api/v1/orders/list", Method: "POST", Group: "orders", Name: "List orders", Authorities: []uint{1, 888}},
		{Path: "/api/v1/orders/delete", Method: "POST", Group: "orders", Name: "Delete orders", Authorities: []uint{1}},
		{Path: "/api/v1/orders/export", Method: "POST", Group: "orders", Name: "Export orders", Authorities: []uint{888}},
		{Path: "/api/v1/orders/:id", Method: "GET", Group: "orders", Name: "Get order", Authorities: []uint{888, 4040}},
	}
	before := policyVersion(t, db)
	require.NoError(t, syncAPICatalog(db, "/api/v1/orders/", catalog))
	require.Greater(t, policyVersion(t, db), before, "running instances reload the policy")

	rules := func(role string) []string {
		var rs []model.SysCasbinRule
		require.NoError(t, db.Where("ptype = 'p' AND v0 = ?", role).Order("v1, v2").Find(&rs).Error)
		out := make([]string, 0, len(rs))
		for _, r := range rs {
			out = append(out, r.V2+" "+r.V1)
		}
		return out
	}
	require.Equal(t, []string{"GET /api/v1/orders/:id", "POST /api/v1/orders/export", "POST /api/v1/orders/list", "GET /api/v1/other/list"}, rules("888"))

	// 管理员撤销了 888 的 list；接口下线了 delete；说明改了名字
	require.NoError(t, db.Where("v0 = '888' AND v1 = ?", "/api/v1/orders/list").Delete(&model.SysCasbinRule{}).Error)
	var deleteAPI model.SysApi
	require.NoError(t, db.Where("path = ?", "/api/v1/orders/delete").First(&deleteAPI).Error)
	require.NoError(t, db.Create(&model.SysApiTokenApi{ApiTokenId: 7, ApiId: deleteAPI.ID}).Error)
	catalog = append(catalog[:1], catalog[2:]...)
	catalog[0].Name = "List all orders"
	require.NoError(t, syncAPICatalog(db, "/api/v1/orders", catalog))

	require.Equal(t, []string{"GET /api/v1/orders/:id", "POST /api/v1/orders/export", "GET /api/v1/other/list"}, rules("888"), "revocation kept, other module untouched")
	require.Equal(t, []string{"POST /api/v1/orders/list"}, rules("1"))
	var n int64
	require.NoError(t, db.Unscoped().Model(&model.SysApi{}).Where("path = ?", "/api/v1/orders/delete").Count(&n).Error)
	require.Zero(t, n)
	require.NoError(t, db.Model(&model.SysApiTokenApi{}).Where("api_id = ?", deleteAPI.ID).Count(&n).Error)
	require.Zero(t, n)
	var list model.SysApi
	require.NoError(t, db.Where("path = ?", "/api/v1/orders/list").First(&list).Error)
	require.Equal(t, "List all orders", list.Description)

	roles, err := rolesWithAPI(db, "/api/v1/orders/:id", "GET")
	require.NoError(t, err)
	require.Equal(t, []uint{888}, roles)
}
