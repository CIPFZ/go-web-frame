import React, { useEffect, useMemo, useState } from 'react';
import { history, useParams } from '@umijs/max';
import {
  Breadcrumb,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
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
import {
  ArrowLeftOutlined,
  DownloadOutlined,
  FileTextOutlined,
  AppstoreOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { getPublishedPluginDetail } from '@/services/api/plugin';

type PublishedPluginVersion = {
  releaseId: number;
  version: string;
  versionConstraint?: string;
  publisher?: string;
  packageX86Url?: string;
  packageArmUrl?: string;
  testReportUrl?: string;
  changelogZh?: string;
  changelogEn?: string;
  performanceSummaryZh?: string;
  performanceSummaryEn?: string;
  releasedAt: string;
};

type PublishedPluginDetail = {
  pluginId: number;
  code: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  capabilityZh: string;
  capabilityEn: string;
  owner: string;
  versions: PublishedPluginVersion[];
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

const PluginMarketDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const pluginId = Number(params.id);

  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<PublishedPluginDetail>();
  const [selectedReleaseId, setSelectedReleaseId] = useState<number>();
  const [lang, setLangState] = useState<'zh' | 'en'>(() => {
    return (localStorage.getItem('plugin-market-lang') as 'zh' | 'en') || 'zh';
  });

  const setLang = (l: 'zh' | 'en') => {
    setLangState(l);
    localStorage.setItem('plugin-market-lang', l);
  };

  useEffect(() => {
    if (!pluginId) return;
    void loadDetail();
  }, [pluginId]);

  const loadDetail = async () => {
    setLoading(true);
    const res = await getPublishedPluginDetail({ pluginId }, { skipErrorHandler: true }).catch((error) => error);
    setLoading(false);
    if (!res || res.code !== 0) {
      message.error(res?.msg || (lang === 'zh' ? '加载插件详情失败' : 'Failed to load plugin details'));
      return;
    }
    const nextDetail = res.data as PublishedPluginDetail;
    setDetail(nextDetail);
    setSelectedReleaseId(nextDetail.versions?.[0]?.releaseId);
  };

  const selectedVersion = useMemo(() => {
    const versions = detail?.versions || [];
    if (!versions.length) return undefined;
    if (!selectedReleaseId) return versions[0];
    return versions.find((item) => item.releaseId === selectedReleaseId) || versions[0];
  }, [detail, selectedReleaseId]);

  const versionColumns: ColumnsType<PublishedPluginVersion> = [
    {
      title: lang === 'zh' ? '版本号' : 'Version',
      dataIndex: 'version',
      width: 100,
      render: (value: string) => <Tag color="blue" bordered={false} style={{ borderRadius: 2 }}>v{value}</Tag>,
    },
    {
      title: lang === 'zh' ? '发布时间' : 'Release Date',
      dataIndex: 'releasedAt',
      width: 140,
      render: (value: string) => formatTime(value),
    },
    {
      title: lang === 'zh' ? '发布人' : 'Publisher',
      dataIndex: 'publisher',
      width: 100,
      render: (value?: string) => value || '-',
    },
    {
      title: lang === 'zh' ? '状态' : 'Status',
      width: 80,
      render: (_, record) =>
        record.releaseId === selectedVersion?.releaseId ? (
          <Tag color="processing" bordered={false} style={{ borderRadius: 2 }}>{lang === 'zh' ? '当前查看' : 'Viewing'}</Tag>
        ) : (
          <Tag bordered={false} style={{ borderRadius: 2 }}>{lang === 'zh' ? '已发布' : 'Published'}</Tag>
        ),
    },
  ];

  const tabItems = selectedVersion
    ? [
        {
          key: 'overview',
          label: lang === 'zh' ? '插件说明' : 'Overview',
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 12 }}>
                {lang === 'zh' ? '基本说明' : 'Basic Info'}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {lang === 'zh' ? (detail?.descriptionZh || '-') : (detail?.descriptionEn || detail?.descriptionZh || '-')}
              </Typography.Paragraph>
              
              <Divider dashed />
              
              <Typography.Title level={5} style={{ marginTop: 16, marginBottom: 12 }}>
                {lang === 'zh' ? '插件能力' : 'Capabilities'}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {lang === 'zh' ? (detail?.capabilityZh || '-') : (detail?.capabilityEn || detail?.capabilityZh || '-')}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'changelog',
          label: lang === 'zh' ? '版本说明' : 'Changelog',
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                {lang === 'zh' ? '更新日志' : 'Changelog'}
              </Typography.Text>
              <Typography.Paragraph style={{ lineHeight: 1.8, background: '#fafafa', padding: 16, borderRadius: 4 }}>
                {lang === 'zh' ? (selectedVersion.changelogZh || '-') : (selectedVersion.changelogEn || selectedVersion.changelogZh || '-')}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'report',
          label: lang === 'zh' ? '测试与性能' : 'Testing & Performance',
          children: (
            <div style={{ padding: '8px 0' }}>
              <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 12 }}>
                {lang === 'zh' ? '测试报告' : 'Test Report'}
              </Typography.Title>
              <div style={{ marginBottom: 24, padding: 16, background: '#fafafa', borderRadius: 4 }}>
                {selectedVersion.testReportUrl ? (
                  <Button type="primary" href={selectedVersion.testReportUrl} target="_blank" icon={<FileTextOutlined style={{ marginRight: 4 }}/>} style={{ borderRadius: 2 }}>
                    {lang === 'zh' ? '下载完整测试报告' : 'Download Complete Test Report'}
                  </Button>
                ) : (
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={lang === 'zh' ? "当前版本暂无测试报告" : "No test report available for current version"} />
                )}
              </div>
              
              <Divider dashed />

              <Typography.Title level={5} style={{ marginTop: 16, marginBottom: 12 }}>
                {lang === 'zh' ? '性能结论' : 'Performance Conclusion'}
              </Typography.Title>
              <Typography.Paragraph style={{ lineHeight: 1.8 }}>
                {lang === 'zh' ? (selectedVersion.performanceSummaryZh || '-') : (selectedVersion.performanceSummaryEn || selectedVersion.performanceSummaryZh || '-')}
              </Typography.Paragraph>
            </div>
          ),
        },
        {
          key: 'history',
          label: lang === 'zh' ? '历史版本' : 'History',
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
                  onClick: () => setSelectedReleaseId(record.releaseId),
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
      {/* Banner / Header Title Area */}
      <div style={{ background: '#fff', borderBottom: '1px solid #e8e8e8', padding: '16px 24px' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Breadcrumb
            items={[
              {
                title: (
                  <Typography.Text
                    type="secondary"
                    style={{ cursor: 'pointer' }}
                    onClick={() => history.push('/plugins')}
                  >
                    <ArrowLeftOutlined style={{ marginRight: 4 }} />
                    {lang === 'zh' ? '插件发布中心' : 'Plugin Market'}
                  </Typography.Text>
                ),
              },
              { title: <Typography.Text strong>{loading ? (lang === 'zh' ? '加载中...' : 'Loading...') : (lang === 'zh' ? detail?.nameZh : (detail?.nameEn || detail?.nameZh)) || (lang === 'zh' ? '插件详情' : 'Plugin Details')}</Typography.Text> },
            ]}
          />
          <Radio.Group size="small" value={lang} onChange={(e) => setLang(e.target.value)} buttonStyle="solid">
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
          <Card bordered={false} style={cardStyle}>
            <Empty description={lang === 'zh' ? "未找到该插件详情，可能已被下架或删除" : "Plugin not found, it might have been removed"} />
          </Card>
        ) : (
          <>
            {/* Plugin Header Summary Card */}
            <Card bordered={false} style={{ ...cardStyle, marginBottom: 24, position: 'relative', overflow: 'hidden' }} bodyStyle={{ padding: 32 }}>
              <Row gutter={[32, 24]} align="middle">
                <Col xs={24} lg={16}>
                  <Space wrap size={[8, 8]} style={{ marginBottom: 16 }}>
                    <Tag color="blue" style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}><AppstoreOutlined style={{ marginRight: 4 }}/>{detail.code}</Tag>
                    <Tag style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}>{lang === 'zh' ? '公开下载' : 'Public Download'}</Tag>
                    <Tag style={{ borderRadius: 2, padding: '2px 8px', fontSize: 13 }}>{lang === 'zh' ? '支持双语' : 'Bilingual Support'}</Tag>
                  </Space>
                  
                  <div style={{ marginBottom: 8 }}>
                    <Typography.Title level={3} style={{ margin: 0, display: 'inline', marginRight: 12 }}>
                      {lang === 'zh' ? detail.nameZh : (detail.nameEn || detail.nameZh)}
                    </Typography.Title>
                  </div>
                  
                  <Typography.Paragraph
                    type="secondary"
                    style={{ fontSize: 14, margin: '16px 0 0', lineHeight: 1.6, maxWidth: 800 }}
                  >
                    {lang === 'zh' ? detail.descriptionZh : (detail.descriptionEn || detail.descriptionZh)}
                  </Typography.Paragraph>
                </Col>

                <Col xs={24} lg={8}>
                  <div style={{ background: '#fafafa', padding: 20, borderRadius: 4, height: '100%' }}>
                    <Descriptions column={1} size="small" labelStyle={{ width: lang === 'zh' ? 90 : 120, color: '#888' }} contentStyle={{ fontWeight: 500 }}>
                      <Descriptions.Item label={lang === 'zh' ? '负责人' : 'Owner'}>{detail.owner || '-'}</Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? '当前版本' : 'Current Version'}>
                        <Tag color="blue" bordered={false} style={{ borderRadius: 2, margin: 0 }}>v{selectedVersion?.version || '-'}</Tag>
                      </Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? '发布时间' : 'Release Date'}>{formatTime(selectedVersion?.releasedAt)}</Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? '发布者' : 'Publisher'}>{selectedVersion?.publisher || detail.owner || '-'}</Descriptions.Item>
                    </Descriptions>
                  </div>
                </Col>
              </Row>
            </Card>

            <Row gutter={[24, 24]}>
              {/* Left Main Content */}
              <Col xs={24} xl={16}>
                <Card bordered={false} style={{ ...cardStyle, height: '100%' }} bodyStyle={{ padding: '0 24px 24px' }}>
                  <Tabs defaultActiveKey="overview" items={tabItems} size="middle" />
                </Card>
              </Col>

              {/* Right Sidebar */}
              <Col xs={24} xl={8}>
                <Space direction="vertical" size={24} style={{ width: '100%' }}>
                  
                  {/* Download Card */}
                  <Card bordered={false} style={cardStyle} bodyStyle={{ padding: 24 }}>
                    <Typography.Title level={5} style={{ marginTop: 0, display: 'flex', alignItems: 'center' }}>
                      <InfoCircleOutlined style={{ marginRight: 8, color: '#1677ff' }} />
                      {lang === 'zh' ? '获取与安装' : 'Get & Install'}
                    </Typography.Title>
                    
                    <div style={{ marginBottom: 16, marginTop: 16 }}>
                      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>{lang === 'zh' ? '选择要下载的版本：' : 'Select version to download:'}</Typography.Text>
                      <Select
                        value={selectedVersion?.releaseId}
                        style={{ width: '100%' }}
                        options={(detail.versions || []).map((item) => ({
                          label: `v${item.version} (${formatTime(item.releasedAt)})`,
                          value: item.releaseId,
                        }))}
                        onChange={(value) => setSelectedReleaseId(value)}
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
                        {lang === 'zh' ? '下载标准版 (x86_64)' : 'Download Standard (x86_64)'}
                      </Button>
                      <Button
                        block
                        icon={<DownloadOutlined />}
                        href={selectedVersion?.packageArmUrl}
                        target="_blank"
                        disabled={!selectedVersion?.packageArmUrl}
                        style={{ borderRadius: 2, height: 38 }}
                      >
                        {lang === 'zh' ? '下载适配版 (ARM64)' : 'Download Adapted (ARM64)'}
                      </Button>
                    </Space>
                  </Card>

                  {/* Version Meta Info */}
                  <Card bordered={false} style={cardStyle} bodyStyle={{ padding: 24 }}>
                    <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 16 }}>
                      {lang === 'zh' ? '版本相关资料' : 'Version Metadata'}
                    </Typography.Title>
                    <Descriptions column={1} size="small" labelStyle={{ width: lang === 'zh' ? 80 : 130, color: '#888' }}>
                      <Descriptions.Item label={lang === 'zh' ? '版本约束' : 'Version Constraint'}>{selectedVersion?.versionConstraint || '-'}</Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? 'x86 包名' : 'x86 Package'}>
                        <Typography.Text ellipsis style={{ width: 140 }} title={extractFileName(selectedVersion?.packageX86Url)}>
                          {extractFileName(selectedVersion?.packageX86Url)}
                        </Typography.Text>
                      </Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? 'ARM 包名' : 'ARM Package'}>
                        <Typography.Text ellipsis style={{ width: 140 }} title={extractFileName(selectedVersion?.packageArmUrl)}>
                          {extractFileName(selectedVersion?.packageArmUrl)}
                        </Typography.Text>
                      </Descriptions.Item>
                      <Descriptions.Item label={lang === 'zh' ? '测试附件' : 'Test Attachment'}>
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

export default PluginMarketDetailPage;