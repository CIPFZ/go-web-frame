import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormDigit,
  ProFormSwitch,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Space, message, Popconfirm, Tag, Form } from 'antd';
// ✨ 导入图标
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FileOutlined,
  FolderOpenOutlined
} from '@ant-design/icons';
// ✨ 导入全局 Model
import { useModel } from '@umijs/max';

import { getMenuList, addBaseMenu, updateBaseMenu, deleteBaseMenu } from '@/services/api/menu';
import { getIcon } from '@/utils/iconMap';
// ✨ 导入我们刚写的组件
import IconPicker from '@/components/IconPicker';
import { clearMenuCache } from '@/utils/menuDataStore';

type MenuItem = {
  ID: number;
  parentId: number;
  path: string;
  name: string;
  component: string;
  sort: number;
  icon: string;
  hideInMenu: boolean;
  access: string;
  target: string;
  locale: string;
  routes?: MenuItem[];
};

const MenuTableList: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  // ✨ 获取全局 initialState 的刷新方法
  const { refresh } = useModel('@@initialState');

  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<MenuItem>();
  const [parentId, setParentId] = useState<number>(0);

  const handleAddRoot = () => {
    setCurrentRow(undefined);
    setParentId(0);
    setModalVisible(true);
  };

  const handleAddChild = (record: MenuItem) => {
    setCurrentRow(undefined);
    setParentId(record.ID);
    setModalVisible(true);
  };

  const handleEdit = (record: MenuItem) => {
    setCurrentRow(record);
    setParentId(record.parentId);
    setModalVisible(true);
  };

  const handleFinish = async (values: any) => {
    const isUpdate = !!currentRow;
    const method = isUpdate ? updateBaseMenu : addBaseMenu;

    const data = {
      ...values,
      id: currentRow?.ID,
      parentId: parentId,
      hideInMenu: values.hideInMenu,
    };

    try {
      const res = await method(data);
      if (res.code === 0) {
        message.success(isUpdate ? '更新成功' : '添加成功');
        setModalVisible(false);
        actionRef.current?.reload();
        // ✨ 关键：刷新左侧全局菜单
        clearMenuCache();
        await refresh();
        console.log("update refresh success");
        return true;
      }
      message.error(res.msg || '操作失败');
      return false;
    } catch (error) {
      message.error('请求出错');
      return false;
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const res = await deleteBaseMenu({ id });
      if (res.code === 0) {
        message.success('删除成功');
        actionRef.current?.reload();
        // ✨ 关键：刷新左侧全局菜单
        clearMenuCache();
        await refresh();
      } else {
        message.error(res.msg || '删除失败');
      }
    } catch (error) {
      message.error('请求出错');
    }
  };

  const columns: ProColumns<MenuItem>[] = [
    {
      title: '展示名称',
      dataIndex: 'name',
      width: 200,
      fixed: 'left',
      search: false,
      // 给父节点加个文件夹图标，子节点加文件图标，更好看
      render: (text, record) => (
        <Space>
          {record.routes ? <FolderOpenOutlined style={{color:'#faad14'}} /> : <FileOutlined style={{color:'#1890ff'}} />}
          {text}
        </Space>
      )
    },
    {
      title: '图标',
      dataIndex: 'icon',
      width: 80,
      align: 'center',
      search: false,
      render: (text) => (
        <div style={{ fontSize: 18, color: '#595959' }}>
          {getIcon(text as string)}
        </div>
      ),
    },
    {
      title: '路由路径',
      dataIndex: 'path',
      copyable: true,
      ellipsis: true,
      search: false,
    },
    {
      title: '组件路径',
      dataIndex: 'component',
      ellipsis: true,
      search: false,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
      align: 'center',
      search: false,
    },
    {
      title: '状态', // 改名，更直观
      dataIndex: 'hideInMenu',
      width: 100,
      align: 'center',
      search: false,
      // ✨ 优化：使用 Tag 渲染
      render: (_, record) => (
        record.hideInMenu ?
          <Tag color="default">隐藏</Tag> :
          <Tag color="success">显示</Tag>
      ),
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      fixed: 'right',
      // ✨ 优化：添加图标，平铺显示
      render: (_, record) => (
        <Space size="small">
          <a key="edit" onClick={() => handleEdit(record)} title="编辑">
            <EditOutlined /> 编辑
          </a>
          <a key="add" onClick={() => handleAddChild(record)} title="添加子菜单">
            <PlusOutlined /> 子菜单
          </a>
          <Popconfirm
            title="确定删除?"
            description="删除后无法恢复"
            onConfirm={() => handleDelete(record.ID)}
            okText="是"
            cancelText="否"
          >
            <a key="delete" style={{ color: '#ff4d4f' }} title="删除">
              <DeleteOutlined /> 删除
            </a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<MenuItem>
        headerTitle={false}
        actionRef={actionRef}
        rowKey="ID"
        search={false}
        pagination={false}
        childrenColumnName="routes"

        request={async (params) => {
          const res = await getMenuList({ pageInfo: { page: 1, pageSize: 999 } });
          return {
            // ✨ 修正：直接使用 res.data，因为截图显示它就是一个数组
            data: res.data || [],
            success: res.code === 0,
            // 如果是全量树形数据，total 甚至可以不传，或者传数组长度
            total: Array.isArray(res.data) ? res.data.length : 0,
          };
        }}
        columns={columns}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={handleAddRoot}>
            <PlusOutlined /> 新建根菜单
          </Button>,
        ]}
        scroll={{ x: 1200 }}
      />

      <ModalForm
        title={currentRow ? '编辑菜单' : '新建菜单'}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        modalProps={{ destroyOnClose: true }}
        initialValues={currentRow}
      >
        <ProFormText
          name="name"
          label="展示名称"
          rules={[{ required: true }]}
        />

        <ProFormText
          name="path"
          label="路由路径"
          rules={[{ required: true }]}
        />

        <ProFormText
          name="component"
          label="组件路径"
          rules={[{ required: true }]}
        />

        <div style={{ display: 'flex', gap: 16 }}>
          {/* ✨ 关键：使用自定义的 IconPicker */}
          <Form.Item name="icon" label="图标" style={{flex: 1}}>
            <IconPicker />
          </Form.Item>

          <ProFormDigit
            name="sort"
            label="排序"
            width="xs"
            initialValue={0}
            fieldProps={{ precision: 0 }}
          />
        </div>

        <div style={{ display: 'flex', gap: 16 }}>
          <ProFormText
            name="access"
            label="权限标识"
            width="md"
          />
          <ProFormSwitch
            name="hideInMenu"
            label="在菜单中隐藏"
            initialValue={false}
          />
        </div>

        {/* 其他字段保持不变 */}
        <div style={{ display: 'flex', gap: 16 }}>
          <ProFormText name="target" label="跳转目标" width="sm" />
          <ProFormText name="locale" label="国际化 Key" width="md" />
        </div>

      </ModalForm>
    </PageContainer>
  );
};

export default MenuTableList;
