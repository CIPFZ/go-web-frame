export default [
  // 1. 登录/注册页 Layout (无侧边栏)
  {
    path: '/user',
    layout: false,
    routes: [
      {
        name: 'login',
        path: '/user/login',
        component: './user/login',
      },
      {
        name: 'register',
        path: '/user/register',
        component: './user/register',
      },
      {
        name: 'register-result',
        path: '/user/register-result',
        component: './user/register-result',
      },
      {
        path: '/user',
        redirect: '/user/login',
      },
      {
        component: '404',
        path: '/user/*',
      },
    ],
  },

  // 2. ✨✨✨ 主系统 Layout (关键补充) ✨✨✨
  // patchClientRoutes 会找到这个路由，并将动态菜单注入到它的 routes 中
  {
    path: '/',
    // 给个 ID 方便 patchClientRoutes 查找 (虽然 path='/' 也能找)
    id: 'ant-design-pro-layout',
    routes: [
      // 这里的 redirect 只是一个静态兜底。
      // 实际运行时，patchClientRoutes 会根据用户角色动态修改这里的跳转目标。
      {
        path: '/',
        redirect: '/dashboard/workplace',
      },
      {
        path: '/sys/api-token',
        component: './sys/api-token',
        hideInMenu: true,
      },
      // ... 动态路由会被注入到这里 ...
    ],
  },

  {
    path: '/account',
    routes: [
      {
        path: '/account/settings',
        component: './account/settings',
        hideInMenu: true,
      }
    ],
  },

  // 3. 全局 404
  {
    path: '/*',
    component: '404',
  },
];
