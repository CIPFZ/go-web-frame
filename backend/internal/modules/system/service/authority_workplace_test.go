package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
)

// 角色的默认工作台：JSON 对象才收；null 恢复预置。
func TestSetAuthorityWorkplace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	pool, _ := db.DB()
	pool.SetMaxOpenConns(1)
	defer pool.Close()
	require.NoError(t, db.AutoMigrate(&model.SysAuthority{}))
	require.NoError(t, db.Create(&model.SysAuthority{AuthorityId: 888, AuthorityName: "user"}).Error)
	svc, ctx := NewAuthorityService(repository.NewAuthorityRepository(db)), context.Background()
	get := func() string {
		var r model.SysAuthority
		require.NoError(t, db.First(&r, "authority_id = ?", 888).Error)
		return string(r.Workplace)
	}

	desk := json.RawMessage(`{"v":1,"items":[{"id":"a","kind":"widget","widget":"notices","size":"m"}]}`)
	require.NoError(t, svc.SetAuthorityWorkplace(ctx, dto.SetAuthorityWorkplaceReq{AuthorityId: 888, Workplace: desk}))
	require.JSONEq(t, string(desk), get())
	require.ErrorIs(t, svc.SetAuthorityWorkplace(ctx, dto.SetAuthorityWorkplaceReq{AuthorityId: 888, Workplace: json.RawMessage(`[1]`)}), ErrInvalidWorkplace)
	require.Error(t, svc.SetAuthorityWorkplace(ctx, dto.SetAuthorityWorkplaceReq{AuthorityId: 9, Workplace: desk}), "unknown role")
	require.NoError(t, svc.SetAuthorityWorkplace(ctx, dto.SetAuthorityWorkplaceReq{AuthorityId: 888, Workplace: json.RawMessage(`null`)}))
	require.Empty(t, get(), "null goes back to the preset")
}

// 本人的工作台：只收不超过上限的 JSON 对象。
func TestValidWorkplace(t *testing.T) {
	require.True(t, validWorkplace([]byte(`{"v":1,"items":[]}`)))
	for _, bad := range []string{`[1]`, `null`, `"x"`, `{`, `{"pad":"` + string(make([]byte, maxWorkplaceBytes)) + `"}`} {
		require.False(t, validWorkplace([]byte(bad)), bad[:min(len(bad), 12)])
	}
}
