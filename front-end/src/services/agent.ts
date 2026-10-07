import { request } from '@umijs/max';

export interface AgentStatus { agent_id: string; agent_version: string; hostname?: string; status: string; registered_at: string; last_heartbeat: string; }
export interface AgentTask { task_id: string; agent_id?: string; status: string; action: Record<string, unknown>; created_at: string; finished_at?: string; result?: { status: string; error_code?: string; error_message?: string; stdout?: string; stderr?: string; exit_code?: number; artifacts?: Array<{ artifact_id: string; name?: string; media_type?: string; size_bytes?: number; sha256?: string; data_base64?: string }> }; }
export interface CreateAgentTaskRequest { caller_id?: string; idempotency_key?: string; action: Record<string, unknown>; timeout_seconds?: number; max_attempts?: number; }
export async function getAgentStatus(name: string) { return request<{ code: number; msg: string; data?: AgentStatus }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}/agent`, { method: 'GET' }); }
export async function checkAgent(name: string) { return request<{ code: number; msg: string; data?: { configured: boolean } }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}/agent/health`, { method: 'GET' }); }
export async function createAgentTask(name: string, data: CreateAgentTaskRequest) { return request<{ code: number; msg: string; data?: AgentTask }>(`/api/v1/virtualization/vms/${encodeURIComponent(name)}/agent/tasks`, { method: 'POST', data }); }
export async function getAgentTask(taskId: string) { return request<{ code: number; msg: string; data?: AgentTask }>(`/api/v1/virtualization/agent/tasks/${encodeURIComponent(taskId)}`, { method: 'GET' }); }
export async function cancelAgentTask(taskId: string) { return request<{ code: number; msg: string; data?: AgentTask }>(`/api/v1/virtualization/agent/tasks/${encodeURIComponent(taskId)}/cancel`, { method: 'POST' }); }
