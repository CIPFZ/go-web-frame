import { EyeOutlined, ReadOutlined } from '@ant-design/icons';
import { Drawer, Tag, Typography } from 'antd';
import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useRef, useState } from 'react';
import { getNovelBookDetail, getNovelBookList, type NovelBook } from '@/services/novel';

const formatDate = (value?: string) => (value ? new Date(value).toLocaleString() : '-');

export default function NovelBooksPage() {
  const actionRef = useRef<ActionType>(null);
  const [detail, setDetail] = useState<NovelBook>();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const columns: ProColumns<NovelBook>[] = [
    { title: '书名', dataIndex: 'title', ellipsis: true, width: 260 },
    { title: '作者', dataIndex: 'author', ellipsis: true, width: 180 },
    { title: '分类', dataIndex: 'category', ellipsis: true, width: 140 },
    { title: '语言', dataIndex: 'language', width: 90, search: false },
    {
      title: '资源', dataIndex: 'sourceUrl', search: false, width: 100,
      render: (_, row) => /^https?:\/\//.test(row.sourceUrl || '') ? <a href={row.sourceUrl} target="_blank" rel="noreferrer">下载</a> : <Tag>暂无链接</Tag>,
    },
    {
      title: '格式', dataIndex: 'formats', search: false, width: 170,
      render: (_, row) => <>{(row.formats || []).map((format) => <Tag key={format}>{format.toUpperCase()}</Tag>)}</>,
    },
    { title: '更新时间', dataIndex: 'updatedAt', valueType: 'dateTime', search: false, width: 180 },
    {
      title: '操作', valueType: 'option', width: 90,
      render: (_, row) => [
        <a key="detail" onClick={async () => { const res = await getNovelBookDetail(row.id); if (res.data) { setDetail(res.data); setDrawerOpen(true); } }}>
          <EyeOutlined /> 查看
        </a>,
      ],
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<NovelBook>
        headerTitle={<><ReadOutlined /> 小说库</>}
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        request={async (params) => {
          const res = await getNovelBookList({
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.title,
            author: params.author,
            category: params.category,
          });
          return { data: res.data?.list || [], success: res.code === 0, total: res.data?.total || 0 };
        }}
        columns={columns}
        pagination={{ defaultPageSize: 20, showSizeChanger: true }}
        scroll={{ x: 1250 }}
        options={{ density: true, fullScreen: true, reload: true, setting: true }}
      />
      <Drawer title="小说详情" width={560} open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        {detail && <div style={{ display: 'grid', gap: 16 }}>
          <Typography.Title level={4} style={{ margin: 0 }}>{detail.title}</Typography.Title>
          <Typography.Paragraph type="secondary">{detail.author || '未知作者'} · {detail.category || '未分类'}</Typography.Paragraph>
          <dl style={{ display: 'grid', gridTemplateColumns: '120px 1fr', gap: '12px 8px', margin: 0 }}>
            <dt>语言</dt><dd>{detail.language || '-'}</dd>
            <dt>文件格式</dt><dd>{(detail.formats || []).map((format) => <Tag key={format}>{format.toUpperCase()}</Tag>)}</dd>
            <dt>资源下载</dt><dd style={{ wordBreak: 'break-all' }}>{/^https?:\/\//.test(detail.sourceUrl || '') ? <a href={detail.sourceUrl} target="_blank" rel="noreferrer">打开下载链接</a> : '暂无可用链接'}</dd>
            <dt>导入时间</dt><dd>{formatDate(detail.createdAt)}</dd>
            <dt>最后更新</dt><dd>{formatDate(detail.updatedAt)}</dd>
          </dl>
        </div>}
      </Drawer>
    </PageContainer>
  );
}
