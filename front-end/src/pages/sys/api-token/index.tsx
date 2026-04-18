import React, { useMemo, useRef, useState } from 'react';
import {
  DrawerForm,
  ProCard,
  ProForm,
  ProFormDateTimePicker,
  ProFormDigit,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  CheckCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  PauseCircleOutlined,
  PlusOutlined,
  RedoOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { Button, Form, message, Modal, Popconfirm, Space, Tag, Typography } from 'antd';
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
import { ApiPermissionSummary } from './components/ApiPermissionSummary';
import {
  ApiPermissionTransfer,
  type ApiPermissionOption,
} from './components/ApiPermissionTransfer';
import {
  buildTokenFormInitialValues,
  buildTokenSubmitPayload,
  type TokenFormValues,
} from './helpers';
import './index.less';

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

const getExpiryStatus = (expiresAt?: string) => {
  if (!expiresAt) {
    return <Tag>未设置</Tag>;
  }

  const expires = dayjs(expiresAt);
  const now = dayjs();

  if (expires.isBefore(now)) {
    return <Tag color="error">已过期</Tag>;
  }

  const remainingDays = expires.diff(now, 'day');
  if (remainingDays <= 7) {
    return <Tag color="warning">{`${remainingDays} 天内到期`}</Tag>;
  }

  return <Tag color="processing">有效中</Tag>;
};

const ApiTokenPage: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [form] = Form.useForm<TokenFormValues>();
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentRow, setCurrentRow] = useState<ApiTokenItem>();
  const [apiOptions, setApiOptions] = useState<ApiPermissionOption[]>([]);
  const [apiOptionsLoading, setApiOptionsLoading] = useState(false);
  const [pageList, setPageList] = useState<ApiTokenItem[]>([]);

  const stats = useMemo(() => {
    const enabledCount = pageList.filter((item) => item.enabled).length;
    const expiringSoonCount = pageList.filter((item) => {
      if (!item.expiresAt) {
        return false;
      }
      const expires = dayjs(item.expiresAt);
      return expires.isAfter(dayjs()) && expires.diff(dayjs(), 'day') <= 7;
    }).length;

    return {
      total: pageList.length,
      enabledCount,
      expiringSoonCount,
    };
  }, [pageList]);

  const loadApiOptions = async () => {
    setApiOptionsLoading(true);
    try {
      const res = await getApiOptions({ page: 1, pageSize: 9999 });
      if (res.code !== 0) {
        message.error(res.msg || '加载 API 列表失败');
        return;
      }

      setApiOptions(
        (res.data?.list || []).map((item: any) => ({
          value: item.ID,
          label: `[${item.method}] ${item.path}`,
          method: item.method,
          path: item.path,
          apiGroup: item.apiGroup,
          description: item.description,
        })),
      );
    } catch (error) {
      message.error('加载 API 列表失败');
    } finally {
      setApiOptionsLoading(false);
    }
  };

  const openCreateDrawer = () => {
    setCurrentRow(undefined);
    form.setFieldsValue(buildTokenFormInitialValues());
    setDrawerVisible(true);
    if (!apiOptions.length) {
      loadApiOptions();
    }
  };

  const openEditDrawer = async (record: ApiTokenItem) => {
    try {
      const [detailRes] = await Promise.all([
        getApiTokenDetail({ id: record.ID }),
        apiOptions.length ? Promise.resolve(null) : loadApiOptions(),
      ]);

      if (detailRes.code !== 0) {
        message.error(detailRes.msg || '获取详情失败');
        return;
      }

      const detail = (detailRes.data || {}) as Partial<ApiTokenItem>;
      const nextRow = {
        ...record,
        ...detail,
        apis: Array.isArray(detail.apis) ? detail.apis : record.apis || [],
      };

      setCurrentRow(nextRow);
      form.setFieldsValue(buildTokenFormInitialValues(nextRow));
      setDrawerVisible(true);
    } catch (error) {
      message.error('请求异常');
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
      message.error('请求异常');
    }
  };

  const handleToggleEnable = async (record: ApiTokenItem) => {
    const request = record.enabled ? disableApiToken : enableApiToken;
    try {
      const res = await request({ id: record.ID });
      if (res.code !== 0) {
        message.error(res.msg || '状态更新失败');
        return;
      }
      message.success(record.enabled ? '已禁用' : '已启用');
      actionRef.current?.reload();
    } catch (error) {
      message.error('请求异常');
    }
  };

  const handleReset = async (id: number) => {
    try {
      const res = await resetApiToken({ id });
      if (res.code !== 0) {
        message.error(res.msg || '重置失败');
        return;
      }

      message.success('重置成功');
      const plainToken = extractPlainToken(res);
      if (plainToken) {
        showPlainTokenModal(plainToken, 'Token 已重置，请复制新的明文 Token');
      }
      actionRef.current?.reload();
    } catch (error) {
      message.error('请求异常');
    }
  };

  const handleSubmit = async (values: TokenFormValues) => {
    const payload = buildTokenSubmitPayload(values, currentRow?.ID);

    try {
      const res = currentRow?.ID ? await updateApiToken(payload) : await createApiToken(payload);
      if (res.code !== 0) {
        message.error(res.msg || '保存失败');
        return false;
      }

      message.success(currentRow?.ID ? '更新成功' : '创建成功');
      if (!currentRow?.ID) {
        const plainToken = extractPlainToken(res);
        if (plainToken) {
          showPlainTokenModal(plainToken, 'Token 创建成功，请立即复制');
        }
      }

      setDrawerVisible(false);
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error('请求异常');
      return false;
    }
  };

  const columns: ProColumns<ApiTokenItem>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      ellipsis: true,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{record.name}</Typography.Text>
          <Typography.Text type="secondary">{record.description || '未填写说明'}</Typography.Text>
        </Space>
      ),
    },
    {
      title: 'Token 前缀',
      dataIndex: 'tokenPrefix',
      width: 220,
      search: false,
      render: (_, record) => (
        <Space direction="vertical" size={4}>
          <span className="tokenPrefix">
            <SafetyCertificateOutlined />
            {record.tokenPrefix}
          </span>
          <span className="tokenPrefixMeta">仅展示前缀，用于快速识别</span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 120,
      valueEnum: {
        true: { text: '启用', status: 'Success' },
        false: { text: '禁用', status: 'Default' },
      },
      render: (_, record) =>
        record.enabled ? (
          <Tag color="success" icon={<CheckCircleOutlined />}>
            启用
          </Tag>
        ) : (
          <Tag icon={<PauseCircleOutlined />}>禁用</Tag>
        ),
    },
    {
      title: '并发上限',
      dataIndex: 'maxConcurrency',
      width: 110,
      search: false,
      align: 'center',
      render: (_, record) => record.maxConcurrency || '-',
    },
    {
      title: '过期时间',
      dataIndex: 'expiresAt',
      width: 220,
      search: false,
      render: (_, record) => (
        <Space direction="vertical" size={4}>
          <Typography.Text>{record.expiresAt ? dayjs(record.expiresAt).format('YYYY-MM-DD HH:mm') : '-'}</Typography.Text>
          {getExpiryStatus(record.expiresAt)}
        </Space>
      ),
    },
    {
      title: '最近使用',
      dataIndex: 'lastUsedAt',
      width: 180,
      search: false,
      render: (_, record) =>
        record.lastUsedAt ? dayjs(record.lastUsedAt).format('YYYY-MM-DD HH:mm') : '暂无记录',
    },
    {
      title: '授权 API',
      dataIndex: 'apis',
      search: false,
      render: (_, record) => <ApiPermissionSummary apis={record.apis} />,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 320,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small" wrap>
          <a onClick={() => openEditDrawer(record)}>
            <EditOutlined /> 编辑
          </a>
          <Popconfirm
            title={record.enabled ? '确认禁用该 Token？' : '确认启用该 Token？'}
            onConfirm={() => handleToggleEnable(record)}
            okText="确认"
            cancelText="取消"
          >
            <a>
              {record.enabled ? <PauseCircleOutlined /> : <CheckCircleOutlined />} {record.enabled ? '禁用' : '启用'}
            </a>
          </Popconfirm>
          <Popconfirm
            title="确认重置 Token？"
            description="重置后旧 Token 会立即失效。"
            onConfirm={() => handleReset(record.ID)}
            okText="确认"
            cancelText="取消"
          >
            <a>
              <RedoOutlined /> 重置
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
    <PageContainer title={false} className="apiTokenPage">
      <ProCard className="apiTokenHero" bordered>
        <Space direction="vertical" size={4} style={{ width: '100%' }}>
          <Typography.Title level={4} className="heroTitle">
            API Token 控制台
          </Typography.Title>
          <Typography.Paragraph className="heroParagraph">
            面向外部服务端和脚本调用场景。这里可以集中管理 Token 生命周期、授权接口范围和可用状态，避免在单条记录里堆叠过多信息。
          </Typography.Paragraph>
          <div className="statsRow">
            <div className="statCard">
              <span className="statLabel">当前页 Token</span>
              <span className="statValue">{stats.total}</span>
              <span className="statHint">用于快速感知当前检索范围</span>
            </div>
            <div className="statCard">
              <span className="statLabel">启用中</span>
              <span className="statValue">{stats.enabledCount}</span>
              <span className="statHint">可直接对外访问的 Token 数量</span>
            </div>
            <div className="statCard">
              <span className="statLabel">7 天内到期</span>
              <span className="statValue">{stats.expiringSoonCount}</span>
              <span className="statHint">需要尽快轮换或续期</span>
            </div>
          </div>
        </Space>
      </ProCard>

      <ProCard className="tableCard" bordered>
        <ProTable<ApiTokenItem>
          actionRef={actionRef}
          rowKey="ID"
          headerTitle="Token 列表"
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

            const list = res.data?.list || [];
            setPageList(list);

            return {
              success: res.code === 0,
              data: list,
              total: res.data?.total || 0,
            };
          }}
          columns={columns}
          scroll={{ x: 1500 }}
          toolBarRender={() => [
            <Button key="create" type="primary" onClick={openCreateDrawer}>
              <PlusOutlined /> 新建 Token
            </Button>,
          ]}
        />
      </ProCard>

      <DrawerForm<TokenFormValues>
        form={form}
        title={currentRow ? '编辑 Token' : '新建 Token'}
        width={760}
        open={drawerVisible}
        onOpenChange={(open) => {
          setDrawerVisible(open);
          if (!open) {
            setCurrentRow(undefined);
            form.resetFields();
          }
        }}
        onFinish={handleSubmit}
        drawerProps={{ destroyOnClose: true }}
        initialValues={buildTokenFormInitialValues(currentRow)}
      >
        <div className="drawerTips">
          <div className="drawerTipsTitle">填写建议</div>
          <Typography.Paragraph className="drawerTipsText">
            过期时间为必填项，建议为不同系统或脚本分别创建独立 Token，并只授予必要的 API 权限。
          </Typography.Paragraph>
        </div>
        <ProFormText
          name="name"
          label="名称"
          placeholder="例如：CI 发布脚本、第三方同步服务"
          rules={[{ required: true, message: '请输入 Token 名称' }]}
        />
        <ProFormTextArea
          name="description"
          label="说明"
          placeholder="填写 Token 的用途、所属系统或负责人"
          fieldProps={{ rows: 3, showCount: true, maxLength: 120 }}
        />
        <ProFormDigit
          name="maxConcurrency"
          label="最大并发"
          min={1}
          fieldProps={{ precision: 0 }}
          rules={[{ required: true, message: '请输入最大并发' }]}
        />
        <ProFormDateTimePicker
          name="expiresAt"
          label="过期时间"
          fieldProps={{ showNow: true }}
          rules={[{ required: true, message: '请选择过期时间' }]}
        />
        <ProForm.Item
          name="apiIds"
          label="授权 API"
          rules={[{ required: true, message: '请选择至少一个授权 API' }]}
        >
          <ApiPermissionTransfer loading={apiOptionsLoading} options={apiOptions} />
        </ProForm.Item>
      </DrawerForm>
    </PageContainer>
  );
};

export default ApiTokenPage;
