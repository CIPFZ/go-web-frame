import { request } from '@umijs/max';

export interface StoragePool {
  name: string;
  uuid: string;
  state: string;
  capacityBytes: number;
  allocationBytes: number;
  availableBytes: number;
  autostart: boolean;
  volumeCount: number;
}

export interface StorageVolume {
  pool: string;
  name: string;
  path: string;
  type: string;
  format: string;
  capacityBytes: number;
  allocationBytes: number;
}

export async function getStoragePools() {
  return request<{ code: number; msg: string; data?: StoragePool[] }>('/api/v1/virtualization/storage/pools', { method: 'GET' });
}

export async function getStorageVolumes(pool?: string) {
  const query = pool ? `?pool=${encodeURIComponent(pool)}` : '';
  return request<{ code: number; msg: string; data?: StorageVolume[] }>(`/api/v1/virtualization/storage/volumes${query}`, { method: 'GET' });
}

export async function createStorageVolume(data: { pool: string; name: string; sizeGiB: number; format: 'qcow2' | 'raw' }) {
  return request<{ code: number; msg: string; data?: StorageVolume }>('/api/v1/virtualization/storage/volumes', { method: 'POST', data });
}

export async function resizeStorageVolume(pool: string, name: string, sizeGiB: number) {
  return request<{ code: number; msg: string; data?: StorageVolume }>(`/api/v1/virtualization/storage/volumes/${encodeURIComponent(pool)}/${encodeURIComponent(name)}`, { method: 'PUT', data: { sizeGiB } });
}

export async function deleteStorageVolume(pool: string, name: string) {
  return request<{ code: number; msg: string }>(`/api/v1/virtualization/storage/volumes/${encodeURIComponent(pool)}/${encodeURIComponent(name)}`, { method: 'DELETE' });
}

