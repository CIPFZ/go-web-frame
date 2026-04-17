package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	tokenCore "github.com/CIPFZ/gowebframe/internal/core/token"
	"github.com/CIPFZ/gowebframe/internal/modules/system/dto"
	"github.com/CIPFZ/gowebframe/internal/modules/system/model"
	"github.com/CIPFZ/gowebframe/internal/modules/system/repository"
	"github.com/CIPFZ/gowebframe/internal/svc"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestApiTokenServiceCreatePersistsHashAndApis(t *testing.T) {
	gormDB := newApiTokenTestDB(t)
	if err := gormDB.Create(&model.SysApi{
		Path:        "/api/v1/poetry/dynasty/list",
		Method:      "GET",
		ApiGroup:    "poetry",
		Description: "List dynasty",
	}).Error; err != nil {
		t.Fatalf("seed sys api error = %v", err)
	}

	var api model.SysApi
	if err := gormDB.First(&api).Error; err != nil {
		t.Fatalf("load seeded api error = %v", err)
	}

	service := NewApiTokenService(
		&svc.ServiceContext{DB: gormDB, Logger: zap.NewNop()},
		repository.NewApiTokenRepository(gormDB),
	)

	resp, err := service.CreateApiToken(context.Background(), 99, dto.CreateApiTokenReq{
		Name:           "poetry-reader",
		Description:    "read poetry api",
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
		Path:        "/api/v1/poetry/poem/list",
		Method:      "GET",
		ApiGroup:    "poetry",
		Description: "List poems",
	}).Error; err != nil {
		t.Fatalf("seed sys api error = %v", err)
	}

	var api model.SysApi
	if err := gormDB.First(&api).Error; err != nil {
		t.Fatalf("load seeded api error = %v", err)
	}

	service := NewApiTokenService(
		&svc.ServiceContext{DB: gormDB, Logger: zap.NewNop()},
		repository.NewApiTokenRepository(gormDB),
	)

	created, err := service.CreateApiToken(context.Background(), 7, dto.CreateApiTokenReq{
		Name:           "reset-me",
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

	resetResp, err := service.ResetApiToken(context.Background(), before.ID)
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
		&model.SysApi{},
		&model.SysApiToken{},
		&model.SysApiTokenApi{},
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

	return gormDB
}

func TestApiTokenServiceCreateRejectsExpiredTimeInPast(t *testing.T) {
	gormDB := newApiTokenTestDB(t)
	service := NewApiTokenService(
		&svc.ServiceContext{DB: gormDB, Logger: zap.NewNop()},
		repository.NewApiTokenRepository(gormDB),
	)

	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	_, err := service.CreateApiToken(context.Background(), 1, dto.CreateApiTokenReq{
		Name:        "expired",
		ExpiresAt:   &past,
		ApiIds:      []uint{},
		NeverExpire: false,
	})
	if err == nil {
		t.Fatal("CreateApiToken() error = nil, want validation error")
	}
}
