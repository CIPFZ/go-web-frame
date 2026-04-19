package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/api"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/repository"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	gormDB, err := db.InitDatabase(config.Database{
		Driver: "sqlite3",
		SQLite: config.SQLite{
			Path:         filepath.Join(t.TempDir(), "market-router.db"),
			MaxIdleConns: 1,
			MaxOpenConns: 1,
		},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("InitDatabase() error = %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("gormDB.DB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	repo := repository.New(gormDB)
	svc := service.New(repo)
	if err := svc.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err := svc.UpsertPlugin(context.Background(), dto.UpsertPluginReq{
		PluginID:      1,
		Code:          "agent-helper",
		NameZh:        "智能助手插件",
		NameEn:        "Agent Helper",
		DescriptionZh: "用于自动巡检的插件",
		DescriptionEn: "Automation inspection plugin",
	}); err != nil {
		t.Fatalf("UpsertPlugin() error = %v", err)
	}
	if err := svc.UpsertVersion(context.Background(), dto.UpsertVersionReq{
		PluginID:      1,
		ReleaseID:     10,
		Version:       "1.0.0",
		NameZh:        "智能助手插件",
		NameEn:        "Agent Helper",
		DescriptionZh: "用于自动巡检的插件",
		DescriptionEn: "Automation inspection plugin",
		ReleasedAt:    "2026-04-19T10:00:00Z",
		PackageX86URL: "https://example.com/plugin-x86.zip",
	}); err != nil {
		t.Fatalf("UpsertVersion() error = %v", err)
	}

	engine := gin.New()
	Init(engine, api.New(svc, "sync-secret"))
	return engine
}

func performJSONRequest(t *testing.T, engine http.Handler, method, path string, payload any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	return resp
}

func TestListPluginsRouteReturnsPublishedPlugins(t *testing.T) {
	engine := newTestRouter(t)

	resp := performJSONRequest(t, engine, http.MethodPost, "/api/v1/market/plugins/list", map[string]any{
		"page":     1,
		"pageSize": 10,
	}, nil)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte(`"pluginId":1`)) {
		t.Fatalf("response body = %s, want pluginId 1", resp.Body.String())
	}
}

func TestGetPluginDetailRouteValidatesPayload(t *testing.T) {
	engine := newTestRouter(t)

	resp := performJSONRequest(t, engine, http.MethodPost, "/api/v1/market/plugins/detail", map[string]any{}, nil)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte(`"code":1000`)) {
		t.Fatalf("response body = %s, want error code", resp.Body.String())
	}
}

func TestSyncRoutesRequireToken(t *testing.T) {
	engine := newTestRouter(t)

	resp := performJSONRequest(t, engine, http.MethodPost, "/api/v1/market/sync/plugin/upsert", map[string]any{
		"pluginId":      2,
		"code":          "ops-center",
		"nameZh":        "运维中心插件",
		"nameEn":        "Ops Center",
		"descriptionZh": "用于运维监控",
		"descriptionEn": "Ops monitoring",
	}, nil)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte("unauthorized")) {
		t.Fatalf("response body = %s, want unauthorized", resp.Body.String())
	}
}

func TestSyncRoutesAcceptValidToken(t *testing.T) {
	engine := newTestRouter(t)

	resp := performJSONRequest(t, engine, http.MethodPost, "/api/v1/market/sync/plugin/upsert", map[string]any{
		"pluginId":      2,
		"code":          "ops-center",
		"nameZh":        "运维中心插件",
		"nameEn":        "Ops Center",
		"descriptionZh": "用于运维监控",
		"descriptionEn": "Ops monitoring",
	}, map[string]string{
		"X-Market-Sync-Token": "sync-secret",
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte(`"code":0`)) {
		t.Fatalf("response body = %s, want success code", resp.Body.String())
	}
}
