package initialize

import (
	"net/http"

	"github.com/CIPFZ/gowebframe/internal/docs"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/router/system"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouters 初始化总路由
func InitRouters(serviceCtx *svc.ServiceContext) *gin.Engine {
	// --------------------
	// 1. 初始化 Gin 引擎
	// --------------------
	r := gin.New()
	r.Use(gin.Recovery())
	if gin.Mode() == gin.DebugMode {
		r.Use(gin.Logger())
	}

	// --------------------
	// 2. 全局中间件 & CORS
	// --------------------
	r.Use(middleware.CorsByRules(serviceCtx.Config.Cors))
	serviceCtx.Logger.Info("use middleware cors")

	// --------------------
	// 3. Swagger 文档
	// --------------------
	routerPrefix := serviceCtx.Config.System.RouterPrefix
	docs.SwaggerInfo.BasePath = routerPrefix
	r.GET(routerPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	serviceCtx.Logger.Info("register swagger handler")

	// --------------------
	// 4. 路由分组
	// --------------------
	publicGroup := r.Group(routerPrefix)
	privateGroup := r.Group(routerPrefix)
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(serviceCtx))

	// --------------------
	// 5. 基础健康检测
	// --------------------
	publicGroup.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --------------------
	// 6. 可观测指标
	// --------------------

	// --------------------
	// 7. 注册系统模块路由
	// --------------------
	systemRouter := system.NewSystemGroup(serviceCtx)
	systemRouter.RegisterRoutes(privateGroup, publicGroup)

	// --------------------
	// 8. 注册业务模块路由（例如 biz、user、api 等）
	// --------------------
	//initBizRouter(privateGroup, publicGroup)

	// --------------------
	// 9. 插件路由
	// --------------------
	//InstallPlugin(privateGroup, publicGroup, r)

	// --------------------
	// 10. 保存路由信息
	// --------------------
	serviceCtx.Routers = r.Routes()
	serviceCtx.Logger.Info("router register success")

	return r
}
