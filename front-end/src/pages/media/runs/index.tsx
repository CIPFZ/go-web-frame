import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Descriptions, Drawer, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { cancelCollectionRun, listCollectionRuns, listMediaSources, type CollectionRun, type MediaSource } from '@/services/media';
import { MediaHeading, modeNames, StatusTag, statusNames, timeText } from '../shared';

export default function Runs() {
 const [runs,setRuns]=useState<CollectionRun[]>([]),[sources,setSources]=useState<MediaSource[]>([]),[loading,setLoading]=useState(false),[source,setSource]=useState<number>(),[mode,setMode]=useState<string>(),[status,setStatus]=useState<string>(),[selected,setSelected]=useState<number>(),[pollError,setPollError]=useState(false);
 const load=async(silent=false)=>{if(!silent)setLoading(true);try{const r=await listCollectionRuns();if(r.code!==0)throw Error();setRuns(r.data||[]);setPollError(false)}catch{setPollError(true)}finally{if(!silent)setLoading(false)}};
 useEffect(()=>{void load();listMediaSources().then(r=>{if(r.code===0)setSources(r.data||[])}).catch(()=>message.error('站点名称加载失败'));const timer=window.setInterval(()=>void load(true),5000);return()=>clearInterval(timer)},[]);
 const active=runs.find(r=>r.id===selected);
 const cancel=async(id:number)=>{try{const r=await cancelCollectionRun(id);if(r.code!==0)throw Error(r.msg);message.success('已请求取消，任务将在当前步骤完成后停止');await load(true)}catch(e){message.error(e instanceof Error?e.message:'取消失败')}};
 const name=(id:number)=>sources.find(s=>s.id===id)?.name||'站点 #'+id;
 return <PageContainer title={false}><div className="media-workspace">
 <MediaHeading eyebrow="MEDIA / ACTIVITY" title="采集日志" description="查看自动任务的执行状态、处理数量与异常原因。" extra={<Space><Tag color="processing">每 5 秒更新</Tag><Button icon={<ReloadOutlined/>} onClick={()=>load()}>刷新</Button></Space>}/>
 <div className="media-stat-grid">{[['运行与排队',runs.filter(r=>['running','queued'].includes(r.status)).length],['已完成',runs.filter(r=>r.status==='success').length],['需要关注',runs.filter(r=>['failed','partial'].includes(r.status)).length],['最近任务',runs.length]].map(([label,value])=><div className="media-stat" key={label}><span className="media-muted">{label}</span><b>{value}</b></div>)}</div>
 {pollError&&<Alert type="warning" showIcon message="暂时无法刷新日志，当前显示上次成功加载的数据。" style={{marginBottom:16}}/>}
 <Card className="media-panel"><div className="media-toolbar"><Space wrap>
 <Select allowClear placeholder="全部站点" style={{width:180}} value={source} onChange={setSource} options={sources.map(s=>({value:s.id,label:s.name}))}/>
 <Select allowClear placeholder="全部类型" style={{width:150}} value={mode} onChange={setMode} options={Object.entries(modeNames).map(([value,label])=>({value,label}))}/>
 <Select allowClear placeholder="全部状态" style={{width:150}} value={status} onChange={setStatus} options={['queued','running','success','partial','failed','cancelled'].map(value=>({value,label:statusNames[value]}))}/>
 </Space><Typography.Text type="secondary">最近 {runs.length} 条记录</Typography.Text></div>
 <Table<CollectionRun> rowKey="id" loading={loading} dataSource={runs.filter(r=>(!source||r.sourceId===source)&&(!mode||r.mode===mode)&&(!status||r.status===status))} pagination={{pageSize:10,showSizeChanger:false}} scroll={{x:1050}} columns={[
 {title:'任务',render:(_,r)=><div><Typography.Text strong>{name(r.sourceId)}</Typography.Text><div className="media-muted">#{r.id} · {modeNames[r.mode]||r.mode}</div></div>},
 {title:'状态',render:(_,r)=><StatusTag value={r.status}/>},
 {title:'采集进度',render:(_,r)=>r.mode==='health'?<Typography.Text type="secondary">接口响应检查</Typography.Text>:<div><Typography.Text>{r.fetched} 条 · {r.pages} 页</Typography.Text><div className="media-muted">{['queued','running'].includes(r.status)?'下一页：'+r.nextPage:'处理结束'}</div></div>},
 {title:'处理结果',render:(_,r)=><div><span style={{color:'#389e0d'}}>+{r.insertedCount} 新增</span> · {r.updatedCount} 更新<div className="media-muted">{r.skippedCount} 跳过 · <span style={{color:r.errorCount?'#cf1322':undefined}}>{r.errorCount} 错误</span></div></div>},
 {title:'任务时间',render:(_,r)=><div>{timeText(r.startedAt)}<div className="media-muted">{r.finishedAt?'结束 '+timeText(r.finishedAt):r.status==='queued'?'等待执行':'正在执行'}</div></div>},
 {title:'操作',width:150,render:(_,r)=><Space><Button type="link" onClick={()=>setSelected(r.id)}>详情</Button>{['running','queued'].includes(r.status)&&<Popconfirm title="取消这个任务？" description="已采集的数据会保留。" onConfirm={()=>cancel(r.id)}><Button type="link" danger>取消</Button></Popconfirm>}</Space>}
 ]}/></Card>
 <Drawer width={560} title={'任务 #'+selected} open={!!selected} onClose={()=>setSelected(undefined)}>
 {active&&<><Space style={{marginBottom:24}}><StatusTag value={active.status}/><Tag>{modeNames[active.mode]}</Tag></Space><Descriptions column={1} bordered size="small" items={[
 {key:'source',label:'站点',children:name(active.sourceId)},{key:'start',label:'开始时间',children:timeText(active.startedAt)},{key:'end',label:'结束时间',children:timeText(active.finishedAt)},
 {key:'progress',label:'采集进度',children:active.fetched+' 条 / '+active.pages+' 页'},{key:'next',label:'下一页断点',children:active.nextPage||'已结束'},
 {key:'new',label:'新增 / 更新',children:active.insertedCount+' / '+active.updatedCount},{key:'skip',label:'跳过 / 错误',children:active.skippedCount+' / '+active.errorCount}
 ]}/><Typography.Title level={5} style={{marginTop:28}}>执行说明</Typography.Title><Alert showIcon type={active.errorMessage?'warning':'info'} message={active.errorMessage||'暂无异常记录。任务进度按已处理页数与条目数更新。'}/></>}
 </Drawer></div></PageContainer>;
}
