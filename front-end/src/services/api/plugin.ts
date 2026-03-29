import { request } from '@umijs/max';

export type ReleaseRequestType = 'initial' | 'maintenance' | 'offline';
export type ReleaseStatus =
  | 'draft'
  | 'release_preparing'
  | 'pending_review'
  | 'approved'
  | 'rejected'
  | 'released'
  | 'offlined';

export interface ReleaseListItem {
  ID: number;
  pluginId: number;
  pluginCode: string;
  pluginNameZh: string;
  pluginNameEn: string;
  requestType: ReleaseRequestType;
  status: ReleaseStatus;
  version: string;
  versionConstraint: string;
  publisher: string;
  reviewerId?: number | null;
  publisherId?: number | null;
  reviewComment: string;
  isOfflined: boolean;
  createdBy: number;
  createdAt: string;
  submittedAt?: string | null;
  approvedAt?: string | null;
  releasedAt?: string | null;
  offlinedAt?: string | null;
}

export interface ReleaseListQuery {
  page?: number;
  pageSize?: number;
  keyword?: string;
  pluginId?: number;
  requestType?: ReleaseRequestType;
  status?: ReleaseStatus;
  createdBy?: number;
  reviewerId?: number;
  publisherId?: number;
}

export interface ReleaseTransitionPayload {
  id: number;
  action: 'submit_review' | 'approve' | 'reject' | 'release' | 'revise';
  reviewComment?: string;
  reviewerId?: number;
  publisherId?: number;
}

export interface ReleaseAssignPayload {
  id: number;
  reviewerId?: number;
  publisherId?: number;
  comment?: string;
}

export async function getPluginList(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/plugin/getPluginList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getProjectDetail(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/plugin/getProjectDetail', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getPluginOverview(options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/plugin/getPluginOverview', {
    method: 'GET',
    ...(options || {}),
  });
}

export async function createPlugin(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/plugin/createPlugin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function updatePlugin(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/plugin/updatePlugin', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getReleaseList(body: ReleaseListQuery, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/getReleaseList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getReleaseDetail(body: { id: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/getReleaseDetail', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function createRelease(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/createRelease', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function updateRelease(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/updateRelease', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function transitRelease(body: ReleaseTransitionPayload, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/transition', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function assignRelease(body: ReleaseAssignPayload, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/release/assign', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getPublishedPluginList(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/public/getPublishedPluginList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getPublishedPluginDetail(body: { pluginId: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/plugin/public/getPublishedPluginDetail', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function uploadPluginAsset(file: File, options?: { [key: string]: any }) {
  const formData = new FormData();
  formData.append('file', file);
  return request<API.CommonResponse>('/api/v1/sys/file/upload', {
    method: 'POST',
    data: formData,
    requestType: 'form',
    ...(options || {}),
  });
}
