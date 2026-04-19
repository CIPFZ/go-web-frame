import { act, fireEvent, render, screen } from '@testing-library/react';
import * as React from 'react';

import PluginDetailPage from './index';

jest.mock('@umijs/max', () => ({
  history: { push: jest.fn() },
  useParams: () => ({ id: '101' }),
}));

jest.mock('@/services/market', () => ({
  getPluginDetail: jest.fn(),
}));

const { getPluginDetail } = jest.requireMock('@/services/market');

describe('PluginDetailPage', () => {
  beforeEach(() => {
    window.localStorage.setItem('plugin-market-standalone-locale', 'en');
  });

  it('renders CMS-style detail sections and download panel', async () => {
    getPluginDetail.mockResolvedValue({
      code: 0,
      data: {
        plugin: {
          ID: 1,
          pluginId: 101,
          code: 'agent-helper',
          nameZh: '智能助手插件',
          nameEn: 'Agent Helper',
          descriptionZh: '用于自动巡检',
          descriptionEn: 'For automation inspection',
          capabilityZh: '自动巡检',
          capabilityEn: 'Automation inspection',
          ownerName: 'Platform Team',
        },
        release: {
          releaseId: 201,
          version: '1.0.0',
          compatibleItems: [],
        },
        versions: [
          {
            releaseId: 201,
            version: '1.0.0',
            publisher: 'CMS',
            packageX86Url: 'https://example.com/x86.zip',
            packageArmUrl: 'https://example.com/arm.zip',
            testReportUrl: 'https://example.com/report.pdf',
            changelogEn: 'Initial release',
            performanceSummaryEn: 'Stable in smoke tests',
            compatibleItems: [],
          },
        ],
      },
    });

    await act(async () => {
      render(React.createElement(PluginDetailPage));
    });

    expect(await screen.findByText('Plugin Ecosystem')).toBeInTheDocument();
    expect(screen.getAllByText('Agent Helper').length).toBeGreaterThan(0);
    expect(screen.getByText('Overview')).toBeInTheDocument();
    expect(screen.getByText('Changelog')).toBeInTheDocument();
    expect(screen.getByText('Testing & Performance')).toBeInTheDocument();
    expect(screen.getByText('Version Metadata')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Download Standard/ })).toBeInTheDocument();
  });

  it('switches selected version from history table', async () => {
    getPluginDetail.mockResolvedValue({
      code: 0,
      data: {
        plugin: {
          ID: 1,
          pluginId: 101,
          code: 'agent-helper',
          nameZh: '智能助手插件',
          nameEn: 'Agent Helper',
          descriptionZh: '用于自动巡检',
          descriptionEn: 'For automation inspection',
          capabilityZh: '自动巡检',
          capabilityEn: 'Automation inspection',
          ownerName: 'Platform Team',
        },
        release: {
          releaseId: 202,
          version: '1.1.0',
          compatibleItems: [],
        },
        versions: [
          {
            releaseId: 202,
            version: '1.1.0',
            publisher: 'CMS',
            changelogEn: 'Second release',
            compatibleItems: [],
          },
          {
            releaseId: 201,
            version: '1.0.0',
            publisher: 'CMS',
            changelogEn: 'Initial release',
            compatibleItems: [],
          },
        ],
      },
    });

    await act(async () => {
      render(React.createElement(PluginDetailPage));
    });

    fireEvent.click(await screen.findByText('History'));
    fireEvent.click(screen.getByText('v1.0.0'));

    expect(await screen.findByText('Viewing')).toBeInTheDocument();
  });
});
