package apispec

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// AuthorizationAction 服务端授权动作。
type AuthorizationAction string

const (
	ActionRead   AuthorizationAction = "read"
	ActionCreate AuthorizationAction = "create"
	ActionUpdate AuthorizationAction = "update"
	ActionDelete AuthorizationAction = "delete"
	ActionReport AuthorizationAction = "report"
	ActionAdmin  AuthorizationAction = "admin"
)

// Sensitivity 副作用与确认等级。
type Sensitivity string

const (
	SensitivityRead           Sensitivity = "read"
	SensitivityWriteSafe      Sensitivity = "write_safe"
	SensitivityWriteSensitive Sensitivity = "write_sensitive"
	SensitivityDestructive    Sensitivity = "destructive"
)

// Visibility 调用面。SELECTED 进入工具面（命令行工具、自动化 Agent 等，由 Command 命名）。
type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityInternal Visibility = "internal"
	VisibilitySelected Visibility = "selected"
)

// Idempotency 实际幂等行为。
type Idempotency string

const (
	IdempotencyNone        Idempotency = "none"
	IdempotencyNatural     Idempotency = "natural"
	IdempotencyOptionalKey Idempotency = "optional_key"
	IdempotencyRequiredKey Idempotency = "required_key"
)

// Authentication 允许的调用身份。
type Authentication string

const (
	AuthAnonymous   Authentication = "anonymous"     // 无需凭据
	AuthUser        Authentication = "user"          // CMS 登录会话
	AuthToken       Authentication = "token"         // API Token
	AuthUserOrToken Authentication = "user_or_token" // 两者之一
)

// MethodPolicy 是方法的机器可读策略，授权、确认、工具面与幂等的唯一来源。
type MethodPolicy struct {
	Action         AuthorizationAction
	Sensitivity    Sensitivity
	Visibility     Visibility
	Idempotency    Idempotency
	Authentication Authentication
	Command        string // 工具面的命令名，SELECTED 必填，如 "orders order get"
}

// 受控动词词典，值为方法名前缀。
var controlledVerbs = map[string]string{
	"list": "List", "get": "Get", "create": "Create", "update": "Update", "upsert": "Upsert",
	"delete": "Delete", "search": "Search", "stop": "Stop", "start": "Start", "retry": "Retry",
	"import": "Import", "scan": "Scan", "watch": "Watch", "exec": "Exec",
	"batch-create": "BatchCreate", "batch-get": "BatchGet", "batch-update": "BatchUpdate", "batch-delete": "BatchDelete",
}

var (
	readOnlyVerbs = []string{"list", "get", "search", "batch-get", "watch"}
	mutatingVerbs = []string{"create", "update", "upsert", "delete", "start", "stop", "retry", "import", "exec", "batch-create", "batch-update", "batch-delete"}

	kebabPlural   = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*s$`)
	pascalName    = regexp.MustCompile(`^[A-Z][A-Za-z0-9]+$`)
	commandFormat = regexp.MustCompile(`^[a-z][a-z0-9-]*( [a-z][a-z0-9-]*){1,2}$`)
)

// verbPrefix 解析受控动词（含 create-from-* 复合动词），返回方法名前缀。
func verbPrefix(verb string) (string, bool) {
	if p, ok := controlledVerbs[verb]; ok {
		return p, true
	}
	if base, mod, ok := strings.Cut(verb, "-from-"); ok && base == "create" && kebabPlural.MatchString(mod+"s") {
		return "CreateFrom", true
	}
	return "", false
}

// validateMethod 校验路由、方法名与策略组合不变式。
func validateMethod(resource, verb, name string, p MethodPolicy) error {
	prefix, ok := verbPrefix(verb)
	if !ok {
		return fmt.Errorf("verb %q is not in the controlled dictionary", verb)
	}
	if !kebabPlural.MatchString(resource) {
		return fmt.Errorf("resource %q must be a plural kebab-case noun", resource)
	}
	if !pascalName.MatchString(name) || !strings.HasPrefix(name, prefix) {
		return fmt.Errorf("method name %q must be PascalCase starting with %q", name, prefix)
	}
	if !slices.Contains([]AuthorizationAction{ActionRead, ActionCreate, ActionUpdate, ActionDelete, ActionReport, ActionAdmin}, p.Action) {
		return fmt.Errorf("authorization_action %q is invalid", p.Action)
	}
	if !slices.Contains([]Sensitivity{SensitivityRead, SensitivityWriteSafe, SensitivityWriteSensitive, SensitivityDestructive}, p.Sensitivity) {
		return fmt.Errorf("sensitivity %q is invalid", p.Sensitivity)
	}
	if !slices.Contains([]Visibility{VisibilityPublic, VisibilityInternal, VisibilitySelected}, p.Visibility) {
		return fmt.Errorf("visibility %q is invalid", p.Visibility)
	}
	if !slices.Contains([]Idempotency{IdempotencyNone, IdempotencyNatural, IdempotencyOptionalKey, IdempotencyRequiredKey}, p.Idempotency) {
		return fmt.Errorf("idempotency %q is invalid", p.Idempotency)
	}
	if !slices.Contains([]Authentication{AuthAnonymous, AuthUser, AuthToken, AuthUserOrToken}, p.Authentication) {
		return fmt.Errorf("authentication %q is invalid", p.Authentication)
	}

	isRead := p.Action == ActionRead
	switch {
	case isRead != (p.Sensitivity == SensitivityRead):
		return fmt.Errorf("authorization_action=read must pair with sensitivity=read")
	case isRead && p.Idempotency != IdempotencyNatural:
		return fmt.Errorf("read methods must be idempotency=natural")
	case isRead && slices.Contains(mutatingVerbs, verb):
		return fmt.Errorf("mutating verb %q cannot be read-only", verb)
	case !isRead && slices.Contains(readOnlyVerbs, verb):
		return fmt.Errorf("read verb %q cannot have side effects", verb)
	case p.Action == ActionDelete && p.Sensitivity != SensitivityDestructive:
		return fmt.Errorf("authorization_action=delete requires sensitivity=destructive")
	case p.Action == ActionReport && p.Visibility != VisibilityInternal:
		return fmt.Errorf("authorization_action=report requires visibility=internal")
	case p.Action == ActionAdmin && p.Visibility != VisibilityPublic:
		return fmt.Errorf("authorization_action=admin cannot be internal or selected")
	case p.Action == ActionAdmin && p.Authentication != AuthUser:
		return fmt.Errorf("admin methods require a CMS user session")
	case p.Authentication == AuthAnonymous && !isRead:
		return fmt.Errorf("anonymous access is only allowed for read methods")
	case p.Visibility == VisibilitySelected && !commandFormat.MatchString(p.Command):
		return fmt.Errorf("selected methods require a command like \"<scope> <resource> <verb>\"")
	case p.Visibility == VisibilitySelected && !isRead && p.Idempotency == IdempotencyNone:
		return fmt.Errorf("selected write methods must be idempotent or accept a key")
	}
	return nil
}
