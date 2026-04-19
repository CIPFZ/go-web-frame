import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import PluginProjectManagementPage from './index';
import { getDepartmentList, getPluginList } from '@/services/api/plugin';

let capturedColumns: any[] = [];
let mockCapturedSelectProps: any[] = [];

jest.mock('@/services/api/plugin', () => ({
  getPluginList: jest.fn(),
  getDepartmentList: jest.fn(),
  createPlugin: jest.fn(),
  updatePlugin: jest.fn(),
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
  const { Form } = require('antd');
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
    ModalForm: ({ children, form, open }: any) =>
      ReactLib.createElement(
        'div',
        { style: { display: open ? 'block' : 'none' } },
        ReactLib.createElement(Form, { form }, children),
      ),
    ProCard: ({ children }: any) => ReactLib.createElement('div', null, children),
    ProFormText: () => null,
    ProFormTextArea: () => null,
    ProFormSelect: (props: any) => {
      mockCapturedSelectProps.push(props);
      return ReactLib.createElement('div', null, props.label);
    },
  };
});

describe('plugin/project-management page', () => {
  beforeEach(() => {
    capturedColumns = [];
    mockCapturedSelectProps = [];
  });

  it('renders management summary and requests project list', async () => {
    (getPluginList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 3 } });
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectManagementPage));

    expect(await screen.findByTestId('plugin-project-table')).toBeTruthy();
    expect(screen.getByText('Plugin Project Management')).toBeTruthy();
    expect(screen.getByText('New Project')).toBeTruthy();

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

  it('renders current locale plugin name as a single line entry', async () => {
    (getPluginList as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            code: 'disk-analyzer',
            nameZh: '磁盘分析插件',
            nameEn: 'Disk Analyzer',
            descriptionZh: '诊断插件',
            descriptionEn: 'Diagnostic plugin',
            departmentId: 1,
            department: 'Storage',
            ownerId: 1,
            createdBy: 1,
            createdAt: '2026-04-18T10:00:00Z',
            repositoryUrl: 'https://example.com/repo.git',
          },
        ],
        total: 1,
      },
    });
    (getDepartmentList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectManagementPage));

    await waitFor(() => {
      expect(capturedColumns.length).toBeGreaterThan(0);
    });

    const nameColumn = capturedColumns.find((column) => column.dataIndex === 'nameZh');
    expect(nameColumn).toBeTruthy();
    const rendered = nameColumn.render(undefined, {
      nameZh: '磁盘分析插件',
      nameEn: 'Disk Analyzer',
      code: 'disk-analyzer',
    });
    expect(String(rendered.props.children)).toContain('Disk Analyzer');
  });

  it('loads department options with localized zh/en labels', async () => {
    (getPluginList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });
    (getDepartmentList as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            nameZh: '存储产品部',
            nameEn: 'Storage Product Dept',
            productLineZh: '基础软件',
            productLineEn: 'Base Software',
          },
        ],
        total: 1,
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectManagementPage));

    await waitFor(() => {
      expect(screen.getByText('New Project')).toBeTruthy();
    });

    fireEvent.click(screen.getByText('New Project'));

    await waitFor(() => {
      expect(getDepartmentList).toHaveBeenCalled();
    });

    const departmentSelect = [...mockCapturedSelectProps]
      .reverse()
      .find((item) => item.name === 'departmentId' && item.options?.length);
    expect(departmentSelect?.options).toEqual([
      { label: '基础软件 / 存储产品部 / Base Software / Storage Product Dept', value: 1 },
    ]);
  });
});
