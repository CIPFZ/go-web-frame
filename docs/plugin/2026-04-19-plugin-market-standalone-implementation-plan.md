# Standalone Plugin Market Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在仓库内新增一套可独立部署的插件生态中心前端和后端，同时保留 CMS 中现有集成版插件生态中心不变。

**Architecture:** 新增独立后端服务 `backend/cmd/plugin-market` 与独立领域模块 `plugin_market`，使用 SQLite 保存公开插件数据，并通过同步接口接收 CMS 推送。新增独立前端应用 `plugin-market-front-end`，只消费独立市场服务的公开接口，提供列表和详情浏览下载能力。

**Tech Stack:** Go, Gin, Gorm, SQLite3, React, Umi, Ant Design, Jest, Docker Compose

---

### Task 1: 独立市场后端服务骨架

**Files:**
- Create: `backend/cmd/plugin-market/main.go`
- Create: `backend/configs/plugin-market.yaml`
- Create: `backend/internal/modules/plugin_market/api/api.go`
- Create: `backend/internal/modules/plugin_market/dto/dto.go`
- Create: `backend/internal/modules/plugin_market/model/model.go`
- Create: `backend/internal/modules/plugin_market/repository/repository.go`
- Create: `backend/internal/modules/plugin_market/router/router.go`
- Create: `backend/internal/modules/plugin_market/service/service.go`
- Test: `backend/internal/modules/plugin_market/service/service_test.go`

- [ ] 先写服务层测试，覆盖 SQLite 初始化、插件 upsert、版本 upsert、版本下架、插件列表、插件详情。
- [ ] 跑 `go test ./internal/modules/plugin_market/...`，确认新测试先失败。
- [ ] 实现独立模块最小闭环：模型、仓储、服务、公开查询 DTO、同步 DTO。
- [ ] 补独立启动入口与配置加载，确保服务能直接启动，不依赖 CMS 主服务路由注册。
- [ ] 再跑 `go test ./internal/modules/plugin_market/...`，确认通过。

### Task 2: 独立市场后端 HTTP 接口与鉴权

**Files:**
- Modify: `backend/cmd/plugin-market/main.go`
- Modify: `backend/configs/plugin-market.yaml`
- Modify: `backend/internal/modules/plugin_market/api/api.go`
- Modify: `backend/internal/modules/plugin_market/router/router.go`
- Modify: `backend/internal/modules/plugin_market/service/service.go`
- Test: `backend/internal/modules/plugin_market/router/router_test.go`

- [ ] 先写路由/API 测试，覆盖公开查询接口和同步接口的 token 鉴权。
- [ ] 跑 `go test ./internal/modules/plugin_market/...`，确认接口测试先失败。
- [ ] 实现 `/api/v1/market/plugins`、`/api/v1/market/plugins/:id`、同步接口组以及 `X-Market-Sync-Token` 校验。
- [ ] 再跑 `go test ./internal/modules/plugin_market/...`，确认通过。

### Task 3: 独立前端应用骨架

**Files:**
- Create: `plugin-market-front-end/package.json`
- Create: `plugin-market-front-end/tsconfig.json`
- Create: `plugin-market-front-end/.gitignore`
- Create: `plugin-market-front-end/config/config.ts`
- Create: `plugin-market-front-end/src/app.tsx`
- Create: `plugin-market-front-end/src/locales/zh-CN.ts`
- Create: `plugin-market-front-end/src/locales/en-US.ts`
- Create: `plugin-market-front-end/src/services/market.ts`
- Test: `plugin-market-front-end/src/pages/plugin-list/index.test.tsx`
- Test: `plugin-market-front-end/src/pages/plugin-detail/index.test.tsx`

- [ ] 先写列表页和详情页的测试文件，定义最小渲染与接口请求契约。
- [ ] 跑 `npm test -- --runInBand`，确认独立前端测试先失败。
- [ ] 搭建独立 Umi 应用骨架，配置最小路由、国际化、请求层。
- [ ] 再跑测试和 `npm run tsc`，确认骨架稳定。

### Task 4: 独立前端页面迁移

**Files:**
- Create: `plugin-market-front-end/src/pages/plugin-list/index.tsx`
- Create: `plugin-market-front-end/src/pages/plugin-detail/index.tsx`
- Modify: `plugin-market-front-end/src/services/market.ts`
- Modify: `plugin-market-front-end/src/locales/zh-CN.ts`
- Modify: `plugin-market-front-end/src/locales/en-US.ts`
- Test: `plugin-market-front-end/src/pages/plugin-list/index.test.tsx`
- Test: `plugin-market-front-end/src/pages/plugin-detail/index.test.tsx`

- [ ] 先补细化测试：列表筛选、详情版本切换、中英文切换、下载入口、空态。
- [ ] 跑测试，确认页面行为测试先失败。
- [ ] 迁移并裁剪当前生态中心列表页和详情页，只保留独立市场需要的 UI 和逻辑。
- [ ] 再跑 `npm test -- --runInBand` 和 `npm run tsc`，确认通过。

### Task 5: Docker 与端到端联调

**Files:**
- Create: `backend/Dockerfile.plugin-market`
- Create: `plugin-market-front-end/Dockerfile`
- Create: `plugin-market-front-end/nginx.conf`
- Create: `docker-compose.plugin-market.yml`
- Create: `docs/plugin/2026-04-19-plugin-market-standalone-test-report.md`

- [ ] 为独立后端和独立前端分别补 Docker 构建文件。
- [ ] 新增独立 compose 编排，挂载 SQLite 数据目录。
- [ ] 启动独立市场前后端。
- [ ] 用脚本或 HTTP 请求模拟 CMS 推送插件和版本数据。
- [ ] 用浏览器自动化验证独立前端列表、详情、下载入口和下架后的行为。
- [ ] 将最终验证结果写入测试报告文档。
