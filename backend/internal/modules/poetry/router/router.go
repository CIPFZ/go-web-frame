package poetry

import (
	"github.com/CIPFZ/gowebframe/internal/middleware"
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

func (r *PoetryRouter) InitPoetryRoutes(privateGroup *gin.RouterGroup, publicGroup *gin.RouterGroup, apiTokenGroup *gin.RouterGroup) {
	_ = apiTokenGroup

	writeGroup := privateGroup.Group("poetry")
	r.initDynastyWriteRoutes(writeGroup)
	r.initGenreWriteRoutes(writeGroup)
	r.initAuthorWriteRoutes(writeGroup)
	r.initPoemWriteRoutes(writeGroup)

	readGroup := publicGroup.Group("poetry")
	readGroup.Use(middleware.PoetryReadAuth(r.svcCtx))
	r.initReadOnlyRoutes(readGroup)
}

func (r *PoetryRouter) initDynastyWriteRoutes(group *gin.RouterGroup) {
	dGroup := group.Group("dynasty")
	{
		dGroup.POST("", r.apis.CreateDynasty)
		dGroup.PUT(":id", r.apis.UpdateDynasty)
		dGroup.DELETE(":id", r.apis.DeleteDynasty)
	}
}

func (r *PoetryRouter) initDynastyReadRoutes(group *gin.RouterGroup) {
	dGroup := group.Group("dynasty")
	{
		dGroup.GET("list", r.apis.ListDynasty)
		dGroup.GET("all", r.apis.AllDynasties)
	}
}

func (r *PoetryRouter) initGenreWriteRoutes(group *gin.RouterGroup) {
	gGroup := group.Group("genre")
	{
		gGroup.POST("", r.apis.CreateGenre)
		gGroup.PUT(":id", r.apis.UpdateGenre)
		gGroup.DELETE(":id", r.apis.DeleteGenre)
	}
}

func (r *PoetryRouter) initGenreReadRoutes(group *gin.RouterGroup) {
	gGroup := group.Group("genre")
	{
		gGroup.GET("list", r.apis.ListGenre)
		gGroup.GET("all", r.apis.AllGenres)
	}
}

func (r *PoetryRouter) initAuthorWriteRoutes(group *gin.RouterGroup) {
	aGroup := group.Group("author")
	{
		aGroup.POST("", r.apis.CreateAuthor)
		aGroup.PUT(":id", r.apis.UpdateAuthor)
		aGroup.DELETE(":id", r.apis.DeleteAuthor)
		aGroup.POST("avatar", r.apis.UploadAuthorAvatar)
	}
}

func (r *PoetryRouter) initAuthorReadRoutes(group *gin.RouterGroup) {
	aGroup := group.Group("author")
	{
		aGroup.GET("list", r.apis.ListAuthor)
		aGroup.GET(":id", r.apis.DetailAuthor)
	}
}

func (r *PoetryRouter) initPoemWriteRoutes(group *gin.RouterGroup) {
	pGroup := group.Group("poem")
	{
		pGroup.POST("", r.apis.CreatePoem)
		pGroup.PUT(":id", r.apis.UpdatePoem)
		pGroup.DELETE(":id", r.apis.DeletePoem)
	}
}

func (r *PoetryRouter) initPoemReadRoutes(group *gin.RouterGroup) {
	pGroup := group.Group("poem")
	{
		pGroup.GET("list", r.apis.ListPoem)
		pGroup.GET(":id", r.apis.DetailPoem)
	}
}

func (r *PoetryRouter) initReadOnlyRoutes(group *gin.RouterGroup) {
	r.initDynastyReadRoutes(group)
	r.initGenreReadRoutes(group)
	r.initAuthorReadRoutes(group)
	r.initPoemReadRoutes(group)
}
