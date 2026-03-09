import React, { useEffect, useMemo, useState } from 'react';
import { PageContainer, ProCard } from '@ant-design/pro-components';
import { Link, useRequest } from '@umijs/max';
import { Alert, Button, List, Progress, Space, Tag, Typography, message } from 'antd';
import dayjs from 'dayjs';
import { getServerState } from '@/services/api/state';
import { getMyNotices, markNoticeRead } from '@/services/api/notice';

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

const toPercent = (used = 0, total = 0) => {
  if (!total) return 0;
  return Number(((used / total) * 100).toFixed(2));
};

const Workplace: React.FC = () => {
  const serverQuery = useRequest(() => getServerState(), { pollingInterval: 10000 });

  const [noticeLoading, setNoticeLoading] = useState(false);
  const [noticeError, setNoticeError] = useState<string | null>(null);
  const [notices, setNotices] = useState<any[]>([]);
  const [noticeTotal, setNoticeTotal] = useState(0);

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

  const server = serverQuery.data?.data?.server;
  const unreadCount = notices.filter((n: any) => !n.readAt).length;

  const cpuAvg = useMemo(() => {
    const cpus: number[] = server?.cpu?.cpus || [];
    if (!cpus.length) return 0;
    return Number((cpus.reduce((sum, n) => sum + n, 0) / cpus.length).toFixed(2));
  }, [server]);

  const ramPercent = toPercent(server?.ram?.used || 0, server?.ram?.total || 0);

  return (
    <PageContainer title="工作台" subTitle="系统入口与个人待办">
      <ProCard gutter={[16, 16]} wrap>
        <ProCard title="快捷入口" colSpan={{ xs: 24, lg: 14 }}>
          <Space size={[8, 8]} wrap>
            {quickLinks.map((item) => (
              <Link key={item.path} to={item.path}>
                <Button>{item.title}</Button>
              </Link>
            ))}
          </Space>
        </ProCard>

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
                      <Text type="secondary">创建于 {dayjs(item.createdAt).format('YYYY-MM-DD HH:mm:ss')}</Text>
                    </Space>
                  }
                />
              </List.Item>
            )}
          />
        </ProCard>
      </ProCard>
    </PageContainer>
  );
};

export default Workplace;
