import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormSelect,
  ProFormTextArea,
  ProFormGroup,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Space, message, Popconfirm, Tag } from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';

// 导入 API
import { getApiList, createApi, updateApi, deleteApi } from '@/services/api/api';

// 定义数据类型
type ApiItem = {
  ID: number;
  path: string;
  description: string;
  apiGroup: string;
  method: string;
};

const ApiTableList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);

  // 模态框状态
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<ApiItem>();
  // 选中的行 (用于批量删除)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // --- 操作函数 ---

  const handleAdd = () => {
    setCurrentRow(undefined);
    setModalVisible(true);
  };

  const handleEdit = (record: ApiItem) => {
    setCurrentRow(record);
    setModalVisible(true);
  };

  // 删除 (单条)
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteApi({ id });
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

  // 批量删除
  const handleBatchDelete = async () => {
    if (!selectedRowKeys.length) return;
    try {
      const res = await deleteApi({ ids: selectedRowKeys as number[] });
      if (res.code === 0) {
        message.success('批量删除成功');
        setSelectedRowKeys([]); // 清空选中
        actionRef.current?.reload();
      } else {
        message.error(res.msg || '删除失败');
      }
    } catch (error) {
      message.error('请求出错');
    }
  };

  // 提交表单
  const handleFinish = async (values: any) => {
    const isUpdate = !!currentRow;
    const method = isUpdate ? updateApi : createApi;
    const data = { ...values, id: currentRow?.ID };

    try {
      const res = await method(data);
      if (res.code === 0) {
        message.success(isUpdate ? '更新成功' : '添加成功');
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

  // --- 列定义 ---
  const columns: ProColumns<ApiItem>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      width: 60,
      search: false,
      align: 'center',
    },
    {
      title: 'API 路径',
      dataIndex: 'path',
      width: 250,
      copyable: true, // 支持一键复制
      ellipsis: true,
    },
    {
      title: 'API 分组',
      dataIndex: 'apiGroup',
      width: 120,
    },
    {
      title: 'API 描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '请求方法',
      dataIndex: 'method',
      width: 100,
      align: 'center',
      valueEnum: {
        POST: { text: 'POST', status: 'Processing' }, // 蓝色
        GET: { text: 'GET', status: 'Success' },      // 绿色
        PUT: { text: 'PUT', status: 'Warning' },      // 橙色
        DELETE: { text: 'DELETE', status: 'Error' },  // 红色
      },
      // 自定义渲染 Tag
      render: (_, record) => {
        const colors: Record<string, string> = {
          POST: 'blue',
          GET: 'green',
          PUT: 'orange',
          DELETE: 'red',
        };
        return <Tag color={colors[record.method] || 'default'}>{record.method}</Tag>;
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a key="edit" onClick={() => handleEdit(record)}>
            <EditOutlined /> 编辑
          </a>
          <Popconfirm
            title="确定删除?"
            description="删除后，关联的角色权限将自动清理。"
            onConfirm={() => handleDelete(record.ID)}
            okText="是"
            cancelText="否"
          >
            <a key="delete" style={{ color: '#ff4d4f' }}>
              <DeleteOutlined /> 删除
            </a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<ApiItem>
        headerTitle={false}
        actionRef={actionRef}
        rowKey="ID"
        search={{ labelWidth: 'auto' }}

        // 开启多选框
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys),
        }}

        request={async (params) => {
          // 转换分页参数并透传查询字段
          const res = await getApiList({
            page: params.current,
            pageSize: params.pageSize,
            path: params.path,
            description: params.description,
            apiGroup: params.apiGroup,
            method: params.method,
          });
          return {
            data: res.data?.list || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        }}

        columns={columns}
        scroll={{ x: 900 }}

        // 工具栏
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={handleAdd}>
            <PlusOutlined /> 新建 API
          </Button>,
          // 只有选中行时才显示批量删除按钮
          selectedRowKeys.length > 0 && (
            <Popconfirm
              key="batchDelete"
              title={`确定删除选中的 ${selectedRowKeys.length} 项 API 吗？`}
              onConfirm={handleBatchDelete}
              okText="确定"
              cancelText="取消"
            >
              <Button danger>
                <DeleteOutlined /> 批量删除
              </Button>
            </Popconfirm>
          ),
        ]}
      />

      <ModalForm
        title={currentRow ? '编辑 API' : '新建 API'}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        initialValues={currentRow}
        modalProps={{ destroyOnClose: true }}
      >
        <ProFormText
          name="path"
          label="API 路径"
          placeholder="e.g. /api/v1/user/info"
          rules={[{ required: true, message: '请输入路径' }]}
        />

        {/* ✨ 2. 使用 ProFormGroup 替代 div，并调整宽度 */}
        <ProFormGroup>
          <ProFormSelect
            name="method"
            label="请求方法"
            valueEnum={{
              POST: 'POST',
              GET: 'GET',
              PUT: 'PUT',
              DELETE: 'DELETE',
            }}
            placeholder="请选择"
            // 修改为 xs (约104px)，对于 POST/GET 足够了，节省空间
            width="xs"
            rules={[{ required: true, message: '请选择方法' }]}
          />
          <ProFormText
            name="apiGroup"
            label="API 分组"
            placeholder="e.g. 用户管理"
            // 保持 md (约328px)，因为前面的 xs 变小了，现在放得下了
            // xs(104) + md(328) + gap(16) = 448px < 容器宽度
            width="md"
            rules={[{ required: true, message: '请输入分组' }]}
          />
        </ProFormGroup>

        <ProFormTextArea
          name="description"
          label="API 描述"
          placeholder="简述 API 功能"
          rules={[{ required: true, message: '请输入描述' }]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default ApiTableList;
