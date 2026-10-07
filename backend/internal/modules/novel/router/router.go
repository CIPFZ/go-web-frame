package router

import (
	"github.com/CIPFZ/gowebframe/internal/modules/novel/api"
	"github.com/gin-gonic/gin"
)

type Router struct{ books *api.BookAPI }

func New(books *api.BookAPI) *Router { return &Router{books: books} }

func (r *Router) Init(private *gin.RouterGroup) {
	books := private.Group("novel/book")
	books.POST("list", r.books.List)
	books.GET("detail", r.books.Detail)
}
