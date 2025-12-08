package i18n

import (
	"github.com/CIPFZ/gowebframe/internal/core/config"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Package i18n -----------------------------
// @file        : i18n_test.go
// @author      : CIPFZ
// @time        : 2025/9/20 14:54
// @description :
// -------------------------------------------

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s failed: %v", path, err)
	}
}

func TestNewI18n_TranslateAndReload(t *testing.T) {
	dir := t.TempDir()

	en := `user.login.success: "Login successful"
user.welcome:
  other: "Welcome, {{.Name}}"
`
	zh := `user.login.success: "登录成功"
user.welcome:
  other: "欢迎, {{.Name}}"
`
	writeFile(t, dir, "en.yaml", en)
	writeFile(t, dir, "zh.yaml", zh)

	logger := zap.NewNop()
	svc, err := NewI18n(config.I18n{Path: dir}, logger)
	if err != nil {
		t.Fatalf("NewI18n failed: %v", err)
	}

	// 基本翻译
	if got := svc.Translate("en", "user.login.success"); got != "Login successful" {
		t.Fatalf("unexpected en translation: %q", got)
	}
	if got := svc.Translate("zh", "user.login.success"); got != "登录成功" {
		t.Fatalf("unexpected zh translation: %q", got)
	}

	// 带模板参数的翻译
	got := svc.Translate("en", "user.welcome", map[string]interface{}{"Name": "Tom"})
	if got != "Welcome, Tom" {
		t.Fatalf("unexpected template translation: %q", got)
	}

	// 缺失翻译返回 fallback 字符串
	miss := svc.Translate("en", "does.not.exist")
	if miss != "missing translation: does.not.exist" {
		t.Fatalf("unexpected missing translation: %q", miss)
	}

	// 修改 en.yaml 内容并验证 watch 能 reload
	newEn := `user.login.success: "Login succeeded"
user.welcome:
  other: "Welcome again, {{.Name}}"
`
	writeFile(t, dir, "en.yaml", newEn)

	// 等待 reload 生效（最多等待 3s）
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if svc.Translate("en", "user.login.success") == "Login succeeded" &&
			svc.Translate("en", "user.welcome", map[string]interface{}{"Name": "Tom"}) == "Welcome again, Tom" {
			// reload 生效
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("i18n reload did not take effect in time")
}
