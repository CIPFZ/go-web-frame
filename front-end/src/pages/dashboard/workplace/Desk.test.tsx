import { render } from '@testing-library/react';
import React from 'react';
import { AppTile } from './Desk';

jest.mock('@umijs/max', () => {
  const R = jest.requireActual('react');
  return {
    getLocale: () => 'zh-CN',
    request: jest.fn(),
    history: { push: jest.fn() },
    Link: ({ to, children, ...rest }: { to: string; children: unknown }) =>
      R.createElement('a', { href: `#${to}`, ...rest }, children),
  };
});

jest.mock('@/i18n/menus', () => ({
  menuLabel: (m: { name: string }) => m.name,
}));

const h = React.createElement;

test('in edit mode an app icon is not a link, so clicking or dragging it never opens the page', () => {
  for (const path of ['/orders/list', 'https://example.com/docs']) {
    const menu = { path, name: 'x' };
    const normal = render(h(AppTile, { menu }));
    expect(normal.container.querySelector('a')).not.toBeNull();
    normal.unmount();
    const editing = render(h(AppTile, { menu, editing: true }));
    expect(editing.container.querySelector('a')).toBeNull();
    editing.unmount();
  }
});
