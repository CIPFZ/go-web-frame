# API Token 使用和接入

API Token 用于脚本及外部服务，CMS 登录和管理仍使用 JWT + Casbin。
只有代码明确开放的接口才能授权。`sys_apis` 中新增一个接口不会改变其鉴权方式。

## 当前可调用接口

`GET /api/v1/open/token-info`（随 RouterPrefix 配置改变前缀）

请求头：`X-API-Token: <token>`。成功返回 `{"code":0,"data":{"tokenId":123},"msg":"..."}`。
此接口用于验证脚本配置和凭据是否有效，不暴露用户、服务器或管理信息。
Token 必须被明确授予此接口，JWT 登录凭据不能代替 API Token。

在后台创建 Token，设置有效期、并发上限和授权接口，复制一次性显示的明文。
后台的 `POST /sys/api-token/options` 只返回明确开放的 API；接口管理仍列出完整的 CMS API。
创建、更新都会在数据库事务内重新校验授权目标；拒绝未开放接口、无效 ID、空白名称、
缺失/过去的过期时间、永久有效，以及 1–1000 以外的并发配置。

## 生命周期和权限边界

- 数据库只保存 SHA256 摘要和展示前缀；列表、详情和操作日志不返回明文。
- 创建/重置时返回一次明文；重置只更新密钥，不覆盖并发修改的权限。
- 禁用、删除、重置、过期或撤销授权后，后续请求被拒绝。已通过鉴权的在途请求不强制终止。
- Token 不继承创建者的角色，不包含用户身份。删除创建者不会自动撤销其创建的服务凭据；管理员应显式禁用/删除凭据。
- 匹配 HTTP 方法和 Gin 注册路径；没有授权、没有明确开放的接口均不能访问。
- 为兼容 CMS，认证失败使用 HTTP 200 + `code=1003`，权限不足 `code=1004`。
  错误请求方法返回 404。并发超限使用 HTTP 429，Redis 限流不可用使用 HTTP 503，均返回 JSON。
- Redis 启用时跨实例共享并发租约；否则只保证单进程并发。处理器必须遵守请求 context 的取消信号。
- 老版本永久 Token 和旧授权不会被迁移自动扩大权限；编辑时需改为有效期并删除未开放授权。

## 新接口接入

1. 在 `internal/core/token/endpoints.go` 添加明确的 method/path 白名单。
2. 在路由层注册业务 handler，并使用 `ApiTokenAuth`；不要接入 CMS 的 JWT/Casbin 组。
3. 同步 seed 和增量迁移，写入 API 目录；不给既有 Token 自动增加授权。
4. handler 若操作任务或业务数据，仍须校验该 Token 对具体资源的访问范围；目前的授权粒度是接口，非资源实例。
5. 增加实际路由测试，覆盖无凭据、错误凭据、无授权、错误方法以及资源越权。

OpenAPI 使用独立的 `ApiTokenAuth` 安全定义。管理权限由角色控制，基础普通角色无 Token 管理权限。

## 验证

实际路由测试：`internal/core/server/api_token_test.go`。
服务回归：`internal/modules/system/service/api_token_service_test.go`，包含模拟重置与撤权交错。
并发回归：`internal/middleware/api_token_test.go` 和 `internal/core/token`。
浏览器/真实 HTTP：`front-end/e2e/token-notice-security.spec.ts`，使用隔离数据库运行。
