import { render, screen, waitFor } from '@testing-library/react';

import ApiTokenPage from './index';
import { getApiTokenList } from '@/services/api/apiToken';

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
    ProTable: ({ request, toolBarRender }: any) => {
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
    ModalForm: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProFormText: () => null,
    ProFormTextArea: () => null,
    ProFormSelect: () => null,
    ProFormDigit: () => null,
    ProFormDateTimePicker: () => null,
  };
});

describe('sys/api-token page', () => {
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
});
