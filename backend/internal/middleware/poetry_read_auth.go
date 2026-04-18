package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func PoetryReadAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader("X-API-Token")) != "" {
			handlePoetryReadWithAPIToken(svcCtx, c)
			return
		}
		handlePoetryReadWithJWT(svcCtx, c)
	}
}

func handlePoetryReadWithAPIToken(svcCtx *svc.ServiceContext, c *gin.Context) {
	if svcCtx.APITokenLimiter == nil {
		svcCtx.APITokenLimiter = tokenCore.NewInMemoryLimiter()
	}

	rawToken := strings.TrimSpace(c.GetHeader("X-API-Token"))
	if rawToken == "" {
		response.FailWithCode(errcode.Unauthorized, c)
		c.Abort()
		return
	}

	var token model.SysApiToken
	err := svcCtx.DB.WithContext(c.Request.Context()).
		Preload("Apis").
		Where("token_hash = ?", tokenCore.HashToken(rawToken)).
		First(&token).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithCode(errcode.Unauthorized, c)
			c.Abort()
			return
		}
		response.FailWithCode(errcode.ServerError, c)
		c.Abort()
		return
	}

	if !token.Enabled {
		response.FailWithCode(errcode.Unauthorized.WithDetails("token 已禁用"), c)
		c.Abort()
		return
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		response.FailWithCode(errcode.Unauthorized.WithDetails("token 已过期"), c)
		c.Abort()
		return
	}

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	if !apiTokenAllowsRequest(&token, c.Request.Method, path) {
		response.FailWithCode(errcode.AssessDenied.WithDetails("token 无权访问该 API"), c)
		c.Abort()
		return
	}

	if !svcCtx.APITokenLimiter.Acquire(token.ID, token.MaxConcurrency) {
		response.FailWithCode(errcode.AssessDenied.WithDetails("token 并发已达上限"), c)
		c.Abort()
		return
	}
	defer svcCtx.APITokenLimiter.Release(token.ID)

	c.Set(CtxKeyAPITokenID, token.ID)
	c.Next()

	_ = svcCtx.DB.WithContext(c.Request.Context()).
		Model(&model.SysApiToken{}).
		Where("id = ?", token.ID).
		Update("last_used_at", time.Now()).
		Error
}

func handlePoetryReadWithJWT(svcCtx *svc.ServiceContext, c *gin.Context) {
	tokenString := c.Request.Header.Get("x-token")
	if tokenString == "" {
		tokenString, _ = c.Cookie("x-token")
	}
	if tokenString == "" {
		response.FailWithCode(errcode.Unauthorized, c)
		c.Abort()
		return
	}

	if svcCtx.JWT.IsBlacklist(c.Request.Context(), tokenString) {
		response.FailWithCode(errcode.Unauthorized.WithDetails("令牌已失效"), c)
		c.Abort()
		return
	}

	parsedClaims, err := svcCtx.JWT.ParseToken(tokenString)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			response.FailWithCode(errcode.Unauthorized.WithDetails("授权已过期"), c)
			c.Abort()
			return
		}
		response.FailWithCode(errcode.Unauthorized.WithDetails("无效的令牌"), c)
		c.Abort()
		return
	}

	refreshTokenIfNeeded(svcCtx, c, tokenString, parsedClaims)
	setClaimsOnContext(c, parsedClaims)

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	if !allowPoetryRouteWithEnforcer(svcCtx.CasbinEnforcer, parsedClaims.AuthorityId, c.Request.Method, path) {
		response.FailWithError(errcode.AssessDenied, c)
		c.Abort()
		return
	}

	c.Next()
}

func refreshTokenIfNeeded(svcCtx *svc.ServiceContext, c *gin.Context, tokenString string, parsedClaims *claims.CustomClaims) {
	if parsedClaims.ExpiresAt.Unix()-time.Now().Unix() >= parsedClaims.BufferTime {
		return
	}

	dr, _ := utils.ParseDuration(svcCtx.Config.JWT.ExpiresTime)
	newToken, _, err := svcCtx.JWT.ResolveToken(c.Request.Context(), tokenString, parsedClaims)
	if err != nil || newToken == "" {
		return
	}

	c.Header("new-token", newToken)
	c.Header("x-token", newToken)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("x-token", newToken, int(dr.Seconds()), "/", "", false, true)
}

func setClaimsOnContext(c *gin.Context, parsedClaims *claims.CustomClaims) {
	c.Set(CtxKeyClaims, parsedClaims)
	c.Set(CtxKeyUserID, parsedClaims.UserID)
	c.Set(CtxKeyUserUUID, parsedClaims.UUID)
	c.Set(CtxKeyAuthorityId, parsedClaims.AuthorityId)
}

func allowPoetryRouteWithEnforcer(e *casbin.SyncedCachedEnforcer, authorityID uint, method, path string) bool {
	if authorityID == 1 {
		return true
	}
	if e == nil {
		return false
	}
	ok, _ := e.Enforce(strconv.Itoa(int(authorityID)), path, method)
	return ok
}
