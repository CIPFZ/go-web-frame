import { request } from '@umijs/max';

// 获取角色列表
export async function getAuthorityList(body?: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/authority/getAuthorityList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 创建角色
export async function createAuthority(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/authority/createAuthority', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 更新角色
export async function updateAuthority(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/authority/updateAuthority', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 删除角色 (body: { id: number })
// 注意：后端Delete使用 request.GetByIdReq (id uint)，前端传 {id: authorityId}
export async function deleteAuthority(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/authority/deleteAuthority', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 设置角色菜单权限
export async function setAuthorityMenus(body: { authorityId: number; menuIds: number[] }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/authority/setAuthorityMenus', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
