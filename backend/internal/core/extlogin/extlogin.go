// Package extlogin 是外部登录提供方（企业单点登录、OAuth 等）的扩展点：项目实现 Provider 并在启动前 Register，
// 登录页就会出现对应的按钮。框架负责跳转、防 CSRF 的 state、账号关联与一次性登录票据；提供方只负责
// “把浏览器送去登录”和“用回调参数换回核验过的身份”。见 docs/EXTERNAL_LOGIN.md。
package extlogin

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"sync"
)

// Identity 是提供方核验过的身份。
type Identity struct {
	// Subject 是提供方内稳定、唯一的用户标识（不随改名变化），账号关联按它进行
	Subject string
	// Username 是建议的本地用户名（自动建号、按用户名匹配时使用）
	Username string
	NickName string
	Email    string
}

// Provider 是一个外部登录提供方。
type Provider interface {
	// Key 是路由里的标识：小写字母、数字与横线，如 "corp-sso"
	Key() string
	// Name 是登录页按钮上的名字
	Name() string
	// LoginURL 返回浏览器要去的外部登录页。登录后外部系统要把浏览器带回 callback，并原样保留 callback 自带的查询参数。
	LoginURL(ctx context.Context, callback string) (string, error)
	// Exchange 处理回调（query 是回调请求的查询参数，如 code），返回核验过的身份；核验失败返回错误。
	Exchange(ctx context.Context, query url.Values) (*Identity, error)
}

// Options 决定外部身份怎样对应到本地账号。默认（都不开）只接受已经关联过的身份。
type Options struct {
	// MatchUsername：没有关联时，按 Identity.Username 找同名的本地用户并关联。
	// 只对用户名由组织统一分配、不能自己改的提供方开启（否则别人可以冒用同名账号）。
	MatchUsername bool
	// AutoCreate：没有关联、也没有匹配到时，新建本地用户（随机密码，角色为 RoleID）并关联
	AutoCreate bool
	// RoleID 是自动新建的用户的角色；AutoCreate 时必填
	RoleID uint
}

type entry struct {
	provider Provider
	options  Options
}

var (
	mu        sync.RWMutex
	providers = map[string]entry{}
	keyFormat = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
)

// Register 登记一个提供方；只能在启动前（包初始化或路由注册前）调用。配置不合法时 panic（启动失败）。
func Register(p Provider, opt Options) {
	if !keyFormat.MatchString(p.Key()) {
		panic(fmt.Sprintf("extlogin: provider key %q must be lowercase letters, digits and dashes", p.Key()))
	}
	if opt.AutoCreate && opt.RoleID == 0 {
		panic(fmt.Sprintf("extlogin: provider %q creates users but has no RoleID", p.Key()))
	}
	mu.Lock()
	defer mu.Unlock()
	if _, dup := providers[p.Key()]; dup {
		panic(fmt.Sprintf("extlogin: provider %q registered twice", p.Key()))
	}
	providers[p.Key()] = entry{provider: p, options: opt}
}

// Get 返回登记过的提供方。
func Get(key string) (Provider, Options, bool) {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := providers[key]
	return e.provider, e.options, ok
}

// Info 是登录页显示的提供方。
type Info struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// List 返回登记过的提供方（按 Key 排序）。
func List() []Info {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Info, 0, len(providers))
	for _, e := range providers {
		out = append(out, Info{Key: e.provider.Key(), Name: e.provider.Name()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// reset 清空登记（测试用）。
func reset() {
	mu.Lock()
	providers = map[string]entry{}
	mu.Unlock()
}
