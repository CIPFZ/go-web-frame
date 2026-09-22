import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Col, Divider, Drawer, Empty, Image, Input, List, Row, Skeleton, Space, Tag, Typography, message } from 'antd';
import { LinkOutlined, ReloadOutlined, SyncOutlined } from '@ant-design/icons';
import { getMediaItem, listMediaItems, refreshMediaItem, type MediaDetail, type MediaItem, type PlayEpisode } from '@/services/media';

const typeColor: Record<string, string> = { 电影: 'blue', 电视剧: 'purple', 动漫: 'cyan', 综艺: 'gold' };
const textOrDash = (value?: string) => value?.trim() || '-';

export default function MediaLibraryPage() {
  const [rows, setRows] = useState<MediaItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [query, setQuery] = useState('');
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState<MediaDetail>();
  const [refreshing, setRefreshing] = useState(false);

  const load = async (page = 1) => {
    setLoading(true);
    try {
      const result = await listMediaItems(page, query);
      if (result.code === 0) { setRows(result.data.list || []); setTotal(result.data.total || 0); }
      else message.error(result.msg || '加载影视库失败');
    } catch { message.error('加载影视库失败'); }
    finally { setLoading(false); }
  };
  useEffect(() => { load(); }, []);

  const openDetail = async (item: MediaItem) => {
    setDetailOpen(true); setDetailLoading(true); setDetail(undefined);
    try {
      const result = await getMediaItem(item.id);
      if (result.code === 0) setDetail(result.data);
      else message.error(result.msg || '加载资源详情失败');
    } catch { message.error('加载资源详情失败'); }
    finally { setDetailLoading(false); }
  };
  const refreshDetail = async () => {
    if (!detail) return;
    setRefreshing(true);
    try {
      const result = await refreshMediaItem(detail.item.id);
      if (result.code !== 0) { message.error(result.msg || '刷新播放链接失败'); return; }
      const latest = await getMediaItem(detail.item.id);
      if (latest.code === 0) setDetail(latest.data);
      message.success('播放链接已刷新');
    } catch { message.error('刷新播放链接失败'); }
    finally { setRefreshing(false); }
  };
  const goPlayer = (id: number) => { window.location.href = '/media/player?id=' + id; };

  return <PageContainer
    title="影视库"
    subTitle="按统一资源浏览，详情中会列出所有资源站、播放线路和剧集链接"
    extra={<Space><Input allowClear placeholder="搜索标题" value={query} onChange={event => setQuery(event.target.value)} onPressEnter={() => load()} style={{ width: 260 }} /><Button icon={<ReloadOutlined />} onClick={() => load()}>刷新</Button></Space>}
  >
    <List
      loading={loading}
      grid={{ gutter: [16, 16], xs: 1, sm: 2, md: 3, lg: 4, xl: 5 }}
      dataSource={rows}
      pagination={{ total, pageSize: 20, showSizeChanger: false, onChange: load }}
      locale={{ emptyText: <Empty description="还没有采集到影视资源" /> }}
      renderItem={item => <List.Item>
        <Card hoverable onClick={() => openDetail(item)} cover={item.coverUrl ? <Image preview={false} height={250} style={{ objectFit: 'cover' }} src={item.coverUrl} fallback="/default_avatar.jpg" /> : <div style={{ height: 250, display: 'grid', placeItems: 'center', background: 'linear-gradient(135deg,#eef2ff,#f8fafc)', color: '#94a3b8' }}>暂无封面</div>} bodyStyle={{ padding: 14 }}>
          <Typography.Title level={5} ellipsis={{ tooltip: item.title }} style={{ margin: 0 }}>{item.title}</Typography.Title>
          <Space size={6} style={{ marginTop: 8 }} wrap>
            <Tag color={typeColor[item.mediaType] || 'default'}>{item.mediaType || '未分类'}</Tag>
            {item.year && <Tag>{item.year}</Tag>}
            {item.area && <Typography.Text type="secondary">{item.area}</Typography.Text>}
          </Space>
          <Typography.Paragraph type="secondary" ellipsis={{ rows: 2 }} style={{ margin: '10px 0 0', minHeight: 42 }}>{item.description || '暂无简介，点击查看可用来源。'}</Typography.Paragraph>
        </Card>
      </List.Item>}
    />
    <Drawer
      title={detail?.item.title || '资源详情'}
      width={920}
      open={detailOpen}
      onClose={() => setDetailOpen(false)}
      extra={detail && <Space><Button size="small" icon={<SyncOutlined />} loading={refreshing} onClick={refreshDetail}>刷新链接</Button><Button size="small" type="primary" onClick={() => goPlayer(detail.item.id)}>播放测试</Button></Space>}
    >
      {detailLoading ? <Skeleton active paragraph={{ rows: 8 }} /> : detail ? <Space direction="vertical" size={18} style={{ width: '100%' }}>
        <Row gutter={24}>
          <Col flex="180px">{detail.item.coverUrl ? <Image width={180} height={250} style={{ objectFit: 'cover', borderRadius: 8 }} src={detail.item.coverUrl} /> : <div style={{ width: 180, height: 250, display: 'grid', placeItems: 'center', background: '#f1f5f9', borderRadius: 8, color: '#94a3b8' }}>暂无封面</div>}</Col>
          <Col flex="auto">
            <Typography.Title level={3} style={{ marginTop: 0 }}>{detail.item.title}</Typography.Title>
            <Space wrap><Tag color={typeColor[detail.item.mediaType] || 'default'}>{detail.item.mediaType || '未分类'}</Tag><Typography.Text>年份：{textOrDash(detail.item.year)}</Typography.Text><Typography.Text>地区：{textOrDash(detail.item.area)}</Typography.Text><Typography.Text>语言：{textOrDash(detail.item.language)}</Typography.Text></Space>
            <DescriptionsBlock label="导演" value={detail.item.director} />
            <DescriptionsBlock label="主演" value={detail.item.actors} />
            <Typography.Paragraph style={{ marginTop: 16, whiteSpace: 'pre-wrap' }}>{detail.item.description || '暂无介绍内容。'}</Typography.Paragraph>
          </Col>
        </Row>
        <Divider orientation="left">资源站与播放链接（{detail.playbacks.length} 个来源）</Divider>
        {detail.playbacks.length === 0 ? <Empty description="当前没有可用播放来源" /> : detail.playbacks.map(playback => <Card key={playback.sourceItem.id} size="small" title={<Space><Typography.Text strong>{playback.source.name}</Typography.Text><Tag>外部 ID：{playback.sourceItem.externalId || '-'}</Tag></Space>} bordered>
          {playback.lines.length === 0 ? <Typography.Text type="secondary">该来源暂时没有播放线路</Typography.Text> : <Space direction="vertical" size={14} style={{ width: '100%' }}>
            {playback.lines.map(line => <div key={line.id}>
              <Space><Tag color="blue">{line.lineName}</Tag><Typography.Text type="secondary">{line.episodes.length} 集</Typography.Text></Space>
              <List<PlayEpisode> size="small" dataSource={line.episodes} renderItem={episode => <List.Item actions={[<Typography.Link key="link" href={episode.streamUrl} target="_blank" rel="noreferrer"><LinkOutlined /> 打开链接</Typography.Link>]}>
                <Space wrap><Tag>{episode.episodeName || `第 ${episode.episodeNo} 集`}</Tag><Typography.Text copyable={{ text: episode.streamUrl }} ellipsis style={{ maxWidth: 620 }}>{episode.streamUrl}</Typography.Text></Space>
              </List.Item>} />
            </div>)}
          </Space>}
        </Card>)}
      </Space> : <Empty description="未找到资源详情" />}
    </Drawer>
  </PageContainer>;
}

function DescriptionsBlock({ label, value }: { label: string; value?: string }) {
  return <Typography.Paragraph type="secondary" style={{ margin: '14px 0 0' }}><Typography.Text strong>{label}：</Typography.Text>{textOrDash(value)}</Typography.Paragraph>;
}