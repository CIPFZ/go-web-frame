import React,{useEffect,useState} from 'react';
import {PageContainer} from '@ant-design/pro-components';
import {Button,Input,Space,Table,Tag,message} from 'antd';
import {PlayCircleOutlined,ReloadOutlined} from '@ant-design/icons';
import {listMediaItems,type MediaItem} from '@/services/media';

export default function MediaLibraryPage(){
 const [rows,setRows]=useState<MediaItem[]>([]);const [total,setTotal]=useState(0);const [loading,setLoading]=useState(false);const [query,setQuery]=useState('');
 const load=async(page=1)=>{setLoading(true);try{const r=await listMediaItems(page,query);if(r.code===0){setRows(r.data.list||[]);setTotal(r.data.total||0)}else message.error(r.msg||'加载影视库失败')}finally{setLoading(false)}};
 useEffect(()=>{load()},[]);
 return <PageContainer title="影视库" extra={<Space><Input placeholder="搜索标题" value={query} onChange={e=>setQuery(e.target.value)} onPressEnter={()=>load()} style={{width:220}}/><Button icon={<ReloadOutlined/>} onClick={()=>load()}>刷新</Button></Space>}>
  <Table<MediaItem> rowKey="id" loading={loading} dataSource={rows} pagination={{total,pageSize:20,onChange:load}} columns={[
   {title:'标题',dataIndex:'title'}, {title:'类型',dataIndex:'mediaType',render:v=><Tag>{v}</Tag>},{title:'年份',dataIndex:'year'},{title:'地区',dataIndex:'area'},
   {title:'状态',dataIndex:'status'}, {title:'操作',render:(_,item)=><Button type="link" icon={<PlayCircleOutlined/>} onClick={()=>{window.location.href='/media/player?id='+item.id}}>播放测试</Button>},
  ]}/>
 </PageContainer>
}
