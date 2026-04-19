import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import PluginPublicListPage from './index';
import { getPublishedPluginList } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getPublishedPluginList: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  history: { push: jest.fn() },
}));

describe('plugin/public-list page', () => {
  it('renders market hero content and filters plugin cards', async () => {
    (getPublishedPluginList as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            code: 'disk-analyzer',
            nameZh: '磁盘分析插件',
            nameEn: 'Disk Analyzer',
            descriptionZh: '用于磁盘诊断',
            descriptionEn: 'Disk diagnostics',
            latestVersion: '1.0.0',
            compatibleItems: [],
            releasedAt: '2026-04-18T10:00:00Z',
            packageX86Url: 'https://files.example.com/x86.zip',
          },
        ],
        total: 1,
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginPublicListPage));

    expect(await screen.findByText('插件生态中心')).toBeTruthy();
    await waitFor(() => {
      expect(getPublishedPluginList).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText('磁盘分析插件')).toBeTruthy();

    fireEvent.change(screen.getByPlaceholderText('搜索插件名称、编码或功能关键词'), {
      target: { value: 'disk-analyzer' },
    });
    fireEvent.click(screen.getByRole('button', { name: /搜\s*索/ }));

    await waitFor(() => {
      expect(getPublishedPluginList).toHaveBeenLastCalledWith(
        { page: 1, pageSize: 60, keyword: 'disk-analyzer' },
        { skipErrorHandler: true },
      );
    });
  });

  it('renders release date as a single horizontal footer item', async () => {
    (getPublishedPluginList as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 2,
            code: 'network-auditor',
            nameZh: '缃戠粶瀹¤鎻掍欢',
            nameEn: 'Network Auditor',
            descriptionZh: '鐢ㄤ簬缃戠粶瀹¤',
            descriptionEn: 'Network audit plugin',
            latestVersion: '2.0.0',
            compatibleItems: [],
            releasedAt: '2026-04-18T10:00:00Z',
            packageX86Url: 'https://files.example.com/x86.zip',
            packageArmUrl: 'https://files.example.com/arm.zip',
          },
        ],
        total: 1,
      },
    });

    const ReactLib = require('react');
    render(ReactLib.createElement(PluginPublicListPage));

    expect(await screen.findByText('2026-04-18')).toBeTruthy();
    expect(screen.getByText('x86_64')).toBeTruthy();
    expect(screen.getByText('ARM64')).toBeTruthy();
  });
});
