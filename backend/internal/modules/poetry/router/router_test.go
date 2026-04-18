package poetry

import (
	"net/http"
	"testing"

	poetryApi "github.com/CIPFZ/gowebframe/internal/modules/poetry/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

func TestInitPoetryRoutesRegistersReadRoutesOnce(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svcCtx := svc.NewServiceContext()
	router := NewPoetryRouter(svcCtx, poetryApi.NewPoetryApi(svcCtx, nil))
	engine := gin.New()
	publicGroup := engine.Group("/api/v1")
	privateGroup := engine.Group("/api/v1")
	apiTokenGroup := engine.Group("/api/v1")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("InitPoetryRoutes panicked: %v", r)
		}
	}()

	router.InitPoetryRoutes(privateGroup, publicGroup, apiTokenGroup)

	var dynastyListGetCount int
	for _, route := range engine.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/v1/poetry/dynasty/list" {
			dynastyListGetCount++
		}
	}
	if dynastyListGetCount != 1 {
		t.Fatalf("GET /api/v1/poetry/dynasty/list route count = %d, want 1", dynastyListGetCount)
	}
}
