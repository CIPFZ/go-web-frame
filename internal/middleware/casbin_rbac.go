package middleware

import (
	"strconv"

	"github.com/CIPFZ/gowebframe/internal/core/claims"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/response"
	"github.com/gin-gonic/gin"
)

func CasbinHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取请求信息
		waitUse, _ := c.Get("claims")
		waitUseClaims := waitUse.(*claims.CustomClaims)

		// 获取请求的 PATH 和 METHOD
		obj := c.Request.URL.Path
		act := c.Request.Method
		// 获取用户的角色ID (Casbin 中通常存为字符串)
		sub := strconv.Itoa(int(waitUseClaims.AuthorityId))

		e := svcCtx.CasbinEnforcer

		// 2. 判断权限
		// 格式: Enforce(sub, obj, act) -> (角色ID, 路径, 方法)
		success, _ := e.Enforce(sub, obj, act)

		// 如果是超级管理员，直接放行
		if waitUseClaims.AuthorityId == 1 {
			success = true
		}

		if !success {
			response.FailWithDetailed(gin.H{}, "权限不足", c)
			c.Abort()
			return
		}

		c.Next()
	}
}
