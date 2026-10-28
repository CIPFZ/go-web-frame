import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Drawer, Empty, Input, Pagination, Segmented, Skeleton, Space, Spin, Tabs, Tooltip, Typography, message } from 'antd';
import { CaretRightFilled, CloseOutlined, CopyOutlined, HistoryOutlined, ReloadOutlined, SortAscendingOutlined, SortDescendingOutlined, SyncOutlined } from '@ant-design/icons';
import { getMediaItem, listMediaItems, refreshMediaItem, type MediaDetail, type MediaItem, type PlayLine, type SourcePlayback } from '@/services/media';
import { Cover, EpisodeRanges, Facts, GenreTags, People, TypeTags, copyText, describe, episodeTitle, mediaTypeMeta, mediaTypeOptions, openPlayer, playerURL, readProgress, splitList } from '../shared';

const PAGE_SIZE = 20;
const EPISODE_RANGE = 60;
const SKELETON_KEYS = ['s1', 's2', 's3', 's4', 's5'];

export default function Library() {
  const [rows, setRows] = useState<MediaItem[]>([]), [total, setTotal] = useState(0), [page, setPage] = useState(1), [loading, setLoading] = useState(false);
  const [query, setQuery] = useState(''), [kind, setKind] = useState('');
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
  const search = (value = query) => { setPage(1); void load(1, value.trim(), kind); };
  const changePage = (nextPage: number) => { void load(nextPage); window.scrollTo({ top: 0, behavior: 'smooth' }); };
  const clearFilters = () => { setQuery(''); if (kind) setKind(''); else void load(1, '', ''); };
  const open = async (item: MediaItem) => {
    // The list row already carries the metadata, so the hero renders immediately while sources load.
    setSelected(item); setDetail(undefined); setDetailLoading(true);
    try { const result = await getMediaItem(item.id); if (result.code !== 0) throw Error(result.msg); setDetail(result.data); }
    catch { message.error('加载资源详情失败'); } finally { setDetailLoading(false); }
  };
  const close = () => { setSelected(undefined); setDetail(undefined); };
  const refresh = async () => {
    if (!detail) return; setRefreshing(true);
    try {
      const result = await refreshMediaItem(detail.item.id); if (result.code !== 0) throw Error(result.msg);
      const latest = await getMediaItem(detail.item.id); if (latest.code === 0) setDetail(latest.data); message.success('播放链接已刷新');
    } catch (error) { message.error(error instanceof Error ? error.message : '刷新失败'); } finally { setRefreshing(false); }
  };

  return <PageContainer title={false}><div className="media-workspace">
    <div className="media-library-filter">
      <Input.Search size="large" allowClear enterButton="搜索" className="media-library-search" value={query} onChange={event => setQuery(event.target.value)} onSearch={value => search(value)} placeholder="搜索片名、别名或分类" />
      <div className="media-library-filter-row">
        <Segmented value={kind} onChange={value => setKind(String(value))} options={[{ value: '', label: '全部' }, ...mediaTypeOptions.map(value => ({ value, label: mediaTypeMeta[value].label }))]} />
        <Space size={12}><Typography.Text type="secondary">共<b className="media-library-total">{total.toLocaleString()}</b>部</Typography.Text><Button icon={<ReloadOutlined />} loading={loading} onClick={() => load()}>刷新</Button></Space>
      </div>
    </div>
    <Spin spinning={loading && rows.length > 0}>
      <div className="media-library-rows">
        {loading && !rows.length ? SKELETON_KEYS.map(key => <div key={key} className="media-library-skeleton"><Skeleton active avatar={{ shape: 'square', size: 76 }} paragraph={{ rows: 2 }} /></div>)
          : !rows.length ? <Empty className="media-library-empty" description={query || kind ? '没有找到符合条件的影视' : '还没有采集到影视资源'}>{(query || kind) && <Button onClick={clearFilters}>清除筛选</Button>}</Empty>
          : rows.map(item => <LibraryRow key={item.id} item={item} onOpen={() => open(item)} />)}
      </div>
    </Spin>
    {total > PAGE_SIZE && <div className="media-pagination"><Pagination current={page} total={total} pageSize={PAGE_SIZE} showSizeChanger={false} showQuickJumper onChange={changePage} /></div>}
    <Drawer width="min(1000px, 100vw)" open={!!selected} onClose={close} closable={false} destroyOnHidden rootClassName="media-detail-drawer" styles={{ header: { display: 'none' } }}>
      {selected && <Detail item={detail?.item || selected} detail={detail} loading={detailLoading} refreshing={refreshing} onRefresh={refresh} onClose={close} />}
    </Drawer>
  </div></PageContainer>;
}

function LibraryRow({ item, onOpen }: { item: MediaItem; onOpen: () => void }) {
  const director = splitList(item.director), actors = splitList(item.actors);
  // The title button stretches over the whole row (see .media-row-name::after) so the row stays clickable.
  return <div className="media-row">
    <Cover url={item.coverUrl} title={item.title} className="media-cover-list" />
    <div className="media-row-main">
      <div className="media-row-title"><button type="button" className="media-row-name" title={item.title} onClick={onOpen}>{item.title}</button>{item.originalTitle && item.originalTitle !== item.title && <span className="media-row-original" title={item.originalTitle}>{item.originalTitle}</span>}</div>
      <div className="media-row-meta"><TypeTags value={item.mediaType} /><GenreTags value={item.genres} /><Facts item={item} /></div>
      <p className="media-row-description">{describe(item.description) || '暂无简介'}</p>
      {(director.length > 0 || actors.length > 0) && <div className="media-row-people">{director.length > 0 && <><b>导演</b>{director.slice(0, 2).join(' / ')}</>}{actors.length > 0 && <><b>主演</b>{actors.slice(0, 4).join(' / ')}</>}</div>}
    </div>
    <div className="media-row-actions">
      <Button type="primary" icon={<CaretRightFilled />} onClick={() => openPlayer(playerURL(item.id))}>播放</Button>
      <Button onClick={onOpen}>详情</Button>
    </div>
  </div>;
}

function Detail({ item, detail, loading, refreshing, onRefresh, onClose }: { item: MediaItem; detail?: MediaDetail; loading: boolean; refreshing: boolean; onRefresh: () => void; onClose: () => void }) {
  const [progress] = useState(() => readProgress(item.id));
  const playbacks = detail?.playbacks || [];
  const firstLine = playbacks.flatMap(playback => playback.lines).find(line => line.episodes.length);
  const lineCount = playbacks.reduce((sum, playback) => sum + playback.lines.length, 0);
  const description = describe(item.description);
  const playFirst = (restart = false) => openPlayer(playerURL(item.id, firstLine?.id, firstLine?.episodes[0]?.id, restart));
  return <div>
    <div className="media-detail-hero">
      {item.coverUrl && <img className="media-detail-backdrop" src={item.coverUrl} alt="" aria-hidden referrerPolicy="no-referrer" onError={event => { event.currentTarget.style.display = 'none'; }} />}
      <Button className="media-detail-close" type="text" shape="circle" icon={<CloseOutlined />} aria-label="关闭" onClick={onClose} />
      <div className="media-detail-hero-body">
        <Cover url={item.coverUrl} title={item.title} className="media-cover-hero" />
        <div>
          <TypeTags value={item.mediaType} />
          <h2 className="media-detail-title">{item.title}</h2>
          {item.originalTitle && item.originalTitle !== item.title && <div className="media-detail-original">{item.originalTitle}</div>}
          <div><Facts item={item} /></div>
          <GenreTags value={item.genres} />
          <People item={item} />
          <Space wrap size={10}>
            {progress && firstLine
              ? <Button type="primary" size="large" icon={<HistoryOutlined />} onClick={() => openPlayer(playerURL(item.id, progress.lineId, progress.episodeId))}>继续观看 · {progress.episodeTitle}</Button>
              : <Button type="primary" size="large" icon={<CaretRightFilled />} disabled={!firstLine} loading={loading} onClick={() => playFirst()}>立即播放</Button>}
            {progress && firstLine && <Button size="large" ghost onClick={() => playFirst(true)}>从头播放</Button>}
            <Button size="large" ghost icon={<SyncOutlined />} loading={refreshing} disabled={!detail} onClick={onRefresh}>刷新播放链接</Button>
          </Space>
        </div>
      </div>
    </div>
    <div className="media-detail-body">
      <section className="media-detail-section">
        <h3 className="media-section-title">剧情简介</h3>
        <Typography.Paragraph className="media-detail-description" ellipsis={description ? { rows: 4, expandable: 'collapsible', symbol: expanded => expanded ? '收起' : '展开全部' } : false}>{description || '暂无简介。'}</Typography.Paragraph>
      </section>
      <section className="media-detail-section">
        <div className="media-section-head"><h3 className="media-section-title">播放来源</h3>{detail && playbacks.length > 0 && <Typography.Text type="secondary">{playbacks.length} 个来源 · {lineCount} 条线路</Typography.Text>}</div>
        {loading ? <Skeleton active paragraph={{ rows: 4 }} />
          : !playbacks.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有可用播放来源" />
          : <Tabs className="media-source-tabs" items={playbacks.map(playback => ({
            key: String(playback.sourceItem.id),
            label: <span>{playback.source.name}<span className="media-tab-count">{playback.lines.reduce((sum, line) => sum + line.episodes.length, 0)}</span></span>,
            children: <SourcePanel playback={playback} itemId={item.id} watchedId={progress?.episodeId} />,
          }))} />}
      </section>
    </div>
  </div>;
}

function SourcePanel({ playback, itemId, watchedId }: { playback: SourcePlayback; itemId: number; watchedId?: number }) {
  const [lineId, setLineId] = useState(playback.lines[0]?.id);
  const line = playback.lines.find(entry => entry.id === lineId) || playback.lines[0];
  const reference = playback.sourceItem.externalTitle || playback.sourceItem.externalId;
  if (!line) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有可用线路" />;
  return <div>
    {reference && <div className="media-source-ref">来源条目：{reference}</div>}
    {playback.lines.length > 1 && <div className="media-line-switch"><Segmented value={line.id} onChange={value => setLineId(Number(value))} options={playback.lines.map(entry => ({ value: entry.id, label: `${entry.lineName} · ${entry.episodes.length} 集` }))} /></div>}
    <EpisodeGrid key={line.id} line={line} itemId={itemId} watchedId={watchedId} />
  </div>;
}

function EpisodeGrid({ line, itemId, watchedId }: { line: PlayLine; itemId: number; watchedId?: number }) {
  const [reverse, setReverse] = useState(false), [range, setRange] = useState(0);
  const ordered = reverse ? [...line.episodes].reverse() : line.episodes;
  const visible = ordered.slice(range * EPISODE_RANGE, (range + 1) * EPISODE_RANGE);
  // "title$url" per line is the format most players and CMS importers accept.
  const copyAll = () => copyText(line.episodes.map(episode => episodeTitle(episode) + '$' + episode.streamUrl).join('\n'), '已复制本线路全部播放地址');
  if (!line.episodes.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="该线路暂无剧集" />;
  return <div>
    <div className="media-episode-toolbar">
      <Typography.Text type="secondary">{line.lineName} · 共 {line.episodes.length} 集</Typography.Text>
      <Space size={4}>
        {line.episodes.length > 1 && <Button size="small" type="text" icon={reverse ? <SortDescendingOutlined /> : <SortAscendingOutlined />} onClick={() => { setReverse(!reverse); setRange(0); }}>{reverse ? '倒序' : '正序'}</Button>}
        <Button size="small" type="text" icon={<CopyOutlined />} onClick={copyAll}>复制全部地址</Button>
      </Space>
    </div>
    <EpisodeRanges total={ordered.length} size={EPISODE_RANGE} reverse={reverse} value={range} onChange={setRange} />
    <div className="media-episode-grid">{visible.map(episode => <div key={episode.id} className={'media-episode-chip' + (episode.id === watchedId ? ' is-watched' : '')}>
      <a href={playerURL(itemId, line.id, episode.id)} target="_blank" rel="noreferrer" title={episode.id === watchedId ? '上次看到这里' : episodeTitle(episode)}>{episode.id === watchedId && <HistoryOutlined />}<span>{episodeTitle(episode)}</span></a>
      <Tooltip title="复制播放地址"><button type="button" className="media-episode-copy" aria-label="复制播放地址" onClick={() => copyText(episode.streamUrl, '播放地址已复制')}><CopyOutlined /></button></Tooltip>
    </div>)}</div>
  </div>;
}
