import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ModalForm, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App, Button, Card, Col, Empty, Pagination, Row, Space, Typography } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { createPlugin, getPluginList, getReleaseList, updatePlugin } from '@/services/api/plugin';
import { getCurrentUserInfo } from '@/services/api/user';
import ProjectFilterBar from '../components/ProjectFilterBar';
import ProjectSummaryCard from '../components/ProjectSummaryCard';

type ViewMode = 'card' | 'list';
type ProjectStatus = 'planning' | 'active' | 'offlined';
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
  status: ReleaseStatus;
  version?: string;
  createdAt?: string;
  isOfflined?: boolean;
};

type ProjectRecord = PluginItem & {
  releases: ReleaseItem[];
  latestWorkflow?: ReleaseItem;
  activeRelease?: ReleaseItem;
  latestReleased?: ReleaseItem;
  workflowSummary: string;
};

const requesterRoleIds = new Set([1, 9528, 10010]);

const cardStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const projectStatusMeta: Record<ProjectStatus, { label: string; color: string }> = {
  planning: { label: '筹备中', color: 'gold' },
  active: { label: '已发布', color: 'success' },
  offlined: { label: '已归档', color: 'default' },
};

const releaseStatusLabel = (status?: ReleaseStatus) => {
  if (!status) return '无进行中流程';
  if (status === 'pending_review') return '审核中';
  if (status === 'approved') return '待发布';
  if (status === 'released') return '已发布';
  if (status === 'offlined') return '已下架';
  if (status === 'rejected') return '已打回';
  return '资料提交中';
};

const getTime = (value?: string) => (value ? new Date(value).getTime() : 0);

const buildWorkflowSummary = (
  record: Pick<ProjectRecord, 'latestWorkflow' | 'activeRelease' | 'latestReleased' | 'latestVersion'>,
) => {
  const version =
    record.latestWorkflow?.version ||
    record.activeRelease?.version ||
    record.latestReleased?.version ||
    record.latestVersion ||
    '-';
  const status = releaseStatusLabel(record.latestWorkflow?.status || record.activeRelease?.status || record.latestReleased?.status);
  return `${version} ${status}`;
};

const PluginProjectCenterPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [projects, setProjects] = useState<PluginItem[]>([]);
  const [releases, setReleases] = useState<ReleaseItem[]>([]);
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [viewMode, setViewMode] = useState<ViewMode>('card');
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<ProjectStatus | 'all'>('all');
  const [ownerFilter, setOwnerFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(12);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<PluginItem>();

  useEffect(() => {
    void bootstrap();
  }, []);

  useEffect(() => {
    setPage(1);
  }, [keyword, statusFilter, ownerFilter, viewMode]);

  const bootstrap = async () => {
    await Promise.all([loadCurrentUser(), loadData()]);
  };

  const loadCurrentUser = async () => {
    const res: any = await getCurrentUserInfo({ skipErrorHandler: true }).catch((error) => error);
    if (res?.code === 0) setCurrentUser(res.data);
  };

  const loadData = async () => {
    setLoading(true);
    try {
      const [projectRes, releaseRes]: any = await Promise.all([
        getPluginList({ page: 1, pageSize: 200 }, { skipErrorHandler: true }).catch((error) => error),
        getReleaseList({ page: 1, pageSize: 500 }, { skipErrorHandler: true }).catch((error) => error),
      ]);

      if (!projectRes || projectRes.code !== 0) {
        message.error(projectRes?.msg || '加载项目列表失败');
        return;
      }
      if (!releaseRes || releaseRes.code !== 0) {
        message.error(releaseRes?.msg || '加载版本流程失败');
        return;
      }

      setProjects(projectRes.data?.list || []);
      setReleases(releaseRes.data?.list || []);
    } finally {
      setLoading(false);
    }
  };

  const authorityIds = useMemo(() => {
    const ids = new Set<number>();
    if (currentUser?.authorityId) ids.add(currentUser.authorityId);
    (currentUser?.authorities || []).forEach((item) => {
      if (item?.authorityId) {
        ids.add(item.authorityId);
      }
    });
    return ids;
  }, [currentUser]);

  const canManageProject = useMemo(
    () => Array.from(authorityIds).some((id) => requesterRoleIds.has(id)),
    [authorityIds],
  );

  const projectRecords = useMemo<ProjectRecord[]>(() => {
    return projects.map((project) => {
      const projectReleases = releases
        .filter((item) => item.pluginId === project.ID)
        .sort((left, right) => getTime(right.createdAt) - getTime(left.createdAt));
      const latestWorkflow = projectReleases[0];
      const activeRelease = projectReleases.find((item) =>
        ['draft', 'release_preparing', 'pending_review', 'approved', 'rejected'].includes(item.status),
      );
      const latestReleased = projectReleases.find((item) => item.status === 'released' && !item.isOfflined);

      return {
        ...project,
        releases: projectReleases,
        latestWorkflow,
        activeRelease,
        latestReleased,
        workflowSummary: buildWorkflowSummary({
          latestWorkflow,
          activeRelease,
          latestReleased,
          latestVersion: project.latestVersion,
        }),
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

  const filteredProjects = useMemo(() => {
    const normalizedKeyword = keyword.trim().toLowerCase();
    return projectRecords.filter((item) => {
      const keywordHit =
        !normalizedKeyword ||
        [item.code, item.nameZh, item.nameEn, item.descriptionZh, item.descriptionEn, item.owner, item.repositoryUrl]
          .join(' ')
          .toLowerCase()
          .includes(normalizedKeyword);
      const statusHit = statusFilter === 'all' || item.currentStatus === statusFilter;
      const ownerHit = ownerFilter === 'all' || item.owner === ownerFilter;
      return keywordHit && statusHit && ownerHit;
    });
  }, [keyword, ownerFilter, projectRecords, statusFilter]);

  const pagedProjects = useMemo(
    () => filteredProjects.slice((page - 1) * pageSize, page * pageSize),
    [filteredProjects, page, pageSize],
  );

  const handleOpenCreate = () => {
    setEditingProject(undefined);
    setProjectModalOpen(true);
  };

  const handleOpenEdit = (record: PluginItem) => {
    setEditingProject(record);
    setProjectModalOpen(true);
  };

  return (
    <PageContainer
      loading={loading}
      title={false}
      content="项目管理首页只展示项目入口和最近流程摘要，完整版本与流程上下文请进入项目详情页。"
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <ProjectFilterBar
          keyword={keyword}
          onKeywordChange={setKeyword}
          statusFilter={statusFilter}
          onStatusChange={setStatusFilter}
          ownerFilter={ownerFilter}
          onOwnerChange={setOwnerFilter}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          canCreateProject={canManageProject}
          onCreateProject={handleOpenCreate}
          statusOptions={[
            { label: '全部状态', value: 'all' },
            { label: '筹备中', value: 'planning' },
            { label: '已发布', value: 'active' },
            { label: '已归档', value: 'offlined' },
          ]}
          ownerOptions={[{ label: '全部负责人', value: 'all' }, ...ownerOptions]}
        />

        {!filteredProjects.length && !loading ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description="暂无符合条件的项目" />
          </Card>
        ) : viewMode === 'card' ? (
          <>
            <Row gutter={[12, 12]}>
              {pagedProjects.map((record) => (
                <Col xs={24} sm={12} md={8} lg={6} xl={6} xxl={6} key={record.ID}>
                  <div style={{ position: 'relative' }}>
                    {canManageProject ? (
                      <Button
                        type="text"
                        size="small"
                        icon={<EditOutlined />}
                        aria-label="编辑项目"
                        style={{
                          position: 'absolute',
                          top: 8,
                          right: 8,
                          zIndex: 1,
                          height: 28,
                          width: 28,
                          padding: 0,
                          background: 'rgba(255, 255, 255, 0.92)',
                          boxShadow: '0 1px 2px rgba(15, 23, 42, 0.08)',
                        }}
                        onClick={(event) => {
                          event.stopPropagation();
                          handleOpenEdit(record);
                        }}
                      />
                    ) : null}
                    <ProjectSummaryCard
                      layout="card"
                      code={record.code}
                      nameZh={record.nameZh}
                      nameEn={record.nameEn}
                      latestVersion={
                        record.latestVersion || record.activeRelease?.version || record.latestReleased?.version || '-'
                      }
                      workflowSummary={record.workflowSummary}
                      statusLabel={projectStatusMeta[record.currentStatus].label}
                      statusColor={projectStatusMeta[record.currentStatus].color}
                      onClick={() => history.push(`/plugin/project/${record.ID}`)}
                    />
                  </div>
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
          <>
            <Space direction="vertical" size={12} style={{ width: '100%' }}>
              {pagedProjects.map((record) => (
                <Space key={record.ID} align="start" size={12} style={{ width: '100%' }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <ProjectSummaryCard
                      layout="list"
                      code={record.code}
                      nameZh={record.nameZh}
                      nameEn={record.nameEn}
                      latestVersion={
                        record.latestVersion || record.activeRelease?.version || record.latestReleased?.version || '-'
                      }
                      workflowSummary={record.workflowSummary}
                      statusLabel={projectStatusMeta[record.currentStatus].label}
                      statusColor={projectStatusMeta[record.currentStatus].color}
                      onClick={() => history.push(`/plugin/project/${record.ID}`)}
                    />
                  </div>
                  {canManageProject ? (
                    <Button
                      type="link"
                      size="small"
                      style={{ paddingInline: 0, height: 'auto', flex: 'none', marginTop: 6 }}
                      onClick={() => handleOpenEdit(record)}
                    >
                      编辑
                    </Button>
                  ) : null}
                </Space>
              ))}
            </Space>

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
        )}
      </Space>

      <ModalForm
        key={editingProject?.ID || 'new-project'}
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
          if (!res || res.code !== 0) {
            message.error(res?.msg || '保存项目失败');
            return false;
          }
          message.success(editingProject ? '项目已更新' : '项目已创建');
          setEditingProject(undefined);
          setProjectModalOpen(false);
          await loadData();
          return true;
        }}
      >
        <Typography.Paragraph type="secondary">
          项目首页只维护基础信息。发布包、测试报告、变更说明等版本资料请进入项目详情页后再维护。
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
