package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/CIPFZ/gowebframe/internal/core/apispec"
	"github.com/CIPFZ/gowebframe/internal/svc"
)

// 契约接口（core/apispec）的认证：复用 CMS 会话与 API Token 的校验，失败时返回契约错误。
// 会话按用户的全部角色授权当前路由；Token 按 Token 的权限 ∩ 主人现在的权限。

// contractFailure 把认证结果转换为契约错误。
func contractFailure(f AuthFailure) *apispec.Error {
	switch f {
	case AuthMissing:
		return apispec.Unauthorized("auth.credentials_missing", "authentication is required")
	case AuthExpired:
		return apispec.Unauthorized("auth.credentials_expired", "credentials have expired")
	case AuthRevoked, AuthInvalid, AuthSessionInvalid:
		return apispec.Unauthorized("auth.credentials_invalid", "credentials are invalid")
	case AuthDisabled:
		return apispec.Unauthorized("auth.token_disabled", "API token is disabled")
	case AuthForbidden:
		return apispec.PermissionDenied("auth.permission_denied", "caller is not allowed to call this method")
	case AuthQuota:
		return apispec.QuotaExceeded("auth.concurrency_exceeded", "API token concurrency limit reached", 1)
	case AuthUnavailable:
		return apispec.Unavailable("auth.limiter_unavailable", nil).WithRetryAfter(1)
	default:
		return apispec.Internal(nil)
	}
}

func hasAPIToken(c *gin.Context) bool { return c.GetHeader("X-API-Token") != "" }

func hasSession(c *gin.Context) bool {
	if c.GetHeader("x-token") != "" {
		return true
	}
	cookie, _ := c.Cookie("x-token")
	return cookie != ""
}

// ContractAuthenticators 返回契约接口各调用身份的认证中间件，交给 apispec.RouterOptions.Authenticators。
func ContractAuthenticators(svcCtx *svc.ServiceContext) map[apispec.Authentication]gin.HandlerFunc {
	user, token := contractUser(svcCtx), contractToken(svcCtx)
	return map[apispec.Authentication]gin.HandlerFunc{
		apispec.AuthUser:  user,
		apispec.AuthToken: token,
		apispec.AuthUserOrToken: func(c *gin.Context) {
			if hasAPIToken(c) {
				token(c)
				return
			}
			user(c)
		},
		apispec.AuthAnonymous: contractOptional(svcCtx),
	}
}

// contractUser 校验 CMS 会话并按 Casbin 授权当前路由。
func contractUser(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl, f := AuthenticateSession(c, svcCtx)
		if f == AuthOK {
			f = AuthorizeRoute(c, svcCtx, cl)
		}
		if f != AuthOK {
			apispec.AbortWithError(c, contractFailure(f))
			return
		}
		apispec.SetPrincipal(c, apispec.Principal{Kind: apispec.PrincipalUser, ID: cl.UUID.String()})
		c.Next()
	}
}

// contractToken 校验 API Token（接口授权、并发租约），请求结束后释放租约。
func contractToken(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, release, f := AuthenticateAPIToken(c, svcCtx)
		if f != AuthOK {
			apispec.AbortWithError(c, contractFailure(f))
			return
		}
		defer release()
		apispec.SetPrincipal(c, apispec.Principal{Kind: apispec.PrincipalToken, ID: strconv.FormatUint(uint64(token.ID), 10)})
		c.Next()
		TouchAPIToken(c, svcCtx, token.ID)
	}
}

// contractOptional 用于匿名可读的方法：携带的 API Token 必须有效（有效但没有授权这个接口时按匿名读取）；
// 携带的会话有效时识别调用方，无效时按匿名处理。
func contractOptional(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasAPIToken(c) {
			token, release, f := AuthenticateAPIToken(c, svcCtx)
			switch f {
			case AuthOK:
				defer release()
				apispec.SetPrincipal(c, apispec.Principal{Kind: apispec.PrincipalToken, ID: strconv.FormatUint(uint64(token.ID), 10)})
				c.Next()
				TouchAPIToken(c, svcCtx, token.ID)
			case AuthForbidden:
				c.Next()
			default:
				apispec.AbortWithError(c, contractFailure(f))
			}
			return
		}
		if hasSession(c) {
			if cl, f := AuthenticateSession(c, svcCtx); f == AuthOK {
				apispec.SetPrincipal(c, apispec.Principal{Kind: apispec.PrincipalUser, ID: cl.UUID.String()})
			}
		}
		c.Next()
	}
}

// ContractAudit 记录契约接口写操作的操作日志：模块名与“调用方 方法名”写入日志的模块与备注。交给 apispec.RouterOptions.Audit。
func ContractAudit(svcCtx *svc.ServiceContext, module string) gin.HandlerFunc {
	record := OperationRecord(svcCtx)
	return func(c *gin.Context) {
		c.Set(CtxKeyAuditModule, module)
		c.Set(CtxKeyAuditRemark, apispec.PrincipalOf(c).String()+" "+apispec.MethodName(c))
		record(c)
	}
}
