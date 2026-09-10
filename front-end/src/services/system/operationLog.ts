import { request } from '@umijs/max';

// 获取列表
export async function getOperationLogList(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/operationLog/getOperationLogList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 批量删除
export async function deleteOperationLogByIds(body: { ids: number[] }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/operationLog/deleteOperationLogByIds', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
