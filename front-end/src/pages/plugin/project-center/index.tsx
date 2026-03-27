import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ModalForm, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import {
  App,
  Avatar,
  Button,
  Card,
  Col,
  Empty,
  Input,
  Pagination,
  Row,
  Segmented,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  AppstoreOutlined,
  CopyOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  SearchOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import { createPlugin, getPluginList, getReleaseList, updatePlugin } from '@/services/api/plugin';
import { getCurrentUserInfo } from '@/services/api/user';

type ViewMode = 'card' | 'list';
type QuickFilter =
  | 'all'
  | 'mine'
  | 'prepare'
  | 'my_reviewing'
  | 'my_published'
  | 'pending_review'
  | 'pending_offline'
  | 'reviewed'
  | 'pending_publish'
  | 'published_all';
type ProjectPhase = 'planning' | 'reviewing' | 'published' | 'archived';
type ProjectStatus = 'planning' | 'active' | 'offlined';
type RequestType = 'initial' | 'maintenance' | 'offline';
type ReleaseStatus =
  | 'draft'
  | 'release_preparing'
  | 'pending_review'
  | 'approved'
  | 'rejected'
  | 'released'
  | 'offlined';

type PluginItem = {
  ID: number;
  code: string;
  repositoryUrl: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  capabilityZh: string;
  capabilityEn: string;
  owner: string;
  currentStatus: ProjectStatus;
  latestVersion?: string;
};

type ReleaseItem = {
  ID: number;
  pluginId: number;
  requestType: RequestType;
  status: ReleaseStatus;
  version?: string;
  versionConstraint?: string;
  createdBy: number;
  reviewerId?: number;
  publisherId?: number;
  isOfflined?: boolean;
  createdAt?: string;
};

type ProjectRecord = PluginItem & {
  releases: ReleaseItem[];
  activeRelease?: ReleaseItem;
  latestReleased?: ReleaseItem;
  phase: ProjectPhase;
  hciVersion?: string;
  acliVersion?: string;
};

const requesterRoleIds = new Set([1, 9528, 10010]);
const reviewerRoleIds = new Set([1, 9528, 10013]);
const publisherRoleIds = new Set([1, 9528, 10014]);

const phaseMeta: Record<ProjectPhase, { label: string; color: string }> = {
  planning: { label: '筹备中', color: 'gold' },
  reviewing: { label: '审核中', color: 'processing' },
  published: { label: '已发布', color: 'success' },
  archived: { label: '已归档', color: 'default' },
};

const cardStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const getTime = (value?: string) => (value ? new Date(value).getTime() : 0);

const extractCompatVersion = (constraint: string | undefined, token: 'HCI' | 'ACLI') =>
  new RegExp(`${token}\\s*([0-9A-Za-z._-]+)`, 'i').exec(constraint || '')?.[1] || '';

const workflowLabel = (status?: ReleaseStatus) => {
  if (!status) return '无进行中流程';
  if (status === 'pending_review') return '审核';
  if (status === 'approved') return '发布';
  if (status === 'offlined') return '下架';
  return '提交资料';
};

const buildProjectPhase = (project: PluginItem, releases: ReleaseItem[]): ProjectPhase => {
  const activeRelease = releases.find((item) =>
    ['draft', 'release_preparing', 'pending_review', 'approved', 'rejected'].includes(item.status),
  );
  const publishedRelease = releases.find((item) => item.status === 'released' && !item.isOfflined);

  if (project.currentStatus === 'offlined' || (!publishedRelease && releases.some((item) => item.status === 'offlined'))) {
    return 'archived';
  }
  if (activeRelease && ['pending_review', 'approved'].includes(activeRelease.status)) {
    return 'reviewing';
  }
  if (publishedRelease || project.currentStatus === 'active') {
    return 'published';
  }
  return 'planning';
};

const PluginProjectCenterPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [projects, setProjects] = useState<PluginItem[]>([]);
  const [releases, setReleases] = useState<ReleaseItem[]>([]);
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [viewMode, setViewMode] = useState<ViewMode>('card');
  const [quickFilter, setQuickFilter] = useState<QuickFilter>('all');
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<ProjectPhase | 'all'>('all');
  const [ownerFilter, setOwnerFilter] = useState('all');
  const [hciFilter, setHciFilter] = useState('all');
  const [acliFilter, setAcliFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(12);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<PluginItem>();

  useEffect(() => {
    void bootstrap();
  }, []);

  useEffect(() => {
    setPage(1);
  }, [keyword, statusFilter, ownerFilter, hciFilter, acliFilter, quickFilter, viewMode]);

  const bootstrap = async () => {
    await Promise.all([loadCurrentUser(), loadData()]);
  };

  const loadCurrentUser = async () => {
    const res: any = await getCurrentUserInfo({ skipErrorHandler: true });
    if (res?.code === 0) setCurrentUser(res.data);
  };

  const loadData = async () => {
    setLoading(true);
    const [projectRes, releaseRes]: any = await Promise.all([
      getPluginList({ page: 1, pageSize: 200 }, { skipErrorHandler: true }).catch((error) => error),
      getReleaseList({ page: 1, pageSize: 500 }, { skipErrorHandler: true }).catch((error) => error),
    ]);
    setLoading(false);
    if (!projectRes || projectRes.code !== 0) return message.error(projectRes?.msg || '加载项目列表失败');
    if (!releaseRes || releaseRes.code !== 0) return message.error(releaseRes?.msg || '加载版本流程失败');
    setProjects(projectRes.data?.list || []);
    setReleases(releaseRes.data?.list || []);
  };

  const authorityIds = useMemo(() => {
    const ids = new Set<number>();
    if (currentUser?.authorityId) ids.add(currentUser.authorityId);
    (currentUser?.authorities || []).forEach((item) => item?.authorityId && ids.add(item.authorityId));
    return ids;
  }, [currentUser]);

  const canManageProject = useMemo(
    () => Array.from(authorityIds).some((id) => requesterRoleIds.has(id)),
    [authorityIds],
  );
  const canReview = useMemo(() => Array.from(authorityIds).some((id) => reviewerRoleIds.has(id)), [authorityIds]);
  const canPublish = useMemo(() => Array.from(authorityIds).some((id) => publisherRoleIds.has(id)), [authorityIds]);

  const projectRecords = useMemo<ProjectRecord[]>(() => {
    return projects.map((project) => {
      const projectReleases = releases
        .filter((item) => item.pluginId === project.ID)
        .sort((left, right) => getTime(right.createdAt) - getTime(left.createdAt));
      const activeRelease = projectReleases.find((item) =>
        ['draft', 'release_preparing', 'pending_review', 'approved', 'rejected'].includes(item.status),
      );
      const latestReleased = projectReleases.find((item) => item.status === 'released' && !item.isOfflined);
      const versionConstraint =
        activeRelease?.versionConstraint || latestReleased?.versionConstraint || projectReleases[0]?.versionConstraint;

      return {
        ...project,
        releases: projectReleases,
        activeRelease,
        latestReleased,
        phase: buildProjectPhase(project, projectReleases),
        hciVersion: extractCompatVersion(versionConstraint, 'HCI'),
        acliVersion: extractCompatVersion(versionConstraint, 'ACLI'),
      };
    });
  }, [projects, releases]);

  const ownerOptions = useMemo(
    () =>
      Array.from(new Set(projectRecords.map((item) => item.owner).filter(Boolean))).map((item) => ({
        label: item,
        value: item,
      })),
    [projectRecords],
  );

  const hciOptions = useMemo(
    () =>
      Array.from(new Set(projectRecords.map((item) => item.hciVersion).filter(Boolean))).map((item) => ({
        label: item,
        value: item,
      })),
    [projectRecords],
  );

  const acliOptions = useMemo(
    () =>
      Array.from(new Set(projectRecords.map((item) => item.acliVersion).filter(Boolean))).map((item) => ({
        label: item,
        value: item,
      })),
    [projectRecords],
  );

  const isMyProject = (record: ProjectRecord) =>
    record.releases.some((item) => item.createdBy === currentUser?.ID) || record.owner === currentUser?.username;

  const roleCards = useMemo(() => {
    if (canManageProject) {
      return [
        { key: 'mine' as QuickFilter, title: '我的项目', value: projectRecords.filter(isMyProject).length },
        {
          key: 'prepare' as QuickFilter,
          title: '待我完善（筹备中）',
          value: projectRecords.filter((item) => isMyProject(item) && item.phase === 'planning').length,
        },
        {
          key: 'my_reviewing' as QuickFilter,
          title: '审核中（我提交的）',
          value: releases.filter((item) => item.createdBy === currentUser?.ID && item.status === 'pending_review').length,
        },
        {
          key: 'my_published' as QuickFilter,
          title: '已发布（我的）',
          value: releases.filter(
            (item) => item.createdBy === currentUser?.ID && item.status === 'released' && !item.isOfflined,
          ).length,
        },
      ];
    }
    if (canReview) {
      return [
        {
          key: 'pending_review' as QuickFilter,
          title: '待审核（新的申请）',
          value: releases.filter(
            (item) =>
              item.status === 'pending_review' &&
              item.requestType !== 'offline' &&
              item.reviewerId === currentUser?.ID,
          ).length,
        },
        {
          key: 'pending_offline' as QuickFilter,
          title: '待审核（下架申请）',
          value: releases.filter(
            (item) =>
              item.status === 'pending_review' &&
              item.requestType === 'offline' &&
              item.reviewerId === currentUser?.ID,
          ).length,
        },
        {
          key: 'reviewed' as QuickFilter,
          title: '已审核（通过/打回）',
          value: releases.filter(
            (item) =>
              item.reviewerId === currentUser?.ID &&
              ['approved', 'rejected', 'released', 'offlined'].includes(item.status),
          ).length,
        },
      ];
    }
    if (canPublish) {
      return [
        {
          key: 'pending_publish' as QuickFilter,
          title: '待发布（已通过审核）',
          value: releases.filter((item) => item.status === 'approved' && item.publisherId === currentUser?.ID).length,
        },
        {
          key: 'published_all' as QuickFilter,
          title: '已发布（全部项目）',
          value: projectRecords.filter((item) => item.phase === 'published').length,
        },
      ];
    }
    return [{ key: 'all' as QuickFilter, title: '全部项目', value: projectRecords.length }];
  }, [canManageProject, canPublish, canReview, currentUser?.ID, projectRecords, releases]);

  const filteredProjects = useMemo(() => {
    const normalizedKeyword = keyword.trim().toLowerCase();
    return projectRecords.filter((item) => {
      const keywordHit =
        !normalizedKeyword ||
        [
          item.code,
          item.nameZh,
          item.nameEn,
          item.descriptionZh,
          item.descriptionEn,
          item.owner,
          item.repositoryUrl,
        ]
          .join(' ')
          .toLowerCase()
          .includes(normalizedKeyword);

      const statusHit = statusFilter === 'all' || item.phase === statusFilter;
      const ownerHit = ownerFilter === 'all' || item.owner === ownerFilter;
      const hciHit = hciFilter === 'all' || item.hciVersion === hciFilter;
      const acliHit = acliFilter === 'all' || item.acliVersion === acliFilter;

      if (!keywordHit || !statusHit || !ownerHit || !hciHit || !acliHit) {
        return false;
      }

      if (canManageProject && !canReview && !canPublish && !isMyProject(item)) {
        return false;
      }

      switch (quickFilter) {
        case 'mine':
          return isMyProject(item);
        case 'prepare':
          return isMyProject(item) && item.phase === 'planning';
        case 'my_reviewing':
          return item.releases.some((release) => release.createdBy === currentUser?.ID && release.status === 'pending_review');
        case 'my_published':
          return item.releases.some(
            (release) => release.createdBy === currentUser?.ID && release.status === 'released' && !release.isOfflined,
          );
        case 'pending_review':
          return item.releases.some(
            (release) =>
              release.status === 'pending_review' &&
              release.requestType !== 'offline' &&
              release.reviewerId === currentUser?.ID,
          );
        case 'pending_offline':
          return item.releases.some(
            (release) =>
              release.status === 'pending_review' &&
              release.requestType === 'offline' &&
              release.reviewerId === currentUser?.ID,
          );
        case 'reviewed':
          return item.releases.some(
            (release) =>
              release.reviewerId === currentUser?.ID &&
              ['approved', 'rejected', 'released', 'offlined'].includes(release.status),
          );
        case 'pending_publish':
          return item.releases.some(
            (release) => release.status === 'approved' && release.publisherId === currentUser?.ID,
          );
        case 'published_all':
          return item.phase === 'published';
        default:
          return true;
      }
    });
  }, [
    acliFilter,
    canManageProject,
    canPublish,
    canReview,
    currentUser?.ID,
    hciFilter,
    keyword,
    ownerFilter,
    projectRecords,
    quickFilter,
    statusFilter,
  ]);

  const pagedProjects = useMemo(
    () => filteredProjects.slice((page - 1) * pageSize, page * pageSize),
    [filteredProjects, page, pageSize],
  );

  const handleCopy = async (value: string) => {
    await navigator.clipboard.writeText(value);
    message.success('仓库地址已复制');
  };

  const listColumns: ColumnsType<ProjectRecord> = [
    {
      title: '序号',
      width: 72,
      render: (_, __, index) => (page - 1) * pageSize + index + 1,
    },
    {
      title: '项目名称',
      width: 260,
      render: (_, record) => (
        <Space align="start" size={12}>
          <Avatar shape="square" size={40} style={{ background: '#e8f3ff', color: '#1677ff', borderRadius: 10 }}>
            {(record.code || 'P').slice(0, 1).toUpperCase()}
          </Avatar>
          <div>
            <Typography.Text strong>{record.nameZh || '-'}</Typography.Text>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {record.nameEn || '-'}
            </Typography.Paragraph>
          </div>
        </Space>
      ),
    },
    {
      title: '简短描述',
      render: (_, record) => (
        <div>
          <Typography.Paragraph ellipsis={{ rows: 1 }} style={{ marginBottom: 4 }}>
            {record.descriptionZh || '-'}
          </Typography.Paragraph>
          <Typography.Paragraph ellipsis={{ rows: 1 }} type="secondary" style={{ marginBottom: 0 }}>
            {record.descriptionEn || '-'}
          </Typography.Paragraph>
        </div>
      ),
    },
    { title: '项目编码', dataIndex: 'code', width: 120 },
    {
      title: '仓库地址',
      width: 220,
      render: (_, record) => (
        <Space size={4}>
          <Typography.Text ellipsis style={{ maxWidth: 150 }}>
            {record.repositoryUrl || '-'}
          </Typography.Text>
          {record.repositoryUrl ? (
            <Tooltip title="复制仓库地址">
              <Button
                type="text"
                size="small"
                icon={<CopyOutlined />}
                onClick={(event) => {
                  event.stopPropagation();
                  void handleCopy(record.repositoryUrl);
                }}
              />
            </Tooltip>
          ) : null}
        </Space>
      ),
    },
    { title: '负责人', dataIndex: 'owner', width: 120 },
    {
      title: '最新版本',
      width: 120,
      render: (_, record) => record.latestVersion || record.activeRelease?.version || '-',
    },
    {
      title: '项目状态',
      width: 120,
      render: (_, record) => <Tag color={phaseMeta[record.phase].color}>{phaseMeta[record.phase].label}</Tag>,
    },
    {
      title: '当前流程',
      width: 120,
      render: (_, record) => workflowLabel(record.activeRelease?.status),
    },
    {
      title: '适配 HCI',
      width: 120,
      render: (_, record) => record.hciVersion || '-',
    },
    {
      title: '操作',
      width: 220,
      render: (_, record) => (
        <Space
          onClick={(event) => {
            event.stopPropagation();
          }}
        >
          <Button type="link" icon={<EyeOutlined />} onClick={() => history.push(`/plugin/project/${record.ID}`)}>
            查看项目
          </Button>
          {canManageProject ? (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => {
                setEditingProject(record);
                setProjectModalOpen(true);
              }}
            >
              编辑
            </Button>
          ) : null}
          {canManageProject ? (
            <Button
              type="link"
              icon={<PlusOutlined />}
              onClick={() => history.push(`/plugin/project/${record.ID}?action=new-version`)}
            >
              创建版本
            </Button>
          ) : null}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false} content="项目管理列表：浏览、筛选和管理插件项目。点击项目后进入详情页查看所有版本与流程。">
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Card bordered={false} style={cardStyle}>
          <Row gutter={[16, 16]}>
            {roleCards.map((item) => (
              <Col xs={12} md={8} xl={6} key={item.key}>
                <Card
                  hoverable
                  bordered={quickFilter === item.key}
                  style={{
                    ...cardStyle,
                    borderColor: quickFilter === item.key ? '#1677ff' : '#e5e6eb',
                  }}
                  bodyStyle={{ padding: 18 }}
                  onClick={() => setQuickFilter(item.key)}
                >
                  <Typography.Text type="secondary">{item.title}</Typography.Text>
                  <Typography.Title level={3} style={{ margin: '8px 0 0' }}>
                    {item.value}
                  </Typography.Title>
                </Card>
              </Col>
            ))}
          </Row>
        </Card>

        <Card bordered={false} style={cardStyle}>
          <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space wrap size={12}>
              <Input
                allowClear
                prefix={<SearchOutlined />}
                placeholder="搜索中文 / 英文项目名、编码、仓库、负责人"
                style={{ width: 300 }}
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
              />
              <Select
                style={{ width: 150 }}
                value={statusFilter}
                onChange={setStatusFilter}
                options={[
                  { label: '全部状态', value: 'all' },
                  { label: '筹备中', value: 'planning' },
                  { label: '审核中', value: 'reviewing' },
                  { label: '已发布', value: 'published' },
                  { label: '已归档', value: 'archived' },
                ]}
              />
              <Select
                style={{ width: 150 }}
                value={ownerFilter}
                onChange={setOwnerFilter}
                options={[{ label: '全部负责人', value: 'all' }, ...ownerOptions]}
              />
              <Select
                style={{ width: 150 }}
                value={hciFilter}
                onChange={setHciFilter}
                options={[{ label: '适配 HCI 版本', value: 'all' }, ...hciOptions]}
              />
              <Select
                style={{ width: 150 }}
                value={acliFilter}
                onChange={setAcliFilter}
                options={[{ label: '适配 ACLI 版本', value: 'all' }, ...acliOptions]}
              />
            </Space>
            <Space>
              <Segmented
                value={viewMode}
                onChange={(value) => setViewMode(value as ViewMode)}
                options={[
                  { label: <AppstoreOutlined />, value: 'card' },
                  { label: <UnorderedListOutlined />, value: 'list' },
                ]}
              />
              {canManageProject ? (
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => {
                    setEditingProject(undefined);
                    setProjectModalOpen(true);
                  }}
                >
                  新建项目
                </Button>
              ) : null}
            </Space>
          </Space>
        </Card>

        {!filteredProjects.length && !loading ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description="暂无符合条件的项目" />
          </Card>
        ) : viewMode === 'card' ? (
          <>
            <Row gutter={[12, 12]}>
              {pagedProjects.map((record) => (
                <Col xs={24} sm={12} md={8} lg={6} xl={6} xxl={6} key={record.ID}>
                  <Card
                    hoverable
                    bordered={false}
                    style={{
                      ...cardStyle,
                      height: '100%',
                      transition: 'transform .2s ease, box-shadow .2s ease',
                      cursor: 'pointer',
                    }}
                    bodyStyle={{ padding: 14, display: 'flex', flexDirection: 'column', height: '100%', minHeight: 220 }}
                    onClick={() => history.push(`/plugin/project/${record.ID}`)}
                  >
                    <Space wrap size={[6, 6]} style={{ marginBottom: 12 }}>
                      <Tag color="blue" bordered={false} style={{ margin: 0 }}>
                        {record.hciVersion ? `HCI ${record.hciVersion}` : 'HCI -'}
                      </Tag>
                      <Tag color="cyan" bordered={false} style={{ margin: 0 }}>
                        {record.acliVersion ? `ACLI ${record.acliVersion}` : 'ACLI -'}
                      </Tag>
                    </Space>

                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 10 }}>
                      <Avatar
                        shape="square"
                        size={48}
                        style={{
                          background: '#e8f3ff',
                          color: '#1677ff',
                          borderRadius: 12,
                          fontSize: 20,
                          fontWeight: 700,
                        }}
                      >
                        {(record.code || 'P').slice(0, 1).toUpperCase()}
                      </Avatar>
                      <Tag color={phaseMeta[record.phase].color} style={{ margin: 0 }}>
                        {phaseMeta[record.phase].label}
                      </Tag>
                    </div>

                    <div style={{ marginBottom: 8 }}>
                      <Typography.Title level={5} style={{ margin: 0, fontSize: 16, lineHeight: 1.4 }} ellipsis>
                        {record.nameZh || '-'}
                      </Typography.Title>
                      <Typography.Paragraph type="secondary" style={{ marginBottom: 0, minHeight: 20, fontSize: 12 }} ellipsis>
                        {record.nameEn || '-'}
                      </Typography.Paragraph>
                    </div>

                    <div style={{ minHeight: 40, marginBottom: 10 }}>
                      <Typography.Paragraph ellipsis={{ rows: 1 }} style={{ marginBottom: 2, fontSize: 13 }}>
                        {record.descriptionZh || '-'}
                      </Typography.Paragraph>
                      <Typography.Paragraph ellipsis={{ rows: 1 }} type="secondary" style={{ marginBottom: 0, fontSize: 12 }}>
                        {record.descriptionEn || '-'}
                      </Typography.Paragraph>
                    </div>

                    <div
                      style={{
                        marginTop: 'auto',
                        paddingTop: 10,
                        borderTop: '1px solid #f0f0f0',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                          最新版本
                        </Typography.Text>
                        <div style={{ fontWeight: 600, fontSize: 12 }}>
                          {record.latestVersion || record.activeRelease?.version || '-'}
                        </div>
                      </div>
                      <Button
                        type="link"
                        size="small"
                        style={{ paddingInline: 0, height: 'auto' }}
                        onClick={(event) => {
                          event.stopPropagation();
                          history.push(`/plugin/project/${record.ID}`);
                        }}
                      >
                        查看
                      </Button>
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>

            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
              <Pagination
                current={page}
                pageSize={pageSize}
                total={filteredProjects.length}
                showSizeChanger
                onChange={(nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                }}
              />
            </div>
          </>
        ) : (
          <Card bordered={false} style={cardStyle} bodyStyle={{ padding: 0 }}>
            <Table<ProjectRecord>
              rowKey="ID"
              loading={loading}
              columns={listColumns}
              dataSource={pagedProjects}
              pagination={false}
              onRow={(record) => ({
                onClick: () => history.push(`/plugin/project/${record.ID}`),
                style: { cursor: 'pointer' },
              })}
            />
            <div style={{ display: 'flex', justifyContent: 'flex-end', padding: 16 }}>
              <Pagination
                current={page}
                pageSize={pageSize}
                total={filteredProjects.length}
                showSizeChanger
                onChange={(nextPage, nextPageSize) => {
                  setPage(nextPage);
                  setPageSize(nextPageSize);
                }}
              />
            </div>
          </Card>
        )}
      </Space>

      <ModalForm
        title={editingProject ? '编辑项目' : '新建项目'}
        width={840}
        open={projectModalOpen}
        initialValues={editingProject}
        modalProps={{
          destroyOnClose: true,
          onCancel: () => {
            setEditingProject(undefined);
            setProjectModalOpen(false);
          },
        }}
        onFinish={async (values) => {
          const api = editingProject ? updatePlugin : createPlugin;
          const payload = editingProject ? { ...values, id: editingProject.ID } : values;
          const res: any = await api(payload).catch((error) => error);
          if (!res || res.code !== 0) return message.error(res?.msg || '保存项目失败'), false;
          message.success(editingProject ? '项目已更新' : '项目已创建');
          setEditingProject(undefined);
          setProjectModalOpen(false);
          await loadData();
          return true;
        }}
      >
        <Typography.Paragraph type="secondary">
          项目层只维护基础信息。发布包、测试报告、变更说明等版本资料，请进入项目详情后通过“创建新版本”完成。
        </Typography.Paragraph>
        <Row gutter={16}>
          <Col span={12}>
            <ProFormText name="code" label="项目编码" rules={[{ required: true }]} />
          </Col>
          <Col span={12}>
            <ProFormText name="owner" label="负责人" rules={[{ required: true }]} />
          </Col>
        </Row>
        <ProFormText name="repositoryUrl" label="Git 仓库地址" rules={[{ required: true }]} />
        <Card size="small" title="中文信息" style={{ marginBottom: 16 }}>
          <ProFormText name="nameZh" label="项目名称（中文）" rules={[{ required: true }]} />
          <ProFormTextArea name="descriptionZh" label="简短描述（中文）" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
          <ProFormTextArea name="capabilityZh" label="项目能力（中文）" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
        </Card>
        <Card size="small" title="English Info">
          <ProFormText name="nameEn" label="Project Name (English)" rules={[{ required: true }]} />
          <ProFormTextArea
            name="descriptionEn"
            label="Short Description (English)"
            fieldProps={{ rows: 3 }}
            rules={[{ required: true }]}
          />
          <ProFormTextArea
            name="capabilityEn"
            label="Capabilities (English)"
            fieldProps={{ rows: 4 }}
            rules={[{ required: true }]}
          />
        </Card>
      </ModalForm>
    </PageContainer>
  );
};

export default PluginProjectCenterPage;
