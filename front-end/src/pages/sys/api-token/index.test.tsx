import { render, screen, waitFor } from '@testing-library/react';

import ApiTokenPage from './index';
import { getApiTokenList } from '@/services/api/apiToken';

const mockProFormDateTimePicker = jest.fn((_: any) => null);
let capturedColumns: any[] = [];

jest.mock('@/services/api/apiToken', () => ({
  getApiTokenList: jest.fn(),
  createApiToken: jest.fn(),
  getApiTokenDetail: jest.fn(),
  updateApiToken: jest.fn(),
  deleteApiToken: jest.fn(),
  resetApiToken: jest.fn(),
  enableApiToken: jest.fn(),
  disableApiToken: jest.fn(),
  getApiOptions: jest.fn(),
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
      const { request, toolBarRender, columns } = props;
      capturedColumns = columns || [];
      ReactLib.useEffect(() => {
        request?.({ current: 1, pageSize: 10 });
      }, [request]);
      return ReactLib.createElement(
        'div',
        null,
        ReactLib.createElement('div', null, toolBarRender?.()),
        ReactLib.createElement('div', { 'data-testid': 'api-token-table' }),
      );
    },
    DrawerForm: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProCard: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProForm: {
      Item: ({ children }: any) => ReactLib.createElement('div', null, children),
    },
    ProFormText: () => null,
    ProFormTextArea: () => null,
    ProFormDigit: () => null,
    ProFormDateTimePicker: (props: any) => mockProFormDateTimePicker(props),
  };
});

describe('sys/api-token page', () => {
  beforeEach(() => {
    mockProFormDateTimePicker.mockClear();
    capturedColumns = [];
  });

  it('renders and requests list data', async () => {
    (getApiTokenList as jest.Mock).mockResolvedValue({
      code: 0,
      data: { list: [], total: 0 },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(ApiTokenPage));

    expect(await screen.findByTestId('api-token-table')).toBeTruthy();
    expect(screen.getByText('新建 Token')).toBeTruthy();

    await waitFor(() => {
      expect(getApiTokenList).toHaveBeenCalledTimes(1);
    });
  });

  it('marks expiresAt as required in the token form', async () => {
    (getApiTokenList as jest.Mock).mockResolvedValue({
      code: 0,
      data: { list: [], total: 0 },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(ApiTokenPage));

    await waitFor(() => {
      expect(mockProFormDateTimePicker).toHaveBeenCalled();
    });

    const firstCall = mockProFormDateTimePicker.mock.calls[0]?.[0];

    expect(firstCall?.rules).toEqual(
      expect.arrayContaining([expect.objectContaining({ required: true })]),
    );
  });

  it('places authorized api column before status and usage metadata', async () => {
    (getApiTokenList as jest.Mock).mockResolvedValue({
      code: 0,
      data: { list: [], total: 0 },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(ApiTokenPage));

    await waitFor(() => {
      expect(capturedColumns.length).toBeGreaterThan(0);
    });

    const order = capturedColumns.map((column) => String(column.dataIndex));

    expect(order.indexOf('apis')).toBeGreaterThan(-1);
    expect(order.indexOf('apis')).toBeLessThan(order.indexOf('enabled'));
    expect(order.indexOf('apis')).toBeLessThan(order.indexOf('expiresAt'));
  });
});
