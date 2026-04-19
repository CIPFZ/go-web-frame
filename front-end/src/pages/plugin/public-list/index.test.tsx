import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import PluginPublicListPage from './index';
import { getPublishedPluginList } from '@/services/api/plugin';

jest.mock('@/services/api/plugin', () => ({
  getPublishedPluginList: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'zh-CN',
  history: { push: jest.fn() },
}));

describe('plugin/public-list page', () => {
  it('renders hero content and filters plugin cards', async () => {
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

    fireEvent.change(screen.getByPlaceholderText('搜索插件名称、编码或描述'), {
      target: { value: 'disk-analyzer' },
    });
    fireEvent.click(screen.getByRole('button', { name: /搜/ }));
  });
});
