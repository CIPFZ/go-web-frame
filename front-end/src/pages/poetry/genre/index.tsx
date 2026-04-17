import React, { useRef } from 'react';
import { PageContainer, ProTable, ActionType, ProColumns } from '@ant-design/pro-components';
import { Button, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { ModalForm, ProFormText, ProFormDigit } from '@ant-design/pro-components';
import { getGenreList, createGenre, updateGenre, deleteGenre, Genre } from '@/services/api/poetry';

const GenreList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);

  // ✨ 修改参数为 ID
  const handleDelete = async (ID: number) => {
    try {
      await deleteGenre(ID);
      message.success('删除成功');
      actionRef.current?.reload();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      // ✨ 检查 values.ID
      if (values.ID) {
        await updateGenre(values.ID, values);
        message.success('更新成功');
      } else {
        await createGenre(values);
        message.success('创建成功');
      }
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error('操作失败');
      return false;
    }
  };

  const columns: ProColumns<Genre>[] = [
    {
      title: 'ID',
      dataIndex: 'ID', // ✨✨✨ 关键修改：id -> ID (匹配后端)
      hideInForm: true,
      width: 80,
      hideInSearch: true,
    },
    {
      title: '体裁名称',
      dataIndex: 'name',
      formItemProps: {
        rules: [{ required: true, message: '请输入名称' }],
      },
    },
    {
      title: '排序权重',
      dataIndex: 'sortOrder',
      valueType: 'digit',
      sorter: true,
      hideInSearch: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 180,
      render: (_, record) => [
        <ModalForm<Genre>
          key={`edit-${record.ID}`}
          title="编辑体裁"
          trigger={
            <a><EditOutlined /> 编辑</a>
          }
          onFinish={handleSubmit}
          initialValues={record}
        >
          {/* ✨✨✨ 关键修改：隐藏域的 name 也要改成 ID */}
          <ProFormText name="ID" hidden />

          <ProFormText name="name" label="体裁名称" rules={[{ required: true }]} />
          <ProFormDigit name="sortOrder" label="排序权重" min={0} />
        </ModalForm>,
        <Popconfirm
          key="delete"
          title="确定删除吗？"
          // ✨✨✨ 关键修改：传入 record.ID
          onConfirm={() => handleDelete(record.ID)}
        >
          <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<Genre>
        headerTitle="体裁列表"
        actionRef={actionRef}
        // ✨✨✨ 关键修改：rowKey 必须指定为 'ID'，否则选中/展开会失效
        rowKey="ID"
        search={{
          labelWidth: 'auto',
        }}
        toolBarRender={() => [
          <ModalForm
            key="create"
            title="新建体裁"
            trigger={
              <Button type="primary"><PlusOutlined /> 新建</Button>
            }
            onFinish={handleSubmit}
          >
            <ProFormText name="name" label="体裁名称" rules={[{ required: true }]} />
            <ProFormDigit name="sortOrder" label="排序权重" initialValue={0} />
          </ModalForm>,
        ]}
        request={async (params) => {
          const res = await getGenreList({ ...params, page: params.current, pageSize: params.pageSize });
          return {
            data: res.data.list,
            success: res.code === 0,
            total: res.data.total,
          };
        }}
        columns={columns}
      />
    </PageContainer>
  );
};

export default GenreList;
