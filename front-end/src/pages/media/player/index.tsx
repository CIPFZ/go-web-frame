import React,{useEffect,useRef,useState} from 'react';
import {PageContainer} from '@ant-design/pro-components';
import {Alert,Button,Card,Empty,List,Select,Space,Spin,Typography,message} from 'antd';
import Hls from 'hls.js';
import {getMediaItem,type MediaDetail} from '@/services/media';

export default function MediaPlayerPage(){
 const id=Number(new URLSearchParams(window.location.search).get('id')||0);
 const [detail,setDetail]=useState<MediaDetail>();const [line,setLine]=useState<number>();const [episode,setEpisode]=useState<number>();const [loading,setLoading]=useState(true);const video=useRef<HTMLVideoElement>(null);const hls=useRef<Hls | undefined>(undefined);
 useEffect(()=>{getMediaItem(id).then(r=>{if(r.code===0){setDetail(r.data);const first=r.data.playbacks[0]?.lines[0];setLine(first?.id);setEpisode(first?.episodes[0]?.id)}else message.error(r.msg||'加载失败')}).finally(()=>setLoading(false));return()=>hls.current?.destroy()},[id]);
 const selected=detail?.playbacks.flatMap(p=>p.lines).find(x=>x.id===line);const ep=selected?.episodes.find(x=>x.id===episode)||selected?.episodes[0];
 useEffect(()=>{if(!video.current||!ep?.streamUrl)return;hls.current?.destroy();if(Hls.isSupported()){const x=new Hls({enableWorker:true});x.loadSource(ep.streamUrl);x.attachMedia(video.current);hls.current=x}else{video.current.src=ep.streamUrl}},[ep?.streamUrl]);
 if(loading)return <PageContainer><Spin/></PageContainer>;if(!detail)return <PageContainer><Empty/></PageContainer>;
 return <PageContainer title={detail.item.title} subTitle="后台播放测试，CMS 不代理或缓存视频流"><Space direction="vertical" style={{width:'100%'}} size="large"><Card><video ref={video} controls playsInline style={{width:'100%',maxHeight:560,background:'#000'}}/><Alert type="info" showIcon message="播放地址由资源站直接提供，可能包含临时参数；CMS 仅在浏览器端播放。" style={{marginTop:12}}/></Card><Card title="播放线路"><Space wrap>{detail.playbacks.flatMap(p=>p.lines).map(l=><Button key={l.id} type={line===l.id?'primary':'default'} onClick={()=>{setLine(l.id);setEpisode(l.episodes[0]?.id)}}>{l.sourceName||l.lineName}</Button>)}</Space>{selected&&<Select style={{minWidth:260,marginTop:16}} value={episode||selected.episodes[0]?.id} onChange={setEpisode} options={selected.episodes.map(e=>({value:e.id,label:e.episodeName||'第'+e.episodeNo+'集'}))}/>}</Card><Typography.Text type="secondary">线路健康检查只验证接口和播放清单，不下载视频分片，也不保存播放内容。</Typography.Text></Space></PageContainer>;
}
