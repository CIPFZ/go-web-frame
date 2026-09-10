import React, { useCallback, useEffect, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Skeleton, Tag, Typography } from 'antd';
import { ApiOutlined, ArrowUpOutlined, CheckCircleFilled, ClockCircleOutlined, CloudServerOutlined, DatabaseOutlined, DeploymentUnitOutlined, LineChartOutlined, ReloadOutlined, SafetyCertificateOutlined, WarningFilled } from '@ant-design/icons';
import { createStyles } from 'antd-style';
import { getServerState, type SystemStatus } from '@/services/system/state';

const useStyles = createStyles(({ token }) => ({
  stack: {display: 'grid', gap: 24},
  hero: {position: 'relative', overflow: 'hidden', border: `1px solid ${token.colorBorderSecondary}`, borderRadius: 16, padding: '32px 36px', background: `linear-gradient(115deg, ${token.colorBgContainer} 30%, ${token.colorPrimaryBg} 100%)`, display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 24, '@media (max-width: 700px)': {padding: 24, alignItems: 'flex-start', flexDirection: 'column'}},
  eyebrow: {color: token.colorTextSecondary, fontSize: 12, letterSpacing: 2, marginBottom: 16},
  headline: {display: 'flex', alignItems: 'center', gap: 14, fontSize: 28, fontWeight: 600, lineHeight: 1.35, marginBottom: 10, '@media (max-width: 700px)': {fontSize: 23}},
  heroMeta: {display: 'flex', gap: 20, flexWrap: 'wrap', color: token.colorTextSecondary, fontSize: 13},
  heroIcon: {fontSize: 86, color: token.colorPrimary, opacity: 0.15, transform: 'rotate(-12deg)', marginRight: 12, '@media (max-width: 700px)': {display: 'none'}},
  sectionTitle: {display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 14, gap: 12, '& h3': {fontSize: 16, fontWeight: 600, margin: 0}},
  grid: {display: 'grid', gridTemplateColumns: 'repeat(4, minmax(0, 1fr))', gap: 16, '@media (max-width: 1100px)': {gridTemplateColumns: 'repeat(2, minmax(0, 1fr))'}, '@media (max-width: 550px)': {gridTemplateColumns: '1fr'}},
  service: {border: `1px solid ${token.colorBorderSecondary}`, background: token.colorBgContainer, borderRadius: 12, padding: 22, minWidth: 0},
  serviceTop: {display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24},
  serviceIcon: {height: 40, width: 40, borderRadius: 10, background: token.colorFillAlter, display: 'grid', placeItems: 'center', fontSize: 20, color: token.colorPrimary},
  serviceName: {fontSize: 16, fontWeight: 600, marginBottom: 6},
  muted: {color: token.colorTextSecondary, fontSize: 13, lineHeight: 1.7},
  lower: {display: 'grid', gridTemplateColumns: 'minmax(0, 1.3fr) minmax(0, 1fr)', gap: 20, '@media (max-width: 900px)': {gridTemplateColumns: '1fr'}},
  panel: {border: `1px solid ${token.colorBorderSecondary}`, background: token.colorBgContainer, borderRadius: 12, padding: 26, minWidth: 0},
  facts: {display: 'grid', gridTemplateColumns: '1fr 1fr', columnGap: 24, rowGap: 26, marginTop: 26, '& dt': {color: token.colorTextSecondary, fontSize: 12, marginBottom: 8}, '& dd': {margin: 0, fontSize: 16, fontWeight: 500, overflowWrap: 'anywhere'}, '@media (max-width: 550px)': {gridTemplateColumns: '1fr'}},
  observability: {display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 18},
  features: {display: 'flex', gap: 8, flexWrap: 'wrap', '& span': {padding: '5px 10px', background: token.colorFillAlter, color: token.colorTextSecondary, borderRadius: 6, fontSize: 12}},
  footer: {display: 'flex', alignItems: 'center', gap: 8, fontSize: 12, color: token.colorTextTertiary},
}));
const services = [
  {key: 'backend', name: '后端服务', description: '处理平台请求与权限校验', icon: <CloudServerOutlined />},
  {key: 'database', name: '数据库', description: '保存账号、菜单和平台数据', icon: <DatabaseOutlined />},
  {key: 'redis', name: 'Redis', description: '提供缓存与共享限流', icon: <DeploymentUnitOutlined />},
  {key: 'mongo', name: 'MongoDB', description: '按需启用的文档存储', icon: <ApiOutlined />},
];
function uptime(seconds: number) {
  const minutes = Math.floor(seconds / 60);
  if (minutes < 1) return '不足 1 分钟';
  if (minutes < 60) return `${minutes} 分钟`;
  if (minutes < 1440) return `${Math.floor(minutes / 60)} 小时 ${minutes % 60} 分钟`;
  return `${Math.floor(minutes / 1440)} 天 ${Math.floor(minutes % 1440 / 60)} 小时`;
}
export default function SystemStatusPage() {
  const { styles, theme } = useStyles();
  const [status, setStatus] = useState<SystemStatus>();
  const [error, setError] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [checkedAt, setCheckedAt] = useState<Date>();
  const mounted = useRef(false);
  const pending = useRef(false);
  const refresh = useCallback(async () => {
    if (pending.current) return;
    pending.current = true; setRefreshing(true);
    try {
      const result = await getServerState();
      if (mounted.current) {
        setError(result.code !== 0);
        if (result.code === 0) { setStatus(result.data.server); setCheckedAt(new Date()); }
      }
    } catch { if (mounted.current) setError(true); }
    finally { pending.current = false; if (mounted.current) setRefreshing(false); }
  }, []);
  useEffect(() => {
    mounted.current = true; void refresh();
    const timer = setInterval(() => { if (!document.hidden) void refresh(); }, 15000);
    return () => { mounted.current = false; clearInterval(timer); };
  }, [refresh]);
  const unavailable = status && Object.values(status.checks).some(value => value === 'unavailable');
  const active = status ? Object.values(status.checks).filter(value => value !== 'disabled').length : 0;
  const good = status ? Object.values(status.checks).filter(value => value === 'ok').length : 0;
  const healthy = !!status && !unavailable && !error;
  const title = error ? '暂时无法确认平台状态' : unavailable ? '部分服务需要关注' : status ? '平台运行正常' : '正在检查平台状态';
  return <PageContainer title="系统状态" subTitle="服务可用性与运行概览" extra={<Button icon={<ReloadOutlined spin={refreshing} />} loading={refreshing} onClick={() => void refresh()}>刷新状态</Button>}>
    <div className={styles.stack}>
      <section className={styles.hero} aria-label="整体运行状态">
        <div>
          <div className={styles.eyebrow}>运行概览 · BASE FRAME</div>
          <div className={styles.headline} role="status">
            {status || error ? healthy ? <CheckCircleFilled style={{color: theme.colorSuccess}} /> : <WarningFilled style={{color: theme.colorWarning}} /> : <ClockCircleOutlined />}
            {title}
          </div>
          <div className={styles.heroMeta}>
            <span>{status ? `${good} / ${active} 项已启用服务检查正常` : '正在连接后端服务'}</span>
            <span><ClockCircleOutlined /> {checkedAt ? `最近更新 ${checkedAt.toLocaleTimeString('zh-CN', {hour12: false})}` : '等待首次检查'}</span>
          </div>
        </div>
        <SafetyCertificateOutlined className={styles.heroIcon} />
      </section>
      {error && <Alert type="warning" showIcon message="状态更新失败" description="以下数据来自最近一次成功检查，请刷新重试。" />}
      {!status ? !error && <Skeleton active paragraph={{rows: 5}} /> : <>
        <section aria-label="服务可用性">
          <div className={styles.sectionTitle}><h3>服务可用性</h3><span className={styles.muted}>自动检查 · 每 15 秒</span></div>
          <div className={styles.grid}>{services.map(service => {
            const value = status.checks[service.key];
            const label = error ? '待确认' : value === 'ok' ? '正常' : value === 'disabled' ? '未启用' : '不可用';
            return <article className={styles.service} key={service.key} aria-label={`${service.name}：${label}`}>
              <div className={styles.serviceTop}><div className={styles.serviceIcon}>{service.icon}</div><Tag bordered={false} color={error ? 'warning' : value === 'ok' ? 'success' : value === 'disabled' ? 'default' : 'error'} style={{margin: 0}}>{label}</Tag></div>
              <div className={styles.serviceName}>{service.name}</div><div className={styles.muted}>{service.description}</div>
            </article>;
          })}</div>
        </section>
        <div className={styles.lower}>
          <section className={styles.panel} aria-label="平台信息">
            <div className={styles.sectionTitle}><h3>平台信息</h3><Tag bordered={false}>当前实例</Tag></div>
            <dl className={styles.facts}>
              <div><dt>平台版本</dt><dd>{status.version}</dd></div>
              <div><dt>本次运行时长</dt><dd>{uptime(status.uptimeSeconds)}</dd></div>
              <div><dt>Go 运行环境</dt><dd>{status.goVersion}</dd></div>
              <div><dt>数据库结构</dt><dd style={{fontSize: 12, fontFamily: 'monospace'}}>{status.schemaVersion}</dd></div>
            </dl>
          </section>
          <section className={`${styles.panel} ${styles.observability}`} aria-label="可观测性">
            <div className={styles.sectionTitle} style={{marginBottom: 0, width: '100%'}}><h3><LineChartOutlined style={{marginRight: 8, color: theme.colorPrimary}} />深入查看运行情况</h3></div>
            <Typography.Paragraph className={styles.muted} style={{margin: 0}}>从当前状态进一步追踪历史变化，查看应用日志与请求调用链，定位异常发生的原因。</Typography.Paragraph>
            <div className={styles.features}><span>请求趋势</span><span>应用日志</span><span>调用链</span><span>告警状态</span></div>
            {status.observabilityEnabled ? <Button type="primary" href="/observability/d/base-frame/base-frame" target="_blank" rel="noopener noreferrer" style={{marginTop: 'auto'}}>打开观测平台 <ArrowUpOutlined rotate={45} /></Button> : <Tag style={{marginTop: 'auto'}} bordered={false}>观测平台未启用</Tag>}
          </section>
        </div>
      </>}
      <div className={styles.footer}><SafetyCertificateOutlined /> 当前页展示服务即时状态，历史记录保存在观测平台。</div>
    </div>
  </PageContainer>;
}
