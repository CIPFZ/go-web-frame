package utils

import (
	"context"
	"log"
)

// ShutdownFunc 定义了一个统一的关闭函数签名
type ShutdownFunc func(context.Context) error

// 定义一个通用接口：所有 OTel Provider 都实现 Shutdown(context.Context) error
type shutdowner interface {
	Shutdown(ctx context.Context) error
}

// SafeShutdown 封装 Shutdown，保证 nil 安全，错误打印
func SafeShutdown(p shutdowner) {
	if p == nil {
		return
	}
	if err := p.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown provider: %v", err)
	}
}
