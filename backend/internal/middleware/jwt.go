package middleware

import (
	"errors"
	"net/http"
	"time"

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

func JWTAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("x-token")
		if token == "" {
			token, _ = c.Cookie("x-token")
		}

		if token == "" {
			response.FailWithCode(errcode.Unauthorized, c)
			c.Abort()
			return
		}

		if svcCtx.JWT.IsBlacklist(c.Request.Context(), token) {
			response.FailWithCode(errcode.Unauthorized.WithDetails("令牌已失效"), c)
			c.Abort()
			return
		}

		j := svcCtx.JWT
		claims, err := j.ParseToken(token)
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

		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(svcCtx.Config.JWT.ExpiresTime)
			newToken, _, err := j.ResolveToken(c.Request.Context(), token, claims)
			if err == nil && newToken != "" {
				c.Header("new-token", newToken)
				c.Header("x-token", newToken)
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(
					"x-token",
					newToken,
					int(dr.Seconds()),
					"/",
					"",
					false,
					true,
				)
			}
		}

		c.Set(CtxKeyClaims, claims)
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUserUUID, claims.UUID)
		c.Set(CtxKeyAuthorityId, claims.AuthorityId)

		c.Next()
	}
}
