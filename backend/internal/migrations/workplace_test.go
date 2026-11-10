package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestWorkplaceGrantsFollowExistingPermissions(t *testing.T) {
	db := helperTestDB(t)
	require.NoError(t, db.Migrator().DropColumn(&model.SysUser{}, "workplace"))
	require.NoError(t, db.Migrator().DropColumn(&model.SysAuthority{}, "workplace"))
	for _, statement := range []string{
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/user/ui-config', 'PUT'), (2, '/api/v1/sys/authority/updateAuthority', 'PUT')",
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (1, 1), (1, 2), (888, 1)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, applyWorkplace(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()))
	require.NoError(t, applyWorkplace(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()), "repeatable")
	require.True(t, db.Migrator().HasColumn(&model.SysUser{}, "workplace"))
	require.True(t, db.Migrator().HasColumn(&model.SysAuthority{}, "workplace"))

	roles := func(path string) []string {
		var out []string
		require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v1 = ? AND v2 = 'PUT'", path).Order("v0").Pluck("v0", &out).Error)
		return out
	}
	require.Equal(t, []string{"1", "888"}, roles("/api/v1/sys/user/workplace"))
	require.Equal(t, []string{"1"}, roles("/api/v1/sys/authority/workplace"))
}
