import { render, screen, waitFor } from '@testing-library/react';

import PluginPublicDetailPage from './index';
import { getPublishedPluginDetail } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getPublishedPluginDetail: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'zh-CN',
  history: { push: jest.fn() },
  useParams: () => ({ id: '1' }),
}));

describe('plugin/public-detail page', () => {
  it('loads public detail and renders version history', async () => {
    (getPublishedPluginDetail as jest.Mock).mockResolvedValue({
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
        release: {
          ID: 11,
          pluginId: 1,
          pluginCode: 'disk-analyzer',
          pluginNameZh: '磁盘分析插件',
          requestType: 1,
          status: 5,
          processStatus: 3,
          version: '1.0.0',
          compatibleItems: [],
          createdBy: 1,
          createdAt: '2026-04-18T10:00:00Z',
        },
        versions: [
          {
            ID: 11,
            version: '1.0.0',
            changelogZh: '首发版本',
            releasedAt: '2026-04-18T10:00:00Z',
            compatibleItems: [],
          },
        ],
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginPublicDetailPage));

    expect(await screen.findByText('磁盘分析插件')).toBeTruthy();
    expect(screen.getByText('版本历史')).toBeTruthy();

    await waitFor(() => {
      expect(getPublishedPluginDetail).toHaveBeenCalledWith({ id: 1 }, { skipErrorHandler: true });
    });
  });
});
