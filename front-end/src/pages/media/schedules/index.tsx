import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Empty, Pagination, Select, Space, Spin, Tag, Typography, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { listMediaSchedules, listMediaSources, type MediaSource, type Schedule } from '@/services/media';
import { MediaHeading, modeNames, StatusTag, timeText } from '../shared';
import ScheduleEditor from '../ScheduleEditor';

export default function Schedules() {
 const [sources,setSources]=useState<MediaSource[]>([]),[plans,setPlans]=useState<Schedule[]>([]),[loading,setLoading]=useState(false),[editing,setEditing]=useState<Schedule>(),[filter,setFilter]=useState<number>(),[page,setPage]=useState(1);
 const load=async()=>{setLoading(true);try{const [a,b]=await Promise.all([listMediaSources(),listMediaSchedules()]);if(a.code!==0||b.code!==0)throw Error();setSources(a.data||[]);setPlans(b.data||[])}catch{message.error('加载采集计划失败')}finally{setLoading(false)}};
 useEffect(()=>{void load()},[]);
 return <PageContainer title={false}><div className="media-workspace">
 <MediaHeading eyebrow="MEDIA / SCHEDULES" title="采集计划" description="每个站点独立选择采集类型和周期，由 CMS 自动执行。" extra={<Button icon={<ReloadOutlined/>} onClick={load}>刷新</Button>}/>
 <Alert type="info" showIcon message="增量用于日常更新，全量用于完整建库与周期校准" description={<div>增量首次读取最近 24 小时；之后根据上次成功时间扩大更新窗口，跳过未变化条目。全量遍历整个目录，可从断点继续，完整且无错误后才标记已下架资源。</div>} style={{marginBottom:24}}/>
 <div className="media-toolbar"><Select allowClear placeholder="全部站点" style={{width:240}} value={filter} onChange={value=>{setFilter(value);setPage(1)}} options={sources.map(s=>({value:s.id,label:s.name}))}/><Typography.Text type="secondary">所有时间均为北京时间</Typography.Text></div>
 <Spin spinning={loading}><Space direction="vertical" size={20} style={{width:'100%'}}>
 {sources.filter(s=>!filter||s.id===filter).slice((page-1)*10,page*10).map(source=><Card className="media-panel" key={source.id} title={<Space><span className="media-source-dot">{source.name.slice(0,1)}</span>{source.name}</Space>} extra={<Tag color={source.enabled?'green':'default'}>{source.enabled?'站点已启用':'站点已停用'}</Tag>}>
 {!source.enabled&&<Alert message="站点停用期间，定时任务将被跳过。" type="warning" style={{marginBottom:16}}/>}
 {['incremental','full'].map(kind=>{const p=plans.find(s=>s.sourceId===source.id&&s.kind===kind);return p?<div key={kind} className="media-schedule-row">
 <div><Space><Typography.Text strong>{modeNames[kind]}</Typography.Text><Tag color={p.enabled?'blue':'default'}>{p.enabled?'已启用':'未启用'}</Tag></Space><div className="media-muted" style={{marginTop:8}}>{kind==='incremental'?'跟进新增与更新，适合高频执行':'遍历目录并校准，适合低频执行'}</div></div>
 <div><Typography.Text code>{p.cron}</Typography.Text><div className="media-muted" style={{marginTop:8}}>下次：{p.enabled?timeText(p.nextRunAt):'未安排'}</div></div>
 <div><StatusTag value={p.lastStatus}/><div className="media-muted" style={{marginTop:8}}>{timeText(p.lastRunAt)}</div>{p.lastError&&<Typography.Text type="danger" ellipsis={{tooltip:p.lastError}} style={{maxWidth:170}}>{p.lastError}</Typography.Text>}</div>
 <Button onClick={()=>setEditing(p)}>设置计划</Button>
 </div>:<Empty key={kind} image={Empty.PRESENTED_IMAGE_SIMPLE} description={modeNames[kind]+'计划尚未初始化'}/>})}
 </Card>)}
 {!sources.length&&!loading&&<Empty description="请先添加资源站点"/>}{sources.filter(s=>!filter||s.id===filter).length>10&&<div className="media-pagination"><Pagination current={page} pageSize={10} total={sources.filter(s=>!filter||s.id===filter).length} showSizeChanger={false} onChange={setPage}/></div>}
 </Space></Spin>
 <ScheduleEditor schedule={editing} sourceName={sources.find(s=>s.id===editing?.sourceId)?.name} fullAllowed={sources.find(s=>s.id===editing?.sourceId)?.fullEnabled} onClose={()=>setEditing(undefined)} onSaved={()=>void load()}/>
 </div></PageContainer>;
}
