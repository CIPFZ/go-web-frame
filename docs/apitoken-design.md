# API Token 使用和接入

API Token 用于脚本及外部服务，CMS 登录和管理仍使用 JWT + Casbin。
只有代码明确开放的接口才能授权。`sys_apis` 中新增一个接口不会改变其鉴权方式。

## 当前可调用接口

`GET /api/v1/open/token-info`（随 RouterPrefix 配置改变前缀）

请求头：`X-API-Token: <token>`。成功返回 `{"code":0,"data":{"tokenId":123},"msg":"..."}`。
此接口用于验证脚本配置和凭据是否有效，不暴露用户、服务器或管理信息。
JWT 登录凭据不能代替 API Token。

## Token 跟人绑定

- 每个人在“我的 API Token”（`/account/tokens`，头像菜单进入）申请、管理自己的 Token；
  管理员在“系统管理 → API Token”看所有人的 Token，并设置申请限制。
- Token 的主人是申请人（`created_by`）。能调用的接口 = Token 的权限 ∩ 主人**现在**的权限：
  主人的角色被收回某接口，Token 也随之不能调用；主人被禁用或删除，Token 立即失效。
- 权限有两种：**跟随本人**（默认，不存接口列表，按“可被 Token 调用的接口”中本人有权限的部分）
  或**自己选接口**（只能从本人有权限的接口里选）。
- 通过 Token 的操作记在主人名下，操作日志同时记录经过的 Token（页面显示“某人 · 通过 Token「名称」”）。
- 申请限制（`sys_token_limits`）：每人最多几个 Token、最长有效期、并发上限。`userId=0` 是全局默认，
  单个用户的设置在“用户管理 → Token 设置”里改，留空跟随默认。管理员不受限制。
  默认值：5 个、180 天、并发 5。限制只影响之后的申请和修改，已有的 Token 不变。
- 申请时只填名字也可以：有效期取本人的最长有效期，并发取上限。

`POST /sys/api-token/options` 返回本人可选的接口、生效的限制与已有的 Token 数。
创建、更新都会在数据库事务内重新校验授权目标；拒绝未开放接口、本人没有的权限、无效 ID、空白名称、
过去的过期时间、永久有效、超过上限的有效期与并发。

## 生命周期和权限边界

- 数据库只保存 SHA256 摘要和展示前缀；列表、详情和操作日志不返回明文。
- 创建/重置时返回一次明文；重置只更新密钥，不覆盖并发修改的权限。
- 禁用、删除、重置、过期、撤销授权或主人失去权限后，后续请求被拒绝。已通过鉴权的在途请求不强制终止。
- 别人的 Token 与不存在的 Token 不可区分（管理员除外）。
- 匹配 HTTP 方法和 Gin 注册路径；没有授权、没有明确开放的接口均不能访问。
- 为兼容 CMS，认证失败使用 HTTP 200 + `code=1003`，权限不足 `code=1004`。
  错误请求方法返回 404。并发超限使用 HTTP 429，Redis 限流不可用使用 HTTP 503，均返回 JSON。
- Redis 启用时跨实例共享并发租约；否则只保证单进程并发。处理器必须遵守请求 context 的取消信号。
- 升级时没有主人的旧 Token 归给最早的超级管理员（迁移 `20261109_personal_tokens_v1`）。

## 新接口接入

1. 业务模块在包初始化时调用 `token.Expose(相对路径, 方法)` 声明可以授予 Token 的接口（`internal/core/token/endpoints.go`）。
2. 在路由层注册业务 handler，并使用 `ApiTokenAuth`；不要接入 CMS 的 JWT/Casbin 组。
3. 同步 seed 和增量迁移，写入 API 目录并授予角色（见 [迁移写法](MIGRATIONS.md)）；不给既有 Token 自动增加授权。
4. handler 若操作任务或业务数据，仍须校验 Token 主人对具体资源的访问范围；目前的授权粒度是接口，非资源实例。
5. 增加实际路由测试，覆盖无凭据、错误凭据、无授权、错误方法以及资源越权。

OpenAPI 使用独立的 `ApiTokenAuth` 安全定义。

## 验证

实际路由测试：`internal/core/server/api_token_test.go`。
服务回归：`internal/modules/system/service/api_token_service_test.go`，包含个人 Token、申请限制与模拟重置/撤权交错。
并发回归：`internal/middleware/api_token_test.go` 和 `internal/core/token`。
浏览器/真实 HTTP：`front-end/e2e/token-notice-security.spec.ts`，使用隔离数据库运行。
