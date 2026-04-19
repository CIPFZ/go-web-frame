import { render, screen, waitFor } from '@testing-library/react';

import PluginProjectManagementPage from './index';
import { getDepartmentList, getPluginList } from '@/services/api/plugin';

let capturedColumns: any[] = [];

jest.mock('@/services/api/plugin', () => ({
  getPluginList: jest.fn(),
  getDepartmentList: jest.fn(),
  createPlugin: jest.fn(),
  updatePlugin: jest.fn(),
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
      const { request, toolBarRender, columns } = props;
      capturedColumns = columns || [];
      ReactLib.useEffect(() => {
        request?.({ current: 1, pageSize: 10 });
      }, [request]);
      return ReactLib.createElement(
        'div',
        null,
        ReactLib.createElement('div', null, toolBarRender?.()),
        ReactLib.createElement('div', { 'data-testid': 'plugin-project-table' }),
      );
    },
    ModalForm: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProCard: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProFormText: () => null,
    ProFormTextArea: () => null,
    ProFormSelect: () => null,
  };
});

describe('plugin/project-management page', () => {
  beforeEach(() => {
    capturedColumns = [];
  });

  it('renders management summary and requests project list', async () => {
    (getPluginList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 3 } });
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectManagementPage));

    expect(await screen.findByTestId('plugin-project-table')).toBeTruthy();
    expect(screen.getByText('插件项目管理')).toBeTruthy();
    expect(screen.getByText('新建项目')).toBeTruthy();

    await waitFor(() => {
      expect(getPluginList).toHaveBeenCalledTimes(1);
    });
  });

  it('keeps the detail action column', async () => {
    (getPluginList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectManagementPage));

    await waitFor(() => {
      expect(capturedColumns.length).toBeGreaterThan(0);
    });

    expect(capturedColumns.some((column) => column.dataIndex === 'option')).toBe(true);
  });
});
