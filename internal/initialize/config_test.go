package initialize

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CIPFZ/gowebframe/internal/svc"
	"github.com/fsnotify/fsnotify"
)

func TestInitConfig_Success(t *testing.T) {
	// 创建一个临时 YAML 文件
	//tmpFile := "/home/ytq/work/go-web-frame/etc/config.yaml"
	// 创建一个临时 YAML 文件
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("system:\n  name: GoWebFrame\n  environment: dev\n  port: 8080\n")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}

	ctx := &svc.ServiceContext{}

	v, err := InitConfig(tmpFile, ctx)
	if err != nil {
		t.Fatalf("InitConfig 应该成功，但报错: %v", err)
	}
	if v == nil {
		t.Fatal("viper 返回不应为 nil")
	}
}

func TestInitConfig_FileNotExist(t *testing.T) {
	ctx := &svc.ServiceContext{}
	_, err := InitConfig("not_exist.yaml", ctx)
	if err == nil {
		t.Fatal("应该报错: 文件不存在")
	}
}

func TestInitConfig_UnmarshalError(t *testing.T) {
	// 这里我们模拟 Unmarshal 失败：通过注入不匹配的类型
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("value")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}

	// 使用一个不能被 Unmarshal 的结构（比如 channel）
	ctx := &svc.ServiceContext{}

	_, err := InitConfig(tmpFile, ctx)
	t.Logf("err: %v", err)
	if err == nil {
		t.Fatal("应该报错: Unmarshal config failed")
	}
}

func TestInitConfig_OnConfigChange(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("system:\n  name: GoWebFrame\n")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}

	ctx := &svc.ServiceContext{}
	v, err := InitConfig(tmpFile, ctx)
	if err != nil {
		t.Fatalf("InitConfig 应该成功: %v", err)
	}

	// 模拟配置文件变化（触发 fsnotify 事件）
	v.OnConfigChange(func(e fsnotify.Event) {
		t.Logf("配置变化回调触发: %s", e.Name)
		_ = v.Unmarshal(&ctx.Config)
		t.Logf("name: %+v", ctx.Config)
	})

	// 修改文件
	if err := os.WriteFile(tmpFile, []byte("system:\n  name: changed\n"), 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 给 WatchConfig 一点时间处理
	time.Sleep(200 * time.Millisecond)
}
