import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import Page from './index';
import { previewMagnet } from '@/services/magnet';

jest.mock('@umijs/max', () => ({ useIntl: () => ({ locale: 'zh-CN', formatMessage: ({ id }: {id:string}) => require('@/locales/zh-CN/magnet').default[id] }) }));
jest.mock('@ant-design/pro-components', () => ({ PageContainer: ({ children }: any) => require('react').createElement('div', null, children) }));
jest.mock('@/services/magnet', () => ({ previewMagnet: jest.fn(), loadCover: jest.fn() }));

const preview = previewMagnet as jest.Mock;
describe('Magnet resource preview', () => {
  beforeAll(() => {
    window.matchMedia = jest.fn().mockImplementation(() => ({
      matches: false, addListener: jest.fn(), removeListener: jest.fn(),
      addEventListener: jest.fn(), removeEventListener: jest.fn(), dispatchEvent: jest.fn(),
    }));
  });
  beforeEach(() => jest.clearAllMocks());

  it.each(['video', 'audio', 'image', 'document', 'archive', 'disk_image', 'other'])('shows %s metadata without requiring artwork', async type => {
    preview.mockResolvedValue({ code: 0, data: {
      info_hash: '0123456789abcdef0123456789abcdef01234567', name: 'Example resource',
      content_type: type, total_size_human: '4.0 GiB', file_count: 1, retrieved_at: '2026-09-18T10:00:00Z',
      files: [{ index: 0, path: 'folder/example.bin', type, size: 4096, size_human: '4.0 KiB' }],
    } });
    render(React.createElement(Page));
    fireEvent.change(screen.getByRole('textbox', { name: '磁力链接' }), { target: { value: 'magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567' } });
    fireEvent.click(screen.getByRole('button', { name: /解析资源/ }));
    await waitFor(() => expect(screen.getByText('Example resource')).toBeTruthy());
    expect(screen.getByText('暂无可用封面')).toBeTruthy();
    expect(screen.getByText('folder/example.bin')).toBeTruthy();
    expect(screen.getByText('4.0 GiB')).toBeTruthy();
  });

  it('renders resolver errors without pretending a preview succeeded', async () => {
    preview.mockResolvedValue({ code: 7, msg: 'No peers available' });
    render(React.createElement(Page));
    fireEvent.change(screen.getByRole('textbox', { name: '磁力链接' }), { target: { value: 'magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567' } });
    fireEvent.click(screen.getByRole('button', { name: /解析资源/ }));
    await waitFor(() => expect(screen.getByText('No peers available')).toBeTruthy());
    expect(screen.queryByText('暂无可用封面')).toBeNull();
  });
});
