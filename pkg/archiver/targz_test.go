package archiver

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestTarGzCompressDecompressCycle 测试一个完整的 tar.gz 压缩和解压流程
func TestTarGzCompressDecompressCycle(t *testing.T) {
	// 1. 准备环境
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	archivePath := filepath.Join(t.TempDir(), "test.tar.gz")

	// 复用已有的辅助函数来创建测试文件
	setupTestFiles(t, sourceDir)

	taskOpts := &Task{
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		cpuLimit: 2,
	}
	processor := &tarGzProcessor{}

	// 2. 测试压缩
	err := processor.Compress(context.Background(), []string{sourceDir}, archivePath, taskOpts)
	if err != nil {
		t.Fatalf("tar.gz 压缩失败: %v", err)
	}

	// 3. 测试解压
	err = processor.Decompress(context.Background(), archivePath, destDir, taskOpts)
	if err != nil {
		t.Fatalf("tar.gz 解压失败: %v", err)
	}

	// 4. 验证解压后的文件内容
	content, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(content) != "hello" {
		t.Fatalf("解压后的 file1.txt 内容不匹配或读取失败")
	}
	content, err = os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil || string(content) != "world" {
		t.Fatalf("解压后的 file2.txt 内容不匹配或读取失败")
	}
}

// TestTarGzDecompress_PathTraversal 安全测试: 防止路径遍历攻击
func TestTarGzDecompress_PathTraversal(t *testing.T) {
	// 1. 手动在内存中创建一个恶意的 tar.gz 文件
	buf := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(buf)
	tarWriter := tar.NewWriter(gzWriter)

	// 创建一个包含向上遍历路径的恶意 tar 头
	hdr := &tar.Header{
		Name:     "../../pwned.txt",
		Mode:     0644,
		Size:     int64(len("pwned")),
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(hdr); err != nil {
		t.Fatalf("写入恶意 tar 头失败: %v", err)
	}
	if _, err := tarWriter.Write([]byte("pwned")); err != nil {
		t.Fatalf("写入恶意 tar 内容失败: %v", err)
	}

	// 关键：必须按正确的顺序关闭写入器
	tarWriter.Close()
	gzWriter.Close()

	maliciousArchivePath := filepath.Join(t.TempDir(), "malicious.tar.gz")
	if err := os.WriteFile(maliciousArchivePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("写入恶意 tar.gz 文件失败: %v", err)
	}

	// 2. 尝试解压恶意文件
	destDir := t.TempDir()
	processor := &tarGzProcessor{}
	err := processor.Decompress(context.Background(), maliciousArchivePath, destDir, &Task{logger: slog.Default()})

	// 3. 验证是否返回了预期的安全错误
	if err == nil {
		t.Fatal("期望解压时发生路径不安全的错误，但没有发生")
	}
	if !strings.Contains(err.Error(), "检测到不安全的路径") {
		t.Errorf("期望错误信息包含'检测到不安全的路径'，实际为: %v", err)
	}

	// 4. 确保恶意文件没有被创建在目标目录之外
	pwnedFilePath := filepath.Join(filepath.Dir(destDir), "pwned.txt")
	if _, err := os.Stat(pwnedFilePath); !os.IsNotExist(err) {
		t.Fatal("安全漏洞！恶意文件被创建在了目标目录之外！")
	}
}

// TestTarGz_WithProgress 测试带进度回调的 tar.gz 压缩和解压
func TestTarGz_WithProgress(t *testing.T) {
	sourceDir := t.TempDir()
	setupTestFiles(t, sourceDir)
	archivePath := filepath.Join(t.TempDir(), "progress.tar.gz")

	// --- 1. 测试压缩进度 ---
	var compressUpdates []Progress
	var mu sync.Mutex
	compressProgressFunc := func(p Progress) {
		mu.Lock()
		defer mu.Unlock()
		t.Logf("压缩进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		compressUpdates = append(compressUpdates, p)
	}

	// 使用公共 API 来进行测试
	task := Compress(sourceDir).To(archivePath).WithFormat(TARGZ).WithProgress(compressProgressFunc)
	err := task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的 tar.gz 压缩失败: %v", err)
	}

	if len(compressUpdates) == 0 {
		t.Fatal("压缩进度回调函数从未被调用")
	}
	// 验证最终的进度报告是否正确
	lastCompressUpdate := compressUpdates[len(compressUpdates)-1]
	if lastCompressUpdate.TotalFiles != 2 || lastCompressUpdate.TotalBytes != 10 {
		t.Errorf("压缩: 期望总文件数/字节数为 2/10, 得到 %d/%d", lastCompressUpdate.TotalFiles, lastCompressUpdate.TotalBytes)
	}

	// --- 2. 测试解压进度 ---
	destDir := t.TempDir()
	var decompressUpdates []Progress
	decompressProgressFunc := func(p Progress) {
		t.Logf("解压进度: %d/%d 文件, %d/%d 字节, 当前: %s", p.FilesProcessed, p.TotalFiles, p.BytesProcessed, p.TotalBytes, p.CurrentFile)
		decompressUpdates = append(decompressUpdates, p)
	}

	task = Decompress(archivePath).To(destDir).WithProgress(decompressProgressFunc)
	err = task.Execute()
	if err != nil {
		t.Fatalf("带进度回调的 tar.gz 解压失败: %v", err)
	}

	if len(decompressUpdates) < 2 {
		t.Fatal("解压进度回调函数调用次数不足")
	}

	// 验证 tar.gz 解压时总数是否如预期般未知 (-1)
	firstDecompressUpdate := decompressUpdates[0]
	if firstDecompressUpdate.TotalFiles != -1 || firstDecompressUpdate.TotalBytes != -1 {
		t.Errorf("解压: 期望总数未知 (-1), 得到 %d/%d", firstDecompressUpdate.TotalFiles, firstDecompressUpdate.TotalBytes)
	}

	// 验证最终处理的文件数是正确的
	lastDecompressUpdate := decompressUpdates[len(decompressUpdates)-1]
	if lastDecompressUpdate.FilesProcessed != 2 {
		t.Errorf("解压: 期望最终处理文件数为 2, 得到 %d", lastDecompressUpdate.FilesProcessed)
	}
}
