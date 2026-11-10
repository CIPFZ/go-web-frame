import {
  appItem,
  defaultDesk,
  presetDesk,
  resolveDesk,
  roleDesk,
  sortRoles,
  widgetItem,
} from './layout';
import {
  registerRolePreset,
  registerWidget,
  setRolePriority,
  setStarterDesk,
} from './registry';

jest.mock('@umijs/max', () => ({ request: jest.fn() }));

const Body = () => null;
for (const [key, sizes] of [
  ['greeting', ['w', 's', 'm']],
  ['notices', ['m', 's', 'l']],
  ['service', ['s', 'm']],
  ['orders', ['m', 's', 'l']],
] as const) {
  registerWidget({
    key,
    group: 'test',
    title: () => key,
    description: () => key,
    icon: null,
    tint: ['#000', '#000'],
    sizes: [...sizes],
    Body,
  });
}
setStarterDesk({
  widgets: [
    ['greeting', 'w'],
    ['orders', 'm'],
    ['service', 's'],
  ],
  apps: ['/orders/list', '/sys/user'],
});
registerRolePreset(1, {
  widgets: [
    ['greeting', 'w'],
    ['service', 'm'],
  ],
  apps: ['/sys/user'],
});
registerRolePreset(888, {
  widgets: [
    ['greeting', 'w'],
    ['notices', 'm'],
  ],
  apps: ['/orders/list'],
});
registerRolePreset(20, {
  widgets: [
    ['orders', 'l'],
    ['notices', 's'],
  ],
  apps: ['/orders/list', '/orders/report'],
});
setRolePriority([1, 20, 888]);

// 只看每一项的种类与尺寸（id 是随机的）
const shape = (
  d: {
    items: (
      | { kind: 'widget'; widget: string; size: string }
      | { kind: 'app'; path: string }
    )[];
  } | null,
) =>
  d?.items.map((i) =>
    i.kind === 'widget' ? `${i.widget}:${i.size}` : `app:${i.path}`,
  );

test('nothing stored or an unknown format means the default desk', () => {
  expect(resolveDesk(undefined)).toBeNull();
  expect(resolveDesk({ v: 9, items: [] })).toBeNull();
  expect(resolveDesk({ v: 1 })).toBeNull();
});

test('the starter desk only has what the user can use', () => {
  const desk = defaultDesk({
    widget: (w) => w !== 'service',
    page: (p) => p.startsWith('/orders/'),
  });
  expect(shape(desk)).toEqual(['greeting:w', 'orders:m', 'app:/orders/list']);
});

test('stored items are cleaned: unknown widgets, bad sizes, duplicate apps and ids, unsafe links; component ids are kept', () => {
  const a = widgetItem('orders', 'l');
  const d = resolveDesk({
    v: 1,
    items: [
      a,
      { ...a },
      { kind: 'widget', widget: 'service', size: 'l' },
      { kind: 'widget', widget: 'nope', size: 's' },
      appItem('/x'),
      appItem('/x'),
      { kind: 'app', path: 'not a page!' },
      appItem('sys/user'),
      appItem('https://example.com/docs'),
      appItem('http://example.com'),
      appItem('//example.com'),
      appItem('javascript:alert(1)'),
    ],
  });
  expect(shape(d)).toEqual([
    'orders:l',
    'orders:l',
    'service:s',
    'app:/x',
    'app:sys/user',
    'app:https://example.com/docs',
  ]);
  expect(new Set(d?.items.map((i) => i.id)).size).toBe(6);
});

test('several roles merge their default desks in priority order, without duplicates', () => {
  // 20 排在 888 前面：20 的在前，888 的接在后面；问候、通知、/orders/list 只留一个
  expect(shape(roleDesk([{ authorityId: 888 }, { authorityId: 20 }]))).toEqual([
    'orders:l',
    'notices:s',
    'app:/orders/list',
    'app:/orders/report',
    'greeting:w',
  ]);
  // 设计过的用设计的
  const designed = {
    v: 1,
    items: [
      { id: 'x', kind: 'widget', widget: 'notices', size: 's' },
      { id: 'y', kind: 'app', path: '/orders/list' },
    ],
  };
  expect(shape(roleDesk([{ authorityId: 888, workplace: designed }]))).toEqual([
    'notices:s',
    'app:/orders/list',
  ]);
  // 没有预置、也没设计过的角色跳过
  expect(shape(roleDesk([{ authorityId: 77 }, { authorityId: 888 }]))).toEqual(
    shape(presetDesk(888)),
  );
  expect(roleDesk([{ authorityId: 77 }])).toBeNull();
  expect(roleDesk([])).toBeNull();
});

test('roles outside the priority list come last, by id', () => {
  expect(
    sortRoles([
      { authorityId: 888 },
      { authorityId: 30 },
      { authorityId: 1 },
      { authorityId: 5 },
    ]).map((r) => r.authorityId),
  ).toEqual([1, 888, 5, 30]);
});
