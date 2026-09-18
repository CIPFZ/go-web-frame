package migrations

import (
	"context"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestMagnetMigrationInstallsAdminMenuIdempotently(t *testing.T) {
	db := cleanupTestDB(t)
	t.Setenv("SEED_ADMIN_ENABLED", "true")
	t.Setenv("SEED_ADMIN_PASSWORD", "migration-password")
	cfg := &config.Config{System: config.System{Environment: "dev", RouterPrefix: "/api/v1"}}
	require.NoError(t, applyTokenNotice(db, cfg, zap.NewNop()))
	require.NoError(t, Run(context.Background(), db, cfg, zap.NewNop()))
	require.NoError(t, Run(context.Background(), db, cfg, zap.NewNop()))
	var menu model.SysMenu
	require.NoError(t, db.Where("path = ?", "/magnet/preview").First(&menu).Error)
	require.Equal(t, "magnet/preview", menu.Component)
	var grants []model.SysAuthorityMenu
	require.NoError(t, db.Where("menu_id = ?", menu.ID).Find(&grants).Error)
	require.Len(t, grants, 1)
	require.EqualValues(t, 1, grants[0].AuthorityId)
	var count int64
	require.NoError(t, db.Model(&model.SysApi{}).Where("api_group = ?", "magnet").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, Check(db))
	// A later run must preserve administrator edits and revoked grants.
	require.NoError(t, db.Model(&menu).Update("name", "My preview").Error)
	require.NoError(t, db.Where("menu_id = ?", menu.ID).Delete(&model.SysAuthorityMenu{}).Error)
	require.NoError(t, Run(context.Background(), db, cfg, zap.NewNop()))
	require.NoError(t, db.First(&menu, menu.ID).Error)
	require.Equal(t, "My preview", menu.Name)
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("menu_id = ?", menu.ID).Count(&count).Error)
	require.Zero(t, count)
}
