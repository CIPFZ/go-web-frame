package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestRemoveBusinessModulesKeepsTheFramework(t *testing.T) {
	db := helperTestDB(t)
	for _, statement := range []string{
		"INSERT INTO sys_menus (id, path, name, parent_id) VALUES (1, '/sys', 'sys', 0), (2, '/virtualization/vms', 'vms', 0), (3, '/virtualization/agent', 'agent', 2), (4, '/novel/books', 'books', 0)",
		"INSERT INTO sys_authority_menus (authority_id, menu_id) VALUES (1, 1), (1, 2), (1, 3), (1, 4), (888, 4)",
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/user/getUserList', 'POST'), (2, '/api/v1/virtualization/vms', 'GET'), (3, '/api/v1/novel/book/list', 'POST')",
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (1, 1), (1, 2), (1, 3)",
		"INSERT INTO sys_api_token_apis (api_token_id, api_id) VALUES (5, 3)",
		"INSERT INTO sys_casbin_rules (ptype, v0, v1, v2) VALUES ('p', '1', '/api/v1/sys/user/getUserList', 'POST'), ('p', '1', '/api/v1/virtualization/vms', 'GET'), ('p', '1', '/api/v1/novel/book/list', 'POST')",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	before := policyVersion(t, db)
	cfg := &config.Config{System: config.System{RouterPrefix: "/api/v1"}}
	require.NoError(t, applyRemoveBusinessModules(db, cfg, zap.NewNop()))
	require.NoError(t, applyRemoveBusinessModules(db, cfg, zap.NewNop()), "repeatable")
	require.Greater(t, policyVersion(t, db), before)

	var menus, apis, rules []string
	require.NoError(t, db.Unscoped().Model(&model.SysMenu{}).Pluck("path", &menus).Error)
	require.NoError(t, db.Unscoped().Model(&model.SysApi{}).Pluck("path", &apis).Error)
	require.NoError(t, db.Model(&model.SysCasbinRule{}).Pluck("v1", &rules).Error)
	require.Equal(t, []string{"/sys"}, menus)
	require.Equal(t, []string{"/api/v1/sys/user/getUserList"}, apis)
	require.Equal(t, []string{"/api/v1/sys/user/getUserList"}, rules)
	var grants, tokenGrants int64
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Count(&grants).Error)
	require.NoError(t, db.Model(&model.SysApiTokenApi{}).Count(&tokenGrants).Error)
	require.EqualValues(t, 1, grants)
	require.Zero(t, tokenGrants)
}
