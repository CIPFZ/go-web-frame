package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/model"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin/repository"
	sysModel "github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/CIPFZ/gowebframe/pkg/errcode"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	testProviderAuthorityID = uint(10010)
	testReviewerAuthorityID = uint(10013)
	testAdminAuthorityID    = uint(1)
)

func TestCreateReleaseValidatesVersionAndCompatibility(t *testing.T) {
	service, _ := newPluginTestService(t)
	ctx := context.Background()

	project := mustCreatePlugin(t, service, ctx, 11, dto.CreatePluginReq{
		Code:          "plugin-versioning",
		NameZh:        "版本校验插件",
		NameEn:        "Versioning Plugin",
		RepositoryURL: "https://example.com/versioning.git",
		DescriptionZh: "用于测试版本约束",
		DescriptionEn: "Used to validate version rules",
		DepartmentID:  1,
		OwnerID:       11,
	})

	_, err := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:    project.ID,
		RequestType: model.ReleaseRequestTypeVersion,
		Version:     "1.0.1",
		ChangelogZh: "错误首版",
		ChangelogEn: "Wrong first release",
		Compatibility: dto.ReleaseCompatibilityReq{
			ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
		},
	})
	assertErrCode(t, err, errcode.PluginVersionInvalid)

	_, err = service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:      project.ID,
		RequestType:   model.ReleaseRequestTypeVersion,
		Version:       "1.0.0",
		ChangelogZh:   "缺少兼容性",
		ChangelogEn:   "Missing compatibility",
		TestReportURL: "https://example.com/report.pdf",
	})
	assertErrCode(t, err, errcode.PluginCompatibilityRequired)

	first, err := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:      project.ID,
		RequestType:   model.ReleaseRequestTypeVersion,
		Version:       "1.0.0",
		TestReportURL: "https://example.com/report.pdf",
		PackageX86URL: "https://example.com/pkg-x86.zip",
		PackageARMURL: "https://example.com/pkg-arm.zip",
		ChangelogZh:   "首个版本",
		ChangelogEn:   "Initial release",
		Compatibility: dto.ReleaseCompatibilityReq{
			ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
			AcliItems:    []dto.UpsertCompatibleProductReq{{ProductID: 2, VersionConstraint: ">=2.1"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateRelease(first) error = %v", err)
	}
	if first.Version != "1.0.0" {
		t.Fatalf("first version = %s, want 1.0.0", first.Version)
	}
	if len(first.Compatibility.ProductItems) != 1 {
		t.Fatalf("product compatibility count = %d, want 1", len(first.Compatibility.ProductItems))
	}
	if len(first.Compatibility.AcliItems) != 1 {
		t.Fatalf("acli compatibility count = %d, want 1", len(first.Compatibility.AcliItems))
	}

	_, err = service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:    project.ID,
		RequestType: model.ReleaseRequestTypeVersion,
		Version:     "1.0.0",
		ChangelogZh: "重复版本",
		ChangelogEn: "Duplicate version",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	assertErrCode(t, err, errcode.PluginVersionDuplicate)

	_, err = service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:    project.ID,
		RequestType: model.ReleaseRequestTypeVersion,
		Version:     "0.9.0",
		ChangelogZh: "回退版本",
		ChangelogEn: "Non increasing version",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	assertErrCode(t, err, errcode.PluginVersionInvalid)

	second, err := service.CreateRelease(ctx, 11, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:      project.ID,
		RequestType:   model.ReleaseRequestTypeVersion,
		Version:       "1.0.1",
		TestReportURL: "https://example.com/report-v101.pdf",
		ChangelogZh:   "第二个版本",
		ChangelogEn:   "Second version",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	if err != nil {
		t.Fatalf("CreateRelease(second) error = %v", err)
	}
	if !second.Compatibility.Universal {
		t.Fatal("second release universal = false, want true")
	}
}

func TestUpdateReleaseOnlyProviderCanEditReadyOrRejected(t *testing.T) {
	service, _ := newPluginTestService(t)
	ctx := context.Background()

	project := mustCreatePlugin(t, service, ctx, 11, dto.CreatePluginReq{
		Code:          "plugin-editability",
		NameZh:        "可编辑性插件",
		NameEn:        "Editability Plugin",
		RepositoryURL: "https://example.com/editability.git",
		DescriptionZh: "用于测试编辑限制",
		DescriptionEn: "Used to validate edit restrictions",
		DepartmentID:  1,
		OwnerID:       11,
	})
	release := mustCreateRelease(t, service, ctx, 11, project.ID, "1.0.0", dto.ReleaseCompatibilityReq{
		ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
	})

	err := service.UpdateRelease(ctx, 1, testAdminAuthorityID, dto.UpdateReleaseReq{
		ID:      release.ID,
		Version: "1.0.1",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	assertErrCode(t, err, errcode.PluginForbidden)

	err = service.UpdateRelease(ctx, 11, testProviderAuthorityID, dto.UpdateReleaseReq{
		ID:            release.ID,
		Version:       "1.0.0",
		ChangelogZh:   "已编辑的首版",
		ChangelogEn:   "Edited initial release",
		TestReportURL: "https://example.com/report-ready.pdf",
		Compatibility: dto.ReleaseCompatibilityReq{
			ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.4"}},
			AcliItems:    []dto.UpsertCompatibleProductReq{{ProductID: 2, VersionConstraint: ">=2.1"}},
		},
	})
	if err != nil {
		t.Fatalf("UpdateRelease(ready) error = %v", err)
	}

	if _, err := service.TransitionRelease(ctx, 11, testProviderAuthorityID, dto.TransitionReleaseReq{
		ID:     release.ID,
		Action: model.ReleaseActionSubmitReview,
	}); err != nil {
		t.Fatalf("TransitionRelease(submit) error = %v", err)
	}

	err = service.UpdateRelease(ctx, 11, testProviderAuthorityID, dto.UpdateReleaseReq{
		ID:      release.ID,
		Version: "1.0.1",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	assertErrCode(t, err, errcode.PluginReleaseNotEditable)

	if _, err := service.ClaimWorkOrder(ctx, 21, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID}); err != nil {
		t.Fatalf("ClaimWorkOrder() error = %v", err)
	}
	if _, err := service.TransitionRelease(ctx, 21, testReviewerAuthorityID, dto.TransitionReleaseReq{
		ID:            release.ID,
		Action:        model.ReleaseActionReject,
		ReviewComment: "Need more compatibility detail",
	}); err != nil {
		t.Fatalf("TransitionRelease(reject) error = %v", err)
	}

	err = service.UpdateRelease(ctx, 11, testProviderAuthorityID, dto.UpdateReleaseReq{
		ID:            release.ID,
		Version:       "1.0.0",
		ChangelogZh:   "驳回后修订",
		ChangelogEn:   "Revised after rejection",
		TestReportURL: "https://example.com/report-rejected.pdf",
		Compatibility: dto.ReleaseCompatibilityReq{
			Universal: true,
		},
	})
	if err != nil {
		t.Fatalf("UpdateRelease(rejected) error = %v", err)
	}
}

func TestGetWorkOrderPoolSupportsAllMineAndFilters(t *testing.T) {
	service, _ := newPluginTestService(t)
	ctx := context.Background()

	processing := mustCreateSubmittedClaimedRelease(t, service, ctx, 11, 21, "plugin-pending", "待审核插件", "1.0.0")
	approved := mustCreateSubmittedClaimedRelease(t, service, ctx, 11, 21, "plugin-approved", "已批准插件", "1.0.0")
	if _, err := service.TransitionRelease(ctx, 21, testReviewerAuthorityID, dto.TransitionReleaseReq{
		ID:            approved.ID,
		Action:        model.ReleaseActionApprove,
		ReviewComment: "Looks good",
	}); err != nil {
		t.Fatalf("TransitionRelease(approve) error = %v", err)
	}

	rejected := mustCreateSubmittedClaimedRelease(t, service, ctx, 12, 22, "plugin-rejected", "已驳回插件", "1.0.0")
	if _, err := service.TransitionRelease(ctx, 22, testReviewerAuthorityID, dto.TransitionReleaseReq{
		ID:            rejected.ID,
		Action:        model.ReleaseActionReject,
		ReviewComment: "Need more detail",
	}); err != nil {
		t.Fatalf("TransitionRelease(reject) error = %v", err)
	}

	allItems, total, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:    dto.WorkOrderScopeAll,
		PageInfo: dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(all) error = %v", err)
	}
	if total != 3 {
		t.Fatalf("GetWorkOrderPool(all) total = %d, want 3", total)
	}
	if len(allItems) != 3 {
		t.Fatalf("GetWorkOrderPool(all) items = %d, want 3", len(allItems))
	}

	var foundClaimer bool
	for _, item := range allItems {
		if item.ID == processing.ID {
			foundClaimer = true
			if item.ClaimerName != "Reviewer A" {
				t.Fatalf("claimer name = %s, want Reviewer A", item.ClaimerName)
			}
			if item.ClaimerUsername != "reviewer-a" {
				t.Fatalf("claimer username = %s, want reviewer-a", item.ClaimerUsername)
			}
		}
	}
	if !foundClaimer {
		t.Fatalf("processing work order %d not found in all scope", processing.ID)
	}

	mineItems, mineTotal, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:    dto.WorkOrderScopeMine,
		PageInfo: dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(mine) error = %v", err)
	}
	if mineTotal != 2 {
		t.Fatalf("GetWorkOrderPool(mine) total = %d, want 2", mineTotal)
	}
	if len(mineItems) != 2 {
		t.Fatalf("GetWorkOrderPool(mine) items = %d, want 2", len(mineItems))
	}

	status := model.ReleaseStatusRejected
	rejectedItems, rejectedTotal, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:  dto.WorkOrderScopeAll,
		Status: &status,
		PageInfo: dto.PageInfo{
			Page:     1,
			PageSize: 20,
		},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(status filter) error = %v", err)
	}
	if rejectedTotal != 1 || len(rejectedItems) != 1 || rejectedItems[0].ID != rejected.ID {
		t.Fatalf("rejected filter returned total=%d items=%d firstID=%d, want total=1 with rejected release", rejectedTotal, len(rejectedItems), firstID(rejectedItems))
	}

	processStatus := model.ReleaseProcessStatusProcessing
	processingItems, processingTotal, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:         dto.WorkOrderScopeAll,
		ProcessStatus: &processStatus,
		PageInfo:      dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(process filter) error = %v", err)
	}
	if processingTotal != 2 || len(processingItems) != 2 {
		t.Fatalf("processing filter returned total=%d items=%d, want 2", processingTotal, len(processingItems))
	}

	keywordItems, keywordTotal, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:    dto.WorkOrderScopeAll,
		Keyword:  "plugin-approved",
		PageInfo: dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(keyword) error = %v", err)
	}
	if keywordTotal != 1 || len(keywordItems) != 1 || keywordItems[0].ID != approved.ID {
		t.Fatalf("keyword filter returned total=%d items=%d firstID=%d, want approved release", keywordTotal, len(keywordItems), firstID(keywordItems))
	}

	claimerID := uint(22)
	claimerItems, claimerTotal, err := service.GetWorkOrderPool(ctx, 21, testReviewerAuthorityID, dto.SearchWorkOrderReq{
		Scope:     dto.WorkOrderScopeAll,
		ClaimerID: &claimerID,
		PageInfo:  dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetWorkOrderPool(claimer) error = %v", err)
	}
	if claimerTotal != 1 || len(claimerItems) != 1 || claimerItems[0].ID != rejected.ID {
		t.Fatalf("claimer filter returned total=%d items=%d firstID=%d, want rejected release", claimerTotal, len(claimerItems), firstID(claimerItems))
	}
}

func TestMasterDataReturnsLocalizedDepartmentsAndTypedProducts(t *testing.T) {
	service, gormDB := newPluginTestService(t)
	ctx := context.Background()

	departments, total, err := service.GetDepartmentList(ctx, dto.SearchDepartmentReq{
		PageInfo: dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetDepartmentList() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("department total = %d, want 2", total)
	}
	if departments[0].NameZh == "" || departments[0].NameEn == "" {
		t.Fatalf("department localized names missing: %+v", departments[0])
	}

	if err := service.CreateProduct(ctx, testAdminAuthorityID, dto.CreateProductReq{
		Code:        "OBS",
		Name:        "Object Storage",
		Type:        model.CompatibleTargetTypeProduct,
		Description: "Object storage platform",
	}); err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	var created model.PluginProduct
	if err := gormDB.Where("code = ?", "OBS").First(&created).Error; err != nil {
		t.Fatalf("query created product error = %v", err)
	}
	if created.Type != model.CompatibleTargetTypeProduct {
		t.Fatalf("created product type = %s, want %s", created.Type, model.CompatibleTargetTypeProduct)
	}

	if err := service.UpdateProduct(ctx, testAdminAuthorityID, dto.UpdateProductReq{
		ID:          created.ID,
		Name:        "Object Storage Updated",
		Type:        model.CompatibleTargetTypeAcli,
		Description: "CLI support pack",
		Status:      true,
	}); err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	var updated model.PluginProduct
	if err := gormDB.First(&updated, created.ID).Error; err != nil {
		t.Fatalf("query updated product error = %v", err)
	}
	if updated.Type != model.CompatibleTargetTypeAcli {
		t.Fatalf("updated product type = %s, want %s", updated.Type, model.CompatibleTargetTypeAcli)
	}
}

func TestDepartmentMasterCRUDAndInactiveQuery(t *testing.T) {
	service, gormDB := newPluginTestService(t)
	ctx := context.Background()

	if err := service.CreateDepartment(ctx, testAdminAuthorityID, dto.CreateDepartmentReq{
		NameZh:      "云产品部",
		NameEn:      "Cloud Product Dept",
		ProductLine: "Cloud Platform",
	}); err != nil {
		t.Fatalf("CreateDepartment() error = %v", err)
	}

	var created model.PluginDepartment
	if err := gormDB.Where("name_zh = ?", "云产品部").First(&created).Error; err != nil {
		t.Fatalf("query created department error = %v", err)
	}
	if !created.Status {
		t.Fatalf("created department status = false, want true")
	}

	if err := service.UpdateDepartment(ctx, testAdminAuthorityID, dto.UpdateDepartmentReq{
		ID:          created.ID,
		NameZh:      "云产品一部",
		NameEn:      "Cloud Product Dept A",
		ProductLine: "Cloud Platform",
		Status:      false,
	}); err != nil {
		t.Fatalf("UpdateDepartment() error = %v", err)
	}

	var updated model.PluginDepartment
	if err := gormDB.First(&updated, created.ID).Error; err != nil {
		t.Fatalf("query updated department error = %v", err)
	}
	if updated.NameZh != "云产品一部" || updated.Status {
		t.Fatalf("updated department = %+v, want renamed and disabled", updated)
	}

	activeItems, _, err := service.GetDepartmentList(ctx, dto.SearchDepartmentReq{
		PageInfo: dto.PageInfo{Page: 1, PageSize: 20},
	})
	if err != nil {
		t.Fatalf("GetDepartmentList(active) error = %v", err)
	}
	for _, item := range activeItems {
		if item.ID == created.ID {
			t.Fatalf("inactive department unexpectedly returned in active list")
		}
	}

	allItems, _, err := service.GetDepartmentList(ctx, dto.SearchDepartmentReq{
		PageInfo:        dto.PageInfo{Page: 1, PageSize: 20},
		IncludeInactive: true,
	})
	if err != nil {
		t.Fatalf("GetDepartmentList(all) error = %v", err)
	}
	found := false
	for _, item := range allItems {
		if item.ID == created.ID {
			found = true
			if item.NameEn != "Cloud Product Dept A" || item.Status {
				t.Fatalf("unexpected department item = %+v", item)
			}
		}
	}
	if !found {
		t.Fatalf("inactive department not returned when includeInactive=true")
	}
}

func TestGetPublishedPluginDetailUsesPluginID(t *testing.T) {
	service, gormDB := newPluginTestService(t)
	ctx := context.Background()

	project := mustCreatePlugin(t, service, ctx, 11, dto.CreatePluginReq{
		Code:          "public-plugin-detail",
		NameZh:        "公开插件详情",
		NameEn:        "Public Plugin Detail",
		RepositoryURL: "https://example.com/public-plugin-detail.git",
		DescriptionZh: "用于测试公开插件详情",
		DescriptionEn: "Used to test public plugin detail",
		DepartmentID:  1,
		OwnerID:       11,
	})

	release := mustCreateRelease(t, service, ctx, 11, project.ID, "1.0.0", dto.ReleaseCompatibilityReq{
		ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
	})

	now := release.CreatedAt
	if err := gormDB.Model(&model.PluginRelease{}).
		Where("id = ?", release.ID).
		Updates(map[string]interface{}{
			"status":         model.ReleaseStatusReleased,
			"process_status": model.ReleaseProcessStatusDone,
			"released_at":    now,
		}).Error; err != nil {
		t.Fatalf("mark release as published error = %v", err)
	}

	detail, err := service.GetPublishedPluginDetail(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetPublishedPluginDetail() error = %v", err)
	}
	if detail.Plugin.ID != project.ID {
		t.Fatalf("detail plugin id = %d, want %d", detail.Plugin.ID, project.ID)
	}
	if len(detail.Versions) != 1 {
		t.Fatalf("detail versions = %d, want 1", len(detail.Versions))
	}
	if detail.Versions[0].ReleaseID != release.ID {
		t.Fatalf("detail release id = %d, want %d", detail.Versions[0].ReleaseID, release.ID)
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
		{BaseModel: model.ToBaseModel(1), Username: "admin", NickName: "Administrator", Password: "hashed", AuthorityID: testAdminAuthorityID},
		{BaseModel: model.ToBaseModel(11), Username: "provider-a", NickName: "Provider A", Password: "hashed", AuthorityID: testProviderAuthorityID},
		{BaseModel: model.ToBaseModel(12), Username: "provider-b", NickName: "Provider B", Password: "hashed", AuthorityID: testProviderAuthorityID},
		{BaseModel: model.ToBaseModel(21), Username: "reviewer-a", NickName: "Reviewer A", Password: "hashed", AuthorityID: testReviewerAuthorityID},
		{BaseModel: model.ToBaseModel(22), Username: "reviewer-b", NickName: "Reviewer B", Password: "hashed", AuthorityID: testReviewerAuthorityID},
	}
	for _, item := range users {
		if err := gormDB.Create(&item).Error; err != nil {
			return err
		}
	}

	departments := []model.PluginDepartment{
		{Name: "存储产品部", NameZh: "存储产品部", NameEn: "Storage Products", ProductLine: "Storage", Sort: 1, Status: true},
		{Name: "网络产品部", NameZh: "网络产品部", NameEn: "Network Products", ProductLine: "Network", Sort: 2, Status: true},
	}
	for _, item := range departments {
		if err := gormDB.Create(&item).Error; err != nil {
			return err
		}
	}

	products := []model.PluginProduct{
		{Code: "HCI", Name: "Hyper Converged", Type: model.CompatibleTargetTypeProduct, Description: "HCI platform", Sort: 1, Status: true},
		{Code: "ACLI", Name: "aCLI", Type: model.CompatibleTargetTypeAcli, Description: "aCLI runtime", Sort: 2, Status: true},
		{Code: "VDC", Name: "VDC", Type: model.CompatibleTargetTypeProduct, Description: "VDC platform", Sort: 3, Status: true},
	}
	for _, item := range products {
		if err := gormDB.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func mustCreatePlugin(t *testing.T, service IPluginService, ctx context.Context, userID uint, req dto.CreatePluginReq) *dto.PluginItem {
	t.Helper()
	item, err := service.CreatePlugin(ctx, userID, testProviderAuthorityID, req)
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}
	return item
}

func mustCreateRelease(t *testing.T, service IPluginService, ctx context.Context, userID, pluginID uint, version string, compatibility dto.ReleaseCompatibilityReq) *dto.PluginReleaseItem {
	t.Helper()
	item, err := service.CreateRelease(ctx, userID, testProviderAuthorityID, dto.CreateReleaseReq{
		PluginID:      pluginID,
		RequestType:   model.ReleaseRequestTypeVersion,
		Version:       version,
		TestReportURL: "https://example.com/report.pdf",
		PackageX86URL: "https://example.com/pkg-x86.zip",
		PackageARMURL: "https://example.com/pkg-arm.zip",
		ChangelogZh:   "测试发布",
		ChangelogEn:   "Test release",
		Compatibility: compatibility,
	})
	if err != nil {
		t.Fatalf("CreateRelease() error = %v", err)
	}
	return item
}

func mustCreateSubmittedClaimedRelease(t *testing.T, service IPluginService, ctx context.Context, providerID, reviewerID uint, code, nameZh, version string) *dto.PluginReleaseItem {
	t.Helper()
	project := mustCreatePlugin(t, service, ctx, providerID, dto.CreatePluginReq{
		Code:          code,
		NameZh:        nameZh,
		NameEn:        code,
		RepositoryURL: "https://example.com/" + code + ".git",
		DescriptionZh: nameZh + "描述",
		DescriptionEn: code + " description",
		DepartmentID:  1,
		OwnerID:       providerID,
	})
	release := mustCreateRelease(t, service, ctx, providerID, project.ID, version, dto.ReleaseCompatibilityReq{
		ProductItems: []dto.UpsertCompatibleProductReq{{ProductID: 1, VersionConstraint: ">=5.2"}},
	})
	if _, err := service.TransitionRelease(ctx, providerID, testProviderAuthorityID, dto.TransitionReleaseReq{
		ID:     release.ID,
		Action: model.ReleaseActionSubmitReview,
	}); err != nil {
		t.Fatalf("TransitionRelease(submit) error = %v", err)
	}
	claimed, err := service.ClaimWorkOrder(ctx, reviewerID, testReviewerAuthorityID, dto.ClaimWorkOrderReq{ID: release.ID})
	if err != nil {
		t.Fatalf("ClaimWorkOrder() error = %v", err)
	}
	return claimed
}

func assertErrCode(t *testing.T, err error, want *errcode.Error) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want code %d", want.Code)
	}
	var got *errcode.Error
	if !errors.As(err, &got) {
		t.Fatalf("error type = %T, want *errcode.Error", err)
	}
	if got.Code != want.Code {
		t.Fatalf("error code = %d, want %d", got.Code, want.Code)
	}
}

func firstID(items []dto.WorkOrderItem) uint {
	if len(items) == 0 {
		return 0
	}
	return items[0].ID
}
