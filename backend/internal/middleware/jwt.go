package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/core/session"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxKeyClaims      = "claims"
	CtxKeyUserID      = "userId"
	CtxKeyUserUUID    = "userUUID"
	CtxKeyAuthorityId = "authorityId"
)

// AuthFailure 是认证 / 授权的结果分类，由不同协议的中间件各自转换为响应。
type AuthFailure int

const (
	AuthOK AuthFailure = iota
	AuthMissing
	AuthRevoked
	AuthExpired
	AuthInvalid
	AuthSessionInvalid
	AuthDisabled
	AuthForbidden
	AuthQuota
	AuthUnavailable
	AuthInternal
)

// AuthenticateSession 校验会话令牌，临近过期时续期，并写入用户上下文。
func AuthenticateSession(c *gin.Context, svcCtx *svc.ServiceContext) (*claims.CustomClaims, AuthFailure) {
	token := c.Request.Header.Get("x-token")
	if token == "" {
		token, _ = c.Cookie("x-token")
	}
	if token == "" {
		return nil, AuthMissing
	}
	if svcCtx.JWT.IsBlacklist(c.Request.Context(), token) {
		return nil, AuthRevoked
	}

	j := svcCtx.JWT
	cl, err := j.ParseToken(token)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, AuthExpired
		}
		return nil, AuthInvalid
	}

	identity := session.Identity{ID: cl.ID, UserID: cl.UserID, Version: cl.TokenVersion, AuthorityID: cl.AuthorityId}
	if svcCtx.Sessions == nil || svcCtx.Sessions.Validate(c.Request.Context(), identity) != nil {
		return nil, AuthSessionInvalid
	}
	if cl.ExpiresAt.Unix()-time.Now().Unix() < cl.BufferTime {
		dr, _ := utils.ParseDuration(svcCtx.Config.JWT.ExpiresTime)
		newToken, renewed, err := j.ResolveToken(c.Request.Context(), token, cl)
		if err == nil && newToken != "" {
			if err := svcCtx.Sessions.Renew(c.Request.Context(), identity, renewed.ExpiresAt.Time); err != nil {
				return nil, AuthUnavailable
			}
			c.Header("new-token", newToken)
			c.Header("x-token", newToken)
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie("x-token", newToken, int(dr.Seconds()), "/", "", false, true)
		}
	}

	c.Set(CtxKeyClaims, cl)
	c.Set(CtxKeyUserID, cl.UserID)
	c.Set(CtxKeyUserUUID, cl.UUID)
	c.Set(CtxKeyAuthorityId, cl.AuthorityId)
	return cl, AuthOK
}

// AuthorizeRoute 校验用户对当前路由的访问权限：用户的所有角色同时生效，有一个角色允许即可。
func AuthorizeRoute(c *gin.Context, svcCtx *svc.ServiceContext, cl *claims.CustomClaims) AuthFailure {
	ok, err := UserAllows(c.Request.Context(), svcCtx, cl.UserID, c.Request.URL.Path, c.Request.Method)
	if err != nil {
		return AuthInternal
	}
	if !ok {
		return AuthForbidden
	}
	return AuthOK
}

func JWTAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, failure := AuthenticateSession(c, svcCtx); failure != AuthOK {
			response.FailWithCode(sessionError(failure), c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// sessionError 把会话认证失败转换为 {code,data,msg} 接口的错误。
func sessionError(f AuthFailure) *errcode.Error {
	switch f {
	case AuthRevoked:
		return errcode.Unauthorized.WithDetails("令牌已失效")
	case AuthExpired:
		return errcode.Unauthorized.WithDetails("授权已过期")
	case AuthInvalid:
		return errcode.Unauthorized.WithDetails("无效的令牌")
	case AuthSessionInvalid:
		return errcode.Unauthorized.WithDetails("会话已失效，请重新登录")
	default:
		return errcode.Unauthorized
	}
}
