import { Outlet } from '@umijs/max';

/**
 * 这是一个用于父级菜单的占位符组件
 * 它只渲染一个 <Outlet />，以便子路由可以被正确渲染
 */
const RouterLayout = () => {
  return <Outlet />;
};

export default RouterLayout;
