import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, ProCard, ProFormSwitch, ProFormText, ProTable } from '@ant-design/pro-components';
import { getLocale } from '@umijs/max';
import { Button, Tag, Typography, message } from 'antd';
import { EditOutlined, PlusOutlined } from '@ant-design/icons';

import {
  createDepartment,
  getDepartmentList,
  updateDepartment,
  type DepartmentItem,
} from '@/services/api/plugin';
import { isEnglishLocale, pickLocaleText } from '@/utils/plugin';

type DepartmentFormValues = {
  nameZh: string;
  nameEn: string;
  productLine: string;
  status?: boolean;
};

const copyMap = {
  zh: {
    title: '部门管理',
    subtitle: '统一维护插件归属部门，供插件项目创建和详情展示复用。',
    tableTitle: '部门列表',
    create: '新增部门',
    edit: '编辑部门',
    departmentName: '部门名称',
    englishName: '英文名称',
    productLine: '产品线',
    status: '状态',
    enabled: '启用',
    disabled: '停用',
    actions: '操作',
    saveSuccess: '保存成功',
    saveFailed: '保存失败',
  },
  en: {
    title: 'Department Management',
    subtitle: 'Maintain plugin ownership departments used by project creation and project detail pages.',
    tableTitle: 'Departments',
    create: 'New Department',
    edit: 'Edit Department',
    departmentName: 'Department Name',
    englishName: 'English Name',
    productLine: 'Product Line',
    status: 'Status',
    enabled: 'Enabled',
    disabled: 'Disabled',
    actions: 'Actions',
    saveSuccess: 'Saved',
    saveFailed: 'Failed to save',
  },
};

const PluginMasterPage: React.FC = () => {
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const actionRef = useRef<ActionType>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingDepartment, setEditingDepartment] = useState<DepartmentItem>();

  const columns: ProColumns<DepartmentItem>[] = [
    {
      title: copy.departmentName,
      dataIndex: 'nameZh',
      render: (_, record) => (
        <Typography.Text strong>
          {pickLocaleText(locale, record.nameZh || record.name, record.nameEn)}
        </Typography.Text>
      ),
    },
    {
      title: copy.productLine,
      dataIndex: 'productLine',
    },
    {
      title: copy.status,
      dataIndex: 'status',
      width: 120,
      render: (_, record) => (
        <Tag color={record.status ? 'success' : 'default'}>
          {record.status ? copy.enabled : copy.disabled}
        </Tag>
      ),
    },
    {
      title: copy.actions,
      dataIndex: 'option',
      valueType: 'option',
      width: 140,
      render: (_, record) => [
        <a
          key="edit"
          onClick={() => {
            setEditingDepartment(record);
            setModalOpen(true);
          }}
        >
          <EditOutlined /> {copy.edit}
        </a>,
      ],
    },
  ];

  return (
    <PageContainer title={false}>
      <ProCard bordered style={{ borderRadius: 24, marginBottom: 20 }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          {copy.title}
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ margin: '6px 0 0', maxWidth: 760 }}>
          {copy.subtitle}
        </Typography.Paragraph>
      </ProCard>

      <ProTable<DepartmentItem>
        actionRef={actionRef}
        rowKey="ID"
        headerTitle={copy.tableTitle}
        search={false}
        columns={columns}
        request={async (params) => {
          const res = await getDepartmentList({
            page: params.current,
            pageSize: params.pageSize,
            includeInactive: true,
          });
          return {
            data: res.data?.list || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        }}
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            onClick={() => {
              setEditingDepartment(undefined);
              setModalOpen(true);
            }}
          >
            <PlusOutlined /> {copy.create}
          </Button>,
        ]}
      />

      <ModalForm<DepartmentFormValues>
        title={editingDepartment?.ID ? copy.edit : copy.create}
        open={modalOpen}
        modalProps={{ destroyOnHidden: true }}
        initialValues={{
          nameZh: editingDepartment?.nameZh || editingDepartment?.name || '',
          nameEn: editingDepartment?.nameEn || '',
          productLine: editingDepartment?.productLine || '',
          status: editingDepartment?.status ?? true,
        }}
        onOpenChange={(open) => {
          setModalOpen(open);
          if (!open) setEditingDepartment(undefined);
        }}
        onFinish={async (values) => {
          const payload = editingDepartment?.ID ? { id: editingDepartment.ID, ...values } : values;
          const res = editingDepartment?.ID ? await updateDepartment(payload) : await createDepartment(payload);
          if (res.code !== 0) {
            message.error(res.msg || copy.saveFailed);
            return false;
          }
          message.success(copy.saveSuccess);
          actionRef.current?.reload();
          return true;
        }}
      >
        <ProFormText name="nameZh" label={copy.departmentName} rules={[{ required: true }]} />
        <ProFormText name="nameEn" label={copy.englishName} rules={[{ required: true }]} />
        <ProFormText name="productLine" label={copy.productLine} rules={[{ required: true }]} />
        {editingDepartment?.ID ? <ProFormSwitch name="status" label={copy.status} /> : null}
      </ModalForm>
    </PageContainer>
  );
};

export default PluginMasterPage;
