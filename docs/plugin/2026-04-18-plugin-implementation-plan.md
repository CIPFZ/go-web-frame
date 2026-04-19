# Plugin 模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前仓库中完整落地 `plugin` 模块的一期能力，包含后端业务模块、前端管理与公开页面、动态路由接入、数据库迁移与测试覆盖。

**Architecture:** 后端新增独立 `plugin` 模块，沿用现有 `api/dto/model/repository/service/router` 分层；前端新增 `plugin` 页面和服务层，接入现有动态菜单与静态公开路由；围绕“发布单 + 工单池认领”的状态机实现，并以 SQLite 测试库和 Jest 组件测试保障回归。

**Tech Stack:** Go, Gin, Gorm, SQLite, React, Umi Max, Ant Design Pro, Jest, Testing Library

---

### Task 1: 搭建 plugin 后端骨架与数据库模型

**Files:**
- Create: `backend/internal/modules/plugin/model/*.go`
- Create: `backend/internal/modules/plugin/dto/*.go`
- Create: `backend/internal/modules/plugin/repository/*.go`
- Create: `backend/internal/modules/plugin/service/*.go`
- Create: `backend/internal/modules/plugin/api/*.go`
- Create: `backend/internal/modules/plugin/router/router.go`
- Modify: `backend/internal/core/server/router.go`
- Modify: `backend/cmd/migrate/main.go`
- Test: `backend/internal/modules/plugin/service/*_test.go`

- [ ] Step 1: 先为状态机、模型字段和核心 service 行为写失败测试。
- [ ] Step 2: 运行对应 `go test`，确认因模块缺失或行为未实现而失败。
- [ ] Step 3: 实现 `plugins`、`plugin_releases`、`plugin_compatible_products`、`plugin_release_events`、`plugin_products`、`plugin_departments` 模型及仓储。
- [ ] Step 4: 将 plugin 模块接入主路由和 `AutoMigrate`。
- [ ] Step 5: 再次运行后端测试，确认骨架层测试通过。

### Task 2: 实现项目、发布单、工单池与公开查询后端能力

**Files:**
- Modify: `backend/internal/modules/plugin/service/*.go`
- Modify: `backend/internal/modules/plugin/repository/*.go`
- Modify: `backend/internal/modules/plugin/api/*.go`
- Modify: `backend/internal/modules/plugin/dto/*.go`
- Test: `backend/internal/modules/plugin/service/*_test.go`
- Test: `backend/internal/modules/plugin/router/router_test.go`

- [ ] Step 1: 为项目创建/更新、发布单创建/更新、提交审核、认领、打回、重提、重置、发布、下架、公开查询写失败测试。
- [ ] Step 2: 运行目标 `go test`，确认状态流转和权限校验测试先失败。
- [ ] Step 3: 实现 service 事务逻辑、并发认领保护、事件记录与聚合查询。
- [ ] Step 4: 暴露对应 API 与路由，并补充路由注册测试。
- [ ] Step 5: 运行模块测试，确认 service/router 测试通过。

### Task 3: 实现前端服务层、页面路由与动态组件映射

**Files:**
- Create: `front-end/src/services/api/plugin.ts`
- Create: `front-end/src/pages/plugin/project-management/index.tsx`
- Create: `front-end/src/pages/plugin/project-detail/index.tsx`
- Create: `front-end/src/pages/plugin/work-order-pool/index.tsx`
- Create: `front-end/src/pages/plugin/public-list/index.tsx`
- Create: `front-end/src/pages/plugin/public-detail/index.tsx`
- Create: `front-end/src/pages/plugin/components/*`
- Modify: `front-end/src/utils/componentMap.tsx`
- Modify: `front-end/config/routes.ts`
- Test: `front-end/src/utils/componentMap.test.tsx`
- Test: `front-end/src/pages/plugin/**/*.test.tsx`

- [ ] Step 1: 先写组件映射、页面基本渲染、工单池操作显隐和公开页渲染的失败测试。
- [ ] Step 2: 运行 `npm test -- plugin` 和相关 jest 用例，确认先失败。
- [ ] Step 3: 实现服务层请求函数、后台页面和公开页面。
- [ ] Step 4: 接入静态路由与动态组件映射，避免出现未找到组件映射。
- [ ] Step 5: 重新运行前端测试，确认页面与映射测试通过。

### Task 4: 前后端联调与测试补齐

**Files:**
- Modify: `backend/internal/modules/plugin/service/*_test.go`
- Modify: `front-end/src/pages/plugin/**/*.test.tsx`
- Modify: `docs/plugin/README.md`

- [ ] Step 1: 为完整业务链路补充测试场景，至少覆盖提交审核、认领、打回重提、发布、下架、公开可见性。
- [ ] Step 2: 跑完整后端 `go test ./...` 与前端 `npm test -- --runInBand`。
- [ ] Step 3: 跑前端 `npm run tsc`，确认类型通过。
- [ ] Step 4: 若存在新增文档或页面入口差异，补充 `docs/plugin/README.md`。

### Task 5: 最终验证与交付整理

**Files:**
- Modify: 仅在必要时修复最终验证中发现的问题

- [ ] Step 1: 执行后端格式化与前端测试最终验证。
- [ ] Step 2: 检查功能覆盖是否与设计文档一致。
- [ ] Step 3: 汇总变更、测试结果和剩余风险后再进入 commit/push。
