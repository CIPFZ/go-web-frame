import { formatDate, t, useI18n } from '@/i18n';
import type { ApiPermissionOption } from '@/pages/sys/api-token/components/ApiPermissionTransfer';
import { TokenPermissionFields } from '@/pages/sys/api-token/components/TokenPermissionFields';
import { buildTokenFormInitialValues, buildTokenSubmitPayload, type TokenFormValues } from '@/pages/sys/api-token/helpers';
import {
  type ApiTokenItem,
  type TokenLimits,
  createApiToken,
  deleteApiToken,
  disableApiToken,
  enableApiToken,
  getApiOptions,
  getApiTokenList,
  resetApiToken,
  updateApiToken,
} from '@/services/system/apiToken';
import { PlusOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { ModalForm, PageContainer, ProFormDateTimePicker, ProFormDigit, ProFormText } from '@ant-design/pro-components';
import { Button, Card, Collapse, Form, Modal, Popconfirm, Space, Table, Tag, Tooltip, Typography, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

/** 申请 / 修改自己的 Token：申请只填名字，其余按本人的上限；更多设置里可以在上限内调整。 */
function TokenForm({ token, limits, options, trigger, onSaved }: { token?: ApiTokenItem; limits?: TokenLimits; options: ApiPermissionOption[]; trigger: React.ReactElement; onSaved: (plain?: string) => void }) {
  const [form] = Form.useForm<TokenFormValues>();
  const initial = useMemo(() => {
    if (token) return buildTokenFormInitialValues(token);
    return {
      ...buildTokenFormInitialValues(),
      maxConcurrency: limits?.maxConcurrency,
      expiresAt: limits ? dayjs().add(limits.maxTtlDays, 'day').toISOString() : undefined,
    };
  }, [token, limits]);
  return (
    <ModalForm<TokenFormValues>
      form={form}
      title={token ? t('cms.editToken') : t('cms.myTokens.apply')}
      trigger={trigger}
      width={720}
      initialValues={initial}
      modalProps={{ destroyOnClose: true, afterOpenChange: (open) => open && form.setFieldsValue(initial) }}
      onFinish={async (values) => {
        const res = await (token ? updateApiToken(buildTokenSubmitPayload(values, token.ID)) : createApiToken(buildTokenSubmitPayload(values)));
        if (res.code !== 0) {
          message.error(res.msg || t('cms.saveFailed'));
          return false;
        }
        onSaved(token ? undefined : (res.data as { token?: string })?.token);
        return true;
      }}
    >
      <ProFormText name="name" label={t('cms.name')} placeholder={t('cms.myTokens.namePlaceholder')} rules={[{ required: true, whitespace: true, max: 100, message: t('cms.enterATokenName') }]} />
      <Collapse
        ghost
        size="small"
        items={[
          {
            key: 'more',
            label: t('cms.myTokens.moreSettings'),
            forceRender: true,
            children: (
              <>
                <TokenPermissionFields options={options} />
                <ProFormDateTimePicker name="expiresAt" label={t('cms.expiresAt')} tooltip={limits ? t('cms.myTokens.expiryHelp', { days: limits.maxTtlDays }) : undefined} rules={[{ required: true, message: t('cms.selectAnExpirationDate') }]} />
                <ProFormDigit name="maxConcurrency" label={t('cms.maximumConcurrency')} min={1} max={limits?.maxConcurrency} width="sm" fieldProps={{ precision: 0 }} />
              </>
            ),
          },
        ]}
      />
    </ModalForm>
  );
}

/** 我的 API Token：每个人申请、管理自己的 Token；Token 能做的事不超过本人现在的权限。 */
export default function MyTokensPage() {
  useI18n();
  const [tokens, setTokens] = useState<ApiTokenItem[]>([]);
  const [limits, setLimits] = useState<TokenLimits>();
  const [used, setUsed] = useState(0);
  const [options, setOptions] = useState<ApiPermissionOption[]>([]);
  const [loading, setLoading] = useState(true);
  // 明文 Token 只在申请或重置后的这一次拿得到：弹窗里显示一次，不保存
  const [shown, setShown] = useState<string>();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [list, opts] = await Promise.all([getApiTokenList({ page: 1, pageSize: 100 }), getApiOptions({})]);
      setTokens(list.data?.list || []);
      const d = opts.data as { list?: any[]; limits?: TokenLimits; usedTokens?: number };
      setLimits(d?.limits);
      setUsed(d?.usedTokens || 0);
      setOptions((d?.list || []).map((x) => ({ value: x.ID, label: `[${x.method}] ${x.path}`, method: x.method, path: x.path, apiGroup: x.apiGroup, description: x.description })));
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    load();
  }, [load]);

  const act = async (run: () => Promise<{ code: number; msg?: string; data?: any }>, done: string) => {
    const res = await run();
    if (res.code !== 0) {
      message.error(res.msg || t('cms.requestError'));
      return;
    }
    message.success(done);
    if (res.data?.token) setShown(res.data.token);
    load();
  };
  const atLimit = !!limits && used >= limits.maxTokens;

  const columns: ColumnsType<ApiTokenItem> = [
    { title: t('cms.name'), dataIndex: 'name', width: 180, render: (name) => <Typography.Text strong ellipsis={{ tooltip: name }}>{name}</Typography.Text> },
    {
      title: (
        <Space size={4}>
          Token
          <Tooltip title={t('cms.myTokens.prefixHelp')}>
            <QuestionCircleOutlined style={{ color: 'rgba(0,0,0,0.45)' }} />
          </Tooltip>
        </Space>
      ),
      width: 130,
      render: (_, r) => <Typography.Text code>{r.tokenPrefix}…</Typography.Text>,
    },
    { title: t('cms.tokenPermissionMode'), width: 140, render: (_, r) => (r.followOwner ? <Tag color="blue">{t('cms.tokenPermissionFollow')}</Tag> : t('cms.myTokens.apiCount', { n: r.apis?.length || 0 })) },
    { title: t('cms.status'), width: 80, render: (_, r) => (r.enabled ? <Tag color="success">{t('cms.enabled')}</Tag> : <Tag>{t('cms.disabled')}</Tag>) },
    { title: t('cms.expiresAt'), width: 160, render: (_, r) => (r.expiresAt ? (dayjs(r.expiresAt).isBefore(dayjs()) ? <Tag color="error">{t('cms.expired')}</Tag> : formatDate(r.expiresAt)) : '-') },
    { title: t('cms.lastUsed'), width: 160, render: (_, r) => (r.lastUsedAt ? formatDate(r.lastUsedAt) : t('cms.noRecords')) },
    {
      title: t('cms.actions'),
      width: 210,
      fixed: 'right',
      render: (_, r) => (
        <Space size="small">
          <TokenForm token={r} limits={limits} options={options} trigger={<a>{t('cms.edit')}</a>} onSaved={() => load()} />
          <Popconfirm title={r.enabled ? t('cms.disableThisToken') : t('cms.enableThisToken')} onConfirm={() => act(() => (r.enabled ? disableApiToken : enableApiToken)({ id: r.ID }), r.enabled ? t('cms.disabled') : t('cms.enabled'))}>
            <a>{r.enabled ? t('cms.disable') : t('cms.enable')}</a>
          </Popconfirm>
          <Popconfirm title={t('cms.resetThisToken')} description={t('cms.theOldTokenWillStopWorking')} onConfirm={() => act(() => resetApiToken({ id: r.ID }), t('cms.resetSuccessfully'))}>
            <a>{t('cms.reset')}</a>
          </Popconfirm>
          <Popconfirm title={t('cms.deleteThisToken')} onConfirm={() => act(() => deleteApiToken({ id: r.ID }), t('cms.deletedSuccessfully'))}>
            <a style={{ color: '#ff4d4f' }}>{t('cms.delete')}</a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={t('cms.myTokens.title')} content={t('cms.myTokens.description')}>
      <Card
        title={limits ? t('cms.myTokens.usage', { used, max: limits.maxTokens }) : undefined}
        extra={
          <TokenForm
            limits={limits}
            options={options}
            trigger={
              <Button type="primary" icon={<PlusOutlined />} disabled={atLimit}>
                {t('cms.myTokens.apply')}
              </Button>
            }
            onSaved={(token) => {
              if (token) setShown(token);
              load();
            }}
          />
        }
      >
        <Typography.Paragraph type="secondary">{t('cms.myTokens.header')}</Typography.Paragraph>
        <Table<ApiTokenItem> rowKey="ID" size="middle" tableLayout="fixed" scroll={{ x: 1060 }} loading={loading} columns={columns} dataSource={tokens} pagination={false} locale={{ emptyText: t('cms.myTokens.empty') }} />
      </Card>
      <Modal
        open={!!shown}
        title={t('cms.myTokens.ready')}
        width={640}
        onCancel={() => setShown(undefined)}
        footer={
          <Button type="primary" onClick={() => setShown(undefined)}>
            {t('cms.confirm')}
          </Button>
        }
      >
        <Typography.Paragraph type="secondary">{t('cms.myTokens.readyHint')}</Typography.Paragraph>
        <Typography.Paragraph code copyable={{ text: shown }} style={{ wordBreak: 'break-all' }}>
          {shown}
        </Typography.Paragraph>
      </Modal>
    </PageContainer>
  );
}
