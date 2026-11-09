import { t } from '@/i18n';
import {
  ExclamationCircleOutlined,
  LoginOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { Button, Skeleton } from 'antd';
import { createStyles } from 'antd-style';
import React, { useCallback, useEffect, useSyncExternalStore } from 'react';
import type { WidgetSize } from '../registry';

// 小组件的数据：同一份数据只查一次（同时放了几个相同的小组件也一样），60 秒后过期，页面在前台时自动刷新。

type Entry = {
  data?: unknown;
  error?: unknown;
  at: number;
  loading: boolean;
  subs: Set<() => void>;
  snapshot: Snapshot<unknown>;
};
type Snapshot<T> = { data?: T; error?: unknown; loading: boolean };
const store = new Map<string, Entry>();
const TTL = 60_000;
const IDLE: Snapshot<unknown> = { loading: false };

function entry(key: string): Entry {
  let e = store.get(key);
  if (!e) {
    e = { at: 0, loading: false, subs: new Set(), snapshot: { loading: true } };
    store.set(key, e);
  }
  return e;
}

function publish(e: Entry) {
  e.snapshot = { data: e.data, error: e.error, loading: e.loading };
  for (const fn of e.subs) fn();
}

async function fetchEntry(
  key: string,
  load: () => Promise<unknown>,
  force = false,
) {
  const e = entry(key);
  if (e.loading || (!force && Date.now() - e.at < TTL)) return;
  e.loading = true;
  publish(e);
  try {
    e.data = await load();
    e.error = undefined;
  } catch (err) {
    e.error = err;
  } finally {
    e.at = Date.now();
    e.loading = false;
    publish(e);
  }
}

/** 共享、自动刷新的数据。key 为空时不查询。 */
export function useShared<T>(key: string | null, load: () => Promise<T>) {
  const subscribe = useCallback(
    (fn: () => void) => {
      if (!key) return () => undefined;
      const e = entry(key);
      e.subs.add(fn);
      return () => e.subs.delete(fn);
    },
    [key],
  );
  // 快照必须是同一个对象（内容变了才换），否则 React 会不停重绘
  const snapshot = useSyncExternalStore(subscribe, () =>
    key ? (entry(key).snapshot as Snapshot<T>) : (IDLE as Snapshot<T>),
  );
  useEffect(() => {
    if (!key) return;
    const tick = () => !document.hidden && void fetchEntry(key, load);
    tick();
    const timer = window.setInterval(tick, TTL / 2);
    document.addEventListener('visibilitychange', tick);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener('visibilitychange', tick);
    };
    // load 随 key 一起变化
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key]);
  return { ...snapshot, reload: () => key && void fetchEntry(key, load, true) };
}

/** 查询失败时显示的一句话：Error 用它的 message，其余显示通用的“请求失败”。 */
const errorText = (error: unknown) =>
  error instanceof Error && error.message
    ? error.message
    : t('cms.requestError');

/** 小组件的状态：正常 / 告警 / 需要登录 / 未配置 / 查询失败。除正常与告警外显示一句说明与对应操作。 */
export type WidgetStatus =
  | { kind: 'ok' | 'alert' }
  | { kind: 'loading' }
  | { kind: 'login'; text: string; href: string }
  | { kind: 'unconfigured' }
  | { kind: 'error'; error: unknown; retry: () => void };

export const useWidgetStyles = createStyles(({ token }) => ({
  body: {
    flex: 1,
    minHeight: 0,
    display: 'flex',
    flexDirection: 'column',
    gap: 8,
  },
  big: {
    fontSize: 30,
    fontWeight: 600,
    lineHeight: 1.15,
    fontVariantNumeric: 'tabular-nums',
    letterSpacing: -0.5,
    '& small': {
      fontSize: 13,
      fontWeight: 400,
      color: token.colorTextTertiary,
      marginLeft: 4,
      letterSpacing: 0,
    },
  },
  label: { color: token.colorTextSecondary, fontSize: 12 },
  muted: { color: token.colorTextTertiary, fontSize: 12 },
  alert: { color: token.colorError },
  bars: { display: 'grid', gap: 10 },
  bar: {
    display: 'grid',
    gridTemplateColumns: '64px minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: 10,
    fontSize: 12,
    '& > span:first-child': { color: token.colorTextSecondary },
    '& > span:last-child': {
      fontVariantNumeric: 'tabular-nums',
      color: token.colorText,
      textAlign: 'right',
      minWidth: 96,
    },
  },
  stats: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(72px, 1fr))',
    gap: 8,
  },
  stat: {
    padding: '8px 10px',
    borderRadius: 12,
    background: token.colorFillQuaternary,
    '& strong': {
      display: 'block',
      fontSize: 20,
      fontWeight: 600,
      fontVariantNumeric: 'tabular-nums',
      lineHeight: 1.3,
    },
    '& span': { color: token.colorTextSecondary, fontSize: 12 },
  },
  list: {
    display: 'grid',
    gap: 6,
    fontSize: 13,
    minHeight: 0,
    overflow: 'hidden',
  },
  listRow: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: 12,
    padding: '6px 0',
    borderBottom: `1px solid ${token.colorBorderSecondary}`,
    '&:last-child': { borderBottom: 0 },
    '& > span:first-child': {
      minWidth: 0,
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
    },
    '& > span:last-child': {
      color: token.colorTextTertiary,
      flex: 'none',
      fontVariantNumeric: 'tabular-nums',
    },
  },
  center: {
    flex: 1,
    display: 'grid',
    placeItems: 'center',
    textAlign: 'center',
  },
  state: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    textAlign: 'center',
    color: token.colorTextSecondary,
    fontSize: 12,
    '& .anticon': { fontSize: 20, color: token.colorTextQuaternary },
  },
  actions: { display: 'flex', gap: 8, marginTop: 'auto', flexWrap: 'wrap' },
}));

/** 非正常状态的内容；正常返回 null。 */
export function StatusView({
  status,
  size,
}: {
  status: WidgetStatus;
  size: WidgetSize;
}) {
  const { styles } = useWidgetStyles();
  const stop = (e: React.SyntheticEvent) => e.stopPropagation();
  if (status.kind === 'loading')
    return (
      <Skeleton
        active
        title={false}
        paragraph={{ rows: size === 's' ? 2 : 3 }}
      />
    );
  if (status.kind === 'login')
    return (
      <div className={styles.state}>
        <LoginOutlined />
        {size !== 's' ? <span>{status.text}</span> : null}
        <Button
          size="small"
          href={status.href}
          target="_blank"
          rel="noopener noreferrer"
          onClick={stop}
        >
          {t('cms.widget.signIn')} ↗
        </Button>
      </div>
    );
  if (status.kind === 'unconfigured')
    return (
      <div className={styles.state}>
        <ExclamationCircleOutlined />
        <span>{t('cms.widget.notConfigured')}</span>
      </div>
    );
  if (status.kind === 'error')
    return (
      <div className={styles.state}>
        <ExclamationCircleOutlined />
        {size !== 's' ? <span>{errorText(status.error)}</span> : null}
        <Button
          size="small"
          icon={<ReloadOutlined />}
          onClick={(e) => {
            stop(e);
            status.retry();
          }}
        >
          {t('cms.retry')}
        </Button>
      </div>
    );
  return null;
}

/** 一行进度：名称、进度条（超出变红）、已用 / 上限。 */
export function Bar({
  label,
  used,
  limit,
  unit,
  color,
}: {
  label: string;
  used?: number | null;
  limit?: number | null;
  unit?: string;
  color: string;
}) {
  const { styles, theme } = useWidgetStyles();
  const ratio = used != null && limit ? used / limit : 0;
  return (
    <div className={styles.bar}>
      <span>{label}</span>
      <div
        style={{
          height: 6,
          borderRadius: 3,
          background: theme.colorFillSecondary,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: `${Math.min(100, ratio * 100)}%`,
            height: '100%',
            borderRadius: 3,
            background: ratio > 1 ? theme.colorError : color,
          }}
        />
      </div>
      <span>
        {fmt(used)} / {fmt(limit)}
        {unit ? ` ${unit}` : ''}
      </span>
    </div>
  );
}

/** 数字：千分位、最多一位小数；一百万以上才缩写（“120万 / 1.2M”），免得同一行里写法不一。 */
export function fmt(v?: number | null): string {
  if (v == null || !Number.isFinite(v)) return '—';
  return new Intl.NumberFormat(undefined, {
    notation: Math.abs(v) >= 1_000_000 ? 'compact' : 'standard',
    maximumFractionDigits: 1,
  }).format(v);
}
