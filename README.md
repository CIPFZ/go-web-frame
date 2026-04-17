# Web CMS

一个前后端分离的内容管理系统项目。

- 后端目录：`backend`
- 前端目录：`front-end`
- 根目录职责：编排开发环境、沉淀协作文档、承载 Docker 与部署配置

## 项目结构

```text
.
├─ backend/                 Go 后端
├─ front-end/               Umi + Ant Design Pro 前端
├─ deploy/                  部署相关资源
├─ docs/                    过程文档与补充资料
├─ scripts/                 辅助脚本
├─ docker-compose.yml       本地联调编排
├─ LOCAL_RUN.md             本地运行说明
├─ ENVIRONMENTS.md          环境说明
└─ TEMPLATE_QUALITY.md      模板质量基线
```

## 后端模块地图

### 启动入口

- `backend/cmd/server`
  负责应用启动、核心依赖初始化、HTTP 服务启动、优雅停机。
- `backend/cmd/migrate`
  负责数据库迁移。
- `backend/configs`
  后端运行配置目录。当前默认启动配置路径是 `./configs/config.yaml`，以 `backend` 目录为执行基准。

### 后端分层

```text
backend
├─ cmd/                     启动与迁移入口
├─ configs/                 配置文件
├─ internal/
│  ├─ core/                 基础能力
│  ├─ middleware/           Gin 中间件
│  ├─ modules/              业务模块
│  └─ svc/                  ServiceContext，统一依赖装配
├─ pkg/                     可复用公共包
├─ logs/                    日志输出
└─ uploads/                 本地文件上传目录
```

### `internal/core` 基础能力

- `config`
  配置加载与配置结构定义。
- `db`
  数据库初始化与连接管理。
- `jwt`
  JWT 生成、解析、续期、黑名单处理。
- `file`
  文件上传相关基础能力。现在包含服务端文件名清洗与上传策略校验。
- `log`
  日志封装。
- `i18n`
  国际化支持。
- `observability`
  可观测性相关能力。
- `server`
  HTTP 服务基础装配。
- `claims`
  JWT claims 类型定义。
- `audit`
  审计相关能力。

### `internal/middleware` 请求链路

- `jwt.go`
  登录态校验、中间件续签、黑名单检查。
- `casbin.go`
  权限控制。
- `operation_record.go`
  操作日志记录。
- 其他中间件
  一般负责跨切面能力，如鉴权、日志、审计、权限等。

### `internal/modules` 业务模块

当前主要有两个模块：

- `system`
  后台管理核心模块，包含用户、角色、菜单、API 权限、Casbin、操作日志、文件上传、系统信息、通知。
- `poetry`
  诗词业务模块，包含朝代、体裁、作者、作品等 CRUD 能力。

每个模块基本按统一结构组织：

```text
module
├─ api/                     Gin Handler，处理请求入参与响应
├─ dto/                     请求/响应 DTO
├─ model/                   GORM 模型
├─ repository/              数据访问层
├─ router/                  路由注册
└─ service/                 业务逻辑层
```

### 后端主要路由域

基于当前路由定义，可把后端接口理解为以下几组：

- 公共接口
  - `/user/login`
  - `/user/register`
- 系统管理接口前缀：`/sys`
  - `/sys/user/*`
    用户信息、用户管理、头像上传、角色切换、UI 设置
  - `/sys/menu/*`
    动态菜单树、菜单管理、角色菜单回显
  - `/sys/authority/*`
    角色管理、角色菜单分配
  - `/sys/api/*`
    API 元数据管理
  - `/sys/casbin/*`
    角色 API 权限策略
  - `/sys/operationLog/*`
    操作日志查询与删除
  - `/sys/file/upload`
    通用文件上传
  - `/sys/system/getServerInfo`
    系统状态与服务器信息
  - `/sys/notice/*`
    通知列表、已读、创建
- 诗词业务接口前缀：`/poetry`
  - `/poetry/dynasty/*`
  - `/poetry/genre/*`
  - `/poetry/author/*`
  - `/poetry/poem/*`

### 后端当前实现重点

- `svc.ServiceContext`
  是后端装配中心。数据库、JWT、OSS、配置、日志等依赖都从这里下发到模块层。
- `system` 模块承担了后台管理的大部分基础设施能力。
- 动态菜单是前后端联动的关键点：
  后端按角色输出菜单树，前端再把菜单树转换成运行时路由。
- 文件上传链路已经补上两类服务端保护：
  - 清洗客户端传入文件名，阻断路径穿越
  - 执行配置中的大小与扩展名限制

## 前端页面与路由地图

### 前端目录分层

```text
front-end
├─ config/                  Umi 静态配置
├─ src/
│  ├─ app.tsx               运行时入口，含登录态与动态路由注入
│  ├─ components/           通用组件
│  ├─ config/               前端测试等运行期辅助配置
│  ├─ locales/              多语言
│  ├─ pages/                页面目录
│  ├─ services/             API 请求封装
│  └─ utils/                工具函数
├─ e2e/                     Playwright 端到端测试
└─ jest.config.ts           Jest 配置
```

### 静态路由骨架

根路由定义位于 `front-end/config/routes.ts`，当前结构可以概括为：

- `/user`
  登录、注册、注册结果页，独立布局
- `/`
  主系统布局壳。这里只保留静态兜底重定向，真正业务页通过运行时注入
- `/account/settings`
  账户设置页
- `/*`
  全局 404

当前静态兜底首页是：

- `/` -> `/dashboard/workplace`

### 动态路由注入机制

前端真正的后台业务路由，不是全部写死在 `config/routes.ts` 里，而是依赖 `front-end/src/app.tsx` 在运行时完成：

- `getInitialState`
  拉取当前用户信息与菜单数据。
- `patchClientRoutes`
  读取后端菜单树，把菜单转换为前端路由并注入主 Layout。
- `menuDataRender`
  直接使用运行时菜单数据渲染左侧菜单。

这意味着：

- 后端菜单树结构错误，会直接影响前端菜单显示与页面可达性。
- 角色默认首页 `defaultRouter` 会覆盖静态兜底首页。
- 当菜单或用户态拉取失败时，静态兜底首页必须指向真实存在页面，否则用户会掉进 404。

### 页面地图

按 `front-end/src/pages` 可以把当前页面分成以下几组：

- `user`
  - `login`
  - `register`
  - `register-result`
- `dashboard`
  - `workplace`
- `account`
  - `settings`
- `sys`
  - `user`
  - `menu`
  - `authority`
  - `api`
  - `notice`
  - `operation`
- `poetry`
  - `dynasty`
  - `genre`
  - `author`
  - `poem`
- `exception`
  - `403`
  - `404`
  - `500`
- 其他示例或扩展页
  - `about`
  - `plugin/project-center`
  - `profile/basic`
  - `profile/advanced`
  - `result/success`
  - `result/fail`
  - `state`

### 前端服务层地图

`front-end/src/services/api` 基本按业务域拆分：

- `user.ts`
  登录、当前用户、设置、头像、用户管理
- `menu.ts`
  菜单数据与菜单管理
- `authority.ts`
  角色管理
- `api.ts`
  API 管理
- `casbin.ts`
  权限策略
- `operationLog.ts`
  操作日志
- `notice.ts`
  通知
- `state.ts`
  系统状态
- `poetry.ts`
  诗词业务相关接口

## 前后端联动关系

### 登录态

- 后端通过 JWT + Cookie/Header 维护登录态
- 前端通过 `localStorage token`、请求头与运行时初始化联合判断是否已登录
- 中间件会在接近过期时续签 token，并更新 Cookie/Header

### 权限与菜单

- 后端 `system/menu` 与 `system/authority` 共同决定用户能看到的菜单和可访问资源
- 前端通过菜单树派生页面路由，而不是单独维护一份完整业务路由表

### 上传

- 前端发起上传请求
- 后端统一在服务端做扩展名、大小、文件名校验
- OSS 实现负责真正落盘或对象存储写入

## 数据库驱动

当前后端支持三种数据库驱动：

- `mysql`：使用 `database.mysql`
- `postgres`：使用 `database.postgres`
- `sqlite3`：使用 `database.sqlite`

`sqlite3` 的定位是本地开发优先，同时支持单实例、小规模生产。推荐配置：

- `database.sqlite.wal=true`
- `database.sqlite.foreign_keys=true`
- `database.sqlite.busy_timeout_ms=5000`

支持通过环境变量 `SQLITE_PATH` 覆盖配置文件中的 sqlite 数据库文件路径。

不建议将 `sqlite3` 用于：

- 多副本部署
- 共享网络存储数据库文件
- 高频并发写入场景

## 开发入口

### 后端

在 `backend` 目录执行：

```bash
go run ./cmd/server
go run ./cmd/migrate
go test ./...
```

如果要使用 `sqlite3`，可在 `backend/configs/config.yaml` 中设置：

```yaml
database:
  driver: sqlite3
  sqlite:
    path: data/app.db
    wal: true
    busy_timeout_ms: 5000
    foreign_keys: true
```

也可以通过环境变量覆盖 sqlite 数据库文件路径：

```bash
SQLITE_PATH=./data/app.db
```

### 前端

在 `front-end` 目录执行：

```bash
npm install
npm start
npm test
```

## 最近一次梳理结论

本轮修复已经覆盖以下关键风险：

- 后端默认配置路径错误，导致本地启动与迁移失败
- Redis 关闭时 JWT 黑名单逻辑会触发 panic
- 菜单树构建会丢失孙级及更深层节点
- 上传接口存在文件名路径穿越风险
- 服务端声明了上传策略但未真正执行
- 登录 Cookie 的 `SameSite` 设置顺序错误
- 前端静态首页兜底指向不存在页面

并且已经补齐对应的后端与前端测试覆盖。
