import { request } from '@umijs/max';

export type ApiTokenItem = {
  ID: number;
  tokenPrefix: string;
  name: string;
  description?: string;
  enabled: boolean;
  maxConcurrency?: number;
  expiresAt?: string;
  lastUsedAt?: string;
  apis?: Array<{
    ID: number;
    path: string;
    method: string;
    apiGroup?: string;
    description?: string;
  }>;
};

export type ApiTokenListParams = {
  page?: number;
  pageSize?: number;
  name?: string;
  description?: string;
  enabled?: boolean;
};

export type ApiTokenFormPayload = {
  id?: number;
  name: string;
  description?: string;
  maxConcurrency?: number;
  expiresAt?: string;
  apiIds?: number[];
};

export async function getApiTokenList(body: ApiTokenListParams, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/getApiTokenList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function createApiToken(body: ApiTokenFormPayload, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getApiTokenDetail(params: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/detail', {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

export async function updateApiToken(body: ApiTokenFormPayload, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/update', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function deleteApiToken(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/delete', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function resetApiToken(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/reset', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function enableApiToken(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/enable', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function disableApiToken(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api-token/disable', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getApiOptions(body: { page?: number; pageSize?: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/api/getApiList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
