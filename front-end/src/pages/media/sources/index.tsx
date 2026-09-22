import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag, message } from 'antd';
import { CheckCircleOutlined, CloudDownloadOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  checkMediaSource, collectMediaSource, createMediaSource, listMediaSources, updateMediaSource,
  type MediaSource, type MediaSourceInput,
} from '@/services/media';

const emptySource: MediaSourceInput = {
  name: '', apiUrl: '', apiType: 'json', enabled: true, priority: 100,
  allowCategories: '', denyCategories: '', requestIntervalSeconds: 3, timeoutSeconds: 15,
};

export default function MediaSourcesPage() {
  const [items, setItems] = useState<MediaSource[]>([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<MediaSource | undefined>();
  const [editorOpen, setEditorOpen] = useState(false);
  const [form] = Form.useForm<MediaSourceInput>();
  const refresh = async () => {
    setLoading(true);
    try {
      const result = await listMediaSources();
      if (result.code === 0) setItems(result.data || []);
      else message.error(result.msg || '获取资源站失败');
    } catch { message.error('获取资源站失败'); }
    finally { setLoading(false); }
  };

  useEffect(() => { refresh(); }, []);

  const openEditor = (item?: MediaSource) => {
    setEditing(item);
    setEditorOpen(true);
    form.setFieldsValue(item ? {
      name: item.name, apiUrl: item.apiUrl, apiType: item.apiType, enabled: item.enabled,
      priority: item.priority, allowCategories: item.allowCategories || '', denyCategories: item.denyCategories || '',
      requestIntervalSeconds: item.requestIntervalSeconds, timeoutSeconds: item.timeoutSeconds,
    } : emptySource);
  };

  const save = async () => {
    const values = await form.validateFields();
    try {
      const result = editing ? await updateMediaSource(editing.id, values) : await createMediaSource(values);
      if (result.code !== 0) { message.error(result.msg || '保存失败'); return; }
      message.success('来源已保存');
      setEditing(undefined);
      setEditorOpen(false);
      await refresh();
    } catch { message.error('保存失败'); }
  };

  const health = async (item: MediaSource) => {
    try {
      const result = await checkMediaSource(item.id);
      if (result.code === 0 && result.data.result === 'ok') message.success('健康检查通过');
      else message.warning(result.data?.errorMessage || result.msg || '来源不可用');
      await refresh();
    } catch { message.error('健康检查失败'); }
  };

  const collect = async (item: MediaSource) => {
    try {
      const result = await collectMediaSource(item.id, 'sample', 100);
      if (result.code === 0) {
        const data = result.data;
        message.success(`任务已排队：#${data.id}（后台执行）`);
      } else message.error(result.msg || '采集失败');
    } catch { message.error('采集失败'); }
  };

  return <PageContainer
    title="影视资源采集"
    extra={<Space><Button icon={<ReloadOutlined />} onClick={refresh}>刷新</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>新增来源</Button></Space>}
  >
    <Table<MediaSource>
      rowKey="id"
      loading={loading}
      dataSource={items}
      columns={[
        { title: '名称', dataIndex: 'name' },
        { title: '接口', dataIndex: 'apiUrl', ellipsis: true },
        { title: '优先级', dataIndex: 'priority', width: 90, sorter: (a, b) => a.priority - b.priority },
        { title: '状态', dataIndex: 'enabled', width: 90, render: (value: boolean) => <Tag color={value ? 'green' : 'default'}>{value ? '启用' : '停用'}</Tag> },
        { title: '最近成功', dataIndex: 'lastOkAt', width: 180, render: (value?: string) => value || '-' },
        { title: '操作', width: 270, render: (_: unknown, item: MediaSource) => <Space>
          <Button size="small" icon={<CheckCircleOutlined />} onClick={() => health(item)}>检查</Button>
          <Button size="small" icon={<CloudDownloadOutlined />} onClick={() => collect(item)} disabled={!item.enabled}>采集100条</Button>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditor(item)}>编辑</Button>
        </Space> },
      ]}
    />
    <Modal open={editorOpen} title={editing ? '编辑来源' : '新增来源'} onCancel={() => { setEditing(undefined); setEditorOpen(false); }} onOk={save} destroyOnHidden>
      <Form form={form} layout="vertical" initialValues={emptySource}>
        <Form.Item name="name" label="来源名称" rules={[{ required: true }]}><Input maxLength={100} /></Form.Item>
        <Form.Item name="apiUrl" label="HTTPS 接口地址" rules={[{ required: true }, { type: 'url' }]}><Input placeholder="https://example.com/api.php/provide/vod/" /></Form.Item>
        <Space style={{ width: '100%' }} align="start">
          <Form.Item name="apiType" label="格式"><Input disabled value="json" /></Form.Item>
          <Form.Item name="priority" label="优先级"><InputNumber min={1} max={10000} /></Form.Item>
          <Form.Item name="timeoutSeconds" label="超时(秒)"><InputNumber min={1} max={120} /></Form.Item>
        </Space>
        <Form.Item name="enabled" label="启用" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="allowCategories" label="分类白名单（逗号分隔）"><Input placeholder="电影,电视剧,动漫,综艺" /></Form.Item>
        <Form.Item name="denyCategories" label="分类黑名单（逗号分隔）"><Input /></Form.Item>
      </Form>
    </Modal>
  </PageContainer>;
}
