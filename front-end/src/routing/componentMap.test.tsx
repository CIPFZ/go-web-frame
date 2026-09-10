import { getComponent } from './componentMap';

describe('componentMap', () => {
  it('maps sys/api-token to a page component', () => {
    expect(getComponent('sys/api-token')).toBeDefined();
  });

  it('maps system pages and hidden account settings used by backend menus', () => {
    for (const name of ['dashboard/workplace', 'components/RouterLayout', 'sys/user',
      'sys/authority', 'sys/menu', 'sys/api', 'sys/operation', 'sys/notice', 'user/info']) {
      expect(getComponent(name)).toBeDefined();
    }
  });
});
