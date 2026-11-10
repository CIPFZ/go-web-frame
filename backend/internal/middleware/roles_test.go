package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

func rolesTestContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "roles.db")), &gorm.Config{})
	require.NoError(t, err)
	pool, _ := db.DB()
	t.Cleanup(func() { _ = pool.Close() })
	require.NoError(t, db.AutoMigrate(&model.SysAuthority{}, &model.SysUserAuthority{}, &model.SysCasbinRule{}, &claims.PolicyRevision{}))
	require.NoError(t, db.Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error)
	deleted := time.Now()
	for _, r := range []model.SysAuthority{{AuthorityId: 1}, {AuthorityId: 888}, {AuthorityId: 30}, {AuthorityId: 40, DeletedAt: &deleted}} {
		require.NoError(t, db.Create(&r).Error)
	}
	// 用户 7：888 + 30 + 已删除的 40；用户 8：超级管理员
	for _, ua := range []model.SysUserAuthority{{UserId: 7, AuthorityId: 888}, {UserId: 7, AuthorityId: 30}, {UserId: 7, AuthorityId: 40}, {UserId: 8, AuthorityId: 1}} {
		require.NoError(t, db.Create(&ua).Error)
	}
	for _, r := range []model.SysCasbinRule{{Ptype: "p", V0: "888", V1: "/a", V2: "GET"}, {Ptype: "p", V0: "30", V1: "/b", V2: "GET"}, {Ptype: "p", V0: "40", V1: "/c", V2: "GET"}} {
		require.NoError(t, db.Create(&r).Error)
	}
	policy, err := claims.NewPolicyManager(db)
	require.NoError(t, err)
	sc := svc.NewServiceContext()
	sc.DB, sc.Policy = db, policy
	return sc
}

func TestAllRolesOfAUserApplyAtOnce(t *testing.T) {
	sc := rolesTestContext(t)
	ctx := context.Background()
	roles, err := UserRoles(ctx, sc.DB, 7)
	require.NoError(t, err)
	require.Equal(t, []uint{30, 888}, roles, "deleted roles do not count")

	for path, want := range map[string]bool{"/a": true, "/b": true, "/c": false} {
		ok, err := UserAllows(ctx, sc, 7, path, "GET")
		require.NoError(t, err)
		require.Equal(t, want, ok, path)
	}
	ok, err := UserAllows(ctx, sc, 8, "/anything", "DELETE")
	require.NoError(t, err)
	require.True(t, ok, "the super admin is not checked against Casbin")
}

func TestCasbinHandlerUsesTheUnionOfRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sc := rolesTestContext(t)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// 主角色是 888；只有 30 能调用 /b
		c.Set(CtxKeyClaims, &claims.CustomClaims{BaseClaims: dto.BaseClaims{UserID: 7, AuthorityId: 888}})
		c.Next()
	}, CasbinHandler(sc))
	r.GET("/b", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/c", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/b", nil))
	require.Equal(t, http.StatusNoContent, w.Code)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/c", nil))
	require.NotEqual(t, http.StatusNoContent, w.Code)
}
