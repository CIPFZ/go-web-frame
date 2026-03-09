// src/services/api/casbin.ts
import { request } from '@umijs/max';

// 更新 Casbin 权限
export async function updateCasbin(body: { authorityId: string, casbinInfos: { path: string, method: string }[] }) {
  return request<API.CommonResponse>('/api/v1/sys/casbin/updateCasbin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
  });
}

// 获取角色当前的 Casbin 权限
export async function getPolicyPathByAuthorityId(body: { authorityId: string }) {
  return request<API.CommonResponse>('/api/v1/sys/casbin/getPolicyPathByAuthorityId', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
  });
}
