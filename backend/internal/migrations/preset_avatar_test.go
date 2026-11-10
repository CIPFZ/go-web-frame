package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestPresetAvatarGrantedToRolesThatCanUploadAvatars(t *testing.T) {
	db := helperTestDB(t)
	for _, statement := range []string{
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/user/avatar', 'POST')",
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (888, 1)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, applyPresetAvatar(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()))
	var roles []string
	require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v1 = ? AND v2 = 'PUT'", "/api/v1/sys/user/avatar/preset").Pluck("v0", &roles).Error)
	require.Equal(t, []string{"888"}, roles)
}
