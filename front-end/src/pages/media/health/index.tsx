import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Col, Drawer, Empty, Pagination, Row, Space, Spin, Table, Tag, Typography, message } from 'antd';
import { ApiOutlined, ReloadOutlined } from '@ant-design/icons';
import { checkMediaSource, listMediaSources, listMediaSchedules, listSourceHealth, type MediaSource, type Schedule } from '@/services/media';
import { MediaHeading, StatusTag, timeText } from '../shared';
import ScheduleEditor from '../ScheduleEditor';
type Health = {id:number;checkedAt:string;result:string;httpStatus:number;latencyMs:number;errorMessage?:string};
export default function HealthPage() {
 const [sources,setSources]=useState<MediaSource[]>([]),[plans,setPlans]=useState<Schedule[]>([]),[history,setHistory]=useState<Record<number,Health[]>>({}),[loading,setLoading]=useState(false),[checking,setChecking]=useState<number[]>([]),[selected,setSelected]=useState<number>(),[editing,setEditing]=useState<Schedule>(),[page,setPage]=useState(1);
 const load=async()=>{setLoading(true);try{const [a,b]=await Promise.all([listMediaSources(),listMediaSchedules()]);if(a.code!==0||b.code!==0)throw Error();setSources(a.data||[]);setPlans(b.data||[]);const checks=await Promise.all((a.data||[]).map(async s=>{const r=await listSourceHealth(s.id);if(r.code!==0)throw Error();return [s.id,r.data||[]] as const}));setHistory(Object.fromEntries(checks))}catch{message.error('读取连通性状态失败')}finally{setLoading(false)}};
 useEffect(()=>{void load()},[]);
 const check=async(s:MediaSource)=>{setChecking(v=>[...v,s.id]);try{const r=await checkMediaSource(s.id);if(r.code!==0)throw Error(r.msg);if(r.data.result==='ok')message.success(s.name+'连接正常');else message.warning(s.name+'：'+(r.data.errorMessage||'连接异常'));const h=await listSourceHealth(s.id);if(h.code===0)setHistory(v=>({...v,[s.id]:h.data||[]}))}catch(e){message.error(e instanceof Error?e.message:'检查失败')}finally{setChecking(v=>v.filter(id=>id!==s.id))}};
 return <PageContainer title={false}><div className="media-workspace"><MediaHeading eyebrow="MEDIA / HEALTH" title="连通性检查" description="检查资源站接口是否可访问，并追踪响应耗时与异常。" extra={<Button icon={<ReloadOutlined/>} onClick={load}>刷新状态</Button>}/>
 <Alert type="info" showIcon message="仅检查接口与响应，不采集影视条目，不下载或缓存视频。" style={{marginBottom:24}}/>
 <Spin spinning={loading}><Row gutter={[20,20]}>{sources.slice((page-1)*10,page*10).map(s=>{const latest=history[s.id]?.[0],plan=plans.find(p=>p.sourceId===s.id&&p.kind==='health');return <Col xs={24} lg={12} xl={8} key={s.id}><Card className="media-panel" title={<Space><span className="media-source-dot"><ApiOutlined/></span>{s.name}</Space>} extra={latest?<StatusTag value={latest.result}/>:<Tag>未检查</Tag>}>
 <div style={{display:'flex',justifyContent:'space-between',marginBottom:22}}><div><div className="media-muted">响应耗时</div><Typography.Title level={2} style={{margin:'6px 0'}}>{latest?latest.latencyMs:'—'} <Typography.Text type="secondary" style={{fontSize:13}}>ms</Typography.Text></Typography.Title></div><div style={{textAlign:'right'}}><div className="media-muted">最近检查</div><div style={{marginTop:12}}>{timeText(latest?.checkedAt)}</div></div></div>
 {latest?.errorMessage&&<Alert type="warning" message={latest.errorMessage} style={{marginBottom:20}}/>}
 <div className="media-episode-panel" style={{margin:'0 0 20px'}}><Space><Typography.Text strong>自动检查</Typography.Text><Tag color={plan?.enabled?'blue':'default'}>{plan?.enabled?'已启用':'未启用'}</Tag></Space><div className="media-muted" style={{marginTop:10}}>下次：{plan?.enabled?timeText(plan.nextRunAt):'未安排'}</div><Space style={{marginTop:8}}><Typography.Text code>{plan?.cron||'尚未初始化'}</Typography.Text>{plan&&<Button type="link" size="small" onClick={()=>setEditing(plan)}>调整周期</Button>}</Space></div>
 <Space wrap><Button type="primary" ghost icon={<ApiOutlined/>} loading={checking.includes(s.id)} onClick={()=>check(s)}>立即检查</Button><Button onClick={()=>setSelected(s.id)}>检查历史</Button></Space>
 </Card></Col>})}</Row>{!sources.length&&!loading&&<Empty description="暂无资源站点"/>}</Spin>
 <Drawer title={(sources.find(s=>s.id===selected)?.name||'')+' · 检查历史'} width={680} open={!!selected} onClose={()=>setSelected(undefined)}><Table<Health> rowKey="id" size="small" dataSource={selected?history[selected]:[]} pagination={{pageSize:10}} columns={[
 {title:'时间',dataIndex:'checkedAt',render:timeText},{title:'结果',dataIndex:'result',render:v=><StatusTag value={v}/>},{title:'HTTP',dataIndex:'httpStatus',render:v=>v||'—'},{title:'耗时',dataIndex:'latencyMs',render:v=>v+' ms'}
 ]} expandable={{rowExpandable:r=>!!r.errorMessage,expandedRowRender:r=><Typography.Text type="danger">{r.errorMessage}</Typography.Text>}}/></Drawer>
 <ScheduleEditor schedule={editing} sourceName={sources.find(s=>s.id===editing?.sourceId)?.name} onClose={()=>setEditing(undefined)} onSaved={()=>void load()}/>
 </div></PageContainer>;
}
