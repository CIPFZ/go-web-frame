import { render, screen, waitFor } from '@testing-library/react';

import PluginWorkOrderPoolPage from './index';
import { getWorkOrderPool } from '@/services/api/plugin';

let capturedColumns: any[] = [];

jest.mock('@/services/api/plugin', () => ({
  getWorkOrderPool: jest.fn(),
  claimWorkOrder: jest.fn(),
  resetWorkOrder: jest.fn(),
  transitionRelease: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'zh-CN',
  history: { push: jest.fn() },
}));

jest.mock('@ant-design/pro-layout', () => ({
  PageContainer: ({ children }: any) => {
    const ReactLib = require('react');
    return ReactLib.createElement('div', null, children);
  },
}));

jest.mock('@ant-design/pro-components', () => {
  const ReactLib = require('react');
  return {
    ProTable: (props: any) => {
      const { request, columns } = props;
      capturedColumns = columns || [];
      ReactLib.useEffect(() => {
        request?.({ current: 1, pageSize: 10 });
      }, [request]);
      return ReactLib.createElement('div', { 'data-testid': 'plugin-work-order-table' });
    },
    ModalForm: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProFormTextArea: () => null,
    ProCard: ({ children }: any) => ReactLib.createElement('div', null, children),
  };
});

describe('plugin/work-order-pool page', () => {
  beforeEach(() => {
    capturedColumns = [];
  });

  it('renders and requests work order pool data', async () => {
    (getWorkOrderPool as jest.Mock).mockResolvedValue({
      code: 0,
      data: { list: [], total: 0 },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginWorkOrderPoolPage));

    expect(await screen.findByTestId('plugin-work-order-table')).toBeTruthy();
    expect(screen.getByText('插件工单池')).toBeTruthy();

    await waitFor(() => {
      expect(getWorkOrderPool).toHaveBeenCalledTimes(1);
    });
  });

  it('includes an action column for review operations', async () => {
    (getWorkOrderPool as jest.Mock).mockResolvedValue({
      code: 0,
      data: { list: [], total: 0 },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginWorkOrderPoolPage));

    await waitFor(() => {
      expect(capturedColumns.length).toBeGreaterThan(0);
    });

    expect(capturedColumns.some((column) => column.dataIndex === 'option')).toBe(true);
  });
});
