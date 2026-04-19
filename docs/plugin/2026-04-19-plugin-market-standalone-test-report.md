# 独立插件生态中心测试报告

日期：2026-04-19

## 测试目标

验证独立拆分后的插件生态中心具备以下能力：

1. 独立后端可基于 SQLite3 启动并提供公开查询接口。
2. CMS 侧可通过同步接口推送插件与版本数据。
3. 独立前端可展示插件列表、详情、版本切换与下载入口。
4. 下线版本、删除插件后，公开市场侧能正确反映变更。
5. Docker 部署形态可与现有 CMS 并行运行，不互相覆盖。

## 测试环境

- 工作分支：`feat/plugin-market-standalone`
- 后端：`backend/cmd/plugin-market`
- 前端：`plugin-market-front-end`
- 编排文件：`docker-compose.plugin-market.yml`
- 独立后端地址：`http://127.0.0.1:18081`
- 独立前端地址：`http://127.0.0.1:18082`
- 同步鉴权头：`X-Market-Sync-Token: plugin-market-sync-token`
- 数据库：SQLite3

## 静态验证

### 后端

- `go test ./internal/modules/plugin_market/...`
- `go test ./cmd/plugin-market`

结果：全部通过。

### 前端

- `npm test`
- `npm run tsc`
- `npm run build`

结果：全部通过。

## Docker 部署验证

执行命令：

```powershell
docker compose -f docker-compose.plugin-market.yml up -d --build
docker compose -f docker-compose.plugin-market.yml ps
```

验证结果：

- `plugin-market-backend` 正常启动，暴露 `18081`
- `plugin-market-front-end` 正常启动，暴露 `18082`
- 现有 CMS 容器仍保持运行，未被替换

健康检查：

```http
GET /healthz
```

返回：

```json
{"name":"plugin-market","status":"ok"}
```

## 真实链路测试

### 用例 1：同步插件基础信息

操作：

- 调用 `/api/v1/market/sync/plugin/upsert`
- 推送 `pluginId=1001`，编码 `agent-helper`

结果：

- 返回 `code=0`
- 后端接受并创建插件主记录

### 用例 2：同步两个已发布版本

操作：

- 调用 `/api/v1/market/sync/version/upsert`
- 推送 `releaseId=5001`, `version=1.0.0`
- 推送 `releaseId=5002`, `version=1.1.0`

结果：

- 返回均为 `code=0`
- 列表接口显示最新版本为 `1.1.0`
- 详情接口返回两个版本，按发布时间倒序

### 用例 3：公开列表接口验证

操作：

```http
POST /api/v1/market/plugins/list
```

断言：

- 能返回 `agent-helper`
- `latestVersion=1.1.0`
- 返回下载地址与兼容项

结果：通过。

### 用例 4：公开详情接口验证

操作：

```http
POST /api/v1/market/plugins/detail
```

断言：

- 能返回插件基础信息
- `release` 指向最新发布版本 `1.1.0`
- `versions` 包含 `1.1.0` 与 `1.0.0`

结果：通过。

### 用例 5：前端列表页验证

验证方式：

- 使用 Playwright 脚本打开 `http://127.0.0.1:18082/plugins`
- 切换 EN

断言：

- 页面标题为 `Plugin Market`
- 卡片标题为 `Agent Helper`
- 卡片展示 `Latest Version: v1.1.0`

结果：通过。

### 用例 6：前端详情页验证

验证方式：

- 从列表页进入详情页
- 切换 EN

断言：

- 详情页标题为 `Agent Helper`
- 版本选择器默认指向 `v1.1.0`
- 版本历史表格可见
- x86、ARM、测试报告下载入口可见

结果：通过。

### 用例 7：下线最新版本

操作：

- 调用 `/api/v1/market/sync/version/offline`
- 下线 `releaseId=5002`

断言：

- 列表接口最新版本回退到 `1.0.0`
- 详情接口 `release` 回退到 `1.0.0`
- `versions` 仅保留仍处于 published 的版本
- 前端列表页显示 `Latest Version: v1.0.0`

结果：通过。

说明：

- 初次验证时我将“下线请求”和“查询请求”并行发出，导致读到了旧结果。
- 复查后确认不是代码缺陷，而是测试步骤并行造成的读写时序误判。
- 按串行流程重新验证后，结果稳定正确。

### 用例 8：删除插件

操作：

- 调用 `/api/v1/market/sync/plugin/delete`

断言：

- 列表接口返回空数组
- 前端列表页显示 `No published plugins`

结果：通过。

说明：

- 初次验证同样因为并行执行了删除与查询，短暂读到旧结果。
- 按串行顺序重跑后，删除生效且结果稳定。

## 结论

独立插件生态中心当前已经满足本轮设计目标：

1. 已形成独立前端应用与独立后端服务。
2. 后端仅保留公开市场所需的最小数据模型与同步接口。
3. SQLite3 部署可行，适合轻量独立部署。
4. 前端可独立发布，不依赖 CMS 菜单与权限体系。
5. Docker 形态已验证可与现有 CMS 共存。

## 后续建议

1. 后续由 CMS 增加正式“推送到独立市场”的同步任务，而不是手工调用接口。
2. 若要进一步对外发布，建议给同步接口补充请求签名与来源白名单。
3. 若需要提升部署效率，可继续为独立前端与后端补充更细粒度的健康检查与 CI 流程。
