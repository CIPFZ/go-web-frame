package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestMultiRoleRemovesSwitchAndPrunesChildRoles(t *testing.T) {
	db := helperTestDB(t)
	for _, statement := range []string{
		"INSERT INTO sys_authorities (authority_id, authority_name, parent_id) VALUES (8881, 'child', 888)",
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/user/switchAuthority', 'POST'), (2, '/api/v1/a', 'GET'), (3, '/api/v1/b', 'GET')",
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (888, 1), (888, 2), (8881, 2), (8881, 3)",
		"INSERT INTO sys_casbin_rules (ptype, v0, v1, v2) VALUES ('p', '888', '/api/v1/sys/user/switchAuthority', 'POST'), ('p', '888', '/api/v1/a', 'GET'), ('p', '8881', '/api/v1/a', 'GET'), ('p', '8881', '/api/v1/b', 'GET')",
		"INSERT INTO sys_menus (id, path, name) VALUES (1, '/m1', 'm1'), (2, '/m2', 'm2')",
		"INSERT INTO sys_authority_menus (authority_id, menu_id) VALUES (888, 1), (8881, 1), (8881, 2)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	before := policyVersion(t, db)
	require.NoError(t, applyMultiRole(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()))
	require.Greater(t, policyVersion(t, db), before)

	var n int64
	require.NoError(t, db.Model(&model.SysApi{}).Where("path LIKE ?", "%switchAuthority").Count(&n).Error)
	require.Zero(t, n)
	require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v1 LIKE ?", "%switchAuthority").Count(&n).Error)
	require.Zero(t, n)
	var childRules []string
	require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v0 = '8881'").Pluck("v1", &childRules).Error)
	require.Equal(t, []string{"/api/v1/a"}, childRules, "the child keeps only what its parent has")
	var childMenus []uint
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("authority_id = 8881").Pluck("menu_id", &childMenus).Error)
	require.Equal(t, []uint{1}, childMenus)
}
