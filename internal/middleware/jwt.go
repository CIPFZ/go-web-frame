package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Token
		token := c.Request.Header.Get("x-token")
		if token == "" {
			response.FailWithDetailed(gin.H{"reload": true}, "未登录或非法访问", c)
			c.Abort()
			return
		}

		// 2. 检查黑名单 (如果实现了 logout)
		if svcCtx.JWT.IsBlacklist(c.Request.Context(), token) {
			response.FailWithDetailed(gin.H{"reload": true}, "您的帐户异地登陆或令牌失效", c)
			c.Abort()
			return
		}

		// 3. 解析 Token
		j := svcCtx.JWT
		claims, err := j.ParseToken(token)
		if err != nil {
			// 处理 Token 过期/无效
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.FailWithDetailed(gin.H{"reload": true}, "授权已过期", c)
				c.Abort()
				return
			}
			response.FailWithDetailed(gin.H{"reload": true}, err.Error(), c)
			c.Abort()
			return
		}

		// 4. Token 自动续期 (Buffer Time 逻辑)
		// 如果 claims.ExpiresAt - now < BufferTime，则生成新 Token
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(svcCtx.Config.JWT.ExpiresTime)
			// 使用 singleflight 避免并发刷新
			newToken, _, err := j.ResolveToken(c.Request.Context(), token, claims)
			if err == nil && newToken != "" {
				// 将新 Token 放入 Header
				c.Header("new-token", newToken)
				c.Header("x-token", newToken)
				// 同时更新 Cookie
				c.SetCookie(
					"x-token",         // name
					newToken,          // value
					int(dr.Seconds()), // maxAge (秒)
					"/",               // path
					"",                // domain (留空是最稳健的选择)
					false,             // secure (HTTPS only)
					true,              // httpOnly (关键安全设置: 禁止 JS 读取)
				)

				// 5. 设置 SameSite (防止 CSRF)
				c.SetSameSite(http.SameSiteLaxMode)
			}
		}

		// 5. 将 Claims 注入上下文
		c.Set("claims", claims)
		// 方便后续使用，单独设置 UUID 或 ID
		c.Set("userId", claims.UserID)
		c.Set("userUUID", claims.UUID)
		c.Set("authorityId", claims.AuthorityId)

		c.Next()
	}
}
