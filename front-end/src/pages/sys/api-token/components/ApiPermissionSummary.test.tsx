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

    fireEvent.click(screen.getByRole('button', { name: '查看全部' }));

    expect(await screen.findByText('[DELETE] /sys/user/remove')).toBeTruthy();
  });
});
