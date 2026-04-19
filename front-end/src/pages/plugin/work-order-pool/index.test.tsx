import { fireEvent, render, screen, waitFor } from '@testing-library/react';

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
  getLocale: () => 'en-US',
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
        void request?.({ current: 1, pageSize: 10 });
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
    expect(screen.getByText('Plugin Work Orders')).toBeTruthy();
    expect(screen.queryByText('Claimer')).toBeNull();
    expect(screen.queryByText(/Reviewers claim/)).toBeNull();

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

  it('renders all and mine tabs with search and reset actions', async () => {
    (getWorkOrderPool as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            pluginId: 1,
            pluginCode: 'disk-analyzer',
            pluginNameZh: '纾佺洏鍒嗘瀽鎻掍欢',
            version: '1.0.0',
            requestType: 1,
            status: 2,
            processStatus: 1,
          },
        ],
        total: 1,
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginWorkOrderPoolPage));

    expect(await screen.findByText('All')).toBeTruthy();
    expect(screen.getByText('Mine')).toBeTruthy();
    expect(screen.getByPlaceholderText('Search plugin / version / TD')).toBeTruthy();
    expect(screen.getByText('Search')).toBeTruthy();
    expect(screen.getByText('Reset')).toBeTruthy();

    fireEvent.click(screen.getByText('Mine'));

    await waitFor(() => {
      expect(getWorkOrderPool).toHaveBeenCalled();
    });

    expect(getWorkOrderPool).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 10 }),
    );
  });
});
