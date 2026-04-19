package router

import (
	"net/http"
	"testing"

	pluginApi "github.com/CIPFZ/gowebframe/internal/modules/plugin/api"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/gin-gonic/gin"
)

func TestInitPluginRoutesRegistersCoreRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svcCtx := svc.NewServiceContext()
	router := NewPluginRouter(svcCtx, pluginApi.NewPluginApi(svcCtx, nil))
	engine := gin.New()
	publicGroup := engine.Group("/api/v1")
	privateGroup := engine.Group("/api/v1")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("InitPluginRoutes panicked: %v", r)
		}
	}()

	router.InitPluginRoutes(privateGroup, publicGroup)

	wantRoutes := map[string]string{
		http.MethodPost + " /api/v1/plugin/plugin/getPluginList":            "",
		http.MethodPost + " /api/v1/plugin/release/createRelease":           "",
		http.MethodPost + " /api/v1/plugin/release/reset":                   "",
		http.MethodPost + " /api/v1/plugin/work-order/getWorkOrderPool":     "",
		http.MethodPut + " /api/v1/plugin/product/updateProduct":            "",
		http.MethodPost + " /api/v1/plugin/department/getDepartmentList":    "",
		http.MethodPost + " /api/v1/plugin/public/getPublishedPluginList":   "",
		http.MethodPost + " /api/v1/plugin/public/getPublishedPluginDetail": "",
	}

	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := wantRoutes[key]; ok {
			delete(wantRoutes, key)
		}
	}
	if len(wantRoutes) != 0 {
		t.Fatalf("missing routes: %#v", wantRoutes)
	}
}
