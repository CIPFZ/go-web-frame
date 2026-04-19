import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { getLocale, history, useParams } from '@umijs/max';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Form,
  Input,
  Modal,
  Row,
  Select,
  Space,
  Timeline,
  Typography,
  Upload,
  message,
} from 'antd';
import { DeleteOutlined, PlusOutlined, UploadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd/es/upload/interface';
import dayjs from 'dayjs';

import {
  createRelease,
  getProductList,
  getProjectDetail,
  transitionRelease,
  updateRelease,
  type CompatibleInfo,
  type CompatibleProductItem,
  type PluginReleaseItem,
  type ProductItem,
  type ProjectDetail,
} from '@/services/api/plugin';
import {
  getDisplayChangelog,
  getDisplayDescription,
  getDisplayName,
  getDisplayOfflineReason,
  isEnglishLocale,
  pickLocaleText,
} from '@/utils/plugin';
import { ProcessStatusTag, ReleaseStatusTag, RequestTypeTag } from '../components/status';

type CompatibleFormItem = {
  productId?: number;
  versionConstraint?: string;
};

type ReleaseFormValues = {
  version: string;
  productCompatibles?: CompatibleFormItem[];
  acliVersionConstraint?: string;
  testReportUrl?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  changelogZh?: string;
  changelogEn?: string;
};

type OfflineFormValues = {
  tdId?: string;
  offlineReasonZh: string;
  offlineReasonEn?: string;
};

const FILE_UPLOAD_ACTION = '/api/v1/sys/file/upload';

const copyMap = {
  zh: {
    back: '返回项目列表',
    code: '插件编码',
    department: '归属部门',
    repository: '仓库地址',
    description: '插件描述',
    releaseList: '发布单列表',
    newRelease: '新建发布单',
    currentRelease: '当前发布单',
    editRelease: '编辑发布单',
    submitReview: '提交审核',
    resubmit: '重新提审',
    requestOffline: '发起下架',
    searchVersion: '搜索版本号',
    noRelease: '未找到发布单',
    version: '版本号',
    requestType: '请求类型',
    lifecycle: '生命周期',
    workflow: '工单状态',
    createdAt: '创建时间',
    changelog: '变更说明',
    offlineReason: '下架原因',
    compatibility: '兼容信息',
    productCompatibility: '产品兼容',
    acliCompatibility: 'aCLI 兼容',
    notConfigured: '未配置',
    testReport: '测试报告',
    selectRelease: '请选择发布单查看',
    timeline: '流转时间线',
    noTimelineNote: '无补充说明',
    edit: '编辑',
    releaseModalTitleCreate: '新建发布单',
    releaseModalTitleEdit: '编辑发布单',
    firstReleaseHint: '首个发布单的版本必须为 1.0.0。',
    versionPlaceholder: '请输入版本号，例如 1.0.0',
    uploadTestReport: '上传测试报告',
    uploadX86: '上传 x86 安装包',
    uploadArm: '上传 ARM 安装包',
    changelogZh: '变更说明（中文）',
    changelogEn: '变更说明（英文）',
    addProductCompatibility: '新增产品兼容',
    remove: '移除',
    requestOfflineTitle: '发起下架',
    offlineReasonZh: '下架原因（中文）',
    offlineReasonEn: '下架原因（英文）',
    tdId: 'TD ID',
    saveFailed: '保存失败',
    actionFailed: '操作失败',
    loadFailed: '加载项目详情失败',
    uploadFailed: '上传失败',
    uploadSucceeded: '上传成功',
    compatibilityRequired: '请至少配置一项产品兼容或填写 aCLI 兼容。',
    productLabel: '产品',
    versionConstraintLabel: '版本约束',
    acliPlaceholder: '请输入 aCLI 兼容版本，例如 >= 1.0.0',
  },
  en: {
    back: 'Back to Projects',
    code: 'Code',
    department: 'Department',
    repository: 'Repository',
    description: 'Description',
    releaseList: 'Release List',
    newRelease: 'New Release',
    currentRelease: 'Current Release',
    editRelease: 'Edit Release',
    submitReview: 'Submit Review',
    resubmit: 'Resubmit',
    requestOffline: 'Request Offline',
    searchVersion: 'Search Version',
    noRelease: 'No releases found',
    version: 'Version',
    requestType: 'Request Type',
    lifecycle: 'Lifecycle',
    workflow: 'Workflow',
    createdAt: 'Created At',
    changelog: 'Changelog',
    offlineReason: 'Offline Reason',
    compatibility: 'Compatibility',
    productCompatibility: 'Product Compatibility',
    acliCompatibility: 'aCLI Compatibility',
    notConfigured: 'Not configured',
    testReport: 'Test Report',
    selectRelease: 'Select a release to inspect',
    timeline: 'Timeline',
    noTimelineNote: 'No additional note',
    edit: 'Edit',
    releaseModalTitleCreate: 'New Release',
    releaseModalTitleEdit: 'Edit Release',
    firstReleaseHint: 'The first release must use version 1.0.0.',
    versionPlaceholder: 'Enter version, for example 1.0.0',
    uploadTestReport: 'Upload Test Report',
    uploadX86: 'Upload x86 Package',
    uploadArm: 'Upload ARM Package',
    changelogZh: 'Changelog (ZH)',
    changelogEn: 'Changelog (EN)',
    addProductCompatibility: 'Add Product Compatibility',
    remove: 'Remove',
    requestOfflineTitle: 'Request Offline',
    offlineReasonZh: 'Offline Reason (ZH)',
    offlineReasonEn: 'Offline Reason (EN)',
    tdId: 'TD ID',
    saveFailed: 'Failed to save release',
    actionFailed: 'Action failed',
    loadFailed: 'Failed to load project detail',
    uploadFailed: 'Upload failed',
    uploadSucceeded: 'Upload succeeded',
    compatibilityRequired: 'Configure at least one product compatibility or fill in the aCLI compatibility.',
    productLabel: 'Product',
    versionConstraintLabel: 'Version Constraint',
    acliPlaceholder: 'Enter aCLI version compatibility, for example >= 1.0.0',
  },
};

const timelineActionCopy: Record<string, { zh: string; en: string; shortZh: string; shortEn: string }> = {
  create: { zh: '创建发布单', en: 'Created release', shortZh: '创建', shortEn: 'Create' },
  submit_review: { zh: '提交审核', en: 'Submitted for review', shortZh: '提审', shortEn: 'Submit' },
  approve: { zh: '审核通过', en: 'Approved', shortZh: '通过', shortEn: 'Approve' },
  reject: { zh: '审核驳回', en: 'Rejected', shortZh: '驳回', shortEn: 'Reject' },
  revise: { zh: '重新提审', en: 'Resubmitted', shortZh: '重提', shortEn: 'Resubmit' },
  release: { zh: '执行发布', en: 'Released', shortZh: '发布', shortEn: 'Release' },
  request_offline: { zh: '发起下架', en: 'Requested offline', shortZh: '申请下架', shortEn: 'Request Offline' },
  offline: { zh: '执行下架', en: 'Offlined', shortZh: '下架', shortEn: 'Offline' },
  claim: { zh: '认领工单', en: 'Claimed', shortZh: '认领', shortEn: 'Claim' },
  reset: { zh: '重置工单', en: 'Reset', shortZh: '重置', shortEn: 'Reset' },
};

const getToken = () => localStorage.getItem('token') || '';

const formatTime = (value?: string) => (value ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '-');

const isEditableRelease = (release?: PluginReleaseItem) =>
  Boolean(release && (release.editable ?? [1, 4].includes(release.status)));

const normalizeCompatibility = (release?: PluginReleaseItem): CompatibleInfo => {
  if (!release) return { universal: false, products: [], acli: [] };
  if (release.compatibility) {
    return {
      universal: Boolean(release.compatibility.universal),
      products: release.compatibility.products || [],
      acli: release.compatibility.acli || [],
    };
  }
  if (release.compatibleInfo) {
    return {
      universal: Boolean(release.compatibleInfo.universal),
      products: release.compatibleInfo.products || [],
      acli: release.compatibleInfo.acli || [],
    };
  }
  return {
    universal: false,
    products: release.compatibleItems || [],
    acli: [],
  };
};

const normalizeCompatibleFormItems = (items?: CompatibleProductItem[]) =>
  (items || []).map((item) => ({
    productId: item.productId,
    versionConstraint: item.versionConstraint,
  }));

const getTimelineActionText = (locale: string, action: string) => {
  const item = timelineActionCopy[action];
  if (!item) return action;
  return isEnglishLocale(locale) ? item.en : item.zh;
};

const getTimelineActionLabel = (locale: string, action: string) => {
  const item = timelineActionCopy[action];
  if (!item) return action;
  return isEnglishLocale(locale) ? item.shortEn : item.shortZh;
};

const renderCompatibilityBlock = (
  title: string,
  emptyText: string,
  items?: CompatibleProductItem[],
) => (
  <Space direction="vertical" size={4} style={{ width: '100%' }}>
    <Typography.Text strong>{title}</Typography.Text>
    {items?.length ? (
      items.map((item) => (
        <Typography.Text key={`${title}-${item.productId}-${item.versionConstraint || 'any'}`}>
          {(item.productName || item.productCode) + (item.versionConstraint ? ` ${item.versionConstraint}` : '')}
        </Typography.Text>
      ))
    ) : (
      <Typography.Text type="secondary">{emptyText}</Typography.Text>
    )}
  </Space>
);

const FileUploadField: React.FC<{
  value?: string;
  onChange?: (value?: string) => void;
  buttonText: string;
  uploadFailedText: string;
  uploadSuccessText: string;
}> = ({ value, onChange, buttonText, uploadFailedText, uploadSuccessText }) => {
  const [uploading, setUploading] = useState(false);

  const props: UploadProps = {
    action: FILE_UPLOAD_ACTION,
    headers: { 'x-token': getToken() },
    showUploadList: false,
    onChange: (info) => {
      if (info.file.status === 'uploading') {
        setUploading(true);
        return;
      }
      if (info.file.status === 'done') {
        setUploading(false);
        const url = info.file.response?.data?.url;
        if (!url) {
          message.error(info.file.response?.msg || uploadFailedText);
          return;
        }
        onChange?.(url);
        message.success(uploadSuccessText);
      }
      if (info.file.status === 'error') {
        setUploading(false);
        message.error(uploadFailedText);
      }
    },
  };

  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Upload {...props}>
        <Button icon={<UploadOutlined />} loading={uploading}>
          {buttonText}
        </Button>
      </Upload>
      {value ? (
        <Typography.Link href={value} target="_blank">
          {value}
        </Typography.Link>
      ) : (
        <Typography.Text type="secondary">-</Typography.Text>
      )}
    </Space>
  );
};

const ProductCompatibilityEditor: React.FC<{
  form: any;
  title: string;
  addText: string;
  removeText: string;
  productLabel: string;
  versionConstraintLabel: string;
  options: Array<{ label: string; value: number }>;
}> = ({ form, title, addText, removeText, productLabel, versionConstraintLabel, options }) => (
  <Form.List name="productCompatibles">
    {(fields, { add, remove }) => (
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Typography.Text strong>{title}</Typography.Text>
          <Button
            type="text"
            icon={<PlusOutlined />}
            aria-label={addText}
            onClick={() => {
              add({ productId: undefined, versionConstraint: '' });
              queueMicrotask(() => form.validateFields(['productCompatibles', 'acliVersionConstraint']));
            }}
          />
        </Space>
        {fields.map((field) => (
          <Row gutter={12} key={field.key}>
            <Col span={10}>
              <Form.Item {...field} name={[field.name, 'productId']} label={productLabel} rules={[{ required: true }]}>
                <Select options={options} />
              </Form.Item>
            </Col>
            <Col span={10}>
              <Form.Item {...field} name={[field.name, 'versionConstraint']} label={versionConstraintLabel}>
                <Input placeholder=">= 1.0.0" />
              </Form.Item>
            </Col>
            <Col span={4} style={{ display: 'flex', alignItems: 'center' }}>
              <Button
                type="text"
                icon={<DeleteOutlined />}
                aria-label={`${removeText}-${field.name}`}
                onClick={() => remove(field.name)}
              />
            </Col>
          </Row>
        ))}
      </Space>
    )}
  </Form.List>
);

const PluginProjectDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const projectId = Number(params.id);
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const [detail, setDetail] = useState<ProjectDetail>();
  const [loading, setLoading] = useState(false);
  const [versionKeyword, setVersionKeyword] = useState('');
  const [releaseModalOpen, setReleaseModalOpen] = useState(false);
  const [offlineModalOpen, setOfflineModalOpen] = useState(false);
  const [editingRelease, setEditingRelease] = useState<PluginReleaseItem>();
  const [products, setProducts] = useState<ProductItem[]>([]);
  const [releaseForm] = Form.useForm<ReleaseFormValues>();
  const [offlineForm] = Form.useForm<OfflineFormValues>();

  const selectedRelease = detail?.selectedRelease;
  const compatible = normalizeCompatibility(selectedRelease);

  const productOptions = useMemo(
    () => products.map((item) => ({ label: `${item.code} / ${item.name}`, value: item.ID })),
    [products],
  );

  const acliProduct = useMemo(
    () => products.find((item) => /acli/i.test(item.code || '') || /acli/i.test(item.name || '')),
    [products],
  );

  const filteredReleases = useMemo(() => {
    const keyword = versionKeyword.trim().toLowerCase();
    return (detail?.releases || []).filter((release) =>
      keyword ? release.version?.toLowerCase().includes(keyword) : true,
    );
  }, [detail?.releases, versionKeyword]);

  const loadProducts = async () => {
    if (products.length) return;
    const res = await getProductList({ page: 1, pageSize: 999 });
    if (res.code === 0) setProducts(res.data?.list || []);
  };

  const loadDetail = async (releaseId?: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const res = await getProjectDetail({ id: projectId, releaseId });
      if (res.code !== 0) {
        message.error(res.msg || copy.loadFailed);
        return;
      }
      setDetail(res.data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadDetail();
    void loadProducts();
  }, [projectId]);

  const openCreateRelease = async () => {
    setEditingRelease(undefined);
    releaseForm.resetFields();
    releaseForm.setFieldsValue({
      version: '1.0.0',
      productCompatibles: [{ productId: undefined, versionConstraint: '' }],
      acliVersionConstraint: '',
    });
    await loadProducts();
    setReleaseModalOpen(true);
  };

  const openEditRelease = async (release: PluginReleaseItem) => {
    if (!isEditableRelease(release)) return;
    setEditingRelease(release);
    await loadProducts();
    const normalized = normalizeCompatibility(release);
    releaseForm.setFieldsValue({
      version: release.version,
      productCompatibles: normalizeCompatibleFormItems(normalized.products).length
        ? normalizeCompatibleFormItems(normalized.products)
        : [{ productId: undefined, versionConstraint: '' }],
      acliVersionConstraint: normalized.acli?.[0]?.versionConstraint || '',
      testReportUrl: release.testReportUrl,
      packageX86Url: release.packageX86Url,
      packageArmUrl: release.packageArmUrl,
      changelogZh: release.changelogZh,
      changelogEn: release.changelogEn,
    });
    setReleaseModalOpen(true);
  };

  const submitReleaseForm = async () => {
    const values = await releaseForm.validateFields();
    const productCompatibles = (values.productCompatibles || []).filter((item) => item.productId);
    const acliCompatibles =
      values.acliVersionConstraint && acliProduct?.ID
        ? [{ productId: acliProduct.ID, versionConstraint: values.acliVersionConstraint }]
        : [];

    if (!productCompatibles.length && !acliCompatibles.length) {
      message.error(copy.compatibilityRequired);
      return;
    }

    const payload = {
      ...(editingRelease?.ID ? { id: editingRelease.ID } : { pluginId: projectId, requestType: 1 }),
      version: values.version,
      testReportUrl: values.testReportUrl,
      packageX86Url: values.packageX86Url,
      packageArmUrl: values.packageArmUrl,
      changelogZh: values.changelogZh,
      changelogEn: values.changelogEn,
      compatibleItems: [...productCompatibles, ...acliCompatibles],
      compatibility: {
        products: productCompatibles,
        acli: acliCompatibles,
      },
    };

    const res = editingRelease?.ID ? await updateRelease(payload) : await createRelease(payload);
    if (res.code !== 0) {
      message.error(res.msg || copy.saveFailed);
      return;
    }
    setReleaseModalOpen(false);
    await loadDetail(editingRelease?.ID || res.data?.ID);
  };

  const handleTransition = async (action: string, extra?: Record<string, any>) => {
    if (!selectedRelease) return;
    const res = await transitionRelease({ id: selectedRelease.ID, action, ...(extra || {}) });
    if (res.code !== 0) {
      message.error(res.msg || copy.actionFailed);
      return;
    }
    if (offlineModalOpen) setOfflineModalOpen(false);
    await loadDetail(selectedRelease.ID);
  };

  const openOfflineModal = () => {
    if (!selectedRelease) return;
    offlineForm.setFieldsValue({
      tdId: selectedRelease.tdId,
      offlineReasonZh: selectedRelease.offlineReasonZh,
      offlineReasonEn: selectedRelease.offlineReasonEn,
    });
    setOfflineModalOpen(true);
  };

  return (
    <PageContainer
      title={false}
      onBack={() => history.push('/plugin/project-management')}
      extra={[
        <Button key="back" onClick={() => history.push('/plugin/project-management')}>
          {copy.back}
        </Button>,
      ]}
    >
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <Card loading={loading} style={{ borderRadius: 24 }}>
          {detail?.plugin ? (
            <Descriptions column={1} title={getDisplayName(locale, detail.plugin)}>
              <Descriptions.Item label={copy.code}>{detail.plugin.code}</Descriptions.Item>
              <Descriptions.Item label={copy.department}>
                {pickLocaleText(
                  locale,
                  detail.plugin.departmentNameZh || detail.plugin.department,
                  detail.plugin.departmentNameEn || detail.plugin.department,
                )}
              </Descriptions.Item>
              <Descriptions.Item label={copy.repository}>
                <Typography.Link href={detail.plugin.repositoryUrl} target="_blank">
                  {detail.plugin.repositoryUrl}
                </Typography.Link>
              </Descriptions.Item>
              <Descriptions.Item label={copy.description}>
                {getDisplayDescription(locale, detail.plugin)}
              </Descriptions.Item>
            </Descriptions>
          ) : (
            <Empty />
          )}
        </Card>

        <Row gutter={[20, 20]} align="stretch">
          <Col xs={24} xl={9}>
            <Card
              title={copy.releaseList}
              extra={
                <Button type="primary" onClick={openCreateRelease}>
                  {copy.newRelease}
                </Button>
              }
              style={{ borderRadius: 24, height: '100%' }}
              styles={{ body: { maxHeight: 720, overflowY: 'auto' } }}
            >
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                <Input.Search
                  placeholder={copy.searchVersion}
                  value={versionKeyword}
                  onChange={(event) => setVersionKeyword(event.target.value)}
                  allowClear
                />
                {filteredReleases.length ? (
                  filteredReleases.map((release) => {
                    const active = release.ID === selectedRelease?.ID;
                    return (
                      <Card
                        key={release.ID}
                        data-testid={`release-card-${release.ID}`}
                        size="small"
                        hoverable
                        onClick={() => void loadDetail(release.ID)}
                        style={{
                          borderRadius: 18,
                          borderColor: active ? '#1677ff' : '#d9d9d9',
                          background: active ? '#f5f9ff' : '#fff',
                        }}
                        styles={{ body: { padding: 16 } }}
                      >
                        <Space direction="vertical" size={10} style={{ width: '100%' }}>
                          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                            <Typography.Text strong>{release.version || '-'}</Typography.Text>
                            <ReleaseStatusTag status={release.status} locale={locale} />
                          </Space>
                          <Space wrap>
                            <RequestTypeTag type={release.requestType} locale={locale} />
                            <ProcessStatusTag status={release.processStatus} locale={locale} />
                          </Space>
                          <Typography.Text type="secondary">{formatTime(release.createdAt)}</Typography.Text>
                          {isEditableRelease(release) ? (
                            <Button
                              size="small"
                              onClick={(event) => {
                                event.stopPropagation();
                                void openEditRelease(release);
                              }}
                            >
                              {copy.edit}
                            </Button>
                          ) : null}
                        </Space>
                      </Card>
                    );
                  })
                ) : (
                  <Empty description={copy.noRelease} />
                )}
              </Space>
            </Card>
          </Col>

          <Col xs={24} xl={15}>
            <Space direction="vertical" size={20} style={{ width: '100%' }}>
              <Card
                title={copy.currentRelease}
                style={{ borderRadius: 24 }}
                extra={
                  selectedRelease ? (
                    <Space wrap>
                      {isEditableRelease(selectedRelease) ? (
                        <Button onClick={() => void openEditRelease(selectedRelease)}>{copy.editRelease}</Button>
                      ) : null}
                      {selectedRelease.status === 1 && (
                        <Button type="primary" onClick={() => void handleTransition('submit_review')}>
                          {copy.submitReview}
                        </Button>
                      )}
                      {selectedRelease.status === 4 && (
                        <Button type="primary" onClick={() => void handleTransition('revise')}>
                          {copy.resubmit}
                        </Button>
                      )}
                      {selectedRelease.status === 5 && (
                        <Button danger onClick={openOfflineModal}>
                          {copy.requestOffline}
                        </Button>
                      )}
                    </Space>
                  ) : undefined
                }
              >
                {selectedRelease ? (
                  <Descriptions column={1} styles={{ label: { width: 120 } }}>
                    <Descriptions.Item label={copy.version}>{selectedRelease.version || '-'}</Descriptions.Item>
                    <Descriptions.Item label={copy.requestType}>
                      <RequestTypeTag type={selectedRelease.requestType} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.lifecycle}>
                      <ReleaseStatusTag status={selectedRelease.status} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.workflow}>
                      <ProcessStatusTag status={selectedRelease.processStatus} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.createdAt}>{formatTime(selectedRelease.createdAt)}</Descriptions.Item>
                    <Descriptions.Item label={copy.changelog}>
                      {getDisplayChangelog(locale, selectedRelease)}
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.offlineReason}>
                      {getDisplayOfflineReason(locale, selectedRelease)}
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.compatibility}>
                      <Space direction="vertical" size={12} style={{ width: '100%' }}>
                        {renderCompatibilityBlock(copy.productCompatibility, copy.notConfigured, compatible.products)}
                        {renderCompatibilityBlock(copy.acliCompatibility, copy.notConfigured, compatible.acli)}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label={copy.testReport}>
                      {selectedRelease.testReportUrl ? (
                        <Typography.Link href={selectedRelease.testReportUrl} target="_blank">
                          {selectedRelease.testReportUrl}
                        </Typography.Link>
                      ) : (
                        '-'
                      )}
                    </Descriptions.Item>
                    <Descriptions.Item label="x86">
                      {selectedRelease.packageX86Url ? (
                        <Typography.Link href={selectedRelease.packageX86Url} target="_blank">
                          {selectedRelease.packageX86Url}
                        </Typography.Link>
                      ) : (
                        '-'
                      )}
                    </Descriptions.Item>
                    <Descriptions.Item label="ARM">
                      {selectedRelease.packageArmUrl ? (
                        <Typography.Link href={selectedRelease.packageArmUrl} target="_blank">
                          {selectedRelease.packageArmUrl}
                        </Typography.Link>
                      ) : (
                        '-'
                      )}
                    </Descriptions.Item>
                  </Descriptions>
                ) : (
                  <Empty description={copy.selectRelease} />
                )}
              </Card>

              <Card title={copy.timeline} style={{ borderRadius: 24 }}>
                <div style={{ maxHeight: 360, overflowY: 'auto', paddingRight: 8 }}>
                  <Timeline
                    items={(detail?.events || []).map((item) => ({
                      children: (
                        <div style={{ padding: '4px 0 10px', borderBottom: '1px solid #f0f0f0' }}>
                          <Space direction="vertical" size={4} style={{ width: '100%' }}>
                            <Space style={{ justifyContent: 'space-between', width: '100%' }} align="start">
                              <Typography.Text strong>{getTimelineActionLabel(locale, item.action)}</Typography.Text>
                              <Typography.Text type="secondary">{formatTime(item.createdAt)}</Typography.Text>
                            </Space>
                            <Typography.Text>{getTimelineActionText(locale, item.action)}</Typography.Text>
                            <Typography.Paragraph type="secondary" style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                              {item.comment || copy.noTimelineNote}
                            </Typography.Paragraph>
                          </Space>
                        </div>
                      ),
                    }))}
                  />
                </div>
              </Card>
            </Space>
          </Col>
        </Row>
      </Space>

      <Modal
        title={editingRelease?.ID ? copy.releaseModalTitleEdit : copy.releaseModalTitleCreate}
        open={releaseModalOpen}
        onCancel={() => setReleaseModalOpen(false)}
        onOk={() => void submitReleaseForm()}
        width={900}
        forceRender
        destroyOnHidden
      >
        <Form form={releaseForm} layout="vertical">
          <Form.Item name="version" label={copy.version} rules={[{ required: true }]} extra={copy.firstReleaseHint}>
            <Input placeholder={copy.versionPlaceholder} disabled={Boolean(editingRelease?.ID)} />
          </Form.Item>

          <ProductCompatibilityEditor
            form={releaseForm}
            title={copy.productCompatibility}
            addText={copy.addProductCompatibility}
            removeText={copy.remove}
            productLabel={copy.productLabel}
            versionConstraintLabel={copy.versionConstraintLabel}
            options={productOptions}
          />

          <Form.Item name="acliVersionConstraint" label={copy.acliCompatibility} style={{ marginTop: 16 }}>
            <Input placeholder={copy.acliPlaceholder} />
          </Form.Item>

          <Form.Item name="testReportUrl" label={copy.testReport}>
            <FileUploadField
              buttonText={copy.uploadTestReport}
              uploadFailedText={copy.uploadFailed}
              uploadSuccessText={copy.uploadSucceeded}
            />
          </Form.Item>
          <Form.Item name="packageX86Url" label="x86">
            <FileUploadField
              buttonText={copy.uploadX86}
              uploadFailedText={copy.uploadFailed}
              uploadSuccessText={copy.uploadSucceeded}
            />
          </Form.Item>
          <Form.Item name="packageArmUrl" label="ARM">
            <FileUploadField
              buttonText={copy.uploadArm}
              uploadFailedText={copy.uploadFailed}
              uploadSuccessText={copy.uploadSucceeded}
            />
          </Form.Item>
          <Form.Item name="changelogZh" label={copy.changelogZh}>
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="changelogEn" label={copy.changelogEn}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={copy.requestOfflineTitle}
        open={offlineModalOpen}
        onCancel={() => setOfflineModalOpen(false)}
        onOk={async () => {
          const values = await offlineForm.validateFields();
          await handleTransition('request_offline', values);
        }}
        forceRender
        destroyOnHidden
      >
        <Form form={offlineForm} layout="vertical">
          <Form.Item name="tdId" label={copy.tdId}>
            <Input placeholder={copy.tdId} />
          </Form.Item>
          <Form.Item name="offlineReasonZh" label={copy.offlineReasonZh} rules={[{ required: true }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="offlineReasonEn" label={copy.offlineReasonEn}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default PluginProjectDetailPage;
