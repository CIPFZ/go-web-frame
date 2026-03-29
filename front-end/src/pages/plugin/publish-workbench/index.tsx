import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { history } from '@umijs/max';
import { App, Card, Col, Empty, Row, Space, Statistic, Table, Tag, Typography } from 'antd';
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

type PublishSummary = {
  approved: number;
  released: number;
  offlined: number;
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

const statusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '待发布', value: 'approved' },
  { label: '已发布', value: 'released' },
  { label: '已下架', value: 'offlined' },
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

const pickCompletedAt = (record: ReleaseListItem) => {
  if (record.releasedAt) return record.releasedAt;
  if (record.approvedAt) return record.approvedAt;
  if (record.offlinedAt) return record.offlinedAt;
  return record.createdAt || '-';
};

const PublishWorkbenchPage: React.FC = () => {
  const { message } = App.useApp();
  const [initialized, setInitialized] = useState(false);
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [loading, setLoading] = useState(false);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [records, setRecords] = useState<ReleaseListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<PublishSummary>({
    approved: 0,
    released: 0,
    offlined: 0,
    mineQueue: 0,
  });
  const [keyword, setKeyword] = useState('');
  const [requestType, setRequestType] = useState<QueryRequestType>('all');
  const [status, setStatus] = useState<QueryStatus>('approved');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    void bootstrap();
  }, []);

  useEffect(() => {
    if (initialized) {
      setPage(1);
    }
  }, [keyword, requestType, status, initialized]);

  useEffect(() => {
    if (!initialized) return;
    void loadList();
  }, [initialized, currentUser, keyword, requestType, status, page, pageSize]);

  useEffect(() => {
    if (!initialized) return;
    void loadSummary();
  }, [initialized, currentUser]);

  const bootstrap = async () => {
    setLoading(true);
    const userRes: any = await getCurrentUserInfo({ skipErrorHandler: true }).catch((error) => error);
    if (userRes?.code === 0) {
      setCurrentUser(userRes.data);
    }
    setInitialized(true);
    setLoading(false);
  };

  const buildQuery = () => {
    const trimmedKeyword = keyword.trim();
    return {
      page,
      pageSize,
      keyword: trimmedKeyword || undefined,
      requestType: requestType === 'all' ? undefined : requestType,
      status: status === 'all' ? undefined : status,
      publisherId: currentUser?.ID || undefined,
    };
  };

  const loadList = async () => {
    setLoading(true);
    try {
      const res: any = await getReleaseList(buildQuery(), { skipErrorHandler: true }).catch((error) => error);
      if (!res || res.code !== 0) {
        setRecords([]);
        setTotal(0);
        message.error(res?.msg || 'Failed to load publish workbench');
        return;
      }
      const normalized = normalizeReleaseList(res);
      setRecords(normalized.list);
      setTotal(normalized.total);
    } catch (error) {
      setRecords([]);
      setTotal(0);
      message.error('Failed to load publish workbench');
    } finally {
      setLoading(false);
    }
  };

  const countReleases = async (params: Partial<Parameters<typeof getReleaseList>[0]>) => {
    if (!currentUser?.ID) return 0;
    const res: any = await getReleaseList(
      {
        page: 1,
        pageSize: 1,
        publisherId: currentUser.ID,
        ...params,
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
        setSummary({ approved: 0, released: 0, offlined: 0, mineQueue: 0 });
        return;
      }
      const [approved, released, offlined, mineQueue] = await Promise.all([
        countReleases({ status: 'approved' }),
        countReleases({ status: 'released' }),
        countReleases({ status: 'offlined' }),
        countReleases({}),
      ]);
      setSummary({ approved, released, offlined, mineQueue });
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
        title: '发布人',
        dataIndex: 'publisherId',
        width: 140,
        render: (_, record) => record.publisher || participantLabel(record.publisherId, currentUser?.ID),
      },
      {
        title: '版本限制',
        dataIndex: 'versionConstraint',
        width: 200,
        render: (value: string) => value || '-',
      },
      {
        title: '完成时间',
        dataIndex: 'completedAt',
        width: 180,
        render: (_, record) => pickCompletedAt(record),
      },
    ],
    [currentUser?.ID],
  );

  const handleRowClick = (record: ReleaseListItem) => {
    history.push({
      pathname: `/plugin/project/${record.pluginId}`,
      query: {
        releaseId: String(record.ID),
        from: 'publish',
        tab: 'review',
      },
    });
  };

  return (
    <PageContainer
      title="发布工作台"
      content="这里集中处理待发布和待下架执行的版本。点击一行会进入统一项目详情页，并自动定位到对应版本。"
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
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
              <Statistic title="待发布" value={summary.approved} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="我已发布" value={summary.released} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="我已下架" value={summary.offlined} />
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
            dataSource={records}
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
              emptyText: <Empty description="暂无发布数据" />,
            }}
            scroll={{ x: 1040 }}
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

export default PublishWorkbenchPage;
