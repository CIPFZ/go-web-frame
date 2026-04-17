import routes from '../../config/routes';

describe('routes config', () => {
  it('redirects the root route to dashboard workplace', () => {
    const layoutRoute = routes.find((route) => route.path === '/');

    expect(layoutRoute).toBeDefined();
    expect(layoutRoute?.routes?.[0]).toMatchObject({
      path: '/',
      redirect: '/dashboard/workplace',
    });
  });
});
