package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/repository"
	"go.uber.org/zap"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	gormDB, err := db.InitDatabase(config.Database{
		Driver: "sqlite3",
		SQLite: config.SQLite{
			Path:         filepath.Join(t.TempDir(), "market.db"),
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
	svc := New(repo)
	if err := svc.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return svc
}

func TestUpsertPluginAndVersionThenQuery(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if err := svc.UpsertPlugin(ctx, dto.UpsertPluginReq{
		PluginID:      100,
		Code:          "disk-analyzer",
		NameZh:        "磁盘分析插件",
		NameEn:        "Disk Analyzer",
		DescriptionZh: "用于磁盘诊断",
		DescriptionEn: "Disk diagnostics",
		CapabilityZh:  "磁盘健康诊断",
		CapabilityEn:  "Disk health",
		OwnerName:     "Plugin Team",
	}); err != nil {
		t.Fatalf("UpsertPlugin() error = %v", err)
	}

	if err := svc.UpsertVersion(ctx, dto.UpsertVersionReq{
		PluginID:          100,
		ReleaseID:         200,
		Version:           "1.0.0",
		NameZh:            "磁盘分析插件",
		NameEn:            "Disk Analyzer",
		DescriptionZh:     "用于磁盘诊断",
		DescriptionEn:     "Disk diagnostics",
		CapabilityZh:      "磁盘健康诊断",
		CapabilityEn:      "Disk health",
		OwnerName:         "Plugin Team",
		Publisher:         "CMS",
		ChangelogZh:       "首发版本",
		ChangelogEn:       "Initial release",
		TestReportURL:     "https://example.com/report.pdf",
		PackageX86URL:     "https://example.com/x86.zip",
		PackageARMURL:     "https://example.com/arm.zip",
		ReleasedAt:        "2026-04-19T10:00:00Z",
		VersionConstraint: ">=1.0.0",
		CompatibleItems: []dto.CompatibilityItem{
			{TargetType: "product", ProductCode: "HCI", ProductName: "超融合", VersionConstraint: ">=5.2"},
			{TargetType: "acli", ProductCode: "ACLI", ProductName: "aCLI", VersionConstraint: ">=2.1"},
		},
	}); err != nil {
		t.Fatalf("UpsertVersion() error = %v", err)
	}

	list, total, err := svc.ListPublishedPlugins(ctx, dto.ListPluginsReq{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("ListPublishedPlugins() error = %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("list size = %d total = %d, want 1/1", len(list), total)
	}
	if list[0].LatestVersion != "1.0.0" {
		t.Fatalf("latest version = %s, want 1.0.0", list[0].LatestVersion)
	}

	detail, err := svc.GetPublishedPluginDetail(ctx, 100)
	if err != nil {
		t.Fatalf("GetPublishedPluginDetail() error = %v", err)
	}
	if detail.Plugin.Code != "disk-analyzer" {
		t.Fatalf("plugin code = %s, want disk-analyzer", detail.Plugin.Code)
	}
	if len(detail.Versions) != 1 || detail.Versions[0].ReleaseID != 200 {
		t.Fatalf("versions mismatch: %+v", detail.Versions)
	}
}

func TestOfflineVersionRemovesPublishedDetail(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	mustSeedVersion(t, svc, ctx)

	if err := svc.OfflineVersion(ctx, 200); err != nil {
		t.Fatalf("OfflineVersion() error = %v", err)
	}

	list, total, err := svc.ListPublishedPlugins(ctx, dto.ListPluginsReq{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}})
	if err != nil {
		t.Fatalf("ListPublishedPlugins() error = %v", err)
	}
	if total != 0 || len(list) != 0 {
		t.Fatalf("list size = %d total = %d, want 0/0", len(list), total)
	}
}

func TestDeletePluginIsIdempotent(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	mustSeedVersion(t, svc, ctx)

	if err := svc.DeletePlugin(ctx, 100); err != nil {
		t.Fatalf("DeletePlugin() error = %v", err)
	}
	if err := svc.DeletePlugin(ctx, 100); err != nil {
		t.Fatalf("DeletePlugin(second) error = %v", err)
	}
}

func TestSeedDemoDataIfEmptySeedsPublishedPlugins(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if err := svc.SeedDemoDataIfEmpty(ctx); err != nil {
		t.Fatalf("SeedDemoDataIfEmpty() error = %v", err)
	}

	list, total, err := svc.ListPublishedPlugins(ctx, dto.ListPluginsReq{PageInfo: dto.PageInfo{Page: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("ListPublishedPlugins() error = %v", err)
	}
	if total < 4 || len(list) < 4 {
		t.Fatalf("seeded plugin size = %d total = %d, want at least 4", len(list), total)
	}

	detail, err := svc.GetPublishedPluginDetail(ctx, list[0].PluginID)
	if err != nil {
		t.Fatalf("GetPublishedPluginDetail() error = %v", err)
	}
	if len(detail.Versions) == 0 {
		t.Fatalf("detail versions should not be empty")
	}
}

func TestSeedDemoDataIfEmptyDoesNotDuplicateExistingData(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if err := svc.SeedDemoDataIfEmpty(ctx); err != nil {
		t.Fatalf("SeedDemoDataIfEmpty(first) error = %v", err)
	}
	firstList, firstTotal, err := svc.ListPublishedPlugins(ctx, dto.ListPluginsReq{PageInfo: dto.PageInfo{Page: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("ListPublishedPlugins(first) error = %v", err)
	}

	if err := svc.SeedDemoDataIfEmpty(ctx); err != nil {
		t.Fatalf("SeedDemoDataIfEmpty(second) error = %v", err)
	}
	secondList, secondTotal, err := svc.ListPublishedPlugins(ctx, dto.ListPluginsReq{PageInfo: dto.PageInfo{Page: 1, PageSize: 20}})
	if err != nil {
		t.Fatalf("ListPublishedPlugins(second) error = %v", err)
	}

	if firstTotal != secondTotal || len(firstList) != len(secondList) {
		t.Fatalf("seed should be idempotent, first = %d/%d second = %d/%d", len(firstList), firstTotal, len(secondList), secondTotal)
	}
}

func mustSeedVersion(t *testing.T, svc *Service, ctx context.Context) {
	t.Helper()
	if err := svc.UpsertPlugin(ctx, dto.UpsertPluginReq{
		PluginID:      100,
		Code:          "disk-analyzer",
		NameZh:        "磁盘分析插件",
		NameEn:        "Disk Analyzer",
		DescriptionZh: "用于磁盘诊断",
		DescriptionEn: "Disk diagnostics",
	}); err != nil {
		t.Fatalf("UpsertPlugin() error = %v", err)
	}
	if err := svc.UpsertVersion(ctx, dto.UpsertVersionReq{
		PluginID:      100,
		ReleaseID:     200,
		Version:       "1.0.0",
		NameZh:        "磁盘分析插件",
		NameEn:        "Disk Analyzer",
		DescriptionZh: "用于磁盘诊断",
		DescriptionEn: "Disk diagnostics",
		ReleasedAt:    "2026-04-19T10:00:00Z",
	}); err != nil {
		t.Fatalf("UpsertVersion() error = %v", err)
	}
}
