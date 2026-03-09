import { request } from '@umijs/max';

// 获取 API 列表 (分页)
export async function getApiList(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api/getApiList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 新增 API
export async function createApi(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api/createApi', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 更新 API
export async function updateApi(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api/updateApi', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 删除 API (支持单删和批删)
// 后端 DTO: { id: uint, ids: []uint }
export async function deleteApi(body: { id?: number; ids?: number[] }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api/deleteApi', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
