import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ModalForm,
  ProFormDateTimePicker,
  ProFormDigit,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { DeleteOutlined, EditOutlined, KeyOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, message, Modal, Popconfirm, Space, Tag, Typography } from 'antd';
import dayjs from 'dayjs';
import {
  createApiToken,
  deleteApiToken,
  disableApiToken,
  enableApiToken,
  getApiOptions,
  getApiTokenDetail,
  getApiTokenList,
  resetApiToken,
  updateApiToken,
  type ApiTokenItem,
} from '@/services/api/apiToken';

type ApiOption = {
  label: string;
  value: number;
};

type TokenFormValues = {
  name: string;
  description?: string;
  maxConcurrency?: number;
  expiresAt?: string;
  apiIds?: number[];
};

const extractPlainToken = (res: API.CommonResponse): string | undefined => {
  const data = res?.data || {};
  return data.token || data.plainToken || data.accessToken || data.value;
};

const showPlainTokenModal = (token: string, title: string) => {
  Modal.info({
    title,
    width: 640,
    content: (
      <div>
        <Typography.Paragraph type="secondary">
          明文 Token 只展示一次，请立即复制并妥善保管。
        </Typography.Paragraph>
        <Typography.Paragraph copyable code>
          {token}
        </Typography.Paragraph>
      </div>
    ),
  });
};

const ApiTokenPage: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [currentRow, setCurrentRow] = useState<ApiTokenItem>();

  const openCreateModal = () => {
    setCurrentRow(undefined);
    setModalVisible(true);
  };

  const openEditModal = async (record: ApiTokenItem) => {
    try {
      const res = await getApiTokenDetail({ id: record.ID });
      if (res.code !== 0) {
        message.error(res.msg || '获取详情失败');
        return;
      }

      const detail = (res.data || {}) as Partial<ApiTokenItem>;
      setCurrentRow({
        ...record,
        ...detail,
        apis: Array.isArray(detail.apis) ? detail.apis : record.apis || [],
      });
      setModalVisible(true);
    } catch (error) {
      message.error('请求出错');
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const res = await deleteApiToken({ id });
      if (res.code !== 0) {
        message.error(res.msg || '删除失败');
        return;
      }
      message.success('删除成功');
      actionRef.current?.reload();
    } catch (error) {
      message.error('请求出错');
    }
  };

  const handleToggleEnable = async (record: ApiTokenItem) => {
    const method = record.enabled ? disableApiToken : enableApiToken;
    try {
      const res = await method({ id: record.ID });
      if (res.code !== 0) {
        message.error(res.msg || '状态更新失败');
        return;
      }
      message.success(record.enabled ? '已禁用' : '已启用');
      actionRef.current?.reload();
    } catch (error) {
      message.error('请求出错');
    }
  };

  const handleReset = async (id: number) => {
    try {
      const res = await resetApiToken({ id });
      if (res.code !== 0) {
        message.error(res.msg || '重置失败');
        return;
      }

      const plainToken = extractPlainToken(res);
      message.success('重置成功');
      if (plainToken) {
        showPlainTokenModal(plainToken, '重置成功，请复制新 Token');
      }
      actionRef.current?.reload();
    } catch (error) {
      message.error('请求出错');
    }
  };

  const handleSubmit = async (values: TokenFormValues) => {
    const isUpdate = Boolean(currentRow?.ID);
    const payload = {
      ...(isUpdate ? { id: currentRow?.ID } : {}),
      name: values.name,
      description: values.description,
      maxConcurrency: values.maxConcurrency,
      expiresAt: values.expiresAt ? dayjs(values.expiresAt).toISOString() : undefined,
      apiIds: values.apiIds || [],
    };

    try {
      const res = isUpdate ? await updateApiToken(payload) : await createApiToken(payload);
      if (res.code !== 0) {
        message.error(res.msg || '保存失败');
        return false;
      }

      message.success(isUpdate ? '更新成功' : '创建成功');
      if (!isUpdate) {
        const plainToken = extractPlainToken(res);
        if (plainToken) {
          showPlainTokenModal(plainToken, '创建成功，请复制 Token');
        }
      }

      setModalVisible(false);
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error('请求出错');
      return false;
    }
  };

  const columns: ProColumns<ApiTokenItem>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      width: 70,
      search: false,
      align: 'center',
    },
    {
      title: 'Token 前缀',
      dataIndex: 'tokenPrefix',
      width: 180,
      search: false,
      copyable: true,
      ellipsis: true,
    },
    {
      title: '名称',
      dataIndex: 'name',
      ellipsis: true,
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 90,
      valueEnum: {
        true: { text: '启用', status: 'Success' },
        false: { text: '禁用', status: 'Default' },
      },
      render: (_, record) =>
        record.enabled ? <Tag color="success">启用</Tag> : <Tag color="default">禁用</Tag>,
    },
    {
      title: '最大并发',
      dataIndex: 'maxConcurrency',
      width: 100,
      search: false,
    },
    {
      title: '过期时间',
      dataIndex: 'expiresAt',
      valueType: 'dateTime',
      width: 180,
      search: false,
    },
    {
      title: '最近使用',
      dataIndex: 'lastUsedAt',
      valueType: 'dateTime',
      width: 180,
      search: false,
    },
    {
      title: '授权 API',
      dataIndex: 'apis',
      search: false,
      render: (_, record) => (
        <Space size={[4, 4]} wrap>
          {(record.apis || []).length > 0
            ? record.apis?.map((api) => (
                <Tag key={`${record.ID}-${api.ID}`} color="blue">
                  [{api.method}] {api.path}
                </Tag>
              ))
            : '-'}
        </Space>
      ),
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 260,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small" wrap>
          <a onClick={() => openEditModal(record)}>
            <EditOutlined /> 编辑
          </a>
          <Popconfirm
            title={record.enabled ? '确认禁用该 Token？' : '确认启用该 Token？'}
            onConfirm={() => handleToggleEnable(record)}
            okText="确认"
            cancelText="取消"
          >
            <a>{record.enabled ? '禁用' : '启用'}</a>
          </Popconfirm>
          <Popconfirm
            title="确认重置 Token？"
            description="重置后原 Token 将立即失效。"
            onConfirm={() => handleReset(record.ID)}
            okText="确认"
            cancelText="取消"
          >
            <a>
              <KeyOutlined /> 重置
            </a>
          </Popconfirm>
          <Popconfirm
            title="确认删除该 Token？"
            onConfirm={() => handleDelete(record.ID)}
            okText="确认"
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
      <ProTable<ApiTokenItem>
        actionRef={actionRef}
        rowKey="ID"
        headerTitle={false}
        search={{ labelWidth: 'auto' }}
        request={async (params) => {
          const enabled =
            params.enabled === undefined
              ? undefined
              : params.enabled === 'true' || params.enabled === true;

          const res = await getApiTokenList({
            page: params.current,
            pageSize: params.pageSize,
            name: params.name as string,
            enabled,
          });

          return {
            success: res.code === 0,
            data: res.data?.list || [],
            total: res.data?.total || 0,
          };
        }}
        columns={columns}
        scroll={{ x: 1350 }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={openCreateModal}>
            <PlusOutlined /> 新建 Token
          </Button>,
        ]}
      />

      <ModalForm<TokenFormValues>
        title={currentRow ? '编辑 Token' : '新建 Token'}
        width={620}
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true }}
        initialValues={
          currentRow
            ? {
                name: currentRow.name,
                description: currentRow.description,
                maxConcurrency: currentRow.maxConcurrency,
                expiresAt: currentRow.expiresAt,
                apiIds: (currentRow.apis || []).map((api) => api.ID),
              }
            : {
                maxConcurrency: 5,
                apiIds: [],
              }
        }
      >
        <ProFormText
          name="name"
          label="名称"
          placeholder="请输入 Token 名称"
          rules={[{ required: true, message: '请输入名称' }]}
        />
        <ProFormTextArea
          name="description"
          label="描述"
          placeholder="可选：用途说明"
          fieldProps={{ rows: 3 }}
        />
        <ProFormDigit
          name="maxConcurrency"
          label="最大并发"
          min={1}
          fieldProps={{ precision: 0 }}
          rules={[{ required: true, message: '请输入最大并发' }]}
        />
        <ProFormDateTimePicker name="expiresAt" label="过期时间" />
        <ProFormSelect
          name="apiIds"
          label="授权 API"
          mode="multiple"
          placeholder="请选择可调用 API"
          request={async (): Promise<ApiOption[]> => {
            const res = await getApiOptions({ page: 1, pageSize: 9999 });
            if (res.code !== 0) {
              return [];
            }

            return (res.data?.list || []).map((item: any) => ({
              value: item.ID,
              label: `[${item.method}] ${item.path}`,
            }));
          }}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default ApiTokenPage;
