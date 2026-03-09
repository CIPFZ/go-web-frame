import React, { useRef } from 'react';
import { PageContainer, ProTable, ActionType, ProColumns } from '@ant-design/pro-components';
import { Button, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { ModalForm, ProFormText, ProFormDigit } from '@ant-design/pro-components';
import { getDynastyList, createDynasty, updateDynasty, deleteDynasty, Dynasty } from '@/services/api/poetry';

const DynastyList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);

  const handleDelete = async (ID: number) => {
    try {
      await deleteDynasty(ID);
      message.success('删除成功');
      actionRef.current?.reload();
    } catch (error: any) {
      // ✨ 优化：如果后端返回了具体错误（如“该朝代下仍有诗人”），优先显示
      // Ant Design Pro 的 request 拦截器可能会处理，如果没有拦截，这里 catch
      console.error(error);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (values.ID) {
        await updateDynasty(values.ID, values);
        message.success('更新成功');
      } else {
        await createDynasty(values);
        message.success('创建成功');
      }
      actionRef.current?.reload();
      return true;
    } catch (error) {
      return false;
    }
  };

  const columns: ProColumns<Dynasty>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      hideInForm: true,
      width: 80,
      hideInSearch: true,
    },
    {
      title: '朝代名称',
      dataIndex: 'name',
      formItemProps: { rules: [{ required: true }] },
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
        <ModalForm<Dynasty>
          key={`edit-${record.ID}`} // ✨ 强制刷新表单
          title="编辑朝代"
          trigger={<a><EditOutlined /> 编辑</a>}
          onFinish={handleSubmit}
          initialValues={record}
        >
          <ProFormText name="ID" hidden />
          <ProFormText name="name" label="朝代名称" rules={[{ required: true }]} />
          <ProFormDigit name="sortOrder" label="排序权重" min={0} />
        </ModalForm>,
        <Popconfirm
          key="del"
          title="确定删除？"
          description="如果该朝代下有诗人，删除将失败。"
          onConfirm={() => handleDelete(record.ID)}
        >
          <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<Dynasty>
        headerTitle="朝代列表"
        actionRef={actionRef}
        rowKey="ID"
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <ModalForm
            key="create"
            title="新建朝代"
            trigger={<Button type="primary"><PlusOutlined /> 新建</Button>}
            onFinish={handleSubmit}
          >
            <ProFormText name="name" label="朝代名称" rules={[{ required: true }]} />
            <ProFormDigit name="sortOrder" label="排序权重" initialValue={0} />
          </ModalForm>,
        ]}
        request={async (params) => {
          const res = await getDynastyList({ ...params, page: params.current, pageSize: params.pageSize });
          return { data: res.data.list, success: res.code === 0, total: res.data.total };
        }}
        columns={columns}
      />
    </PageContainer>
  );
};

export default DynastyList;
