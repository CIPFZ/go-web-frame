package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/extlogin"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

type testSSO struct{}

func (testSSO) Key() string  { return "test-sso" }
func (testSSO) Name() string { return "Test SSO" }
func (testSSO) LoginURL(_ context.Context, callback string) (string, error) {
	return "https://sso.example.com/login?" + url.Values{"return": {callback}}.Encode(), nil
}
func (testSSO) Exchange(_ context.Context, q url.Values) (*extlogin.Identity, error) {
	return &extlogin.Identity{Subject: "sub-" + q.Get("code"), Username: "u-" + q.Get("code")}, nil
}

type extUsers struct {
	service.IUserService
	err      error
	identity *extlogin.Identity
}

func (u *extUsers) ExternalLogin(_ context.Context, provider string, _ extlogin.Options, id *extlogin.Identity) (string, error) {
	u.identity = id
	return "tkt", u.err
}

func init() { extlogin.Register(testSSO{}, extlogin.Options{}) }

func TestExternalLoginFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := &extUsers{}
	api := &UserApi{userService: users, svcCtx: &svc.ServiceContext{Config: &config.Config{System: config.System{RouterPrefix: "/api/v1"}}}}
	r := gin.New()
	r.GET("/api/v1/user/ext-login/:provider", api.ExternalLogin)
	r.GET("/api/v1/user/ext-callback/:provider", api.ExternalCallback)
	get := func(path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "cms.example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// 第一步：跳到提供方，回调地址带着 state 与站内的 redirect，Cookie 里也有同一个 state
	w := get("/api/v1/user/ext-login/test-sso?redirect=" + url.QueryEscape("/sys/user"))
	require.Equal(t, http.StatusFound, w.Code)
	loc, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "sso.example.com", loc.Host)
	callback, err := url.Parse(loc.Query().Get("return"))
	require.NoError(t, err)
	require.Equal(t, "https://cms.example.com/api/v1/user/ext-callback/test-sso", callback.Scheme+"://"+callback.Host+callback.Path)
	require.Equal(t, "/sys/user", callback.Query().Get("redirect"))
	state := callback.Query().Get("state")
	require.Len(t, state, 32)
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == extStateCookie {
			cookie = c
		}
	}
	require.NotNil(t, cookie)
	require.Equal(t, state, cookie.Value)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, "/api/v1/user/ext-callback/test-sso", cookie.Path)

	// 站外的 redirect 被丢掉
	w = get("/api/v1/user/ext-login/test-sso?redirect=" + url.QueryEscape("//evil.example.com"))
	loc, _ = url.Parse(w.Header().Get("Location"))
	callback, _ = url.Parse(loc.Query().Get("return"))
	require.Empty(t, callback.Query().Get("redirect"))

	back := func(w *httptest.ResponseRecorder) url.Values {
		require.Equal(t, http.StatusFound, w.Code)
		loc := w.Header().Get("Location")
		require.True(t, strings.HasPrefix(loc, "/#/user/login?"), loc)
		q, err := url.ParseQuery(strings.TrimPrefix(loc, "/#/user/login?"))
		require.NoError(t, err)
		return q
	}
	cb := "/api/v1/user/ext-callback/test-sso?code=42&redirect=%2Fsys%2Fuser&state=" + state
	// 没有 Cookie 或 state 不一致：拒绝，不去换身份
	require.Equal(t, "state", back(get(cb)).Get("ext_error"))
	require.Equal(t, "state", back(get(strings.Replace(cb, state, strings.Repeat("0", 32), 1), cookie)).Get("ext_error"))
	require.Nil(t, users.identity)
	require.Equal(t, "unknown", back(get("/api/v1/user/ext-login/nope")).Get("ext_error"))

	// 一致：换回身份、发票据，回到登录页时带着原来要去的页面
	q := back(get(cb, cookie))
	require.Equal(t, "tkt", q.Get("ext_ticket"))
	require.Equal(t, "/sys/user", q.Get("redirect"))
	require.Equal(t, "sub-42", users.identity.Subject)

	users.err = service.ErrExternalNotLinked
	require.Equal(t, "not_linked", back(get(cb, cookie)).Get("ext_error"))
	users.err = service.ErrUserDisabled
	require.Equal(t, "user_disabled", back(get(cb, cookie)).Get("ext_error"))
}
