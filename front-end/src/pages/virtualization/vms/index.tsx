import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PoweroffOutlined,
  ReloadOutlined,
  StopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import {
  Button,
  Descriptions,
  Drawer,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Space,
  Switch,
  Tag,
} from 'antd';
import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useRef, useState } from 'react';
import {
  createVirtualMachine,
  deleteVirtualMachine,
  getVirtualMachine,
  getVirtualMachineList,
  updateVirtualMachine,
  virtualMachineAction,
  type VirtualMachine,
} from '@/services/virtualization';

const stateColors: Record<string, string> = { running: 'green', paused: 'orange', shutoff: 'default', crashed: 'red' };

export default function VirtualMachinesPage() {
  const actionRef = useRef<ActionType>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [createOpen, setCreateOpen] = useState(false);
  const [editVm, setEditVm] = useState<VirtualMachine>();
  const [detail, setDetail] = useState<VirtualMachine>();
  const [detailOpen, setDetailOpen] = useState(false);

  const refresh = () => actionRef.current?.reload();
  const run = async (name: string, action: 'start' | 'shutdown' | 'reboot' | 'force-stop') => {
    const result = await virtualMachineAction(name, action);
    if (result.code === 0) { message.success('操作成功'); refresh(); }
  };

  const columns: ProColumns<VirtualMachine>[] = [
    { title: '名称', dataIndex: 'name', ellipsis: true, width: 180 },
    { title: '状态', dataIndex: 'state', search: false, width: 100, render: (_, row) => <Tag color={stateColors[row.state] || 'blue'}>{row.state}</Tag> },
    { title: '内存', dataIndex: 'memoryMiB', search: false, width: 110, render: (_, row) => row.memoryMiB + ' MiB' },
    { title: 'CPU', dataIndex: 'vcpus', search: false, width: 80, render: (_, row) => row.vcpus + ' vCPU' },
    { title: '网络', dataIndex: 'network', search: false, width: 120 },
    { title: '自启动', dataIndex: 'autostart', search: false, width: 90, render: (_, row) => row.autostart ? <Tag color="green">是</Tag> : <Tag>否</Tag> },
    {
      title: '操作', valueType: 'option', width: 330,
      render: (_, row) => [
        <a key="detail" onClick={async () => { const result = await getVirtualMachine(row.name); if (result.data) { setDetail(result.data); setDetailOpen(true); } }}><EyeOutlined /> 详情</a>,
        <a key="edit" onClick={() => { editForm.setFieldsValue({ memoryMiB: row.memoryMiB, vcpus: row.vcpus, autostart: row.autostart }); setEditVm(row); }}><EditOutlined /> 编辑</a>,
        row.state === 'running'
          ? <Popconfirm key="shutdown" title="确认正常关机？" onConfirm={() => run(row.name, 'shutdown')}><a><StopOutlined /> 关机</a></Popconfirm>
          : <a key="start" onClick={() => run(row.name, 'start')}><PoweroffOutlined /> 启动</a>,
        row.state === 'running' && <a key="reboot" onClick={() => run(row.name, 'reboot')}><ReloadOutlined /> 重启</a>,
        row.state === 'running' && <Popconfirm key="force-stop" title="强制停止可能丢失未保存数据，确认继续？" onConfirm={() => run(row.name, 'force-stop')}><a><ThunderboltOutlined /> 强停</a></Popconfirm>,
        row.state !== 'running' && <Popconfirm key="delete" title="只删除虚拟机定义，不删除磁盘文件，确认继续？" onConfirm={async () => { const result = await deleteVirtualMachine(row.name); if (result.code === 0) { message.success('已删除'); refresh(); } }}><a><DeleteOutlined /> 删除</a></Popconfirm>,
      ].filter(Boolean) as React.ReactNode[],
    },
  ];

  return (
    <PageContainer title="虚拟机管理">
      <ProTable<VirtualMachine>
        headerTitle="QEMU / KVM"
        actionRef={actionRef}
        rowKey="uuid"
        search={false}
        request={async () => { const result = await getVirtualMachineList(); return { data: result.data || [], success: result.code === 0 }; }}
        columns={columns}
        options={{ density: true, fullScreen: true, reload: true, setting: true }}
        toolBarRender={() => [<Button key="create" type="primary" onClick={() => { createForm.resetFields(); setCreateOpen(true); }}>创建虚拟机</Button>]}
      />
      <Modal title="创建虚拟机" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={async () => {
        const values = await createForm.validateFields();
        const result = await createVirtualMachine(values);
        if (result.code === 0) { message.success('创建成功'); setCreateOpen(false); refresh(); }
      }} width={620}>
        <Form form={createForm} layout="vertical" initialValues={{ memoryMiB: 2048, vcpus: 2, diskFormat: 'qcow2', pool: 'default', network: 'default', autostart: false, start: false }}>
          <Space.Compact block>
            <Form.Item name="name" label="名称" rules={[{ required: true, pattern: /^[A-Za-z0-9._-]+$/, message: '仅支持字母、数字、点、下划线和短横线' }]} style={{ width: '50%' }}><Input placeholder="例如 vm-test" /></Form.Item>
            <Form.Item name="network" label="网络" style={{ width: '50%' }}><Input /></Form.Item>
          </Space.Compact>
          <Space.Compact block>
            <Form.Item name="memoryMiB" label="内存 MiB" rules={[{ required: true }]} style={{ width: '33%' }}><InputNumber min={128} max={1048576} style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="vcpus" label="vCPU" rules={[{ required: true }]} style={{ width: '33%' }}><InputNumber min={1} max={256} style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="diskSizeGiB" label="新建磁盘 GiB" style={{ width: '34%' }}><InputNumber min={1} max={1024} style={{ width: '100%' }} /></Form.Item>
          </Space.Compact>
          <Form.Item name="diskPath" label="已有磁盘路径"><Input placeholder="留空则可使用新建磁盘；路径需位于 libvirt 允许目录" /></Form.Item>
          <Space.Compact block>
            <Form.Item name="pool" label="存储池" style={{ width: '50%' }}><Input /></Form.Item>
            <Form.Item name="diskFormat" label="磁盘格式" style={{ width: '50%' }}><Input /></Form.Item>
          </Space.Compact>
          <Space>
            <Form.Item name="autostart" label="开机自启" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="start" label="创建后启动" valuePropName="checked"><Switch /></Form.Item>
          </Space>
        </Form>
      </Modal>
      <Modal title={'编辑虚拟机：' + (editVm?.name || '')} open={Boolean(editVm)} onCancel={() => setEditVm(undefined)} onOk={async () => {
        if (!editVm) return;
        const values = await editForm.validateFields();
        const result = await updateVirtualMachine(editVm.name, values);
        if (result.code === 0) { message.success('保存成功'); setEditVm(undefined); refresh(); }
      }}>
        <Form form={editForm} layout="vertical">
          <Form.Item name="memoryMiB" label="内存 MiB" rules={[{ required: true }]}><InputNumber min={128} max={1048576} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="vcpus" label="vCPU" rules={[{ required: true }]}><InputNumber min={1} max={256} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="autostart" label="开机自启" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Modal>
      <Drawer title={'虚拟机详情：' + (detail?.name || '')} width={720} open={detailOpen} onClose={() => setDetailOpen(false)}>
        {detail && <Descriptions column={1} bordered>
          <Descriptions.Item label="名称">{detail.name}</Descriptions.Item>
          <Descriptions.Item label="UUID">{detail.uuid}</Descriptions.Item>
          <Descriptions.Item label="状态">{detail.state}</Descriptions.Item>
          <Descriptions.Item label="内存">{detail.memoryMiB} MiB</Descriptions.Item>
          <Descriptions.Item label="vCPU">{detail.vcpus}</Descriptions.Item>
          <Descriptions.Item label="网络">{detail.network || '-'}</Descriptions.Item>
          <Descriptions.Item label="磁盘">{detail.diskPaths.length ? detail.diskPaths.join('\n') : '-'}</Descriptions.Item>
        </Descriptions>}
      </Drawer>
    </PageContainer>
  );
}
