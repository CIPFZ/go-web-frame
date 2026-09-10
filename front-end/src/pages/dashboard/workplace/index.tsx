import React, { useMemo, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Link, useModel, useRequest } from '@umijs/max';
import {
  Alert,
  Avatar,
  Button,
  Modal,
  Pagination,
  Skeleton,
  Space,
  Tag,
  message,
} from 'antd';
import {
  ApiOutlined,
  AppstoreOutlined,
  ArrowRightOutlined,
  BellOutlined,
  CheckCircleFilled,
  CloudServerOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
  WarningFilled,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { getComponent } from '@/routing/componentMap';
import { getServerState } from '@/services/system/state';
import { markNoticeRead, type MyNotice } from '@/services/system/notice';
import { noticePageSize, useNotices } from './useNotices';
import { useStyles } from './styles';

interface WorkplaceMenu {
  path: string;
  name?: string;
  component?: string;
  icon?: React.ReactNode;
  hideInMenu?: boolean;
  routes?: WorkplaceMenu[];
}

function pageMenus(
  menus: WorkplaceMenu[],
  includeHidden = false,
): WorkplaceMenu[] {
  return menus.flatMap((menu) => {
    if (
      (!includeHidden && menu.hideInMenu) ||
      !menu.path?.startsWith('/') ||
      menu.path.startsWith('//')
    )
      return [];
    if (menu.routes?.length) return pageMenus(menu.routes, includeHidden);
    return menu.component &&
      getComponent(menu.component) &&
      menu.component !== 'components/RouterLayout'
      ? [menu]
      : [];
  });
}

const descriptions: Record<string, string> = {
  'sys/user': '维护账号与成员信息',
  'sys/authority': '分配角色与访问权限',
  'sys/menu': '配置导航与页面入口',
  'sys/api': '维护接口与权限资源',
  'sys/api-token': '管理程序访问凭证',
  'sys/notice': '发布公告与定向通知',
  'sys/operation': '追溯操作与变更记录',
  state: '查看服务与运行情况',
  about: '了解平台与技术信息',
};
const services = [
  { key: 'backend', name: '后端服务', icon: <CloudServerOutlined /> },
  { key: 'database', name: '数据库', icon: <DatabaseOutlined /> },
  { key: 'redis', name: 'Redis', icon: <DeploymentUnitOutlined /> },
  { key: 'mongo', name: 'MongoDB', icon: <ApiOutlined /> },
];
const levels = {
  info: { label: '通知', color: 'blue' },
  warning: { label: '提醒', color: 'orange' },
  error: { label: '重要', color: 'red' },
};

export default function Workplace() {
  const { styles, theme } = useStyles();
  const { initialState } = useModel('@@initialState');
  const user = initialState?.currentUser;
  const menus = initialState?.menuData;
  const shortcuts = useMemo(
    () =>
      pageMenus(menus || []).filter(
        (menu) => menu.component !== 'dashboard/workplace',
      ),
    [menus],
  );
  const allPages = useMemo(() => pageMenus(menus || [], true), [menus]);
  const statePage = shortcuts.find((menu) => menu.component === 'state');
  const accountPage = allPages.find((menu) => menu.component === 'user/info');
  const serverQuery = useRequest(
    async () => {
      const result = await getServerState();
      if (result.code !== 0) throw new Error(result.msg || '状态加载失败');
      return { data: { ...result.data, checkedAt: new Date() } };
    },
    { ready: !!statePage, pollingInterval: 15000, pollingWhenHidden: false },
  );
  const server = serverQuery.data?.server;
  const serverError = !!serverQuery.error;
  const unavailable =
    server && Object.values(server.checks).includes('unavailable');
  const healthLabel = serverError
    ? '状态待确认'
    : !server
      ? '正在检查'
      : unavailable
        ? '部分服务异常'
        : '服务运行正常';
  const healthColor =
    serverError || unavailable
      ? theme.colorWarning
      : !server
        ? theme.colorTextSecondary
        : theme.colorSuccess;
  const notices = useNotices();
  const [selected, setSelected] = useState<MyNotice>();
  const [marking, setMarking] = useState(false);
  const markingRef = useRef(false);
  const [messageApi, messageContext] = message.useMessage();
  const now = new Date();
  const hour = now.getHours();
  const greeting =
    hour < 6
      ? '夜深了'
      : hour < 12
        ? '上午好'
        : hour < 18
          ? '下午好'
          : '晚上好';
  const nickname = user?.nickName || user?.username || '用户';
  const role = user?.authority?.authorityName || '未分配角色';
  const unread =
    notices.data?.list.filter((notice) => !notice.readAt).length || 0;

  const confirmRead = async () => {
    if (!selected || markingRef.current) return;
    markingRef.current = true;
    setMarking(true);
    try {
      const result = await markNoticeRead({ noticeId: selected.ID });
      if (result.code !== 0) throw new Error(result.msg || '操作失败');
      messageApi.success('已确认已读');
      setSelected(undefined);
      notices.refresh();
    } catch {
      messageApi.error('标记已读失败，请重试');
    } finally {
      markingRef.current = false;
      setMarking(false);
    }
  };

  return (
    <PageContainer title="工作台" subTitle="从这里开始今天的工作">
      {messageContext}
      <div className={styles.stack}>
        <section className={styles.hero} aria-label="工作概览">
          <div className={styles.welcome}>
            <div>
              <div className={styles.eyebrow}>工作概览 · BASE FRAME</div>
              <h1 className={styles.greeting}>
                {greeting}，{nickname}
              </h1>
              <div className={styles.muted}>
                常用功能、个人通知与平台状态，尽在此处。
              </div>
            </div>
            <time
              className={styles.calendar}
              dateTime={dayjs(now).format('YYYY-MM-DD')}
            >
              <strong>{String(now.getDate()).padStart(2, '0')}</strong>
              <span>
                {now.getFullYear()} 年 {now.getMonth() + 1} 月<br />
                {now.toLocaleDateString('zh-CN', { weekday: 'long' })}
              </span>
            </time>
          </div>
          <div className={styles.summary}>
            <div className={styles.metric}>
              <div className={styles.icon}>
                <AppstoreOutlined />
              </div>
              <div>
                <div className={styles.muted}>可用入口</div>
                <strong>{shortcuts.length}</strong>
                <span className={styles.muted}> 项功能</span>
              </div>
            </div>
            <div className={styles.metric}>
              <div
                className={styles.icon}
                style={{
                  color: theme.colorWarning,
                  background: theme.colorWarningBg,
                }}
              >
                <BellOutlined />
              </div>
              <div>
                <div className={styles.muted}>个人通知</div>
                <strong>
                  {notices.loading || notices.error
                    ? '—'
                    : (notices.data?.total ?? 0)}
                </strong>
                <span className={styles.muted}> 条通知</span>
              </div>
            </div>
            <div className={styles.metric}>
              <div
                className={styles.icon}
                style={{
                  color: theme.colorSuccess,
                  background: theme.colorSuccessBg,
                }}
              >
                <SafetyCertificateOutlined />
              </div>
              <div>
                <div className={styles.muted}>当前角色</div>
                <strong style={{ fontSize: 17 }}>{role}</strong>
              </div>
            </div>
          </div>
        </section>
        <div className={styles.columns}>
          <div className={styles.stack}>
            <section className={styles.panel} aria-label="快捷入口">
              <div className={styles.sectionTitle}>
                <h2>快捷入口</h2>
                <span className={styles.muted}>常用功能，一步直达</span>
              </div>
              {shortcuts.length ? (
                <nav className={styles.links} aria-label="工作台快捷入口">
                  {shortcuts.map((menu) => (
                    <Link
                      className={styles.quickLink}
                      to={menu.path}
                      key={menu.path}
                      aria-label={menu.name}
                    >
                      <div className={styles.linkTop}>
                        <span className={styles.linkIcon}>
                          {menu.icon || <AppstoreOutlined />}
                        </span>
                        <ArrowRightOutlined />
                      </div>
                      <div>
                        <strong>{menu.name}</strong>
                        <br />
                        <small>
                          {descriptions[menu.component || ''] || '进入功能页面'}
                        </small>
                      </div>
                    </Link>
                  ))}
                </nav>
              ) : (
                <div className={styles.empty}>
                  <AppstoreOutlined className={styles.emptyIcon} />
                  <h3>暂无可用入口</h3>
                  <span className={styles.muted}>
                    可联系管理员分配所需功能。
                  </span>
                </div>
              )}
            </section>
            <section className={styles.panel} aria-label="我的通知">
              <div className={styles.sectionTitle}>
                <h2>我的通知</h2>
                <Button
                  type="text"
                  size="small"
                  icon={<ReloadOutlined />}
                  loading={notices.loading}
                  onClick={notices.refresh}
                >
                  刷新通知
                </Button>
              </div>
              {notices.error ? (
                <Alert
                  type="warning"
                  showIcon
                  message="通知加载失败"
                  description="暂时无法获取通知，请稍后重试。"
                  action={
                    <Button size="small" onClick={notices.refresh}>
                      重试
                    </Button>
                  }
                />
              ) : notices.loading ? (
                <Skeleton active paragraph={{ rows: 4 }} />
              ) : (
                <>
                  <div className={styles.noticeMeta}>
                    <span>总数 {notices.data?.total || 0}</span>
                    <span>本页 {unread} 条未读 · 未读优先</span>
                  </div>
                  {!notices.data?.list.length ? (
                    <div className={styles.empty}>
                      <div className={styles.emptyIcon}>
                        <BellOutlined />
                      </div>
                      <h3>暂时没有通知</h3>
                      <span className={styles.muted}>
                        发送给你的公告与提醒会显示在这里。
                      </span>
                    </div>
                  ) : (
                    notices.data.list.map((notice) => {
                      const level = levels[notice.level] || levels.info;
                      return (
                        <article
                          key={notice.ID}
                          className={styles.notice}
                          aria-label={notice.title}
                        >
                          <div className={styles.noticeHeading}>
                            {!notice.readAt && (
                              <span
                                className={styles.unreadDot}
                                aria-hidden="true"
                              />
                            )}
                            <Tag
                              color={level.color}
                              bordered={false}
                              style={{ margin: 0 }}
                            >
                              {level.label}
                            </Tag>
                            <h3>{notice.title}</h3>
                            {notice.needConfirm && !notice.readAt && (
                              <Tag bordered={false}>待确认</Tag>
                            )}
                          </div>
                          <p className={styles.noticePreview}>
                            {notice.content}
                          </p>
                          <div className={styles.noticeFooter}>
                            <span>
                              {dayjs(notice.createdAt).format(
                                'YYYY-MM-DD HH:mm',
                              )}{' '}
                              · {notice.readAt ? '已读' : '未读'}
                            </span>
                            <Button
                              type="link"
                              size="small"
                              onClick={() => setSelected(notice)}
                            >
                              查看详情 <ArrowRightOutlined />
                            </Button>
                          </div>
                        </article>
                      );
                    })
                  )}
                  {(notices.data?.total || 0) > noticePageSize && (
                    <div className={styles.pagination}>
                      <Pagination
                        size="small"
                        current={notices.page}
                        total={notices.data?.total}
                        pageSize={noticePageSize}
                        showSizeChanger={false}
                        onChange={notices.changePage}
                      />
                    </div>
                  )}
                </>
              )}
            </section>
          </div>
          <aside className={styles.stack}>
            {statePage && (
              <section className={styles.panel} aria-label="运行概览">
                <div className={styles.sectionTitle}>
                  <h2>运行概览</h2>
                  <Button
                    type="text"
                    size="small"
                    aria-label="刷新运行概览"
                    icon={<ReloadOutlined spin={serverQuery.loading} />}
                    loading={serverQuery.loading}
                    onClick={serverQuery.refresh}
                  />
                </div>
                <div
                  className={styles.health}
                  role="status"
                  style={{
                    color: healthColor,
                    background:
                      serverError || unavailable
                        ? theme.colorWarningBg
                        : !server
                          ? theme.colorFillAlter
                          : theme.colorSuccessBg,
                  }}
                >
                  {serverError || unavailable ? (
                    <WarningFilled />
                  ) : !server ? (
                    <ClockCircleOutlined />
                  ) : (
                    <CheckCircleFilled />
                  )}
                  {healthLabel}
                </div>
                {serverError ? (
                  <div className={styles.muted} style={{ marginTop: 14 }}>
                    状态更新失败，请刷新重试。
                  </div>
                ) : !server ? (
                  <Skeleton active title={false} paragraph={{ rows: 4 }} />
                ) : (
                  services.map((service) => {
                    const status = server.checks[service.key];
                    return (
                      <div key={service.key} className={styles.service}>
                        <span className={styles.serviceName}>
                          {service.icon}
                          {service.name}
                        </span>
                        <Tag
                          bordered={false}
                          color={
                            status === 'ok'
                              ? 'success'
                              : status === 'disabled'
                                ? 'default'
                                : 'warning'
                          }
                        >
                          {status === 'ok'
                            ? '正常'
                            : status === 'disabled'
                              ? '未启用'
                              : '不可用'}
                        </Tag>
                      </div>
                    );
                  })
                )}
                <div className={styles.panelFooter}>
                  <span>
                    {serverQuery.data?.checkedAt
                      ? `${dayjs(serverQuery.data.checkedAt).format('HH:mm:ss')} 更新`
                      : '每 15 秒自动检查'}
                  </span>
                  <Link to={statePage.path}>
                    查看完整状态 <ArrowRightOutlined />
                  </Link>
                </div>
              </section>
            )}
            <section className={styles.panel} aria-label="当前账号">
              <div className={styles.sectionTitle}>
                <h2>当前账号</h2>
                <Tag bordered={false} color="blue" style={{ margin: 0 }}>
                  已登录
                </Tag>
              </div>
              <div className={styles.account}>
                <Avatar
                  size={44}
                  src={user?.avatar}
                  icon={<UserOutlined />}
                  style={{
                    flexShrink: 0,
                    background: theme.colorPrimaryBg,
                    color: theme.colorPrimary,
                  }}
                />
                <div style={{ minWidth: 0 }}>
                  <strong>{nickname}</strong>
                  <span className={styles.muted}>@{user?.username}</span>
                </div>
              </div>
              <dl className={styles.accountFacts}>
                <div>
                  <dt>当前角色</dt>
                  <dd>{role}</dd>
                </div>
                <div>
                  <dt>联系邮箱</dt>
                  <dd>{user?.email || '未设置'}</dd>
                </div>
              </dl>
              {accountPage && (
                <div className={styles.panelFooter}>
                  <span>维护个人资料与账号安全</span>
                  <Link to={accountPage.path}>
                    账号设置 <ArrowRightOutlined />
                  </Link>
                </div>
              )}
            </section>
          </aside>
        </div>
      </div>
      <Modal
        open={!!selected}
        title={selected?.title}
        onCancel={() => {
          if (!marking) setSelected(undefined);
        }}
        maskClosable={!marking}
        closable={!marking}
        keyboard={!marking}
        footer={
          <Space>
            <Button disabled={marking} onClick={() => setSelected(undefined)}>
              关闭
            </Button>
            {selected && !selected.readAt && (
              <Button
                type="primary"
                loading={marking}
                onClick={() => void confirmRead()}
              >
                {selected.needConfirm ? '确认已读' : '标记已读'}
              </Button>
            )}
          </Space>
        }
      >
        {selected && (
          <>
            <Space wrap>
              <Tag color={(levels[selected.level] || levels.info).color}>
                {(levels[selected.level] || levels.info).label}
              </Tag>
              <span className={styles.muted}>
                {dayjs(selected.createdAt).format('YYYY-MM-DD HH:mm')}
              </span>
              {selected.readAt && <Tag>已读</Tag>}
            </Space>
            <div className={styles.noticeBody}>{selected.content}</div>
            {selected.needConfirm && !selected.readAt && (
              <Alert type="info" showIcon message="这条通知需要你确认已读。" />
            )}
          </>
        )}
      </Modal>
    </PageContainer>
  );
}
