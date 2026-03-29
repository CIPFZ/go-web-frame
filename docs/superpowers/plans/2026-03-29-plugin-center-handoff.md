# 插件发布 - 插件中心当前接力文档

日期：2026-03-29
分支：`codex/plugin-release-platform`
仓库：`C:\Users\ytq\work\ai\web-cms`

## 1. 文档用途

这份文档用于记录插件发布模块当前的真实进展，而不是仅记录目标方案。

适用场景：

- 对话被 `/compact` 压缩后继续接力
- 更换新的 AI / Agent 继续开发
- 临时中断后快速恢复上下文

建议新的接手机器人先读这三个文件：

1. `docs/superpowers/specs/2026-03-28-plugin-center-design.md`
2. `docs/superpowers/plans/2026-03-28-plugin-center-refactor.md`
3. `docs/superpowers/plans/2026-03-29-plugin-center-handoff.md`

其中：

- `design` 说明产品信息架构和交互原则
- `refactor plan` 说明原始实施分解
- `handoff` 说明当前代码已经做到哪、卡在哪、下一步先做什么

## 2. 当前已确认的信息架构

### 2.1 模块定位

插件发布只是 `web-cms` 中的一个模块，不单独拆系统。

- 全局首页仍然是控制台
- 插件模块沿用 CMS 现有角色、菜单、权限模型
- 是否能看到菜单、是否能访问子页面，完全由 CMS 角色管理控制

### 2.2 管理侧角色

管理侧只考虑 3 类角色：

1. 插件提供方
2. 插件审核者
3. 插件发布者

插件下载使用者属于公开侧，不参与 CMS 管理流程。

### 2.3 菜单结构

插件发布模块最终方向：

- 插件发布
  - 项目管理
  - 审核工作台
  - 发布工作台

说明：

- 子菜单由角色权限决定是否可见
- 不额外做“进入插件中心后再切角色”的逻辑
- 多角色用户可以同时看到多个子菜单

### 2.4 三个核心对象

管理侧只保留 3 个核心对象：

1. 项目 `Project`
2. 版本 `Version`
3. 流程记录 `Workflow / Audit Trail`

原则：

- 项目是长期主体
- 版本是操作主体
- 流程记录负责审计追踪
- 版本当前状态与流程历史同时保留
- 下架既要改变版本状态，也要保留下架流程记录

### 2.5 页面组织方式

采用：

- 列表分开
- 详情统一

即：

- 项目管理：项目列表页
- 审核工作台：待审核版本列表页
- 发布工作台：待发布版本列表页
- 统一项目详情页：所有列表最终都进入这里

### 2.6 项目详情页骨架

统一骨架已经确定：

- 左侧：项目基础信息
- 中间：版本列表
- 右侧顶部：流程区
- 右侧下方：Tabs
  - 版本概览
  - 文件资料
  - 审核与发布
  - 时间轴

不同入口的默认落点：

- 从项目管理进入：默认看版本概览
- 从审核工作台进入：默认打开审核与发布，并定位目标版本
- 从发布工作台进入：默认打开审核与发布，并定位目标版本

### 2.7 项目管理首页规则

项目管理首页只显示项目轻量信息，不展开完整流程。

每个项目只保留最近一条流程摘要，例如：

- `v1.2.0 审核中`
- `v1.3.0 待补资料`
- `v1.1.0 已发布`

首页不要展开多个并行版本流程，避免重新变成“大卡片”。

### 2.8 创建新版本策略

已确认采用：

- 先创建版本壳
- 再在项目详情中补齐资料

即：

1. 先创建最小版本记录
2. 新版本进入“筹备中”
3. 在详情页继续补变更说明、测试报告、x86/ARM 包
4. 资料齐全后再提交资料进入审核

## 3. 当前代码状态

### 3.1 已存在的关键文档

- `docs/superpowers/specs/2026-03-28-plugin-center-design.md`
- `docs/superpowers/plans/2026-03-28-plugin-center-refactor.md`

### 3.2 近期已落地的后端改动

后端已经做了“项目归属”和“请求方范围过滤”的基础修正：

- 给插件项目补了稳定归属字段 `CreatedBy`
- 创建项目时保存当前用户 ID
- 请求方项目列表改成服务端过滤，不再只是前端隐藏
- 项目编辑权限按项目归属控制

相关文件：

- `backend/internal/modules/plugin/model/model.go`
- `backend/internal/modules/plugin/dto/plugin.go`
- `backend/internal/modules/plugin/repository/repository.go`
- `backend/internal/modules/plugin/service/service.go`
- `backend/internal/modules/plugin/api/api.go`
- `backend/cmd/server/seed_admin.go`

已验证：

- `cd backend && go build ./...` 通过

### 3.3 近期已落地的前端改动

#### 项目管理页

当前主页面：

- `front-end/src/pages/plugin/project-center/index.tsx`

已完成：

- 不再依赖全量 release list 拼首页摘要
- 首页直接吃项目摘要字段
- 编辑按钮改为按项目归属显示

兼容页：

- `front-end/src/pages/plugin/center/index.tsx`

说明：

- 该页应继续作为兼容出口，避免旧菜单/旧路径失效

#### 动态菜单组件映射

已在：

- `front-end/src/utils/componentMap.tsx`

补上以下页面组件映射：

- `plugin/project-center`
- `plugin/project`
- `plugin/review-workbench`
- `plugin/publish-workbench`

这一步是必须的，否则动态菜单命中路由后会出现空白页。

#### 路由与菜单种子

后端菜单种子已补隐藏详情路由：

- `/plugin/project/:id`

用于支持从项目列表进入项目详情页。

### 3.4 运行态已验证内容

已确认：

- Docker 已可用
- `http://localhost:8080/health` 返回 200
- 请求方账号 `plugin.requester / Plugin@123456`
  可以正常获取自己的项目列表
- 项目列表现在返回 6 个自己的项目
- 之前的 SQL `created_by` 歧义错误已修复

## 4. 当前未完成的核心实施项

### 4.1 任务 1：路由/菜单骨架

基本已落地。

### 4.2 任务 2：项目管理轻量首页

已推进到可用状态，但还需要继续压缩信息密度和样式。

用户明确反馈：

- 现在项目卡片仍然太大
- 希望更像“小卡片网格”
- 不要每个项目展示太多信息

视觉方向：

- 更紧凑
- 更轻量
- 类似 `skillshub` 那种小卡片总览，但仍要符合后台风格

### 4.3 任务 3：审核工作台

尚未正式开始重构。

目标已明确：

- 高密度列表
- 默认显示“待审核 + 我审核过的”
- 点击一行进入统一项目详情页，并定位到对应版本

### 4.4 任务 4：发布工作台

尚未正式开始重构。

目标已明确：

- 高密度列表
- 默认显示“待发布 + 我发布过的”
- 点击一行进入统一项目详情页，并定位到对应版本

### 4.5 任务 5：统一项目详情页

页面骨架已经确定，但当前实现还不稳定，仍需继续重构和修 bug。

## 5. 当前正在处理的线上问题

### 5.1 问题：请求方进入项目详情时“权限不足”

现象：

- `plugin.requester` 可以看到项目列表
- 点击项目详情页后提示权限不足

已定位根因：

- 不是前端路由本身的问题
- 是 API `POST /api/v1/plugin/plugin/getProjectDetail` 没有授权给请求方角色

已确认现象：

- 使用 requester token 调用 `getProjectDetail`
- 返回 `code: 1004`
- 返回信息为“权限不足”

根因细节：

- 权限角色 `10010` 已有很多 plugin API
- 但缺少 `getProjectDetail`

### 5.2 当前修复状态

已在：

- `backend/cmd/server/seed_admin.go`

补入 `getProjectDetail` 到插件相关 API policy 列表：

- 全量 plugin API 列表
- `pluginOnlyAccess()`
- `pluginOnlyPolicyWithBasic()`

注意：

这一步代码已改，但在写这份 handoff 时，需要再次确认：

1. 后端容器是否已用最新镜像重建
2. 权限种子是否已经真正写回数据库
3. requester 重新登录后是否拿到新权限

也就是说：

- 代码层面的修复已做
- 运行态最终闭环仍需要再验证一次

## 6. 当前最优先的下一步

下一位接手的 Agent 建议严格按下面顺序执行：

### Step 1：先闭环当前权限问题

1. 重建后端镜像
2. 重启后端容器
3. requester 重新登录
4. 验证 `getProjectDetail` 是否返回 `code: 0`
5. 验证 `/#/plugin/project/:id` 是否能正常展示详情页

如果仍失败，优先检查：

- `sys_authority_apis` 是否真的写入了 `getProjectDetail`
- 菜单/API 权限缓存是否需要重新登录刷新

### Step 2：稳定项目详情页

在权限问题解决后，继续检查：

- 项目详情页是否能正常展示项目基础信息
- 版本列表是否正常显示
- 路由 query 是否能正确定位目标版本
- 当前页面是否仍有空白区、乱码或未注册组件

### Step 3：继续任务 3 审核工作台

当前任务优先级建议：

1. 修好项目详情访问权限
2. 稳定统一项目详情页
3. 再进入审核工作台

不要直接跳去做审核工作台，否则详情页会继续成为阻塞点。

## 7. 建议新的 Agent 如何接手

建议新的 Agent 按下面顺序读代码：

1. `backend/cmd/server/seed_admin.go`
2. `backend/internal/modules/plugin/service/service.go`
3. `backend/internal/modules/plugin/repository/repository.go`
4. `front-end/src/utils/componentMap.tsx`
5. `front-end/src/pages/plugin/project-center/index.tsx`
6. `front-end/src/pages/plugin/project/index.tsx`

然后先验证这条链：

1. requester 登录
2. 打开 `/#/plugin/center`
3. 点击一个项目
4. 验证是否能打开 `/#/plugin/project/:id`
5. 验证 `getProjectDetail` 是否成功

## 8. 演示账号

当前本地可用账号：

- `admin / Admin@123456`
- `plugin.requester / Plugin@123456`
- `plugin.reviewer / Plugin@123456`
- `plugin.publisher / Plugin@123456`

推荐验证顺序：

1. requester 验证项目管理和项目详情
2. reviewer 验证审核工作台
3. publisher 验证发布工作台

## 9. 注意事项

### 9.1 不要被终端中文乱码误导

当前环境里用终端直接 `Get-Content` 查看中文文档时，经常会出现乱码。

这不一定代表源文件本身有问题。

优先依据：

- 浏览器页面表现
- 实际接口响应
- 编辑器文件内容

而不是仅凭终端输出判断中文内容损坏。

### 9.2 继续保持轻量汇报

当前用户已经要求进入 compact 模式。

后续汇报建议只保留：

- 当前修什么
- 改了哪些文件
- 验证结果
- 下一步做什么

### 9.3 不要遗忘的产品原则

后续前端重构时要始终坚持：

- 项目首页只放轻量摘要
- 项目详情才承载完整上下文
- 审核/发布以待办列表为入口
- 所有深度处理最终回到统一项目详情页
- 不为了“好看”牺牲后台效率

## 10. 当前一句话状态

插件中心的信息架构已经定清楚，项目管理首页已初步落地，当前阻塞点是 requester 对 `getProjectDetail` 的 API 权限未完全在运行态闭环，修完这个点后应继续稳定统一项目详情页，再推进审核工作台。
