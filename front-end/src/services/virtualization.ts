import { request } from '@umijs/max';

export interface VirtualMachine {
  name: string;
  uuid: string;
  state: string;
  memoryMiB: number;
  vcpus: number;
  autostart: boolean;
  diskPaths: string[];
  network?: string;
  xml?: string;
}

export interface CreateVirtualMachineRequest {
  name: string;
  memoryMiB?: number;
  vcpus?: number;
  diskPath?: string;
  diskSizeGiB?: number;
  diskFormat?: 'qcow2' | 'raw';
  pool?: string;
  network?: string;
  autostart?: boolean;
  start?: boolean;
}

export interface UpdateVirtualMachineRequest {
  memoryMiB?: number;
  vcpus?: number;
  autostart?: boolean;
}

export async function getVirtualMachineList() {
  return request<{ code: number; msg: string; data?: VirtualMachine[] }>('/api/v1/virtualization/vms', { method: 'GET' });
}

export async function getVirtualMachine(name: string) {
  return request<{ code: number; msg: string; data?: VirtualMachine }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}`, { method: 'GET' });
}

export async function createVirtualMachine(data: CreateVirtualMachineRequest) {
  return request<{ code: number; msg: string; data?: VirtualMachine }>('/api/v1/virtualization/vms', { method: 'POST', data });
}

export async function updateVirtualMachine(name: string, data: UpdateVirtualMachineRequest) {
  return request<{ code: number; msg: string; data?: VirtualMachine }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}`, { method: 'PUT', data });
}

export async function virtualMachineAction(name: string, action: 'start' | 'shutdown' | 'reboot' | 'force-stop') {
  return request<{ code: number; msg: string }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}/${action}`, { method: 'POST' });
}

export async function deleteVirtualMachine(name: string) {
  return request<{ code: number; msg: string }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}`, { method: 'DELETE' });
}
