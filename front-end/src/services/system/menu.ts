import { request } from '@umijs/max';

/** 获取菜单接口 GET /api/v1/menu/getMenu */
export async function getMenu(options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/getMenu', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    ...(options || {}),
  });
}

// 获取菜单列表 (树形)
export async function getMenuList(body?: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/getMenuList', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body, // 虽然是全量，但保持分页参数格式
    ...(options || {}),
  });
}

// 新增菜单
export async function addBaseMenu(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/addBaseMenu', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 更新菜单
export async function updateBaseMenu(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/updateBaseMenu', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

// 删除菜单
export async function deleteBaseMenu(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/deleteBaseMenu', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getMenuAuthority(body: { authorityId: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/menu/getMenuAuthority', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
