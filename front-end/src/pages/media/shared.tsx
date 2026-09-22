import React from 'react';
import { Tag, Typography } from 'antd';
import './style.less';
export const timeText = (v?: string) => v && !v.startsWith('0001') ? new Date(v).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai', hour12: false }) : '—';
export const modeNames: Record<string, string> = { health: '连通性检查', sample: '试运行', incremental: '增量采集', full: '全量采集' };
export const statusNames: Record<string, string> = { queued: '排队中', running: '执行中', success: '已完成', partial: '部分完成', failed: '失败', cancelled: '已取消', skipped: '已跳过', ok: '正常', error: '异常' };
const colors: Record<string, string> = { queued: 'blue', running: 'processing', success: 'green', partial: 'orange', failed: 'red', ok: 'green', error: 'red' };
export const StatusTag = ({ value }: { value?: string }) => <Tag color={colors[value || '']}>{statusNames[value || ''] || value || '未执行'}</Tag>;
export function MediaHeading({ eyebrow, title, description, extra }: { eyebrow: string; title: string; description: string; extra?: React.ReactNode }) {
 return <div className="media-heading"><div><div className="media-eyebrow">{eyebrow}</div><Typography.Title level={2}>{title}</Typography.Title><Typography.Text type="secondary">{description}</Typography.Text></div><div className="media-heading-actions">{extra}</div></div>;
}
export function Cover({ url, title, small = false }: { url?: string; title: string; small?: boolean }) {
 const [failed, setFailed] = React.useState(false);
 React.useEffect(() => setFailed(false), [url]);
 return <div className={small ? 'media-cover media-cover-small' : 'media-cover'}>{url && !failed ? <img src={url} alt={title} loading="lazy" onError={() => setFailed(true)} /> : <div className="media-cover-placeholder"><span>映画</span><small>暂无封面</small></div>}</div>;
}
