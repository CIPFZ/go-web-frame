package middleware

import (
	"errors"
	"time"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/CIPFZ/gowebframe/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 定义上下文 Key，方便后续 Handler 获取
const (
	CtxKeyClaims      = "claims"
	CtxKeyUserID      = "userId"
	CtxKeyUserUUID    = "userUUID"
	CtxKeyAuthorityId = "authorityId"
)

func JWTAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Token (优先 Header，其次 Cookie)
		token := c.Request.Header.Get("x-token")
		if token == "" {
			token, _ = c.Cookie("x-token")
		}

		// 没有任何 Token，直接返回未授权
		if token == "" {
			response.FailWithCode(errcode.Unauthorized, c)
			c.Abort()
			return
		}

		// 2. 检查黑名单 (Redis)
		// 建议：黑名单检查通常包含“注销”或“强制下线”逻辑
		if svcCtx.JWT.IsBlacklist(c.Request.Context(), token) {
			response.FailWithCode(errcode.Unauthorized.WithDetails("令牌已失效"), c)
			c.Abort()
			return
		}

		// 3. 解析 Token
		j := svcCtx.JWT
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.FailWithCode(errcode.Unauthorized.WithDetails("授权已过期"), c)
				c.Abort()
				return
			}
			// 其他解析错误（如格式错误、签名无效）
			response.FailWithCode(errcode.Unauthorized.WithDetails("无效的令牌"), c)
			c.Abort()
			return
		}

		// 4. Token 自动续期 (Sliding Window / Buffer Time)
		// 逻辑：如果 (过期时间 - 当前时间) < 缓冲时间，则刷新
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(svcCtx.Config.JWT.ExpiresTime)

			// ResolveToken 内部应该处理并发问题 (SingleFlight)
			newToken, _, err := j.ResolveToken(c.Request.Context(), token, claims)

			if err == nil && newToken != "" {
				// A. 设置 Header
				c.Header("new-token", newToken)
				c.Header("x-token", newToken)

				// B. 更新 Cookie (属性应从 Config 读取更佳，这里保持硬编码或默认)
				c.SetCookie(
					"x-token",
					newToken,
					int(dr.Seconds()),
					"/",
					"",    // domain
					false, // secure: 生产环境建议为 true
					true,  // httpOnly: 禁止 JS 读取，防止 XSS 窃取
				)

				// C. 更新本次请求使用的是新 Token (可选，视业务需求)
				// token = newToken
			}
		}

		// 5. 注入上下文 (使用常量 Key)
		c.Set(CtxKeyClaims, claims)
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUserUUID, claims.UUID)
		c.Set(CtxKeyAuthorityId, claims.AuthorityId)

		c.Next()
	}
}
