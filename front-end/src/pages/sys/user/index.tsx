import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
import type { API } from '@/services/system/types';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormTreeSelect,
  ProFormSwitch,
  ProFormGroup,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Space, message, Popconfirm, Tag, Avatar, Divider } from 'antd';
import {
  PlusOutlined,
  UserOutlined,
  KeyOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';

// 导入 API
import {
  getUserList,
  addUser,
  updateUser,
  deleteUser,
  resetPassword,
} from '@/services/system/user';
import { getAuthorityList } from '@/services/system/authority';
type AuthorityTreeNode = {
  title: string;
  value: number;
  key: number;
  children: AuthorityTreeNode[];
};
const UserTableList: React.FC = () => {
  const localeFormRef2 = useFormLocale();
  const localeFormRef1 = useFormLocale();
  useI18n();
  const actionRef = useRef<ActionType>(null);

  // --- 状态管理 ---
  // 用户模态框
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.UserInfo>();

  // 重置密码模态框
  const [pwdModalVisible, setPwdModalVisible] = useState<boolean>(false);
  const [pwdCurrentRow, setPwdCurrentRow] = useState<API.UserInfo>();

  // --- 操作处理 ---

  const handleAdd = () => {
    setCurrentRow(undefined);
    setModalVisible(true);
  };
  const handleEdit = (record: API.UserInfo) => {
    setCurrentRow(record);
    setModalVisible(true);
  };
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteUser({
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

  // 打开重置密码框
  const handleResetPwdClick = (record: API.UserInfo) => {
    setPwdCurrentRow(record);
    setPwdModalVisible(true);
  };

  // 提交用户表单 (新增/编辑)
  const handleFinish = async (values: any) => {
    const isUpdate = !!currentRow?.ID;
    const method = isUpdate ? updateUser : addUser;

    // 数据转换
    const reqData = {
      ...values,
      id: currentRow?.ID,
      // switch true -> 1 (正常), false -> 0 (禁用)
      status: values.status ? 1 : 0,
      // 确保 authorityIds 是数组
      authorityIds: values.authorityIds || [],
    };
    try {
      const res = await method(reqData);
      if (res.code === 0) {
        message.success(
          isUpdate
            ? t('cms.updatedSuccessfully')
            : t('cms.createdSuccessfully'),
        );
        setModalVisible(false);
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

  // 提交重置密码
  const handleResetPwdFinish = async (values: any) => {
    if (!pwdCurrentRow?.ID) return false;
    try {
      const res = await resetPassword({
        id: pwdCurrentRow.ID,
        password: values.password,
      });
      if (res.code === 0) {
        message.success(t('cms.passwordResetSuccessfully'));
        setPwdModalVisible(false);
        return true;
      }
      message.error(res.msg || t('cms.resetFailed'));
      return false;
    } catch (error) {
      message.error(t('cms.requestFailed'));
      return false;
    }
  };

  // --- 列定义 ---
  const columns: ProColumns<API.UserInfo>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      width: 60,
      search: false,
      align: 'center',
    },
    {
      title: t('cms.avatar'),
      dataIndex: 'avatar',
      width: 60,
      search: false,
      align: 'center',
      render: (_, record) => (
        <Avatar src={record.avatar} icon={<UserOutlined />} />
      ),
    },
    {
      title: t('cms.username'),
      dataIndex: 'username',
      copyable: true,
      width: 120,
    },
    {
      title: t('cms.nickname'),
      dataIndex: 'nickName',
      width: 120,
    },
    {
      title: t('cms.phoneNumber.5a9cc5'),
      dataIndex: 'phone',
      width: 120,
    },
    {
      title: t('cms.userRoles'),
      dataIndex: 'authorityId',
      width: 200,
      search: false,
      render: (_, record) => (
        <Space wrap>
          {/* 遍历所有角色 */}
          {record.authorities?.map((auth) => {
            // 如果是当前角色，高亮显示
            const isCurrent = auth.authorityId === record.authorityId;
            return (
              <Tag
                key={auth.authorityId}
                color={isCurrent ? 'blue' : 'default'}
              >
                {auth.authorityName}
              </Tag>
            );
          })}
        </Space>
      ),
    },
    {
      title: t('cms.status'),
      dataIndex: 'status',
      width: 100,
      // 稍微调宽一点以容纳标签
      align: 'center',
      // valueEnum 用于搜索栏的下拉筛选，必须保留
      valueEnum: {
        1: {
          text: t('cms.apiToken.enabledLabel'),
          status: 'Success',
        },
        0: {
          text: t('cms.disable'),
          status: 'Error',
        },
      },
      // ✨ 关键修改：使用 render 自定义渲染为 Tag
      render: (_, record) => {
        // 定义状态映射
        const statusMap: Record<
          number,
          {
            color: string;
            text: string;
          }
        > = {
          1: {
            color: 'success',
            text: t('cms.apiToken.enabledLabel'),
          },
          // 绿色胶囊
          0: {
            color: 'error',
            text: t('cms.disable'),
          }, // 红色胶囊
        };
        const current = statusMap[record.status] || {
          color: 'default',
          text: t('cms.unknown'),
        };
        return (
          <Tag
            color={current.color}
            style={{
              minWidth: 60,
              textAlign: 'center',
            }}
          >
            {current.text}
          </Tag>
        );
      },
    },
    {
      title: t('cms.email'),
      dataIndex: 'email',
      hideInTable: true, // 列表隐藏，搜索显示
    },
    {
      title: t('cms.actions'),
      valueType: 'option',
      width: 220,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a onClick={() => handleEdit(record)}>
            <EditOutlined />
            {t('cms.edit')}
          </a>
          <a onClick={() => handleResetPwdClick(record)}>
            <KeyOutlined />
            {t('cms.resetPassword')}
          </a>
          <Popconfirm
            title={t('cms.deleteThisUser')}
            description={t('cms.thisActionCannotBeUndone')}
            onConfirm={() => handleDelete(record.ID!)}
            okText={t('cms.ok')}
            cancelText={t('cms.cancel')}
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
      <ProTable<API.UserInfo>
        headerTitle={t('cms.users')}
        actionRef={actionRef}
        rowKey="ID"
        search={{
          labelWidth: 'auto',
        }}
        request={async (params) => {
          const res = await getUserList({
            page: params.current,
            pageSize: params.pageSize,
            username: params.username,
            nickName: params.nickName,
            phone: params.phone,
            email: params.email,
          });
          return {
            data: res.data?.list || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        }}
        columns={columns}
        scroll={{
          x: 1000,
        }}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={handleAdd}>
            <PlusOutlined />
            {t('cms.newUser')}
          </Button>,
        ]}
      />

      {/* --- 1. 用户信息表单 (新增/编辑) --- */}
      <ModalForm
        title={currentRow ? t('cms.editUser') : t('cms.newUser')}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        modalProps={{
          destroyOnClose: true,
        }}
        // 回显数据
        initialValues={
          currentRow
            ? {
                ...currentRow,
                // 将后端 1/2 转换为 switch 的 true/false
                status: currentRow.status === 1,
                // 回显多角色 (提取 ID 数组)
                authorityIds: currentRow.authorities?.map((a) => a.authorityId),
              }
            : {
                status: true,
                // 默认启用
                authorityIds: [],
              }
        }
        formRef={localeFormRef1}
      >
        <ProFormText
          name="username"
          label={t('cms.username')}
          placeholder={t('cms.loginAccount')}
          disabled={!!currentRow} // 编辑时不可改用户名
          rules={[
            {
              required: true,
              message: t('cms.enterAUsername'),
            },
          ]}
        />

        {/* 只有新增时才显示密码输入框 */}
        {!currentRow && (
          <ProFormText.Password
            name="password"
            label={t('cms.initialPassword')}
            placeholder={t('cms.enterAPassword')}
            rules={[
              {
                required: true,
                message: t('cms.enterAPassword'),
              },
            ]}
          />
        )}

        {/* ✨ 使用 ProFormGroup 并将宽度改为 sm */}
        <ProFormGroup>
          <ProFormText
            name="nickName"
            label={t('cms.nickname')}
            placeholder={t('cms.displayName.75ae6a')}
            rules={[
              {
                required: true,
                message: t('cms.enterANickname'),
              },
            ]}
            width="sm"
          />
          <ProFormText
            name="phone"
            label={t('cms.phoneNumber.5a9cc5')}
            width="sm"
          />
        </ProFormGroup>

        {/* ✅ 新增：统一的角色选择框 */}
        <ProFormTreeSelect
          name="authorityIds"
          label={t('cms.roleAssignment')}
          placeholder={t('cms.selectOneOrMoreRoles')}
          // 直接使用 fetchRoleData (我们之前写的稳健函数)
          request={async () => {
            const res = await getAuthorityList({
              page: 1,
              pageSize: 9999,
            });
            const list = res.data?.list || [];
            const loop = (data: any[]): any[] =>
              data.map((item) => ({
                title: item.authorityName,
                value: item.authorityId,
                key: item.authorityId,
                children: item.children ? loop(item.children) : [],
              }));
            return loop(list);
          }}
          fieldProps={{
            multiple: true,
            // 多选
            treeDefaultExpandAll: true,
            showSearch: true,
            treeNodeFilterProp: 'title',
          }}
          // 必填校验：至少选一个
          rules={[
            {
              required: true,
              message: t('cms.assignAtLeastOneRole'),
            },
          ]}
        />

        <ProFormText
          name="email"
          label={t('cms.email')}
          rules={[
            {
              type: 'email',
              message: t('cms.enterAValidEmailAddress'),
            },
          ]}
        />

        <ProFormSwitch
          name="status"
          label={t('cms.userStatus')}
          checkedChildren={t('cms.apiToken.enabledLabel')}
          unCheckedChildren={t('cms.disable')}
        />
      </ModalForm>

      {/* --- 2. 重置密码模态框 --- */}
      <ModalForm
        title={t('cms.resetPassword.9434c7', {
          value0: pwdCurrentRow?.username,
        })}
        width="400px"
        open={pwdModalVisible}
        onOpenChange={setPwdModalVisible}
        onFinish={handleResetPwdFinish}
        modalProps={{
          destroyOnClose: true,
        }}
        formRef={localeFormRef2}
      >
        <ProFormText.Password
          name="password"
          label={t('cms.newPassword')}
          rules={[
            {
              required: true,
              message: t('cms.enterANewPassword'),
            },
          ]}
        />
      </ModalForm>
    </PageContainer>
  );
};
export default UserTableList;
