import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ModalForm, ProFormSelect, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { history, useLocation, useParams } from '@umijs/max';
import { App, Avatar, Button, Card, Col, Descriptions, Empty, Form, Input, Modal, Row, Select, Space, Steps, Tabs, Tag, Timeline, Tooltip, Typography, Upload } from 'antd';
import type { UploadFile } from 'antd/es/upload/interface';
import { ArrowLeftOutlined, CheckCircleOutlined, CopyOutlined, FileTextOutlined, InboxOutlined, PlusOutlined, RocketOutlined, StopOutlined } from '@ant-design/icons';
import { createRelease, getProjectDetail, getReleaseDetail, transitRelease, updatePlugin, uploadPluginAsset } from '@/services/api/plugin';
import { getCurrentUserInfo, getUserList } from '@/services/api/user';

type ProjectStatus = 'planning' | 'active' | 'offlined';
type RequestType = 'initial' | 'maintenance' | 'offline';
type ReleaseStatus = 'draft' | 'release_preparing' | 'pending_review' | 'approved' | 'rejected' | 'released' | 'offlined';

const requesterRoleIds = new Set([1, 9528, 10010]);
const reviewerRoleIds = new Set([1, 9528, 10013]);
const publisherRoleIds = new Set([1, 9528, 10014]);
const projectStatusMap: Record<ProjectStatus, { label: string; color: string }> = {
  planning: { label: '筹备中', color: 'gold' },
  active: { label: '已发布', color: 'success' },
  offlined: { label: '已归档', color: 'default' },
};
const releaseStatusMap: Record<ReleaseStatus, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  release_preparing: { label: '提交资料', color: 'gold' },
  pending_review: { label: '审核中', color: 'processing' },
  approved: { label: '待发布', color: 'cyan' },
  rejected: { label: '已打回', color: 'error' },
  released: { label: '已发布', color: 'success' },
  offlined: { label: '已下架', color: 'volcano' },
};
const requestTypeMap: Record<RequestType, string> = {
  initial: '首发',
  maintenance: '版本更新',
  offline: '下架申请',
};
const cardStyle: React.CSSProperties = { borderRadius: 8, border: '1px solid #e5e6eb', boxShadow: '0 1px 2px rgba(15,23,42,0.04)' };
const getTime = (v?: string) => (v ? new Date(v).getTime() : 0);
const fileName = (url?: string) => decodeURIComponent((url || '').split('?')[0].split('/').pop() || '');
const pickCompat = (source: string | undefined, token: 'HCI' | 'ACLI') => new RegExp(`${token}\\s*([0-9A-Za-z._-]+)`, 'i').exec(source || '')?.[1] || '';

const AssetUploadField: React.FC<{ value?: string; onChange?: (value?: string) => void; hint: string; accept?: string }> = ({ value, onChange, hint, accept }) => {
  const { message } = App.useApp();
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  useEffect(() => setFileList(value ? [{ uid: value, name: fileName(value), status: 'done', url: value }] : []), [value]);
  return (
    <Upload.Dragger accept={accept} maxCount={1} fileList={fileList} customRequest={async (options) => {
      const res: any = await uploadPluginAsset(options.file as File).catch((e) => e);
      if (!res || res.code !== 0) return message.error(res?.msg || '鏂囦欢涓婁紶澶辫触'), options.onError?.(new Error('upload failed'));
      const url = res?.data?.url || res?.data?.fileUrl || res?.data;
      if (!url) return message.error('涓婁紶鎴愬姛浣嗘湭杩斿洖鏂囦欢鍦板潃'), options.onError?.(new Error('missing file url'));
      setFileList([{ uid: url, name: (options.file as File).name, status: 'done', url }]);
      onChange?.(url); options.onSuccess?.(res);
    }} onRemove={() => { setFileList([]); onChange?.(undefined); return true; }}>
      <p className="ant-upload-drag-icon"><InboxOutlined /></p>
      <p className="ant-upload-text">{hint}</p>
      <p className="ant-upload-hint">仅支持上传文件，不支持填写外部链接。</p>
    </Upload.Dragger>
  );
};

const PluginProjectPage: React.FC = () => {
  const { message, modal } = App.useApp();
  const params = useParams<{ id: string }>();
  const location = useLocation();
  const query = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const projectId = Number(params.id);
  const source = query.get('from') || 'project';
  const preferredReleaseId = Number(query.get('releaseId') || 0);
  const initialTab = query.get('tab') || (source === 'review' || source === 'publish' ? 'review' : 'overview');
  const [projectForm] = Form.useForm();
  const [versionForm] = Form.useForm();
  const [offlineForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [project, setProject] = useState<any>();
  const [selectedVersionId, setSelectedVersionId] = useState<number>();
  const [currentUser, setCurrentUser] = useState<API.UserInfo>();
  const [userOptions, setUserOptions] = useState<{ label: string; value: number }[]>([]);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [versionModalOpen, setVersionModalOpen] = useState(false);
  const [offlineModalOpen, setOfflineModalOpen] = useState(false);
  const [reviewModal, setReviewModal] = useState<{ action: 'approve' | 'reject' }>();
  const [reviewComment, setReviewComment] = useState('');
  const [activeTab, setActiveTab] = useState(initialTab);
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => { if (projectId) void bootstrap(); }, [projectId]);
  useEffect(() => { if (location.search.includes('action=new-version')) setVersionModalOpen(true); }, [location.search]);
  useEffect(() => { setActiveTab(initialTab); }, [initialTab]);

  const bootstrap = async () => { await Promise.all([loadCurrentUser(), loadUsers(), loadProject()]); };
  const loadCurrentUser = async () => { const res: any = await getCurrentUserInfo({ skipErrorHandler: true }); if (res?.code === 0) setCurrentUser(res.data); };
  const loadUsers = async () => { const res: any = await getUserList({ page: 1, pageSize: 500 }); if (res?.code === 0) setUserOptions((res.data?.list || []).map((item: any) => ({ label: `${item.nickName || item.username} (${item.username})`, value: item.ID }))); };
  const loadProject = async () => {
    setLoading(true);
    const res: any = await getProjectDetail({ id: projectId }, { skipErrorHandler: true }).catch((e) => e);
    setLoading(false);
    if (!res || res.code !== 0) return message.error(res?.msg || '鍔犺浇椤圭洰璇︽儏澶辫触');
    const versions = [...(res.data?.versions || [])].sort((a, b) => getTime(b.createdAt) - getTime(a.createdAt));
    const detail = { ...res.data, versions };
    setProject(detail);
    const preferred = versions.find((item) => item.ID === preferredReleaseId) || versions.find((item) => ['pending_review', 'approved', 'rejected', 'release_preparing', 'draft'].includes(item.status)) || versions[0];
    if (preferred?.ID) { setSelectedVersionId(preferred.ID); await hydrateReleaseDetail(preferred.ID, detail); }
  };
  const hydrateReleaseDetail = async (id: number, current?: any) => {
    const res: any = await getReleaseDetail({ id }, { skipErrorHandler: true }).catch((e) => e);
    if (!res || res.code !== 0) return;
    setProject((prev: any) => { const target = current || prev; return !target ? target : { ...target, versions: target.versions.map((item: any) => item.ID === id ? { ...item, ...res.data } : item) }; });
  };

  const authorityIds = useMemo(() => { const ids = new Set<number>(); if (currentUser?.authorityId) ids.add(currentUser.authorityId); (currentUser?.authorities || []).forEach((item) => item?.authorityId && ids.add(item.authorityId)); return ids; }, [currentUser]);
  const canManageProject = useMemo(() => Array.from(authorityIds).some((id) => requesterRoleIds.has(id)), [authorityIds]);
  const canReview = useMemo(() => Array.from(authorityIds).some((id) => reviewerRoleIds.has(id)), [authorityIds]);
  const canPublish = useMemo(() => Array.from(authorityIds).some((id) => publisherRoleIds.has(id)), [authorityIds]);
  const selectedVersion = useMemo(() => (project?.versions || []).find((item: any) => item.ID === selectedVersionId) || project?.versions?.[0], [project, selectedVersionId]);
  const userMap = useMemo(() => Object.fromEntries(userOptions.map((item) => [item.value, item.label])), [userOptions]);
  const currentStep = selectedVersion ? (['draft', 'release_preparing'].includes(selectedVersion.status) ? 0 : ['pending_review', 'rejected'].includes(selectedVersion.status) ? 1 : 2) : 0;
  const stepItems = selectedVersion?.requestType === 'offline' ? [{ title: '涓嬫灦鐢宠' }, { title: '涓嬫灦瀹℃牳' }, { title: '鎵ц涓嬫灦' }] : [{ title: '鎻愪氦璧勬枡' }, { title: '瀹℃牳' }, { title: '鍙戝竷' }];
  const canSubmitRelease = !!(selectedVersion && canManageProject && currentUser?.ID === selectedVersion.createdBy && ['draft', 'release_preparing'].includes(selectedVersion.status));
  const canReviseRelease = !!(selectedVersion && canManageProject && currentUser?.ID === selectedVersion.createdBy && selectedVersion.status === 'rejected');
  const canApproveRelease = !!(selectedVersion && canReview && currentUser?.ID === selectedVersion.reviewerId && selectedVersion.status === 'pending_review');
  const canPublishRelease = !!(selectedVersion && canPublish && currentUser?.ID === selectedVersion.publisherId && selectedVersion.status === 'approved');
  const releasedVersionOptions = useMemo(() => (project?.versions || []).filter((item: any) => item.status === 'released' && !item.isOfflined).map((item: any) => ({ label: `v${item.version}`, value: item.ID })), [project]);
  const backTarget = source === 'review' ? '/plugin/review-workbench' : source === 'publish' ? '/plugin/publish-workbench' : '/plugin/center';
  const backLabel = source === 'review' ? '返回审核工作台' : source === 'publish' ? '返回发布工作台' : '返回项目管理';
  const sectionLabel = source === 'review' ? '审核工作台' : source === 'publish' ? '发布工作台' : '项目管理';
  const pageContent = source === 'review'
    ? '当前从审核工作台进入，页面已自动定位到待审核版本。你可以在项目上下文中查看资料、时间轴并完成审核。'
    : source === 'publish'
      ? '当前从发布工作台进入，页面已自动定位到待发布版本。你可以在项目上下文中查看资料、时间轴并执行发布。'
      : '项目详情页承载项目基础信息、版本管理和版本流程。先选版本，再查看对应资料、审核与发布时间轴。';

  const handleReleaseAction = async (action: string, comment?: string) => {
    if (!selectedVersion) return false;
    setActionLoading(true);
    const res: any = await transitRelease({ id: selectedVersion.ID, action, reviewComment: comment }, { skipErrorHandler: true }).catch((e) => e);
    setActionLoading(false);
    if (!res || res.code !== 0) return message.error(res?.msg || '流程操作失败'), false;
    message.success('版本流程已更新');
    await loadProject();
    return true;
  };

  const bottomActions = (
    <Space wrap style={{ marginTop: 16, paddingTop: 16, borderTop: '1px solid #f0f0f0', width: '100%', justifyContent: 'flex-end' }}>
      {canManageProject ? <Button icon={<PlusOutlined />} onClick={() => setVersionModalOpen(true)}>创建新版本</Button> : null}
      {selectedVersion?.status === 'released' && !selectedVersion?.isOfflined && canManageProject ? <Button danger icon={<StopOutlined />} onClick={() => setOfflineModalOpen(true)}>下架申请</Button> : null}
      {selectedVersion?.status === 'released' && !selectedVersion?.isOfflined && canManageProject ? <Button onClick={() => modal.info({ title: '归档功能待后端支持', content: '当前后端还没有独立的归档接口，这里先保留操作入口和交互位置。' })}>归档</Button> : null}
      {canReviseRelease ? <Button onClick={() => void handleReleaseAction('revise')}>返回筹备</Button> : null}
      {canSubmitRelease ? <Button type="primary" onClick={() => void handleReleaseAction('submit_review')}>提交资料</Button> : null}
      {canApproveRelease ? <Button danger onClick={() => { setReviewModal({ action: 'reject' }); setReviewComment(''); }}>打回</Button> : null}
      {canApproveRelease ? <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => { setReviewModal({ action: 'approve' }); setReviewComment(''); }}>审核通过</Button> : null}
      {canPublishRelease ? <Button type="primary" icon={<RocketOutlined />} onClick={() => modal.confirm({ title: selectedVersion?.requestType === 'offline' ? '执行下架' : '执行发布', content: '将先视为已完成全面扫描，再执行最终动作。', onOk: () => handleReleaseAction('release') })}>{selectedVersion?.requestType === 'offline' ? '执行下架' : '执行发布'}</Button> : null}
    </Space>
  );

  return (
    <PageContainer
      loading={loading}
      title={false}
      breadcrumb={{
        routes: [
          { path: '/plugin', breadcrumbName: '插件发布' },
          { path: backTarget, breadcrumbName: sectionLabel },
          { path: location.pathname, breadcrumbName: project?.nameZh || '项目详情' },
        ],
      }}
      content={pageContent}
    >
      {project ? (
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={8} xl={6}>
            <Space direction="vertical" size={16} style={{ width: '100%', position: 'sticky', top: 24 }}>
              <Card bordered={false} style={cardStyle} bodyStyle={{ padding: 24, textAlign: 'center' }}>
                <Button type="link" icon={<ArrowLeftOutlined />} style={{ paddingInline: 0, marginBottom: 16, float: 'left' }} onClick={() => history.push(backTarget)}>{backLabel}</Button>
                <div style={{ clear: 'both' }} />
                <Avatar shape="square" size={100} style={{ background: '#e8f3ff', color: '#1677ff', borderRadius: 24, fontSize: 36, fontWeight: 700, marginBottom: 16 }}>{(project.code || 'P').slice(0, 1).toUpperCase()}</Avatar>
                <Typography.Title level={4} style={{ margin: 0, marginBottom: 4 }}>{project.nameZh || '-'}</Typography.Title>
                <Typography.Paragraph type="secondary" style={{ marginBottom: 12 }}>{project.nameEn || '-'}</Typography.Paragraph>
                <Tag color={projectStatusMap[project.currentStatus || 'planning']?.color}>{projectStatusMap[project.currentStatus || 'planning']?.label}</Tag>
                <Descriptions column={1} size="small" style={{ marginTop: 24, textAlign: 'left' }} labelStyle={{ width: 88 }}>
                  <Descriptions.Item label="项目编码">{project.code || '-'}</Descriptions.Item>
                  <Descriptions.Item label="负责人">{project.owner || '-'}</Descriptions.Item>
                  <Descriptions.Item label="仓库地址">
                    <Space size={6}>
                      <Typography.Text ellipsis style={{ maxWidth: 150 }}>
                        {project.repositoryUrl || '-'}
                      </Typography.Text>
                      {project.repositoryUrl ? (
                        <Tooltip title="复制仓库地址">
                          <Button
                            type="text"
                            size="small"
                            icon={<CopyOutlined />}
                            onClick={() => navigator.clipboard.writeText(project.repositoryUrl)}
                          />
                        </Tooltip>
                      ) : null}
                    </Space>
                  </Descriptions.Item>
                </Descriptions>
                <Card size="small" title="中文描述" style={{ marginTop: 16 }}><Typography.Paragraph ellipsis={{ rows: 3, expandable: true, symbol: '展开' }} style={{ marginBottom: 0 }}>{project.descriptionZh || '-'}</Typography.Paragraph></Card>
                <Card size="small" title="English Description" style={{ marginTop: 12 }}><Typography.Paragraph ellipsis={{ rows: 3, expandable: true, symbol: 'Expand' }} style={{ marginBottom: 0 }}>{project.descriptionEn || '-'}</Typography.Paragraph></Card>
                {canManageProject ? <Button style={{ marginTop: 16 }} icon={<CheckCircleOutlined />} onClick={() => { projectForm.setFieldsValue(project); setProjectModalOpen(true); }}>编辑项目</Button> : null}
              </Card>
            </Space>
          </Col>
          <Col xs={24} lg={16} xl={18}>
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <Card bordered={false} style={cardStyle}>
                <Typography.Title level={4} style={{ margin: 0, marginBottom: 16 }}>版本管理</Typography.Title>
                <div style={{ background: '#f8f9fa', padding: 16, borderRadius: 8 }}>
                  <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>选择当前查看的版本</Typography.Text>
                  <Space align="center" size={16} wrap>
                    <Select
                      value={selectedVersion?.ID}
                      style={{ width: 320 }}
                      options={(project.versions || []).map((item: any) => ({
                        label: `${item.version ? `v${item.version}` : requestTypeMap[item.requestType]} / ${releaseStatusMap[item.status]?.label || item.status}`,
                        value: item.ID,
                      }))}
                      onChange={(value) => {
                        setSelectedVersionId(value);
                        void hydrateReleaseDetail(value);
                      }}
                    />
                    {selectedVersion ? (
                      <Space wrap size={8}>
                        <Tag color="blue">{requestTypeMap[selectedVersion.requestType]}</Tag>
                        <Tag color={releaseStatusMap[selectedVersion.status]?.color}>{releaseStatusMap[selectedVersion.status]?.label}</Tag>
                      </Space>
                    ) : (
                      <Tag>暂无版本</Tag>
                    )}
                  </Space>
                </div>
              </Card>
              {selectedVersion ? (
                <>
                  <Card bordered={false} style={cardStyle}>
                    <Typography.Text type="secondary">版本流程</Typography.Text>
                    <Steps current={currentStep} status={selectedVersion.status === 'rejected' ? 'error' : undefined} items={stepItems} style={{ marginTop: 16 }} />
                  </Card>
                  <Card bordered={false} style={cardStyle}>
                    <Tabs
                      activeKey={activeTab}
                      onChange={setActiveTab}
                      items={[
                        {
                          key: 'overview',
                          label: '版本概览',
                          children: (
                            <Descriptions column={2} size="small" labelStyle={{ width: 120 }}>
                              <Descriptions.Item label="版本号">v{selectedVersion.version || '-'}</Descriptions.Item>
                              <Descriptions.Item label="发布人">{selectedVersion.publisher || '-'}</Descriptions.Item>
                              <Descriptions.Item label="适配 HCI">{pickCompat(selectedVersion.versionConstraint, 'HCI') || selectedVersion.versionConstraint || '-'}</Descriptions.Item>
                              <Descriptions.Item label="适配 ACLI">{pickCompat(selectedVersion.versionConstraint, 'ACLI') || '-'}</Descriptions.Item>
                              <Descriptions.Item label="中文变更说明" span={2}>{selectedVersion.changelogZh || '-'}</Descriptions.Item>
                              <Descriptions.Item label="English Changelog" span={2}>{selectedVersion.changelogEn || '-'}</Descriptions.Item>
                            </Descriptions>
                          ),
                        },
                        {
                          key: 'files',
                          label: '文件资料',
                          children: (
                            <Row gutter={[16, 16]}>
                              {[
                                { title: '测试报告', url: selectedVersion.testReportUrl },
                                { title: 'x86 插件包', url: selectedVersion.packageX86Url },
                                { title: 'ARM 插件包', url: selectedVersion.packageArmUrl },
                              ].map((item) => (
                                <Col xs={24} md={8} key={item.title}>
                                  <Card size="small" style={cardStyle} bodyStyle={{ minHeight: 150 }}>
                                    <Space direction="vertical" size={10} style={{ width: '100%' }}>
                                      <FileTextOutlined style={{ fontSize: 20, color: '#1677ff' }} />
                                      <Typography.Text strong>{item.title}</Typography.Text>
                                      <Typography.Paragraph type="secondary" ellipsis={{ rows: 2 }} style={{ minHeight: 42 }}>
                                        {fileName(item.url) || '暂未上传'}
                                      </Typography.Paragraph>
                                      <Button href={item.url} target="_blank" disabled={!item.url}>下载文件</Button>
                                    </Space>
                                  </Card>
                                </Col>
                              ))}
                            </Row>
                          ),
                        },
                        {
                          key: 'review',
                          label: '审核与发布',
                          children: (
                            <Descriptions column={2} size="small" labelStyle={{ width: 120 }}>
                              <Descriptions.Item label="提交人">{userMap[selectedVersion.createdBy] || `#${selectedVersion.createdBy}`}</Descriptions.Item>
                              <Descriptions.Item label="提交时间">{selectedVersion.createdAt || '-'}</Descriptions.Item>
                              <Descriptions.Item label="审核人">{selectedVersion.reviewerId ? userMap[selectedVersion.reviewerId] || `#${selectedVersion.reviewerId}` : '-'}</Descriptions.Item>
                              <Descriptions.Item label="审核状态">{releaseStatusMap[selectedVersion.status]?.label || '-'}</Descriptions.Item>
                              <Descriptions.Item label="审核意见" span={2}>{selectedVersion.reviewComment || '-'}</Descriptions.Item>
                              <Descriptions.Item label="AI 扫描结果摘要" span={2}>待后端补充 AI 扫描字段，当前先保留展示区域。</Descriptions.Item>
                              <Descriptions.Item label="发布管理员">{selectedVersion.publisherId ? userMap[selectedVersion.publisherId] || `#${selectedVersion.publisherId}` : '-'}</Descriptions.Item>
                              <Descriptions.Item label="发布时间">{selectedVersion.releasedAt || selectedVersion.offlinedAt || '-'}</Descriptions.Item>
                              <Descriptions.Item label="全面扫描结果" span={2}>当前发布动作前提供确认步骤，全面扫描结果摘要待后端字段补充。</Descriptions.Item>
                            </Descriptions>
                          ),
                        },
                        {
                          key: 'timeline',
                          label: '流程时间轴',
                          children: selectedVersion.events?.length ? (
                            <Timeline
                              items={selectedVersion.events.map((item: any) => ({
                                color: item.toStatus === 'rejected' ? 'red' : ['released', 'offlined'].includes(item.toStatus) ? 'green' : 'blue',
                                children: (
                                  <Space direction="vertical" size={0}>
                                    <Typography.Text strong>{item.action} / {item.fromStatus || 'none'} → {item.toStatus}</Typography.Text>
                                    <Typography.Text type="secondary">{item.createdAt} / {userMap[item.operatorId] || `#${item.operatorId}`}</Typography.Text>
                                    {item.comment ? <Typography.Paragraph style={{ marginBottom: 0 }}>{item.comment}</Typography.Paragraph> : null}
                                  </Space>
                                ),
                              }))}
                            />
                          ) : (
                            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无流程时间轴" />
                          ),
                        },
                      ]}
                    />
                    {bottomActions}
                  </Card>
                </>
              ) : (
                <Card bordered={false} style={cardStyle}>
                  <Empty description="当前项目还没有版本" />
                </Card>
              )}
            </Space>
          </Col>
        </Row>
      ) : <Card bordered={false} style={cardStyle}><Empty description="未找到该项目" /></Card>}

      <ModalForm form={projectForm} title="编辑项目" width={820} open={projectModalOpen} initialValues={project} modalProps={{ destroyOnClose: true, onCancel: () => setProjectModalOpen(false) }} onFinish={async (values) => { const res: any = await updatePlugin({ ...values, id: projectId }, { skipErrorHandler: true }).catch((e) => e); if (!res || res.code !== 0) return message.error(res?.msg || '更新项目失败'), false; message.success('项目已更新'); setProjectModalOpen(false); await loadProject(); return true; }}>
        <ProFormText name="code" label="项目编码" rules={[{ required: true }]} />
        <ProFormText name="owner" label="负责人" rules={[{ required: true }]} />
        <ProFormText name="repositoryUrl" label="Git 仓库地址" rules={[{ required: true }]} />
        <ProFormText name="nameZh" label="项目名称（中文）" rules={[{ required: true }]} />
        <ProFormTextArea name="descriptionZh" label="中文描述" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
        <ProFormTextArea name="capabilityZh" label="中文能力说明" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
        <ProFormText name="nameEn" label="Project Name (English)" rules={[{ required: true }]} />
        <ProFormTextArea name="descriptionEn" label="English Description" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
        <ProFormTextArea name="capabilityEn" label="Capabilities (English)" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
      </ModalForm>

      <ModalForm form={versionForm} title="创建新版本" width={840} open={versionModalOpen} modalProps={{ destroyOnClose: true, onCancel: () => setVersionModalOpen(false) }} onFinish={async (values) => { const res: any = await createRelease({ ...values, pluginId: projectId }, { skipErrorHandler: true }).catch((e) => e); if (!res || res.code !== 0) return message.error(res?.msg || '保存版本失败'), false; message.success('版本已创建'); setVersionModalOpen(false); await loadProject(); return true; }}>
        <ProFormSelect name="requestType" label="版本类型" initialValue="maintenance" valueEnum={{ initial: '首发', maintenance: '版本更新' }} rules={[{ required: true }]} />
        <ProFormText name="version" label="版本号" rules={[{ required: true }]} />
        <ProFormText name="versionConstraint" label="版本限制 / 兼容信息" extra="当前后端还没有独立的 HCI / ACLI 字段，可先按 “HCI 3.1 / ACLI 1.2” 方式填写。" />
        <ProFormText name="publisher" label="发布人" rules={[{ required: true }]} />
        <ProFormSelect name="reviewerId" label="审核人" options={userOptions} rules={[{ required: true }]} />
        <ProFormSelect name="publisherId" label="发布管理员" options={userOptions} rules={[{ required: true }]} />
        <ProFormTextArea name="performanceSummaryZh" label="性能结论（中文）" fieldProps={{ rows: 3 }} />
        <ProFormTextArea name="performanceSummaryEn" label="Performance Summary (English)" fieldProps={{ rows: 3 }} />
        <Form.Item name="testReportUrl" label="测试报告" rules={[{ required: true }]}><AssetUploadField hint="上传测试报告文件" accept=".pdf,.doc,.docx,.xls,.xlsx" /></Form.Item>
        <Form.Item name="packageX86Url" label="x86 插件包" rules={[{ required: true }]}><AssetUploadField hint="上传 x86 插件包" accept=".zip,.tar,.gz,.tgz" /></Form.Item>
        <Form.Item name="packageArmUrl" label="ARM 插件包" rules={[{ required: true }]}><AssetUploadField hint="上传 ARM 插件包" accept=".zip,.tar,.gz,.tgz" /></Form.Item>
        <ProFormTextArea name="changelogZh" label="变更说明（中文）" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
        <ProFormTextArea name="changelogEn" label="Changelog (English)" fieldProps={{ rows: 3 }} rules={[{ required: true }]} />
      </ModalForm>

      <ModalForm form={offlineForm} title="发起下架申请" width={760} open={offlineModalOpen} modalProps={{ destroyOnClose: true, onCancel: () => setOfflineModalOpen(false) }} onFinish={async (values) => { const res: any = await createRelease({ ...values, pluginId: projectId, requestType: 'offline' }, { skipErrorHandler: true }).catch((e) => e); if (!res || res.code !== 0) return message.error(res?.msg || '保存下架申请失败'), false; message.success('下架申请已创建'); setOfflineModalOpen(false); await loadProject(); return true; }}>
        <ProFormSelect name="targetReleaseId" label="目标版本" options={releasedVersionOptions} rules={[{ required: true }]} />
        <ProFormSelect name="reviewerId" label="审核人" options={userOptions} rules={[{ required: true }]} />
        <ProFormSelect name="publisherId" label="执行发布管理员" options={userOptions} rules={[{ required: true }]} />
        <ProFormTextArea name="offlineReasonZh" label="下架原因（中文）" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
        <ProFormTextArea name="offlineReasonEn" label="Offline Reason (English)" fieldProps={{ rows: 4 }} rules={[{ required: true }]} />
      </ModalForm>

      <Modal
        open={!!reviewModal}
        title={reviewModal?.action === 'approve' ? '审核通过' : '打回版本'}
        confirmLoading={actionLoading}
        onCancel={() => {
          setReviewModal(undefined);
          setReviewComment('');
        }}
        onOk={async () => {
          if (reviewModal?.action === 'reject' && !reviewComment.trim()) {
            return message.warning('打回时必须填写意见');
          }
          const ok = await handleReleaseAction(reviewModal?.action || 'approve', reviewComment.trim());
          if (ok) {
            setReviewModal(undefined);
            setReviewComment('');
          }
        }}
      >
        <Typography.Paragraph>
          {reviewModal?.action === 'approve'
            ? '可填写审核备注，便于提交人和发布管理员了解审核结论。'
            : '请填写打回意见，提交人会根据这里的意见补充资料。'}
        </Typography.Paragraph>
        <Input.TextArea
          rows={4}
          value={reviewComment}
          onChange={(e) => setReviewComment(e.target.value)}
          placeholder={reviewModal?.action === 'approve' ? '请输入审核备注（可选）' : '请输入打回意见（必填）'}
        />
      </Modal>
    </PageContainer>
  );
};

export default PluginProjectPage;

