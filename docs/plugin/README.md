# Plugin 模块文档

本目录用于沉淀 `plugin` 相关设计、实现计划、重构方案与测试报告。

## 文档列表

| 文档 | 说明 |
| --- | --- |
| [2026-04-18-plugin-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-design.md) | 插件 CMS 一期设计方案 |
| [2026-04-18-plugin-implementation-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-implementation-plan.md) | 插件 CMS 一期实施计划 |
| [2026-04-18-plugin-market-refactor-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-market-refactor-design.md) | 集成版插件生态中心重构方案 |
| [2026-04-18-plugin-real-env-test-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-18-plugin-real-env-test-plan.md) | 插件模块真实环境测试方案 |
| [2026-04-19-plugin-workflow-hardening-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-workflow-hardening-plan.md) | 插件工作流加固计划 |
| [2026-04-19-plugin-market-standalone-design.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-market-standalone-design.md) | 独立插件生态中心拆分设计方案 |
| [2026-04-19-plugin-market-standalone-implementation-plan.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-market-standalone-implementation-plan.md) | 独立插件生态中心实施计划 |
| [2026-04-19-plugin-market-standalone-test-report.md](/Users/ytq/work/ai/web-cms/docs/plugin/2026-04-19-plugin-market-standalone-test-report.md) | 独立插件生态中心 Docker 与端到端测试报告 |

## 当前分层

当前 `plugin` 相关能力分为两条主线：

1. 插件 CMS 主线
2. 插件生态中心主线

其中：

- 插件 CMS 主线负责项目管理、发布单、审核流、工单池与后台治理。
- 插件生态中心主线负责公开浏览、版本查询、下载入口与对外展示。

## 独立生态中心定位

`2026-04-19` 之后，插件生态中心同时存在两种形态：

1. 集成版：继续保留在 CMS 仓库主应用中，不删除。
2. 独立版：新增独立前端 `plugin-market-front-end` 与独立后端 `backend/cmd/plugin-market`。

独立版的目标是：

- 轻量部署
- 使用 SQLite3
- 仅保留公开插件市场最小能力
- 通过 CMS 同步接口接收插件与版本数据

后续新增需求时，需先明确变更属于 CMS 治理侧，还是独立公开市场侧，避免边界继续混杂。
