import React, { useRef, useState } from 'react';
import { PageContainer, ProTable, ActionType, ProColumns, DrawerForm } from '@ant-design/pro-components';
import { Button, message, Popconfirm, Tag, Space, Typography } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { ProFormText, ProFormSelect, ProFormTextArea } from '@ant-design/pro-components';
import {
  getPoemList, createPoem, updatePoem, deletePoem,
  getAllGenres, getAuthorList,
  PoemWork
} from '@/services/api/poetry';

const PoemList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentRow, setCurrentRow] = useState<PoemWork | undefined>(undefined);

  // 辅助数据加载
  const requestGenres = async () => (await getAllGenres()).data.map((i: any) => ({ label: i.name, value: i.ID }));

  // 远程搜索诗人 (支持输入关键词)
  const requestAuthors = async (params: { keyWords?: string }) => {
    const res = await getAuthorList({ name: params.keyWords, pageSize: 50 });
    return res.data.list.map((i: any) => ({ label: i.name, value: i.ID }));
  };

  const handleSubmit = async (values: any) => {
    try {
      if (values.ID) {
        await updatePoem(values.ID, values);
        message.success('保存成功');
      } else {
        await createPoem(values);
        message.success('创建成功');
      }
      setDrawerVisible(false);
      actionRef.current?.reload();
      return true;
    } catch (e) { return false; }
  };

  const columns: ProColumns<PoemWork>[] = [
    { title: 'ID', dataIndex: 'ID', width: 60, search: false },
    {
      title: '标题',
      dataIndex: 'title',
      copyable: true,
      formItemProps: { rules: [{ required: true }] }
    },
    {
      title: '作者',
      dataIndex: 'authorId',
      valueType: 'select',
      // 这里 request 会自动处理远程搜索
      request: requestAuthors,
      debounceTime: 500, // 防抖
      fieldProps: { showSearch: true },
      render: (_, row) => (
        <Space>
          <span style={{ fontWeight: 'bold' }}>{row.author?.name}</span>
          {row.author?.dynasty && <Tag>{row.author.dynasty.name}</Tag>}
        </Space>
      )
    },
    {
      title: '体裁',
      dataIndex: 'genreId',
      valueType: 'select',
      request: requestGenres,
      render: (_, row) => row.genre ? <Tag color="cyan">{row.genre.name}</Tag> : '-'
    },
    {
      title: '正文',
      dataIndex: 'content',
      ellipsis: true,
      search: false,
      width: 300,
      render: (text) => <Typography.Text type="secondary" style={{ fontSize: 13 }}>{text}</Typography.Text>
    },
    {
      title: '操作',
      valueType: 'option',
      fixed: 'right',
      width: 150,
      render: (_, record) => [
        <a key="edit" onClick={() => { setCurrentRow(record); setDrawerVisible(true); }}>
          <EditOutlined /> 编辑
        </a>,
        <Popconfirm key="del" title="确认删除?" onConfirm={async () => { await deletePoem(record.ID); actionRef.current?.reload(); }}>
          <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
        </Popconfirm>
      ]
    },
  ];

  return (
    <PageContainer>
      <ProTable<PoemWork>
        headerTitle="诗词作品库"
        actionRef={actionRef}
        rowKey="ID"
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => { setCurrentRow(undefined); setDrawerVisible(true); }}>
            <PlusOutlined /> 新建作品
          </Button>
        ]}
        request={async (params) => {
          // 参数映射：ProTable 默认传 title, authorId 等，需适配后端 searchReq
          const res = await getPoemList({
            ...params,
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.title // 如果用户在标题栏输入，传给后端 keyword
          });
          return { data: res.data.list, success: true, total: res.data.total };
        }}
        columns={columns}
      />

      <DrawerForm<PoemWork>
        title={currentRow ? "编辑作品" : "新建作品"}
        width={600}
        open={drawerVisible}
        onOpenChange={setDrawerVisible}
        initialValues={currentRow}
        onFinish={handleSubmit}
        // 当关闭时销毁表单，防止数据缓存
        drawerProps={{ destroyOnClose: true }}
      >
        <ProFormText name="ID" hidden />

        <ProFormText
          name="title"
          label="标题/词牌名"
          rules={[{ required: true }]}
          placeholder="例如：静夜思"
        />

        <div style={{ display: 'flex', gap: 16 }}>
          <ProFormSelect
            name="authorId"
            label="诗人"
            width="md"
            rules={[{ required: true }]}
            showSearch
            request={requestAuthors}
            placeholder="输入名字搜索..."
          />
          <ProFormSelect
            name="genreId"
            label="体裁"
            width="sm"
            rules={[{ required: true }]}
            request={requestGenres}
          />
        </div>

        <ProFormTextArea
          name="content"
          label="正文"
          rules={[{ required: true }]}
          fieldProps={{ rows: 6, showCount: true }}
        />

        <ProFormTextArea name="translation" label="译文" fieldProps={{ rows: 4 }} />
        <ProFormTextArea name="annotation" label="注释" fieldProps={{ rows: 3 }} placeholder="支持 Markdown" />
        <ProFormTextArea name="appreciation" label="赏析" fieldProps={{ rows: 4 }} />
        <ProFormText name="audioUrl" label="音频 URL" />

      </DrawerForm>
    </PageContainer>
  );
};

export default PoemList;
