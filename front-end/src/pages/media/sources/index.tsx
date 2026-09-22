import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Col, Form, Input, InputNumber, Modal, Progress, Row, Space, Switch, Table, Tag, Typography, message } from 'antd';
import { CheckCircleOutlined, EditOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons';
import {
  cancelCollectionRun, checkMediaSource, createMediaSource, listCollectionRuns, listMediaSchedules, listMediaSources,
  updateMediaSchedule, updateMediaSource, type CollectionRun, type MediaSource, type MediaSourceInput, type Schedule,
} from '@/services/media';

const emptySource: MediaSourceInput = {
  name: '', apiUrl: '', apiType: 'json', enabled: true, priority: 100, fullEnabled: false,
  allowCategories: '', denyCategories: '', requestIntervalSeconds: 3, timeoutSeconds: 15,
};
const scheduleNames: Record<string, string> = { health: '健康检查', incremental: '增量采集', full: '全量采集' };
const statusColor: Record<string, string> = { queued: 'blue', running: 'processing', success: 'green', partial: 'orange', failed: 'red', cancelled: 'default' };

const formatTime = (value?: string) => value ? new Date(value).toLocaleString() : '-';
const modeName = (value: string) => value === 'sample' ? '试运行' : value === 'incremental' ? '增量' : value === 'full' ? '全量' : value;
const runPercent = (run: CollectionRun) => {
  if (run.status === 'success') return 100;
  if (run.maxItems > 0) return Math.min(99, Math.round((run.fetched / run.maxItems) * 100));
  return run.status === 'running' ? 50 : 0;
};

export default function MediaSourcesPage() {
  const [items, setItems] = useState<MediaSource[]>([]);
  const [runs, setRuns] = useState<CollectionRun[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<MediaSource | undefined>();
  const [editorOpen, setEditorOpen] = useState(false);
  const [form] = Form.useForm<MediaSourceInput>();

  const loadRuns = async () => {
    const result = await listCollectionRuns();
    if (result.code === 0) setRuns(result.data || []);
  };
  const refresh = async () => {
    setLoading(true);
    try {
      const [sourceResult, runResult, scheduleResult] = await Promise.all([listMediaSources(), listCollectionRuns(), listMediaSchedules()]);
      if (sourceResult.code === 0) setItems(sourceResult.data || []); else message.error(sourceResult.msg || '获取资源站失败');
      if (runResult.code === 0) setRuns(runResult.data || []);
      if (scheduleResult.code === 0) setSchedules(scheduleResult.data || []);
    } catch { message.error('加载影视采集配置失败'); }
    finally { setLoading(false); }
  };
  useEffect(() => {
    refresh();
    const timer = window.setInterval(() => { loadRuns().catch(() => undefined); }, 5000);
    return () => window.clearInterval(timer);
  }, []);

  const sourceMap = useMemo(() => new Map(items.map(item => [item.id, item.name])), [items]);
  const openEditor = (item?: MediaSource) => {
    setEditing(item);
    setEditorOpen(true);
    form.setFieldsValue(item ? {
      name: item.name, apiUrl: item.apiUrl, apiType: item.apiType, enabled: item.enabled, fullEnabled: item.fullEnabled,
      priority: item.priority, allowCategories: item.allowCategories || '', denyCategories: item.denyCategories || '',
      requestIntervalSeconds: item.requestIntervalSeconds, timeoutSeconds: item.timeoutSeconds,
    } : emptySource);
  };
  const save = async () => {
    const values = await form.validateFields();
    try {
      const result = editing ? await updateMediaSource(editing.id, values) : await createMediaSource(values);
      if (result.code !== 0) { message.error(result.msg || '保存失败'); return; }
      message.success('来源配置已保存');
      setEditing(undefined); setEditorOpen(false); await refresh();
    } catch { message.error('保存来源失败'); }
  };
  const health = async (item: MediaSource) => {
    try {
      const result = await checkMediaSource(item.id);
      if (result.code === 0 && result.data.result === 'ok') message.success(`${item.name} 检查通过`);
      else message.warning(result.data?.errorMessage || result.msg || '来源不可用');
      await refresh();
    } catch { message.error('健康检查失败'); }
  };
  const saveSchedule = async (schedule: Schedule) => {
    const result = await updateMediaSchedule(schedule.id, { cron: schedule.cron, enabled: schedule.enabled });
    if (result.code === 0) { message.success('调度已保存'); await refresh(); }
    else message.error(result.msg || '保存调度失败');
  };
  const changeSchedule = (id: number, patch: Partial<Schedule>) => {
    setSchedules(prev => prev.map(item => item.id === id ? { ...item, ...patch } : item));
  };

  return <PageContainer
    title="影视资源采集"
    subTitle="采集由 CMS 内置调度器执行，所有运行结果都会保留在任务日志中"
    extra={<Space><Button icon={<ReloadOutlined />} onClick={refresh}>刷新全部</Button><Button type="primary" icon={<EditOutlined />} onClick={() => openEditor()}>新增来源</Button></Space>}
  >
    <Row gutter={[16, 16]}>
      <Col span={24}>
        <Card title="资源站配置" extra={<Typography.Text type="secondary">只负责来源和策略配置</Typography.Text>} bodyStyle={{ padding: 0 }}>
          <Table<MediaSource> rowKey="id" loading={loading} dataSource={items} pagination={false} columns={[
            { title: '来源', dataIndex: 'name', render: (value, item) => <Space direction="vertical" size={0}><Typography.Text strong>{value}</Typography.Text><Typography.Text type="secondary" ellipsis style={{ maxWidth: 360 }}>{item.apiUrl}</Typography.Text></Space> },
            { title: '优先级', dataIndex: 'priority', width: 90 },
            { title: '状态', dataIndex: 'enabled', width: 100, render: (value: boolean, item) => <Tag color={value && !item.suspended ? 'green' : 'default'}>{item.suspended ? '暂挂' : value ? '启用' : '停用'}</Tag> },
            { title: '成功/失败', width: 110, render: (_, item) => <Typography.Text>{item.successes} / {item.failures}</Typography.Text> },
            { title: '最近成功', dataIndex: 'lastOkAt', width: 170, render: formatTime },
            { title: '操作', width: 190, render: (_, item) => <Space><Button size="small" icon={<CheckCircleOutlined />} onClick={() => health(item)}>检查</Button><Button size="small" icon={<EditOutlined />} onClick={() => openEditor(item)}>编辑</Button></Space> },
          ]} />
        </Card>
      </Col>
      <Col span={24}>
        <Card title="定时采集策略" extra={<Typography.Text type="secondary">启用后按 Cron 自动执行</Typography.Text>} bodyStyle={{ padding: 0 }}>
          <Table<Schedule> rowKey="id" loading={loading} dataSource={schedules} pagination={false} columns={[
            { title: '来源', width: 150, render: (_, item) => sourceMap.get(item.sourceId) || `来源 #${item.sourceId}` },
            { title: '任务类型', width: 130, render: (_, item) => <Tag color={item.kind === 'health' ? 'blue' : item.kind === 'full' ? 'orange' : 'purple'}>{scheduleNames[item.kind] || item.kind}</Tag> },
            { title: 'Cron', width: 220, render: (_, item) => <Input value={item.cron} onChange={event => changeSchedule(item.id, { cron: event.target.value })} /> },
            { title: '启用', width: 80, render: (_, item) => <Switch checked={item.enabled} onChange={checked => changeSchedule(item.id, { enabled: checked })} /> },
            { title: '下次执行', width: 180, render: (_, item) => formatTime(item.nextRunAt) },
            { title: '最近结果', width: 120, render: (_, item) => item.lastStatus ? <Tag color={statusColor[item.lastStatus] || 'default'}>{item.lastStatus}</Tag> : '-' },
            { title: '操作', width: 100, render: (_, item) => <Button size="small" icon={<SaveOutlined />} onClick={() => saveSchedule(item)}>保存</Button> },
          ]} />
        </Card>
      </Col>
      <Col span={24}>
        <Card title="采集日志" extra={<Space size="small"><Tag color="processing">每 5 秒刷新</Tag><Typography.Text type="secondary">保留最近 100 条任务</Typography.Text></Space>} bodyStyle={{ padding: 0 }}>
          <Table<CollectionRun> rowKey="id" dataSource={runs} pagination={{ pageSize: 10 }} columns={[
            { title: '任务', width: 85, render: (_, item) => <Typography.Text strong>#{item.id}</Typography.Text> },
            { title: '来源', width: 140, render: (_, item) => sourceMap.get(item.sourceId) || `来源 #${item.sourceId}` },
            { title: '类型', width: 100, render: (_, item) => modeName(item.mode) },
            { title: '状态 / 进度', width: 220, render: (_, item) => <Space direction="vertical" size={2} style={{ width: '100%' }}><Tag color={statusColor[item.status] || 'default'}>{item.status}</Tag><Progress percent={runPercent(item)} size="small" status={item.status === 'failed' ? 'exception' : item.status === 'success' ? 'success' : 'active'} format={() => `${item.fetched} 条 / ${item.pages} 页`} /></Space> },
            { title: '结果统计', width: 190, render: (_, item) => <Typography.Text type="secondary">新增 {item.insertedCount} · 更新 {item.updatedCount} · 跳过 {item.skippedCount} · 错误 {item.errorCount}</Typography.Text> },
            { title: '开始时间', width: 180, render: (_, item) => formatTime(item.startedAt) },
            { title: '结束时间', width: 180, render: (_, item) => item.finishedAt ? formatTime(item.finishedAt) : '执行中' },
            { title: '错误信息', render: (_, item) => item.errorMessage ? <Typography.Text type="danger" ellipsis={{ tooltip: item.errorMessage }}>{item.errorMessage}</Typography.Text> : '-' },
            { title: '操作', width: 90, render: (_, item) => item.status === 'queued' || item.status === 'running' ? <Button danger size="small" onClick={async () => { await cancelCollectionRun(item.id); message.success('已请求取消'); await loadRuns(); }}>取消</Button> : null },
          ]} />
        </Card>
      </Col>
    </Row>
    <Modal open={editorOpen} title={editing ? '编辑来源' : '新增来源'} onCancel={() => { setEditing(undefined); setEditorOpen(false); }} onOk={save} destroyOnHidden width={720}>
      <Form form={form} layout="vertical" initialValues={emptySource}>
        <Row gutter={16}>
          <Col span={12}><Form.Item name="name" label="来源名称" rules={[{ required: true }]}><Input maxLength={100} /></Form.Item></Col>
          <Col span={12}><Form.Item name="priority" label="优先级"><InputNumber min={1} max={10000} style={{ width: '100%' }} /></Form.Item></Col>
        </Row>
        <Form.Item name="apiUrl" label="HTTPS JSON 接口地址" rules={[{ required: true }, { type: 'url' }]}><Input placeholder="https://example.com/api.php/provide/vod/" /></Form.Item>
        <Row gutter={16}>
          <Col span={8}><Form.Item name="timeoutSeconds" label="超时（秒）"><InputNumber min={1} max={120} style={{ width: '100%' }} /></Form.Item></Col>
          <Col span={8}><Form.Item name="requestIntervalSeconds" label="请求间隔（秒）"><InputNumber min={0} max={60} style={{ width: '100%' }} /></Form.Item></Col>
          <Col span={8}><Form.Item name="enabled" label="来源启用" valuePropName="checked"><Switch /></Form.Item></Col>
        </Row>
        <Form.Item name="allowCategories" label="分类白名单（逗号分隔）"><Input placeholder="电影,电视剧,动漫,综艺" /></Form.Item>
        <Form.Item name="denyCategories" label="分类黑名单（逗号分隔）"><Input /></Form.Item>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>全量采集仍需先在来源配置中明确开启，避免定时任务意外拉取全部数据。</Typography.Paragraph>
      </Form>
    </Modal>
  </PageContainer>;
}