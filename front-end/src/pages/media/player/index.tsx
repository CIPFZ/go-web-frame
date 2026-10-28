import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Button, ConfigProvider, Empty, Select, Space, Spin, Switch, Tooltip, Typography, message, theme } from 'antd';
import { CopyOutlined, ExportOutlined, LeftOutlined, ReloadOutlined, SortAscendingOutlined, SortDescendingOutlined, StepBackwardOutlined, StepForwardOutlined, SwapOutlined } from '@ant-design/icons';
import Hls from 'hls.js';
import { getMediaItem, type MediaDetail, type PlayEpisode, type PlayLine } from '@/services/media';
import { Cover, EpisodeRanges, Facts, GenreTags, People, TypeTags, clockText, copyText, describe, episodeTitle, playerURL, readProgress, writeProgress } from '../shared';

const RANGE_SIZE = 100;
const AUTO_NEXT_KEY = 'media-auto-next';
const darkTheme = { algorithm: theme.darkAlgorithm, token: { colorPrimary: '#3b82f6', borderRadius: 8 } };

export default function MediaPlayerPage() {
  const params = useMemo(() => {
    const hashQuery = window.location.hash.includes('?') ? window.location.hash.split('?')[1] : '';
    return new URLSearchParams(hashQuery || window.location.search);
  }, []);
  const itemId = Number(params.get('id') || 0);
  const requestedLineId = Number(params.get('lineId') || 0);
  const requestedEpisodeId = Number(params.get('episodeId') || 0);
  const restart = params.get('restart') === '1';
  const [detail, setDetail] = useState<MediaDetail>();
  const [lineId, setLineId] = useState<number>();
  const [episodeId, setEpisodeId] = useState<number>();
  const [loading, setLoading] = useState(true);
  const [playerState, setPlayerState] = useState<'idle' | 'loading' | 'ready' | 'error'>('idle');
  const [playerError, setPlayerError] = useState('');
  const [playing, setPlaying] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);
  const [reverse, setReverse] = useState(false);
  const [range, setRange] = useState(0);
  const [autoNext, setAutoNext] = useState(() => localStorage.getItem(AUTO_NEXT_KEY) !== '0');
  const video = useRef<HTMLVideoElement>(null);
  const grid = useRef<HTMLDivElement>(null);
  const hls = useRef<Hls | undefined>(undefined);
  // Position to seek to once the next stream reports its metadata.
  const resumeAt = useRef(0);
  const onEnded = useRef<() => void>(() => undefined);

  const lines = useMemo(() => detail?.playbacks.flatMap((playback) => playback.lines.map((line) => ({ ...line, sourceName: line.sourceName || playback.source.name }))) || [], [detail]);
  const selectedLine = lines.find((line) => line.id === lineId) || lines[0];
  const selectedEpisode = selectedLine?.episodes.find((episode) => episode.id === episodeId) || selectedLine?.episodes[0];
  const episodeIndex = selectedLine && selectedEpisode ? selectedLine.episodes.indexOf(selectedEpisode) : -1;
  const hasPrev = episodeIndex > 0;
  const hasNext = !!selectedLine && episodeIndex >= 0 && episodeIndex < selectedLine.episodes.length - 1;
  const ordered = useMemo(() => (selectedLine ? (reverse ? [...selectedLine.episodes].reverse() : selectedLine.episodes) : []), [selectedLine, reverse]);
  const nextLine = lines.length > 1 && selectedLine ? lines[(lines.indexOf(selectedLine) + 1) % lines.length] : undefined;

  useEffect(() => {
    if (!itemId) {
      setLoading(false);
      return;
    }
    getMediaItem(itemId)
      .then((result) => {
        if (result.code !== 0) throw Error(result.msg || '加载失败');
        setDetail(result.data);
        const all = result.data.playbacks.flatMap((playback) => playback.lines);
        const saved = restart ? undefined : readProgress(itemId);
        // An explicit episode wins; otherwise continue where this browser left off.
        let wantedLine = requestedLineId;
        let wantedEpisode = requestedEpisodeId;
        if (!wantedEpisode && saved && (!wantedLine || wantedLine === saved.lineId)) {
          wantedLine = saved.lineId;
          wantedEpisode = saved.episodeId;
        }
        const line = all.find((entry) => entry.id === wantedLine) || all.find((entry) => entry.episodes.some((episode) => episode.id === wantedEpisode)) || all.find((entry) => entry.episodes.length) || all[0];
        const episode = line?.episodes.find((entry) => entry.id === wantedEpisode) || line?.episodes[0];
        resumeAt.current = saved && episode && saved.episodeId === episode.id ? saved.time : 0;
        setLineId(line?.id);
        setEpisodeId(episode?.id);
      })
      .catch((error) => message.error(error instanceof Error ? error.message : '加载播放内容失败'))
      .finally(() => setLoading(false));
  }, [itemId, requestedEpisodeId, requestedLineId, restart]);

  useEffect(() => {
    const element = video.current;
    if (!element || !selectedLine || !selectedEpisode?.streamUrl) return;
    let disposed = false;
    let lastSaved = 0;
    setPlayerState('loading');
    setPlayerError('');
    setPlaying(false);
    hls.current?.destroy();
    hls.current = undefined;
    element.pause();
    element.removeAttribute('src');
    element.load();

    const line = selectedLine;
    const episode = selectedEpisode;
    const url = episode.streamUrl;
    const playbackURL = `/api/v1/media/episodes/${episode.id}/stream`;
    const type = (episode.streamType || '').toLowerCase();
    const isHls = type === 'm3u8' || type === 'hls' || /\.m3u8(?:$|[?#])/i.test(url);
    const save = (time = element.currentTime || 0) => writeProgress(itemId, { lineId: line.id, episodeId: episode.id, episodeTitle: episodeTitle(episode), time: Math.floor(time), updatedAt: Date.now() });
    const fail = (text: string) => {
      if (disposed) return;
      setPlayerState('error');
      setPlayerError(text);
    };
    const onMeta = () => {
      if (disposed) return;
      setPlayerState('ready');
      const at = resumeAt.current;
      resumeAt.current = 0;
      if (at > 5 && (!Number.isFinite(element.duration) || at < element.duration - 10)) {
        element.currentTime = at;
        message.info(`已从 ${clockText(at)} 继续播放`);
      }
      void element.play().catch(() => undefined);
    };
    const onTime = () => {
      if (Math.abs(element.currentTime - lastSaved) >= 5) {
        lastSaved = element.currentTime;
        save();
      }
    };
    const onPlay = () => setPlaying(true);
    const onPause = () => {
      setPlaying(false);
      save();
    };
    const onEnd = () => {
      setPlaying(false);
      onEnded.current();
    };
    const onVideoError = () => fail('播放器无法加载该地址，可能是来源不允许跨域播放或链接已失效。');
    const onLeave = () => save();
    element.addEventListener('loadedmetadata', onMeta);
    element.addEventListener('timeupdate', onTime);
    element.addEventListener('play', onPlay);
    element.addEventListener('pause', onPause);
    element.addEventListener('ended', onEnd);
    element.addEventListener('error', onVideoError);
    window.addEventListener('pagehide', onLeave);
    // Record the episode right away so "continue watching" follows episode switches.
    save(resumeAt.current);

    if (isHls && Hls.isSupported()) {
      const instance = new Hls({ enableWorker: true, lowLatencyMode: false });
      instance.on(Hls.Events.ERROR, (_event, data) => {
        if (data.fatal) fail('m3u8 播放失败，可能是来源跨域限制或播放地址已过期。');
      });
      instance.loadSource(playbackURL);
      instance.attachMedia(element);
      hls.current = instance;
    } else if (!isHls || element.canPlayType('application/vnd.apple.mpegurl')) {
      element.src = playbackURL;
      element.load();
    } else {
      fail('当前浏览器不支持 m3u8 播放，请使用新版 Chrome、Edge 或 Safari。');
    }

    return () => {
      if (element.currentTime > 0) save();
      disposed = true;
      element.removeEventListener('loadedmetadata', onMeta);
      element.removeEventListener('timeupdate', onTime);
      element.removeEventListener('play', onPlay);
      element.removeEventListener('pause', onPause);
      element.removeEventListener('ended', onEnd);
      element.removeEventListener('error', onVideoError);
      window.removeEventListener('pagehide', onLeave);
      hls.current?.destroy();
      hls.current = undefined;
    };
  }, [selectedEpisode?.id, selectedEpisode?.streamUrl, selectedEpisode?.streamType, reloadKey]);

  // Keep the range containing the current episode visible and scroll it into the middle of the grid.
  useEffect(() => {
    const index = ordered.findIndex((episode) => episode.id === selectedEpisode?.id);
    if (index >= 0) setRange(Math.floor(index / RANGE_SIZE));
  }, [selectedEpisode?.id, ordered]);
  useEffect(() => {
    const box = grid.current;
    const active = box?.querySelector<HTMLElement>('.is-active');
    if (box && active) box.scrollTop = active.offsetTop - box.clientHeight / 2 + active.offsetHeight / 2;
  }, [selectedEpisode?.id, range, reverse, detail]);
  useEffect(() => {
    if (!detail) return undefined;
    const previous = document.title;
    document.title = [detail.item.title, selectedEpisode && episodeTitle(selectedEpisode)].filter(Boolean).join(' · ');
    return () => {
      document.title = previous;
    };
  }, [detail, selectedEpisode?.id]);

  const selectEpisode = (line: PlayLine, episode: PlayEpisode) => {
    const saved = readProgress(itemId);
    resumeAt.current = saved?.episodeId === episode.id ? saved.time : 0;
    setLineId(line.id);
    setEpisodeId(episode.id);
    window.history.replaceState({}, '', playerURL(itemId, line.id, episode.id));
  };
  // Switching sources keeps the same episode number when the other line has it.
  const switchLine = (line?: PlayLine) => {
    if (!line?.episodes.length) return;
    selectEpisode(line, line.episodes.find((episode) => episode.episodeNo === selectedEpisode?.episodeNo) || line.episodes[0]);
  };
  const moveEpisode = (offset: number) => {
    const next = selectedLine?.episodes[episodeIndex + offset];
    if (selectedLine && next) selectEpisode(selectedLine, next);
  };
  onEnded.current = () => {
    if (autoNext && hasNext) moveEpisode(1);
  };
  const toggleAutoNext = (checked: boolean) => {
    setAutoNext(checked);
    localStorage.setItem(AUTO_NEXT_KEY, checked ? '1' : '0');
  };
  const openOriginal = () => {
    if (selectedEpisode) window.open(selectedEpisode.streamUrl, '_blank', 'noopener,noreferrer');
  };

  if (loading || !detail) {
    return (
      <ConfigProvider theme={darkTheme}>
        <div className="mp-page mp-center">
          {loading ? <Spin size="large" /> : <Empty description="没有找到播放内容"><Button href="/#/media/library">返回影视库</Button></Empty>}
        </div>
      </ConfigProvider>
    );
  }

  const item = detail.item;
  const description = describe(item.description);
  const visible = ordered.slice(range * RANGE_SIZE, (range + 1) * RANGE_SIZE);
  return (
    <ConfigProvider theme={darkTheme}>
      <div className="mp-page">
        {item.coverUrl && <img className="mp-backdrop" src={item.coverUrl} alt="" aria-hidden referrerPolicy="no-referrer" onError={(event) => { event.currentTarget.style.display = 'none'; }} />}
        <header className="mp-topbar">
          <a className="mp-back" href="/#/media/library"><LeftOutlined /> 影视库</a>
          <div className="mp-topbar-title">{item.title}{selectedEpisode && <em>{episodeTitle(selectedEpisode)}</em>}</div>
        </header>

        <div className="mp-layout">
          <main className="mp-main">
            <div className="mp-stage">
              <video ref={video} controls playsInline preload="metadata" />
              {playerState === 'loading' && <div className="mp-overlay"><Spin size="large" /><div className="mp-overlay-desc">正在加载播放内容…</div></div>}
              {playerState === 'error' && (
                <div className="mp-overlay">
                  <div className="mp-overlay-title">暂时无法播放</div>
                  <div className="mp-overlay-desc">{playerError}</div>
                  <Space wrap>
                    <Button type="primary" icon={<ReloadOutlined />} onClick={() => setReloadKey((key) => key + 1)}>重试</Button>
                    {nextLine && <Button icon={<SwapOutlined />} onClick={() => switchLine(nextLine)}>换一条线路</Button>}
                    <Button icon={<ExportOutlined />} onClick={openOriginal}>打开原始链接</Button>
                  </Space>
                </div>
              )}
              {!selectedEpisode && <div className="mp-overlay"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有可用剧集" /></div>}
            </div>

            <div className="mp-controls">
              <div className="mp-now">
                <span className="mp-now-label">正在播放</span>
                <b>{selectedEpisode ? episodeTitle(selectedEpisode) : '—'}</b>
                {selectedLine && <span className="mp-now-line">{selectedLine.sourceName} · {selectedLine.lineName}</span>}
              </div>
              <Space wrap size={8}>
                <Button icon={<StepBackwardOutlined />} disabled={!hasPrev} onClick={() => moveEpisode(-1)}>上一集</Button>
                <Button icon={<StepForwardOutlined />} disabled={!hasNext} onClick={() => moveEpisode(1)}>下一集</Button>
                <span className="mp-autonext"><Switch id="mp-autonext" size="small" checked={autoNext} onChange={toggleAutoNext} /><label htmlFor="mp-autonext">自动连播</label></span>
                <Tooltip title="复制播放地址"><Button type="text" icon={<CopyOutlined />} disabled={!selectedEpisode} onClick={() => selectedEpisode && copyText(selectedEpisode.streamUrl, '播放地址已复制')} /></Tooltip>
                <Tooltip title="打开原始链接"><Button type="text" icon={<ExportOutlined />} disabled={!selectedEpisode} onClick={openOriginal} /></Tooltip>
              </Space>
            </div>

            <section className="mp-info">
              <Cover url={item.coverUrl} title={item.title} className="mp-cover" />
              <div>
                <h1>{item.title}</h1>
                {item.originalTitle && item.originalTitle !== item.title && <div className="mp-original">{item.originalTitle}</div>}
                <div><TypeTags value={item.mediaType} /><Facts item={item} /></div>
                <GenreTags value={item.genres} />
                <People item={item} />
                <Typography.Paragraph className="mp-description" ellipsis={description ? { rows: 3, expandable: 'collapsible', symbol: (expanded) => (expanded ? '收起' : '展开') } : false}>
                  {description || '暂无简介。'}
                </Typography.Paragraph>
              </div>
            </section>
          </main>

          <aside className="mp-sidebar">
            <div className="mp-sidebar-head">
              <b>选集</b>
              {selectedLine && <span className="mp-count">共 {selectedLine.episodes.length} 集</span>}
              {(selectedLine?.episodes.length || 0) > 1 && (
                <Button size="small" type="text" icon={reverse ? <SortDescendingOutlined /> : <SortAscendingOutlined />} onClick={() => setReverse(!reverse)}>{reverse ? '倒序' : '正序'}</Button>
              )}
            </div>
            {lines.length > 1 && (
              <Select
                className="mp-line-select"
                value={selectedLine?.id}
                onChange={(id) => switchLine(lines.find((line) => line.id === id))}
                options={lines.map((line) => ({ value: line.id, label: `${line.sourceName} · ${line.lineName}（${line.episodes.length} 集）` }))}
              />
            )}
            <EpisodeRanges total={ordered.length} size={RANGE_SIZE} reverse={reverse} value={range} onChange={setRange} />
            <div className="mp-episode-grid" ref={grid}>
              {visible.map((episode) => {
                const active = episode.id === selectedEpisode?.id;
                return (
                  <button
                    key={episode.id}
                    type="button"
                    title={episodeTitle(episode)}
                    aria-pressed={active}
                    className={'mp-episode' + (active ? ' is-active' : '')}
                    onClick={() => selectedLine && selectEpisode(selectedLine, episode)}
                  >
                    {active && <span className={'mp-eq' + (playing ? ' is-playing' : '')}><i /><i /><i /></span>}
                    <span>{episodeTitle(episode)}</span>
                  </button>
                );
              })}
              {!visible.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有可用剧集" />}
            </div>
            <p className="mp-note">播放内容来自资源站，仅实时转发，不保存视频文件。观看进度保存在当前浏览器。</p>
          </aside>
        </div>
      </div>
    </ConfigProvider>
  );
}
