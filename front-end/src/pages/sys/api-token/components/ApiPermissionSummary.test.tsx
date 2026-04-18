import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';

import { ApiPermissionSummary } from './ApiPermissionSummary';

describe('ApiPermissionSummary', () => {
  it('shows compact summary and reveals hidden permissions on demand', async () => {
    render(
      <ApiPermissionSummary
        apis={[
          { ID: 1, method: 'GET', path: '/sys/user/list' },
          { ID: 2, method: 'POST', path: '/sys/user/create' },
          { ID: 3, method: 'DELETE', path: '/sys/user/remove' },
        ]}
      />,
    );

    expect(screen.getByText('3 个 API')).toBeTruthy();
    expect(screen.getByText('[GET] /sys/user/list')).toBeTruthy();
    expect(screen.getByText('[POST] /sys/user/create')).toBeTruthy();
    expect(screen.getByText('+1')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: /查看全部/ }));

    expect(await screen.findByText('[DELETE] /sys/user/remove')).toBeTruthy();
  });

  it('renders count on its own row and stacks summarized apis vertically', () => {
    render(
      <ApiPermissionSummary
        apis={[
          { ID: 1, method: 'GET', path: '/sys/user/list' },
          { ID: 2, method: 'POST', path: '/sys/user/create' },
          { ID: 3, method: 'DELETE', path: '/sys/user/remove' },
        ]}
      />,
    );

    expect(screen.getByTestId('api-permission-count-row')).toBeTruthy();
    expect(screen.getByTestId('api-permission-visible-list')).toBeTruthy();
    expect(screen.getAllByTestId('api-permission-visible-api-row')).toHaveLength(2);
    expect(screen.getByTestId('api-permission-overflow-row')).toBeTruthy();
  });

  it('limits full permission popover height and enables scrolling for long lists', async () => {
    render(
      <ApiPermissionSummary
        apis={Array.from({ length: 8 }).map((_, index) => ({
          ID: index + 1,
          method: 'GET',
          path: `/sys/api/${index + 1}`,
        }))}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: /查看全部/ }));

    const fullList = await screen.findByTestId('api-permission-popover-list');
    expect(fullList.style.maxHeight).toBe('240px');
    expect(fullList.style.overflowY).toBe('auto');
  });
});
