# Archiver 🗜️

[![Go Report Card](https://goreportcard.com/badge/github.com/your-repo/go-archiver)](https://goreportcard.com/report/github.com/your-repo/go-archiver)
[![GoDoc](https://godoc.org/github.com/your-repo/go-archiver/pkg/archiver?status.svg)](https://godoc.org/github.com/your-repo/go-archiver/pkg/archiver)
[![Test Coverage](https://img.shields.io/badge/coverage-98%25+-brightgreen.svg)](https://goreportcard.com/report/github.com/your-repo/go-archiver)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Archiver** 是一个现代化、高性能的 Go 归档库。它旨在通过一个统一、简洁的链式 API，优雅地解决 Go 语言中常见的压缩与解压需求。无论是简单的文件打包，还是需要并发处理、强加密和实时进度的复杂场景，Archiver 都能轻松胜任。

---

### ✨ 核心特性

* **现代化 API**: 采用链式调用 (Fluent API)，让归档操作如自然语言般流畅、易读。
* **统一接口**: 无论是 `zip`, `tar.gz` 还是 `tar.zst`，都使用完全相同的 API 进行操作，无需关心底层细节。
* **高性能并发**: 自动利用多核 CPU 并发压缩文件，并通过工作池 (Worker Pool) 控制资源占用，兼顾速度与稳定。
* **强加密支持**:
  * 为 `.zip` 提供 **AES-256** 加密，兼容主流压缩软件。
  * 为 `.tar.*` 系列格式提供基于 **AES-256-GCM** 和 **PBKDF2** 的安全流加密。
* **实时进度反馈**: 支持注册回调函数，轻松实现进度条或向用户报告实时进度。
* **安全第一**: 内置“Zip Slip”路径遍历攻击防御，确保解压过程的安全性。

---

### 📦 支持的格式

| 格式 | 压缩 | 解压 | 加密 | 并发压缩 | 备注 |
| :--- | :---: | :---: | :---: | :---: | :--- |
| `.zip` | ✅ | ✅ | ✅ | ✅ | 采用 AES-256 加密 |
| `.tar.gz` | ✅ | ✅ | ❌ | ✅ | 经典 `gzip` 格式，通用性最强 |
| `.tar.zst` | ✅ | ✅ | ✅ | ✅ | 高性能 `zstd` 格式，现代 Linux 首选 |

---

### 💾 安装

```bash
go get [github.com/your-repo/go-archiver/pkg/archiver](https://github.com/your-repo/go-archiver/pkg/archiver)
```
*(请将 `your-repo/go-archiver` 替换为您自己的仓库路径)*

---

### 🚀 使用示例

#### 示例 1: 压缩为 `.tar.gz` (最常用)

```go
package main

import "[github.com/your-repo/go-archiver/pkg/archiver](https://github.com/your-repo/go-archiver/pkg/archiver)"

func main() {
    // 假设我们有 a.txt 文件和 my_dir 目录
	err := archiver.Compress("a.txt", "my_dir").
		To("backup.tar.gz").
		Execute()

	if err != nil {
		// 处理错误
	}
}
```

#### 示例 2: 创建带密码的 `.zip` 文件

```go
package main

import "[github.com/your-repo/go-archiver/pkg/archiver](https://github.com/your-repo/go-archiver/pkg/archiver)"

func main() {
	err := archiver.Compress("secret-docs/", "project.go").
		To("secret-archive.zip").
		WithPassword("my-super-strong-password-!@#$%").
		Execute()

	if err != nil {
		// 处理错误
	}
}
```

#### 示例 3: 解压加密文件并显示进度条

这个例子展示了如何解压一个加密的 `.tar.zst` 文件，并集成一个流行的进度条库 `progressbar`。

*首先获取进度条库: `go get github.com/schollz/progressbar/v3`*

```go
package main

import (
	"fmt"
	"[github.com/your-repo/go-archiver/pkg/archiver](https://github.com/your-repo/go-archiver/pkg/archiver)"
	"[github.com/schollz/progressbar/v3](https://github.com/schollz/progressbar/v3)"
)

func main() {
	// 初始化一个进度条
	bar := progressbar.NewOptions64(
		-1, // 由于 tar.zst 解压时总大小未知，设为 -1
		progressbar.OptionSetDescription("正在解压..."),
		progressbar.OptionSetBytes(-1), // 同样设为 -1
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
	)

	// 定义进度回调函数
	progressFunc := func(p archiver.Progress) {
		// 首次回调时，如果我们能知道总字节数，就设置进度条的最大值
		if p.TotalBytes > 0 && bar.GetMax64() == -1 {
			bar.ChangeMax64(p.TotalBytes)
		}
		// 更新进度条的当前值
		bar.Set64(p.BytesProcessed)
		bar.Describe(fmt.Sprintf("正在解压: %s", p.CurrentFile))
	}

	err := archiver.Decompress("encrypted_archive.tar.zst").
		To("output_directory").
		WithPassword("my-super-strong-password-!@#$%").
		WithProgress(progressFunc).
		Execute()

	if err != nil {
		// 处理错误
	}
    fmt.Println("\n解压完成!")
}
```

---

### 📜 API 概览

本库 API 的核心是 `Task` 结构体，通过链式调用进行配置：

* **入口**: `archiver.Compress(...)` 或 `archiver.Decompress(...)`
* **配置**:
  * `.To(destination)`: **(必须)** 设置目标路径
  * `.WithFormat(format)`: (可选) 显式指定格式
  * `.WithPassword(password)`: (可选) 设置密码
  * `.WithCPU(limit)`: (可选) 设置并发数
  * `.WithLogger(logger)`: (可选) 注入 `slog` 日志记录器
  * `.WithProgress(fn)`: (可选) 注册进度回调
  * `.WithContext(ctx)`: (可选) 传入 `context` 以支持取消
* **执行**: `.Execute()`

---

### 🤝 贡献

我们欢迎任何形式的贡献，无论是提交 Issue 还是 Pull Request。

---

### 📄 许可证

本项目基于 [MIT License](LICENSE) 授权。