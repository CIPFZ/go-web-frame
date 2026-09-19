import { request } from '@umijs/max';

export type ProxyStatus = {
  state: string;
  subState?: string;
  mainPid?: number;
  error?: string;
  updatedAt: string;
};
export type ProxyInstance = {
  id: number;
  name: string;
  engine: string;
  scope: string;
  unit: string;
  binaryPath: string;
  configPath: string;
  enabled: boolean;
  description?: string;
  status: ProxyStatus;
};
export type ProxyMetrics = {
  pid: number;
  readBytes: number;
  writeBytes: number;
  readCalls: number;
  writeCalls: number;
  connections: number;
  updatedAt: string;
};
export type ProxyConfig = { content: string; digest: string; size: number; modified: string };
type API<T> = { code: number; msg?: string; data: T };

const base = '/api/v1/proxy/instances';
const pathFor = (id: number, suffix: string) => base + '/' + id + suffix;
export const listProxyInstances = () => request<API<ProxyInstance[]>>(base);
export const actionProxy = (id: number, action: string) =>
  request<API<ProxyStatus>>(pathFor(id, '/action'), { method: 'POST', data: { action } });
export const readProxyConfig = (id: number) => request<API<ProxyConfig>>(pathFor(id, '/config'));
export const validateProxyConfig = (id: number) =>
  request<API<{ message: string }>>(pathFor(id, '/config/validate'), { method: 'POST' });
export const saveProxyConfig = (id: number, content: string) =>
  request<API<{ digest: string; backup?: string; validated: boolean }>>(pathFor(id, '/config'), { method: 'PUT', data: { content } });
export const rollbackProxyConfig = (id: number) =>
  request<API<{ digest: string; backup?: string; validated: boolean }>>(pathFor(id, '/config/rollback'), { method: 'POST' });
export const readProxyMetrics = (id: number) => request<API<ProxyMetrics>>(pathFor(id, '/metrics'));
