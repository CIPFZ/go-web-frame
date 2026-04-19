import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';

import PluginProjectDetailPage from './index';
import { getProductList, getProjectDetail } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getProjectDetail: jest.fn(),
  getProductList: jest.fn(),
  createRelease: jest.fn(),
  updateRelease: jest.fn(),
  transitionRelease: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'en-US',
  history: { push: jest.fn() },
  useParams: () => ({ id: '1' }),
}));

jest.mock('@ant-design/pro-layout', () => ({
  PageContainer: ({ children, extra }: any) => {
    const ReactLib = require('react');
    return ReactLib.createElement('div', null, extra, children);
  },
}));

const buildDetail = () => ({
  code: 0,
  data: {
    plugin: {
      ID: 1,
      code: 'disk-analyzer',
      repositoryUrl: 'https://example.com/repo.git',
      nameZh: '磁盘分析插件',
      nameEn: 'Disk Analyzer',
      descriptionZh: '用于磁盘诊断',
      descriptionEn: 'Disk diagnostics',
      departmentId: 1,
      department: '存储产品部',
      ownerId: 1,
      createdBy: 1,
      createdAt: '2026-04-18T10:00:00Z',
    },
    selectedRelease: {
      ID: 11,
      pluginId: 1,
      pluginCode: 'disk-analyzer',
      pluginNameZh: '磁盘分析插件',
      requestType: 1,
      status: 1,
      processStatus: 0,
      version: '1.0.0',
      testReportUrl: 'https://files.example.com/report.pdf',
      packageX86Url: 'https://files.example.com/x86.zip',
      packageArmUrl: 'https://files.example.com/arm.zip',
      changelogZh: '初始版本',
      changelogEn: 'Initial release',
      tdId: 'TD-101',
      createdBy: 1,
      createdAt: '2026-04-18T10:00:00Z',
      compatibleItems: [],
    },
    releases: [
      {
        ID: 11,
        pluginId: 1,
        pluginCode: 'disk-analyzer',
        pluginNameZh: '磁盘分析插件',
        requestType: 1,
        status: 1,
        processStatus: 0,
        version: '1.0.0',
        createdBy: 1,
        createdAt: '2026-04-18T10:00:00Z',
        compatibleItems: [],
      },
      {
        ID: 12,
        pluginId: 1,
        pluginCode: 'disk-analyzer',
        pluginNameZh: '磁盘分析插件',
        requestType: 1,
        status: 2,
        processStatus: 1,
        version: '1.1.0',
        createdBy: 1,
        createdAt: '2026-04-19T10:00:00Z',
        compatibleItems: [],
      },
      {
        ID: 13,
        pluginId: 1,
        pluginCode: 'disk-analyzer',
        pluginNameZh: '磁盘分析插件',
        requestType: 1,
        status: 4,
        processStatus: 2,
        version: '1.2.0',
        createdBy: 1,
        createdAt: '2026-04-20T10:00:00Z',
        compatibleItems: [],
      },
    ],
    events: [
      { ID: 101, action: 'submit_review', comment: 'Submitted', createdAt: '2026-04-18T12:00:00Z' },
      { ID: 102, action: 'approve', comment: 'Approved', createdAt: '2026-04-18T14:00:00Z' },
    ],
  },
});

describe('plugin/project-detail page', () => {
  beforeEach(() => {
    (getProjectDetail as jest.Mock).mockResolvedValue(buildDetail());
    (getProductList as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          { ID: 1, code: 'EulerOS', name: 'EulerOS' },
          { ID: 2, code: 'aCLI', name: 'aCLI' },
        ],
        total: 2,
      },
    });
  });

  it('loads project detail and shows back action', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectDetailPage));

    expect(await screen.findByText('Disk Analyzer')).toBeTruthy();
    expect(screen.getByText('Back to Projects')).toBeTruthy();

    await waitFor(() => {
      expect(getProjectDetail).toHaveBeenCalledWith({ id: 1, releaseId: undefined });
    });
  });

  it('filters releases through a single version search box', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectDetailPage));

    const search = await screen.findByPlaceholderText('Search Version');
    expect(search).toBeTruthy();
    expect(screen.queryByText('Filter by Status')).toBeNull();

    fireEvent.change(search, { target: { value: '1.1' } });

    await waitFor(() => {
      expect(screen.queryByTestId('release-card-11')).toBeNull();
      expect(screen.getByTestId('release-card-12')).toBeTruthy();
      expect(screen.queryByTestId('release-card-13')).toBeNull();
    });
  });

  it('renders adjusted compatibility fields in release form', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectDetailPage));

    fireEvent.click(await screen.findByRole('button', { name: 'New Release' }));

    const dialog = await screen.findByRole('dialog');
    const modal = within(dialog);

    expect(modal.getAllByText('Product Compatibility').length).toBeGreaterThan(0);
    expect(modal.getAllByText('aCLI Compatibility').length).toBeGreaterThan(0);
    expect(modal.queryByText('Universal Support')).toBeNull();
    expect(modal.queryByText('TD ID')).toBeNull();
    expect(modal.getByPlaceholderText('Enter aCLI version compatibility, for example >= 1.0.0')).toBeTruthy();
    expect(modal.getAllByText('Product').length).toBeGreaterThan(0);
    expect(modal.getByRole('button', { name: /Add Product Compatibility/i })).toBeTruthy();

    const modalText = dialog.textContent || '';
    expect(modalText.indexOf('Product Compatibility')).toBeLessThan(modalText.indexOf('Upload Test Report'));
  });

  it('keeps edit action only for editable statuses', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectDetailPage));

    const readyCard = await screen.findByTestId('release-card-11');
    const rejectedCard = await screen.findByTestId('release-card-13');
    const pendingCard = await screen.findByTestId('release-card-12');

    expect(within(readyCard).getByRole('button', { name: 'Edit' })).toBeTruthy();
    expect(within(rejectedCard).getByRole('button', { name: 'Edit' })).toBeTruthy();
    expect(within(pendingCard).queryByRole('button', { name: 'Edit' })).toBeNull();
  });
});
