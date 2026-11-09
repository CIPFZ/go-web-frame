package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/extlogin"
	"github.com/CIPFZ/gowebframe/internal/core/jwt"
	"github.com/CIPFZ/gowebframe/internal/core/session"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

func externalLoginService(t *testing.T) (IUserService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "ext.db")), &gorm.Config{})
	require.NoError(t, err)
	pool, _ := db.DB()
	t.Cleanup(func() { _ = pool.Close() })
	require.NoError(t, db.AutoMigrate(&model.SysUser{}, &model.SysAuthority{}, &model.SysUserAuthority{}, &model.SysUserIdentity{},
		&session.Record{}, &claims.PolicyRevision{}))
	require.NoError(t, db.Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error)
	require.NoError(t, db.Create(&model.SysAuthority{AuthorityId: 888, AuthorityName: "user"}).Error)
	cfg := &config.Config{JWT: config.JWT{SigningKey: "a-long-signing-key-for-external-login", Issuer: "tests", ExpiresTime: "1h", BufferTime: "5m"}}
	sc := svc.NewServiceContext()
	sc.DB, sc.Config, sc.Logger = db, cfg, zap.NewNop()
	sc.Sessions = session.NewStore(db)
	sc.JWT = jwt.NewJWT(cfg.JWT, sc.Logger, nil)
	return NewUserService(sc, repository.NewUserRepository(db)), db
}

func TestExternalLoginLinksMatchesCreatesOrRefuses(t *testing.T) {
	users, db := externalLoginService(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	require.NoError(t, repo.Create(ctx, &model.SysUser{Username: "alice", Password: "x", Status: model.UserActive, AuthorityID: 888, Authorities: []model.SysAuthority{{AuthorityId: 888}}}))
	login := func(ticket string) uint {
		resp, err := users.RedeemLoginTicket(ctx, ticket)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Token)
		var id model.SysUserIdentity
		_ = db.Last(&id).Error
		return id.UserID
	}

	// 默认只接受关联过的身份：同名用户不会被冒用
	_, err := users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Subject: "s-1", Username: "alice"})
	require.ErrorIs(t, err, ErrExternalUsernameTaken)
	_, err = users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Subject: "s-9", Username: "nobody"})
	require.ErrorIs(t, err, ErrExternalNotLinked)
	_, err = users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Username: "alice"})
	require.Error(t, err, "an identity needs a subject")

	// 按用户名关联到已有用户；之后按关联登录，改名也不受影响
	ticket, err := users.ExternalLogin(ctx, "corp", extlogin.Options{MatchUsername: true}, &extlogin.Identity{Subject: "s-1", Username: "alice"})
	require.NoError(t, err)
	alice := login(ticket)
	ticket, err = users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Subject: "s-1", Username: "alice-renamed"})
	require.NoError(t, err)
	_, err = users.RedeemLoginTicket(ctx, ticket)
	require.NoError(t, err)
	_, err = users.RedeemLoginTicket(ctx, ticket)
	require.Error(t, err, "a ticket works once")

	// 自动建号：角色由提供方决定；同一身份第二次登录用同一个账号
	opt := extlogin.Options{AutoCreate: true, RoleID: 888}
	ticket, err = users.ExternalLogin(ctx, "corp", opt, &extlogin.Identity{Subject: "s-2", Username: "bob", NickName: "Bob", Email: "bob@example.com"})
	require.NoError(t, err)
	bob := login(ticket)
	require.NotEqual(t, alice, bob)
	var created model.SysUser
	require.NoError(t, db.Preload("Authorities").First(&created, bob).Error)
	require.Equal(t, "Bob", created.NickName)
	require.Len(t, created.Authorities, 1)
	require.EqualValues(t, 888, created.Authorities[0].AuthorityId)
	_, err = users.ExternalLogin(ctx, "corp", opt, &extlogin.Identity{Subject: "s-2", Username: "bob"})
	require.NoError(t, err)
	var count int64
	require.NoError(t, db.Model(&model.SysUser{}).Where("username = ?", "bob").Count(&count).Error)
	require.EqualValues(t, 1, count)
	// 另一个提供方的同一个标识是另一个身份
	_, err = users.ExternalLogin(ctx, "other", extlogin.Options{}, &extlogin.Identity{Subject: "s-2", Username: "bob2"})
	require.ErrorIs(t, err, ErrExternalNotLinked)

	// 被禁用的用户不能登录，已发出的票据也换不到凭证
	ticket, err = users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Subject: "s-2"})
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.SysUser{}).Where("id = ?", bob).Update("status", model.UserInactive).Error)
	_, err = users.RedeemLoginTicket(ctx, ticket)
	require.ErrorIs(t, err, ErrUserDisabled)
	_, err = users.ExternalLogin(ctx, "corp", extlogin.Options{}, &extlogin.Identity{Subject: "s-2"})
	require.ErrorIs(t, err, ErrUserDisabled)
}
