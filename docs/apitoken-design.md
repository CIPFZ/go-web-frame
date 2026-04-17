# API Token 设计文档

## 目标

为 `web-cms` 增加一套仅供外部服务端或脚本调用后端 API 的 `API Token` 机制。

这套机制必须满足以下约束：

- 不参与浏览器后台登录态
- 不写 Cookie
- 不替代现有 `JWT + Casbin` 的后台用户权限体系
- 由后台管理员在 system 模块中创建和管理
- 授权范围复用现有 `sys_apis`

## 设计原则

- 认证与后台用户体系隔离：`API Token` 只认 `X-API-Token`
- 授权复用现有 API 资源模型：按 `path + method` 精确授权
- 明文 token 永不落库：数据库只保存 `token_hash`
- 默认单实例并发限制：满足当前小规模生产场景
- 管理入口统一放入 `system` 模块

## 范围

本次实现包含：

- `sys_api_tokens` 与 `sys_api_token_apis` 数据模型
- 后台管理接口与管理页面
- token 生成、重置、启停、删除、列表、详情、更新
- `X-API-Token` 认证中间件
- 基于 `sys_apis` 的接口授权
- 单实例内存并发限制
- 至少一组业务路由支持 `API Token` 访问
- 后端与前端测试

本次不包含：

- 浏览器端使用 token 登录后台
- 多实例分布式并发控制
- token 使用明细审计表
- token 自助申请流程

## 核心模型

### 1. sys_api_tokens

字段：

- `id`
- `token_hash`：完整 token 的 SHA256 摘要，唯一索引
- `token_prefix`：仅用于展示，形如 `cms_ab12`
- `name`
- `description`
- `expires_at`
- `max_concurrency`
- `enabled`
- `last_used_at`
- `created_by`
- `created_at`
- `updated_at`
- `deleted_at`

### 2. sys_api_token_apis

用于维护 token 与 `sys_apis` 的多对多关系：

- `api_token_id`
- `api_id`

## 认证与授权设计

### 1. 请求入口

外部调用方通过请求头传递：

`X-API-Token: <raw-token>`

### 2. 中间件职责

新增 `ApiTokenAuth` 中间件，执行顺序如下：

1. 从请求头读取 `X-API-Token`
2. 对原始 token 做 SHA256
3. 根据 `token_hash` 查询 token
4. 校验 `enabled`
5. 校验 `expires_at`
6. 校验当前请求 `path + method` 是否在该 token 的授权 API 集合内
7. 校验并发数
8. 将 `apiTokenId` 等上下文写入 gin context
9. 请求结束后释放并发占用
10. 尝试更新 `last_used_at`

### 3. 授权粒度

授权基于现有 `sys_apis`，按以下维度精确匹配：

- HTTP Method
- Router 注册路径

不做按菜单、按角色、按模块的粗粒度授权。

## 路由设计

### 1. 后台管理接口

继续放在 `system` 模块，由后台管理员通过 JWT 登录后调用：

- `POST /api/v1/sys/api-token/create`
- `POST /api/v1/sys/api-token/getApiTokenList`
- `GET /api/v1/sys/api-token/detail`
- `PUT /api/v1/sys/api-token/update`
- `DELETE /api/v1/sys/api-token/delete`
- `POST /api/v1/sys/api-token/reset`
- `POST /api/v1/sys/api-token/enable`
- `POST /api/v1/sys/api-token/disable`

### 2. 外部 token 可访问业务接口

本次不新开 `/openapi` 前缀，而是在现有业务路由上增加一组独立路由分组：

- 该分组只挂 `ApiTokenAuth`
- 只暴露明确允许外部调用的接口

第一阶段选择诗词模块只读接口作为样板：

- `GET /api/v1/poetry/dynasty/list`
- `GET /api/v1/poetry/dynasty/all`
- `GET /api/v1/poetry/genre/list`
- `GET /api/v1/poetry/genre/all`
- `GET /api/v1/poetry/author/list`
- `GET /api/v1/poetry/author/:id`
- `GET /api/v1/poetry/poem/list`
- `GET /api/v1/poetry/poem/:id`

## 后台管理页面

前端新增 `sys/api-token` 页面，能力如下：

- token 列表
- 新建 token
- 编辑 token 元数据
- 选择授权 API
- 启用/禁用
- 删除
- 重置 token
- 创建或重置成功后一次性展示完整 token

页面风格复用当前 `sys/api` 的 `ProTable + ModalForm` 模式。

## 安全策略

- 数据库存 hash，不存明文 token
- 明文 token 只在创建和重置成功时返回一次
- 失效 token 不可恢复，只能重置
- 禁用或删除后立即失效
- 默认返回统一的 401/403，避免泄露过多内部细节

## 并发限制策略

当前版本采用进程内内存计数：

- 优点：实现简单，满足单实例部署
- 限制：多实例场景下无法保证全局并发上限

因此文档明确约束：第一阶段的小规模生产仅支持单实例下的严格并发控制。

## 测试策略

后端测试覆盖：

- token 生成与 hash
- service 创建、重置、更新逻辑
- repository API 绑定逻辑
- 中间件授权成功与失败路径
- 并发限制

前端测试覆盖：

- 路由配置包含 `sys/api-token`
- 页面关键交互函数
- 成功创建/重置场景的展示逻辑

## 与现有系统的关系

- 管理 `API Token` 的权限，仍由现有后台管理员体系控制
- `API Token` 自身不拥有角色，不参与 Casbin
- `sys_apis` 继续作为唯一接口资源目录

## 结果

实现完成后，项目将同时具备两套互不混淆的访问方式：

- 后台管理员/用户：`JWT + Casbin`
- 外部脚本/服务：`API Token + sys_apis`
