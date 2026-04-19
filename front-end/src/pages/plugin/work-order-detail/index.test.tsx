import { fireEvent, render, screen, waitFor } from '@testing-library/react';

import PluginWorkOrderDetailPage from './index';
import { getProjectDetail } from '@/services/api/plugin';
import { history } from '@umijs/max';

jest.mock('@/services/api/plugin', () => ({
  getProjectDetail: jest.fn(),
  transitionRelease: jest.fn(),
  claimWorkOrder: jest.fn(),
  resetWorkOrder: jest.fn(),
}));

jest.mock('@umijs/max', () => ({
  getLocale: () => 'zh-CN',
  history: { push: jest.fn() },
  useParams: () => ({ id: '22', pluginId: '1' }),
}));

jest.mock('@ant-design/pro-layout', () => ({
  PageContainer: ({ children, extra }: any) => {
    const ReactLib = require('react');
    return ReactLib.createElement('div', null, extra, children);
  },
}));

describe('plugin/work-order-detail page', () => {
  beforeEach(() => {
    (history.push as jest.Mock).mockReset();
    (getProjectDetail as jest.Mock).mockResolvedValue({
      code: 0,
      data: {
        plugin: {
          ID: 1,
          code: 'disk-analyzer',
          nameZh: '磁盘分析插件',
          nameEn: 'Disk Analyzer',
          descriptionZh: '用于磁盘诊断',
          descriptionEn: 'Disk diagnostics',
          repositoryUrl: 'https://example.com/repo.git',
          department: '存储产品部',
          createdAt: '2026-04-18T10:00:00Z',
        },
        selectedRelease: {
          ID: 22,
          pluginId: 1,
          pluginCode: 'disk-analyzer',
          pluginNameZh: '磁盘分析插件',
          version: '1.1.0',
          requestType: 1,
          status: 2,
          processStatus: 1,
          claimerName: '张三',
          claimerUsername: 'zhangsan',
          changelogZh: '修复兼容问题',
          testReportUrl: 'https://files.example.com/report.pdf',
          packageX86Url: 'https://files.example.com/x86.zip',
          packageArmUrl: 'https://files.example.com/arm.zip',
          compatibleInfo: {
            universal: false,
            products: [{ productId: 1, productCode: 'EulerOS', productName: '欧拉系统', versionConstraint: '>= 5.0' }],
            acli: [{ productId: 2, productCode: 'aCLI', productName: 'aCLI', versionConstraint: '>= 1.2' }],
          },
          createdBy: 1,
          createdAt: '2026-04-19T10:00:00Z',
        },
        releases: [],
        events: [{ ID: 1, action: 'submit_review', comment: '提交审核', createdAt: '2026-04-19T11:00:00Z' }],
      },
    });
  });

  it('renders read-only work order detail and back navigation', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginWorkOrderDetailPage));

    expect(await screen.findByText('工单详情')).toBeTruthy();
    expect(await screen.findByText('磁盘分析插件')).toBeTruthy();
    expect(screen.getByText('产品兼容')).toBeTruthy();
    expect(screen.getByText('aCLI 兼容')).toBeTruthy();
    expect(screen.getByText('张三')).toBeTruthy();
    expect(screen.queryByRole('textbox', { name: '版本号' })).toBeNull();
    expect(screen.queryByRole('button', { name: '编辑发布单' })).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: '返回工单池' }));
    expect(history.push).toHaveBeenCalledWith('/plugin/work-order-pool');
  });

  it('shows reviewer transition actions only for allowed states', async () => {
    const ReactLib = require('react');
    render(ReactLib.createElement(PluginWorkOrderDetailPage));

    await waitFor(() => {
      expect(getProjectDetail).toHaveBeenCalledWith({ id: 1, releaseId: 22 });
    });

    expect(await screen.findByRole('button', { name: '审核通过' })).toBeTruthy();
    expect(screen.getByRole('button', { name: /打\s*回/ })).toBeTruthy();
    expect(screen.queryByRole('button', { name: '认领' })).toBeNull();
  });
});
