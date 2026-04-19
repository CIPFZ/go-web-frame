import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, ProCard, ProFormTextArea, ProTable } from '@ant-design/pro-components';
import { getLocale, history } from '@umijs/max';
import { Button, Space, Tag, Typography, message } from 'antd';

import {
  claimWorkOrder,
  getWorkOrderPool,
  resetWorkOrder,
  transitionRelease,
  type PluginReleaseItem,
} from '@/services/api/plugin';
import { isEnglishLocale } from '@/utils/plugin';
import { ProcessStatusTag, ReleaseStatusTag, RequestTypeTag } from '../components/status';

const copyMap = {
  zh: {
    title: '插件工单池',
    subtitle: '审核员在这里认领、审核、发布和下架插件发布单；管理员可重置卡住的工单。',
    list: '待处理工单',
    refresh: '刷新',
    detail: '项目详情',
    claim: '认领',
    approve: '审核通过',
    reject: '打回',
    publish: '执行发布',
    offline: '执行下架',
    reset: '重置工单',
    review: '审核意见',
    resetReason: '重置原因',
    actions: '操作',
  },
  en: {
    title: 'Plugin Work Orders',
    subtitle: 'Reviewers claim, review, publish, and offline release requests here. Admins can reset blocked work orders.',
    list: 'Pending Work Orders',
    refresh: 'Refresh',
    detail: 'Project Detail',
    claim: 'Claim',
    approve: 'Approve',
    reject: 'Reject',
    publish: 'Release',
    offline: 'Offline',
    reset: 'Reset',
    review: 'Review Comment',
    resetReason: 'Reset Reason',
    actions: 'Actions',
  },
};

const PluginWorkOrderPoolPage: React.FC = () => {
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const actionRef = useRef<ActionType>(null);
  const [reviewTarget, setReviewTarget] = useState<PluginReleaseItem>();
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>();
  const [resetTarget, setResetTarget] = useState<PluginReleaseItem>();

  const reload = () => actionRef.current?.reload();

  const handleClaim = async (record: PluginReleaseItem) => {
    const res = await claimWorkOrder({ id: record.ID });
    if (res.code !== 0) {
      message.error(res.msg || copy.claim);
      return;
    }
    reload();
  };

  const handleDirectAction = async (record: PluginReleaseItem, action: string) => {
    const res = await transitionRelease({ id: record.ID, action });
    if (res.code !== 0) {
      message.error(res.msg || action);
      return;
    }
    reload();
  };

  const columns: ProColumns<PluginReleaseItem>[] = [
    {
      title: isEnglishLocale(locale) ? 'Plugin' : '插件',
      dataIndex: 'pluginNameZh',
      width: 220,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{record.pluginNameZh}</Typography.Text>
          <Typography.Text type="secondary">{record.pluginCode}</Typography.Text>
        </Space>
      ),
    },
    {
      title: isEnglishLocale(locale) ? 'Version' : '版本',
      dataIndex: 'version',
      width: 120,
    },
    {
      title: isEnglishLocale(locale) ? 'Request Type' : '请求类型',
      dataIndex: 'requestType',
      width: 140,
      render: (_, record) => <RequestTypeTag type={record.requestType} locale={locale} />,
    },
    {
      title: isEnglishLocale(locale) ? 'Lifecycle' : '生命周期',
      dataIndex: 'status',
      width: 120,
      render: (_, record) => <ReleaseStatusTag status={record.status} locale={locale} />,
    },
    {
      title: isEnglishLocale(locale) ? 'Workflow' : '工单状态',
      dataIndex: 'processStatus',
      width: 120,
      render: (_, record) => <ProcessStatusTag status={record.processStatus} locale={locale} />,
    },
    {
      title: isEnglishLocale(locale) ? 'Claim' : '认领状态',
      dataIndex: 'claimerId',
      width: 140,
      search: false,
      render: (_, record) => (record.claimerId ? <Tag color="processing">#{record.claimerId}</Tag> : <Tag>-</Tag>),
    },
    {
      title: copy.actions,
      dataIndex: 'option',
      valueType: 'option',
      width: 360,
      render: (_, record) => (
        <Space wrap>
          <a onClick={() => history.push(`/plugin/project/${record.pluginId}`)}>{copy.detail}</a>
          {record.processStatus === 0 && <a onClick={() => void handleClaim(record)}>{copy.claim}</a>}
          {record.status === 2 && record.processStatus === 1 && (
            <>
              <a
                onClick={() => {
                  setReviewTarget(record);
                  setReviewAction('approve');
                }}
              >
                {copy.approve}
              </a>
              <a
                onClick={() => {
                  setReviewTarget(record);
                  setReviewAction('reject');
                }}
              >
                {copy.reject}
              </a>
            </>
          )}
          {record.status === 3 && record.processStatus === 1 && record.requestType === 1 && (
            <a onClick={() => void handleDirectAction(record, 'release')}>{copy.publish}</a>
          )}
          {record.status === 3 && record.processStatus === 1 && record.requestType === 2 && (
            <a onClick={() => void handleDirectAction(record, 'offline')}>{copy.offline}</a>
          )}
          {record.processStatus === 1 && <a onClick={() => setResetTarget(record)}>{copy.reset}</a>}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <ProCard bordered style={{ borderRadius: 24 }}>
          <Space direction="vertical" size={6}>
            <Typography.Title level={3} style={{ margin: 0 }}>
              {copy.title}
            </Typography.Title>
            <Typography.Paragraph type="secondary" style={{ margin: 0 }}>
              {copy.subtitle}
            </Typography.Paragraph>
          </Space>
        </ProCard>

        <ProTable<PluginReleaseItem>
          actionRef={actionRef}
          rowKey="ID"
          headerTitle={copy.list}
          search={false}
          columns={columns}
          request={async (params) => {
            const res = await getWorkOrderPool({ page: params.current, pageSize: params.pageSize });
            return {
              data: res.data?.list || [],
              success: res.code === 0,
              total: res.data?.total || 0,
            };
          }}
          toolBarRender={() => [<Button key="refresh" onClick={reload}>{copy.refresh}</Button>]}
        />
      </Space>

      <ModalForm<{ reviewComment?: string }>
        title={reviewAction === 'approve' ? copy.approve : copy.reject}
        open={Boolean(reviewTarget)}
        onOpenChange={(open) => {
          if (!open) {
            setReviewTarget(undefined);
            setReviewAction(undefined);
          }
        }}
        onFinish={async (values) => {
          if (!reviewTarget || !reviewAction) return false;
          const res = await transitionRelease({
            id: reviewTarget.ID,
            action: reviewAction,
            reviewComment: values.reviewComment,
          });
          if (res.code !== 0) {
            message.error(res.msg || copy.review);
            return false;
          }
          setReviewTarget(undefined);
          setReviewAction(undefined);
          reload();
          return true;
        }}
      >
        <ProFormTextArea
          name="reviewComment"
          label={copy.review}
          rules={reviewAction === 'reject' ? [{ required: true }] : undefined}
        />
      </ModalForm>

      <ModalForm<{ reason: string }>
        title={copy.reset}
        open={Boolean(resetTarget)}
        onOpenChange={(open) => {
          if (!open) setResetTarget(undefined);
        }}
        onFinish={async (values) => {
          if (!resetTarget) return false;
          const res = await resetWorkOrder({ id: resetTarget.ID, reason: values.reason });
          if (res.code !== 0) {
            message.error(res.msg || copy.reset);
            return false;
          }
          setResetTarget(undefined);
          reload();
          return true;
        }}
      >
        <ProFormTextArea name="reason" label={copy.resetReason} rules={[{ required: true }]} />
      </ModalForm>
    </PageContainer>
  );
};

export default PluginWorkOrderPoolPage;
