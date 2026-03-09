package poetry

import (
	"github.com/CIPFZ/gowebframe/internal/modules/poetry/api"
	"github.com/CIPFZ/gowebframe/internal/svc"

	"github.com/gin-gonic/gin"
)

type PoetryRouter struct {
	svcCtx *svc.ServiceContext
	apis   *api.PoetryApi
}

func NewPoetryRouter(svcCtx *svc.ServiceContext, apis *api.PoetryApi) *PoetryRouter {
	return &PoetryRouter{
		svcCtx: svcCtx,
		apis:   apis,
	}
}

func (r *PoetryRouter) InitPoetryRoutes(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup) {
	group := privateGroup.Group("poetry")
	r.initDynastyRoutes(group)
	r.initGenreRoutes(group)
	r.initAuthorRoutes(group)
	r.initPoemRoutes(group)
}

// 初始化朝代路由
func (r *PoetryRouter) initDynastyRoutes(group *gin.RouterGroup) {
	dGroup := group.Group("dynasty")
	{
		dGroup.POST("", r.apis.CreateDynasty)
		dGroup.PUT(":id", r.apis.UpdateDynasty)
		dGroup.DELETE(":id", r.apis.DeleteDynasty)
		dGroup.GET("list", r.apis.ListDynasty)
		dGroup.GET("all", r.apis.AllDynasties)
	}
}

// 初始化体裁路由
func (r *PoetryRouter) initGenreRoutes(group *gin.RouterGroup) {
	gGroup := group.Group("genre")
	{
		gGroup.POST("", r.apis.CreateGenre)
		gGroup.PUT(":id", r.apis.UpdateGenre)
		gGroup.DELETE(":id", r.apis.DeleteGenre)
		gGroup.GET("list", r.apis.ListGenre)
		gGroup.GET("all", r.apis.AllGenres)
	}
}

// 初始化诗人路由
func (r *PoetryRouter) initAuthorRoutes(group *gin.RouterGroup) {
	aGroup := group.Group("author")
	{
		aGroup.POST("", r.apis.CreateAuthor)
		aGroup.PUT(":id", r.apis.UpdateAuthor)
		aGroup.DELETE(":id", r.apis.DeleteAuthor)
		aGroup.GET("list", r.apis.ListAuthor)
		aGroup.GET(":id", r.apis.DetailAuthor)
		aGroup.POST("avatar", r.apis.UploadAuthorAvatar)
	}
}

// 初始化诗词内容路由
func (r *PoetryRouter) initPoemRoutes(group *gin.RouterGroup) {
	pGroup := group.Group("poem")
	{
		pGroup.POST("", r.apis.CreatePoem)
		pGroup.PUT(":id", r.apis.UpdatePoem)
		pGroup.DELETE(":id", r.apis.DeletePoem)
		pGroup.GET("list", r.apis.ListPoem)
		pGroup.GET(":id", r.apis.DetailPoem)
	}
}
