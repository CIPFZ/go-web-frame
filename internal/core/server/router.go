package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/docs"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	"github.com/CIPFZ/gowebframe/internal/modules/system/api"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	systemRouter "github.com/CIPFZ/gowebframe/internal/modules/system/router"
	"github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

// InitRouters 是初始化 Gin 引擎和所有应用路由的主入口。
// 它负责设置服务器、注册全局中间件，并挂载所有业务模块的子路由。
func InitRouters(svcCtx *svc.ServiceContext) *gin.Engine {
	// 1. 初始化 Gin 引擎
	// 使用 gin.New() 而非 gin.Default()，以便完全控制中间件栈。
	// gin.Recovery() 中间件用于从任何 panic 中恢复并返回 500 错误。
	r := gin.New()
	r.Use(gin.Recovery())

	// 2. 注册全局中间件
	registerGlobalMiddleware(r, svcCtx)

	// 3. 注册基础路由 (Swagger, Health Check)
	// 这些路由是应用的基础，不属于任何特定的业务模块。
	registerBaseRoutes(r, svcCtx)

	// 4. 创建不同权限级别的 API 路由组
	routerPrefix := svcCtx.Config.System.RouterPrefix

	// 公共路由组：用于无需认证的路由 (如登录、注册)。
	publicGroup := r.Group(routerPrefix)

	// 私有路由组：用于需要认证和授权的路由。
	privateGroup := r.Group(routerPrefix)
	privateGroup.Use(middleware.JWTAuth(svcCtx), middleware.CasbinHandler(svcCtx))

	// 5. 组装依赖并注册模块路由
	// --- System 模块 ---
	sysRouter := wireSystemModule(svcCtx)
	sysRouter.InitSystemRoutes(privateGroup, publicGroup)

	// --- 在此注册其他业务模块 ---

	// 将所有已注册的路由信息存入 ServiceContext，便于调试或展示。
	svcCtx.Routers = r.Routes()
	svcCtx.Logger.Info("✅ 所有路由注册完成")

	return r
}

// registerGlobalMiddleware 注册需要应用于所有请求的全局中间件。
func registerGlobalMiddleware(r *gin.Engine, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config

	// OpenTelemetry 中间件，用于分布式追踪和指标收集。
	if cfg.Observable.Exporter != "none" {
		r.Use(otelgin.Middleware(
			cfg.Observable.ServiceName,
			otelgin.WithTracerProvider(otel.GetTracerProvider()),
			otelgin.WithMeterProvider(otel.GetMeterProvider()),
			// 过滤掉高频、低价值的路由，避免产生过多追踪数据。
			otelgin.WithFilter(func(r *http.Request) bool {
				path := r.URL.Path
				// 排除健康检查、监控指标和 Swagger 文档相关的路由。
				return path != "/health" && path != "/metrics" && !strings.HasPrefix(path, "/swagger")
			}),
		))
	}

	// 3. 注册治理中间件
	// 3.1 限流 (Rate Limit) - 将恶意流量挡在最外面
	r.Use(middleware.RateLimitMiddleware(cfg.RateLimit))

	// 3.2 熔断 (Circuit Breaker) - 保护内部服务
	r.Use(middleware.BreakerMiddleware(svcCtx.Logger))

	// 自定义的日志中间件，用于结构化日志记录。
	r.Use(middleware.GinLoggerMiddleware(svcCtx.Logger))

	// CORS (跨域资源共享) 中间件，根据配置规则处理跨域请求。
	r.Use(middleware.CorsByRules(svcCtx.Config.Cors))
}

// registerBaseRoutes 注册基础路由，如 Swagger 文档和健康检查端点。
func registerBaseRoutes(r *gin.Engine, svcCtx *svc.ServiceContext) {
	routerPrefix := svcCtx.Config.System.RouterPrefix

	// Swagger API 文档路由。
	docs.SwaggerInfo.BasePath = routerPrefix
	r.GET(routerPrefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查端点，用于服务存活探测。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Unix()})
	})
}

// wireSystemModule 负责处理 `system` 模块内部的所有依赖注入（DI）。
// 它创建并连接了 Repository、Service 和 API 各层，最终返回一个完全配置好的模块化路由器。
// 这种方式将模块的创建细节封装起来，保持了 InitRouters 的整洁。
func wireSystemModule(svcCtx *svc.ServiceContext) *systemRouter.SystemRouter {
	// --- 1. 实例化 Repositories (数据访问层) ---
	userRepo := repository.NewUserRepository(svcCtx.DB)
	menuRepo := repository.NewMenuRepository(svcCtx.DB)
	authRepo := repository.NewAuthorityRepository(svcCtx.DB)
	apiRepo := repository.NewApiRepository(svcCtx.DB)
	casbinRepo := repository.NewCasbinRepository(svcCtx.CasbinEnforcer)
	opLogRepo := repository.NewOperationLogRepository(svcCtx.DB)

	// --- 2. 实例化 Services (业务逻辑层) ---
	opLogService := service.NewOperationLogService(svcCtx, opLogRepo)
	userService := service.NewUserService(svcCtx, userRepo)
	menuService := service.NewMenuService(svcCtx, menuRepo)
	authService := service.NewAuthorityService(svcCtx, authRepo)
	apiService := service.NewApiService(svcCtx, apiRepo)
	casbinService := service.NewCasbinService(svcCtx, casbinRepo)

	// --- 3. 实例化 APIs (接口表现层) ---
	// 使用 SystemApis 结构体来聚合所有 API 处理器，使传递更简洁。
	apis := &systemRouter.SystemApis{
		UserApi:      api.NewUserApi(svcCtx, userService),
		MenuApi:      api.NewMenuApi(svcCtx, menuService),
		AuthorityApi: api.NewAuthorityApi(svcCtx, authService),
		SysApiApi:    api.NewSysApiApi(svcCtx, apiService),
		CasbinApi:    api.NewCasbinApi(svcCtx, casbinService),
		OpLogApi:     api.NewOperationLogApi(svcCtx, opLogService),
	}

	// --- 4. 创建并返回模块的路由器 ---
	return systemRouter.NewSystemRouter(svcCtx, apis)
}
