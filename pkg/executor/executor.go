package executor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ExecutionResult 封装了命令执行的所有结果信息 (保持不变)
type ExecutionResult struct {
	Command   string        // 执行的命令和参数
	Stdout    string        // 标准输出 (如果未重定向到文件)
	Stderr    string        // 标准错误 (如果未重定向到文件)
	ExitCode  int           // 命令退出码
	Success   bool          // 是否成功 (ExitCode == 0)
	Error     error         // 执行过程中发生的错误 (例如：超时、命令未找到)
	Duration  time.Duration // 执行耗时
	StartTime time.Time     // 开始时间
	EndTime   time.Time     // 结束时间
}

// Executor 是命令执行器的配置结构体
type Executor struct {
	// 使用非导出字段来存储命令，避免混淆
	commandSlice  []string
	commandString string

	Env        []string      // 环境变量
	Dir        string        // 工作目录
	Timeout    time.Duration // 超时设置
	Input      io.Reader     // 标准输入
	StdoutFile string        // 标准输出重定向文件路径
	StderrFile string        // 标准错误重定向文件路径
}

// NewExecutor 通过 []string 创建一个新的 Executor 实例 (保留)
func NewExecutor(command []string) *Executor {
	return &Executor{
		commandSlice: command,
	}
}

// NewExecutorFromString 通过单个 string 创建一个新的 Executor 实例 (新增)
// 注意：此方法在类Unix系统上使用 'sh -c' 执行命令，在Windows上使用 'cmd /C'
// 这允许用户使用shell语法，但也意味着命令的解析依赖于系统shell。
func NewExecutorFromString(command string) *Executor {
	return &Executor{
		commandString: command,
	}
}

// Run 执行命令并返回结果
func (e *Executor) Run() *ExecutionResult {
	result := &ExecutionResult{
		ExitCode:  -1,
		StartTime: time.Now(),
	}

	// 1. 设置超时上下文
	ctx := context.Background()
	var cancel context.CancelFunc
	if e.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, e.Timeout)
		defer cancel()
	}

	var cmd *exec.Cmd

	// 2. 根据输入类型创建 Command 对象
	if e.commandString != "" {
		// 用户提供了单个字符串，使用 shell 来执行
		// 这可以正确处理引号和复杂的参数
		// --- 跨平台核心修改 ---
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd", "/C", e.commandString)
		} else {
			cmd = exec.CommandContext(ctx, "/bin/bash", "-c", e.commandString)
		}
		result.Command = e.commandString
	} else if len(e.commandSlice) > 0 {
		// 用户提供了字符串切片，精确控制参数
		cmd = exec.CommandContext(ctx, e.commandSlice[0], e.commandSlice[1:]...)
		result.Command = strings.Join(e.commandSlice, " ")
	} else {
		result.Error = errors.New("command cannot be empty")
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// 3. 设置工作目录和环境变量
	cmd.Dir = e.Dir
	if len(e.Env) > 0 {
		cmd.Env = append(os.Environ(), e.Env...)
	}

	// 4. 设置标准输入
	if e.Input != nil {
		cmd.Stdin = e.Input
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	// 5. 设置标准输出和标准错误
	var stdoutWriter io.Writer = &stdoutBuf
	var stderrWriter io.Writer = &stderrBuf

	if e.StdoutFile != "" {
		file, err := os.Create(e.StdoutFile)
		if err != nil {
			result.Error = err
			return result
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(file)
		stdoutWriter = io.MultiWriter(&stdoutBuf, file)
	}

	if e.StderrFile != "" {
		file, err := os.Create(e.StderrFile)
		if err != nil {
			result.Error = err
			return result
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(file)
		stderrWriter = io.MultiWriter(&stderrBuf, file)
	}

	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	// 6. 启动和等待命令
	if err := cmd.Start(); err != nil {
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	cmdErrChan := make(chan error, 1)
	go func() {
		cmdErrChan <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		<-cmdErrChan
		result.Error = ctx.Err()
		result.ExitCode = -1

	case waitErr := <-cmdErrChan:
		result.Error = waitErr
		if waitErr != nil {
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				result.ExitCode = exitErr.ExitCode()
			}
		} else {
			result.ExitCode = 0
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = result.ExitCode == 0
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	return result
}
