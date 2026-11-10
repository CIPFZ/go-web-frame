package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/CIPFZ/gowebframe/internal/core/apispec"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
)

func TestContractAuthenticators(t *testing.T) {
	_, svcCtx := newAPITokenMiddlewareTestEngine(t, func(*gin.RouterGroup) {})
	// 允许的 Token 获得 GetOrder；另一个 Token 没有授权
	api := model.SysApi{Path: "/api/v1/orders/orders/get", Method: "POST", ApiGroup: "orders"}
	require.NoError(t, svcCtx.DB.Create(&api).Error)
	var allowed model.SysApiToken
	require.NoError(t, svcCtx.DB.Where("name = ?", "allowed").First(&allowed).Error)
	require.NoError(t, svcCtx.DB.Model(&allowed).Association("Apis").Append(&api))

	engine := gin.New()
	r := apispec.NewRouter(engine.Group("/api/v1/orders"), apispec.RouterOptions{Authenticators: ContractAuthenticators(svcCtx)})
	read := apispec.MethodPolicy{Action: apispec.ActionRead, Sensitivity: apispec.SensitivityRead, Visibility: apispec.VisibilityPublic, Idempotency: apispec.IdempotencyNatural}
	whoami := func(c *gin.Context) { apispec.WriteOK(c, gin.H{"principal": apispec.PrincipalOf(c).String()}) }
	tokenOnly := read
	tokenOnly.Authentication = apispec.AuthToken
	r.Handle("orders", "get", "GetOrder", tokenOnly, whoami)
	anonymous := read
	anonymous.Authentication = apispec.AuthAnonymous
	r.Handle("orders", "list", "ListOrders", anonymous, whoami)

	call := func(path, token string) (int, map[string]any) {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		if token != "" {
			req.Header.Set("X-API-Token", token)
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		return w.Code, body
	}
	reason := func(body map[string]any) string { return body["error"].(map[string]any)["reason"].(string) }
	principal := func(body map[string]any) string { return body["data"].(map[string]any)["principal"].(string) }

	code, body := call("/api/v1/orders/orders/get", "")
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, "auth.credentials_missing", reason(body))
	code, body = call("/api/v1/orders/orders/get", "wrong")
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, "auth.credentials_invalid", reason(body))
	code, body = call("/api/v1/orders/orders/get", "cms_forbidden_token")
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, "auth.permission_denied", reason(body))
	code, body = call("/api/v1/orders/orders/get", "cms_allow_token")
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "token:1", principal(body))

	// 匿名可读：不带凭据、带着没有授权的 Token 都按匿名读取；无效的 Token 被拒绝
	code, body = call("/api/v1/orders/orders/list", "")
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "anonymous:192.0.2.1", principal(body))
	code, _ = call("/api/v1/orders/orders/list", "cms_forbidden_token")
	require.Equal(t, http.StatusOK, code)
	code, body = call("/api/v1/orders/orders/list", "wrong")
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, "auth.credentials_invalid", reason(body))
}
