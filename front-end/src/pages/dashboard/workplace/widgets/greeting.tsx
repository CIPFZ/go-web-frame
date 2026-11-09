import { formatDate, t, useI18n } from '@/i18n';
import {
  AppstoreOutlined,
  BellOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { history } from '@umijs/max';
import { createStyles } from 'antd-style';
import dayjs from 'dayjs';
import React from 'react';
import type { WidgetSize } from '../registry';
import { useStyles } from '../styles';
import { useDesk } from './context';

// 问候：小号像日历，中号是问候与日期，宽号是工作概览（问候、日期与三项：可用入口、未读通知、我的角色）。

const useGreetingStyles = createStyles(({ token }) => ({
  card: {
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden',
    borderRadius: 20,
    border: `1px solid ${token.colorBorderSecondary}`,
    background: `linear-gradient(115deg, ${token.colorBgContainer} 25%, ${token.colorPrimaryBg} 100%)`,
    boxShadow:
      '0 1px 2px rgba(15, 23, 42, 0.04), 0 4px 16px rgba(15, 23, 42, 0.04)',
  },
  calendar: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    padding: '14px 18px',
    '& small': { color: token.colorError, fontSize: 13, fontWeight: 600 },
    '& strong': {
      fontSize: 54,
      fontWeight: 300,
      lineHeight: 1.05,
      color: token.colorText,
      fontVariantNumeric: 'tabular-nums',
    },
    '& span': { color: token.colorTextSecondary, fontSize: 12 },
  },
  medium: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 16,
    padding: '16px 22px',
    '& h2': {
      fontSize: 22,
      fontWeight: 600,
      margin: '0 0 6px',
      lineHeight: 1.35,
      overflowWrap: 'anywhere',
    },
  },
  metricButton: {
    border: 0,
    background: 'transparent',
    font: 'inherit',
    textAlign: 'left',
    cursor: 'pointer',
    width: '100%',
    '&:hover strong': { color: token.colorPrimary },
    '&:focus-visible': {
      outline: `2px solid ${token.colorPrimary}`,
      outlineOffset: -2,
    },
  },
}));

function useGreetingText() {
  const { user } = useDesk();
  const hour = new Date().getHours();
  const greeting =
    hour < 6
      ? t('cms.goodEvening')
      : hour < 12
        ? t('cms.goodMorning')
        : hour < 18
          ? t('cms.goodAfternoon')
          : t('cms.goodEvening.509f35');
  return t('cms.welcomeUser', {
    greeting,
    name: user?.nickName || user?.username || t('cms.user'),
  });
}

export function GreetingWidget({ size }: { size: WidgetSize }) {
  const locale = useI18n();
  const { styles } = useGreetingStyles();
  const hero = useStyles().styles;
  const { theme } = useStyles();
  const desk = useDesk();
  const now = new Date();
  const text = useGreetingText();
  const weekday = now.toLocaleDateString(locale, { weekday: 'long' });
  const stop = (e: React.SyntheticEvent) => e.stopPropagation();

  if (size === 's')
    return (
      <div className={styles.card}>
        <time
          className={styles.calendar}
          dateTime={dayjs(now).format('YYYY-MM-DD')}
        >
          <small>{weekday}</small>
          <strong>{now.getDate()}</strong>
          <span>{formatDate(now, { year: 'numeric', month: 'long' })}</span>
        </time>
      </div>
    );
  if (size === 'm')
    return (
      <div className={styles.card}>
        <div className={styles.medium}>
          <div style={{ minWidth: 0 }}>
            <h2>{text}</h2>
            <div className={hero.muted}>
              {formatDate(now, {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                weekday: 'long',
              })}
            </div>
          </div>
        </div>
      </div>
    );

  const unread = (desk.notices.data?.list || []).filter((n) => !n.readAt);
  const metrics = [
    {
      key: 'tools',
      icon: <AppstoreOutlined />,
      color: undefined,
      label: t('cms.availableTools'),
      value: (
        <>
          <strong>{desk.pagesCount}</strong>
          <span className={hero.muted}>
            {' '}
            {t('cms.toolsUnit', { count: desk.pagesCount })}
          </span>
        </>
      ),
      onClick: () => desk.openGallery('apps'),
    },
    {
      key: 'notices',
      icon: <BellOutlined />,
      color: [theme.colorWarning, theme.colorWarningBg],
      label: t('cms.personalNotices'),
      value: (
        <>
          <strong>
            {desk.notices.loading || desk.notices.error
              ? '—'
              : (desk.notices.data?.total ?? 0)}
          </strong>
          <span className={hero.muted}>
            {' '}
            {t('cms.noticesUnit', { count: desk.notices.data?.total || 0 })}
          </span>
        </>
      ),
      onClick: () => {
        const first = unread[0] || desk.notices.data?.list[0];
        if (first) desk.openNotice(first);
      },
    },
    {
      key: 'role',
      icon: <SafetyCertificateOutlined />,
      color: [theme.colorSuccess, theme.colorSuccessBg],
      label: t('cms.currentRole'),
      value: <strong style={{ fontSize: 17 }}>{desk.role}</strong>,
      onClick: () => desk.accountPath && history.push(desk.accountPath),
    },
  ];
  return (
    <section
      className={hero.hero}
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        borderRadius: 20,
      }}
      aria-label={t('cms.workOverview')}
    >
      <div className={hero.welcome} style={{ flex: 1 }}>
        <div>
          <div className={hero.eyebrow}>{t('cms.workOverview')}</div>
          <h1 className={hero.greeting}>{text}</h1>
          <div className={hero.muted}>
            {t('cms.yourToolsNoticesAndPlatformStatus')}
          </div>
        </div>
        <time
          className={hero.calendar}
          dateTime={dayjs(now).format('YYYY-MM-DD')}
        >
          <strong>{String(now.getDate()).padStart(2, '0')}</strong>
          <span>
            {formatDate(now, { year: 'numeric', month: 'long' })}
            <br />
            {weekday}
          </span>
        </time>
      </div>
      <div className={hero.summary}>
        {metrics.map((m) => (
          <button
            key={m.key}
            type="button"
            className={`${hero.metric} ${styles.metricButton}`}
            onClick={(e) => {
              stop(e);
              m.onClick();
            }}
          >
            <div
              className={hero.icon}
              style={
                m.color
                  ? { color: m.color[0], background: m.color[1] }
                  : undefined
              }
            >
              {m.icon}
            </div>
            <div>
              <div className={hero.muted}>{m.label}</div>
              {m.value}
            </div>
          </button>
        ))}
      </div>
    </section>
  );
}
