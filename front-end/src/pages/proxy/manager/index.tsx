import React, { useEffect, useState } from 'react';
import { useIntl } from '@umijs/max';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Col, Descriptions, Drawer, Empty, Input, List, Modal, Row, Space, Statistic, Tag, Typography, message } from 'antd';
import { CheckCircleOutlined, CodeOutlined, LineChartOutlined, PauseCircleOutlined, PlayCircleOutlined, ReloadOutlined, RollbackOutlined, SaveOutlined } from '@ant-design/icons';
import {
  actionProxy, listProxyInstances, readProxyConfig, readProxyMetrics, rollbackProxyConfig, saveProxyConfig, validateProxyConfig,
  type ProxyInstance, type ProxyMetrics,
} from '@/services/proxy';

const stateColor: Record<string, string> = { active: 'green', running: 'green', inactive: 'default', failed: 'red', unavailable: 'orange', unknown: 'gold' };
const HISTORY_LIMIT = 30;

type MetricSample = { at: number; metrics: ProxyMetrics };

const isRunning = (state?: string) => state === 'active' || state === 'running';

const formatBytes = (value?: number) => {
  const bytes = Math.max(0, Number(value || 0));
  if (bytes < 1024) return bytes.toFixed(0) + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB';
};

function MetricChart({ label, samples, color, value, formatValue }: {
  label: string;
  samples: MetricSample[];
  color: string;
  value: (sample: MetricSample) => number;
  formatValue: (value: number) => string;
}) {
  if (!samples.length) return <Typography.Text type="secondary">-</Typography.Text>;
  const values = samples.map(value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const flat = min === max;
  const padding = flat ? (max === 0 ? 1 : Math.max(Math.abs(max) * 0.05, 1)) : 0;
  const axisMin = flat ? Math.max(0, min - padding) : min;
  const axisMax = flat ? max + padding : max;
  const axisRange = axisMax - axisMin || 1;
  const width = 640;
  const height = 190;
  const left = 62;
  const right = 18;
  const top = 24;
  const bottom = 36;
  const plotWidth = width - left - right;
  const plotHeight = height - top - bottom;
  const xAt = (index: number) => left + (samples.length === 1 ? plotWidth / 2 : (index / (samples.length - 1)) * plotWidth);
  const yAt = (item: number) => top + plotHeight - ((item - axisMin) / axisRange) * plotHeight;
  const points = values.map((item, index) => xAt(index).toFixed(2) + ',' + yAt(item).toFixed(2)).join(' ');
  const latestIndex = values.length - 1;
  const latestX = xAt(latestIndex);
  const latestY = yAt(values[latestIndex]);
  const tickValues = flat ? [min] : Array.from({ length: 5 }, (_, index) => axisMax - (axisRange * index) / 4);
  const timeIndexes = Array.from(new Set([0, Math.floor(latestIndex / 2), latestIndex]));
  const formatTime = (valueAt: number) => new Date(valueAt).toLocaleTimeString('zh-CN', { hour12: false });
  return <div style={{ marginBottom: 22 }}>
    <Typography.Text strong>{label}</Typography.Text>
    <svg role="img" aria-label={label} viewBox={'0 0 ' + width + ' ' + height} style={{ display: 'block', width: '100%', height: 190, marginTop: 6, background: 'rgba(0,0,0,0.02)', borderRadius: 6 }}>
      <title>{label}</title>
      {tickValues.map((tick, index) => {
        const y = yAt(tick);
        return <g key={'y-' + index}>
          <line x1={left} y1={y} x2={width - right} y2={y} stroke="rgba(0,0,0,0.12)" strokeDasharray="3 3" />
          <text x={left - 8} y={y + 4} textAnchor="end" fontSize="10" fill="rgba(0,0,0,0.55)">{formatValue(tick)}</text>
        </g>;
      })}
      <line x1={left} y1={top} x2={left} y2={height - bottom} stroke="rgba(0,0,0,0.35)" />
      <line x1={left} y1={height - bottom} x2={width - right} y2={height - bottom} stroke="rgba(0,0,0,0.35)" />
      {timeIndexes.map(index => <text key={'x-' + index} x={xAt(index)} y={height - 12} textAnchor="middle" fontSize="10" fill="rgba(0,0,0,0.55)">{formatTime(samples[index].at)}</text>)}
      <polyline fill="none" stroke={color} strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" points={points} />
      <circle cx={latestX} cy={latestY} r="3.5" fill={color} />
      <text x={Math.min(latestX + 8, width - right - 42)} y={Math.max(latestY - 8, top + 10)} fontSize="11" fill={color}>{formatValue(values[latestIndex])}</text>
    </svg>
  </div>;
}

export default function ProxyManagerPage() {
  const intl = useIntl();
  const [items, setItems] = useState<ProxyInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<ProxyInstance>();
  const [monitoring, setMonitoring] = useState<ProxyInstance>();
  const [config, setConfig] = useState('');
  const [configDigest, setConfigDigest] = useState('');
  const [metrics, setMetrics] = useState<Record<number, ProxyMetrics>>({});
  const [metricHistory, setMetricHistory] = useState<Record<number, MetricSample[]>>({});
  const [actionLoading, setActionLoading] = useState<Record<number, string | undefined>>({});
  const [saving, setSaving] = useState(false);

  const t = (id: string) => intl.formatMessage({ id });
  const refresh = async (showLoading = true) => {
    if (showLoading) setLoading(true);
    try {
      const res = await listProxyInstances();
      if (res.code === 0) setItems(res.data || []);
      else message.error(res.msg || t('proxy.loadFailed'));
    } catch { message.error(t('proxy.loadFailed')); }
    finally { if (showLoading) setLoading(false); }
  };

  useEffect(() => {
    refresh();
    const timer = window.setInterval(async () => {
      try {
        const res = await listProxyInstances();
        if (res.code === 0) setItems(res.data || []);
      } catch { /* keep the last known service state */ }
    }, 5000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    let disposed = false;
    const poll = async () => {
      const results = await Promise.all(items.map(async item => {
        try {
          const res = await readProxyMetrics(item.id);
          return res.code === 0 ? { id: item.id, metrics: res.data } : undefined;
        } catch { return undefined; }
      }));
      if (disposed) return;
      const samples = results.filter((item): item is { id: number; metrics: ProxyMetrics } => !!item);
      if (!samples.length) return;
      const now = Date.now();
      setMetrics(prev => {
        const next = { ...prev };
        samples.forEach(item => { next[item.id] = item.metrics; });
        return next;
      });
      setMetricHistory(prev => {
        const next = { ...prev };
        samples.forEach(item => {
          const history = next[item.id] || [];
          const last = history[history.length - 1];
          if (!last || last.metrics.updatedAt !== item.metrics.updatedAt) {
            next[item.id] = [...history, { at: now, metrics: item.metrics }].slice(-HISTORY_LIMIT);
          }
        });
        return next;
      });
    };
    if (items.length) poll();
    const timer = window.setInterval(poll, 5000);
    return () => { disposed = true; window.clearInterval(timer); };
  }, [items]);

  const runAction = async (item: ProxyInstance, action: string) => {
    const running = isRunning(item.status?.state);
    if ((action === 'start' && running) || (action === 'stop' && !running) || (action === 'restart' && !running)) return;
    setActionLoading(prev => ({ ...prev, [item.id]: action }));
    try {
      const res = await actionProxy(item.id, action);
      if (res.code !== 0) message.error(res.msg || t('proxy.actionFailed'));
      else { message.success(t('proxy.actionDone')); await refresh(false); }
    } catch { message.error(t('proxy.actionFailed')); }
    finally { setActionLoading(prev => ({ ...prev, [item.id]: undefined })); }
  };

  const openConfig = async (item: ProxyInstance) => {
    try {
      const res = await readProxyConfig(item.id);
      if (res.code !== 0) { message.error(res.msg || t('proxy.configReadFailed')); return; }
      setEditing(item); setConfig(res.data.content); setConfigDigest(res.data.digest);
    } catch { message.error(t('proxy.configReadFailed')); }
  };

  const validate = async () => {
    if (!editing) return;
    try {
      const res = await validateProxyConfig(editing.id);
      if (res.code === 0) message.success(res.data.message || t('proxy.valid'));
      else message.error(res.msg || t('proxy.invalid'));
    } catch { message.error(t('proxy.invalid')); }
  };

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      const res = await saveProxyConfig(editing.id, config);
      if (res.code !== 0) { message.error(res.msg || t('proxy.saveFailed')); return; }
      setConfigDigest(res.data.digest);
      message.success(t('proxy.saved'));
      refresh();
    } catch { message.error(t('proxy.saveFailed')); }
    finally { setSaving(false); }
  };

  const rollback = async () => {
    if (!editing) return;
    setSaving(true);
    try {
      const res = await rollbackProxyConfig(editing.id);
      if (res.code !== 0) { message.error(res.msg || t('proxy.rollbackFailed')); return; }
      const current = await readProxyConfig(editing.id);
      if (current.code === 0) { setConfig(current.data.content); setConfigDigest(current.data.digest); }
      message.success(t('proxy.rolledBack'));
      refresh();
    } catch { message.error(t('proxy.rollbackFailed')); }
    finally { setSaving(false); }
  };

  const monitorItem = monitoring ? items.find(item => item.id === monitoring.id) || monitoring : undefined;
  const monitorHistory = monitorItem ? (metricHistory[monitorItem.id] || []) : [];
  const currentMetrics = monitorItem ? metrics[monitorItem.id] || (monitorHistory.length ? monitorHistory[monitorHistory.length - 1].metrics : undefined) : undefined;
  const previousMetrics = monitorHistory.length > 1 ? monitorHistory[monitorHistory.length - 2].metrics : undefined;
  const currentTime = currentMetrics ? Date.parse(currentMetrics.updatedAt) : 0;
  const previousTime = previousMetrics ? Date.parse(previousMetrics.updatedAt) : 0;
  const intervalSeconds = previousMetrics && currentMetrics && Number.isFinite(currentTime) && Number.isFinite(previousTime)
    ? Math.max((currentTime - previousTime) / 1000, 1) : 0;
  const readRate = intervalSeconds > 0 && previousMetrics && currentMetrics ? Math.max(0, (currentMetrics.readBytes - previousMetrics.readBytes) / intervalSeconds) : 0;
  const writeRate = intervalSeconds > 0 && previousMetrics && currentMetrics ? Math.max(0, (currentMetrics.writeBytes - previousMetrics.writeBytes) / intervalSeconds) : 0;

  return <PageContainer title={t('proxy.title')} extra={<Button icon={<ReloadOutlined />} onClick={() => refresh()}>{t('proxy.refresh')}</Button>}>
    <Alert showIcon type="info" message={t('proxy.notice')} style={{ marginBottom: 16 }} />
    <List loading={loading} grid={{ gutter: 16, xs: 1, md: 2 }} dataSource={items} renderItem={item => {
      const state = item.status?.state || 'unknown';
      const running = isRunning(state);
      const busy = !!actionLoading[item.id];
      return <List.Item><Card title={<Space><Typography.Text strong>{item.name}</Typography.Text><Tag color={stateColor[state] || 'gold'}>{state}</Tag></Space>} extra={<Tag>{item.engine}</Tag>}>
        <Descriptions size="small" column={1}>
          <Descriptions.Item label={t('proxy.unit')}>{item.unit}</Descriptions.Item>
          <Descriptions.Item label={t('proxy.configPath')}>{item.configPath}</Descriptions.Item>
          <Descriptions.Item label="PID">{item.status?.mainPid || '-'}</Descriptions.Item>
        </Descriptions>
        {item.status?.error && <Alert type="warning" showIcon message={item.status.error} style={{ marginBottom: 12 }} />}
        <Space wrap>
          <Button icon={<PlayCircleOutlined />} onClick={() => runAction(item, 'start')} disabled={running || busy} loading={actionLoading[item.id] === 'start'}>{t('proxy.start')}</Button>
          <Button icon={<PauseCircleOutlined />} onClick={() => runAction(item, 'stop')} disabled={!running || busy} loading={actionLoading[item.id] === 'stop'}>{t('proxy.stop')}</Button>
          <Button icon={<ReloadOutlined />} onClick={() => runAction(item, 'restart')} disabled={!running || busy} loading={actionLoading[item.id] === 'restart'}>{t('proxy.restart')}</Button>
          <Button icon={<LineChartOutlined />} onClick={() => setMonitoring(item)}>{t('proxy.monitor')}</Button>
          <Button icon={<CodeOutlined />} onClick={() => openConfig(item)}>{t('proxy.config')}</Button>
        </Space>
      </Card></List.Item>;
    }} />
    <Drawer open={!!monitoring} title={monitorItem ? t('proxy.monitor') + ' · ' + monitorItem.name : ''} width={820} onClose={() => setMonitoring(undefined)}>
      {monitorItem && <Space direction="vertical" size={18} style={{ width: '100%' }}>
        <Alert showIcon type="info" message={t('proxy.monitorNotice')} />
        <Descriptions size="small" column={2}>
          <Descriptions.Item label={t('proxy.unit')}>{monitorItem.unit}</Descriptions.Item>
          <Descriptions.Item label={t('proxy.status')}>{monitorItem.status?.state || 'unknown'}</Descriptions.Item>
        </Descriptions>
        {currentMetrics ? <Row gutter={[12, 12]}>
          <Col xs={24} sm={8}><Statistic title={t('proxy.connections')} value={currentMetrics.connections} /></Col>
          <Col xs={24} sm={8}><Statistic title={t('proxy.readRate')} value={formatBytes(readRate)} suffix="/s" /></Col>
          <Col xs={24} sm={8}><Statistic title={t('proxy.writeRate')} value={formatBytes(writeRate)} suffix="/s" /></Col>
        </Row> : <Empty description={t('proxy.monitorNoData')} />}
        {monitorHistory.length > 0 && <Card size="small" title={t('proxy.trends')}>
          <MetricChart label={t('proxy.connectionsTrend')} samples={monitorHistory} color="#1677ff" value={sample => sample.metrics.connections} formatValue={value => Math.round(value).toString()} />
          <MetricChart label={t('proxy.readTrend')} samples={monitorHistory} color="#13c2c2" value={sample => sample.metrics.readBytes} formatValue={formatBytes} />
          <MetricChart label={t('proxy.writeTrend')} samples={monitorHistory} color="#722ed1" value={sample => sample.metrics.writeBytes} formatValue={formatBytes} />
        </Card>}
      </Space>}
    </Drawer>
    <Modal open={!!editing} title={editing ? t('proxy.config') + ' · ' + editing.name : ''} width={900} onCancel={() => setEditing(undefined)} footer={<Space>
      <Button icon={<CheckCircleOutlined />} onClick={validate}>{t('proxy.validate')}</Button>
      <Button icon={<RollbackOutlined />} onClick={rollback} disabled={saving}>{t('proxy.rollback')}</Button>
      <Button type="primary" icon={<SaveOutlined />} onClick={save} loading={saving}>{t('proxy.save')}</Button>
    </Space>}>
      <Typography.Paragraph type="secondary">{t('proxy.digest')}: {configDigest}</Typography.Paragraph>
      <Input.TextArea value={config} onChange={e => setConfig(e.target.value)} autoSize={{ minRows: 18, maxRows: 36 }} spellCheck={false} />
    </Modal>
  </PageContainer>;
}
