package archiver

import (
	"archive/tar"
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// TestTarZstCompressDecompressCycle 测试一个完整的 tar.zst 压缩和解压流程
func TestTarZstCompressDecompressCycle(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	tarzstPath := filepath.Join(t.TempDir(), "test.tar.zst")
	setupTestFiles(t, sourceDir) // 复用 zip 测试的辅助函数

	taskOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cpuLimit: 2,
	}
	tzp := &tarZstProcessor{}

	err := tzp.Compress(context.Background(), []string{sourceDir}, tarzstPath, taskOpts)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	err = tzp.Decompress(context.Background(), tarzstPath, destDir, taskOpts)
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

// TestTarZstDecompress_PathTraversal 安全测试: 防止路径遍历攻击
func TestTarZstDecompress_PathTraversal(t *testing.T) {
	buf := new(bytes.Buffer)
	zstdWriter, _ := zstd.NewWriter(buf)
	tarWriter := tar.NewWriter(zstdWriter)

	hdr := &tar.Header{
		Name:     "../../evil.txt",
		Mode:     0644,
		Size:     int64(len("pwned")),
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(hdr); err != nil {
		t.Fatalf("写入恶意 tar头失败: %v", err)
	}
	if _, err := tarWriter.Write([]byte("pwned")); err != nil {
		t.Fatalf("写入恶意 tar内容失败: %v", err)
	}
	tarWriter.Close()
	zstdWriter.Close()

	maliciousTarPath := filepath.Join(t.TempDir(), "malicious.tar.zst")
	if err := os.WriteFile(maliciousTarPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("写入恶意 tar.zst 文件失败: %v", err)
	}

	destDir := t.TempDir()
	tzp := &tarZstProcessor{}
	err := tzp.Decompress(context.Background(), maliciousTarPath, destDir, &Task{logger: slog.Default()})

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

// TestTarZstCompressDecompressCycle_WithPassword 测试带密码的 tar.zst 压缩和解压流程
func TestTarZstCompressDecompressCycle_WithPassword(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	archivePath := filepath.Join(t.TempDir(), "test-encrypted.tar.zst")
	setupTestFiles(t, sourceDir)

	password := "another-very-secure-password-&*%^"
	processor := &tarZstProcessor{}

	compressOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cpuLimit: 2,
		password: password,
	}
	if err := processor.Compress(context.Background(), []string{sourceDir}, archivePath, compressOpts); err != nil {
		t.Fatalf("带密码压缩 tar.zst 失败: %v", err)
	}

	decompressOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		password: password,
	}
	if err := processor.Decompress(context.Background(), archivePath, destDir, decompressOpts); err != nil {
		t.Fatalf("使用正确密码解压 tar.zst 失败: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(content) != "hello" {
		t.Fatalf("正确密码解压 tar.zst 后，file1.txt 内容不匹配或读取失败")
	}
	content, err = os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil || string(content) != "world" {
		t.Fatalf("正确密码解压 tar.zst 后，file2.txt 内容不匹配或读取失败")
	}

	wrongPassOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		password: "wrong-password",
	}
	err = processor.Decompress(context.Background(), archivePath, t.TempDir(), wrongPassOpts)
	if err == nil {
		t.Fatal("期望使用错误密码解压 tar.zst 时发生错误，但没有发生")
	}
	t.Logf("使用错误密码解压 tar.zst 时捕获到预期错误: %v", err)

	noPassOpts := &Task{
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	err = processor.Decompress(context.Background(), archivePath, t.TempDir(), noPassOpts)
	if err == nil {
		t.Fatal("期望不提供密码解压 tar.zst 时发生错误，但没有发生")
	}
	t.Logf("不提供密码解压 tar.zst 时捕获到预期错误: %v", err)
}

// TestTarZst_WithProgress 测试带进度回调的 tar.zst 压缩和解压
func TestTarZst_WithProgress(t *testing.T) {
	sourceDir := t.TempDir()
	setupTestFiles(t, sourceDir)
	archivePath := filepath.Join(t.TempDir(), "progress.tar.zst")

	// --- 测试压缩进度 ---
	var compressUpdates []Progress
	var mu sync.Mutex
	compressProgressFunc := func(p Progress) {
		mu.Lock()
		defer mu.Unlock()
		t.Logf("压缩进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		compressUpdates = append(compressUpdates, p)
	}
	task := Compress(sourceDir).To(archivePath).WithFormat(TARZST).WithProgress(compressProgressFunc)
	err := task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的压缩失败: %v", err)
	}
	if len(compressUpdates) == 0 {
		t.Fatal("压缩进度回调函数从未被调用")
	}
	lastCompressUpdate := compressUpdates[len(compressUpdates)-1]
	if lastCompressUpdate.TotalFiles != 2 || lastCompressUpdate.TotalBytes != 10 {
		t.Errorf("压缩: 期望总文件数/字节数为 2/10, 得到 %d/%d", lastCompressUpdate.TotalFiles, lastCompressUpdate.TotalBytes)
	}

	// --- 测试解压进度 ---
	destDir := t.TempDir()
	var decompressUpdates []Progress
	decompressProgressFunc := func(p Progress) {
		t.Logf("解压进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		decompressUpdates = append(decompressUpdates, p)
	}
	task = Decompress(archivePath).To(destDir).WithProgress(decompressProgressFunc)
	err = task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的解压失败: %v", err)
	}
	if len(decompressUpdates) < 2 {
		t.Fatal("解压进度回调函数调用次数不足")
	}

	// 验证 tar.zst 解压时总数未知 (-1)
	firstDecompressUpdate := decompressUpdates[0]
	if firstDecompressUpdate.TotalFiles != -1 || firstDecompressUpdate.TotalBytes != -1 {
		t.Errorf("解压: 期望总数未知 (-1), 得到 %d/%d", firstDecompressUpdate.TotalFiles, firstDecompressUpdate.TotalBytes)
	}

	lastDecompressUpdate := decompressUpdates[len(decompressUpdates)-1]
	if lastDecompressUpdate.FilesProcessed != 2 {
		t.Errorf("解压: 期望最终处理文件数为 2, 得到 %d", lastDecompressUpdate.FilesProcessed)
	}
}
