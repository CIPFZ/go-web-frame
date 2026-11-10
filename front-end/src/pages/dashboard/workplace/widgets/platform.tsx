import { formatDate, t } from '@/i18n';
import {
  ApiOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  UserOutlined,
  WarningFilled,
} from '@ant-design/icons';
import { Avatar, Pagination, Skeleton, Tag } from 'antd';
import React from 'react';
import type { WidgetSize } from '../registry';
import { noticePageSize } from '../useNotices';
import { useDesk } from './context';
import { StatusView, useWidgetStyles } from './shared';

const stop = (e: React.SyntheticEvent) => e.stopPropagation();

// ---------- 我的通知 ----------

export function NoticesWidget({ size }: { size: WidgetSize }) {
  const { styles, theme } = useWidgetStyles();
  const { notices, openNotice } = useDesk();
  const list = notices.data?.list || [];
  const unread = list.filter((n) => !n.readAt).length;
  if (notices.loading && !notices.data)
    return <Skeleton active title={false} paragraph={{ rows: 2 }} />;
  if (notices.error && !notices.data)
    return (
      <StatusView
        size={size}
        status={{
          kind: 'error',
          error: new Error(t('cms.unableToLoadNotices')),
          retry: notices.refresh,
        }}
      />
    );
  if (size === 's')
    return (
      <div className={styles.body}>
        <div
          className={styles.big}
          style={{ color: unread ? theme.colorPrimary : undefined }}
        >
          {unread}
        </div>
        <div className={styles.muted}>{t('cms.widget.unreadNotices')}</div>
      </div>
    );
  if (!list.length)
    return <div className={styles.state}>{t('cms.noNoticesYet')}</div>;
  const total = notices.data?.total || 0;
  return (
    <div className={styles.body}>
      <div className={styles.list}>
        {list.slice(0, size === 'm' ? 3 : noticePageSize).map((n) => (
          <button
            key={n.ID}
            type="button"
            className={styles.listRow}
            style={{
              border: 0,
              borderBottom: `1px solid ${theme.colorBorderSecondary}`,
              background: 'transparent',
              cursor: 'pointer',
              font: 'inherit',
              textAlign: 'left',
            }}
            onClick={(e) => {
              stop(e);
              openNotice(n);
            }}
          >
            <span
              style={{
                fontWeight: n.readAt ? 400 : 600,
                color: theme.colorText,
              }}
            >
              {!n.readAt ? (
                <span style={{ color: theme.colorPrimary }}>● </span>
              ) : null}
              {n.title}
            </span>
            <span>
              {formatDate(n.createdAt, { month: '2-digit', day: '2-digit' })}
            </span>
          </button>
        ))}
      </div>
      {/* 大号可以翻页看到全部通知（未读在前） */}
      {size === 'l' && total > noticePageSize ? (
        <div
          style={{
            marginTop: 'auto',
            display: 'flex',
            justifyContent: 'flex-end',
          }}
          onClick={stop}
          onKeyDown={stop}
        >
          <Pagination
            simple
            size="small"
            current={notices.page}
            pageSize={noticePageSize}
            total={total}
            onChange={notices.changePage}
          />
        </div>
      ) : null}
    </div>
  );
}

// ---------- 服务概况 ----------

const SERVICES = [
  {
    key: 'backend',
    icon: <CloudServerOutlined />,
    name: () => t('cms.backendService'),
  },
  {
    key: 'database',
    icon: <DatabaseOutlined />,
    name: () => t('cms.database'),
  },
  { key: 'redis', icon: <DeploymentUnitOutlined />, name: () => 'Redis' },
  { key: 'mongo', icon: <ApiOutlined />, name: () => 'MongoDB' },
];

export function ServiceWidget({ size }: { size: WidgetSize }) {
  const { styles, theme } = useWidgetStyles();
  const { server } = useDesk();
  const bad =
    server.error || Object.values(server.checks || {}).includes('unavailable');
  const label = server.error
    ? t('cms.statusUnconfirmed')
    : !server.checks
      ? t('cms.checkingStatus')
      : bad
        ? t('cms.someServicesAreUnavailable')
        : t('cms.servicesAreHealthy');
  const color = bad
    ? theme.colorWarning
    : !server.checks
      ? theme.colorTextTertiary
      : theme.colorSuccess;
  const icon = bad ? (
    <WarningFilled />
  ) : !server.checks ? (
    <ClockCircleOutlined />
  ) : (
    <CheckCircleFilled />
  );
  if (size === 's')
    return (
      <div className={styles.center}>
        <div>
          <div style={{ fontSize: 34, color }}>{icon}</div>
          <div className={styles.label} style={{ color }}>
            {label}
          </div>
        </div>
      </div>
    );
  if (server.error)
    return (
      <StatusView
        size={size}
        status={{
          kind: 'error',
          error: new Error(t('cms.statusUnconfirmed')),
          retry: server.refresh,
        }}
      />
    );
  return (
    <div className={styles.list}>
      {SERVICES.map((s) => {
        const status = server.checks?.[s.key];
        return (
          <div key={s.key} className={styles.listRow}>
            <span>
              {s.icon} {s.name()}
            </span>
            <Tag
              bordered={false}
              color={
                status === 'ok'
                  ? 'success'
                  : status === 'disabled'
                    ? 'default'
                    : status
                      ? 'warning'
                      : undefined
              }
              style={{ margin: 0 }}
            >
              {status === 'ok'
                ? t('cms.healthy')
                : status === 'disabled'
                  ? t('cms.notEnabled')
                  : status
                    ? t('cms.unavailable')
                    : '—'}
            </Tag>
          </div>
        );
      })}
    </div>
  );
}

// ---------- 当前账号 ----------

export function AccountWidget({ size }: { size: WidgetSize }) {
  const { styles, theme } = useWidgetStyles();
  const { user, role } = useDesk();
  const avatar = (
    <Avatar
      size={size === 's' ? 52 : 44}
      src={user?.avatar}
      icon={<UserOutlined />}
      style={{
        background: theme.colorPrimaryBg,
        color: theme.colorPrimary,
        flexShrink: 0,
      }}
    />
  );
  if (size === 's')
    return (
      <div className={styles.center}>
        <div>
          {avatar}
          <div style={{ fontWeight: 600, marginTop: 6 }}>
            {user?.nickName || user?.username}
          </div>
          <div className={styles.muted}>{role}</div>
        </div>
      </div>
    );
  return (
    <div className={styles.body}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        {avatar}
        <div style={{ minWidth: 0 }}>
          <div style={{ fontWeight: 600 }}>
            {user?.nickName || user?.username}
          </div>
          <div className={styles.muted}>@{user?.username}</div>
        </div>
      </div>
      <div className={styles.listRow}>
        <span className={styles.muted}>{t('cms.currentRole')}</span>
        <span>{role}</span>
      </div>
      <div className={styles.listRow}>
        <span className={styles.muted}>{t('cms.contactEmail')}</span>
        <span>{user?.email || t('cms.notSet')}</span>
      </div>
    </div>
  );
}
