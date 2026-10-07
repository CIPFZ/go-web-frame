import { CheckCircleOutlined, CloseCircleOutlined, PlayCircleOutlined, StopOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, InputNumber, message, Space, Tag, Typography } from 'antd';
import { PageContainer } from '@ant-design/pro-components';
import { useState } from 'react';
import { cancelAgentTask, createAgentTask, getAgentStatus, getAgentTask, type AgentStatus, type AgentTask } from '@/services/agent';

const terminalStatuses = new Set(['succeeded', 'failed', 'cancelled', 'expired']);

export default function AgentPage() {
  const [form] = Form.useForm();
  const [status, setStatus] = useState<AgentStatus>();
  const [task, setTask] = useState<AgentTask>();
  const [loading, setLoading] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const loadStatus = async (name: string) => {
    const result = await getAgentStatus(name);
    if (result.code === 0 && result.data) setStatus(result.data);
    else message.error(result.msg);
    return result;
  };

  const submit = async (values: { vm: string; command: string; timeout_seconds?: number }) => {
    setLoading(true);
    try {
      const statusResult = await loadStatus(values.vm);
      if (statusResult.code !== 0) return;
      const timeout = values.timeout_seconds || 30;
      const result = await createAgentTask(values.vm, {
        caller_id: 'cms-ui',
        idempotency_key: `cms-ui-${values.vm}-${Date.now()}-${Math.random().toString(16).slice(2)}`,
        timeout_seconds: timeout,
        action: {
          type: 'command.exec',
          command: { argv: values.command.trim().split(/\s+/), timeout_seconds: timeout, output_limit_bytes: 1048576 },
        },
      });
      if (result.code === 0 && result.data) {
        setTask(result.data);
        message.success('任务已提交');
      } else {
        message.error(result.msg);
      }
    } finally {
      setLoading(false);
    }
  };

  const refreshTask = async () => {
    if (!task) return;
    const result = await getAgentTask(task.task_id);
    if (result.code === 0 && result.data) setTask(result.data);
    else message.error(result.msg);
  };

  const cancelTask = async () => {
    if (!task || terminalStatuses.has(task.status)) return;
    setCancelling(true);
    try {
      const result = await cancelAgentTask(task.task_id);
      if (result.code === 0 && result.data) setTask(result.data);
      else message.error(result.msg);
    } finally {
      setCancelling(false);
    }
  };

  return <PageContainer title="Agent 联动">
    <Card title="通过 Gateway 执行受控任务" extra={status && <Tag color={status.status === 'online' ? 'green' : 'orange'}>{status.status}</Tag>}>
      <Form form={form} layout="inline" onFinish={submit} initialValues={{ timeout_seconds: 30 }}>
        <Form.Item name="vm" label="虚拟机" rules={[{ required: true }]}><Input placeholder="VM 名称，同时作为 Agent ID" /></Form.Item>
        <Form.Item name="command" label="命令" rules={[{ required: true }]}><Input placeholder="printf hello" style={{ width: 260 }} /></Form.Item>
        <Form.Item name="timeout_seconds" label="超时"><InputNumber min={1} max={86400} /></Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} icon={<PlayCircleOutlined />}>提交</Button>
      </Form>
      {status && <Typography.Paragraph>Agent：{status.agent_id}，版本：{status.agent_version || '-'}，最近心跳：{status.last_heartbeat || '-'}</Typography.Paragraph>}
    </Card>
    {task && <Card title={`任务 ${task.task_id}`} style={{ marginTop: 16 }} extra={<Space><Button onClick={refreshTask}>刷新</Button>{!terminalStatuses.has(task.status) && <Button danger loading={cancelling} onClick={cancelTask} icon={<StopOutlined />}>取消</Button>}<Tag color={task.status === 'succeeded' ? 'green' : task.status === 'failed' ? 'red' : 'blue'}>{task.status}</Tag></Space>}>
      {task.result?.stdout && <Typography.Paragraph><pre style={{ whiteSpace: 'pre-wrap' }}>{task.result.stdout}</pre></Typography.Paragraph>}
      {task.result?.stderr && <Typography.Paragraph type="danger">{task.result.stderr}</Typography.Paragraph>}
      {task.result?.error_message && <Typography.Paragraph type="danger">{task.result.error_message}</Typography.Paragraph>}
      {task.status === 'succeeded' ? <CheckCircleOutlined style={{ color: 'green' }} /> : task.status === 'failed' ? <CloseCircleOutlined style={{ color: 'red' }} /> : null}
    </Card>}
  </PageContainer>;
}
