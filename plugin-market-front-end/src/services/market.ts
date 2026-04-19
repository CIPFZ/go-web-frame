import { request } from '@umijs/max';

import type { ApiResponse, PageResult, PublishedPluginDetail, PublishedPluginItem } from '@/types';

export async function getPluginList(keyword?: string) {
  return request<ApiResponse<PageResult<PublishedPluginItem>>>('/api/v1/market/plugins/list', {
    method: 'POST',
    data: {
      page: 1,
      pageSize: 60,
      keyword,
    },
  });
}

export async function getPluginDetail(pluginId: number) {
  return request<ApiResponse<PublishedPluginDetail>>('/api/v1/market/plugins/detail', {
    method: 'POST',
    data: { pluginId },
  });
}
