import { render, screen, waitFor } from '@testing-library/react';

import PluginPublicDetailPage from './index';
import { getPublishedPluginDetail } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getPublishedPluginDetail: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  history: { push: jest.fn() },
  useParams: () => ({ id: '1' }),
}));

describe('plugin/public-detail page', () => {
  it('loads market detail and renders version history', async () => {
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
          capabilityZh: '分析磁盘健康度与容量占用',
          capabilityEn: 'Analyze disk health and capacity',
          ownerName: '张三',
        },
        versions: [
          {
            releaseId: 11,
            version: '1.0.0',
            changelogZh: '首发版本',
            releasedAt: '2026-04-18T10:00:00Z',
            compatibleItems: [],
            packageX86Url: 'https://files.example.com/x86.zip',
            testReportUrl: 'https://files.example.com/report.pdf',
            publisher: '张三',
            versionConstraint: '>= 1.0.0',
          },
        ],
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginPublicDetailPage));

    expect((await screen.findAllByText('磁盘分析插件')).length).toBeGreaterThan(0);
    expect(screen.getByText('历史版本')).toBeTruthy();
    expect(screen.getByText('获取与安装')).toBeTruthy();

    await waitFor(() => {
      expect(getPublishedPluginDetail).toHaveBeenCalledWith({ pluginId: 1 }, { skipErrorHandler: true });
    });
  });
});
