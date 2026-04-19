import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import PluginMasterPage from './index';
import { getDepartmentList } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getDepartmentList: jest.fn(),
  createDepartment: jest.fn(),
  updateDepartment: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'en-US',
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
      const { request, headerTitle, toolBarRender, columns } = props;
      ReactLib.useEffect(() => {
        void request?.({ current: 1, pageSize: 10 });
      }, [request]);
      const toolbar = toolBarRender?.() || [];
      const firstColumnTitle = columns?.[0]?.title;
      return ReactLib.createElement(
        'div',
        { 'data-testid': 'department-table' },
        headerTitle,
        firstColumnTitle,
        toolbar,
      );
    },
    ModalForm: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProCard: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProFormText: ({ label }: any) => ReactLib.createElement('div', null, label),
    ProFormSwitch: ({ label }: any) => ReactLib.createElement('div', null, label),
  };
});

describe('sys/plugin-master page', () => {
  it('renders department management and requests inactive-inclusive department data', async () => {
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginMasterPage));

    expect(await screen.findByText('Department Management')).toBeTruthy();
    expect(await screen.findByTestId('department-table')).toBeTruthy();
    expect(screen.queryByText('Plugin Master Data')).toBeNull();

    await waitFor(() => {
      expect(getDepartmentList).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, pageSize: 10, includeInactive: true }),
      );
    });
  });

  it('exposes create department action only', async () => {
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginMasterPage));

    expect(await screen.findByText('New Department')).toBeTruthy();
    fireEvent.click(screen.getByText('New Department'));
  });
});
