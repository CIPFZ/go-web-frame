package archiver

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

// mockProcessor 用来替代真正的压缩/解压实现
type mockProcessor struct {
	compressCalled   bool
	decompressCalled bool
}

func (m *mockProcessor) Compress(ctx context.Context, sources []string, destination string, opts *Task) error {
	m.compressCalled = true
	if len(sources) == 0 || destination == "" {
		return errors.New("invalid args")
	}
	return nil
}

func (m *mockProcessor) Decompress(ctx context.Context, source string, destination string, opts *Task) error {
	m.decompressCalled = true
	if source == "" || destination == "" {
		return errors.New("invalid args")
	}
	return nil
}

// TestAutodetectByFilename 自动检测文件后缀
func TestAutodetectByFilename(t *testing.T) {
	cases := map[string]Format{
		"file.zip":     ZIP,
		"file.tar":     TAR,
		"file.tar.zst": TARZST,
		"file.tzst":    TARZST,
		"file.unknown": "",
	}
	for name, want := range cases {
		got := AutodetectByFilename(name)
		if got != want {
			t.Errorf("AutodetectByFilename(%s) = %v, want %v", name, got, want)
		}
	}
}

// TestTaskBuilder 链式构建器
func TestTaskBuilder(t *testing.T) {
	task := Compress("a.txt").
		To("a.zip").
		WithFormat(ZIP).
		WithCPU(4).
		WithPassword("123").
		WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))).
		WithContext(context.TODO())

	if !task.isCompress {
		t.Error("expected isCompress=true")
	}
	if task.destination != "a.zip" {
		t.Errorf("got %s, want a.zip", task.destination)
	}
	if task.format != ZIP {
		t.Errorf("got %s, want zip", task.format)
	}
	if task.cpuLimit != 4 {
		t.Errorf("got %d, want 4", task.cpuLimit)
	}
	if task.password != "123" {
		t.Errorf("got %s, want 123", task.password)
	}
	if task.ctx == nil {
		t.Error("ctx should not be nil")
	}
}

// TestExecuteErrors 执行路径错误
func TestExecuteErrors(t *testing.T) {
	// 缺少 source
	task := Compress()
	task.destination = "out.zip"
	task.format = ZIP
	if err := task.Execute(); !errors.Is(err, ErrSourceRequired) {
		t.Errorf("expected ErrSourceRequired, got %v", err)
	}

	// 缺少 destination
	task = Compress("file.txt")
	task.destination = ""
	task.format = ZIP
	if err := task.Execute(); !errors.Is(err, ErrDestinationRequired) {
		t.Errorf("expected ErrDestinationRequired, got %v", err)
	}

	// 格式无法检测
	task = Compress("file.txt").To("output.unknown") // 格式无法检测
	task.format = ""
	if err := task.Execute(); !errors.Is(err, ErrFormatNotSpecified) {
		t.Errorf("expected ErrFormatNotSpecified, got %v", err)
	}
}

// TestExecuteSuccess 执行成功路径（使用 mockProcessor）
func TestExecuteSuccess(t *testing.T) {
	mock := &mockProcessor{}
	processors[ZIP] = mock

	// 测试压缩
	task := Compress("file.txt").To("file.zip").WithFormat(ZIP)
	if err := task.Execute(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !mock.compressCalled {
		t.Error("expected compressCalled=true")
	}

	// 测试解压
	mock = &mockProcessor{}
	processors[ZIP] = mock
	task = Decompress("file.zip").To("outdir").WithFormat(ZIP)
	if err := task.Execute(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !mock.decompressCalled {
		t.Error("expected decompressCalled=true")
	}
}
