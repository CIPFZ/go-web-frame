import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormDigit,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Space, message, Popconfirm } from 'antd';
import { PlusOutlined, SettingOutlined, CopyOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';

import { getAuthorityList, createAuthority, updateAuthority, deleteAuthority } from '@/services/api/authority';
// ✨ 导入尚未创建的 PermissionDrawer
import PermissionDrawer from './components/PermissionDrawer';

export type AuthorityItem = {
  authorityId: number;
  authorityName: string;
  parentId: number;
  defaultRouter: string;
  children: AuthorityItem[] | null;
};

const AuthorityTableList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<AuthorityItem>();
  const [currentParentId, setCurrentParentId] = useState<number>(0);

  // 权限抽屉状态
  const [isDrawerOpen, setIsDrawerOpen] = useState<boolean>(false);
  const [permissionRow, setPermissionRow] = useState<AuthorityItem>();

  // --- CRUD 操作 ---
  const handleAddNew = () => {
    setCurrentRow(undefined);
    setCurrentParentId(0);
    setIsModalOpen(true);
  };

  const handleAddChild = (record: AuthorityItem) => {
    setCurrentRow(undefined);
    setCurrentParentId(record.authorityId);
    setIsModalOpen(true);
  };

  const handleEdit = (record: AuthorityItem) => {
    setCurrentRow(record);
    setCurrentParentId(record.parentId);
    setIsModalOpen(true);
  };

  const handleCopy = (record: AuthorityItem) => {
    setCurrentRow({ ...record, authorityId: undefined } as any); // 清空ID
    setCurrentParentId(record.parentId);
    setIsModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const res = await deleteAuthority({ id });
      if (res.code === 0) {
        message.success('删除成功');
        actionRef.current?.reload();
      } else {
        message.error(res.msg || '删除失败');
      }
    } catch (error) {
      message.error('请求出错');
    }
  };

  const handleModalFinish = async (values: any) => {
    const isUpdate = !!currentRow?.authorityId;
    const method = isUpdate ? updateAuthority : createAuthority;
    const data = {
      ...values,
      parentId: currentParentId,
      // 如果是更新，需要传原来的 authorityId
      authorityId: isUpdate ? currentRow?.authorityId : values.authorityId,
    };

    try {
      const res = await method(data);
      if (res.code === 0) {
        message.success('操作成功');
        setIsModalOpen(false);
        actionRef.current?.reload();
        return true;
      }
      message.error(res.msg || '操作失败');
      return false;
    } catch (error) {
      message.error('请求出错');
      return false;
    }
  };

  const columns: ProColumns<AuthorityItem>[] = [
    { title: '角色ID', dataIndex: 'authorityId', width: 100, fixed: 'left' },
    { title: '角色名称', dataIndex: 'authorityName', width: 200 },
    { title: '默认路由', dataIndex: 'defaultRouter' },
    {
      title: '操作',
      valueType: 'option',
      width: 350,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a onClick={() => { setPermissionRow(record); setIsDrawerOpen(true); }}>
            <SettingOutlined /> 设置权限
          </a>
          <a onClick={() => handleAddChild(record)}>
            <PlusOutlined />新增子角色
          </a>
          <a onClick={() => handleCopy(record)}>
            <CopyOutlined /> 拷贝
          </a>
          <a onClick={() => handleEdit(record)}>
            <EditOutlined />编辑
          </a>
          <Popconfirm
            title="确定删除?"
            onConfirm={() => handleDelete(record.authorityId)}
            cancelText="取消"
            okText="确定"
          >
            <a style={{ color: '#ff4d4f' }}><DeleteOutlined />删除</a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<AuthorityItem>
        headerTitle={false}
        actionRef={actionRef}
        rowKey="authorityId"
        search={false}
        pagination={false}
        expandable = {{ childrenColumnName: "children" }}
        scroll={{ x: 'max-content' }}
        request={async () => {
          const res = await getAuthorityList();
          return { data: res.data?.list || [], success: res.code === 0 };
        }}
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={handleAddNew}>
            <PlusOutlined /> 新增角色
          </Button>,
        ]}
      />

      <ModalForm
        title={currentRow?.authorityId ? '编辑角色' : '新增角色'}
        width="500px"
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        onFinish={handleModalFinish}
        modalProps={{ destroyOnClose: true }}
        initialValues={currentRow}
      >
        <ProFormDigit
          name="authorityId"
          label="角色ID"
          tooltip="必须是唯一的数字"
          placeholder="例如 888"
          disabled={!!currentRow?.authorityId} // 编辑时不可改
          rules={[{ required: true, message: '角色ID为必填项' }]}
        />
        <ProFormText
          name="authorityName"
          label="角色名称"
          placeholder="例如：管理员"
          rules={[{ required: true, message: '角色名称为必填项' }]}
        />
      </ModalForm>

      {/* 权限设置抽屉 */}
      {permissionRow && (
        <PermissionDrawer
          open={isDrawerOpen}
          role={permissionRow}
          onClose={() => { setIsDrawerOpen(false); setPermissionRow(undefined); }}
          onSuccess={() => {
            actionRef.current?.reload();
          }}
        />
      )}
    </PageContainer>
  );
};

export default AuthorityTableList;
