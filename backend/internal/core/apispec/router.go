package apispec

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxPolicy    = "apispec.policy"
	ctxMethod    = "apispec.method"
	ctxPrincipal = "apispec.principal"
)

// PrincipalKind 调用方身份类别；审计中据此区分人工与 Agent。
type PrincipalKind string

const (
	PrincipalAnonymous PrincipalKind = "anonymous"
	PrincipalUser      PrincipalKind = "user"
	PrincipalToken     PrincipalKind = "token"
)

// Principal 是认证后的调用方。
type Principal struct {
	Kind PrincipalKind
	ID   string // user uuid 或 token id
}

func (p Principal) String() string { return string(p.Kind) + ":" + p.ID }

func SetPrincipal(c *gin.Context, p Principal) { c.Set(ctxPrincipal, p) }

// PrincipalOf 返回当前调用方，未认证时为匿名。
func PrincipalOf(c *gin.Context) Principal {
	if v, ok := c.Get(ctxPrincipal); ok {
		return v.(Principal)
	}
	return Principal{Kind: PrincipalAnonymous, ID: c.ClientIP()}
}

// PolicyOf 返回当前方法策略。
func PolicyOf(c *gin.Context) MethodPolicy {
	v, _ := c.Get(ctxPolicy)
	p, _ := v.(MethodPolicy)
	return p
}

// MethodInfo 描述一个已注册的契约方法，用于生成文档与工具面（命令行工具、自动化 Agent 等）。
type MethodInfo struct {
	Path     string
	Resource string
	Verb     string
	Name     string
	Policy   MethodPolicy
}

// Router 注册契约方法：路由 /api/<version>/<service>/<resources>/<verb>，统一 POST。
// 策略在注册时校验，非法即 panic（启动失败，fail-closed）。
type Router struct {
	group          *gin.RouterGroup
	authenticators map[Authentication]gin.HandlerFunc
	idempotency    gin.HandlerFunc
	audit          gin.HandlerFunc
	methods        []MethodInfo
	commands       map[string]string
}

// RouterOptions 由业务模块提供：各身份的认证中间件、幂等中间件与写操作审计中间件。
// AuthAnonymous 对应的认证器（可选）用于在携带凭据时识别调用方。
type RouterOptions struct {
	Authenticators map[Authentication]gin.HandlerFunc
	Idempotency    gin.HandlerFunc
	Audit          gin.HandlerFunc // 只作用于非只读方法
}

func NewRouter(group *gin.RouterGroup, opts RouterOptions) *Router {
	group.Use(Observe())
	return &Router{group: group, authenticators: opts.Authenticators, idempotency: opts.Idempotency, audit: opts.Audit, commands: map[string]string{}}
}

// Handle 注册方法，如 Handle("orders", "get", "GetOrder", policy, h)。
func (r *Router) Handle(resource, verb, name string, p MethodPolicy, h gin.HandlerFunc) {
	path := "/" + resource + "/" + verb
	if err := validateMethod(resource, verb, name, p); err != nil {
		panic(fmt.Sprintf("apispec: %s %s: %v", name, path, err))
	}
	auth, ok := r.authenticators[p.Authentication]
	if !ok && p.Authentication != AuthAnonymous {
		panic(fmt.Sprintf("apispec: %s: no authenticator for %q", name, p.Authentication))
	}
	if p.Command != "" {
		if other, dup := r.commands[p.Command]; dup {
			panic(fmt.Sprintf("apispec: command %q used by both %s and %s", p.Command, other, name))
		}
		r.commands[p.Command] = name
	}
	usesKey := p.Idempotency == IdempotencyOptionalKey || p.Idempotency == IdempotencyRequiredKey
	if usesKey && r.idempotency == nil {
		panic(fmt.Sprintf("apispec: %s requires idempotency middleware", name))
	}

	chain := []gin.HandlerFunc{func(c *gin.Context) {
		c.Set(ctxPolicy, p)
		c.Set(ctxMethod, name)
		c.Next()
	}}
	if auth != nil {
		chain = append(chain, auth)
	}
	if r.audit != nil && p.Action != ActionRead {
		chain = append(chain, r.audit)
	}
	if usesKey {
		chain = append(chain, r.idempotency)
	}
	chain = append(chain, h)
	r.group.POST(path, chain...)
	r.methods = append(r.methods, MethodInfo{Path: r.group.BasePath() + path, Resource: resource, Verb: verb, Name: name, Policy: p})
}

// Methods 返回已注册方法（按路径排序）。
func (r *Router) Methods() []MethodInfo {
	out := slices.Clone(r.methods)
	slices.SortFunc(out, func(a, b MethodInfo) int { return strings.Compare(a.Path, b.Path) })
	return out
}

// MethodName 返回当前请求对应的方法名。
func MethodName(c *gin.Context) string { return c.GetString(ctxMethod) }
