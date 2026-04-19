export default [
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
        redirect: '/dashboard/workplace',
      },
      {
        path: '/sys/api-token',
        component: './sys/api-token',
        hideInMenu: true,
      },
      {
        path: '/plugins',
        component: './plugin/public-list',
        layout: false,
        hideInMenu: true,
      },
      {
        path: '/plugins/:id',
        component: './plugin/public-detail',
        layout: false,
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
