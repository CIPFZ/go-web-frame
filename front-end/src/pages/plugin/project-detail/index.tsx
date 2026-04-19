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
  message,
} from 'antd';
import dayjs from 'dayjs';

import {
  createRelease,
  getProductList,
  getProjectDetail,
  transitionRelease,
  updateRelease,
  type CompatibleProductItem,
  type PluginReleaseItem,
  type ProductItem,
  type ProjectDetail,
} from '@/services/api/plugin';
import { getDisplayChangelog, getDisplayDescription, getDisplayName, getDisplayOfflineReason, isEnglishLocale } from '@/utils/plugin';
import { ProcessStatusTag, ReleaseStatusTag, RequestTypeTag } from '../components/status';

type ReleaseFormValues = {
  version: string;
  testReportUrl?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  changelogZh?: string;
  changelogEn?: string;
  tdId?: string;
  compatibleItems?: Array<{ productId: number; versionConstraint?: string }>;
};

type OfflineFormValues = {
  tdId?: string;
  offlineReasonZh: string;
  offlineReasonEn?: string;
};

const formatTime = (value?: string) => (value ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '-');

const renderCompatibles = (items?: CompatibleProductItem[]) => {
  if (!items?.length) {
    return <Typography.Text type="secondary">-</Typography.Text>;
  }
  return (
    <Space wrap>
      {items.map((item) => (
        <Typography.Text key={`${item.productId}-${item.versionConstraint || 'any'}`}>
          {item.productName || item.productCode}
          {item.versionConstraint ? ` ${item.versionConstraint}` : ''}
        </Typography.Text>
      ))}
    </Space>
  );
};

const PluginProjectDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const projectId = Number(params.id);
  const locale = getLocale();
  const isEnglish = isEnglishLocale(locale);
  const [detail, setDetail] = useState<ProjectDetail>();
  const [loading, setLoading] = useState(false);
  const [releaseModalOpen, setReleaseModalOpen] = useState(false);
  const [offlineModalOpen, setOfflineModalOpen] = useState(false);
  const [editingRelease, setEditingRelease] = useState<PluginReleaseItem>();
  const [products, setProducts] = useState<ProductItem[]>([]);
  const [releaseForm] = Form.useForm<ReleaseFormValues>();
  const [offlineForm] = Form.useForm<OfflineFormValues>();

  const selectedRelease = detail?.selectedRelease;

  const productOptions = useMemo(
    () => products.map((item) => ({ label: `${item.code} / ${item.name}`, value: item.ID })),
    [products],
  );

  const loadProducts = async () => {
    if (products.length) return;
    const res = await getProductList({ page: 1, pageSize: 999 });
    if (res.code === 0) {
      setProducts(res.data?.list || []);
    }
  };

  const loadDetail = async (releaseId?: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const res = await getProjectDetail({ id: projectId, releaseId });
      if (res.code !== 0) {
        message.error(res.msg || 'Failed to load project detail');
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
    releaseForm.setFieldsValue({ compatibleItems: [] });
    await loadProducts();
    setReleaseModalOpen(true);
  };

  const openEditRelease = async (release: PluginReleaseItem) => {
    setEditingRelease(release);
    await loadProducts();
    releaseForm.setFieldsValue({
      version: release.version,
      testReportUrl: release.testReportUrl,
      packageX86Url: release.packageX86Url,
      packageArmUrl: release.packageArmUrl,
      changelogZh: release.changelogZh,
      changelogEn: release.changelogEn,
      tdId: release.tdId,
      compatibleItems: (release.compatibleItems || []).map((item) => ({
        productId: item.productId,
        versionConstraint: item.versionConstraint,
      })),
    });
    setReleaseModalOpen(true);
  };

  const submitReleaseForm = async () => {
    const values = await releaseForm.validateFields();
    const payload = editingRelease?.ID ? { id: editingRelease.ID, ...values } : { pluginId: projectId, requestType: 1, ...values };
    const res = editingRelease?.ID ? await updateRelease(payload) : await createRelease(payload);
    if (res.code !== 0) {
      message.error(res.msg || 'Failed to save release');
      return;
    }
    setReleaseModalOpen(false);
    await loadDetail(editingRelease?.ID || res.data?.ID);
  };

  const handleTransition = async (action: string, extra?: Record<string, any>) => {
    if (!selectedRelease) return;
    const res = await transitionRelease({ id: selectedRelease.ID, action, ...(extra || {}) });
    if (res.code !== 0) {
      message.error(res.msg || 'Action failed');
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

  const submitOfflineRequest = async () => {
    const values = await offlineForm.validateFields();
    await handleTransition('request_offline', values);
  };

  return (
    <PageContainer
      title={false}
      onBack={() => history.push('/plugin/project-management')}
      extra={[
        <Button key="back" onClick={() => history.push('/plugin/project-management')}>
          {isEnglish ? 'Back to Projects' : '返回项目列表'}
        </Button>,
      ]}
    >
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <Card loading={loading} style={{ borderRadius: 24 }}>
          {detail?.plugin ? (
            <Descriptions column={2} title={getDisplayName(locale, detail.plugin)}>
              <Descriptions.Item label={isEnglish ? 'Code' : '插件编码'}>{detail.plugin.code}</Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Department' : '归属部门'}>
                {detail.plugin.department || '-'}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Repository' : '仓库地址'} span={2}>
                <Typography.Link href={detail.plugin.repositoryUrl} target="_blank">
                  {detail.plugin.repositoryUrl}
                </Typography.Link>
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Description' : '插件描述'} span={2}>
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
              title={isEnglish ? 'Release List' : '发布单列表'}
              extra={
                <Button type="primary" onClick={openCreateRelease}>
                  {isEnglish ? 'New Release' : '新建发布单'}
                </Button>
              }
              style={{ borderRadius: 24, height: '100%' }}
              styles={{ body: { maxHeight: 720, overflowY: 'auto' } }}
            >
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                {(detail?.releases || []).length ? (
                  detail?.releases?.map((release) => {
                    const active = release.ID === selectedRelease?.ID;
                    return (
                      <Card
                        key={release.ID}
                        size="small"
                        hoverable
                        onClick={() => void loadDetail(release.ID)}
                        style={{
                          borderRadius: 18,
                          borderColor: active ? '#1677ff' : '#d9d9d9',
                          background: active ? '#f0f7ff' : '#fff',
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
                          <Button size="small" onClick={(event) => { event.stopPropagation(); void openEditRelease(release); }}>
                            {isEnglish ? 'Edit' : '编辑'}
                          </Button>
                        </Space>
                      </Card>
                    );
                  })
                ) : (
                  <Empty description={isEnglish ? 'No releases yet' : '暂无发布单'} />
                )}
              </Space>
            </Card>
          </Col>

          <Col xs={24} xl={15}>
            <Space direction="vertical" size={20} style={{ width: '100%' }}>
              <Card
                title={isEnglish ? 'Current Release' : '当前发布单'}
                style={{ borderRadius: 24 }}
                extra={
                  selectedRelease ? (
                    <Space wrap>
                      <Button onClick={() => void openEditRelease(selectedRelease)}>{isEnglish ? 'Edit' : '编辑发布单'}</Button>
                      {selectedRelease.status === 1 && (
                        <Button type="primary" onClick={() => void handleTransition('submit_review')}>
                          {isEnglish ? 'Submit Review' : '提交审核'}
                        </Button>
                      )}
                      {selectedRelease.status === 4 && (
                        <Button type="primary" onClick={() => void handleTransition('revise')}>
                          {isEnglish ? 'Resubmit' : '重新提审'}
                        </Button>
                      )}
                      {selectedRelease.status === 5 && (
                        <Button danger onClick={openOfflineModal}>
                          {isEnglish ? 'Request Offline' : '发起下架'}
                        </Button>
                      )}
                    </Space>
                  ) : undefined
                }
              >
                {selectedRelease ? (
                  <Descriptions column={2} styles={{ label: { width: 110 } }}>
                    <Descriptions.Item label={isEnglish ? 'Version' : '版本号'}>{selectedRelease.version || '-'}</Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Request Type' : '请求类型'}>
                      <RequestTypeTag type={selectedRelease.requestType} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Lifecycle' : '生命周期'}>
                      <ReleaseStatusTag status={selectedRelease.status} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Workflow' : '工单状态'}>
                      <ProcessStatusTag status={selectedRelease.processStatus} locale={locale} />
                    </Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Compatible' : '兼容产品'} span={2}>
                      {renderCompatibles(selectedRelease.compatibleItems)}
                    </Descriptions.Item>
                    <Descriptions.Item label="TD ID">{selectedRelease.tdId || '-'}</Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Created At' : '创建时间'}>
                      {formatTime(selectedRelease.createdAt)}
                    </Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Changelog' : '变更说明'} span={2}>
                      {getDisplayChangelog(locale, selectedRelease)}
                    </Descriptions.Item>
                    <Descriptions.Item label={isEnglish ? 'Offline Reason' : '下架原因'} span={2}>
                      {getDisplayOfflineReason(locale, selectedRelease)}
                    </Descriptions.Item>
                    <Descriptions.Item label="x86 URL" span={2}>
                      {selectedRelease.packageX86Url ? (
                        <Typography.Link href={selectedRelease.packageX86Url} target="_blank">
                          {selectedRelease.packageX86Url}
                        </Typography.Link>
                      ) : (
                        '-'
                      )}
                    </Descriptions.Item>
                    <Descriptions.Item label="ARM URL" span={2}>
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
                  <Empty description={isEnglish ? 'Select a release to inspect' : '请选择一个发布单查看'} />
                )}
              </Card>

              <Card title={isEnglish ? 'Timeline' : '流转时间轴'} style={{ borderRadius: 24 }}>
                <Timeline
                  items={(detail?.events || []).map((item) => ({
                    children: (
                      <Space direction="vertical" size={0}>
                        <Typography.Text strong>{item.action}</Typography.Text>
                        <Typography.Text type="secondary">{item.comment || '-'}</Typography.Text>
                        <Typography.Text type="secondary">{formatTime(item.createdAt)}</Typography.Text>
                      </Space>
                    ),
                  }))}
                />
              </Card>
            </Space>
          </Col>
        </Row>
      </Space>

      <Modal
        open={releaseModalOpen}
        title={editingRelease?.ID ? (isEnglish ? 'Edit Release' : '编辑发布单') : isEnglish ? 'New Release' : '新建发布单'}
        onCancel={() => setReleaseModalOpen(false)}
        onOk={() => void submitReleaseForm()}
        width={760}
        destroyOnHidden
      >
        <Form form={releaseForm} layout="vertical">
          <Form.Item name="version" label={isEnglish ? 'Version' : '版本号'} rules={[{ required: true }]}>
            <Input placeholder="1.0.0" />
          </Form.Item>
          <Form.Item name="testReportUrl" label={isEnglish ? 'Test Report URL' : '测试报告地址'}>
            <Input />
          </Form.Item>
          <Form.Item name="tdId" label="TD ID">
            <Input />
          </Form.Item>
          <Form.Item name="packageX86Url" label="x86 URL">
            <Input />
          </Form.Item>
          <Form.Item name="packageArmUrl" label="ARM URL">
            <Input />
          </Form.Item>
          <Form.Item name="changelogZh" label="中文变更说明">
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="changelogEn" label="English Changelog">
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.List name="compatibleItems">
            {(fields, { add, remove }) => (
              <Space direction="vertical" style={{ width: '100%' }}>
                {fields.map((field) => (
                  <Row gutter={12} key={field.key}>
                    <Col span={10}>
                      <Form.Item {...field} name={[field.name, 'productId']} label={isEnglish ? 'Product' : '产品'} rules={[{ required: true }]}>
                        <Select options={productOptions} />
                      </Form.Item>
                    </Col>
                    <Col span={10}>
                      <Form.Item {...field} name={[field.name, 'versionConstraint']} label={isEnglish ? 'Constraint' : '版本约束'}>
                        <Input />
                      </Form.Item>
                    </Col>
                    <Col span={4} style={{ display: 'flex', alignItems: 'center' }}>
                      <Button onClick={() => remove(field.name)}>{isEnglish ? 'Remove' : '删除'}</Button>
                    </Col>
                  </Row>
                ))}
                <Button onClick={() => add({ productId: undefined, versionConstraint: '' })}>
                  {isEnglish ? 'Add Compatible Product' : '添加兼容产品'}
                </Button>
              </Space>
            )}
          </Form.List>
        </Form>
      </Modal>

      <Modal
        open={offlineModalOpen}
        title={isEnglish ? 'Request Offline' : '发起下架申请'}
        onCancel={() => setOfflineModalOpen(false)}
        onOk={() => void submitOfflineRequest()}
        destroyOnHidden
      >
        <Form form={offlineForm} layout="vertical">
          <Form.Item name="tdId" label="TD ID">
            <Input />
          </Form.Item>
          <Form.Item name="offlineReasonZh" label="中文下架原因" rules={[{ required: true }]}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="offlineReasonEn" label="English Offline Reason">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default PluginProjectDetailPage;
