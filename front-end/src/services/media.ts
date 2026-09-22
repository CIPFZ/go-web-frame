import { request } from '@umijs/max';

export type MediaSource = {
  id: number;
  name: string;
  apiUrl: string;
  apiType: 'json' | 'xml';
  enabled: boolean;
  priority: number;
  allowCategories?: string;
  denyCategories?: string;
  requestIntervalSeconds: number;
  timeoutSeconds: number;
  lastOkAt?: string;
  lastErrorAt?: string;
  lastError?: string;
};

export type MediaItem = {
  id: number;
  title: string;
  mediaType: string;
  year?: string;
  area?: string;
  coverUrl?: string;
  status: string;
};

type API<T> = { code: number; msg?: string; data: T };
export type MediaSourceInput = Omit<MediaSource, 'id' | 'lastOkAt' | 'lastErrorAt' | 'lastError'>;

const base = '/api/v1/media';
const sourcePath = (id?: number) => id ? base + '/sources/' + id : base + '/sources';

export const listMediaSources = () => request<API<MediaSource[]>>(sourcePath());
export const createMediaSource = (data: MediaSourceInput) =>
  request<API<MediaSource>>(sourcePath(), { method: 'POST', data });
export const updateMediaSource = (id: number, data: MediaSourceInput) =>
  request<API<MediaSource>>(sourcePath(id), { method: 'PUT', data });
export const checkMediaSource = (id: number) =>
  request<API<{ result: string; latencyMs: number; errorMessage?: string }>>(sourcePath(id) + '/health', { method: 'POST' });
export const collectMediaSource = (id: number, limit = 100) =>
  request<API<{ fetched: number; insertedCount: number; updatedCount: number; skippedCount: number; errorCount: number }>>(
    sourcePath(id) + '/collect?limit=' + limit,
    { method: 'POST' },
  );
export const listMediaItems = (limit = 100) =>
  request<API<MediaItem[]>>(base + '/items?limit=' + limit);
