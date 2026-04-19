import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { getLocale, history, useParams } from '@umijs/max';
import { Button, Card, Descriptions, Empty, Space, Timeline, Typography, message } from 'antd';
import dayjs from 'dayjs';

import {
  claimWorkOrder,
  getProjectDetail,
  transitionRelease,
  type ProjectDetail,
} from '@/services/api/plugin';
import { getDisplayChangelog, getDisplayDescription, getDisplayName, isEnglishLocale, pickLocaleText } from '@/utils/plugin';
import { ProcessStatusTag, ReleaseStatusTag, RequestTypeTag } from '../components/status';

const formatTime = (value?: string) => (value ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : '-');

const PluginWorkOrderDetailPage: React.FC = () => {
  const locale = getLocale();
  const isEnglish = isEnglishLocale(locale);
  const params = useParams<{ pluginId: string; id: string }>();
  const pluginId = Number(params.pluginId);
  const releaseId = Number(params.id);
  const [detail, setDetail] = useState<ProjectDetail>();
  const [loading, setLoading] = useState(false);

  const selectedRelease = detail?.selectedRelease;
  const compatible = useMemo(() => {
    if (selectedRelease?.compatibility) {
      return {
        universal: Boolean(selectedRelease.compatibility.universal),
        products: selectedRelease.compatibility.products || [],
        acli: selectedRelease.compatibility.acli || [],
      };
    }
    if (selectedRelease?.compatibleInfo) {
      return {
        universal: Boolean(selectedRelease.compatibleInfo.universal),
        products: selectedRelease.compatibleInfo.products || [],
        acli: selectedRelease.compatibleInfo.acli || [],
      };
    }
    return {
      universal: false,
      products: selectedRelease?.compatibleItems || [],
      acli: [],
    };
  }, [selectedRelease]);

  const loadDetail = async () => {
    if (!pluginId || !releaseId) return;
    setLoading(true);
    try {
      const res = await getProjectDetail({ id: pluginId, releaseId });
      if (res.code !== 0) {
        message.error(res.msg || 'Failed to load work order detail');
        return;
      }
      setDetail(res.data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadDetail();
  }, [pluginId, releaseId]);

  const handleAction = async (action: 'approve' | 'reject' | 'release' | 'offline') => {
    if (!selectedRelease) return;
    const res = await transitionRelease({ id: selectedRelease.ID, action });
    if (res.code !== 0) {
      message.error(res.msg || action);
      return;
    }
    await loadDetail();
  };

  const handleClaim = async () => {
    if (!selectedRelease) return;
    const res = await claimWorkOrder({ id: selectedRelease.ID });
    if (res.code !== 0) {
      message.error(res.msg || 'claim');
      return;
    }
    await loadDetail();
  };

  return (
    <PageContainer
      title={false}
      onBack={() => history.push('/plugin/work-order-pool')}
      extra={[
        <Button key="back" onClick={() => history.push('/plugin/work-order-pool')}>
          {isEnglish ? 'Back to Work Orders' : '返回工单池'}
        </Button>,
      ]}
    >
      <Space direction="vertical" size={20} style={{ width: '100%' }}>
        <Card
          loading={loading}
          style={{ borderRadius: 24 }}
          title={isEnglish ? 'Work Order Detail' : '工单详情'}
          extra={
            selectedRelease ? (
              <Space wrap>
                {selectedRelease.processStatus === 0 && selectedRelease.status === 2 ? (
                  <Button onClick={() => void handleClaim()}>{isEnglish ? 'Claim' : '认领'}</Button>
                ) : null}
                {selectedRelease.processStatus === 1 && selectedRelease.status === 2 ? (
                  <>
                    <Button type="primary" onClick={() => void handleAction('approve')}>
                      {isEnglish ? 'Approve' : '审核通过'}
                    </Button>
                    <Button danger onClick={() => void handleAction('reject')}>
                      {isEnglish ? 'Reject' : '打回'}
                    </Button>
                  </>
                ) : null}
                {selectedRelease.processStatus === 1 && selectedRelease.status === 3 && selectedRelease.requestType === 1 ? (
                  <Button type="primary" onClick={() => void handleAction('release')}>
                    {isEnglish ? 'Release' : '执行发布'}
                  </Button>
                ) : null}
                {selectedRelease.processStatus === 1 && selectedRelease.status === 3 && selectedRelease.requestType === 2 ? (
                  <Button danger onClick={() => void handleAction('offline')}>
                    {isEnglish ? 'Offline' : '执行下架'}
                  </Button>
                ) : null}
              </Space>
            ) : undefined
          }
        >
          {selectedRelease && detail?.plugin ? (
            <Descriptions column={2}>
              <Descriptions.Item label={isEnglish ? 'Plugin' : '插件'}>
                {getDisplayName(locale, detail.plugin)}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Code' : '插件编码'}>{detail.plugin.code}</Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Version' : '版本号'}>{selectedRelease.version}</Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'TD ID' : 'TD ID'}>{selectedRelease.tdId || '-'}</Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Department' : '归属部门'}>
                {pickLocaleText(
                  locale,
                  detail.plugin.departmentNameZh || detail.plugin.department,
                  detail.plugin.departmentNameEn || detail.plugin.department,
                )}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Claimer' : '认领人'}>
                {selectedRelease.claimerName || selectedRelease.claimerUsername || '-'}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Request Type' : '请求类型'}>
                <RequestTypeTag type={selectedRelease.requestType} locale={locale} />
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Lifecycle' : '生命周期'}>
                <ReleaseStatusTag status={selectedRelease.status} locale={locale} />
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Workflow' : '工单状态'}>
                <ProcessStatusTag status={selectedRelease.processStatus} locale={locale} />
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Created At' : '创建时间'}>{formatTime(selectedRelease.createdAt)}</Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Description' : '插件描述'} span={2}>
                {getDisplayDescription(locale, detail.plugin)}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Changelog' : '变更说明'} span={2}>
                {getDisplayChangelog(locale, selectedRelease)}
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Compatibility' : '兼容信息'} span={2}>
                <Space direction="vertical" size={12}>
                  <Space direction="vertical" size={4}>
                    <Typography.Text strong>{isEnglish ? 'Product Compatibility' : '产品兼容'}</Typography.Text>
                    {compatible.products.length ? (
                      compatible.products.map((item) => (
                        <Typography.Text key={`product-${item.productId}`}>
                          {item.productName || item.productCode}
                          {item.versionConstraint ? ` ${item.versionConstraint}` : ''}
                        </Typography.Text>
                      ))
                    ) : (
                      <Typography.Text type="secondary">-</Typography.Text>
                    )}
                  </Space>
                  <Space direction="vertical" size={4}>
                    <Typography.Text strong>aCLI 兼容</Typography.Text>
                    {compatible.acli.length ? (
                      compatible.acli.map((item) => (
                        <Typography.Text key={`acli-${item.productId}`}>
                          {item.productName || item.productCode}
                          {item.versionConstraint ? ` ${item.versionConstraint}` : ''}
                        </Typography.Text>
                      ))
                    ) : (
                      <Typography.Text type="secondary">-</Typography.Text>
                    )}
                  </Space>
                  <Space>
                    <Typography.Text strong>{isEnglish ? 'Universal Support' : '全支持'}</Typography.Text>
                    <Typography.Text>{compatible.universal ? (isEnglish ? 'Yes' : '是') : (isEnglish ? 'No' : '否')}</Typography.Text>
                  </Space>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label={isEnglish ? 'Test Report' : '测试报告'} span={2}>
                {selectedRelease.testReportUrl ? (
                  <Typography.Link href={selectedRelease.testReportUrl} target="_blank">
                    {selectedRelease.testReportUrl}
                  </Typography.Link>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="x86" span={2}>
                {selectedRelease.packageX86Url ? (
                  <Typography.Link href={selectedRelease.packageX86Url} target="_blank">
                    {selectedRelease.packageX86Url}
                  </Typography.Link>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="ARM" span={2}>
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
            <Empty />
          )}
        </Card>

        <Card title={isEnglish ? 'Timeline' : '流转时间线'} style={{ borderRadius: 24 }}>
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
    </PageContainer>
  );
};

export default PluginWorkOrderDetailPage;
