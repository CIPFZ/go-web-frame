import { request } from '@umijs/max';

export async function createNotice(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/notice/createNotice', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getNoticeList(body: any, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/notice/getNoticeList', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}

export async function getMyNotices(params?: { page?: number; pageSize?: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/notice/getMyNotices', {
    method: 'GET',
    params,
    ...(options || {}),
  });
}

export async function markNoticeRead(body: { noticeId: number }, options?: { [key: string]: any }) {
  return request<API.CommonResponse>('/api/v1/sys/notice/markRead', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
    ...(options || {}),
  });
}
