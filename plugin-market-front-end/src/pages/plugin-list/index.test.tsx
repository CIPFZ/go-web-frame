import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import * as React from 'react';

import PluginListPage from './index';

jest.mock('@umijs/max', () => ({
  history: { push: jest.fn() },
}));

jest.mock('@/services/market', () => ({
  getPluginList: jest.fn(),
}));

const { getPluginList } = jest.requireMock('@/services/market');

describe('PluginListPage', () => {
  beforeEach(() => {
    window.localStorage.setItem('plugin-market-standalone-locale', 'en');
  });

  it('renders CMS-style filter controls and plugin cards', async () => {
    getPluginList.mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            pluginId: 101,
            code: 'agent-helper',
            nameZh: '智能助手插件',
            nameEn: 'Agent Helper',
            descriptionZh: '用于自动巡检',
            descriptionEn: 'For automation inspection',
            latestVersion: '1.0.0',
            releasedAt: '2026-04-19T10:00:00Z',
            packageX86Url: 'https://example.com/x86.zip',
            packageArmUrl: 'https://example.com/arm.zip',
            compatibleItems: [],
          },
        ],
      },
    });

    await act(async () => {
      render(React.createElement(PluginListPage));
    });

    expect(await screen.findByText('Plugin Ecosystem')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
    expect(screen.getByText('Architecture')).toBeInTheDocument();
    expect(screen.getByText('Sort by')).toBeInTheDocument();
    expect(screen.getByText('Agent Helper')).toBeInTheDocument();
    expect(screen.getByText('x86_64')).toBeInTheDocument();
    expect(screen.getByText('ARM64')).toBeInTheDocument();
  });

  it('filters plugins by category and keyword', async () => {
    getPluginList.mockResolvedValue({
      code: 0,
      data: {
        list: [
          {
            ID: 1,
            pluginId: 101,
            code: 'agent-helper',
            nameZh: '智能助手插件',
            nameEn: 'Agent Helper',
            descriptionZh: '用于自动巡检',
            descriptionEn: 'For automation inspection',
            latestVersion: '1.0.0',
            releasedAt: '2026-04-19T10:00:00Z',
            compatibleItems: [],
          },
          {
            ID: 2,
            pluginId: 102,
            code: 'image-optimizer',
            nameZh: '图像优化插件',
            nameEn: 'Image Optimizer',
            descriptionZh: '用于图像处理',
            descriptionEn: 'Image enhancement and optimization',
            latestVersion: '2.0.0',
            releasedAt: '2026-04-18T10:00:00Z',
            compatibleItems: [],
          },
        ],
      },
    });

    await act(async () => {
      render(React.createElement(PluginListPage));
    });

    fireEvent.click(await screen.findByText('Image'));

    await waitFor(() => {
      expect(screen.queryByText('Agent Helper')).not.toBeInTheDocument();
    });
    expect(screen.getByText('Image Optimizer')).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('Search plugins by name, code, or keywords'), {
      target: { value: 'optimizer' },
    });
    fireEvent.keyDown(screen.getByPlaceholderText('Search plugins by name, code, or keywords'), {
      key: 'Enter',
      code: 'Enter',
    });

    await waitFor(() => {
      expect(getPluginList).toHaveBeenCalledWith('optimizer');
    });
  });
});
