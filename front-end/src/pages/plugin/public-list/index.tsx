import React, { useEffect, useMemo, useState } from 'react';
import { getLocale, history } from '@umijs/max';
import {
  AppstoreOutlined,
  ArrowRightOutlined,
  SearchOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Button, Card, Col, Empty, Input, Row, Space, Spin, Tag, Typography, message } from 'antd';

import { getPublishedPluginList, type PublishedPluginItem } from '@/services/api/plugin';
import { getDisplayDescription, getDisplayName, isEnglishLocale } from '@/utils/plugin';

const pageCopy = {
  zh: {
    title: '插件生态中心',
    subtitle: '面向公开访问的插件目录，集中展示已发布版本、兼容产品与下载入口。',
    search: '搜索插件名称、编码或描述',
    searchButton: '搜索',
    stats: '已发布插件',
    latest: '最新版本',
    compatible: '兼容产品',
    detail: '查看详情',
    empty: '暂无可展示的已发布插件',
  },
  en: {
    title: 'Plugin Marketplace',
    subtitle: 'A public catalogue for released plugins, compatibility details, and download entry points.',
    search: 'Search by plugin name, code, or description',
    searchButton: 'Search',
    stats: 'Released Plugins',
    latest: 'Latest Version',
    compatible: 'Compatible Products',
    detail: 'View Details',
    empty: 'No released plugins available.',
  },
};

const cardStyle: React.CSSProperties = {
  height: '100%',
  borderRadius: 20,
  border: '1px solid #d9e2f2',
  boxShadow: '0 18px 40px rgba(15, 23, 42, 0.08)',
};

const PluginPublicListPage: React.FC = () => {
  const locale = getLocale();
  const copy = isEnglishLocale(locale) ? pageCopy.en : pageCopy.zh;
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [searchText, setSearchText] = useState('');
  const [items, setItems] = useState<PublishedPluginItem[]>([]);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const res = await getPublishedPluginList({ page: 1, pageSize: 200 }, { skipErrorHandler: true });
        if (res.code !== 0) {
          message.error(res.msg || copy.empty);
          return;
        }
        setItems(res.data?.list || []);
      } finally {
        setLoading(false);
      }
    };
    void load();
  }, [copy.empty]);

  const filteredItems = useMemo(() => {
    const normalized = searchText.trim().toLowerCase();
    if (!normalized) return items;
    return items.filter((item) => {
      const searchable = [
        item.code,
        item.nameZh,
        item.nameEn,
        item.descriptionZh,
        item.descriptionEn,
      ]
        .join(' ')
        .toLowerCase();
      return searchable.includes(normalized);
    });
  }, [items, searchText]);

  const runSearch = () => setSearchText(keyword);

  return (
    <div
      style={{
        minHeight: '100vh',
        background:
          'linear-gradient(180deg, #eef4ff 0%, #f8fafc 28%, #ffffff 100%)',
        padding: '32px 24px 48px',
      }}
    >
      <div style={{ maxWidth: 1240, margin: '0 auto' }}>
        <div
          style={{
            borderRadius: 32,
            padding: '36px 32px',
            background:
              'radial-gradient(circle at top left, rgba(14,116,144,0.18), transparent 32%), linear-gradient(135deg, #0f172a 0%, #1d4ed8 52%, #0369a1 100%)',
            color: '#fff',
            marginBottom: 28,
          }}
        >
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Tag
              bordered={false}
              style={{ width: 'fit-content', borderRadius: 999, padding: '6px 12px', background: 'rgba(255,255,255,0.14)', color: '#fff' }}
            >
              <AppstoreOutlined /> {copy.stats}: {items.length}
            </Tag>
            <Typography.Title style={{ color: '#fff', margin: 0 }}>{copy.title}</Typography.Title>
            <Typography.Paragraph style={{ color: 'rgba(255,255,255,0.86)', maxWidth: 760, margin: 0 }}>
              {copy.subtitle}
            </Typography.Paragraph>
            <Space.Compact style={{ maxWidth: 680, width: '100%' }}>
              <Input
                value={keyword}
                placeholder={copy.search}
                onChange={(event) => setKeyword(event.target.value)}
                onPressEnter={runSearch}
                prefix={<SearchOutlined />}
                size="large"
              />
              <Button size="large" type="primary" onClick={runSearch}>
                {copy.searchButton}
              </Button>
            </Space.Compact>
          </Space>
        </div>

        <Spin spinning={loading}>
          {filteredItems.length ? (
            <Row gutter={[20, 20]}>
              {filteredItems.map((item) => (
                <Col xs={24} md={12} xl={8} key={item.ID}>
                  <Card style={cardStyle} styles={{ body: { padding: 24 } }}>
                    <Space direction="vertical" size={18} style={{ width: '100%' }}>
                      <Space align="start" style={{ justifyContent: 'space-between', width: '100%' }}>
                        <Space direction="vertical" size={4}>
                          <Typography.Title level={4} style={{ margin: 0 }}>
                            {getDisplayName(locale, item)}
                          </Typography.Title>
                          <Typography.Text type="secondary">{item.code}</Typography.Text>
                        </Space>
                        <Tag color="blue" bordered={false}>
                          {copy.latest}: {item.latestVersion || '-'}
                        </Tag>
                      </Space>

                      <Typography.Paragraph
                        type="secondary"
                        style={{ minHeight: 66, marginBottom: 0 }}
                        ellipsis={{ rows: 3 }}
                      >
                        {getDisplayDescription(locale, item)}
                      </Typography.Paragraph>

                      <Space direction="vertical" size={8} style={{ width: '100%' }}>
                        <Typography.Text strong>
                          <ThunderboltOutlined /> {copy.compatible}
                        </Typography.Text>
                        <Space wrap>
                          {(item.compatibleItems || []).length ? (
                            item.compatibleItems?.map((compatible) => (
                              <Tag key={`${item.ID}-${compatible.productId}-${compatible.versionConstraint || 'any'}`}>
                                {compatible.productName || compatible.productCode}
                                {compatible.versionConstraint ? ` ${compatible.versionConstraint}` : ''}
                              </Tag>
                            ))
                          ) : (
                            <Typography.Text type="secondary">-</Typography.Text>
                          )}
                        </Space>
                      </Space>

                      <Button type="link" style={{ padding: 0 }} onClick={() => history.push(`/plugins/${item.ID}`)}>
                        {copy.detail} <ArrowRightOutlined />
                      </Button>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          ) : (
            <Card style={{ ...cardStyle, textAlign: 'center' }}>
              <Empty description={copy.empty} />
            </Card>
          )}
        </Spin>
      </div>
    </div>
  );
};

export default PluginPublicListPage;
