import { render, screen, waitFor } from '@testing-library/react';

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
  getLocale: () => 'zh-CN',
  history: { push: jest.fn() },
  useParams: () => ({ id: '1' }),
}));

jest.mock('@ant-design/pro-layout', () => ({
  PageContainer: ({ children, extra }: any) => {
    const ReactLib = require('react');
    return ReactLib.createElement('div', null, extra, children);
  },
}));

describe('plugin/project-detail page', () => {
  it('loads project detail and shows back action', async () => {
    (getProjectDetail as jest.Mock).mockResolvedValue({
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
        releases: [],
        events: [],
      },
    });
    (getProductList as jest.Mock).mockResolvedValue({ code: 0, data: { list: [], total: 0 } });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginProjectDetailPage));

    expect(await screen.findByText('磁盘分析插件')).toBeTruthy();
    expect(screen.getByText('返回项目列表')).toBeTruthy();

    await waitFor(() => {
      expect(getProjectDetail).toHaveBeenCalledWith({ id: 1, releaseId: undefined });
    });
  });
});
