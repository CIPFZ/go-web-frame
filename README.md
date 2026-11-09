# Go Web Frame 基础框架

提供账号、权限、菜单与后台管理能力，作为不同业务模块分支的公共基础。
Nexus 跑批、阶段 Agent 和多轮进展等功能可在独立业务分支上开发。

## 保留的能力

- 用户登录、个人设置、用户与角色管理。
- 后端菜单下发、角色菜单授权、前端动态路由。
- API 目录、Casbin 权限、API Token 管理。
- 工作台、通知、操作日志、系统状态、文件上传（本地、MinIO、腾讯云 COS；COS 还能签发只能写入单个文件的临时凭据，供客户端直传大文件）。

## 目录与技术栈

| 目录 | 用途 |
| --- | --- |
| `backend/cmd/server` | Go / Gin 进程入口（组装在 internal/bootstrap） |
| `backend/cmd/migrate` | 独立、带锁的版本迁移（含空库初始化） |
| `backend/internal/modules/system` | CMS 接口、服务、仓储与模型 |
| `backend/internal/core` | 配置、数据库、缓存、文件与可观测性 |
| `front-end/src/pages` | React / Umi / Ant Design Pro 页面 |
| `front-end/src/services/api` | 后端请求 |
| `front-end/src/utils/componentMap.tsx` | 后端组件标识到页面组件的映射 |
| `deploy/local` | 开发机 Docker Compose 配置、初始化与验证 |

本机部署使用 MySQL、Redis 和文件持久卷。参见 [Docker 部署说明](deploy/local/README.md)。

## 角色与权限

- 一个用户可以有多个角色，**所有角色同时生效**：接口权限与菜单取并集，不需要切换角色。
  主角色（`sys_users.authority_id`）只决定默认首页。
- 超级管理员（角色 1）不经 Casbin 判断，拥有全部接口。
- 子角色是父角色权限的一部分：分配给子角色的菜单、接口只能从父角色已有的里面选；
  父角色收回的权限，各级子角色一起收回。子角色**不继承**父角色的权限。
- 发给某个角色的通知，也会发给它的各级子角色。

## 动态路由约定

`GET /api/v1/sys/menu/getMenu` 返回当前用户全部角色可访问的菜单树（多个角色取并集），包含 `path`、`component`、
`routes`、`hideInMenu` 等字段。前端通过 `fetchMenuData` 获取数据，`buildRoutes` 生成页面路由，
`patchClientRoutes` 注入布局，`menuDataRender` 渲染导航。
有子菜单的分组本身没有页面：直接访问分组地址时跳到它的第一个子页面，面包屑里分组只显示文字。
需要在同一分组的页面间来回切换、又不想每次重新请求和重画图表时，把分组 path 登记到
`front-end/src/routing/keepAlive.ts`：去过的页面切走时不卸载，页面可用 `useRouteActive()` 只在显示时轮询。

`front-end/config/routes.ts` 仅保留认证页、根布局与 404。API Token、个人设置等业务页面也由
后端菜单注册，隐藏菜单不代表没有路由。新增业务时需同步添加组件映射、后端菜单、API 和角色授权。

## 运行与验证

在仓库根目录执行：

```bash
python3 deploy/local/init.py
DOCKER_BUILDKIT=0 docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml up -d --build --wait --wait-timeout 180
python3 deploy/local/smoke.py
```

访问 `http://localhost:8080`；管理员初始凭据见本机私有文件 `deploy/local/runtime/admin.txt`。
既有账号密码和持久卷在更新部署时保留。后端镜像构建会运行 Go 测试，前端构建使用锁定依赖。
前端单元测试在 `front-end` 下执行 `npm test -- --runInBand`。完整浏览器验证从仓库根目录执行
`python3 deploy/local/e2e.py`：复用已构建镜像，在随机 Compose 项目中创建独立数据库与随机本机端口，
测试结束后删除该测试项目及其持久卷，避免测试用户和通知进入共享 CMS。需要本机安装前端依赖和 Playwright Chromium。

API Token 跟人绑定：每个人申请自己的 Token，能做的事不超过本人现在的权限，管理员可以设置申请限制。
当前系统接口使用 JWT / Casbin，新增外部接口需要显式接入 `ApiTokenAuth` 并用 `token.Expose` 声明，
仅勾选 API 授权不会改变接口的认证方式。详见 [API Token 使用和接入](docs/apitoken-design.md)。

## 基础分支与业务分支

`base-frame` 保存可复用的 CMS 基础能力、通用修复与部署脚本。
各业务从基础分支创建独立分支，例如：

```bash
git switch base-frame
git pull --ff-only
git switch -c nexus-platform
```

业务对象与页面放在各自模块中，基础能力的修复优先回到 `base-frame`，再同步到业务分支。
全新数据库只初始化一个管理员；普通用户由团队按需创建。菜单 `name` 保存可编辑的展示名称，
`locale` 单独保存国际化 Key，重启会保留已经修改的展示名称。

## 架构与维护

- [架构评审与技术栈升级记录](docs/BASE_FRAME_REVIEW.md)
- [目录结构与模块边界](docs/DIRECTORY_STRUCTURE.md)
- [数据库迁移写法](docs/MIGRATIONS.md)
- [工作台与项目扩展](docs/WORKPLACE.md)
- [资源接口契约（core/apispec）](docs/API_SPEC.md)
- [配置约定](docs/CONFIG.md)
- [外部登录（企业单点登录、OAuth 等）](docs/EXTERNAL_LOGIN.md)
- [质量检查与独立测试环境](TEMPLATE_QUALITY.md)

后端使用 Go 1.27.1，前端构建使用 Node 24.21.0 LTS。业务菜单和路由仍由后端下发。

基础框架第二轮问题闭环、数据恢复与观测说明见 [闭环记录](docs/REMEDIATION.md)。系统状态页展示服务可用性；历史指标、日志与调用链在可选 [观测平台](deploy/observability/README.md) 查看。
