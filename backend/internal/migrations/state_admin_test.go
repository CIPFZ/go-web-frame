package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestStateIsKeptOnlyForAdministrators(t *testing.T) {
	db := helperTestDB(t)
	for _, statement := range []string{
		"INSERT INTO sys_authorities (authority_id, authority_name, parent_id) VALUES (9528, 'ops', 0)",
		"INSERT INTO sys_menus (id, path, name, component) VALUES (1, '/sys', 'sys', 'components/RouterLayout'), (2, '/health', 'state', 'state')",
		"INSERT INTO sys_authority_menus (authority_id, menu_id) VALUES (1, 1), (1, 2), (9528, 1), (9528, 2), (888, 2)",
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/system/getServerInfo', 'POST')",
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (1, 1), (9528, 1), (888, 1)",
		"INSERT INTO sys_casbin_rules (ptype, v0, v1, v2) VALUES ('p', '1', '/api/v1/sys/system/getServerInfo', 'POST'), ('p', '9528', '/api/v1/sys/system/getServerInfo', 'POST'), ('p', '888', '/api/v1/sys/system/getServerInfo', 'POST')",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, applyStateAdminOnly(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()))
	var menus, apis, rules []string
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("menu_id = 2").Order("authority_id").Pluck("CAST(authority_id AS TEXT)", &menus).Error)
	require.NoError(t, db.Model(&model.SysAuthorityApi{}).Where("api_id = 1").Order("authority_id").Pluck("CAST(authority_id AS TEXT)", &apis).Error)
	require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v1 LIKE ?", "%getServerInfo").Order("v0").Pluck("v0", &rules).Error)
	require.Equal(t, []string{"1", "9528"}, menus)
	require.Equal(t, []string{"1", "9528"}, apis)
	require.Equal(t, []string{"1", "9528"}, rules)
}
