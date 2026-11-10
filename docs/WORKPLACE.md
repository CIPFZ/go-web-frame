# 工作台

工作台是一张“桌面”：小组件与应用图标混排在网格上。小组件有小（2×2）、中（4×2）、大（4×4）三档，
问候还有“宽”（占满一行）；应用图标 1×1，打开一个页面或外部链接。

- 每个人点“编辑工作台”拖动、移除、调尺寸、从组件库添加，“完成”保存到本人账号（`sys_users.workplace`，`PUT /sys/user/workplace`）。
  编辑时图块不是链接，点击、拖动都不会打开页面。
- 管理员在“角色管理 → 默认工作台”里设计角色的默认工作台（`sys_authorities.workplace`，`PUT /sys/authority/workplace`），
  只能放这个角色有权限的内容。没改过工作台的人看到自己各角色的默认工作台按优先顺序拼在一起（重复的只留第一个）；
  角色没设计过时用预置的那一版；都没有时用起步桌面。
- 只显示本人能用的：应用图标需要页面的菜单权限，小组件需要它对应页面的权限（问候、通知与标了 `public` 的人人都有）。

## 项目扩展

框架自带问候、我的通知、运行概览、当前账号四个小组件，以及超级管理员（1）和普通用户（888）的预置
（`front-end/src/pages/dashboard/workplace/widgets/index.tsx`）。项目在 `front-end/src/workplace.tsx` 里注册自己的：

| 函数 | 作用 |
|---|---|
| `registerWidget` | 小组件：键、组件库分组、标题与说明、图标与底色、可用尺寸、打开的页面（`to`）、内容组件 |
| `registerWidgetGroup` | 组件库里的分组 |
| `registerAppTint` | 应用图标的颜色；以 `/` 结尾时对这个前缀下的页面生效 |
| `registerAppBadge` | 应用图标右上角的数字（React Hook，每个图标都会调用，不是自己的 path 返回 0） |
| `registerLink` | 外部链接（只能是 `https://`），人人可放 |
| `registerRolePreset` / `setRolePriority` / `setStarterDesk` | 角色预置、多个角色合并的顺序、起步桌面 |

小组件的 `to` 和预置里的应用图标可以写页面的组件标识（如 `'state'`、`'sys/user'`），按后端菜单找到实际地址，
管理员改了菜单地址也不受影响。小组件的数据用 `widgets/shared.tsx` 的 `useShared`（同一份数据只查一次、60 秒刷新、
页面在后台时不查），非正常状态用 `StatusView` 显示（加载中、需要登录、未配置、失败可重试）。

布局保存的是小组件的键：改了键，已保存的布局里就找不到它（会被去掉）。
