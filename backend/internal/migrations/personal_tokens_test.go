package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestPersonalTokensGiveOwnersAndSplitManagement(t *testing.T) {
	db := helperTestDB(t)
	for _, statement := range []string{
		"INSERT INTO sys_users (id, username, password, authority_id) VALUES (5, 'admin', 'x', 1), (6, 'member', 'x', 888)",
		"INSERT INTO sys_user_authorities (user_id, authority_id) VALUES (5, 1), (6, 888)",
		"INSERT INTO sys_apis (id, path, method) VALUES (1, '/api/v1/sys/api-token/getApiTokenList', 'POST'), (2, '/api/v1/sys/menu/getMenu', 'GET')",
		// 新装时普通角色也有 getApiTokenList（我的 Token）；只有 API Token 菜单的角色是管理员
		"INSERT INTO sys_authority_apis (authority_id, api_id) VALUES (1, 1), (1, 2), (888, 1), (888, 2)",
		"INSERT INTO sys_menus (id, path, name, component) VALUES (10, '/sys/api-token', 'tokens', 'sys/api-token')",
		"INSERT INTO sys_authority_menus (authority_id, menu_id) VALUES (1, 10)",
		"INSERT INTO sys_api_tokens (id, name, token_hash, created_by) VALUES (1, 'legacy', 'h1', 0), (2, 'orphan', 'h2', 404), (3, 'mine', 'h3', 6)",
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	require.NoError(t, applyPersonalTokens(db, &config.Config{System: config.System{RouterPrefix: "/api/v1"}}, zap.NewNop()))

	owners := map[string]uint{}
	var tokens []model.SysApiToken
	require.NoError(t, db.Find(&tokens).Error)
	for _, tk := range tokens {
		owners[tk.Name] = tk.CreatedBy
	}
	require.Equal(t, map[string]uint{"legacy": 5, "orphan": 5, "mine": 6}, owners)

	has := func(role, path, method string) bool {
		var n int64
		require.NoError(t, db.Model(&model.SysCasbinRule{}).Where("v0 = ? AND v1 = ? AND v2 = ?", role, path, method).Count(&n).Error)
		return n > 0
	}
	require.True(t, has("1", "/api/v1/sys/api-token/all", "POST"), "roles with the API Token page manage everyone's tokens")
	require.True(t, has("1", "/api/v1/sys/api-token/limits/update", "PUT"))
	require.False(t, has("888", "/api/v1/sys/api-token/all", "POST"))
	require.True(t, has("888", "/api/v1/sys/api-token/create", "POST"), "everyone manages their own tokens")
	require.True(t, has("888", "/api/v1/sys/api-token/options", "POST"))
	var menu model.SysMenu
	require.NoError(t, db.Where("path = ?", "/account/tokens").First(&menu).Error)
	var grants int64
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("menu_id = ?", menu.ID).Count(&grants).Error)
	require.EqualValues(t, 2, grants)
	require.False(t, has("888", "/api/v1/sys/api-token/limits/get", "POST"))
}
