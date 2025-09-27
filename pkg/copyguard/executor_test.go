package copyguard

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteCopy_Success 测试文件成功拷贝的场景
func TestExecuteCopy_Success(t *testing.T) {
	// 1. 准备环境
	// 使用 t.TempDir() 创建一个临时目录，测试结束后会自动清理
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	sourceFile := filepath.Join(sourceDir, "source.txt")
	destFile := filepath.Join(destDir, "dest.txt")
	fileContent := []byte("hello, copyguard!")

	// 创建并写入源文件
	err := os.WriteFile(sourceFile, fileContent, 0644)
	require.NoError(t, err, "创建源文件不应失败")

	// 创建一个默认配置的 Manager
	manager, err := NewManager()
	require.NoError(t, err)

	// 2. 执行拷贝
	err = manager.executeCopy(context.Background(), sourceFile, destFile, nil)

	// 3. 验证结果
	require.NoError(t, err, "拷贝操作不应返回错误")

	// 检查目标文件是否存在
	_, err = os.Stat(destFile)
	assert.NoError(t, err, "目标文件应当存在")

	// 读取目标文件内容并与源文件内容比较
	copiedContent, err := os.ReadFile(destFile)
	assert.NoError(t, err, "读取目标文件不应失败")
	assert.Equal(t, fileContent, copiedContent, "目标文件内容应与源文件内容完全一致")
}

// TestExecuteCopy_SourceNotFound 测试源文件不存在的场景
func TestExecuteCopy_SourceNotFound(t *testing.T) {
	destDir := t.TempDir()
	manager, _ := NewManager()

	nonExistentFile := "non-existent-file.txt"
	destFile := filepath.Join(destDir, "dest.txt")

	err := manager.executeCopy(context.Background(), nonExistentFile, destFile, nil)

	// 验证返回的错误是否是 os.ErrNotExist 或包含了它
	assert.Error(t, err, "当源文件不存在时，应当返回错误")
	// 使用 errors.Is 来处理被包装过的错误
	assert.True(t, errors.Is(err, os.ErrNotExist), "返回的错误类型应为 'file not found'") // <--- 这里是修正点
}

// TestExecuteCopy_SourceIsDirectory 测试源是目录的场景
func TestExecuteCopy_SourceIsDirectory(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	manager, _ := NewManager()

	destFile := filepath.Join(destDir, "dest.txt")

	err := manager.executeCopy(context.Background(), sourceDir, destFile, nil)

	// 验证返回的错误是否是我们定义的 ErrSourceIsDirectory
	assert.Error(t, err, "当源是目录时，应当返回错误")
	assert.True(t, errors.Is(err, ErrSourceIsDirectory), "错误类型应为 ErrSourceIsDirectory")
}

// TestExecuteCopy_ZeroByteFile 测试拷贝一个零字节大小的文件
func TestExecuteCopy_ZeroByteFile(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "empty.txt")
	destFile := filepath.Join(destDir, "empty_dest.txt")

	// 创建一个空文件
	f, err := os.Create(sourceFile)
	require.NoError(t, err)
	f.Close()

	manager, _ := NewManager()
	err = manager.executeCopy(context.Background(), sourceFile, destFile, nil)
	require.NoError(t, err, "拷贝空文件不应失败")

	// 验证目标文件存在且大小为 0
	stat, err := os.Stat(destFile)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stat.Size(), "目标文件大小应为 0")
}
