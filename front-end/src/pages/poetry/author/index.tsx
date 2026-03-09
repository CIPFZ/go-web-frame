import React, { useRef } from 'react';
import {
  ActionType,
  ModalForm,
  PageContainer,
  ProColumns,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProTable
} from '@ant-design/pro-components';
import { Form, Image, message, Popconfirm, Tag, Button } from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import UploadImage from '@/components/Upload/UploadImage';
import { Author, createAuthor, deleteAuthor, getAllDynasties, getAuthorList, updateAuthor } from '@/services/api/poetry';

const AuthorList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);

  const requestDynasties = async () => {
    const res = await getAllDynasties();
    return res.data.map((item: any) => ({ label: item.name, value: item.ID }));
  };

  const handleSubmit = async (values: any) => {
    try {
      if (values.ID) {
        await updateAuthor(values.ID, values);
        message.success('更新成功');
      } else {
        await createAuthor(values);
        message.success('创建成功');
      }
      actionRef.current?.reload();
      return true;
    } catch (e) { return false; }
  };

  // ✨ 抽取表单字段：为了复用于“新建”和“编辑”
  // 传入 record 用于判断是编辑模式还是新建模式
  const renderFormFields = (record?: Author) => {
    const isEdit = !!record?.ID;
    return (
      <>
        {/* 只有编辑时才有 ID */}
        {isEdit && <ProFormText name="ID" hidden />}

        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 24 }}>
          <Form.Item name="avatarUrl" noStyle>
            <UploadImage
              // ✨ 关键逻辑：
              // 编辑模式 -> 用专用接口 (可以自动整理文件到 ID 文件夹)
              // 新建模式 -> 用通用接口 (因为还没有 ID)
              action={isEdit ? "/api/v1/poetry/author/avatar" : "/api/v1/sys/file/upload"}
              // 编辑模式传 ID，新建模式不传
              data={isEdit ? { id: record.ID } : undefined}
              circle={false}
            />
          </Form.Item>
        </div>

        <ProFormText name="name" label="姓名" rules={[{ required: true }]} />
        <ProFormSelect name="dynastyId" label="朝代" rules={[{ required: true }]} request={requestDynasties} />
        <ProFormTextArea name="intro" label="简介" />
        <ProFormTextArea name="lifeStory" label="生平" />
      </>
    );
  };

  const columns: ProColumns<Author>[] = [
    { title: 'ID', dataIndex: 'ID', hideInForm: true, width: 60, hideInSearch: true },
    {
      title: '头像',
      dataIndex: 'avatarUrl',
      hideInSearch: true,
      width: 80,
      render: (_, record) => {
        // ✨ 解决问题1：列表图片缓存
        // 使用 updatedAt (更新时间) 作为版本号。如果后端没返回 updatedAt，就只能依赖列表刷新
        // 假设 Author 接口定义中有 updatedAt 或 similar 字段 (BaseModel通常有)
        const cacheBuster = record.updatedAt ? `?t=${new Date(record.updatedAt).getTime()}` : '';
        return (
          <Image
            src={`${record.avatarUrl}${cacheBuster}`}
            width={40}
            height={40}
            style={{ borderRadius: '50%', objectFit: 'cover' }}
            fallback="/default_avatar.png"
          />
        );
      },
    },
    { title: '姓名', dataIndex: 'name' },
    {
      title: '朝代',
      dataIndex: 'dynastyId',
      valueType: 'select',
      request: requestDynasties,
      render: (_, record) => record.dynasty ? <Tag color="blue">{record.dynasty.name}</Tag> : '-',
    },
    { title: '简介', dataIndex: 'intro', ellipsis: true, search: false },
    {
      title: '操作',
      valueType: 'option',
      width: 180,
      render: (_, record) => [
        <ModalForm<Author>
          key={`edit-${record.ID}`} // 确保 key 唯一，每次打开都是新的
          title="编辑诗人"
          trigger={<a><EditOutlined /> 编辑</a>}
          onFinish={handleSubmit}
          initialValues={record}
          modalProps={{ destroyOnClose: true }} // 关闭销毁，防缓存
        >
          {renderFormFields(record)}
        </ModalForm>,
        <Popconfirm
          key="del"
          title="确定删除?"
          onConfirm={async () => {
            try { await deleteAuthor(record.ID); message.success('删除成功'); actionRef.current?.reload(); }
            catch(e) { /* Error caught by interceptor */ }
          }}
        >
          <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
        </Popconfirm>
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<Author>
        headerTitle="诗人库"
        actionRef={actionRef}
        rowKey="ID"
        // ✨ 解决问题2：添加工具栏按钮
        toolBarRender={() => [
          <ModalForm
            key="create"
            title="新建诗人"
            trigger={
              <Button type="primary">
                <PlusOutlined /> 新建
              </Button>
            }
            onFinish={handleSubmit}
            modalProps={{ destroyOnClose: true }}
          >
            {/* 新建模式不传 record */}
            {renderFormFields(undefined)}
          </ModalForm>
        ]}
        request={async (params) => {
          const res = await getAuthorList({ ...params, page: params.current, pageSize: params.pageSize });
          return { data: res.data.list, success: true, total: res.data.total };
        }}
        columns={columns}
      />
    </PageContainer>
  );
};

export default AuthorList;
