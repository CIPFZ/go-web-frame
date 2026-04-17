import routes from '../../config/routes';

const hasRoutePath = (list: any[] = [], path: string): boolean => {
  for (const route of list) {
    if (route?.path === path) {
      return true;
    }
    if (Array.isArray(route?.routes) && hasRoutePath(route.routes, path)) {
      return true;
    }
  }
  return false;
};

describe('routes config', () => {
  it('redirects the root route to dashboard workplace', () => {
    const layoutRoute = routes.find((route) => route.path === '/');

    expect(layoutRoute).toBeDefined();
    expect(layoutRoute?.routes?.[0]).toMatchObject({
      path: '/',
      redirect: '/dashboard/workplace',
    });
  });

  it('contains sys api-token route fallback', () => {
    expect(hasRoutePath(routes as any[], '/sys/api-token')).toBe(true);
  });
});
