import React, { useEffect, useMemo, useState } from 'react';
import { history } from '@umijs/max';
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
import { DownloadOutlined, SearchOutlined, AppstoreOutlined, FireOutlined, RightOutlined } from '@ant-design/icons';
import { getPublishedPluginList } from '@/services/api/plugin';

type PublishedPlugin = {
  pluginId: number;
  releaseId: number;
  code: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  capabilityZh: string;
  capabilityEn: string;
  owner: string;
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

type CategoryKey = 'all' | 'audit' | 'image' | 'edge' | 'data' | 'general';
type ArchKey = 'all' | 'dual' | 'x86' | 'arm';
type SortKey = 'latest' | 'name';

const categoryLabels: Record<CategoryKey, {zh: string, en: string}> = {
  all: { zh: '全部插件', en: 'All' },
  audit: { zh: '安全合规', en: 'Audit' },
  image: { zh: '图像处理', en: 'Image' },
  edge: { zh: '边缘设备', en: 'Edge' },
  data: { zh: '数据报表', en: 'Data' },
  general: { zh: '通用组件', en: 'General' },
};

const inferCategory = (plugin: PublishedPlugin): CategoryKey => {
  const text = `${plugin.nameZh} ${plugin.nameEn} ${plugin.descriptionZh} ${plugin.capabilityZh}`.toLowerCase();
  if (text.includes('audit') || text.includes('审计') || text.includes('合规') || text.includes('安全')) return 'audit';
  if (text.includes('image') || text.includes('图像') || text.includes('压缩') || text.includes('optimizer')) return 'image';
  if (text.includes('edge') || text.includes('device') || text.includes('runtime') || text.includes('设备') || text.includes('接入')) return 'edge';
  if (text.includes('data') || text.includes('report') || text.includes('分析') || text.includes('报表') || text.includes('数据')) return 'data';
  return 'general';
};

const matchArchitecture = (plugin: PublishedPlugin, arch: ArchKey) => {
  if (arch === 'all') return true;
  if (arch === 'dual') return Boolean(plugin.packageX86Url && plugin.packageArmUrl);
  if (arch === 'x86') return Boolean(plugin.packageX86Url);
  if (arch === 'arm') return Boolean(plugin.packageArmUrl);
  return true;
};

const formatTime = (value?: string) => value ? value.replace('T', ' ').replace('Z', '').slice(0, 10) : '-';

const PluginMarketPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [plugins, setPlugins] = useState<PublishedPlugin[]>([]);
  const [keyword, setKeyword] = useState('');
  const [searchValue, setSearchValue] = useState('');
  const [category, setCategory] = useState<CategoryKey>('all');
  const [arch, setArch] = useState<ArchKey>('all');
  const [sortKey, setSortKey] = useState<SortKey>('latest');
  const [lang, setLangState] = useState<'zh' | 'en'>(() => {
    return (localStorage.getItem('plugin-market-lang') as 'zh' | 'en') || 'zh';
  });

  const setLang = (l: 'zh' | 'en') => {
    setLangState(l);
    localStorage.setItem('plugin-market-lang', l);
  };

  const loadPlugins = async (nextKeyword = keyword) => {
    setLoading(true);
    const res = await getPublishedPluginList(
      { page: 1, pageSize: 60, keyword: nextKeyword || undefined },
      { skipErrorHandler: true }
    );
    setLoading(false);
    if (res.code !== 0) {
      message.error(res.msg || '加载已发布插件失败');
      return;
    }
    setPlugins(res.data?.list || []);
  };

  useEffect(() => {
    void loadPlugins('');
  }, []);

  const filteredPlugins = useMemo(() => {
    const result = plugins.filter((plugin) => {
      const pluginCategory = inferCategory(plugin);
      const searchText = `${plugin.code} ${plugin.nameZh} ${plugin.nameEn} ${plugin.descriptionZh} ${plugin.capabilityZh}`.toLowerCase();
      const hitKeyword = !searchValue || searchText.includes(searchValue.toLowerCase());
      const hitCategory = category === 'all' || pluginCategory === category;
      const hitArch = matchArchitecture(plugin, arch);
      return hitKeyword && hitCategory && hitArch;
    });

    if (sortKey === 'name') {
      return [...result].sort((a, b) => a.nameZh.localeCompare(b.nameZh));
    }
    return [...result].sort((a, b) => new Date(b.releasedAt).getTime() - new Date(a.releasedAt).getTime());
  }, [arch, category, plugins, searchValue, sortKey]);

  const doSearch = () => {
    setSearchValue(keyword.trim());
    void loadPlugins(keyword.trim());
  };

  return (
    <div style={{ minHeight: '100vh', background: '#f5f7fa', paddingBottom: 64 }}>
      {/* Hero Banner Section */}
      <div
        style={{
          background: '#fff',
          borderBottom: '1px solid #e8e8e8',
          padding: '48px 24px',
          textAlign: 'center',
          position: 'relative',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <Typography.Title level={2} style={{ margin: '0 0 16px', fontWeight: 600 }}>
            {lang === 'zh' ? 'Web-CMS 插件生态中心' : 'Web-CMS Plugin Ecosystem'}
          </Typography.Title>
          <Typography.Paragraph type="secondary" style={{ fontSize: 16, margin: '0 0 32px' }}>
            {lang === 'zh' ? '开箱即用的优质插件，加速构建您的企业级应用' : 'Out-of-the-box premium plugins to accelerate your enterprise application build'}
          </Typography.Paragraph>
          <Input
            size="large"
            placeholder={lang === 'zh' ? "搜索插件名称、编码或功能关键字" : "Search plugins by name, code or keywords"}
            prefix={<SearchOutlined style={{ color: '#bfbfbf' }} />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={doSearch}
            style={{ maxWidth: 640, borderRadius: 4, boxShadow: '0 2px 6px rgba(0,0,0,0.05)' }}
            suffix={
              <Button type="primary" onClick={doSearch} style={{ borderRadius: 2 }}>
                {lang === 'zh' ? '搜 索' : 'Search'}
              </Button>
            }
          />
        </div>
      </div>

      {/* Main Content Area */}
      <div style={{ maxWidth: 1200, margin: '24px auto 0', padding: '0 24px' }}>
        
        {/* Filter Bar */}
        <Card
          bordered={false}
          bodyStyle={{ padding: '20px 24px' }}
          style={{ borderRadius: 4, boxShadow: '0 1px 4px rgba(0,0,0,0.04)', marginBottom: 24 }}
        >
          <Space direction="vertical" size={20} style={{ width: '100%' }}>
            
            <Row align="middle" gutter={16}>
              <Col style={{ width: 80 }}><Typography.Text type="secondary">{lang === 'zh' ? '所属分类' : 'Category'}</Typography.Text></Col>
              <Col flex="auto">
                <Space wrap size={[8, 8]}>
                  {(Object.keys(categoryLabels) as CategoryKey[]).map((item) => (
                    <Tag.CheckableTag
                      key={item}
                      checked={category === item}
                      onChange={() => setCategory(item)}
                      style={{
                        padding: '4px 16px',
                        fontSize: 14,
                        borderRadius: 4,
                      }}
                    >
                      {categoryLabels[item][lang]}
                    </Tag.CheckableTag>
                  ))}
                </Space>
              </Col>
            </Row>

            <Row align="middle" gutter={16}>
              <Col style={{ width: 80 }}><Typography.Text type="secondary">{lang === 'zh' ? '支持架构' : 'Architecture'}</Typography.Text></Col>
              <Col flex="auto">
                <Space wrap size={[8, 8]}>
                  {([
                    { label: '全部架构', labelEn: 'All Arch', value: 'all' },
                    { label: '双架构提供', labelEn: 'Dual Arch', value: 'dual' },
                    { label: 'x86_64 标准版', labelEn: 'x86_64 Standard', value: 'x86' },
                    { label: 'ARM64 适配版', labelEn: 'ARM64 Adapted', value: 'arm' },
                  ] as const).map((item) => (
                    <Tag.CheckableTag
                      key={item.value}
                      checked={arch === item.value}
                      onChange={() => setArch(item.value as ArchKey)}
                      style={{
                        padding: '4px 16px',
                        fontSize: 14,
                        borderRadius: 4,
                      }}
                    >
                      {lang === 'zh' ? item.label : item.labelEn}
                    </Tag.CheckableTag>
                  ))}
                </Space>
              </Col>
            </Row>
            
            <Divider style={{ margin: '4px 0' }} />

            <Row align="middle" justify="space-between">
              <Col>
                <Space size={16}>
                  <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                    {lang === 'zh' ? (
                      <>共找到 <span style={{ color: '#1677ff', fontWeight: 600 }}>{filteredPlugins.length}</span> 个匹配的插件</>
                    ) : (
                      <>Found <span style={{ color: '#1677ff', fontWeight: 600 }}>{filteredPlugins.length}</span> matching plugins{' '}</>
                    )}
                  </Typography.Text>
                </Space>
              </Col>
              <Col>
                <Space align="center" size={16}>
                  <Radio.Group size="small" value={lang} onChange={(e) => setLang(e.target.value)} buttonStyle="solid">
                    <Radio.Button value="zh">中</Radio.Button>
                    <Radio.Button value="en">EN</Radio.Button>
                  </Radio.Group>
                  <Space align="center" size={8}>
                    <Typography.Text type="secondary" style={{ fontSize: 13 }}>{lang === 'zh' ? '排序方式:' : 'Sort by:'}</Typography.Text>
                    <Select
                      size="small"
                      value={sortKey}
                      onChange={(v) => setSortKey(v)}
                      bordered={false}
                      style={{ minWidth: 100 }}
                      options={[
                        { label: lang === 'zh' ? '最新上架' : 'Latest', value: 'latest' },
                        { label: lang === 'zh' ? '名称排序' : 'Name', value: 'name' },
                      ]}
                    />
                  </Space>
                </Space>
              </Col>
            </Row>
          </Space>
        </Card>

        {/* Plugin Grid */}
        <Spin spinning={loading}>
          {filteredPlugins.length > 0 ? (
            <Row gutter={[24, 24]}>
              {filteredPlugins.map((item) => {
                const isNew = new Date().getTime() - new Date(item.releasedAt).getTime() < 14 * 24 * 60 * 60 * 1000;
                
                return (
                  <Col xs={24} sm={12} lg={8} xl={6} key={`${item.pluginId}-${item.releaseId}`}>
                    <Card
                      hoverable
                      bordered
                      style={{
                        height: '100%',
                        display: 'flex',
                        flexDirection: 'column',
                        borderRadius: 4,
                        borderColor: '#f0f0f0',
                        boxShadow: '0 1px 2px -2px rgba(0, 0, 0, 0.08), 0 3px 6px 0 rgba(0, 0, 0, 0.06), 0 5px 12px 4px rgba(0, 0, 0, 0.04)',
                        transition: 'all 0.3s',
                      }}
                      bodyStyle={{ padding: '24px 24px 20px', flex: 1, display: 'flex', flexDirection: 'column' }}
                    >
                      <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <Typography.Title level={5} style={{ margin: 0, fontSize: 16, display: 'flex', alignItems: 'center', gap: 6 }}>
                          {lang === 'zh' ? item.nameZh : (item.nameEn || item.nameZh)}
                          {isNew && <FireOutlined style={{ color: '#ff4d4f' }} title={lang === 'zh' ? '最新发布' : 'New Release'} />}
                        </Typography.Title>
                        <Tag color="blue" bordered={false} style={{ margin: 0, borderRadius: 2 }}>v{item.version}</Tag>
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
                        {lang === 'zh' ? (item.descriptionZh || item.capabilityZh || '-') : (item.descriptionEn || item.capabilityEn || item.descriptionZh || '-')}
                      </Typography.Paragraph>

                      <div style={{ padding: '16px 0 0', borderTop: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Space size={6}>
                          {item.packageX86Url && <Tag color="default" style={{ margin: 0, borderRadius: 2, fontSize: 11, border: '1px solid #d9d9d9', background: '#fafafa' }}>x86_64</Tag>}
                          {item.packageArmUrl && <Tag color="default" style={{ margin: 0, borderRadius: 2, fontSize: 11, border: '1px solid #d9d9d9', background: '#fafafa' }}>ARM64</Tag>}
                        </Space>
                        <Button
                          type="link"
                          size="small"
                          style={{ padding: 0, fontSize: 13 }}
                          onClick={() => history.push(`/plugins/${item.pluginId}`)}
                        >
                          {lang === 'zh' ? '查看详情' : 'Details'} <RightOutlined />
                        </Button>
                      </div>
                    </Card>
                  </Col>
                );
              })}
            </Row>
          ) : (
            <Card bordered={false} style={{ borderRadius: 4, padding: '60px 0' }}>
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={<Typography.Text type="secondary">{lang === 'zh' ? '暂无符合条件的插件' : 'No matching plugins found'}</Typography.Text>}
              />
            </Card>
          )}
        </Spin>
      </div>
    </div>
  );
};

export default PluginMarketPage;