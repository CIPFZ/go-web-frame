import type { API } from '@/services/system/types';
import { request } from '@umijs/max';

export interface MyNotice {
  ID: number;
  createdAt: string;
  title: string;
  content: string;
  level: 'info' | 'warning' | 'error';
  isPopup: boolean;
  needConfirm: boolean;
  startAt: string | null;
  endAt: string | null;
  readAt: string | null;
}

export interface MyNoticePage {
  list: MyNotice[] | null;
  total: number;
  page: number;
  pageSize: number;
}

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
  return request<Omit<API.CommonResponse, 'data'> & { data: MyNoticePage }>('/api/v1/sys/notice/getMyNotices', {
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
