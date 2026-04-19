package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/repository"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	testProviderAuthorityID = uint(10010)
	testReviewerAuthorityID = uint(10013)
	testAdminAuthorityID    = uint(1)
)

func TestPluginServiceRejectReviseFlow(t *testing.T) {
	service, gormDB := newPluginTestService(t)
	ctx := context.Background()

	project, err := service.CreatePlugin(ctx, 11, testProviderAuthorityID, dto.CreatePluginReq{
		Code:          "disk-analyzer",
		NameZh:        "磁盘分析插件",
		NameEn:        "Disk Analyzer",
		RepositoryURL: "https://example.com/disk-analyzer.git",
		DescriptionZh: "分析磁盘使用情况",
		DescriptionEn: "Analyze disk usage",
		DepartmentID:  1,
		OwnerID:       11,
	})
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}

	release, err := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:        project.ID,
		Version:         "1.0.0",
		RequestType:     model.ReleaseRequestTypeVersion,
		TestReportURL:   "https://example.com/report.pdf",
		PackageX86URL:   "https://example.com/pkg-x86.zip",
		PackageARMURL:   "https://example.com/pkg-arm.zip",
		ChangelogZh:     "首次发布",
		ChangelogEn:     "Initial release",
		CompatibleItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
	})
	if err != nil {
		t.Fatalf("CreateRelease() error = %v", err)
	}

	if _, err := service.ClaimWorkOrder(ctx, 21, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID}); err == nil {
		t.Fatal("ClaimWorkOrder() before submit review error = nil, want failure")
	}

	if _, err := service.TransitionRelease(ctx, 11, testProviderAuthorityID, dto.TransitionReleaseReq{
		ID:     release.ID,
		Action: model.ReleaseActionSubmitReview,
	}); err != nil {
		t.Fatalf("submit review error = %v", err)
	}

	claimed, err := service.ClaimWorkOrder(ctx, 21, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID})
	if err != nil {
		t.Fatalf("ClaimWorkOrder() error = %v", err)
	}
	if claimed.ProcessStatus != model.ReleaseProcessStatusProcessing {
		t.Fatalf("claimed process status = %d, want %d", claimed.ProcessStatus, model.ReleaseProcessStatusProcessing)
	}

	if _, err := service.TransitionRelease(ctx, 21, testReviewerAuthorityID, dto.TransitionReleaseReq{
		ID:            release.ID,
		Action:        model.ReleaseActionReject,
		ReviewComment: "缺少更完整的兼容性说明",
	}); err != nil {
		t.Fatalf("reject error = %v", err)
	}

	revised, err := service.TransitionRelease(ctx, 11, testProviderAuthorityID, dto.TransitionReleaseReq{
		ID:     release.ID,
		Action: model.ReleaseActionRevise,
	})
	if err != nil {
		t.Fatalf("revise error = %v", err)
	}
	if revised.ProcessStatus != model.ReleaseProcessStatusPending {
		t.Fatalf("revised process status = %d, want %d", revised.ProcessStatus, model.ReleaseProcessStatusPending)
	}

	var events []model.PluginReleaseEvent
	if err := gormDB.Where("release_id = ?", release.ID).Order("id asc").Find(&events).Error; err != nil {
		t.Fatalf("load events error = %v", err)
	}
	if len(events) < 4 {
		t.Fatalf("events count = %d, want >= 4", len(events))
	}
}

func TestPluginServiceClaimIsExclusive(t *testing.T) {
	service, _ := newPluginTestService(t)
	ctx := context.Background()

	project, err := service.CreatePlugin(ctx, 11, testProviderAuthorityID, dto.CreatePluginReq{
		Code:          "netdiag",
		NameZh:        "网络诊断插件",
		NameEn:        "Netdiag",
		RepositoryURL: "https://example.com/netdiag.git",
		DescriptionZh: "网络诊断",
		DescriptionEn: "Network diagnostics",
		DepartmentID:  1,
		OwnerID:       11,
	})
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}
	release, err := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:      project.ID,
		Version:       "2.0.0",
		RequestType:   model.ReleaseRequestTypeVersion,
		TestReportURL: "https://example.com/report.pdf",
		PackageX86URL: "https://example.com/pkg-x86.zip",
		PackageARMURL: "https://example.com/pkg-arm.zip",
		ChangelogZh:   "稳定版本",
		ChangelogEn:   "Stable",
	})
	if err != nil {
		t.Fatalf("CreateRelease() error = %v", err)
	}
	if _, err := service.TransitionRelease(ctx, 11, testProviderAuthorityID, dto.TransitionReleaseReq{
		ID:     release.ID,
		Action: model.ReleaseActionSubmitReview,
	}); err != nil {
		t.Fatalf("submit review error = %v", err)
	}

	if _, err := service.ClaimWorkOrder(ctx, 21, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID}); err != nil {
		t.Fatalf("first claim error = %v", err)
	}
	if _, err := service.ClaimWorkOrder(ctx, 22, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID}); err == nil {
		t.Fatal("second claim error = nil, want conflict")
	}
}

func TestPluginServicePublicDetailReturnsVersionHistory(t *testing.T) {
	service, _ := newPluginTestService(t)
	ctx := context.Background()

	project, err := service.CreatePlugin(ctx, 11, testProviderAuthorityID, dto.CreatePluginReq{
		Code:          "log-center",
		NameZh:        "日志中心插件",
		NameEn:        "Log Center",
		RepositoryURL: "https://example.com/log-center.git",
		DescriptionZh: "统一日志检索",
		DescriptionEn: "Unified log search",
		DepartmentID:  1,
		OwnerID:       11,
	})
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}

	publish := func(version string) {
		release, releaseErr := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
			PluginID:      project.ID,
			Version:       version,
			RequestType:   model.ReleaseRequestTypeVersion,
			TestReportURL: "https://example.com/report.pdf",
			PackageX86URL: "https://example.com/pkg-x86.zip",
			PackageARMURL: "https://example.com/pkg-arm.zip",
			ChangelogZh:   "正式发布",
			ChangelogEn:   "GA",
		})
		if releaseErr != nil {
			t.Fatalf("CreateRelease(%s) error = %v", version, releaseErr)
		}
		if _, releaseErr = service.TransitionRelease(ctx, 11, testProviderAuthorityID, dto.TransitionReleaseReq{
			ID:     release.ID,
			Action: model.ReleaseActionSubmitReview,
		}); releaseErr != nil {
			t.Fatalf("submit review(%s) error = %v", version, releaseErr)
		}
		if _, releaseErr = service.ClaimWorkOrder(ctx, 21, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID}); releaseErr != nil {
			t.Fatalf("claim(%s) error = %v", version, releaseErr)
		}
		if _, releaseErr = service.TransitionRelease(ctx, 21, testReviewerAuthorityID, dto.TransitionReleaseReq{
			ID:            release.ID,
			Action:        model.ReleaseActionApprove,
			ReviewComment: "审核通过",
		}); releaseErr != nil {
			t.Fatalf("approve(%s) error = %v", version, releaseErr)
		}
		if _, releaseErr = service.TransitionRelease(ctx, 21, testReviewerAuthorityID, dto.TransitionReleaseReq{
			ID:     release.ID,
			Action: model.ReleaseActionRelease,
		}); releaseErr != nil {
			t.Fatalf("release(%s) error = %v", version, releaseErr)
		}
	}

	publish("3.0.0")
	publish("3.1.0")

	detail, err := service.GetPublishedPluginDetail(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetPublishedPluginDetail() error = %v", err)
	}
	if len(detail.Versions) != 2 {
		t.Fatalf("published versions = %d, want 2", len(detail.Versions))
	}
	if detail.Versions[0].Version != "3.1.0" {
		t.Fatalf("latest public version = %s, want 3.1.0", detail.Versions[0].Version)
	}
	if detail.Release.Version != "3.1.0" {
		t.Fatalf("current public release version = %s, want 3.1.0", detail.Release.Version)
	}
}

func newPluginTestService(t *testing.T) (IPluginService, *gorm.DB) {
	t.Helper()

	gormDB, err := db.InitDatabase(config.Database{
		Driver: "sqlite3",
		SQLite: config.SQLite{
			Path:         filepath.Join(t.TempDir(), "plugin.db"),
			MaxIdleConns: 1,
			MaxOpenConns: 1,
		},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("InitDatabase() error = %v", err)
	}

	if err := gormDB.AutoMigrate(
		&sysModel.SysUser{},
		&model.PluginDepartment{},
		&model.PluginProduct{},
		&model.Plugin{},
		&model.PluginRelease{},
		&model.PluginCompatibleProduct{},
		&model.PluginReleaseEvent{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	if err := seedPluginTestData(gormDB); err != nil {
		t.Fatalf("seedPluginTestData() error = %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	svcCtx := &svc.ServiceContext{DB: gormDB, Logger: zap.NewNop()}
	repo := repository.NewPluginRepository(gormDB)
	return NewPluginService(svcCtx, repo), gormDB
}

func seedPluginTestData(gormDB *gorm.DB) error {
	users := []sysModel.SysUser{
		{BaseModel: model.ToBaseModel(11), Username: "provider", NickName: "provider", Password: "hashed", AuthorityID: testProviderAuthorityID},
		{BaseModel: model.ToBaseModel(21), Username: "reviewer-a", NickName: "reviewer-a", Password: "hashed", AuthorityID: testReviewerAuthorityID},
		{BaseModel: model.ToBaseModel(22), Username: "reviewer-b", NickName: "reviewer-b", Password: "hashed", AuthorityID: testReviewerAuthorityID},
		{BaseModel: model.ToBaseModel(1), Username: "admin", NickName: "admin", Password: "hashed", AuthorityID: testAdminAuthorityID},
	}
	for _, item := range users {
		if err := gormDB.Create(&item).Error; err != nil {
			return err
		}
	}
	department := model.PluginDepartment{Name: "存储产品部", ProductLine: "存储", Status: true}
	if err := gormDB.Create(&department).Error; err != nil {
		return err
	}
	product := model.PluginProduct{Code: "HCI", Name: "超融合", Description: "HCI 平台", Status: true}
	return gormDB.Create(&product).Error
}
