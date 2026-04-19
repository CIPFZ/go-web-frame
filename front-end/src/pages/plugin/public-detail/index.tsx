import React, { useEffect, useMemo, useState } from 'react';
import { history, useParams } from '@umijs/max';
import {
  AppstoreOutlined,
  ArrowLeftOutlined,
  DownloadOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import {
  Breadcrumb,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Radio,
  Row,
  Select,
  Space,
  Spin,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';

import { getPublishedPluginDetail, type PublishedPluginDetail } from '@/services/api/plugin';

type MarketVersion = PublishedPluginDetail['versions'][number];

const copyMap = {
  zh: {
    market: '插件生态中心',
    loading: '加载中...',
    detail: '插件详情',
    loadFailed: '加载插件详情失败',
    owner: '负责人',
    currentVersion: '当前版本',
    releaseDate: '发布日期',
    publisher: '发布人',
    overview: '插件说明',
    basicInfo: '基本说明',
    capabilities: '插件能力',
    changelogTab: '版本说明',
    changelog: '更新日志',
    reportTab: '测试与性能',
    report: '测试报告',
    reportDownload: '下载完整测试报告',
    reportEmpty: '当前版本暂无测试报告',
    performance: '性能结论',
    historyTab: '历史版本',
    version: '版本号',
    status: '状态',
    viewing: '当前查看',
    published: '已发布',
    getInstall: '获取与安装',
    selectVersion: '选择要下载的版本',
    downloadX86: '下载标准版 (x86_64)',
    downloadArm: '下载适配版 (ARM64)',
    metadata: '版本相关资料',
    versionConstraint: '版本约束',
    x86Package: 'x86 包名',
    armPackage: 'ARM 包名',
    testAttachment: '测试附件',
    publicDownload: '公开下载',
    bilingual: '支持双语',
    noDetail: '未找到该插件详情，可能已被移除',
  },
  en: {
    market: 'Plugin Ecosystem',
    loading: 'Loading...',
    detail: 'Plugin Details',
    loadFailed: 'Failed to load plugin details',
    owner: 'Owner',
    currentVersion: 'Current Version',
    releaseDate: 'Release Date',
    publisher: 'Publisher',
    overview: 'Overview',
    basicInfo: 'Basic Info',
    capabilities: 'Capabilities',
    changelogTab: 'Changelog',
    changelog: 'Changelog',
    reportTab: 'Testing & Performance',
    report: 'Test Report',
    reportDownload: 'Download Complete Test Report',
    reportEmpty: 'No test report available for this version',
    performance: 'Performance Conclusion',
    historyTab: 'History',
    version: 'Version',
    status: 'Status',
    viewing: 'Viewing',
    published: 'Published',
    getInstall: 'Get & Install',
    selectVersion: 'Select version to download',
    downloadX86: 'Download Standard (x86_64)',
    downloadArm: 'Download Adapted (ARM64)',
    metadata: 'Version Metadata',
    versionConstraint: 'Version Constraint',
    x86Package: 'x86 Package',
    armPackage: 'ARM Package',
    testAttachment: 'Test Attachment',
    publicDownload: 'Public Download',
    bilingual: 'Bilingual Support',
    noDetail: 'Plugin not found or no longer published',
  },
};

const formatTime = (value?: string) => (value ? value.replace('T', ' ').replace('Z', '').slice(0, 16) : '-');

const extractFileName = (url?: string) => {
  if (!url) return '-';
  const cleanUrl = url.split('?')[0];
  const segments = cleanUrl.split('/');
  return decodeURIComponent(segments[segments.length - 1] || 'uploaded-file');
};

const cardStyle: React.CSSProperties = {
  borderRadius: 4,
  border: '1px solid #e8e8e8',
  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)',
};

const getName = (lang: 'zh' | 'en', detail?: PublishedPluginDetail) =>
  lang === 'zh' ? detail?.plugin.nameZh : detail?.plugin.nameEn || detail?.plugin.nameZh || '';

const getDescription = (lang: 'zh' | 'en', detail?: PublishedPluginDetail) =>
  lang === 'zh'
    ? detail?.plugin.descriptionZh || detail?.plugin.descriptionEn || '-'
    : detail?.plugin.descriptionEn || detail?.plugin.descriptionZh || '-';

const getCapability = (lang: 'zh' | 'en', detail?: PublishedPluginDetail) =>
  lang === 'zh'
    ? detail?.plugin.capabilityZh || detail?.plugin.capabilityEn || '-'
    : detail?.plugin.capabilityEn || detail?.plugin.capabilityZh || '-';

const getChangelog = (lang: 'zh' | 'en', version?: MarketVersion) =>
  lang === 'zh' ? version?.changelogZh || version?.changelogEn || '-' : version?.changelogEn || version?.changelogZh || '-';

const getPerformance = (lang: 'zh' | 'en', version?: MarketVersion) =>
  lang === 'zh'
    ? version?.performanceSummaryZh || version?.performanceSummaryEn || '-'
    : version?.performanceSummaryEn || version?.performanceSummaryZh || '-';

const PluginPublicDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const pluginId = Number(params.id);
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<PublishedPluginDetail>();
  const [selectedVersionId, setSelectedVersionId] = useState<number>();
  const [lang, setLang] = useState<'zh' | 'en'>(() => {
    return (localStorage.getItem('plugin-market-lang') as 'zh' | 'en') || 'zh';
  });

  const copy = copyMap[lang];

  const updateLang = (value: 'zh' | 'en') => {
    setLang(value);
    localStorage.setItem('plugin-market-lang', value);
  };

  useEffect(() => {
    if (!pluginId) return;

    const loadDetail = async () => {
      setLoading(true);
      try {
        const res = await getPublishedPluginDetail({ pluginId }, { skipErrorHandler: true });
        if (res.code !== 0) {
          message.error(res.msg || copy.loadFailed);
          return;
        }
        setDetail(res.data);
        setSelectedVersionId(res.data?.versions?.[0]?.releaseId);
      } finally {
        setLoading(false);
      }
    };

    void loadDetail();
  }, [copy.loadFailed, pluginId]);

  const selectedVersion = useMemo(() => {
    const versions = detail?.versions || [];
    if (!versions.length) return undefined;
    return versions.find((item) => item.releaseId === selectedVersionId) || versions[0];
  }, [detail?.versions, selectedVersionId]);

  const versionColumns: ColumnsType<MarketVersion> = [
    {
      title: copy.version,
      dataIndex: 'version',
      width: 110,
      render: (value: string) => (
        <Tag color="blue" bordered={false} style={{ borderRadius: 2 }}>
          v{value}
        </Tag>
      ),
    },
    {
      title: copy.releaseDate,
      dataIndex: 'releasedAt',
      width: 160,
      render: (value?: string) => formatTime(value),
    },
    {
      title: copy.publisher,
      dataIndex: 'publisher',
      width: 120,
      render: (value?: string) => value || '-',
    },
    {
      title: copy.status,
      width: 100,
      render: (_, record) =>
        record.releaseId === selectedVersion?.releaseId ? (
          <Tag color="processing" bordered={false} style={{ borderRadius: 2 }}>
            {copy.viewing}
          </Tag>
        ) : (
          <Tag bordered={false} style={{ borderRadius: 2 }}>
            {copy.published}
          </Tag>
        ),
    },
  ];

  const tabItems = selectedVersion
    ? [
        {
          key: 'overview',
          label: copy.overview,
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 12 }}>
                {copy.basicInfo}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {getDescription(lang, detail)}
              </Typography.Paragraph>

              <Typography.Title level={5} style={{ marginTop: 24, marginBottom: 12 }}>
                {copy.capabilities}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {getCapability(lang, detail)}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'changelog',
          label: copy.changelogTab,
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                {copy.changelog}
              </Typography.Text>
              <Typography.Paragraph style={{ lineHeight: 1.8, background: '#fafafa', padding: 16, borderRadius: 4 }}>
                {getChangelog(lang, selectedVersion)}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'report',
          label: copy.reportTab,
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 12 }}>
                {copy.report}
              </Typography.Title>
              <div style={{ marginBottom: 24, padding: 16, background: '#fafafa', borderRadius: 4 }}>
                {selectedVersion.testReportUrl ? (
                  <Button
                    type="primary"
                    href={selectedVersion.testReportUrl}
                    target="_blank"
                    icon={<FileTextOutlined style={{ marginRight: 4 }} />}
                    style={{ borderRadius: 2 }}
                  >
                    {copy.reportDownload}
                  </Button>
                ) : (
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={copy.reportEmpty} />
                )}
              </div>

              <Typography.Title level={5} style={{ marginTop: 16, marginBottom: 12 }}>
                {copy.performance}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {getPerformance(lang, selectedVersion)}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'history',
          label: copy.historyTab,
          children: (
            <div style={{ padding: '8px 0' }}>
              <Table
                size="middle"
                pagination={{ pageSize: 15 }}
                rowKey="releaseId"
                columns={versionColumns}
                dataSource={detail?.versions || []}
                rowClassName={(record) => (record.releaseId === selectedVersion?.releaseId ? 'ant-table-row-selected' : '')}
                onRow={(record) => ({
                  onClick: () => setSelectedVersionId(record.releaseId),
                  style: { cursor: 'pointer' },
                })}
              />
            </div>
          ),
        },
      ]
    : [];

  return (
    <div style={{ minHeight: '100vh', background: '#f5f7fa', paddingBottom: 64 }}>
      <div style={{ background: '#fff', borderBottom: '1px solid #e8e8e8', padding: '16px 24px' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Breadcrumb
            items={[
              {
                title: (
                  <Typography.Text type="secondary" style={{ cursor: 'pointer' }} onClick={() => history.push('/plugins')}>
                    <ArrowLeftOutlined style={{ marginRight: 4 }} />
                    {copy.market}
                  </Typography.Text>
                ),
              },
              {
                title: (
                  <Typography.Text strong>
                    {loading ? copy.loading : getName(lang, detail) || copy.detail}
                  </Typography.Text>
                ),
              },
            ]}
          />
          <Radio.Group size="small" value={lang} onChange={(event) => updateLang(event.target.value)} buttonStyle="solid">
            <Radio.Button value="zh">中</Radio.Button>
            <Radio.Button value="en">EN</Radio.Button>
          </Radio.Group>
        </div>
      </div>

      <div style={{ maxWidth: 1200, margin: '24px auto 0', padding: '0 24px' }}>
        {loading && !detail ? (
          <div style={{ textAlign: 'center', padding: 80 }}>
            <Spin size="large" />
          </div>
        ) : !detail ? (
          <Card variant="borderless" style={cardStyle}>
            <Empty description={copy.noDetail} />
          </Card>
        ) : (
          <>
            <Card
              variant="borderless"
              style={{ ...cardStyle, marginBottom: 24, position: 'relative', overflow: 'hidden' }}
              styles={{ body: { padding: 32 } }}
            >
              <Row gutter={[32, 24]} align="middle">
                <Col xs={24} lg={16}>
                  <Space wrap size={[8, 8]} style={{ marginBottom: 16 }}>
                    <Tag color="blue" style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}>
                      <AppstoreOutlined style={{ marginRight: 4 }} />
                      {detail.plugin.code}
                    </Tag>
                    <Tag style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}>{copy.publicDownload}</Tag>
                    <Tag style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}>{copy.bilingual}</Tag>
                  </Space>

                  <Typography.Title level={3} style={{ margin: 0 }}>
                    {getName(lang, detail)}
                  </Typography.Title>

                  <Typography.Paragraph type="secondary" style={{ fontSize: 14, margin: '16px 0 0', lineHeight: 1.6, maxWidth: 800 }}>
                    {getDescription(lang, detail)}
                  </Typography.Paragraph>
                </Col>

                <Col xs={24} lg={8}>
                  <div style={{ background: '#fafafa', padding: 20, borderRadius: 4, height: '100%' }}>
                    <Descriptions
                      column={1}
                      size="small"
                      styles={{
                        label: { width: lang === 'zh' ? 90 : 120, color: '#888' },
                        content: { fontWeight: 500 },
                      }}
                    >
                      <Descriptions.Item label={copy.owner}>{detail.plugin.ownerName || '-'}</Descriptions.Item>
                      <Descriptions.Item label={copy.currentVersion}>
                        <Tag color="blue" bordered={false} style={{ borderRadius: 2, margin: 0 }}>
                          v{selectedVersion?.version || '-'}
                        </Tag>
                      </Descriptions.Item>
                      <Descriptions.Item label={copy.releaseDate}>{formatTime(selectedVersion?.releasedAt)}</Descriptions.Item>
                      <Descriptions.Item label={copy.publisher}>
                        {selectedVersion?.publisher || detail.plugin.ownerName || '-'}
                      </Descriptions.Item>
                    </Descriptions>
                  </div>
                </Col>
              </Row>
            </Card>

            <Row gutter={[24, 24]}>
              <Col xs={24} xl={16}>
                <Card variant="borderless" style={{ ...cardStyle, height: '100%' }} styles={{ body: { padding: '0 24px 24px' } }}>
                  <Tabs defaultActiveKey="overview" items={tabItems} size="middle" />
                </Card>
              </Col>

              <Col xs={24} xl={8}>
                <Space direction="vertical" size={24} style={{ width: '100%' }}>
                  <Card variant="borderless" style={cardStyle} styles={{ body: { padding: 24 } }}>
                    <Typography.Title level={5} style={{ marginTop: 0, display: 'flex', alignItems: 'center' }}>
                      <InfoCircleOutlined style={{ marginRight: 8, color: '#1677ff' }} />
                      {copy.getInstall}
                    </Typography.Title>

                    <div style={{ marginBottom: 16, marginTop: 16 }}>
                      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                        {copy.selectVersion}
                      </Typography.Text>
                      <Select
                        value={selectedVersion?.releaseId}
                        style={{ width: '100%' }}
                        options={(detail.versions || []).map((item) => ({
                          label: `v${item.version} (${formatTime(item.releasedAt)})`,
                          value: item.releaseId,
                        }))}
                        onChange={(value) => setSelectedVersionId(value)}
                      />
                    </div>

                    <Space direction="vertical" size={12} style={{ width: '100%' }}>
                      <Button
                        type="primary"
                        block
                        icon={<DownloadOutlined />}
                        href={selectedVersion?.packageX86Url}
                        target="_blank"
                        disabled={!selectedVersion?.packageX86Url}
                        style={{ borderRadius: 2, height: 38 }}
                      >
                        {copy.downloadX86}
                      </Button>
                      <Button
                        block
                        icon={<DownloadOutlined />}
                        href={selectedVersion?.packageArmUrl}
                        target="_blank"
                        disabled={!selectedVersion?.packageArmUrl}
                        style={{ borderRadius: 2, height: 38 }}
                      >
                        {copy.downloadArm}
                      </Button>
                    </Space>
                  </Card>

                  <Card variant="borderless" style={cardStyle} styles={{ body: { padding: 24 } }}>
                    <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 16 }}>
                      {copy.metadata}
                    </Typography.Title>
                    <Descriptions column={1} size="small" styles={{ label: { width: lang === 'zh' ? 90 : 130, color: '#888' } }}>
                      <Descriptions.Item label={copy.versionConstraint}>{selectedVersion?.versionConstraint || '-'}</Descriptions.Item>
                      <Descriptions.Item label={copy.x86Package}>
                        <Typography.Text ellipsis style={{ width: 140 }} title={extractFileName(selectedVersion?.packageX86Url)}>
                          {extractFileName(selectedVersion?.packageX86Url)}
                        </Typography.Text>
                      </Descriptions.Item>
                      <Descriptions.Item label={copy.armPackage}>
                        <Typography.Text ellipsis style={{ width: 140 }} title={extractFileName(selectedVersion?.packageArmUrl)}>
                          {extractFileName(selectedVersion?.packageArmUrl)}
                        </Typography.Text>
                      </Descriptions.Item>
                      <Descriptions.Item label={copy.testAttachment}>
                        <Typography.Text ellipsis style={{ width: 140 }} title={extractFileName(selectedVersion?.testReportUrl)}>
                          {extractFileName(selectedVersion?.testReportUrl)}
                        </Typography.Text>
                      </Descriptions.Item>
                    </Descriptions>
                  </Card>
                </Space>
              </Col>
            </Row>
          </>
        )}
      </div>
    </div>
  );
};

export default PluginPublicDetailPage;
