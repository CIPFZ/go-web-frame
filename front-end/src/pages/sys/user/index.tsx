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
import { PlusOutlined, UserOutlined, KeyOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';

// 导入 API
import { getUserList, addUser, updateUser, deleteUser, resetPassword } from '@/services/api/user';
import { getAuthorityList } from '@/services/api/authority';

type AuthorityTreeNode = {
  title: string;
  value: number;
  key: number;
  children: AuthorityTreeNode[];
};


const UserTableList: React.FC = () => {
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
      const res = await deleteUser({ id });
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
      // switch true -> 1 (正常), false -> 2 (冻结)
      status: values.status ? 1 : 2,
      // 确保 authorityIds 是数组
      authorityIds: values.authorityIds || [],
    };

    try {
      const res = await method(reqData);
      if (res.code === 0) {
        message.success(isUpdate ? '更新成功' : '创建成功');
        setModalVisible(false);
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

  // 提交重置密码
  const handleResetPwdFinish = async (values: any) => {
    if (!pwdCurrentRow?.ID) return false;
    try {
      const res = await resetPassword({
        id: pwdCurrentRow.ID,
        password: values.password
      });
      if (res.code === 0) {
        message.success('密码重置成功');
        setPwdModalVisible(false);
        return true;
      }
      message.error(res.msg || '重置失败');
      return false;
    } catch (error) {
      message.error('请求出错');
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
      title: '头像',
      dataIndex: 'avatar',
      width: 60,
      search: false,
      align: 'center',
      render: (_, record) => (
        <Avatar src={record.avatar} icon={<UserOutlined />} />
      ),
    },
    {
      title: '用户名',
      dataIndex: 'username',
      copyable: true,
      width: 120,
    },
    {
      title: '昵称',
      dataIndex: 'nickName',
      width: 120,
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      width: 120,
    },
    {
      title: '用户角色',
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
              <Tag key={auth.authorityId} color={isCurrent ? "blue" : "default"}>
                {auth.authorityName}
              </Tag>
            );
          })}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100, // 稍微调宽一点以容纳标签
      align: 'center',
      // valueEnum 用于搜索栏的下拉筛选，必须保留
      valueEnum: {
        1: { text: '正常', status: 'Success' },
        2: { text: '冻结', status: 'Error' },
      },
      // ✨ 关键修改：使用 render 自定义渲染为 Tag
      render: (_, record) => {
        // 定义状态映射
        const statusMap = {
          1: { color: 'success', text: '正常' }, // 绿色胶囊
          2: { color: 'error', text: '冻结' },   // 红色胶囊
        };

        const current = statusMap[record.status] || { color: 'default', text: '未知' };

        return (
          <Tag color={current.color} style={{ minWidth: 60, textAlign: 'center' }}>
            {current.text}
          </Tag>
        );
      },
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      hideInTable: true, // 列表隐藏，搜索显示
    },
    {
      title: '操作',
      valueType: 'option',
      width: 220,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a onClick={() => handleEdit(record)}>
            <EditOutlined /> 编辑
          </a>
          <a onClick={() => handleResetPwdClick(record)}>
            <KeyOutlined /> 重置密码
          </a>
          <Popconfirm
            title="确定删除此用户?"
            description="删除后无法恢复"
            onConfirm={() => handleDelete(record.ID!)}
            okText="确定"
            cancelText="取消"
          >
            <a style={{ color: '#ff4d4f' }}>
              <DeleteOutlined /> 删除
            </a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<API.UserInfo>
        headerTitle="用户管理"
        actionRef={actionRef}
        rowKey="ID"
        search={{ labelWidth: 'auto' }}

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
        scroll={{ x: 1000 }}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={handleAdd}>
            <PlusOutlined /> 新建用户
          </Button>,
        ]}
      />

      {/* --- 1. 用户信息表单 (新增/编辑) --- */}
      <ModalForm
        title={currentRow ? '编辑用户' : '新建用户'}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        modalProps={{ destroyOnClose: true }}
        // 回显数据
        initialValues={currentRow ? {
          ...currentRow,
          // 将后端 1/2 转换为 switch 的 true/false
          status: currentRow.status === 1,
          // 回显多角色 (提取 ID 数组)
          authorityIds: currentRow.authorities?.map(a => a.authorityId)
        } : {
          status: true, // 默认启用
          authorityIds: []
        }}
      >
        <ProFormText
          name="username"
          label="用户名"
          placeholder="登录账号"
          disabled={!!currentRow} // 编辑时不可改用户名
          rules={[{ required: true, message: '请输入用户名' }]}
        />

        {/* 只有新增时才显示密码输入框 */}
        {!currentRow && (
          <ProFormText.Password
            name="password"
            label="初始密码"
            placeholder="请输入密码"
            rules={[{ required: true, message: '请输入密码' }]}
          />
        )}

        {/* ✨ 使用 ProFormGroup 并将宽度改为 sm */}
        <ProFormGroup>
          <ProFormText
            name="nickName"
            label="昵称"
            placeholder="显示名称"
            rules={[{ required: true, message: '请输入昵称' }]}
            width="sm"
          />
          <ProFormText
            name="phone"
            label="手机号"
            width="sm"
          />
        </ProFormGroup>

        {/* ✅ 新增：统一的角色选择框 */}
        <ProFormTreeSelect
          name="authorityIds"
          label="角色分配"
          placeholder="请选择用户拥有的角色 (可多选)"
          // 直接使用 fetchRoleData (我们之前写的稳健函数)
          request={async () => {
            const res = await getAuthorityList({ page: 1, pageSize: 9999 });
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
            multiple: true, // 多选
            treeDefaultExpandAll: true,
            showSearch: true,
            treeNodeFilterProp: 'title',
          }}
          // 必填校验：至少选一个
          rules={[{ required: true, message: '请至少分配一个角色' }]}
        />

        <ProFormText
          name="email"
          label="邮箱"
          rules={[{ type: 'email', message: '邮箱格式不正确' }]}
        />

        <ProFormSwitch
          name="status"
          label="用户状态"
          checkedChildren="正常"
          unCheckedChildren="冻结"
        />
      </ModalForm>

      {/* --- 2. 重置密码模态框 --- */}
      <ModalForm
        title={`重置密码 - ${pwdCurrentRow?.username}`}
        width="400px"
        open={pwdModalVisible}
        onOpenChange={setPwdModalVisible}
        onFinish={handleResetPwdFinish}
        modalProps={{ destroyOnClose: true }}
      >
        <ProFormText.Password
          name="password"
          label="新密码"
          rules={[{ required: true, message: '请输入新密码' }]}
        />
      </ModalForm>

    </PageContainer>
  );
};

export default UserTableList;
