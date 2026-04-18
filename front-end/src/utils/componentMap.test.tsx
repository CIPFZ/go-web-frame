import { getComponent } from './componentMap';

describe('componentMap', () => {
  it('maps sys/api-token to a page component', () => {
    expect(getComponent('sys/api-token')).toBeDefined();
  });
});
