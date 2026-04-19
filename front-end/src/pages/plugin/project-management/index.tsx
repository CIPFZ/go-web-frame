import React, { useMemo, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  ModalForm,
  ProCard,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { getLocale, history } from '@umijs/max';
import { EditOutlined, EyeOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Form, Space, Statistic, Typography, message } from 'antd';

import {
  createPlugin,
  getDepartmentList,
  getPluginList,
  type DepartmentItem,
  type PluginItem,
  updatePlugin,
} from '@/services/api/plugin';
import { getDisplayDescription, getDisplayName, isEnglishLocale } from '@/utils/plugin';

type ProjectFormValues = {
  code: string;
  repositoryUrl: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  departmentId: number;
};

const copyMap = {
  zh: {
    title: '插件项目管理',
    subtitle: '维护插件主数据，并从项目详情继续完成发布、重提、下架与流程追踪。',
    listTitle: '项目列表',
    create: '新建项目',
    edit: '编辑项目',
    detail: '查看详情',
    code: '插件编码',
    name: '插件名称',
    department: '归属部门',
    repository: '仓库地址',
    description: '插件描述',
    total: '项目总数',
    actions: '操作',
    saveFailed: '保存项目失败',
    saveSuccess: '项目已保存',
  },
  en: {
    title: 'Plugin Project Management',
    subtitle:
      'Maintain plugin metadata and continue release, review, and offline workflows from project detail.',
    listTitle: 'Projects',
    create: 'New Project',
    edit: 'Edit Project',
    detail: 'View Detail',
    code: 'Code',
    name: 'Name',
    department: 'Department',
    repository: 'Repository',
    description: 'Description',
    total: 'Projects',
    actions: 'Actions',
    saveFailed: 'Failed to save project',
    saveSuccess: 'Project saved',
  },
};

const PluginProjectManagementPage: React.FC = () => {
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const actionRef = useRef<ActionType>(null);
  const [form] = Form.useForm<ProjectFormValues>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<PluginItem>();
  const [departments, setDepartments] = useState<DepartmentItem[]>([]);
  const [summaryTotal, setSummaryTotal] = useState(0);

  const departmentOptions = useMemo(
    () => departments.map((item) => ({ label: `${item.productLine} / ${item.name}`, value: item.ID })),
    [departments],
  );

  const loadDepartments = async () => {
    if (departments.length) return;
    const res = await getDepartmentList({ page: 1, pageSize: 999 });
    if (res.code === 0) {
      setDepartments(res.data?.list || []);
    }
  };

  const openCreateModal = async () => {
    setEditing(undefined);
    form.resetFields();
    await loadDepartments();
    setModalOpen(true);
  };

  const openEditModal = async (record: PluginItem) => {
    setEditing(record);
    await loadDepartments();
    form.setFieldsValue({
      code: record.code,
      repositoryUrl: record.repositoryUrl,
      nameZh: record.nameZh,
      nameEn: record.nameEn,
      descriptionZh: record.descriptionZh,
      descriptionEn: record.descriptionEn,
      departmentId: record.departmentId,
    });
    setModalOpen(true);
  };

  const handleSubmit = async (values: ProjectFormValues) => {
    const payload = editing?.ID ? { id: editing.ID, ...values, ownerId: editing.ownerId } : values;
    const res = editing?.ID ? await updatePlugin(payload) : await createPlugin(payload);
    if (res.code !== 0) {
      message.error(res.msg || copy.saveFailed);
      return false;
    }
    message.success(copy.saveSuccess);
    setModalOpen(false);
    actionRef.current?.reload();
    return true;
  };

  const columns: ProColumns<PluginItem>[] = [
    {
      title: copy.code,
      dataIndex: 'code',
      width: 160,
    },
    {
      title: copy.name,
      dataIndex: 'nameZh',
      width: 240,
      render: (_, record) => (
        <Space direction="vertical" size={2}>
          <Typography.Text strong>{getDisplayName(locale, record)}</Typography.Text>
          <Typography.Text type="secondary">{record.code}</Typography.Text>
        </Space>
      ),
    },
    {
      title: copy.department,
      dataIndex: 'department',
      width: 180,
      search: false,
      render: (_, record) => record.department || '-',
    },
    {
      title: copy.repository,
      dataIndex: 'repositoryUrl',
      search: false,
      ellipsis: true,
      render: (_, record) => (
        <Typography.Link href={record.repositoryUrl} target="_blank">
          {record.repositoryUrl}
        </Typography.Link>
      ),
    },
    {
      title: copy.description,
      dataIndex: 'descriptionZh',
      search: false,
      ellipsis: true,
      render: (_, record) => getDisplayDescription(locale, record),
    },
    {
      title: copy.actions,
      dataIndex: 'option',
      valueType: 'option',
      width: 180,
      render: (_, record) => [
        <a key="detail" onClick={() => history.push(`/plugin/project/${record.ID}`)}>
          <EyeOutlined /> {copy.detail}
        </a>,
        <a key="edit" onClick={() => void openEditModal(record)}>
          <EditOutlined /> {copy.edit}
        </a>,
      ],
    },
  ];

  return (
    <PageContainer title={false}>
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <ProCard
          bordered
          style={{
            borderRadius: 24,
            overflow: 'hidden',
            background: 'linear-gradient(135deg, #f8fafc 0%, #eef4ff 100%)',
          }}
        >
          <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <Space direction="vertical" size={6}>
              <Typography.Title level={3} style={{ margin: 0 }}>
                {copy.title}
              </Typography.Title>
              <Typography.Paragraph type="secondary" style={{ margin: 0, maxWidth: 760 }}>
                {copy.subtitle}
              </Typography.Paragraph>
            </Space>
            <Statistic title={copy.total} value={summaryTotal} />
          </Space>
        </ProCard>

        <ProTable<PluginItem>
          actionRef={actionRef}
          rowKey="ID"
          headerTitle={copy.listTitle}
          search={{ labelWidth: 'auto' }}
          columns={columns}
          request={async (params) => {
            const res = await getPluginList({
              code: params.code,
              name: params.nameZh || params.name,
              page: params.current,
              pageSize: params.pageSize,
            });
            setSummaryTotal(res.data?.total || 0);
            return {
              data: res.data?.list || [],
              success: res.code === 0,
              total: res.data?.total || 0,
            };
          }}
          toolBarRender={() => [
            <Button key="new" type="primary" onClick={() => void openCreateModal()}>
              <PlusOutlined /> {copy.create}
            </Button>,
          ]}
        />
      </Space>

      <ModalForm<ProjectFormValues>
        form={form}
        title={editing?.ID ? copy.edit : copy.create}
        open={modalOpen}
        modalProps={{ destroyOnClose: true }}
        onOpenChange={setModalOpen}
        onFinish={handleSubmit}
      >
        <ProFormText
          name="code"
          label={copy.code}
          disabled={Boolean(editing?.ID)}
          rules={[{ required: true }]}
        />
        <ProFormText name="repositoryUrl" label={copy.repository} rules={[{ required: true }]} />
        <ProFormText name="nameZh" label="中文名称" rules={[{ required: true }]} />
        <ProFormText name="nameEn" label="English Name" rules={[{ required: true }]} />
        <ProFormSelect
          name="departmentId"
          label={copy.department}
          options={departmentOptions}
          rules={[{ required: true }]}
        />
        <ProFormTextArea name="descriptionZh" label="中文描述" rules={[{ required: true }]} />
        <ProFormTextArea name="descriptionEn" label="English Description" rules={[{ required: true }]} />
      </ModalForm>
    </PageContainer>
  );
};

export default PluginProjectManagementPage;
