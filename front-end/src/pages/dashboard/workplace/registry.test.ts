import {
  appTint,
  registerAppBadge,
  registerAppTint,
  registerLink,
  useAppBadge,
} from './registry';

jest.mock('@umijs/max', () => ({ request: jest.fn() }));

test('app colours: exact path first, then the longest registered prefix, then the default', () => {
  registerAppTint('/orders/', ['#111', '#222']);
  registerAppTint('/orders/vip/', ['#333', '#444']);
  registerAppTint('/orders/list', ['#555', '#666']);
  registerLink({
    url: 'https://example.com/docs',
    name: () => 'Docs',
    icon: null,
    tint: ['#777', '#888'],
  });
  expect(appTint('/orders/list')).toEqual(['#555', '#666']);
  expect(appTint('/orders/vip/today')).toEqual(['#333', '#444']);
  expect(appTint('/orders/report')).toEqual(['#111', '#222']);
  expect(appTint('https://example.com/docs')).toEqual(['#777', '#888']);
  expect(appTint('/other')).toEqual(['#94a3b8', '#64748b']);
});

test('only https links can be placed on the desk', () => {
  expect(() =>
    registerLink({ url: 'http://example.com', name: () => 'x', icon: null }),
  ).toThrow();
  expect(() =>
    registerLink({ url: 'javascript:alert(1)', name: () => 'x', icon: null }),
  ).toThrow();
});

test('every badge hook runs for every icon and the first non-zero count wins', () => {
  const calls: string[] = [];
  registerAppBadge((path) => {
    calls.push(`a:${path}`);
    return path === '/inbox' ? 3 : 0;
  });
  registerAppBadge((path) => {
    calls.push(`b:${path}`);
    return path === '/inbox' || path === '/tasks' ? 7 : 0;
  });
  expect([
    useAppBadge('/inbox'),
    useAppBadge('/tasks'),
    useAppBadge('/other'),
  ]).toEqual([3, 7, 0]);
  expect(calls).toEqual([
    'a:/inbox',
    'b:/inbox',
    'a:/tasks',
    'b:/tasks',
    'a:/other',
    'b:/other',
  ]);
});
