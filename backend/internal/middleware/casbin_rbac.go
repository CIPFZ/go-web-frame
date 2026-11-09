package middleware

import (
	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

// CasbinHandler 按用户的全部角色校验当前路由（权限取并集，超级管理员直接放行）。
func CasbinHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := c.MustGet(CtxKeyClaims).(*claims.CustomClaims)
		switch AuthorizeRoute(c, svcCtx, cl) {
		case AuthOK:
			c.Next()
		case AuthInternal:
			response.FailWithCode(errcode.ServerError, c)
			c.Abort()
		default:
			response.FailWithError(errcode.AssessDenied, c)
			c.Abort()
		}
	}
}
