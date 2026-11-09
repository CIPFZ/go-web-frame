package token

import (
	"strings"
	"sync"
)

var (
	exposedMu sync.RWMutex
	exposed   = map[string]string{} // 相对 RouterPrefix 的路径 → HTTP 方法
)

// Expose 声明业务模块中允许授予 API Token 的接口（路径相对 RouterPrefix，如 /orders/list）。
// 只能在包初始化阶段调用；CMS 的 JWT 路由不会因修改数据库中的 API 记录而变得可被 Token 访问。
func Expose(relativePath, method string) {
	exposedMu.Lock()
	defer exposedMu.Unlock()
	exposed[relativePath] = method
}

// Endpoints is the explicit external API allowlist. CMS JWT routes are never
// made token-accessible by adding or editing database API records.
func Endpoints(prefix string) map[string]string {
	prefix = strings.TrimRight(prefix, "/")
	out := map[string]string{TokenInfoPath(prefix): "GET"}
	exposedMu.RLock()
	defer exposedMu.RUnlock()
	for path, method := range exposed {
		out[prefix+path] = method
	}
	return out
}

func AllowsEndpoint(prefix, method, path string) bool {
	return Endpoints(prefix)[path] == method
}

func TokenInfoPath(prefix string) string { return strings.TrimRight(prefix, "/") + "/open/token-info" }
