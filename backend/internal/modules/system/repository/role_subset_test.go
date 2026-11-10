package repository

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

// 子角色只能拿父角色已有的接口与菜单；父角色收回的，各级子角色一起收回。
func TestChildRoleIsSubsetOfParent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	pool, _ := db.DB()
	pool.SetMaxOpenConns(1)
	defer pool.Close()
	require.NoError(t, db.AutoMigrate(&model.SysAuthority{}, &model.SysApi{}, &model.SysCasbinRule{}, &model.SysAuthorityApi{},
		&model.SysMenu{}, &model.SysAuthorityMenu{}, &claims.PolicyRevision{}))
	require.NoError(t, db.Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error)
	for _, r := range []model.SysAuthority{{AuthorityId: 10, AuthorityName: "父"}, {AuthorityId: 11, AuthorityName: "子", ParentId: 10},
		{AuthorityId: 12, AuthorityName: "孙", ParentId: 11}, {AuthorityId: 1, AuthorityName: "超级管理员"},
		{AuthorityId: 20, AuthorityName: "超级管理员的子角色", ParentId: 1}} {
		require.NoError(t, db.Create(&r).Error)
	}
	for _, p := range []string{"/a", "/b", "/c"} {
		require.NoError(t, db.Create(&model.SysApi{Path: p, Method: "POST"}).Error)
	}
	menus := []model.SysMenu{{Path: "/m1", Name: "m1"}, {Path: "/m2", Name: "m2"}}
	require.NoError(t, db.Create(&menus).Error)
	ctx, casbin, auth := context.Background(), NewCasbinRepository(db), NewAuthorityRepository(db)
	rules := func(role string, paths ...string) [][]string {
		out := [][]string{}
		for _, p := range paths {
			out = append(out, []string{role, p, "POST"})
		}
		return out
	}
	policy := func(role string) []string {
		got, err := casbin.GetPolicy(ctx, role)
		require.NoError(t, err)
		out := []string{}
		for _, r := range got {
			out = append(out, r[1])
		}
		return out
	}

	require.NoError(t, casbin.ReplacePolicy(ctx, "10", rules("10", "/a", "/b")))
	require.ErrorContains(t, casbin.ReplacePolicy(ctx, "11", rules("11", "/a", "/c")), "父角色", "/c is not the parent's")
	require.NoError(t, casbin.ReplacePolicy(ctx, "11", rules("11", "/a", "/b")))
	require.NoError(t, casbin.ReplacePolicy(ctx, "12", rules("12", "/b")))
	require.NoError(t, auth.SetMenuAuthority(ctx, 10, []uint{menus[0].ID}))
	require.ErrorContains(t, auth.SetMenuAuthority(ctx, 11, []uint{menus[0].ID, menus[1].ID}), "父角色", "m2 is not the parent's")
	require.NoError(t, auth.SetMenuAuthority(ctx, 10, []uint{menus[0].ID, menus[1].ID}))
	require.NoError(t, auth.SetMenuAuthority(ctx, 11, []uint{menus[0].ID, menus[1].ID}))
	require.NoError(t, auth.SetMenuAuthority(ctx, 12, []uint{menus[1].ID}))

	// 父角色收回 /b 与 m2：子角色、孙角色一起收回
	require.NoError(t, casbin.ReplacePolicy(ctx, "10", rules("10", "/a")))
	require.Equal(t, []string{"/a"}, policy("11"))
	require.Empty(t, policy("12"))
	var apis int64
	require.NoError(t, db.Model(&model.SysAuthorityApi{}).Where("authority_id IN ?", []uint{11, 12}).Count(&apis).Error)
	require.EqualValues(t, 1, apis)
	require.NoError(t, auth.SetMenuAuthority(ctx, 10, []uint{menus[0].ID}))
	var left []uint
	require.NoError(t, db.Model(&model.SysAuthorityMenu{}).Where("authority_id IN ?", []uint{11, 12}).Pluck("menu_id", &left).Error)
	require.Equal(t, []uint{menus[0].ID}, left)

	// 超级管理员拥有全部：它的子角色可以选任何接口与菜单
	require.NoError(t, casbin.ReplacePolicy(ctx, "20", rules("20", "/a", "/b", "/c")))
	require.NoError(t, auth.SetMenuAuthority(ctx, 20, []uint{menus[0].ID, menus[1].ID}))
}

// 发给某角色的通知也发给它的各级子角色（只算启用的用户）。
func TestNoticeRecipientsIncludeChildRoles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	pool, _ := db.DB()
	pool.SetMaxOpenConns(1)
	defer pool.Close()
	require.NoError(t, db.AutoMigrate(&model.SysAuthority{}, &model.SysUser{}, &model.SysUserAuthority{}))
	for _, r := range []model.SysAuthority{{AuthorityId: 10}, {AuthorityId: 11, ParentId: 10}, {AuthorityId: 12, ParentId: 11}, {AuthorityId: 30}} {
		require.NoError(t, db.Create(&r).Error)
	}
	for i, role := range []uint{10, 11, 12, 30, 12} {
		u := model.SysUser{Username: string(rune('a' + i)), Status: model.UserActive}
		require.NoError(t, db.Create(&u).Error)
		if i == 4 {
			require.NoError(t, db.Model(&u).Update("status", model.UserInactive).Error)
		}
		require.NoError(t, db.Create(&model.SysUserAuthority{UserId: u.ID, AuthorityId: role}).Error)
	}
	targets := make([]uint, 1, 4)
	targets[0] = 11
	ids, err := NewNoticeRepository(db).ListUserIDsByAuthorityIDs(context.Background(), targets)
	require.NoError(t, err)
	require.ElementsMatch(t, []uint{2, 3}, ids)
	require.Equal(t, []uint{11}, targets, "the caller's slice is not modified")
}
