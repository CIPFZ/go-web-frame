import { request } from '@umijs/max';
export interface SystemStatus {
  checks: Record<string, 'ok' | 'unavailable' | 'disabled'>;
  version: string;
  schemaVersion: string;
  goVersion: string;
  uptimeSeconds: number;
  observabilityEnabled: boolean;
}
export function getServerState() {
  return request<{code: number; msg: string; data: {server: SystemStatus}}>('/api/v1/sys/system/getServerInfo', {method: 'POST', data: {}});
}
