import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Divider, Drawer, Empty, Input, List, Pagination, Segmented, Select, Skeleton, Space, Spin, Tag, Typography, message } from 'antd';
import { AppstoreOutlined, BarsOutlined, LinkOutlined, ReloadOutlined, SyncOutlined } from '@ant-design/icons';
import { getMediaItem, listMediaItems, refreshMediaItem, type MediaDetail, type MediaItem, type PlayEpisode, type PlayLine } from '@/services/media';
import { Cover, MediaHeading } from '../shared';

const typeMeta: Record<string, { label: string; color: string }> = {
  movie: { label: '电影', color: 'blue' }, series: { label: '电视剧', color: 'purple' },
  anime: { label: '动漫', color: 'cyan' }, variety: { label: '综艺', color: 'gold' },
  电影: { label: '电影', color: 'blue' }, 电视剧: { label: '电视剧', color: 'purple' },
  动漫: { label: '动漫', color: 'cyan' }, 综艺: { label: '综艺', color: 'gold' },
};
const typeOptions = ['movie', 'series', 'anime', 'variety'];
const text = (value?: string) => value?.trim() || '暂无';
const typeTags = (value?: string) => Array.from(new Set((value || '').split(/[,，/|、\s]+/).map(v => v.trim()).filter(Boolean))).map(value => ({ raw: value, ...(typeMeta[value.toLowerCase()] || { label: value, color: 'default' }) }));
const TypeTags = ({ value }: { value?: string }) => <Space size={4} wrap>{(typeTags(value).length ? typeTags(value) : [{ raw: '', label: '未分类', color: 'default' }]).map(tag => <Tag key={tag.raw || tag.label} color={tag.color}>{tag.label}</Tag>)}</Space>;

export default function Library() {
  const [rows, setRows] = useState<MediaItem[]>([]), [total, setTotal] = useState(0), [page, setPage] = useState(1), [loading, setLoading] = useState(false);
  const [query, setQuery] = useState(''), [kind, setKind] = useState(''), [view, setView] = useState<'card' | 'list'>('card');
  const [selected, setSelected] = useState<MediaItem>(), [detail, setDetail] = useState<MediaDetail>(), [detailLoading, setDetailLoading] = useState(false), [refreshing, setRefreshing] = useState(false);

  const load = async (nextPage = page, nextQuery = query, nextKind = kind) => {
    setLoading(true);
    try {
      const result = await listMediaItems(nextPage, nextQuery, nextKind);
      if (result.code !== 0) throw Error(result.msg);
      setRows(result.data.list || []); setTotal(result.data.total || 0); setPage(nextPage);
    } catch { message.error('加载影视库失败'); } finally { setLoading(false); }
  };
  useEffect(() => { void load(1, query, kind); }, [kind]);
  const search = (value = query) => { setPage(1); void load(1, value, kind); };
  const open = async (item: MediaItem) => {
    setSelected(item); setDetail(undefined); setDetailLoading(true);
    try { const result = await getMediaItem(item.id); if (result.code !== 0) throw Error(result.msg); setDetail(result.data); }
    catch { message.error('加载资源详情失败'); } finally { setDetailLoading(false); }
  };
  const refresh = async () => {
    if (!detail) return; setRefreshing(true);
    try {
      const result = await refreshMediaItem(detail.item.id); if (result.code !== 0) throw Error(result.msg);
      const latest = await getMediaItem(detail.item.id); if (latest.code === 0) setDetail(latest.data); message.success('播放链接已刷新');
    } catch (error) { message.error(error instanceof Error ? error.message : '刷新失败'); } finally { setRefreshing(false); }
  };
  return <PageContainer title={false}><div className="media-workspace">
    <MediaHeading eyebrow="MEDIA / LIBRARY" title="影视库" description="统一浏览采集到的影视资源，按来源查看可用线路和选集。" extra={<Space><Segmented value={view} onChange={value => setView(value as 'card' | 'list')} options={[{ value: 'card', icon: <AppstoreOutlined />, label: '卡片' }, { value: 'list', icon: <BarsOutlined />, label: '列表' }]} /><Button icon={<ReloadOutlined />} onClick={() => load()}>刷新</Button></Space>} />
    <Card className="media-panel" bodyStyle={{ padding: '18px 22px' }}><div className="media-toolbar" style={{ marginBottom: 0 }}>
      <Space wrap><Input.Search allowClear value={query} onChange={event => setQuery(event.target.value)} onSearch={search} placeholder="搜索片名、别名" style={{ width: 300 }} /><Select allowClear placeholder="全部类型" style={{ width: 150 }} value={kind || undefined} onChange={value => { setKind(value || ''); setPage(1); }} options={typeOptions.map(value => ({ value, label: typeMeta[value].label }))} /></Space>
      <Typography.Text type="secondary">共 {total} 部</Typography.Text>
    </div></Card>
    <Spin spinning={loading}><div style={{ marginTop: 24 }}>{!rows.length && !loading ? <Empty description="还没有采集到影视资源" /> : view === 'card' ? <div className="media-library-grid">{rows.map(item => <button className="media-poster" key={item.id} onClick={() => open(item)}><Cover url={item.coverUrl} title={item.title} /><span className="media-poster-title">{item.title}</span><Space size={5} align="start"><TypeTags value={item.mediaType} />{item.year && <Typography.Text type="secondary">{item.year}</Typography.Text>}</Space></button>)}</div> : <Card className="media-panel"><List dataSource={rows} renderItem={item => <List.Item key={item.id} actions={[<Button type="link" key="detail" onClick={() => open(item)}>查看详情</Button>]}><List.Item.Meta avatar={<Cover url={item.coverUrl} title={item.title} small />} title={<Typography.Text strong ellipsis={{ tooltip: item.title }}>{item.title}</Typography.Text>} description={<div className="media-list-meta"><Space wrap><TypeTags value={item.mediaType} /><span>{text(item.year)}</span><span>{text(item.area)}</span></Space><Typography.Paragraph className="media-list-description" ellipsis={{ rows: 2, tooltip: item.description || '暂无简介' }}>{item.description || '暂无简介'}</Typography.Paragraph></div>} /></List.Item>} /></Card>}</div></Spin>
    {total > 0 && <div className="media-pagination"><Pagination current={page} total={total} pageSize={20} showSizeChanger={false} showQuickJumper onChange={nextPage => void load(nextPage)} /></div>}
    <Drawer width={960} title={null} open={!!selected} onClose={() => { setSelected(undefined); setDetail(undefined); }} extra={detail && <Space><Button icon={<SyncOutlined />} loading={refreshing} onClick={refresh}>刷新播放链接</Button><Button type="primary" onClick={() => { window.location.href = '/media/player?id=' + detail.item.id; }}>播放测试</Button></Space>}>
      {detailLoading ? <Skeleton active paragraph={{ rows: 10 }} /> : detail ? <Detail detail={detail} /> : <Empty description="未找到详情" />}
    </Drawer>
  </div></PageContainer>;
}

function Detail({ detail }: { detail: MediaDetail }) {
  const item = detail.item;
  return <div><div className="media-detail-hero"><Cover url={item.coverUrl} title={item.title} /><div><Typography.Title level={2}>{item.title}</Typography.Title><Space wrap><TypeTags value={item.mediaType} /><Tag>{text(item.year)}</Tag><Tag>{text(item.area)}</Tag><Tag>{text(item.language)}</Tag></Space><Typography.Paragraph type="secondary" style={{ marginTop: 18 }}>导演：{text(item.director)}<br />主演：{text(item.actors)}</Typography.Paragraph><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{item.description || '暂无介绍内容。'}</Typography.Paragraph></div></div><Divider /><Typography.Title level={4}>播放来源</Typography.Title>
    {!detail.playbacks.length ? <Empty description="没有可用播放来源" /> : <Space direction="vertical" size={18} style={{ width: '100%' }}>{detail.playbacks.map(playback => <Card key={playback.sourceItem.id} className="media-panel" title={<Space><Typography.Text strong>{playback.source.name}</Typography.Text><Tag>{playback.lines.length} 条线路</Tag></Space>} extra={<Typography.Text type="secondary">来源条目：{playback.sourceItem.externalTitle || playback.sourceItem.externalId || '—'}</Typography.Text>}>{!playback.lines.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有可用线路" /> : <Space direction="vertical" size={18} style={{ width: '100%' }}>{playback.lines.map(line => <PlayLineBlock key={line.id} line={line} />)}</Space>}</Card>)}</Space>}
  </div>;
}
function PlayLineBlock({ line }: { line: PlayLine }) {
  const [page, setPage] = useState(1); const pageSize = 12; const start = (page - 1) * pageSize;
  return <div className="media-line-block"><Space><Typography.Text strong>{line.lineName}</Typography.Text><Tag>{line.episodes.length} 集</Tag></Space><div className="media-episodes">{line.episodes.slice(start, start + pageSize).map(episode => <Episode key={episode.id} episode={episode} />)}</div>{line.episodes.length > pageSize && <div className="media-pagination"><Pagination size="small" current={page} total={line.episodes.length} pageSize={pageSize} showSizeChanger={false} onChange={setPage} /></div>}</div>;
}
function Episode({ episode }: { episode: PlayEpisode }) {
  const title = episode.episodeName && !/^(m3u8|mp4|hls|播放)$/i.test(episode.episodeName.trim()) ? episode.episodeName : '第 ' + episode.episodeNo + ' 集';
  return <Card size="small" hoverable title={title} extra={<Tag color={episode.lastStatus === 'ok' ? 'green' : 'default'}>{episode.streamType || '播放'}</Tag>} actions={[<Typography.Link key="copy" copyable={{ text: episode.streamUrl }}><LinkOutlined /> 复制链接</Typography.Link>, <Typography.Link key="open" href={episode.streamUrl} target="_blank" rel="noreferrer">打开</Typography.Link>]}><Typography.Text type="secondary" ellipsis={{ tooltip: episode.streamUrl }} style={{ display: 'block' }}>{episode.streamUrl}</Typography.Text></Card>;
}
