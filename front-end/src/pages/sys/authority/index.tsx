import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
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
import {
  PlusOutlined,
  SettingOutlined,
  CopyOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import {
  getAuthorityList,
  createAuthority,
  updateAuthority,
  deleteAuthority,
} from '@/services/system/authority';
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
  const localeFormRef1 = useFormLocale();
  useI18n();
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
    setCurrentRow({
      ...record,
      authorityId: undefined,
    } as any); // 清空ID
    setCurrentParentId(record.parentId);
    setIsModalOpen(true);
  };
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteAuthority({
        id,
      });
      if (res.code === 0) {
        message.success(t('cms.deletedSuccessfully'));
        actionRef.current?.reload();
      } else {
        message.error(res.msg || t('cms.deleteFailed'));
      }
    } catch (error) {
      message.error(t('cms.requestFailed'));
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
        message.success(t('cms.operationCompleted'));
        setIsModalOpen(false);
        actionRef.current?.reload();
        return true;
      }
      message.error(res.msg || t('cms.operationFailed'));
      return false;
    } catch (error) {
      message.error(t('cms.requestFailed'));
      return false;
    }
  };
  const columns: ProColumns<AuthorityItem>[] = [
    {
      title: t('cms.roleId'),
      dataIndex: 'authorityId',
      width: 100,
      fixed: 'left',
    },
    {
      title: t('cms.roleName'),
      dataIndex: 'authorityName',
      width: 200,
    },
    {
      title: t('cms.defaultRoute'),
      dataIndex: 'defaultRouter',
    },
    {
      title: t('cms.actions'),
      valueType: 'option',
      width: 350,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a
            onClick={() => {
              setPermissionRow(record);
              setIsDrawerOpen(true);
            }}
          >
            <SettingOutlined />
            {t('cms.configurePermissions')}
          </a>
          <a onClick={() => handleAddChild(record)}>
            <PlusOutlined />
            {t('cms.addChildRole')}
          </a>
          <a onClick={() => handleCopy(record)}>
            <CopyOutlined />
            {t('cms.copy')}
          </a>
          <a onClick={() => handleEdit(record)}>
            <EditOutlined />
            {t('cms.edit')}
          </a>
          <Popconfirm
            title={t('cms.deleteThisItem')}
            onConfirm={() => handleDelete(record.authorityId)}
            cancelText={t('cms.cancel')}
            okText={t('cms.ok')}
          >
            <a
              style={{
                color: '#ff4d4f',
              }}
            >
              <DeleteOutlined />
              {t('cms.delete')}
            </a>
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
        expandable={{
          childrenColumnName: 'children',
        }}
        scroll={{
          x: 'max-content',
        }}
        request={async () => {
          const res = await getAuthorityList();
          return {
            data: res.data?.list || [],
            success: res.code === 0,
          };
        }}
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={handleAddNew}>
            <PlusOutlined />
            {t('cms.newRole')}
          </Button>,
        ]}
      />

      <ModalForm
        title={currentRow?.authorityId ? t('cms.editRole') : t('cms.newRole')}
        width="500px"
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        onFinish={handleModalFinish}
        modalProps={{
          destroyOnClose: true,
        }}
        initialValues={currentRow}
        formRef={localeFormRef1}
      >
        <ProFormDigit
          name="authorityId"
          label={t('cms.roleId')}
          tooltip={t('cms.mustBeAUniqueNumber')}
          placeholder={t('cms.eG888')}
          disabled={!!currentRow?.authorityId} // 编辑时不可改
          rules={[
            {
              required: true,
              message: t('cms.roleIdIsRequired'),
            },
          ]}
        />
        <ProFormText
          name="authorityName"
          label={t('cms.roleName')}
          placeholder={t('cms.eGAdministrator')}
          rules={[
            {
              required: true,
              message: t('cms.roleNameIsRequired'),
            },
          ]}
        />
      </ModalForm>

      {/* 权限设置抽屉 */}
      {permissionRow && (
        <PermissionDrawer
          open={isDrawerOpen}
          role={permissionRow}
          onClose={() => {
            setIsDrawerOpen(false);
            setPermissionRow(undefined);
          }}
          onSuccess={() => {
            actionRef.current?.reload();
          }}
        />
      )}
    </PageContainer>
  );
};
export default AuthorityTableList;
