import { request } from '@umijs/max';

export type ResourceType = 'video' | 'audio' | 'image' | 'document' | 'archive' | 'disk_image' | 'other';
export type PreviewFile = {
  index: number; path: string; name: string; size: number; size_human: string;
  extension?: string; type: ResourceType;
};
export type MagnetPreview = {
  info_hash: string; name: string; title: string; display_name: string;
  content_type: ResourceType; total_size_human: string; total_size: number;
  file_count: number; files_truncated: boolean; files: PreviewFile[];
  trackers: string[]; cached: boolean; retrieved_at: string;
  cover?: { url: string; filename: string; size: number };
};

export function previewMagnet(magnet: string) {
  return request<{ code: number; msg: string; data: MagnetPreview }>('/api/v1/magnet/preview', {
    method: 'POST', data: { magnet }, timeout: 60000,
  });
}

export function loadCover(url: string) {
  return request<Blob>(url, { responseType: 'blob', timeout: 10000 });
}
