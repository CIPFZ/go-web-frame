import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ModalForm, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { App, Button, Card, Col, Empty, Pagination, Row, Space, Typography } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { createPlugin, getPluginList, updatePlugin } from '@/services/api/plugin';
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
  createdBy: number;
  currentStatus: ProjectStatus;
  latestVersion?: string;
  currentWorkflowId?: number;
  currentWorkflowType?: string;
  currentWorkflowStatus?: ReleaseStatus;
  currentWorkflowVersion?: string;
};

const requesterRoleIds = new Set([10010]);

const cardStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const projectStatusMeta: Record<ProjectStatus, { label: string; color: string }> = {
  planning: { label: 'Planning', color: 'gold' },
  active: { label: 'Active', color: 'success' },
  offlined: { label: 'Offlined', color: 'default' },
};

const releaseStatusLabel = (status?: ReleaseStatus) => {
  if (!status) return 'No workflow';
  if (status === 'pending_review') return 'Pending review';
  if (status === 'approved') return 'Approved';
  if (status === 'released') return 'Released';
  if (status === 'offlined') return 'Offlined';
  if (status === 'rejected') return 'Rejected';
  return 'Draft';
};

const buildWorkflowSummary = (record: Pick<PluginItem, 'currentWorkflowVersion' | 'currentWorkflowStatus' | 'latestVersion'>) => {
  const version = record.currentWorkflowVersion || record.latestVersion || '-';
  return record.currentWorkflowStatus ? `${version} ${releaseStatusLabel(record.currentWorkflowStatus)}` : 'No active workflow';
};

const PluginProjectCenterPage: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [projects, setProjects] = useState<PluginItem[]>([]);
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
    const userLoaded = await loadCurrentUser();
    if (!userLoaded) {
      setProjects([]);
      return;
    }
    await loadData();
  };

  const loadCurrentUser = async () => {
    const res: any = await getCurrentUserInfo({ skipErrorHandler: true }).catch((error) => error);
    if (res?.code === 0) {
      setCurrentUser(res.data);
      return true;
    }
    return false;
  };

  const loadData = async () => {
    setLoading(true);
    try {
      const projectRes: any = await getPluginList({ page: 1, pageSize: 200 }, { skipErrorHandler: true }).catch((error) => error);
      if (!projectRes || projectRes.code !== 0) {
        message.error(projectRes?.msg || 'Failed to load project list');
        return;
      }
      setProjects(projectRes.data?.list || []);
    } finally {
      setLoading(false);
    }
  };

  const authorityIds = useMemo(() => {
    const ids = new Set<number>();
    if (currentUser?.authorityId) ids.add(currentUser.authorityId);
    (currentUser?.authorities || []).forEach((item: any) => {
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

  const projectRecords = useMemo(
    () =>
      projects.map((project) => ({
        ...project,
        workflowSummary: buildWorkflowSummary(project),
      })),
    [projects],
  );

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

  const canEditProject = (record: PluginItem) => canManageProject && !!currentUser?.ID && record.createdBy === currentUser.ID;

  return (
    <PageContainer
      loading={loading}
      title={false}
      content="Project center shows plugin cards and a lightweight workflow summary. Open the project detail page for release-level information."
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
            { label: 'All', value: 'all' },
            { label: 'Planning', value: 'planning' },
            { label: 'Active', value: 'active' },
            { label: 'Offlined', value: 'offlined' },
          ]}
          ownerOptions={[{ label: 'All owners', value: 'all' }, ...ownerOptions]}
        />

        {!filteredProjects.length && !loading ? (
          <Card bordered={false} style={cardStyle}>
            <Empty description="No matching projects" />
          </Card>
        ) : viewMode === 'card' ? (
          <>
            <Row gutter={[12, 12]}>
              {pagedProjects.map((record) => (
                <Col xs={24} sm={12} md={8} lg={6} xl={6} xxl={6} key={record.ID}>
                  <div style={{ position: 'relative' }}>
                    {canEditProject(record) ? (
                      <Button
                        type="text"
                        size="small"
                        icon={<EditOutlined />}
                        aria-label="Edit project"
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
                      latestVersion={record.latestVersion || record.currentWorkflowVersion || '-'}
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
                      latestVersion={record.latestVersion || record.currentWorkflowVersion || '-'}
                      workflowSummary={record.workflowSummary}
                      statusLabel={projectStatusMeta[record.currentStatus].label}
                      statusColor={projectStatusMeta[record.currentStatus].color}
                      onClick={() => history.push(`/plugin/project/${record.ID}`)}
                    />
                  </div>
                  {canEditProject(record) ? (
                    <Button
                      type="link"
                      size="small"
                      style={{ paddingInline: 0, height: 'auto', flex: 'none', marginTop: 6 }}
                      onClick={() => handleOpenEdit(record)}
                    >
                      Edit
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
        title={editingProject ? 'Edit project' : 'Create project'}
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
            message.error(res?.msg || 'Failed to save project');
            return false;
          }
          message.success(editingProject ? 'Project updated' : 'Project created');
          setEditingProject(undefined);
          setProjectModalOpen(false);
          await loadData();
          return true;
        }}
      >
        <Typography.Paragraph type="secondary">
          Project info here is intentionally lightweight. Release packages, test reports, and changelog details belong on the detail page.
        </Typography.Paragraph>
        <Row gutter={16}>
          <Col span={12}>
            <ProFormText name="code" label="Project code" rules={[{ required: true }]} />
          </Col>
          <Col span={12}>
            <ProFormText name="owner" label="Owner" rules={[{ required: true }]} />
          </Col>
        </Row>
        <ProFormText name="repositoryUrl" label="Git repository URL" rules={[{ required: true }]} />
        <Card size="small" title="Chinese" style={{ marginBottom: 16 }}>
          <ProFormText name="nameZh" label="Project name (zh)" rules={[{ required: true }]} />
          <ProFormTextArea name="descriptionZh" label="Short description (zh)" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
          <ProFormTextArea name="capabilityZh" label="Capabilities (zh)" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
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
