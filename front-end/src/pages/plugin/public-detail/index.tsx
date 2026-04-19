import React, { useEffect, useMemo, useState } from 'react';
import { getLocale, history, useParams } from '@umijs/max';
import { ArrowLeftOutlined, CloudDownloadOutlined, LinkOutlined } from '@ant-design/icons';
import { Button, Card, Col, Descriptions, Empty, Row, Select, Space, Tag, Typography, message } from 'antd';

import { getPublishedPluginDetail, type PublishedPluginDetail } from '@/services/api/plugin';
import { getDisplayChangelog, getDisplayDescription, getDisplayName, isEnglishLocale } from '@/utils/plugin';

const copyMap = {
  zh: {
    back: '返回插件中心',
    version: '版本',
    current: '当前发布版本',
    download: '下载包',
    report: '测试报告',
    repo: '仓库地址',
    intro: '插件说明',
    compatible: '兼容产品',
    history: '版本历史',
    empty: '未找到该插件的公开详情',
  },
  en: {
    back: 'Back to Marketplace',
    version: 'Version',
    current: 'Current Release',
    download: 'Download',
    report: 'Test Report',
    repo: 'Repository',
    intro: 'Overview',
    compatible: 'Compatible Products',
    history: 'Version History',
    empty: 'The plugin detail is not available.',
  },
};

const PluginPublicDetailPage: React.FC = () => {
  const params = useParams<{ id: string }>();
  const pluginId = Number(params.id);
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? copyMap.en : copyMap.zh;
  const [detail, setDetail] = useState<PublishedPluginDetail>();
  const [selectedVersionId, setSelectedVersionId] = useState<number>();

  useEffect(() => {
    const load = async () => {
      const res = await getPublishedPluginDetail({ id: pluginId }, { skipErrorHandler: true });
      if (res.code !== 0) {
        message.error(res.msg || copy.empty);
        return;
      }
      setDetail(res.data);
      setSelectedVersionId(res.data?.release?.ID);
    };
    if (pluginId) {
      void load();
    }
  }, [copy.empty, pluginId]);

  const selectedVersion = useMemo(() => {
    const versions = detail?.versions || [];
    if (!versions.length) return undefined;
    return versions.find((item) => item.ID === selectedVersionId) || versions[0];
  }, [detail?.versions, selectedVersionId]);

  if (!detail) {
    return (
      <div style={{ padding: 40 }}>
        <Card style={{ borderRadius: 24 }}>
          <Empty description={copy.empty} />
        </Card>
      </div>
    );
  }

  return (
    <div style={{ minHeight: '100vh', background: '#f4f8fc', padding: '28px 24px 40px' }}>
      <div style={{ maxWidth: 1240, margin: '0 auto' }}>
        <Button type="link" icon={<ArrowLeftOutlined />} style={{ padding: 0, marginBottom: 16 }} onClick={() => history.push('/plugins')}>
          {copy.back}
        </Button>

        <Card
          style={{
            borderRadius: 28,
            marginBottom: 24,
            background: 'linear-gradient(135deg, #0f172a 0%, #1e3a8a 55%, #0369a1 100%)',
            color: '#fff',
          }}
          styles={{ body: { padding: 28 } }}
        >
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Space wrap>
              <Tag color="cyan" bordered={false}>{detail.plugin.code}</Tag>
              <Tag color="green" bordered={false}>
                {copy.current}: {detail.release.version || '-'}
              </Tag>
            </Space>
            <Typography.Title style={{ color: '#fff', margin: 0 }}>{getDisplayName(locale, detail.plugin)}</Typography.Title>
            <Typography.Paragraph style={{ color: 'rgba(255,255,255,0.82)', margin: 0, maxWidth: 760 }}>
              {getDisplayDescription(locale, detail.plugin)}
            </Typography.Paragraph>
          </Space>
        </Card>

        <Row gutter={[20, 20]}>
          <Col xs={24} xl={16}>
            <Card style={{ borderRadius: 24, height: '100%' }}>
              <Space direction="vertical" size={24} style={{ width: '100%' }}>
                <Descriptions title={copy.intro} column={1} styles={{ label: { width: 120 } }}>
                  <Descriptions.Item label={copy.repo}>
                    <Typography.Link href={detail.plugin.repositoryUrl} target="_blank">
                      <LinkOutlined /> {detail.plugin.repositoryUrl}
                    </Typography.Link>
                  </Descriptions.Item>
                  <Descriptions.Item label={copy.compatible}>
                    <Space wrap>
                      {(selectedVersion?.compatibleItems || []).length ? (
                        selectedVersion?.compatibleItems?.map((item) => (
                          <Tag key={`${item.productId}-${item.versionConstraint || 'any'}`}>
                            {item.productName || item.productCode}
                            {item.versionConstraint ? ` ${item.versionConstraint}` : ''}
                          </Tag>
                        ))
                      ) : (
                        <Typography.Text type="secondary">-</Typography.Text>
                      )}
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label={copy.version}>{selectedVersion?.version || '-'}</Descriptions.Item>
                </Descriptions>

                <Card type="inner" title={copy.history} style={{ borderRadius: 18 }}>
                  <Space direction="vertical" size={16} style={{ width: '100%' }}>
                    <Select
                      value={selectedVersionId}
                      style={{ width: '100%' }}
                      onChange={setSelectedVersionId}
                      options={(detail.versions || []).map((item) => ({
                        value: item.ID,
                        label: `${item.version}${item.releasedAt ? ` · ${item.releasedAt.slice(0, 10)}` : ''}`,
                      }))}
                    />
                    <Typography.Paragraph style={{ marginBottom: 0 }}>
                      {getDisplayChangelog(locale, selectedVersion)}
                    </Typography.Paragraph>
                  </Space>
                </Card>
              </Space>
            </Card>
          </Col>

          <Col xs={24} xl={8}>
            <Space direction="vertical" size={20} style={{ width: '100%' }}>
              <Card title={copy.download} style={{ borderRadius: 24 }}>
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  <Button
                    block
                    type="primary"
                    icon={<CloudDownloadOutlined />}
                    href={selectedVersion?.packageX86Url}
                    target="_blank"
                    disabled={!selectedVersion?.packageX86Url}
                  >
                    x86_64
                  </Button>
                  <Button
                    block
                    href={selectedVersion?.packageArmUrl}
                    target="_blank"
                    disabled={!selectedVersion?.packageArmUrl}
                  >
                    ARM64
                  </Button>
                </Space>
              </Card>

              <Card title={copy.report} style={{ borderRadius: 24 }}>
                {selectedVersion?.testReportUrl ? (
                  <Typography.Link href={selectedVersion.testReportUrl} target="_blank">
                    {selectedVersion.testReportUrl}
                  </Typography.Link>
                ) : (
                  <Typography.Text type="secondary">-</Typography.Text>
                )}
              </Card>
            </Space>
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default PluginPublicDetailPage;
