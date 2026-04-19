package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CIPFZ/gowebframe/internal/docs"
	"github.com/CIPFZ/gowebframe/internal/middleware"
	pluginApi "github.com/CIPFZ/gowebframe/internal/modules/plugin/api"
	pluginRepo "github.com/CIPFZ/gowebframe/internal/modules/plugin/repository"
	pluginRouter "github.com/CIPFZ/gowebframe/internal/modules/plugin/router"
	pluginService "github.com/CIPFZ/gowebframe/internal/modules/plugin/service"
	poetryApi "github.com/CIPFZ/gowebframe/internal/modules/poetry/api"
	poetryRepo "github.com/CIPFZ/gowebframe/internal/modules/poetry/repository"
	poetryRouter "github.com/CIPFZ/gowebframe/internal/modules/poetry/router"
	poetryService "github.com/CIPFZ/gowebframe/internal/modules/poetry/service"
	systemApi "github.com/CIPFZ/gowebframe/internal/modules/system/api"
	systemRepo "github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	systemRouter "github.com/CIPFZ/gowebframe/internal/modules/system/router"
	systemService "github.com/CIPFZ/gowebframe/internal/modules/system/service"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

func InitRouters(svcCtx *svc.ServiceContext) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	registerGlobalMiddleware(r, svcCtx)
	registerBaseRoutes(r, svcCtx)

	routerPrefix := svcCtx.Config.System.RouterPrefix
	publicGroup := r.Group(routerPrefix)
	privateGroup := r.Group(routerPrefix)
	apiTokenGroup := r.Group(routerPrefix)
	privateGroup.Use(middleware.JWTAuth(svcCtx), middleware.CasbinHandler(svcCtx))
	apiTokenGroup.Use(middleware.ApiTokenAuth(svcCtx))

	sysRouter := wireSystemModule(svcCtx)
	sysRouter.InitSystemRoutes(privateGroup, publicGroup)
	wirePluginModule(svcCtx).InitPluginRoutes(privateGroup, publicGroup)
	wirePoetryModule(svcCtx).InitPoetryRoutes(privateGroup, publicGroup, apiTokenGroup)

	svcCtx.Routers = r.Routes()
	svcCtx.Logger.Info("all routes initialized")
	return r
}

func registerGlobalMiddleware(r *gin.Engine, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config

	if cfg.Observable.Exporter != "none" {
		r.Use(otelgin.Middleware(
			cfg.Observable.ServiceName,
			otelgin.WithTracerProvider(otel.GetTracerProvider()),
			otelgin.WithMeterProvider(otel.GetMeterProvider()),
			otelgin.WithFilter(func(r *http.Request) bool {
				path := r.URL.Path
				return path != "/health" && path != "/metrics" && !strings.HasPrefix(path, "/swagger")
			}),
		))
	}

	r.Use(middleware.RateLimitMiddleware(cfg.RateLimit))
	r.Use(middleware.BreakerMiddleware(svcCtx.Logger))
	r.Use(middleware.GinLoggerMiddleware(svcCtx.Logger))
	r.Use(middleware.CorsByRules(svcCtx.Config.Cors))
}

func registerBaseRoutes(r *gin.Engine, svcCtx *svc.ServiceContext) {
	routerPrefix := svcCtx.Config.System.RouterPrefix
	docs.SwaggerInfo.BasePath = routerPrefix

	r.GET(routerPrefix+"/swagger/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, routerPrefix+"/swagger/index.html")
	})
	r.GET(routerPrefix+"/swagger/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerIndexHTML(routerPrefix+"/swagger/doc.json"))
	})
	r.GET(routerPrefix+"/swagger/doc.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Unix()})
	})
}

func swaggerIndexHTML(docURL string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "%s",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`, docURL)
}

func wireSystemModule(svcCtx *svc.ServiceContext) *systemRouter.SystemRouter {
	userRepo := systemRepo.NewUserRepository(svcCtx.DB)
	menuRepo := systemRepo.NewMenuRepository(svcCtx.DB)
	authRepo := systemRepo.NewAuthorityRepository(svcCtx.DB)
	apiRepo := systemRepo.NewApiRepository(svcCtx.DB)
	apiTokenRepo := systemRepo.NewApiTokenRepository(svcCtx.DB)
	casbinRepo := systemRepo.NewCasbinRepository(svcCtx.CasbinEnforcer)
	opLogRepo := systemRepo.NewOperationLogRepository(svcCtx.DB)
	noticeRepo := systemRepo.NewNoticeRepository(svcCtx.DB)

	opLogService := systemService.NewOperationLogService(svcCtx, opLogRepo)
	userService := systemService.NewUserService(svcCtx, userRepo)
	menuService := systemService.NewMenuService(svcCtx, menuRepo)
	authService := systemService.NewAuthorityService(svcCtx, authRepo)
	apiService := systemService.NewApiService(svcCtx, apiRepo)
	apiTokenService := systemService.NewApiTokenService(svcCtx, apiTokenRepo)
	casbinService := systemService.NewCasbinService(svcCtx, casbinRepo)
	noticeService := systemService.NewNoticeService(svcCtx, noticeRepo)

	apis := &systemRouter.SystemApis{
		UserApi:      systemApi.NewUserApi(svcCtx, userService),
		MenuApi:      systemApi.NewMenuApi(svcCtx, menuService),
		AuthorityApi: systemApi.NewAuthorityApi(svcCtx, authService),
		SysApiApi:    systemApi.NewSysApiApi(svcCtx, apiService),
		ApiTokenApi:  systemApi.NewApiTokenApi(svcCtx, apiTokenService),
		CasbinApi:    systemApi.NewCasbinApi(svcCtx, casbinService),
		OpLogApi:     systemApi.NewOperationLogApi(svcCtx, opLogService),
		FileApi:      systemApi.NewFileApi(svcCtx),
		StateApi:     systemApi.NewStateApi(svcCtx),
		NoticeApi:    systemApi.NewNoticeApi(svcCtx, noticeService),
	}

	return systemRouter.NewSystemRouter(svcCtx, apis)
}

func wirePoetryModule(svcCtx *svc.ServiceContext) *poetryRouter.PoetryRouter {
	repo := poetryRepo.NewPoetryRepo(svcCtx.DB)
	service := poetryService.NewPoetryService(svcCtx, repo)
	apis := poetryApi.NewPoetryApi(svcCtx, service)
	return poetryRouter.NewPoetryRouter(svcCtx, apis)
}

func wirePluginModule(svcCtx *svc.ServiceContext) *pluginRouter.PluginRouter {
	repo := pluginRepo.NewPluginRepository(svcCtx.DB)
	service := pluginService.NewPluginService(svcCtx, repo)
	apis := pluginApi.NewPluginApi(svcCtx, service)
	return pluginRouter.NewPluginRouter(svcCtx, apis)
}
