package token

import "strings"

// Endpoints is the explicit external API allowlist. CMS JWT routes are never
// made token-accessible by adding or editing database API records.
func Endpoints(prefix string) map[string]string {
	return map[string]string{
		TokenInfoPath(prefix):        "GET",
		BrowserBootstrapPath(prefix): "GET",
		BrowserHealthPath(prefix):    "GET",
	}
}
func AllowsEndpoint(prefix, method, path string) bool {
	return Endpoints(prefix)[path] == method
}

func TokenInfoPath(prefix string) string { return strings.TrimRight(prefix, "/") + "/open/token-info" }
func BrowserBootstrapPath(prefix string) string {
	return strings.TrimRight(prefix, "/") + "/proxy/browser/bootstrap"
}
func BrowserHealthPath(prefix string) string {
	return strings.TrimRight(prefix, "/") + "/proxy/browser/health"
}
