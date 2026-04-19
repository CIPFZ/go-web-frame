import React, { useMemo, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, ProFormTextArea, ProTable } from '@ant-design/pro-components';
import { getLocale, history } from '@umijs/max';
import { Button, Input, Select, Space, Tabs, Typography, message } from 'antd';

import {
  claimWorkOrder,
  getWorkOrderPool,
  resetWorkOrder,
  transitionRelease,
  type PluginReleaseItem,
} from '@/services/api/plugin';
import { getDisplayName, isEnglishLocale } from '@/utils/plugin';
import { ProcessStatusTag, ReleaseStatusTag, RequestTypeTag } from '../components/status';

type FilterState = {
  keyword?: string;
  status?: number;
  requestType?: number;
  processStatus?: number;
};

const copyMap = {
  zh: {
    title: '插件工单池',
    list: '工单列表',
    detail: '工单详情',
    claim: '认领',
    approve: '审核通过',
    reject: '驳回',
    publish: '执行发布',
    offline: '执行下架',
    reset: '重置工单',
    review: '审核意见',
    resetReason: '重置原因',
    actions: '操作',
    all: '全部',
    mine: '待我处理',
    keyword: '搜索插件/版本/TD',
    plugin: '插件',
    version: '版本',
    requestType: '请求类型',
    lifecycle: '生命周期',
    workflow: '工单状态',
    statusFilter: '按生命周期筛选',
    requestFilter: '按请求类型筛选',
    processFilter: '按工单状态筛选',
    search: '搜索',
    resetFilters: '重置',
  },
  en: {
    title: 'Plugin Work Orders',
    list: 'Work Orders',
    detail: 'Work Order Detail',
    claim: 'Claim',
    approve: 'Approve',
    reject: 'Reject',
    publish: 'Release',
    offline: 'Offline',
    reset: 'Reset',
    review: 'Review Comment',
    resetReason: 'Reset Reason',
    actions: 'Actions',
    all: 'All',
    mine: 'Mine',
    keyword: 'Search plugin / version / TD',
    plugin: 'Plugin',
    version: 'Version',
    requestType: 'Request Type',
    lifecycle: 'Lifecycle',
    workflow: 'Workflow',
    statusFilter: 'Lifecycle',
    requestFilter: 'Request Type',
    processFilter: 'Workflow',
    search: 'Search',
    resetFilters: 'Reset',
  },
};

const PluginWorkOrderPoolPage: React.FC = () => {
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const actionRef = useRef<ActionType>(null);
  const [reviewTarget, setReviewTarget] = useState<PluginReleaseItem>();
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>();
  const [resetTarget, setResetTarget] = useState<PluginReleaseItem>();
  const [tab, setTab] = useState<'all' | 'mine'>('all');
  const [draftFilters, setDraftFilters] = useState<FilterState>({});
  const [filters, setFilters] = useState<FilterState>({});

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
      title: copy.plugin,
      dataIndex: 'pluginNameZh',
      width: 220,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>
            {getDisplayName(locale, { nameZh: record.pluginNameZh, nameEn: record.pluginNameEn })}
          </Typography.Text>
          <Typography.Text type="secondary">{record.pluginCode}</Typography.Text>
        </Space>
      ),
    },
    {
      title: copy.version,
      dataIndex: 'version',
      width: 120,
    },
    {
      title: copy.requestType,
      dataIndex: 'requestType',
      width: 140,
      render: (_, record) => <RequestTypeTag type={record.requestType} locale={locale} />,
    },
    {
      title: copy.lifecycle,
      dataIndex: 'status',
      width: 120,
      render: (_, record) => <ReleaseStatusTag status={record.status} locale={locale} />,
    },
    {
      title: copy.workflow,
      dataIndex: 'processStatus',
      width: 120,
      render: (_, record) => <ProcessStatusTag status={record.processStatus} locale={locale} />,
    },
    {
      title: copy.actions,
      dataIndex: 'option',
      valueType: 'option',
      width: 360,
      render: (_, record) => (
        <Space wrap>
          <a onClick={() => history.push(`/plugin/work-order/${record.pluginId}/${record.ID}`)}>{copy.detail}</a>
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

  const tabItems = useMemo(
    () => [
      { key: 'all', label: copy.all },
      { key: 'mine', label: copy.mine },
    ],
    [copy.all, copy.mine],
  );

  const handleSearch = () => {
    setFilters(draftFilters);
    reload();
  };

  const handleReset = () => {
    const emptyFilters = {};
    setDraftFilters(emptyFilters);
    setFilters(emptyFilters);
    reload();
  };

  return (
    <PageContainer title={false}>
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <div
          style={{
            background: '#fff',
            borderRadius: 24,
            border: '1px solid #f0f0f0',
            padding: 24,
          }}
        >
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Typography.Title level={3} style={{ margin: 0 }}>
              {copy.title}
            </Typography.Title>
            <Tabs
              activeKey={tab}
              items={tabItems}
              onChange={(value) => {
                setTab(value as 'all' | 'mine');
                reload();
              }}
            />
            <Space wrap>
              <Input
                placeholder={copy.keyword}
                value={draftFilters.keyword}
                onChange={(event) => setDraftFilters((prev) => ({ ...prev, keyword: event.target.value }))}
                style={{ width: 240 }}
                allowClear
              />
              <Select
                allowClear
                placeholder={copy.statusFilter}
                value={draftFilters.status}
                onChange={(value) => setDraftFilters((prev) => ({ ...prev, status: value }))}
                style={{ width: 180 }}
                options={[
                  { label: isEnglishLocale(locale) ? 'Ready' : '待提交', value: 1 },
                  { label: isEnglishLocale(locale) ? 'Pending Review' : '待审核', value: 2 },
                  { label: isEnglishLocale(locale) ? 'Approved' : '已通过', value: 3 },
                  { label: isEnglishLocale(locale) ? 'Rejected' : '已驳回', value: 4 },
                  { label: isEnglishLocale(locale) ? 'Released' : '已发布', value: 5 },
                  { label: isEnglishLocale(locale) ? 'Offlined' : '已下架', value: 6 },
                ]}
              />
              <Select
                allowClear
                placeholder={copy.requestFilter}
                value={draftFilters.requestType}
                onChange={(value) => setDraftFilters((prev) => ({ ...prev, requestType: value }))}
                style={{ width: 180 }}
                options={[
                  { label: isEnglishLocale(locale) ? 'Version Release' : '版本发布', value: 1 },
                  { label: isEnglishLocale(locale) ? 'Offline Request' : '下架申请', value: 2 },
                ]}
              />
              <Select
                allowClear
                placeholder={copy.processFilter}
                value={draftFilters.processStatus}
                onChange={(value) => setDraftFilters((prev) => ({ ...prev, processStatus: value }))}
                style={{ width: 180 }}
                options={[
                  { label: isEnglishLocale(locale) ? 'Pending' : '待处理', value: 0 },
                  { label: isEnglishLocale(locale) ? 'Processing' : '处理中', value: 1 },
                  { label: isEnglishLocale(locale) ? 'Rejected' : '已退回', value: 2 },
                  { label: isEnglishLocale(locale) ? 'Done' : '已完成', value: 3 },
                ]}
              />
              <Button type="primary" onClick={handleSearch}>
                {copy.search}
              </Button>
              <Button onClick={handleReset}>{copy.resetFilters}</Button>
            </Space>
          </Space>
        </div>

        <ProTable<PluginReleaseItem>
          actionRef={actionRef}
          rowKey="ID"
          headerTitle={copy.list}
          search={false}
          columns={columns}
          request={async (params) => {
            const res = await getWorkOrderPool({
              page: params.current,
              pageSize: params.pageSize,
              scope: tab,
              keyword: filters.keyword,
              status: filters.status,
              requestType: filters.requestType,
              processStatus: filters.processStatus,
            });
            return {
              data: res.data?.list || [],
              success: res.code === 0,
              total: res.data?.total || 0,
            };
          }}
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
