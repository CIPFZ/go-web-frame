# 外部登录

除了账号密码，登录页可以提供“使用 XX 登录”（企业单点登录、OAuth 2.0 / OIDC 等）。框架负责跳转、防 CSRF、
账号关联与一次性登录票据；项目只实现一个 `extlogin.Provider` 并登记。没有登记提供方时，登录页与原来一样。

## 流程

1. 登录页按钮 → `GET /api/v1/user/ext-login/<provider>?redirect=<站内地址>`：后端生成随机 state，放进回调地址和只发给回调地址的
   HttpOnly Cookie，然后把浏览器送到 `Provider.LoginURL` 返回的地址。
2. 提供方登录后把浏览器带回 `GET /api/v1/user/ext-callback/<provider>?state=…`（必须原样保留回调地址自带的查询参数）。
   后端核对 state 与 Cookie 一致，调用 `Provider.Exchange` 换回核验过的身份，找到或关联本地用户，签发一次性票据（1 分钟内有效、
   只能用一次；有 Redis 时存在 Redis，多实例可用），带着票据回到登录页。登录凭证本身不出现在任何地址里。
3. 登录页用票据 `POST /api/v1/user/ext-ticket` 换到登录凭证，跳到原来要去的页面。

失败时回到登录页并提示原因：来源无效或超时、提供方核验失败、外部账号没有关联、账号被禁用。

## 账号关联

外部身份按“提供方 + `Identity.Subject`”关联到本地用户（`sys_user_identities`）。关联过的身份直接登录，提供方那边改了用户名也不受影响。
没有关联时按登记时的 `Options`：

| 选项 | 行为 |
|---|---|
| 都不开（默认） | 拒绝，提示“没有关联”。适合只让已有账号使用的场合 |
| `MatchUsername` | 找同名（`Identity.Username`）的本地用户并关联。**只对用户名由组织统一分配、用户不能自己改的提供方开启**，否则别人可以用同名账号登录 |
| `AutoCreate` + `RoleID` | 新建本地用户（随机密码、角色为 `RoleID`）并关联；同名用户已存在且没开 `MatchUsername` 时拒绝 |

删除用户时，它关联的外部身份一起删除。

## 实现一个提供方

```go
package corpsso

// Provider 是一个 OAuth 2.0 授权码模式的提供方示例（地址、凭据都来自部署配置，见 docs/CONFIG.md）。
type Provider struct {
	AuthorizeURL, TokenURL, UserInfoURL string
	ClientID, ClientSecret               string
}

func (p *Provider) Key() string  { return "corp-sso" }
func (p *Provider) Name() string { return "企业账号" }

func (p *Provider) LoginURL(_ context.Context, callback string) (string, error) {
	q := url.Values{"response_type": {"code"}, "client_id": {p.ClientID}, "redirect_uri": {callback}, "scope": {"openid profile email"}}
	return p.AuthorizeURL + "?" + q.Encode(), nil
}

func (p *Provider) Exchange(ctx context.Context, q url.Values) (*extlogin.Identity, error) {
	// 用 q.Get("code") 向 TokenURL 换 access token，再用它读 UserInfoURL；任何一步失败都返回错误
	...
	return &extlogin.Identity{Subject: info.Sub, Username: info.PreferredUsername, NickName: info.Name, Email: info.Email}, nil
}
```

在启动前登记（例如业务模块的路由注册里，配置齐全时才登记）：

```go
if cfg.Configured() {
	extlogin.Register(&corpsso.Provider{...}, extlogin.Options{AutoCreate: true, RoleID: 888})
}
```

提供方的回调地址填 `https://<站点>/api/v1/user/ext-callback/<key>`。后端经前端 nginx 转发时，站点协议取自 `X-Forwarded-Proto`。
