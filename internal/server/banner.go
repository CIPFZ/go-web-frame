package server

import "fmt"

func PrintBanner(address string) {
	fmt.Printf(`
🚀 欢迎使用 gin-vue-admin
📦 当前版本: v2.8.3
📜 文档地址: http://127.0.0.1%s/swagger/index.html
`, address)
}
