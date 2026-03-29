import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { history } from '@umijs/max';
import { App, Card, Col, Empty, Row, Segmented, Space, Statistic, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import WorkbenchFilterBar from '../components/WorkbenchFilterBar';
import {
  getReleaseList,
  ReleaseListItem,
  ReleaseRequestType,
  ReleaseStatus,
} from '@/services/api/plugin';
import { getCurrentUserInfo } from '@/services/api/user';

type QueryRequestType = ReleaseRequestType | 'all';
type QueryStatus = ReleaseStatus | 'all';
type ReviewScope = 'pending' | 'reviewed';

type ReviewSummary = {
  pending: number;
  approved: number;
  rejected: number;
  mineQueue: number;
};

const cardStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const requestTypeOptions = [
  { label: '全部类型', value: 'all' },
  { label: '首发', value: 'initial' },
  { label: '版本更新', value: 'maintenance' },
  { label: '下架申请', value: 'offline' },
];

const pendingStatusOptions = [
  { label: '全部待审核', value: 'all' },
  { label: '待审核', value: 'pending_review' },
];

const reviewedStatusOptions = [
  { label: '全部已处理', value: 'all' },
  { label: '已通过', value: 'approved' },
  { label: '已打回', value: 'rejected' },
];

const requestTypeLabel: Record<ReleaseRequestType, string> = {
  initial: '首发',
  maintenance: '版本更新',
  offline: '下架申请',
};

const statusMeta: Record<ReleaseStatus, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  release_preparing: { label: '资料准备中', color: 'gold' },
  pending_review: { label: '待审核', color: 'processing' },
  approved: { label: '待发布', color: 'cyan' },
  rejected: { label: '已打回', color: 'red' },
  released: { label: '已发布', color: 'green' },
  offlined: { label: '已下架', color: 'volcano' },
};

const normalizeReleaseList = (res: any) => {
  const data = res?.data ?? {};
  return {
    list: Array.isArray(data.list) ? (data.list as ReleaseListItem[]) : [],
    total: typeof data.total === 'number' ? data.total : 0,
  };
};

const participantLabel = (userId: number | null | undefined, currentUserId?: number) => {
  if (!userId) return '-';
  if (currentUserId && userId === currentUserId) return '我';
  return `#${userId}`;
};

const ReviewWorkbenchPage: React.FC = () => {
  const { message } = App.useApp();
  const [initialized, setInitialized] = useState(false);
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [loading, setLoading] = useState(false);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [scopeRecords, setScopeRecords] = useState<ReleaseListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<ReviewSummary>({
    pending: 0,
    approved: 0,
    rejected: 0,
    mineQueue: 0,
  });
  const [reviewScope, setReviewScope] = useState<ReviewScope>('pending');
  const [keyword, setKeyword] = useState('');
  const [requestType, setRequestType] = useState<QueryRequestType>('all');
  const [status, setStatus] = useState<QueryStatus>('all');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    void bootstrap();
  }, []);

  useEffect(() => {
    if (initialized) {
      setPage(1);
    }
  }, [keyword, requestType, status, reviewScope, initialized]);

  useEffect(() => {
    if (!initialized) return;
    void loadList();
  }, [initialized, currentUser, reviewScope, keyword, requestType, status, page, pageSize]);

  useEffect(() => {
    if (!initialized) return;
    void loadSummary();
  }, [initialized, currentUser]);

  useEffect(() => {
    setStatus('all');
  }, [reviewScope]);

  const bootstrap = async () => {
    setLoading(true);
    const userRes: any = await getCurrentUserInfo({ skipErrorHandler: true }).catch((error) => error);
    if (userRes?.code === 0) {
      setCurrentUser(userRes.data);
    }
    setInitialized(true);
    setLoading(false);
  };

  const fetchReleaseSlice = async (
    statuses: ReleaseStatus[],
    baseQuery: Omit<Parameters<typeof getReleaseList>[0], 'page' | 'pageSize' | 'status'>,
  ) => {
    const querySize = Math.max(page*pageSize, pageSize);
    const responses = await Promise.all(
      statuses.map((itemStatus) =>
        getReleaseList(
          {
            ...baseQuery,
            status: itemStatus,
            page: 1,
            pageSize: querySize,
          },
          { skipErrorHandler: true },
        ).catch((error) => error),
      ),
    );

    const normalized = responses.map((response) => {
      if (!response || response.code !== 0) {
        throw new Error(response?.msg || 'Failed to load review workbench');
      }
      return normalizeReleaseList(response);
    });

    const merged = Array.from(
      new Map(
        normalized
          .flatMap((item) => item.list)
          .map((item) => [item.ID, item] as const),
      ).values(),
    ).sort((left, right) => {
      const rightTime = new Date(right.submittedAt || right.createdAt || 0).getTime();
      const leftTime = new Date(left.submittedAt || left.createdAt || 0).getTime();
      return rightTime - leftTime;
    });

    const total = normalized.reduce((sum, item) => sum + item.total, 0);
    return {
      list: merged.slice((page - 1) * pageSize, page * pageSize),
      total,
    };
  };

  const loadList = async () => {
    setLoading(true);
    try {
      if (!currentUser?.ID) {
        setScopeRecords([]);
        return;
      }

      const reviewerId = currentUser.ID;
      const baseQuery = {
        reviewerId,
        keyword: keyword.trim() || undefined,
        requestType: requestType === 'all' ? undefined : requestType,
      };
      const scopedStatuses =
        reviewScope === 'pending'
          ? [status === 'all' ? 'pending_review' : status]
          : status === 'all'
            ? (['approved', 'rejected'] as ReleaseStatus[])
            : [status];

      const result = await fetchReleaseSlice(scopedStatuses, baseQuery);
      setScopeRecords(result.list);
      setTotal(result.total);
    } catch (error: any) {
      message.error(error?.message || 'Failed to load review workbench');
      setScopeRecords([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  const countByStatus = async (itemStatus: ReleaseStatus) => {
    if (!currentUser?.ID) return 0;
    const res: any = await getReleaseList(
      {
        page: 1,
        pageSize: 1,
        reviewerId: currentUser.ID,
        status: itemStatus,
      },
      { skipErrorHandler: true },
    ).catch((error) => error);
    if (!res || res.code !== 0) return 0;
    return normalizeReleaseList(res).total;
  };

  const loadSummary = async () => {
    setSummaryLoading(true);
    try {
      if (!currentUser?.ID) {
        setSummary({ pending: 0, approved: 0, rejected: 0, mineQueue: 0 });
        return;
      }
      const [pending, approved, rejected] = await Promise.all([
        countByStatus('pending_review'),
        countByStatus('approved'),
        countByStatus('rejected'),
      ]);
      setSummary({
        pending,
        approved,
        rejected,
        mineQueue: reviewScope === 'pending' ? pending : approved + rejected,
      });
    } finally {
      setSummaryLoading(false);
    }
  };

  const columns = useMemo<ColumnsType<ReleaseListItem>>(
    () => [
      {
        title: 'Plugin',
        title: '项目',
        dataIndex: 'pluginNameZh',
        key: 'plugin',
        render: (_, record) => (
          <Space direction="vertical" size={0}>
            <Typography.Text strong>{record.pluginNameZh || '-'}</Typography.Text>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {record.pluginCode || '-'}
            </Typography.Text>
            {record.pluginNameEn ? (
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                {record.pluginNameEn}
              </Typography.Text>
            ) : null}
          </Space>
        ),
      },
      {
        title: '申请类型',
        dataIndex: 'requestType',
        width: 140,
        render: (value: ReleaseRequestType) => <Tag color="blue">{requestTypeLabel[value] || value}</Tag>,
      },
      {
        title: '版本号',
        dataIndex: 'version',
        width: 140,
        render: (value: string) => value || '-',
      },
      {
        title: '状态',
        dataIndex: 'status',
        width: 160,
        render: (value: ReleaseStatus) => <Tag color={statusMeta[value]?.color}>{statusMeta[value]?.label || value}</Tag>,
      },
      {
        title: '审核人',
        dataIndex: 'reviewerId',
        width: 110,
        render: (value: number | null | undefined) => participantLabel(value, currentUser?.ID),
      },
      {
        title: '提交时间',
        dataIndex: 'submittedAt',
        width: 180,
        render: (_, record) => record.submittedAt || record.createdAt || '-',
      },
    ],
    [currentUser?.ID],
  );

  const statusOptions = reviewScope === 'pending' ? pendingStatusOptions : reviewedStatusOptions;

  const handleRowClick = (record: ReleaseListItem) => {
    history.push({
      pathname: `/plugin/project/${record.pluginId}`,
      query: {
        releaseId: String(record.ID),
        from: 'review',
        tab: 'review',
      },
    });
  };

  return (
    <PageContainer
      title="审核工作台"
      content="这里集中处理待审核版本。点击一行会进入统一项目详情页，并自动定位到对应版本。"
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Segmented
          value={reviewScope}
          onChange={(value) => setReviewScope(value as ReviewScope)}
          options={[
            { label: '待审核', value: 'pending' },
            { label: '我已处理', value: 'reviewed' },
          ]}
        />

        <WorkbenchFilterBar
          keyword={keyword}
          onKeywordChange={setKeyword}
          requestType={requestType}
          onRequestTypeChange={(value) => setRequestType(value as QueryRequestType)}
          status={status}
          onStatusChange={(value) => setStatus(value as QueryStatus)}
          requestTypeOptions={requestTypeOptions}
          statusOptions={statusOptions}
        />

        <Row gutter={12}>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="待审核" value={summary.pending} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="我已通过" value={summary.approved} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="我已打回" value={summary.rejected} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="我的队列" value={summary.mineQueue} />
            </Card>
          </Col>
        </Row>

        <Card bordered={false} bodyStyle={{ padding: 0 }} style={cardStyle}>
          <Table<ReleaseListItem>
            rowKey="ID"
            columns={columns}
            dataSource={scopeRecords}
            loading={loading}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              showTotal: (count) => `共 ${count} 条`,
              onChange: (nextPage, nextPageSize) => {
                setPage(nextPage);
                setPageSize(nextPageSize || 10);
              },
            }}
            locale={{
              emptyText: <Empty description="暂无审核数据" />,
            }}
            scroll={{ x: 920 }}
            onRow={(record) => ({
              onClick: () => handleRowClick(record),
              style: { cursor: 'pointer' },
            })}
          />
        </Card>
      </Space>
    </PageContainer>
  );
};

export default ReviewWorkbenchPage;
