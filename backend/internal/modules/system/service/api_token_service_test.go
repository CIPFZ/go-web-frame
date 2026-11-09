package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/claims"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/modules/common"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
)

func allowAll(context.Context, uint, string, string) (bool, error) { return true, nil }

// tokenAdmin 是不受申请限制约束的管理员。
func tokenAdmin(id uint) TokenActor { return TokenActor{UserID: id, Manage: true} }

func TestApiTokenServiceCreatePersistsHashAndApis(t *testing.T) {
	gormDB := newApiTokenTestDB(t)
	if err := gormDB.Create(&model.SysApi{
		Path:        "/api/v1/open/token-info",
		Method:      "GET",
		ApiGroup:    "test",
		Description: "List resources",
	}).Error; err != nil {
		t.Fatalf("seed sys api error = %v", err)
	}

	var api model.SysApi
	if err := gormDB.First(&api).Error; err != nil {
		t.Fatalf("load seeded api error = %v", err)
	}

	service := NewApiTokenService(repository.NewApiTokenRepository(gormDB), allowAll)

	expires := time.Now().Add(time.Hour).Format(time.RFC3339)
	resp, err := service.CreateApiToken(context.Background(), tokenAdmin(99), dto.CreateApiTokenReq{
		Name:           "test-reader",
		ExpiresAt:      &expires,
		Description:    "read test api",
		MaxConcurrency: 2,
		ApiIds:         []uint{api.ID},
	})
	if err != nil {
		t.Fatalf("CreateApiToken() error = %v", err)
	}
	if resp.Token == "" {
		t.Fatal("CreateApiToken() returned empty raw token")
	}
	if len(resp.Apis) != 1 || resp.Apis[0].ID != api.ID {
		t.Fatalf("CreateApiToken() apis = %#v, want api id %d", resp.Apis, api.ID)
	}

	var stored model.SysApiToken
	if err := gormDB.Preload("Apis").First(&stored).Error; err != nil {
		t.Fatalf("load stored api token error = %v", err)
	}
	if stored.TokenHash == resp.Token {
		t.Fatal("stored token hash should not equal raw token")
	}
	if !tokenCore.VerifyToken(resp.Token, stored.TokenHash) {
		t.Fatal("stored token hash does not match returned raw token")
	}
	if stored.CreatedBy != 99 {
		t.Fatalf("stored created_by = %d, want 99", stored.CreatedBy)
	}
	if len(stored.Apis) != 1 || stored.Apis[0].ID != api.ID {
		t.Fatalf("stored apis = %#v, want api id %d", stored.Apis, api.ID)
	}
}

func TestApiTokenServiceResetReplacesStoredHash(t *testing.T) {
	gormDB := newApiTokenTestDB(t)
	if err := gormDB.Create(&model.SysApi{
		Path:        "/api/v1/open/token-info",
		Method:      "GET",
		ApiGroup:    "test",
		Description: "List records",
	}).Error; err != nil {
		t.Fatalf("seed sys api error = %v", err)
	}

	var api model.SysApi
	if err := gormDB.First(&api).Error; err != nil {
		t.Fatalf("load seeded api error = %v", err)
	}

	service := NewApiTokenService(repository.NewApiTokenRepository(gormDB), allowAll)

	expires := time.Now().Add(time.Hour).Format(time.RFC3339)
	created, err := service.CreateApiToken(context.Background(), tokenAdmin(7), dto.CreateApiTokenReq{
		Name:           "reset-me",
		ExpiresAt:      &expires,
		MaxConcurrency: 1,
		ApiIds:         []uint{api.ID},
	})
	if err != nil {
		t.Fatalf("CreateApiToken() error = %v", err)
	}

	var before model.SysApiToken
	if err := gormDB.First(&before).Error; err != nil {
		t.Fatalf("load before reset error = %v", err)
	}

	resetResp, err := service.ResetApiToken(context.Background(), tokenAdmin(7), before.ID)
	if err != nil {
		t.Fatalf("ResetApiToken() error = %v", err)
	}
	if resetResp.Token == "" {
		t.Fatal("ResetApiToken() returned empty raw token")
	}
	if resetResp.Token == created.Token {
		t.Fatal("ResetApiToken() returned old token, want new token")
	}

	var after model.SysApiToken
	if err := gormDB.First(&after, before.ID).Error; err != nil {
		t.Fatalf("load after reset error = %v", err)
	}
	if before.TokenHash == after.TokenHash {
		t.Fatal("ResetApiToken() did not replace token hash")
	}
	if tokenCore.VerifyToken(created.Token, after.TokenHash) {
		t.Fatal("old token should not match new stored hash")
	}
	if !tokenCore.VerifyToken(resetResp.Token, after.TokenHash) {
		t.Fatal("new token does not match stored hash")
	}
}

func newApiTokenTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gormDB, err := db.InitDatabase(config.Database{
		Driver: "sqlite3",
		SQLite: config.SQLite{
			Path:         filepath.Join(t.TempDir(), "apitoken.db"),
			MaxIdleConns: 1,
			MaxOpenConns: 1,
		},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("InitDatabase() error = %v", err)
	}

	if err := gormDB.AutoMigrate(
		&model.SysApi{}, &claims.PolicyRevision{},
		&model.SysApiToken{},
		&model.SysApiTokenApi{},
		&model.SysTokenLimit{},
		&model.SysUser{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := gormDB.Create(&claims.PolicyRevision{ID: 1, Version: 1}).Error; err != nil {
		t.Fatal(err)
	}
	return gormDB
}

func TestApiTokenServiceCreateRejectsExpiredTimeInPast(t *testing.T) {
	gormDB := newApiTokenTestDB(t)
	service := NewApiTokenService(repository.NewApiTokenRepository(gormDB), allowAll)

	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	_, err := service.CreateApiToken(context.Background(), tokenAdmin(1), dto.CreateApiTokenReq{
		Name:        "expired",
		ExpiresAt:   &past,
		ApiIds:      []uint{},
		NeverExpire: false,
	})
	if err == nil {
		t.Fatal("CreateApiToken() error = nil, want validation error")
	}
}

func TestApiTokenRejectsCMSPermission(t *testing.T) {
	database := newApiTokenTestDB(t)
	api := model.SysApi{Path: "/api/v1/sys/user/getUserList", Method: "POST"}
	database.Create(&api)
	expires := time.Now().Add(time.Hour).Format(time.RFC3339)
	service := NewApiTokenService(repository.NewApiTokenRepository(database), allowAll)
	_, err := service.CreateApiToken(context.Background(), tokenAdmin(1), dto.CreateApiTokenReq{Name: "bad grant", ExpiresAt: &expires, MaxConcurrency: 1, ApiIds: []uint{api.ID}})
	if err == nil {
		t.Fatal("JWT-only CMS API accepted as token permission")
	}
}

func TestApiTokenToggleMissingFails(t *testing.T) {
	service := NewApiTokenService(repository.NewApiTokenRepository(newApiTokenTestDB(t)), allowAll)
	if err := service.DisableApiToken(context.Background(), tokenAdmin(1), 999); err == nil {
		t.Fatal("nonexistent token reported success")
	}
}

// Deterministically interleave permission removal between reset's read and write.
type revokeOnReadRepo struct {
	repository.IApiTokenRepository
	database *gorm.DB
}

func (r revokeOnReadRepo) FindByID(ctx context.Context, id uint) (*model.SysApiToken, error) {
	token, err := r.IApiTokenRepository.FindByID(ctx, id)
	if err == nil {
		err = r.database.Where("api_token_id = ?", id).Delete(&model.SysApiTokenApi{}).Error
	}
	return token, err
}
func TestResetDoesNotRestoreConcurrentlyRevokedPermissions(t *testing.T) {
	database := newApiTokenTestDB(t)
	api := model.SysApi{Path: "/api/v1/open/token-info", Method: "GET"}
	database.Create(&api)
	expires := time.Now().Add(time.Hour).Format(time.RFC3339)
	repo := repository.NewApiTokenRepository(database)
	service := NewApiTokenService(repo, allowAll)
	created, err := service.CreateApiToken(context.Background(), tokenAdmin(1), dto.CreateApiTokenReq{Name: "rotate", ExpiresAt: &expires, MaxConcurrency: 1, ApiIds: []uint{api.ID}})
	if err != nil {
		t.Fatal(err)
	}
	service = NewApiTokenService(revokeOnReadRepo{repo, database}, allowAll)
	if _, err := service.ResetApiToken(context.Background(), tokenAdmin(1), created.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.FindByID(context.Background(), created.ID)
	if err != nil || len(stored.Apis) != 0 {
		t.Fatalf("revoked permissions restored: %v %v", stored, err)
	}
}

// 个人 Token：只能选自己有的权限，受申请限制约束；看不到、改不了别人的。
func TestPersonalTokens(t *testing.T) {
	database := newApiTokenTestDB(t)
	api := model.SysApi{Path: "/api/v1/open/token-info", Method: "GET", ApiGroup: "test", Description: "x"}
	if err := database.Create(&api).Error; err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	granted := false
	svc := NewApiTokenService(repository.NewApiTokenRepository(database), func(context.Context, uint, string, string) (bool, error) { return granted, nil })
	alice, bob := TokenActor{UserID: 10}, TokenActor{UserID: 11}

	// 只填名字：跟随主人的权限，有效期与额度取默认
	first, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "a", FollowOwner: true})
	if err != nil {
		t.Fatal(err)
	}
	if !first.FollowOwner || first.MaxConcurrency != 5 || first.ExpiresAt == nil {
		t.Fatalf("defaults not applied: %+v", first.ApiTokenResponse)
	}
	// 自己没有的权限不能选；有了就能选
	if _, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "b", ApiIds: []uint{api.ID}}); err == nil || err.Error() != "token.permissionDenied" {
		t.Fatalf("err = %v", err)
	}
	granted = true
	if _, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "b", ApiIds: []uint{api.ID}}); err != nil {
		t.Fatal(err)
	}
	// 超过并发上限、有效期上限被拒绝；管理员不受约束
	if _, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "c", FollowOwner: true, MaxConcurrency: 6}); err == nil || err.Error() != "token.quotaTooHigh" {
		t.Fatalf("err = %v", err)
	}
	far := time.Now().AddDate(0, 0, 181).Format(time.RFC3339)
	if _, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "c", FollowOwner: true, ExpiresAt: &far}); err == nil || err.Error() != "token.expiryTooLong" {
		t.Fatalf("err = %v", err)
	}
	if _, err := svc.CreateApiToken(ctx, tokenAdmin(1), dto.CreateApiTokenReq{Name: "boss", FollowOwner: true, MaxConcurrency: 99, ExpiresAt: &far}); err != nil {
		t.Fatal(err)
	}

	// 单个用户的设置优先于全局默认；没设置的项跟随默认，默认改了也跟着变
	two, ten := 2, 10
	if err := svc.UpdateLimit(ctx, dto.TokenLimitSettings{UserID: alice.UserID, MaxTokens: &two}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateApiToken(ctx, alice, dto.CreateApiTokenReq{Name: "c", FollowOwner: true}); err == nil || err.Error() != "token.tooMany" {
		t.Fatalf("err = %v", err)
	}
	if err := svc.UpdateLimit(ctx, dto.TokenLimitSettings{UserID: 0, MaxTokens: &ten, MaxConcurrency: &two}); err != nil {
		t.Fatal(err)
	}
	limits, err := svc.EffectiveLimits(ctx, alice.UserID)
	if err != nil || limits.MaxTokens != 2 || limits.MaxConcurrency != 2 || limits.MaxTTLDays != 180 {
		t.Fatalf("limits = %+v, %v", limits, err)
	}
	if detail, err := svc.GetLimitDetail(ctx, bob.UserID); err != nil || detail.MaxTokens != nil || detail.Effective.MaxTokens != 10 {
		t.Fatalf("detail = %+v, %v", detail, err)
	}

	// 别人的 Token 与不存在不可区分
	if _, err := svc.GetApiTokenDetail(ctx, bob, first.ID); err == nil {
		t.Fatal("bob read alice's token")
	}
	if _, err := svc.ResetApiToken(ctx, bob, first.ID); err == nil {
		t.Fatal("bob reset alice's token")
	}
	if err := svc.DisableApiToken(ctx, bob, first.ID); err == nil {
		t.Fatal("bob disabled alice's token")
	}
	if err := svc.DeleteApiToken(ctx, bob, dto.DeleteApiTokenReq{ID: first.ID}); err == nil {
		t.Fatal("bob deleted alice's token")
	}
	if err := svc.DisableApiToken(ctx, alice, first.ID); err != nil {
		t.Fatal(err)
	}
	list, total, err := svc.GetApiTokenList(ctx, dto.SearchApiTokenReq{OwnerID: alice.UserID, PageInfo: common.PageInfo{Page: 1, PageSize: 10}})
	if err != nil || total != 2 || len(list) != 2 {
		t.Fatalf("list = %d %d %v", len(list), total, err)
	}
}
