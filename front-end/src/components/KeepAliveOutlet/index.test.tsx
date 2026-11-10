import { act, render, screen } from '@testing-library/react';
import { createElement, useEffect } from 'react';

// 文件里有 jest.mock 时，测试转换会先改写 import 再编译 JSX，所以这里不用 JSX，直接 createElement

let mockPathname = '/reports/a';
let mockRenderPage: (name: string) => unknown = () => null;
jest.mock('@umijs/max', () => ({
  useLocation: () => ({ pathname: mockPathname }),
  useOutlet: () => mockRenderPage(mockPathname),
}));

import KeepAliveOutlet, { useRouteActive } from './index';

// jsdom 没有 ResizeObserver
(globalThis as any).ResizeObserver ??= class {
  observe() {}
  disconnect() {}
};

const mounts: string[] = [];
function Page({ name }: { name: string }) {
  const active = useRouteActive();
  useEffect(() => {
    mounts.push(name);
  }, [name]);
  return createElement('div', { 'data-testid': name }, active ? 'active' : 'hidden');
}

describe('KeepAliveOutlet', () => {
  it('keeps visited pages mounted and marks only the current one active', () => {
    mockRenderPage = (name) => createElement(Page, { name });
    const outlet = () => createElement(KeepAliveOutlet);
    const { rerender } = render(outlet());
    expect(screen.getByTestId('/reports/a').textContent).toBe('active');

    mockPathname = '/reports/b';
    act(() => rerender(outlet()));
    expect(screen.getByTestId('/reports/a').textContent).toBe('hidden');
    expect(screen.getByTestId('/reports/b').textContent).toBe('active');
    // 隐藏的页面移到屏幕外、不可交互，读屏也跳过
    const hidden = screen.getByTestId('/reports/a').parentElement as HTMLElement;
    expect(hidden.getAttribute('aria-hidden')).toBe('true');
    expect(hidden.style.position).toBe('fixed');
    expect(hidden.style.pointerEvents).toBe('none');

    mockPathname = '/reports/a';
    act(() => rerender(outlet()));
    expect(screen.getByTestId('/reports/a').textContent).toBe('active');
    expect(mounts).toEqual(['/reports/a', '/reports/b']);
  });
});
