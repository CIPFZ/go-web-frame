package initialize

import (
	"net/http"
	"time"

	"github.com/CIPFZ/gowebframe/internal/docs"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/router/system"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouters 初始化总路由
func InitRouters(serviceCtx *svc.ServiceContext) *gin.Engine {
	// ----------------------------------------------------------------
	// 1. 初始化 Gin 引擎 (无默认中间件)
	// ----------------------------------------------------------------
	r := gin.New()

	// ----------------------------------------------------------------
	// 2. 注册全局核心中间件 (顺序非常重要！)
	// ----------------------------------------------------------------

	// 2.1. Recovery: 防止 Panic 导致进程崩溃 (最外层)
	r.Use(gin.Recovery())

	// 2.2. Otel: 链路追踪 & 指标 (在 Logger 之前，以便 Logger 能获取 TraceID)
	cfg := serviceCtx.Config
	if cfg.Observable.Exporter != "none" {
		// serviceCtx.Logger.Info("Registering OpenTelemetry middleware...") // (可选日志)
		r.Use(otelgin.Middleware(
			cfg.Observable.ServiceName,
			otelgin.WithTracerProvider(otel.GetTracerProvider()),
			otelgin.WithMeterProvider(otel.GetMeterProvider()),
			// 过滤掉健康检查和 swagger，减少噪音
			otelgin.WithFilter(func(r *http.Request) bool {
				path := r.URL.Path
				return path != "/health" && len(path) >= 8 && path[:8] != "/swagger" // 简单的 Swagger 过滤
			}),
		))
	}

	// 2.3. Logger: 访问日志 (使用我们打磨的 Zap 中间件)
	// 它会从 Context 中提取 Otel TraceID 并注入到日志中
	r.Use(logger.GinLoggerMiddleware(serviceCtx.Logger))

	// 2.4. CORS: 跨域处理
	r.Use(middleware.CorsByRules(serviceCtx.Config.Cors))

	// ----------------------------------------------------------------
	// 3. 注册基础路由 (Swagger, Health)
	// ----------------------------------------------------------------
	routerPrefix := serviceCtx.Config.System.RouterPrefix

	// Swagger (通常仅在非生产环境开启，这里假设始终开启或由外部控制)
	docs.SwaggerInfo.BasePath = routerPrefix
	r.GET(routerPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查 (不需要鉴权，通常也不需要详细日志，但 Otel filter 已处理)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Unix()})
	})

	// ----------------------------------------------------------------
	// 4. 注册业务路由分组
	// ----------------------------------------------------------------
	// 创建 Public 和 Private 两个基础组
	publicGroup := r.Group(routerPrefix)
	// Private 组自动挂载鉴权中间件
	privateGroup := r.Group(routerPrefix)
	// 鉴权中间件挂载在这里
	privateGroup.Use(middleware.JWTAuth(serviceCtx), middleware.CasbinHandler(serviceCtx))

	// --- [系统模块] ---
	systemRouter := system.NewSystemGroup(serviceCtx)
	systemRouter.RegisterRoutes(privateGroup, publicGroup)

	// --- [业务模块] (例如 biz、user、api 等) ---
	// initBizRouter(privateGroup, publicGroup)

	// --- [插件模块] ---
	// InstallPlugin(privateGroup, publicGroup, r)

	// ----------------------------------------------------------------
	// 5. 打印路由信息
	// ----------------------------------------------------------------
	// 存储路由表
	serviceCtx.Routers = r.Routes()
	serviceCtx.Logger.Info("✅ 所有路由注册完成")

	return r
}
