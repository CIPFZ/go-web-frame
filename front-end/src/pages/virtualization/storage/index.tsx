import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tag,
} from 'antd';
import { PageContainer } from '@ant-design/pro-components';
import { useEffect, useState } from 'react';
import {
  createStorageVolume,
  deleteStorageVolume,
  getStoragePools,
  getStorageVolumes,
  resizeStorageVolume,
  type StoragePool,
  type StorageVolume,
} from '@/services/virtualizationStorage';

function formatBytes(value: number) {
  if (!value) return '0 B';
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) { size /= 1024; unit += 1; }
  return `${size.toFixed(size >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export default function VirtualizationStoragePage() {
  const [pools, setPools] = useState<StoragePool[]>([]);
  const [volumes, setVolumes] = useState<StorageVolume[]>([]);
  const [poolFilter, setPoolFilter] = useState<string>();
  const [createForm] = Form.useForm();
  const [resizeForm] = Form.useForm();
  const [createOpen, setCreateOpen] = useState(false);
  const [resizeVolume, setResizeVolume] = useState<StorageVolume>();

  const load = async (pool?: string) => {
    const [poolResult, volumeResult] = await Promise.all([getStoragePools(), getStorageVolumes(pool)]);
    if (poolResult.code === 0) setPools(poolResult.data || []);
    if (volumeResult.code === 0) setVolumes(volumeResult.data || []);
  };

  useEffect(() => { load(); }, []);

  const create = async () => {
    const values = await createForm.validateFields();
    const result = await createStorageVolume(values);
    if (result.code === 0) {
      message.success('磁盘创建成功');
      setCreateOpen(false);
      createForm.resetFields();
      load();
    }
  };

  const resize = async () => {
    if (!resizeVolume) return;
    const values = await resizeForm.validateFields();
    const result = await resizeStorageVolume(resizeVolume.pool, resizeVolume.name, values.sizeGiB);
    if (result.code === 0) {
      message.success('磁盘扩容成功');
      setResizeVolume(undefined);
      load();
    }
  };

  const poolColumns = [
    { title: '存储池', dataIndex: 'name', key: 'name' },
    { title: '状态', dataIndex: 'state', key: 'state', render: (state: string) => <Tag color={state === 'running' ? 'green' : 'default'}>{state}</Tag> },
    { title: '容量', dataIndex: 'capacityBytes', key: 'capacityBytes', render: (value: number) => formatBytes(value) },
    { title: '已分配', dataIndex: 'allocationBytes', key: 'allocationBytes', render: (value: number) => formatBytes(value) },
    { title: '可用', dataIndex: 'availableBytes', key: 'availableBytes', render: (value: number) => formatBytes(value) },
    { title: '磁盘卷', dataIndex: 'volumeCount', key: 'volumeCount' },
  ];

  return (
    <PageContainer title="存储管理">
      <Row gutter={[16, 16]}>
        {pools.map((pool) => (
          <Col xs={24} sm={12} lg={8} key={pool.name}>
            <Card size="small" hoverable onClick={() => { setPoolFilter(pool.name); load(pool.name); }}>
              <Statistic title={pool.name} value={formatBytes(pool.availableBytes)} suffix="可用" />
              <div style={{ marginTop: 8 }}>总容量：{formatBytes(pool.capacityBytes)}，磁盘卷：{pool.volumeCount}</div>
            </Card>
          </Col>
        ))}
      </Row>
      <Card
        title={poolFilter ? `磁盘卷（${poolFilter}）` : '全部磁盘卷'}
        style={{ marginTop: 16 }}
        extra={
          <Space>
            <Select
              allowClear
              placeholder="筛选存储池"
              value={poolFilter}
              style={{ width: 160 }}
              options={pools.map((pool) => ({ label: pool.name, value: pool.name }))}
              onChange={(value) => { setPoolFilter(value); load(value); }}
            />
            <Button icon={<PlusOutlined />} type="primary" onClick={() => { createForm.resetFields(); createForm.setFieldsValue({ pool: poolFilter || pools[0]?.name || 'default', format: 'qcow2' }); setCreateOpen(true); }}>创建磁盘</Button>
            <Button onClick={() => load()}>刷新</Button>
          </Space>
        }
      >
        <Table<StorageVolume>
          rowKey={(row) => `${row.pool}/${row.name}`}
          dataSource={volumes}
          pagination={{ pageSize: 10 }}
          columns={[
            { title: '名称', dataIndex: 'name', key: 'name' },
            { title: '存储池', dataIndex: 'pool', key: 'pool' },
            { title: '格式', dataIndex: 'format', key: 'format', render: (value: string) => <Tag>{value || '-'}</Tag> },
            { title: '容量', dataIndex: 'capacityBytes', key: 'capacityBytes', render: (value: number) => formatBytes(value) },
            { title: '实际占用', dataIndex: 'allocationBytes', key: 'allocationBytes', render: (value: number) => formatBytes(value) },
            { title: '路径', dataIndex: 'path', key: 'path', ellipsis: true },
            {
              title: '操作',
              key: 'action',
              render: (_: unknown, row: StorageVolume) => (
                <Space>
                  <a onClick={() => { resizeForm.setFieldsValue({ sizeGiB: Math.ceil(row.capacityBytes / 1024 / 1024 / 1024) }); setResizeVolume(row); }}><EditOutlined /> 扩容</a>
                  <Popconfirm title="确认删除此磁盘？如果仍被虚拟机使用，系统会拒绝删除。" onConfirm={async () => { const result = await deleteStorageVolume(row.pool, row.name); if (result.code === 0) { message.success('磁盘已删除'); load(); } }}>
                    <a><DeleteOutlined /> 删除</a>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      <Card title="存储池详情" style={{ marginTop: 16 }}>
        <Table<StoragePool> rowKey="name" dataSource={pools} columns={poolColumns} pagination={false} />
      </Card>
      <Modal title="创建磁盘卷" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={create}>
        <Form form={createForm} layout="vertical">
          <Form.Item name="pool" label="存储池" rules={[{ required: true }]}><Select options={pools.map((pool) => ({ label: pool.name, value: pool.name }))} /></Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, pattern: /^[A-Za-z0-9._-]+$/, message: '仅支持字母、数字、点、下划线和短横线' }]}><Input placeholder="例如 data-01" /></Form.Item>
          <Form.Item name="sizeGiB" label="容量 GiB" rules={[{ required: true }]}><InputNumber min={1} max={4096} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="format" label="格式" rules={[{ required: true }]}><Select options={[{ label: 'qcow2', value: 'qcow2' }, { label: 'raw', value: 'raw' }]} /></Form.Item>
        </Form>
      </Modal>
      <Modal title={`扩容磁盘：${resizeVolume?.name || ''}`} open={Boolean(resizeVolume)} onCancel={() => setResizeVolume(undefined)} onOk={resize}>
        <Form form={resizeForm} layout="vertical">
          <Form.Item name="sizeGiB" label="新的容量 GiB" rules={[{ required: true }]}><InputNumber min={1} max={4096} style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
}

