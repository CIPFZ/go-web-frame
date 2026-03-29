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
import { buildReviewWorkbenchBaseQuery } from '../utils/pluginWorkbench';

type QueryRequestType = ReleaseRequestType | 'all';
type QueryStatus = ReleaseStatus | 'all';
type ReviewScope = 'assigned' | 'all' | 'reviewed';

type ReviewSummary = {
  assignedPending: number;
  allPending: number;
  approved: number;
  rejected: number;
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
  release_preparing: { label: '提交资料', color: 'gold' },
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
    assignedPending: 0,
    allPending: 0,
    approved: 0,
    rejected: 0,
  });
  const [reviewScope, setReviewScope] = useState<ReviewScope>('assigned');
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
    const querySize = Math.max(page * pageSize, pageSize);
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
        throw new Error(response?.msg || '加载审核工作台失败');
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
        setTotal(0);
        return;
      }

      const baseQuery = buildReviewWorkbenchBaseQuery({
        keyword,
        requestType,
        scope: reviewScope === 'all' ? 'all' : 'assigned',
        currentUserId: currentUser.ID,
      });
      const scopedStatuses =
        reviewScope === 'reviewed'
          ? status === 'all'
            ? (['approved', 'rejected'] as ReleaseStatus[])
            : [status]
          : [status === 'all' ? 'pending_review' : status];

      const result = await fetchReleaseSlice(scopedStatuses, baseQuery);
      setScopeRecords(result.list);
      setTotal(result.total);
    } catch (error: any) {
      message.error(error?.message || '加载审核工作台失败');
      setScopeRecords([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  const countByStatus = async (itemStatus: ReleaseStatus, reviewerId?: number) => {
    const res: any = await getReleaseList(
      {
        page: 1,
        pageSize: 1,
        reviewerId,
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
        setSummary({ assignedPending: 0, allPending: 0, approved: 0, rejected: 0 });
        return;
      }
      const [assignedPending, allPending, approved, rejected] = await Promise.all([
        countByStatus('pending_review', currentUser.ID),
        countByStatus('pending_review'),
        countByStatus('approved', currentUser.ID),
        countByStatus('rejected', currentUser.ID),
      ]);
      setSummary({
        assignedPending,
        allPending,
        approved,
        rejected,
      });
    } finally {
      setSummaryLoading(false);
    }
  };

  const columns = useMemo<ColumnsType<ReleaseListItem>>(
    () => [
      {
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
        title: '当前审核人',
        dataIndex: 'reviewerId',
        width: 140,
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

  const statusOptions = reviewScope === 'reviewed' ? reviewedStatusOptions : pendingStatusOptions;

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
      content="这里集中处理插件版本审核。默认优先展示分配给我的待审核任务，也可以切换查看全队列或我已处理的记录。"
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Segmented
          value={reviewScope}
          onChange={(value) => setReviewScope(value as ReviewScope)}
          options={[
            { label: '分配给我', value: 'assigned' },
            { label: '全部待审核', value: 'all' },
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
              <Statistic title="分配给我的待审核" value={summary.assignedPending} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="全队列待审核" value={summary.allPending} />
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
