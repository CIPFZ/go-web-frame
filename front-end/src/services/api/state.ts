import { request } from '@umijs/max';

/** 获取服务器信息接口 POST /api/v1/system/getServerInfo */
export async function getServerState(options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/system/getServerInfo', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}

