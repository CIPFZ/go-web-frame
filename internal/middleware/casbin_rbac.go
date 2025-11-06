package middleware

import (
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

// CasbinHandler 拦截器
func CasbinHandler(serviceCtx *svc.ServiceContext) gin.HandlerFunc {
	//prefix := serviceCtx.Config.System.RouterPrefix
	return func(c *gin.Context) {
		//waitUse, _ := utils.GetClaims(c)
		////获取请求的PATH
		//path := c.Request.URL.Path
		//obj := strings.TrimPrefix(path, prefix)
		//// 获取请求方法
		//act := c.Request.Method
		//// 获取用户的角色
		//sub := strconv.Itoa(int(waitUse.AuthorityId))
		//e := utils.GetCasbin(serviceCtx) // 判断策略中是否存在
		//success, _ := e.Enforce(sub, obj, act)
		//if !success {
		//	response.FailWithDetailed(gin.H{}, "权限不足", c)
		//	c.Abort()
		//	return
		//}
		c.Next()
	}
}
