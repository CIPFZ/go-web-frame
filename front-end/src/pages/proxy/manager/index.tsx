import React, { useEffect, useState } from 'react';
import { useIntl } from '@umijs/max';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Col, Descriptions, Input, List, Modal, Row, Space, Statistic, Tag, Typography, message } from 'antd';
import { CheckCircleOutlined, CodeOutlined, PauseCircleOutlined, PlayCircleOutlined, ReloadOutlined, RollbackOutlined, SaveOutlined } from '@ant-design/icons';
import {
  actionProxy, listProxyInstances, readProxyConfig, readProxyMetrics, rollbackProxyConfig, saveProxyConfig, validateProxyConfig,
  type ProxyInstance, type ProxyMetrics,
} from '@/services/proxy';

const stateColor: Record<string, string> = { active: 'green', running: 'green', inactive: 'default', failed: 'red', unavailable: 'orange', unknown: 'gold' };

export default function ProxyManagerPage() {
  const intl = useIntl();
  const [items, setItems] = useState<ProxyInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<ProxyInstance>();
  const [config, setConfig] = useState('');
  const [configDigest, setConfigDigest] = useState('');
  const [metrics, setMetrics] = useState<Record<number, ProxyMetrics>>({});
  const [saving, setSaving] = useState(false);

  const t = (id: string) => intl.formatMessage({ id });
  const refresh = async () => {
    setLoading(true);
    try {
      const res = await listProxyInstances();
      if (res.code === 0) setItems(res.data || []);
      else message.error(res.msg || t('proxy.loadFailed'));
    } catch { message.error(t('proxy.loadFailed')); }
    finally { setLoading(false); }
  };
  useEffect(() => { refresh(); }, []);
  useEffect(() => {
    const poll = async () => {
      await Promise.all(items.map(async item => {
        try {
          const res = await readProxyMetrics(item.id);
          if (res.code === 0) setMetrics(prev => ({ ...prev, [item.id]: res.data }));
        } catch { /* status remains visible even when metrics are unavailable */ }
      }));
    };
    if (items.length) poll();
    const timer = window.setInterval(poll, 5000);
    return () => window.clearInterval(timer);
  }, [items]);

  const runAction = async (item: ProxyInstance, action: string) => {
    try {
      const res = await actionProxy(item.id, action);
      if (res.code !== 0) message.error(res.msg || t('proxy.actionFailed'));
      else { message.success(t('proxy.actionDone')); refresh(); }
    } catch { message.error(t('proxy.actionFailed')); }
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

  return <PageContainer title={t('proxy.title')} extra={<Button icon={<ReloadOutlined />} onClick={refresh}>{t('proxy.refresh')}</Button>}>
    <Alert showIcon type="info" message={t('proxy.notice')} style={{ marginBottom: 16 }} />
    <List loading={loading} grid={{ gutter: 16, xs: 1, md: 2 }} dataSource={items} renderItem={item => {
      const m = metrics[item.id];
      const state = item.status?.state || 'unknown';
      return <List.Item><Card title={<Space><Typography.Text strong>{item.name}</Typography.Text><Tag color={stateColor[state]}>{state}</Tag></Space>} extra={<Tag>{item.engine}</Tag>}>
        <Descriptions size="small" column={1}>
          <Descriptions.Item label={t('proxy.unit')}>{item.unit}</Descriptions.Item>
          <Descriptions.Item label={t('proxy.configPath')}>{item.configPath}</Descriptions.Item>
          <Descriptions.Item label="PID">{item.status?.mainPid || '-'}</Descriptions.Item>
        </Descriptions>
        {m && <Row gutter={12} style={{ margin: '12px 0' }}>
          <Col span={8}><Statistic title={t('proxy.connections')} value={m.connections} /></Col>
          <Col span={8}><Statistic title={t('proxy.read')} value={m.readBytes} suffix="B" /></Col>
          <Col span={8}><Statistic title={t('proxy.write')} value={m.writeBytes} suffix="B" /></Col>
        </Row>}
        {item.status?.error && <Alert type="warning" showIcon message={item.status.error} style={{ marginBottom: 12 }} />}
        <Space wrap>
          <Button icon={<PlayCircleOutlined />} onClick={() => runAction(item, 'start')}>{t('proxy.start')}</Button>
          <Button icon={<PauseCircleOutlined />} onClick={() => runAction(item, 'stop')}>{t('proxy.stop')}</Button>
          <Button icon={<ReloadOutlined />} onClick={() => runAction(item, 'restart')}>{t('proxy.restart')}</Button>
          <Button icon={<CodeOutlined />} onClick={() => openConfig(item)}>{t('proxy.config')}</Button>
        </Space>
      </Card></List.Item>;
    }} />
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
