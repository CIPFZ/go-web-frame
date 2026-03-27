import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer, ProCard } from '@ant-design/pro-components';
import { Link, useModel, useRequest } from '@umijs/max';
import { Alert, Button, Checkbox, List, Modal, Progress, Space, Tag, Typography, message } from 'antd';
import dayjs from 'dayjs';
import { getMyNotices, markNoticeRead } from '@/services/api/notice';
import { getPluginOverview } from '@/services/api/plugin';
import { getServerState } from '@/services/api/state';
import { updateUiConfig } from '@/services/api/user';

const { Text, Paragraph } = Typography;

const quickLinks = [
  { title: '服务器状态', path: '/state' },
  { title: '用户管理', path: '/sys/user' },
  { title: '角色管理', path: '/sys/authority' },
  { title: '菜单管理', path: '/sys/menu' },
  { title: 'API 管理', path: '/sys/api' },
  { title: '通知公告', path: '/sys/notice' },
  { title: '操作日志', path: '/sys/operation' },
  { title: '诗词管理', path: '/poetry/poem' },
];

const levelColor: Record<string, string> = {
  info: 'blue',
  warning: 'orange',
  error: 'red',
};

const workplaceModules = [
  { label: '快捷入口', value: 'quickLinks' },
  { label: '系统健康', value: 'systemHealth' },
  { label: '插件模块', value: 'pluginOverview' },
  { label: '我的通知', value: 'myNotices' },
];

const defaultVisibleModules = workplaceModules.map((item) => item.value);

const toPercent = (used = 0, total = 0) => {
  if (!total) return 0;
  return Number(((used / total) * 100).toFixed(2));
};

const Workplace: React.FC = () => {
  const { initialState, setInitialState } = useModel('@@initialState');
  const serverQuery = useRequest(() => getServerState(), { pollingInterval: 10000 });
  const pluginOverviewQuery = useRequest(() => getPluginOverview(), { pollingInterval: 15000 });

  const [noticeLoading, setNoticeLoading] = useState(false);
  const [noticeError, setNoticeError] = useState<string | null>(null);
  const [notices, setNotices] = useState<any[]>([]);
  const [noticeTotal, setNoticeTotal] = useState(0);
  const [configVisible, setConfigVisible] = useState(false);
  const [savingConfig, setSavingConfig] = useState(false);
  const [visibleModules, setVisibleModules] = useState<string[]>(defaultVisibleModules);

  const loadNotices = async () => {
    setNoticeLoading(true);
    setNoticeError(null);
    try {
      const res = await getMyNotices({ page: 1, pageSize: 8 });
      if (res.code !== 0) {
        setNoticeError(res.msg || '加载失败');
        setNotices([]);
        setNoticeTotal(0);
        return;
      }
      setNotices(res.data?.list || []);
      setNoticeTotal(res.data?.total || 0);
    } catch (e: any) {
      setNoticeError(e?.message || '加载失败');
      setNotices([]);
      setNoticeTotal(0);
    } finally {
      setNoticeLoading(false);
    }
  };

  useEffect(() => {
    loadNotices();
  }, []);

  useEffect(() => {
    const savedModules = initialState?.currentUser?.settings?.workplace?.visibleModules;
    if (Array.isArray(savedModules) && savedModules.length > 0) {
      setVisibleModules(savedModules);
      return;
    }
    setVisibleModules(defaultVisibleModules);
  }, [initialState?.currentUser?.settings]);

  const server = serverQuery.data?.data?.server;
  const unreadCount = notices.filter((n: any) => !n.readAt).length;

  const cpuAvg = useMemo(() => {
    const cpus: number[] = server?.cpu?.cpus || [];
    if (!cpus.length) return 0;
    return Number((cpus.reduce((sum, n) => sum + n, 0) / cpus.length).toFixed(2));
  }, [server]);

  const ramPercent = toPercent(server?.ram?.used || 0, server?.ram?.total || 0);
  const pluginOverview = pluginOverviewQuery.data?.data;
  const showModule = (moduleName: string) => visibleModules.includes(moduleName);

  const saveWorkplaceConfig = async () => {
    if (!visibleModules.length) {
      message.warning('至少保留一个工作台模块');
      return;
    }

    const currentSettings = initialState?.currentUser?.settings || {};
    const mergedSettings = {
      ...currentSettings,
      workplace: {
        ...(currentSettings.workplace || {}),
        visibleModules,
      },
    };

    setSavingConfig(true);
    try {
      const res = await updateUiConfig({ settings: mergedSettings });
      if (res.code !== 0) {
        message.error(res.msg || '保存失败');
        return;
      }
      message.success('工作台配置已保存');
      setConfigVisible(false);
      await setInitialState((prev) => ({
        ...prev,
        currentUser: prev?.currentUser
          ? {
              ...prev.currentUser,
              settings: mergedSettings,
            }
          : prev?.currentUser,
      }));
    } catch (error: any) {
      message.error(error?.message || '保存失败');
    } finally {
      setSavingConfig(false);
    }
  };

  return (
    <PageContainer
      title="工作台"
      subTitle="系统入口与个人待办"
      extra={[
        <Button key="config" onClick={() => setConfigVisible(true)}>
          配置工作台
        </Button>,
      ]}
    >
      <ProCard gutter={[16, 16]} wrap>
        {showModule('quickLinks') ? (
          <ProCard title="快捷入口" colSpan={{ xs: 24, lg: 14 }}>
            <Space size={[8, 8]} wrap>
              {quickLinks.map((item) => (
                <Link key={item.path} to={item.path}>
                  <Button>{item.title}</Button>
                </Link>
              ))}
            </Space>
          </ProCard>
        ) : null}

        {showModule('systemHealth') ? (
          <ProCard title="系统健康" colSpan={{ xs: 24, lg: 10 }} loading={serverQuery.loading}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text>CPU 平均占用: {cpuAvg}%</Text>
                <Progress percent={cpuAvg} size="small" />
              </div>
              <div>
                <Text>内存占用: {ramPercent}%</Text>
                <Progress percent={ramPercent} size="small" />
              </div>
              <div>
                <Text>负载 (1/5/15): </Text>
                <Tag>{Number(server?.cpu?.load1 || 0).toFixed(2)}</Tag>
                <Tag>{Number(server?.cpu?.load5 || 0).toFixed(2)}</Tag>
                <Tag>{Number(server?.cpu?.load15 || 0).toFixed(2)}</Tag>
              </div>
            </Space>
          </ProCard>
        ) : null}

        {showModule('pluginOverview') ? (
          <ProCard
            title="插件模块"
            colSpan={{ xs: 24, lg: 12 }}
            loading={pluginOverviewQuery.loading}
            extra={<Link to="/plugin/center">进入插件中心</Link>}
          >
            <Space wrap size={[12, 12]}>
              <Tag color="blue">项目 {pluginOverview?.projectCount || 0}</Tag>
              <Tag color="gold">准备中 {pluginOverview?.preparingCount || 0}</Tag>
              <Tag color="orange">审核中 {pluginOverview?.pendingReviewCount || 0}</Tag>
              <Tag color="cyan">待发布 {pluginOverview?.approvedCount || 0}</Tag>
              <Tag color="green">已发布 {pluginOverview?.publishedCount || 0}</Tag>
              <Tag color="red">已下架 {pluginOverview?.offlinedCount || 0}</Tag>
            </Space>
            <Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0 }}>
              插件统计作为工作台模块接入，当前支持按用户配置显示或隐藏。
            </Paragraph>
          </ProCard>
        ) : null}

        {showModule('myNotices') ? (
          <ProCard title="我的通知" colSpan={24} loading={noticeLoading}>
            <Space style={{ marginBottom: 12 }}>
              <Tag color="blue">总数 {noticeTotal}</Tag>
              <Tag color={unreadCount > 0 ? 'red' : 'default'}>未读 {unreadCount}</Tag>
              <Button size="small" onClick={loadNotices}>
                刷新
              </Button>
            </Space>
            {noticeError ? (
              <Alert
                style={{ marginBottom: 12 }}
                type="error"
                showIcon
                message="通知加载失败"
                description={noticeError}
              />
            ) : null}
            <List
              dataSource={notices}
              locale={{ emptyText: '暂无通知' }}
              renderItem={(item: any) => (
                <List.Item
                  actions={[
                    item.readAt ? (
                      <Tag color="default" key="done">
                        已读
                      </Tag>
                    ) : (
                      <Button
                        key="read"
                        type="link"
                        onClick={async () => {
                          const res = await markNoticeRead({ noticeId: item.ID });
                          if (res.code === 0) {
                            message.success('已标记已读');
                            loadNotices();
                          } else {
                            message.error(res.msg || '操作失败');
                          }
                        }}
                      >
                        标记已读
                      </Button>
                    ),
                  ]}
                >
                  <List.Item.Meta
                    title={
                      <Space>
                        <Tag color={levelColor[item.level] || 'default'}>{item.level}</Tag>
                        <Text strong>{item.title}</Text>
                        {item.needConfirm ? <Tag color="purple">需确认</Tag> : null}
                      </Space>
                    }
                    description={
                      <Space direction="vertical" size={2}>
                        <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>
                          {item.content}
                        </Paragraph>
                        <Text type="secondary">
                          创建于 {dayjs(item.createdAt).format('YYYY-MM-DD HH:mm:ss')}
                        </Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </ProCard>
        ) : null}
      </ProCard>
      <Modal
        title="配置工作台"
        open={configVisible}
        onOk={saveWorkplaceConfig}
        onCancel={() => setConfigVisible(false)}
        confirmLoading={savingConfig}
        okText="保存配置"
        cancelText="取消"
      >
        <Paragraph type="secondary">
          选择需要显示在工作台中的模块。后续我们可以继续补拖拽排序和更多模块接入。
        </Paragraph>
        <Checkbox.Group
          style={{ display: 'flex', flexDirection: 'column', gap: 12 }}
          value={visibleModules}
          onChange={(values) => setVisibleModules(values as string[])}
          options={workplaceModules}
        />
      </Modal>
    </PageContainer>
  );
};

export default Workplace;
