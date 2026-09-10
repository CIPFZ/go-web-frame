import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n, formatDate } from '@/i18n';
import type { API } from '@/services/system/types';
import { loadPagedOptions } from '@/utils/pagedOptions';
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
import {
  Button,
  Form,
  message,
  Modal,
  Popconfirm,
  Space,
  Tag,
  Typography,
} from 'antd';
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
} from '@/services/system/apiToken';
import { ApiPermissionSummary } from './components/ApiPermissionSummary';
import {
  ApiPermissionTransfer,
  type ApiPermissionOption,
} from './components/ApiPermissionTransfer';
import {
  API_TOKEN_TABLE_LAYOUT,
  buildTokenFormInitialValues,
  buildTokenSubmitPayload,
  type TokenFormValues,
} from './helpers';
import './index.less';
const extractPlainToken = (res: API.CommonResponse): string | undefined => {
  const data = res?.data || {};
  return data.token || data.plainToken || data.accessToken || data.value;
};
const getExpiryStatus = (expiresAt?: string) => {
  if (!expiresAt) {
    return <Tag>{t('cms.notSet')}</Tag>;
  }
  const expires = dayjs(expiresAt);
  const now = dayjs();
  if (expires.isBefore(now)) {
    return <Tag color="error">{t('cms.expired')}</Tag>;
  }
  const remainingDays = expires.diff(now, 'day');
  if (remainingDays <= 7) {
    return (
      <Tag color="warning">
        {t('cms.expiresInOther', {
          value0: remainingDays,
        })}
      </Tag>
    );
  }
  return <Tag color="processing">{t('cms.valid')}</Tag>;
};
const ApiTokenPage: React.FC = () => {
  const localeFormRef1 = useFormLocale();
  useI18n();
  const actionRef = useRef<ActionType>(null);
  const [form] = Form.useForm<TokenFormValues>();
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [plainToken, setPlainToken] = useState<{
    token: string;
    reset: boolean;
  }>();
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
      const res = await loadPagedOptions(getApiOptions);
      if (res.code !== 0) {
        message.error(res.msg || t('cms.unableToLoadApis'));
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
      message.error(t('cms.unableToLoadApis'));
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
        getApiTokenDetail({
          id: record.ID,
        }),
        apiOptions.length ? Promise.resolve(null) : loadApiOptions(),
      ]);
      if (detailRes.code !== 0) {
        message.error(detailRes.msg || t('cms.unableToLoadDetails'));
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
      message.error(t('cms.requestError'));
    }
  };
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteApiToken({
        id,
      });
      if (res.code !== 0) {
        message.error(res.msg || t('cms.deleteFailed'));
        return;
      }
      message.success(t('cms.deletedSuccessfully'));
      actionRef.current?.reload();
    } catch (error) {
      message.error(t('cms.requestError'));
    }
  };
  const handleToggleEnable = async (record: ApiTokenItem) => {
    const request = record.enabled ? disableApiToken : enableApiToken;
    try {
      const res = await request({
        id: record.ID,
      });
      if (res.code !== 0) {
        message.error(res.msg || t('cms.statusUpdateFailed'));
        return;
      }
      message.success(record.enabled ? t('cms.disabled') : t('cms.enabled'));
      actionRef.current?.reload();
    } catch (error) {
      message.error(t('cms.requestError'));
    }
  };
  const handleReset = async (id: number) => {
    try {
      const res = await resetApiToken({
        id,
      });
      if (res.code !== 0) {
        message.error(res.msg || t('cms.resetFailed'));
        return;
      }
      message.success(t('cms.resetSuccessfully'));
      const plainToken = extractPlainToken(res);
      if (plainToken) {
        setPlainToken({ token: plainToken, reset: true });
      }
      actionRef.current?.reload();
    } catch (error) {
      message.error(t('cms.requestError'));
    }
  };
  const handleSubmit = async (values: TokenFormValues) => {
    const payload = buildTokenSubmitPayload(values, currentRow?.ID);
    try {
      const res = currentRow?.ID
        ? await updateApiToken(payload)
        : await createApiToken(payload);
      if (res.code !== 0) {
        message.error(res.msg || t('cms.saveFailed'));
        return false;
      }
      message.success(
        currentRow?.ID
          ? t('cms.updatedSuccessfully')
          : t('cms.createdSuccessfully'),
      );
      if (!currentRow?.ID) {
        const plainToken = extractPlainToken(res);
        if (plainToken) {
          setPlainToken({ token: plainToken, reset: false });
        }
      }
      setDrawerVisible(false);
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error(t('cms.requestError'));
      return false;
    }
  };
  const columns: ProColumns<ApiTokenItem>[] = [
    {
      title: t('cms.name'),
      dataIndex: 'name',
      width: API_TOKEN_TABLE_LAYOUT.nameWidth,
      ellipsis: true,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{record.name}</Typography.Text>
          <Typography.Text type="secondary">
            {record.description || t('cms.noDescription')}
          </Typography.Text>
        </Space>
      ),
    },
    {
      title: t('cms.tokenPrefix'),
      dataIndex: 'tokenPrefix',
      width: API_TOKEN_TABLE_LAYOUT.tokenPrefixWidth,
      search: false,
      render: (_, record) => (
        <Space direction="vertical" size={4}>
          <span className="tokenPrefix">
            <SafetyCertificateOutlined />
            {record.tokenPrefix}
          </span>
          <span className="tokenPrefixMeta">
            {t('cms.onlyThePrefixIsDisplayedFor')}
          </span>
        </Space>
      ),
    },
    {
      title: t('cms.authorizedApis.653b57'),
      dataIndex: 'apis',
      width: API_TOKEN_TABLE_LAYOUT.apisWidth,
      search: false,
      render: (_, record) => <ApiPermissionSummary apis={record.apis} />,
    },
    {
      title: t('cms.status'),
      dataIndex: 'enabled',
      width: API_TOKEN_TABLE_LAYOUT.statusWidth,
      valueEnum: {
        true: {
          text: t('cms.enabled'),
          status: 'Success',
        },
        false: {
          text: t('cms.disabled'),
          status: 'Default',
        },
      },
      render: (_, record) =>
        record.enabled ? (
          <Tag color="success" icon={<CheckCircleOutlined />}>
            {t('cms.enabled')}
          </Tag>
        ) : (
          <Tag icon={<PauseCircleOutlined />}>{t('cms.disabled')}</Tag>
        ),
    },
    {
      title: t('cms.concurrencyLimit'),
      dataIndex: 'maxConcurrency',
      width: API_TOKEN_TABLE_LAYOUT.concurrencyWidth,
      search: false,
      align: 'center',
      render: (_, record) => record.maxConcurrency || '-',
    },
    {
      title: t('cms.expiresAt'),
      dataIndex: 'expiresAt',
      width: API_TOKEN_TABLE_LAYOUT.expiresAtWidth,
      search: false,
      render: (_, record) => (
        <Space direction="vertical" size={4}>
          <Typography.Text>
            {record.expiresAt ? formatDate(record.expiresAt) : '-'}
          </Typography.Text>
          {getExpiryStatus(record.expiresAt)}
        </Space>
      ),
    },
    {
      title: t('cms.lastUsed'),
      dataIndex: 'lastUsedAt',
      width: API_TOKEN_TABLE_LAYOUT.lastUsedWidth,
      search: false,
      render: (_, record) =>
        record.lastUsedAt ? formatDate(record.lastUsedAt) : t('cms.noRecords'),
    },
    {
      title: t('cms.actions'),
      dataIndex: 'option',
      valueType: 'option',
      width: API_TOKEN_TABLE_LAYOUT.actionWidth,
      render: (_, record) => (
        <Space size="small" wrap>
          <a onClick={() => openEditDrawer(record)}>
            <EditOutlined />
            {t('cms.edit')}
          </a>
          <Popconfirm
            title={
              record.enabled
                ? t('cms.disableThisToken')
                : t('cms.enableThisToken')
            }
            onConfirm={() => handleToggleEnable(record)}
            okText={t('cms.confirm')}
            cancelText={t('cms.cancel')}
          >
            <a>
              {record.enabled ? (
                <PauseCircleOutlined />
              ) : (
                <CheckCircleOutlined />
              )}{' '}
              {record.enabled ? t('cms.disable') : t('cms.enable')}
            </a>
          </Popconfirm>
          <Popconfirm
            title={t('cms.resetThisToken')}
            description={t('cms.theOldTokenWillStopWorking')}
            onConfirm={() => handleReset(record.ID)}
            okText={t('cms.confirm')}
            cancelText={t('cms.cancel')}
          >
            <a>
              <RedoOutlined />
              {t('cms.reset')}
            </a>
          </Popconfirm>
          <Popconfirm
            title={t('cms.deleteThisToken')}
            onConfirm={() => handleDelete(record.ID)}
            okText={t('cms.confirm')}
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
    <PageContainer title={false} className="apiTokenPage">
      <ProCard className="apiTokenHero" bordered>
        <Space
          direction="vertical"
          size={4}
          style={{
            width: '100%',
          }}
        >
          <Typography.Title level={4} className="heroTitle">
            {t('cms.apiTokenConsole')}
          </Typography.Title>
          <Typography.Paragraph className="heroParagraph">
            {t('cms.manageApiTokensForExternalServices')}
          </Typography.Paragraph>
          <div className="statsRow">
            <div className="statCard">
              <span className="statLabel">{t('cms.tokensOnThisPage')}</span>
              <span className="statValue">{stats.total}</span>
              <span className="statHint">
                {t('cms.overviewOfTheCurrentSearchResults')}
              </span>
            </div>
            <div className="statCard">
              <span className="statLabel">{t('cms.enabled.595aab')}</span>
              <span className="statValue">{stats.enabledCount}</span>
              <span className="statHint">
                {t('cms.enabledTokensOnThisPage')}
              </span>
            </div>
            <div className="statCard">
              <span className="statLabel">{t('cms.expiringWithin7Days')}</span>
              <span className="statValue">{stats.expiringSoonCount}</span>
              <span className="statHint">
                {t('cms.reviewTokensThatNeedRenewalOr')}
              </span>
            </div>
          </div>
        </Space>
      </ProCard>

      <ProCard className="tableCard" bordered>
        <ProTable<ApiTokenItem>
          actionRef={actionRef}
          rowKey="ID"
          headerTitle={t('cms.tokenList')}
          search={{
            labelWidth: 'auto',
          }}
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
          scroll={{
            x: API_TOKEN_TABLE_LAYOUT.scrollX,
          }}
          toolBarRender={() => [
            <Button key="create" type="primary" onClick={openCreateDrawer}>
              <PlusOutlined />
              {t('cms.newToken')}
            </Button>,
          ]}
        />
      </ProCard>

      <DrawerForm<TokenFormValues>
        form={form}
        title={currentRow ? t('cms.editToken') : t('cms.newToken')}
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
        drawerProps={{
          destroyOnClose: true,
        }}
        initialValues={buildTokenFormInitialValues(currentRow)}
        formRef={localeFormRef1}
      >
        <div className="drawerTips">
          <div className="drawerTipsTitle">{t('cms.tokenGuidelines')}</div>
          <Typography.Paragraph className="drawerTipsText">
            {t('cms.anExpirationDateIsRequiredUse')}
          </Typography.Paragraph>
        </div>
        <ProFormText
          name="name"
          label={t('cms.name')}
          placeholder={t('cms.eGCiReleaseScriptOr')}
          rules={[
            {
              required: true,
              message: t('cms.enterATokenName'),
            },
          ]}
        />
        <ProFormTextArea
          name="description"
          label={t('cms.description')}
          placeholder={t('cms.describeItsPurposeSystemOrOwner')}
          fieldProps={{
            rows: 3,
            showCount: true,
            maxLength: 120,
          }}
        />
        <ProFormDigit
          name="maxConcurrency"
          label={t('cms.maximumConcurrency')}
          min={1}
          fieldProps={{
            precision: 0,
          }}
          rules={[
            {
              required: true,
              message: t('cms.enterTheMaximumConcurrency'),
            },
          ]}
        />
        <ProFormDateTimePicker
          name="expiresAt"
          label={t('cms.expiresAt')}
          fieldProps={{
            showNow: true,
          }}
          rules={[
            {
              required: true,
              message: t('cms.selectAnExpirationDate'),
            },
          ]}
        />
        <ProForm.Item
          name="apiIds"
          label={t('cms.authorizedApis.653b57')}
          rules={[
            {
              required: true,
              message: t('cms.selectAtLeastOneApi'),
            },
          ]}
        >
          <ApiPermissionTransfer
            loading={apiOptionsLoading}
            options={apiOptions}
          />
        </ProForm.Item>
      </DrawerForm>
      <Modal
        open={!!plainToken}
        title={t(
          plainToken?.reset
            ? 'cms.tokenResetCopyTheNewToken'
            : 'cms.tokenCreatedCopyItNow',
        )}
        width={640}
        onCancel={() => setPlainToken(undefined)}
        footer={
          <Button type="primary" onClick={() => setPlainToken(undefined)}>
            {t('cms.confirm')}
          </Button>
        }
      >
        <Typography.Paragraph type="secondary">
          {t('cms.thisTokenIsShownOnlyOnce')}
        </Typography.Paragraph>
        <Typography.Paragraph copyable code>
          {plainToken?.token}
        </Typography.Paragraph>
      </Modal>
    </PageContainer>
  );
};
export default ApiTokenPage;
