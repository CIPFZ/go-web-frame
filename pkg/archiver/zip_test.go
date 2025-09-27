package archiver

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// setupTestFiles 是一个辅助函数，用于在指定目录创建测试文件结构
func setupTestFiles(t *testing.T, rootDir string) {
	t.Helper()
	// 创建文件
	err := os.WriteFile(filepath.Join(rootDir, "file1.txt"), []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	// 创建子目录和文件
	err = os.Mkdir(filepath.Join(rootDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}
	err = os.WriteFile(filepath.Join(rootDir, "subdir", "file2.txt"), []byte("world"), 0644)
	if err != nil {
		t.Fatalf("创建子目录文件失败: %v", err)
	}
}

// verifyZipContents 辅助函数用于验证 zip 文件的内容
func verifyZipContents(t *testing.T, zipPath string, expectedFiles map[string]string) {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("打开 zip 文件失败: %v", err)
	}
	defer r.Close()

	if len(r.File) != len(expectedFiles) {
		t.Fatalf("zip 文件数量不匹配: 期望 %d, 得到 %d", len(expectedFiles), len(r.File))
	}

	foundFiles := make(map[string]bool)
	for _, f := range r.File {
		expectedContent, ok := expectedFiles[f.Name]
		if !ok {
			t.Errorf("在 zip 中发现未预期的文件: %s", f.Name)
			continue
		}
		foundFiles[f.Name] = true

		rc, err := f.Open()
		if err != nil {
			t.Fatalf("打开 zip 内文件失败 '%s': %v", f.Name, err)
		}
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("读取 zip 内文件内容失败 '%s': %v", f.Name, err)
		}

		if buf.String() != expectedContent {
			t.Errorf("文件 '%s' 内容不匹配: 期望 '%s', 得到 '%s'", f.Name, expectedContent, buf.String())
		}
	}

	if len(foundFiles) != len(expectedFiles) {
		t.Errorf("zip 中找到的文件数量与预期不符")
	}
}

// TestZipCompressDecompressCycle 测试一个完整的压缩和解压流程
func TestZipCompressDecompressCycle(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	zipPath := filepath.Join(t.TempDir(), "test.zip")
	setupTestFiles(t, sourceDir)

	taskOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cpuLimit: 2,
	}
	zp := &zipProcessor{}

	err := zp.Compress(context.Background(), []string{sourceDir}, zipPath, taskOpts)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	expected := map[string]string{
		"file1.txt":        "hello",
		"subdir/file2.txt": "world",
	}
	verifyZipContents(t, zipPath, expected)

	err = zp.Decompress(context.Background(), zipPath, destDir, taskOpts)
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(content) != "hello" {
		t.Fatalf("解压后的 file1.txt 内容不匹配或读取失败")
	}
	content, err = os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil || string(content) != "world" {
		t.Fatalf("解压后的 file2.txt 内容不匹配或读取失败")
	}
}

// TestDecompress_PathTraversal 安全测试：防止路径遍历攻击
func TestDecompress_PathTraversal(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	_, err := zw.Create("../evil.txt")
	if err != nil {
		t.Fatalf("创建恶意 zip 条目失败: %v", err)
	}
	zw.Close()

	maliciousZipPath := filepath.Join(t.TempDir(), "malicious.zip")
	err = os.WriteFile(maliciousZipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("写入恶意 zip 文件失败: %v", err)
	}

	destDir := t.TempDir()
	zp := &zipProcessor{}
	err = zp.Decompress(context.Background(), maliciousZipPath, destDir, &Task{logger: slog.Default()})

	if err == nil {
		t.Fatal("期望解压时发生路径不安全的错误，但没有发生")
	}
	if !strings.Contains(err.Error(), "检测到不安全的路径") {
		t.Errorf("期望错误信息包含'检测到不安全的路径'，实际为: %v", err)
	}

	evilFilePath := filepath.Join(filepath.Dir(destDir), "evil.txt")
	if _, err := os.Stat(evilFilePath); !os.IsNotExist(err) {
		t.Fatal("安全漏洞！恶意文件被创建在了目标目录之外！")
	}
}

func TestCompress_ContextCancellation(t *testing.T) {
	sourceDir := t.TempDir()
	setupTestFiles(t, sourceDir)
	zipPath := filepath.Join(t.TempDir(), "cancel.zip")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond)

	zp := &zipProcessor{}
	err := zp.Compress(ctx, []string{sourceDir}, zipPath, &Task{logger: slog.Default(), cpuLimit: 1})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("期望错误为 context.DeadlineExceeded, 实际为: %v", err)
	}
}

func TestSecureJoin(t *testing.T) {
	base := "/safe/path"
	p, err := secureJoin(base, "file.txt")
	if err != nil || p != filepath.Join(base, "file.txt") {
		t.Errorf("合法的 join 失败")
	}
	_, err = secureJoin(base, "../unsafe.txt")
	if err == nil {
		t.Errorf("非法的 join 应该返回错误")
	}
}

// TestZipCompressDecompressCycle_WithPassword 测试带密码的压缩和解压流程
func TestZipCompressDecompressCycle_WithPassword(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	archivePath := filepath.Join(t.TempDir(), "test-encrypted.zip")
	setupTestFiles(t, sourceDir)

	password := "my-super-secret-password-!@#$"
	processor := &zipProcessor{}

	compressOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cpuLimit: 2,
		password: password,
	}
	if err := processor.Compress(context.Background(), []string{sourceDir}, archivePath, compressOpts); err != nil {
		t.Fatalf("带密码压缩失败: %v", err)
	}

	decompressOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		password: password,
	}
	if err := processor.Decompress(context.Background(), archivePath, destDir, decompressOpts); err != nil {
		t.Fatalf("使用正确密码解压失败: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(content) != "hello" {
		t.Fatalf("正确密码解压后，file1.txt 内容不匹配或读取失败")
	}
	content, err = os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil || string(content) != "world" {
		t.Fatalf("正确密码解压后，file2.txt 内容不匹配或读取失败")
	}

	wrongPassOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		password: "this-is-the-wrong-password",
	}
	err = processor.Decompress(context.Background(), archivePath, t.TempDir(), wrongPassOpts)
	if err == nil {
		t.Fatal("期望使用错误密码解压时发生错误，但没有发生")
	}
	t.Logf("使用错误密码时捕获到预期错误: %v", err)

	noPassOpts := &Task{
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	err = processor.Decompress(context.Background(), archivePath, t.TempDir(), noPassOpts)
	if err == nil {
		t.Fatal("期望不提供密码解压加密文件时发生错误，但没有发生")
	}
	t.Logf("不提供密码时捕获到预期错误: %v", err)
}

// TestZip_WithProgress 测试带进度回调的压缩和解压
func TestZip_WithProgress(t *testing.T) {
	sourceDir := t.TempDir()
	setupTestFiles(t, sourceDir)
	zipPath := filepath.Join(t.TempDir(), "progress.zip")

	// --- 测试压缩进度 ---
	var compressUpdates []Progress
	var mu sync.Mutex
	compressProgressFunc := func(p Progress) {
		mu.Lock()
		defer mu.Unlock()
		// 打印日志以便调试
		t.Logf("压缩进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		compressUpdates = append(compressUpdates, p)
	}

	task := Compress(sourceDir).To(zipPath).WithProgress(compressProgressFunc)
	err := task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的压缩失败: %v", err)
	}

	if len(compressUpdates) == 0 {
		t.Fatal("压缩进度回调函数从未被调用")
	}

	// 验证最终状态
	lastCompressUpdate := compressUpdates[len(compressUpdates)-1]
	if lastCompressUpdate.TotalFiles != 2 {
		t.Errorf("压缩: 期望总文件数为 2, 得到 %d", lastCompressUpdate.TotalFiles)
	}
	if lastCompressUpdate.TotalBytes != 10 { // "hello" (5) + "world" (5)
		t.Errorf("压缩: 期望总字节数为 10, 得到 %d", lastCompressUpdate.TotalBytes)
	}

	// --- 测试解压进度 ---
	destDir := t.TempDir()
	var decompressUpdates []Progress
	decompressProgressFunc := func(p Progress) {
		t.Logf("解压进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		decompressUpdates = append(decompressUpdates, p)
	}

	task = Decompress(zipPath).To(destDir).WithProgress(decompressProgressFunc)
	err = task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的解压失败: %v", err)
	}

	if len(decompressUpdates) < 2 { // 至少应有 (开始file1, 开始file2, 完成) 3次回调
		t.Fatal("解压进度回调函数调用次数不足")
	}

	lastDecompressUpdate := decompressUpdates[len(decompressUpdates)-1]
	if lastDecompressUpdate.TotalFiles != 2 {
		t.Errorf("解压: 期望总文件数为 2, 得到 %d", lastDecompressUpdate.TotalFiles)
	}
	if lastDecompressUpdate.TotalBytes != 10 {
		t.Errorf("解压: 期望总字节数为 10, 得到 %d", lastDecompressUpdate.TotalBytes)
	}
	if lastDecompressUpdate.FilesProcessed != 2 {
		t.Errorf("解压: 期望最终已处理文件数为 2, 得到 %d", lastDecompressUpdate.FilesProcessed)
	}
	if lastDecompressUpdate.BytesProcessed != 10 {
		t.Errorf("解压: 期望最终已处理字节数为 10, 得到 %d", lastDecompressUpdate.BytesProcessed)
	}
}
