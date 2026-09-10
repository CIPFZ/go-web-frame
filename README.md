# Go Web Frame 基础框架

提供账号、权限、菜单与后台管理能力，作为不同业务模块分支的公共基础。
Nexus 跑批、阶段 Agent 和多轮进展等功能可在独立业务分支上开发。

## 保留的能力

- 用户登录、个人设置、用户与角色管理。
- 后端菜单下发、角色菜单授权、前端动态路由。
- API 目录、Casbin 权限、API Token 管理。
- 工作台、通知、操作日志、系统状态、文件上传。

## 目录与技术栈

| 目录 | 用途 |
| --- | --- |
| `backend/cmd/server` | Go / Gin 服务与基础数据初始化 |
| `backend/cmd/migrate` | GORM 表结构初始化与版本化数据迁移 |
| `backend/internal/modules/system` | CMS 接口、服务、仓储与模型 |
| `backend/internal/core` | 配置、数据库、缓存、文件与可观测性 |
| `front-end/src/pages` | React / Umi / Ant Design Pro 页面 |
| `front-end/src/services/api` | 后端请求 |
| `front-end/src/utils/componentMap.tsx` | 后端组件标识到页面组件的映射 |
| `deploy/local` | 开发机 Docker Compose 配置、初始化与验证 |

本机部署使用 MySQL、Redis 和文件持久卷。参见 [Docker 部署说明](deploy/local/README.md)。

## 动态路由约定

`GET /api/v1/sys/menu/getMenu` 返回当前角色可访问的菜单树，包含 `path`、`component`、
`routes`、`hideInMenu` 等字段。前端通过 `fetchMenuData` 获取数据，`buildRoutes` 生成页面路由，
`patchClientRoutes` 注入布局，`menuDataRender` 渲染导航。

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

API Token 管理保留为后续 Worker 集成的基础能力；当前系统接口使用 JWT / Casbin，
新增外部接口需要显式接入 `ApiTokenAuth`，仅勾选 API 授权不会改变接口的认证方式。

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
