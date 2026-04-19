import { getComponent } from './componentMap';

describe('componentMap', () => {
  it('maps sys/api-token to a page component', () => {
    expect(getComponent('sys/api-token')).toBeDefined();
  });

  it('maps plugin management pages used by dynamic routes', () => {
    expect(getComponent('plugin/project-management')).toBeDefined();
    expect(getComponent('plugin/project-detail')).toBeDefined();
    expect(getComponent('plugin/work-order-pool')).toBeDefined();
    expect(getComponent('plugin/work-order-detail')).toBeDefined();
    expect(getComponent('sys/plugin-master')).toBeDefined();
  });
});
