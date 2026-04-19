import React, { useEffect, useMemo, useState } from 'react';
import { history } from '@umijs/max';
import { FireOutlined, RightOutlined, SearchOutlined } from '@ant-design/icons';
import {
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Input,
  Radio,
  Row,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'antd';

import { getPublishedPluginList, type PublishedPluginItem } from '@/services/api/plugin';

type CategoryKey = 'all' | 'audit' | 'image' | 'edge' | 'data' | 'general';
type ArchKey = 'all' | 'dual' | 'x86' | 'arm';
type SortKey = 'latest' | 'name';

const copyMap = {
  zh: {
    title: '插件生态中心',
    subtitle: '开箱即用的优质插件，集中展示已发布版本、能力说明与下载入口。',
    searchPlaceholder: '搜索插件名称、编码或功能关键词',
    searchButton: '搜索',
    category: '所属分类',
    architecture: '支持架构',
    sort: '排序方式',
    found: '共找到',
    matches: '个匹配插件',
    latest: '最新上架',
    details: '查看详情',
    empty: '暂无符合条件的插件',
    loadFailed: '加载已发布插件失败',
    sortLatest: '最新发布',
    sortName: '名称排序',
    allArch: '全部架构',
    dualArch: '双架构提供',
    x86Arch: 'x86_64 标准版',
    armArch: 'ARM64 适配版',
  },
  en: {
    title: 'Plugin Ecosystem',
    subtitle: 'Out-of-the-box plugins with release details, capability summaries, and download entry points.',
    searchPlaceholder: 'Search plugins by name, code, or keywords',
    searchButton: 'Search',
    category: 'Category',
    architecture: 'Architecture',
    sort: 'Sort by',
    found: 'Found',
    matches: 'matching plugins',
    latest: 'New',
    details: 'Details',
    empty: 'No matching plugins found',
    loadFailed: 'Failed to load released plugins',
    sortLatest: 'Latest',
    sortName: 'Name',
    allArch: 'All Arch',
    dualArch: 'Dual Arch',
    x86Arch: 'x86_64 Standard',
    armArch: 'ARM64 Adapted',
  },
};

const categoryLabels: Record<CategoryKey, { zh: string; en: string }> = {
  all: { zh: '全部插件', en: 'All' },
  audit: { zh: '安全合规', en: 'Audit' },
  image: { zh: '图像处理', en: 'Image' },
  edge: { zh: '边缘设备', en: 'Edge' },
  data: { zh: '数据分析', en: 'Data' },
  general: { zh: '通用组件', en: 'General' },
};

const inferCategory = (plugin: PublishedPluginItem): CategoryKey => {
  const text = [
    plugin.nameZh,
    plugin.nameEn,
    plugin.descriptionZh,
    plugin.descriptionEn,
    plugin.code,
  ]
    .join(' ')
    .toLowerCase();

  if (/(audit|安全|合规|审计)/.test(text)) return 'audit';
  if (/(image|图像|图片|压缩|optimizer)/.test(text)) return 'image';
  if (/(edge|device|runtime|设备|边缘|接入)/.test(text)) return 'edge';
  if (/(data|report|报表|分析|数据)/.test(text)) return 'data';
  return 'general';
};

const matchArchitecture = (plugin: PublishedPluginItem, arch: ArchKey) => {
  const hasX86 = Boolean(plugin.packageX86Url);
  const hasArm = Boolean(plugin.packageArmUrl);
  if (arch === 'all') return true;
  if (arch === 'dual') return hasX86 && hasArm;
  if (arch === 'x86') return hasX86;
  if (arch === 'arm') return hasArm;
  return true;
};

const formatReleaseDate = (value?: string) =>
  value ? value.replace('T', ' ').replace('Z', '').slice(0, 10) : '-';

const getName = (lang: 'zh' | 'en', item: PublishedPluginItem) =>
  lang === 'zh' ? item.nameZh : item.nameEn || item.nameZh;

const getDescription = (lang: 'zh' | 'en', item: PublishedPluginItem) =>
  lang === 'zh' ? item.descriptionZh || item.descriptionEn || '-' : item.descriptionEn || item.descriptionZh || '-';

const PluginPublicListPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<PublishedPluginItem[]>([]);
  const [keyword, setKeyword] = useState('');
  const [searchValue, setSearchValue] = useState('');
  const [category, setCategory] = useState<CategoryKey>('all');
  const [arch, setArch] = useState<ArchKey>('all');
  const [sortKey, setSortKey] = useState<SortKey>('latest');
  const [lang, setLang] = useState<'zh' | 'en'>(() => {
    return (localStorage.getItem('plugin-market-lang') as 'zh' | 'en') || 'zh';
  });

  const copy = copyMap[lang];

  const updateLang = (value: 'zh' | 'en') => {
    setLang(value);
    localStorage.setItem('plugin-market-lang', value);
  };

  const loadPlugins = async (nextKeyword = '') => {
    setLoading(true);
    try {
      const res = await getPublishedPluginList(
        { page: 1, pageSize: 60, keyword: nextKeyword || undefined },
        { skipErrorHandler: true },
      );
      if (res.code !== 0) {
        message.error(res.msg || copy.loadFailed);
        return;
      }
      setItems(res.data?.list || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadPlugins('');
  }, []);

  const filteredItems = useMemo(() => {
    const normalized = searchValue.trim().toLowerCase();
    const result = items.filter((item) => {
      const searchText = [item.code, item.nameZh, item.nameEn, item.descriptionZh, item.descriptionEn]
        .join(' ')
        .toLowerCase();
      const hitKeyword = !normalized || searchText.includes(normalized);
      const hitCategory = category === 'all' || inferCategory(item) === category;
      const hitArch = matchArchitecture(item, arch);
      return hitKeyword && hitCategory && hitArch;
    });

    if (sortKey === 'name') {
      return [...result].sort((a, b) => getName(lang, a).localeCompare(getName(lang, b)));
    }

    return [...result].sort(
      (a, b) => new Date(b.releasedAt || '').getTime() - new Date(a.releasedAt || '').getTime(),
    );
  }, [arch, category, items, lang, searchValue, sortKey]);

  const doSearch = () => {
    const nextValue = keyword.trim();
    setSearchValue(nextValue);
    void loadPlugins(nextValue);
  };

  return (
    <div style={{ minHeight: '100vh', background: '#f5f7fa', paddingBottom: 64 }}>
      <div
        style={{
          background: '#fff',
          borderBottom: '1px solid #e8e8e8',
          padding: '48px 24px',
          textAlign: 'center',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <Typography.Title level={2} style={{ margin: '0 0 16px', fontWeight: 600 }}>
            {copy.title}
          </Typography.Title>
          <Typography.Paragraph type="secondary" style={{ fontSize: 16, margin: '0 0 32px' }}>
            {copy.subtitle}
          </Typography.Paragraph>
          <Input
            size="large"
            placeholder={copy.searchPlaceholder}
            prefix={<SearchOutlined style={{ color: '#bfbfbf' }} />}
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            onPressEnter={doSearch}
            style={{ maxWidth: 640, borderRadius: 4, boxShadow: '0 2px 6px rgba(0,0,0,0.05)' }}
            suffix={
              <Button type="primary" onClick={doSearch} style={{ borderRadius: 2 }}>
                {copy.searchButton}
              </Button>
            }
          />
        </div>
      </div>

      <div style={{ maxWidth: 1200, margin: '24px auto 0', padding: '0 24px' }}>
        <Card
          variant="borderless"
          styles={{ body: { padding: '20px 24px' } }}
          style={{ borderRadius: 4, boxShadow: '0 1px 4px rgba(0,0,0,0.04)', marginBottom: 24 }}
        >
          <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <Row align="middle" gutter={16}>
              <Col style={{ width: 88 }}>
                <Typography.Text type="secondary">{copy.category}</Typography.Text>
              </Col>
              <Col flex="auto">
                <Space wrap size={[8, 8]}>
                  {(Object.keys(categoryLabels) as CategoryKey[]).map((item) => (
                    <Tag.CheckableTag
                      key={item}
                      checked={category === item}
                      onChange={() => setCategory(item)}
                      style={{ padding: '4px 16px', fontSize: 14, borderRadius: 4 }}
                    >
                      {categoryLabels[item][lang]}
                    </Tag.CheckableTag>
                  ))}
                </Space>
              </Col>
            </Row>

            <Row align="middle" gutter={16}>
              <Col style={{ width: 88 }}>
                <Typography.Text type="secondary">{copy.architecture}</Typography.Text>
              </Col>
              <Col flex="auto">
                <Space wrap size={[8, 8]}>
                  {[
                    { label: copy.allArch, value: 'all' },
                    { label: copy.dualArch, value: 'dual' },
                    { label: copy.x86Arch, value: 'x86' },
                    { label: copy.armArch, value: 'arm' },
                  ].map((item) => (
                    <Tag.CheckableTag
                      key={item.value}
                      checked={arch === item.value}
                      onChange={() => setArch(item.value as ArchKey)}
                      style={{ padding: '4px 16px', fontSize: 14, borderRadius: 4 }}
                    >
                      {item.label}
                    </Tag.CheckableTag>
                  ))}
                </Space>
              </Col>
            </Row>

            <Divider style={{ margin: '4px 0' }} />

            <Row align="middle" justify="space-between" gutter={[12, 12]}>
              <Col>
                <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                  {copy.found} <span style={{ color: '#1677ff', fontWeight: 600 }}>{filteredItems.length}</span>{' '}
                  {copy.matches}
                </Typography.Text>
              </Col>
              <Col>
                <Space align="center" size={16}>
                  <Radio.Group
                    size="small"
                    value={lang}
                    onChange={(event) => updateLang(event.target.value)}
                    buttonStyle="solid"
                  >
                    <Radio.Button value="zh">中</Radio.Button>
                    <Radio.Button value="en">EN</Radio.Button>
                  </Radio.Group>
                  <Space align="center" size={8}>
                    <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                      {copy.sort}
                    </Typography.Text>
                    <Select
                      size="small"
                      value={sortKey}
                      onChange={(value) => setSortKey(value)}
                      variant="borderless"
                      style={{ minWidth: 100 }}
                      options={[
                        { label: copy.sortLatest, value: 'latest' },
                        { label: copy.sortName, value: 'name' },
                      ]}
                    />
                  </Space>
                </Space>
              </Col>
            </Row>
          </Space>
        </Card>

        <Spin spinning={loading}>
          {filteredItems.length ? (
            <Row gutter={[24, 24]}>
              {filteredItems.map((item) => {
                const isNew =
                  Boolean(item.releasedAt) &&
                  new Date().getTime() - new Date(item.releasedAt!).getTime() < 14 * 24 * 60 * 60 * 1000;

                return (
                  <Col xs={24} sm={12} lg={8} xl={6} key={item.ID}>
                    <Card
                      hoverable
                      variant="outlined"
                      style={{
                        height: '100%',
                        display: 'flex',
                        flexDirection: 'column',
                        borderRadius: 4,
                        borderColor: '#f0f0f0',
                        boxShadow:
                          '0 1px 2px -2px rgba(0, 0, 0, 0.08), 0 3px 6px 0 rgba(0, 0, 0, 0.06), 0 5px 12px 4px rgba(0, 0, 0, 0.04)',
                      }}
                      styles={{ body: { padding: '24px 24px 20px', flex: 1, display: 'flex', flexDirection: 'column' } }}
                    >
                      <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <Typography.Title level={5} style={{ margin: 0, fontSize: 16, display: 'flex', alignItems: 'center', gap: 6 }}>
                          {getName(lang, item)}
                          {isNew ? <FireOutlined style={{ color: '#ff4d4f' }} title={copy.latest} /> : null}
                        </Typography.Title>
                        <Tag color="blue" bordered={false} style={{ margin: 0, borderRadius: 2 }}>
                          v{item.latestVersion || '-'}
                        </Tag>
                      </div>

                      <Typography.Paragraph
                        type="secondary"
                        style={{
                          fontSize: 13,
                          lineHeight: 1.6,
                          flex: 1,
                          marginBottom: 20,
                          display: '-webkit-box',
                          WebkitLineClamp: 3,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                        }}
                      >
                        {getDescription(lang, item)}
                      </Typography.Paragraph>

                      <div
                        style={{
                          padding: '16px 0 0',
                          borderTop: '1px solid #f0f0f0',
                          display: 'flex',
                          flexDirection: 'column',
                          gap: 14,
                        }}
                      >
                        <div
                          style={{
                            display: 'flex',
                            flexWrap: 'wrap',
                            alignItems: 'center',
                            gap: 8,
                            minHeight: 24,
                          }}
                        >
                          {item.packageX86Url ? (
                            <Tag
                              color="default"
                              style={{
                                margin: 0,
                                borderRadius: 2,
                                fontSize: 11,
                                border: '1px solid #d9d9d9',
                                background: '#fafafa',
                                lineHeight: '20px',
                              }}
                            >
                              x86_64
                            </Tag>
                          ) : null}
                          {item.packageArmUrl ? (
                            <Tag
                              color="default"
                              style={{
                                margin: 0,
                                borderRadius: 2,
                                fontSize: 11,
                                border: '1px solid #d9d9d9',
                                background: '#fafafa',
                                lineHeight: '20px',
                              }}
                            >
                              ARM64
                            </Tag>
                          ) : null}
                          {item.releasedAt ? (
                            <Typography.Text
                              type="secondary"
                              style={{
                                fontSize: 12,
                                whiteSpace: 'nowrap',
                                display: 'inline-flex',
                                alignItems: 'center',
                                padding: '2px 0',
                              }}
                            >
                              {formatReleaseDate(item.releasedAt)}
                            </Typography.Text>
                          ) : null}
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                          <Button
                            type="link"
                            size="small"
                            style={{ padding: 0, fontSize: 13, height: 'auto' }}
                            onClick={() => history.push(`/plugins/${item.ID}`)}
                          >
                          {copy.details} <RightOutlined />
                          </Button>
                        </div>
                      </div>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          ) : (
            <Card variant="borderless" style={{ borderRadius: 4, padding: '60px 0' }}>
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={<Typography.Text type="secondary">{copy.empty}</Typography.Text>} />
            </Card>
          )}
        </Spin>
      </div>
    </div>
  );
};

export default PluginPublicListPage;
