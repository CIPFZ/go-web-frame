# 资源接口契约（core/apispec）

`backend/internal/core/apispec` 是一套面向资源的接口约定，新模块可以选用；系统管理模块继续使用原有的
`{code,data,msg}` 格式，两者并存。

## 路由与方法

```text
POST /api/v1/<模块>/<资源复数>/<动词>      例：POST /api/v1/orders/orders/get
```

- 一律 POST，参数放在 JSON 请求体里。动词来自固定词典：`list get search create update upsert delete start stop
  retry import scan watch exec batch-get batch-create batch-update batch-delete create-from-<资源>`。
- 资源名是复数、kebab-case；方法名是 PascalCase，并以动词开头（`get` → `GetOrder`）。
- 每个方法带一份机器可读的策略（`MethodPolicy`）：授权动作、副作用等级、可见范围、幂等行为、允许的调用身份，
  以及进入工具面时的命令名。策略在注册时检查（例如只读方法必须天然幂等、删除必须标为 destructive、
  匿名只能读、管理类方法只能由登录用户调用），不合法时启动失败。
- `Router.Methods()` 列出已注册的方法，可用来生成文档或命令行工具。

```go
r := apispec.NewRouter(engine.Group("/api/v1/orders"), apispec.RouterOptions{
	Authenticators: middleware.ContractAuthenticators(svcCtx), // 会话、API Token、两者之一、匿名可读
	Idempotency:    apispec.IdempotencyMiddleware(apispec.NewRedisIdempotencyStore(redisClient, "orders:idem:")),
	Audit:          middleware.ContractAudit(svcCtx, "orders"),  // 写操作记操作日志
})
r.Handle("orders", "get", "GetOrder", apispec.MethodPolicy{
	Action: apispec.ActionRead, Sensitivity: apispec.SensitivityRead, Visibility: apispec.VisibilityPublic,
	Idempotency: apispec.IdempotencyNatural, Authentication: apispec.AuthUserOrToken,
}, getOrder)
```

会话按用户的全部角色授权当前路由；API Token 按“Token 的权限 ∩ 主人现在的权限”，可被 Token 调用的接口
用 `token.Expose` 声明（见 [API Token](apitoken-design.md)）。接口仍要登记到 API 目录并授权（见 [迁移](MIGRATIONS.md)
的 `syncAPICatalog`）。

## 响应

```json
{
  "code": 0, "code_name": "", "message": "success", "data": {...},
  "meta": { "request_id": "req_...", "latency_ms": 12, "trace_id": "..." },
  "error": { "reason": "order.already_paid", "retryable": false, "field_violations": [...], "metadata": {...} }
}
```

- HTTP 状态码是真实的（400、401、403、404、409、429、503……）；`code` 是标准错误码（100001 起），模块细分的原因写在
  `error.reason`（稳定的点分字符串，客户端据此判断），不增加数字码。
- 服务层返回 `*apispec.Error`（`InvalidParameter`、`NotFound`、`FailedPrecondition`、`PermissionDenied`……），
  `WriteError` 统一转换；其他错误一律变成“内部错误”，内部细节只进日志。
- 请求 ID 与耗时同时写在响应头和 `meta` 里；客户端可以传入自己的请求 ID。

## 输入、分页、排序、幂等、资源标识

- `Decode` 严格解码请求体：未知字段、类型不符、超过大小上限都会被拒绝；`validate` 标签的校验失败转换为字段错误
  （只说哪个字段、什么问题，不回显被拒绝的值）。
- 分页用不透明的游标：`PageRequest{page_size, page_token}`。游标用服务端密钥签名，并绑定到这次查询的条件，
  换了过滤条件再用旧游标会被拒绝。密钥在启动时由 JWT 密钥派生。
- 排序是结构化的 `[{field, order}]`，只允许 `SortSpec` 里登记的字段。
- 写方法可以声明接受或要求 `Idempotency-Key`：同一个键、同样的请求返回第一次的结果；同一个键、不同的请求返回冲突；
  并发的同键请求等待第一次完成；可重试的失败会释放这个键。键按调用方隔离，多副本部署用 Redis 存储。
- 资源标识用带类型前缀的 ULID（`NewSID("ord")` → `ord_01J…`），不暴露自增 ID。
