import { request } from '@umijs/max';

export interface NovelBook {
  id: number;
  title: string;
  author?: string;
  category?: string;
  sourceUrl?: string;
  sourceHash?: string;
  recordHash?: string;
  language?: string;
  formats?: string[];
  createdAt: string;
  updatedAt: string;
}

export interface NovelBookPage {
  list: NovelBook[];
  total: number;
  page: number;
  pageSize: number;
}

export async function getNovelBookList(body: Record<string, unknown>) {
  return request<{ code: number; msg: string; data?: NovelBookPage }>('/api/v1/novel/book/list', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    data: body,
  });
}

export async function getNovelBookDetail(id: number) {
  return request<{ code: number; msg: string; data?: NovelBook }>(`/api/v1/novel/book/detail?id=${id}`, {
    method: 'GET',
  });
}
