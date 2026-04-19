# Plugin 模块文档

本目录用于沉淀 `plugin` 相关设计、实现约束与后续演进文档。

## 文档列表

| 文档 | 说明 |
| --- | --- |
| [2026-04-18-plugin-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-design.md) | 插件 CMS 模块的一期正式设计方案 |
| [2026-04-18-plugin-implementation-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-implementation-plan.md) | 插件 CMS 模块一期实现计划 |
| [2026-04-18-plugin-market-refactor-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-market-refactor-design.md) | 集成版插件生态中心重构设计 |
| [2026-04-18-plugin-real-env-test-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-real-env-test-plan.md) | 插件模块真实环境测试方案 |
| [2026-04-19-plugin-workflow-hardening-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-workflow-hardening-plan.md) | 插件工作流加固计划 |
| [2026-04-19-plugin-market-standalone-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-market-standalone-design.md) | 独立插件生态中心拆分设计方案 |

## 当前文档分层

当前 `plugin` 相关文档分成两条主线：

1. 插件 CMS 主线
2. 插件生态中心主线

其中：

- 插件 CMS 主线关注项目管理、发布单、审核流、工单池和主数据管理。
- 插件生态中心主线分为两代：
  - 集成版生态中心：仍内嵌于 CMS 前后端内部。
  - 独立版生态中心：作为独立前端和独立后端服务存在。

## 当前定位

当前仓库中的 `plugin` 能力已经不是单一 CRUD 页面，而是包含两套相关但边界不同的能力：

1. CMS 内部插件治理能力
2. 面向用户的公开插件生态中心能力

后续新增工作应优先明确自己属于哪一条主线，避免把 CMS 内部逻辑与公开市场逻辑继续耦合。
