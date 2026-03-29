import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { history } from '@umijs/max';
import { App, Card, Col, Empty, Row, Segmented, Space, Statistic, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import WorkbenchFilterBar from '../components/WorkbenchFilterBar';
import {
  getReleaseList,
  ReleaseListItem,
  ReleaseRequestType,
  ReleaseStatus,
} from '@/services/api/plugin';
import { getCurrentUserInfo } from '@/services/api/user';
import { buildPublishWorkbenchQuery } from '../utils/pluginWorkbench';

type QueryRequestType = ReleaseRequestType | 'all';
type QueryStatus = ReleaseStatus | 'all';
type PublishScope = 'assigned' | 'all';

type PublishSummary = {
  pendingPublish: number;
  pendingOffline: number;
  releasedToday: number;
  offlinedToday: number;
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
  { label: '待执行', value: 'approved' },
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
  release_preparing: { label: '提交资料', color: 'gold' },
  pending_review: { label: '待审核', color: 'processing' },
  approved: { label: '待执行', color: 'cyan' },
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

const isToday = (value?: string | null) => {
  if (!value) return false;
  return dayjs(value).isSame(dayjs(), 'day');
};

const pickCompletedAt = (record: ReleaseListItem) =>
  record.offlinedAt || record.releasedAt || record.approvedAt || record.createdAt || '-';

const assigneeLabel = (userId: number | null | undefined, currentUserId?: number) => {
  if (!userId) return '-';
  if (currentUserId && userId === currentUserId) return '我';
  return `#${userId}`;
};

const PublishWorkbenchPage: React.FC = () => {
  const { message } = App.useApp();
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [loading, setLoading] = useState(false);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [records, setRecords] = useState<ReleaseListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<PublishSummary>({
    pendingPublish: 0,
    pendingOffline: 0,
    releasedToday: 0,
    offlinedToday: 0,
  });
  const [scope, setScope] = useState<PublishScope>('assigned');
  const [keyword, setKeyword] = useState('');
  const [requestType, setRequestType] = useState<QueryRequestType>('all');
  const [status, setStatus] = useState<QueryStatus>('approved');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    void bootstrap();
  }, []);

  useEffect(() => {
    setPage(1);
  }, [scope, keyword, requestType, status]);

  useEffect(() => {
    if (!currentUser?.ID) return;
    void loadList();
  }, [currentUser?.ID, scope, keyword, requestType, status, page, pageSize]);

  useEffect(() => {
    if (!currentUser?.ID) return;
    void loadSummary();
  }, [currentUser?.ID, scope]);

  const bootstrap = async () => {
    const userRes: any = await getCurrentUserInfo({ skipErrorHandler: true }).catch((error) => error);
    if (userRes?.code === 0) {
      setCurrentUser(userRes.data);
    }
  };

  const query = useMemo(
    () =>
      buildPublishWorkbenchQuery({
        page,
        pageSize,
        keyword,
        requestType,
        status,
        scope,
        currentUserId: currentUser?.ID,
      }),
    [currentUser?.ID, keyword, page, pageSize, requestType, scope, status],
  );

  const loadList = async () => {
    setLoading(true);
    try {
      const res: any = await getReleaseList(query, { skipErrorHandler: true }).catch((error) => error);
      if (!res || res.code !== 0) {
        setRecords([]);
        setTotal(0);
        message.error(res?.msg || '加载发布工作台失败');
        return;
      }

      const normalized = normalizeReleaseList(res);
      setRecords(normalized.list);
      setTotal(normalized.total);
    } catch (error) {
      setRecords([]);
      setTotal(0);
      message.error('加载发布工作台失败');
    } finally {
      setLoading(false);
    }
  };

  const countReleases = async (params: Partial<Parameters<typeof getReleaseList>[0]>) => {
    const res: any = await getReleaseList(
      {
        page: 1,
        pageSize: 1,
        ...params,
      },
      { skipErrorHandler: true },
    ).catch((error) => error);

    if (!res || res.code !== 0) return 0;
    return normalizeReleaseList(res).total;
  };

  const getRecentReleases = async (itemStatus: ReleaseStatus) => {
    const res: any = await getReleaseList(
      {
        page: 1,
        pageSize: 200,
        status: itemStatus,
        publisherId: scope === 'assigned' ? currentUser?.ID : undefined,
      },
      { skipErrorHandler: true },
    ).catch((error) => error);

    if (!res || res.code !== 0) return [] as ReleaseListItem[];
    return normalizeReleaseList(res).list;
  };

  const loadSummary = async () => {
    setSummaryLoading(true);
    try {
      const publisherId = scope === 'assigned' ? currentUser?.ID : undefined;
      const [initialPending, maintenancePending, pendingOffline, releasedList, offlinedList] = await Promise.all([
        countReleases({ status: 'approved', requestType: 'initial', publisherId }),
        countReleases({ status: 'approved', requestType: 'maintenance', publisherId }),
        countReleases({ status: 'approved', requestType: 'offline', publisherId }),
        getRecentReleases('released'),
        getRecentReleases('offlined'),
      ]);

      setSummary({
        pendingPublish: initialPending + maintenancePending,
        pendingOffline,
        releasedToday: releasedList.filter(
          (item) => item.requestType !== 'offline' && isToday(item.releasedAt),
        ).length,
        offlinedToday: offlinedList.filter(
          (item) => item.requestType === 'offline' && isToday(item.offlinedAt),
        ).length,
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
        width: 140,
        render: (value: ReleaseStatus) => (
          <Tag color={statusMeta[value]?.color}>{statusMeta[value]?.label || value}</Tag>
        ),
      },
      {
        title: '当前发布人',
        dataIndex: 'publisherId',
        width: 140,
        render: (value: number | null | undefined) => assigneeLabel(value, currentUser?.ID),
      },
      {
        title: '发布人署名',
        dataIndex: 'publisher',
        width: 140,
        render: (value: string) => value || '-',
      },
      {
        title: '兼容信息',
        dataIndex: 'versionConstraint',
        width: 220,
        render: (value: string) => value || '-',
      },
      {
        title: '最近处理时间',
        dataIndex: 'completedAt',
        width: 200,
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
      content="这里集中处理审核通过后的发布和下架任务。默认优先展示分配给我的待执行版本，也可以切换查看全队列。"
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Segmented
          value={scope}
          onChange={(value) => setScope(value as PublishScope)}
          options={[
            { label: '分配给我', value: 'assigned' },
            { label: '全部待处理', value: 'all' },
          ]}
        />

        <WorkbenchFilterBar
          keyword={keyword}
          onKeywordChange={(value) => {
            setKeyword(value);
            setPage(1);
          }}
          requestType={requestType}
          onRequestTypeChange={(value) => {
            setRequestType(value as QueryRequestType);
            setPage(1);
          }}
          status={status}
          onStatusChange={(value) => {
            setStatus(value as QueryStatus);
            setPage(1);
          }}
          requestTypeOptions={requestTypeOptions}
          statusOptions={statusOptions}
        />

        <Row gutter={12}>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="待发布版本" value={summary.pendingPublish} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="待执行下架" value={summary.pendingOffline} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="今日已发布" value={summary.releasedToday} />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card bordered={false} loading={summaryLoading} style={cardStyle}>
              <Statistic title="今日已下架" value={summary.offlinedToday} />
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
            scroll={{ x: 1220 }}
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
