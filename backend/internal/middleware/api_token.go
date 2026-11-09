package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
)

const (
	CtxKeyAPITokenID = "apiTokenId"
)

// AuthenticateAPIToken 校验 X-API-Token、启用状态、有效期、接口授权并获取并发租约。
// 成功时返回的 release 必须在请求处理结束后调用。
func AuthenticateAPIToken(c *gin.Context, svcCtx *svc.ServiceContext) (*model.SysApiToken, func(), AuthFailure) {
	if svcCtx.APITokenLimiter == nil {
		svcCtx.APITokenLimiter = tokenCore.NewInMemoryLimiter()
	}
	rawToken := strings.TrimSpace(c.GetHeader("X-API-Token"))
	if rawToken == "" {
		return nil, nil, AuthMissing
	}

	var token model.SysApiToken
	err := svcCtx.DB.WithContext(c.Request.Context()).
		Preload("Apis").
		Where("token_hash = ?", tokenCore.HashToken(rawToken)).
		First(&token).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, AuthInvalid
		}
		return nil, nil, AuthInternal
	}
	if !token.Enabled {
		return nil, nil, AuthDisabled
	}
	if token.ExpiresAt != nil && !token.ExpiresAt.After(time.Now()) {
		return nil, nil, AuthExpired
	}

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	if failure := authorizeToken(c, svcCtx, &token, path); failure != AuthOK {
		return nil, nil, failure
	}

	release := func() {}
	if svcCtx.Redis != nil {
		ctx, done, err := tokenCore.AcquireRedisLease(c.Request.Context(), svcCtx.Redis, token.ID, token.MaxConcurrency)
		if err != nil {
			if errors.Is(err, tokenCore.ErrQuota) {
				return nil, nil, AuthQuota
			}
			return nil, nil, AuthUnavailable
		}
		c.Request = c.Request.WithContext(ctx)
		release = done
	} else {
		if !svcCtx.APITokenLimiter.Acquire(token.ID, token.MaxConcurrency) {
			return nil, nil, AuthQuota
		}
		release = func() { svcCtx.APITokenLimiter.Release(token.ID) }
	}

	c.Set(CtxKeyAPITokenID, token.ID)
	c.Set(CtxKeyUserID, token.CreatedBy) // 通过 Token 的操作记在主人名下（操作日志等）
	return &token, release, AuthOK
}

// TouchAPIToken 记录 Token 最近使用时间。
func TouchAPIToken(c *gin.Context, svcCtx *svc.ServiceContext, tokenID uint) {
	_ = svcCtx.DB.WithContext(c.Request.Context()).
		Model(&model.SysApiToken{}).
		Where("id = ?", tokenID).
		Update("last_used_at", time.Now()).
		Error
}

func ApiTokenAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, release, failure := AuthenticateAPIToken(c, svcCtx)
		switch failure {
		case AuthOK:
		case AuthMissing, AuthInvalid:
			response.FailWithCode(errcode.Unauthorized, c)
			c.Abort()
			return
		case AuthDisabled:
			response.FailWithCode(errcode.Unauthorized.WithDetails("token 已禁用"), c)
			c.Abort()
			return
		case AuthExpired:
			response.FailWithCode(errcode.Unauthorized.WithDetails("token 已过期"), c)
			c.Abort()
			return
		case AuthForbidden:
			response.FailWithCode(errcode.AssessDenied.WithDetails("token 无权访问该 API"), c)
			c.Abort()
			return
		case AuthQuota:
			c.AbortWithStatusJSON(429, response.Response{Code: 429, Msg: response.LocalizeMessage(c, "token.quota")})
			return
		case AuthUnavailable:
			c.AbortWithStatusJSON(503, response.Response{Code: 503, Msg: response.LocalizeMessage(c, "token.limiterUnavailable")})
			return
		default:
			response.FailWithCode(errcode.ServerError, c)
			c.Abort()
			return
		}
		defer release()
		c.Next()
		TouchAPIToken(c, svcCtx, token.ID)
	}
}

// authorizeToken：Token 跟主人绑定。账号被禁用或删除时 Token 失效；能调用的接口 = Token 的权限 ∩ 主人现在的权限。
func authorizeToken(c *gin.Context, svcCtx *svc.ServiceContext, token *model.SysApiToken, path string) AuthFailure {
	ctx, method := c.Request.Context(), c.Request.Method
	var active int64
	if err := svcCtx.DB.WithContext(ctx).Model(&model.SysUser{}).Where("id = ? AND status = ?", token.CreatedBy, model.UserActive).Count(&active).Error; err != nil {
		return AuthInternal
	}
	if active == 0 {
		return AuthDisabled
	}
	ok, err := TokenAllows(ctx, svcCtx, token, path, method)
	if err != nil {
		return AuthInternal
	}
	if !ok {
		return AuthForbidden
	}
	return AuthOK
}

// TokenAllows 判断 Token 能否调用某接口：Token 的权限 ∩ 主人现在的权限。
func TokenAllows(ctx context.Context, svcCtx *svc.ServiceContext, token *model.SysApiToken, path, method string) (bool, error) {
	prefix := svcCtx.Config.System.RouterPrefix
	if !apiTokenAllowsRequest(token, prefix, method, path) {
		return false, nil
	}
	if path == tokenCore.TokenInfoPath(prefix) { // 只返回 Token 自身的信息，不属于任何角色的权限
		return true, nil
	}
	return UserAllows(ctx, svcCtx, token.CreatedBy, path, method)
}

func apiTokenAllowsRequest(token *model.SysApiToken, prefix, method, path string) bool {
	if token.FollowOwner {
		return tokenCore.AllowsEndpoint(prefix, method, path)
	}
	for _, api := range token.Apis {
		if strings.EqualFold(api.Method, method) && api.Path == path {
			return true
		}
	}
	return false
}
