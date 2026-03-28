export default [
  {
    path: '/plugins/:id',
    layout: false,
    component: './plugin/market/detail',
  },
  {
    path: '/plugins',
    layout: false,
    component: './plugin/market',
  },
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
  {
    path: '/',
    id: 'ant-design-pro-layout',
    routes: [
      {
        path: '/',
        redirect: '/dashboard/analysis',
      },
    ],
  },
  {
    path: '/plugin',
    routes: [
      {
        path: '/plugin/center',
        component: './plugin/project-center',
      },
      {
        path: '/plugin/market',
        component: './plugin/market',
      },
      {
        path: '/plugin/project-management',
        component: './plugin/project-center',
      },
      {
        path: '/plugin/review-workbench',
        component: './plugin/review-workbench',
      },
      {
        path: '/plugin/publish-workbench',
        component: './plugin/publish-workbench',
      },
      {
        path: '/plugin/project/:id',
        component: './plugin/project',
        hideInMenu: true,
      },
    ],
  },
  {
    path: '/account',
    routes: [
      {
        path: '/account/settings',
        component: './account/settings',
        hideInMenu: true,
      },
    ],
  },
  {
    path: '/*',
    component: '404',
  },
];
