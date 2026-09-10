import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
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
  FolderOpenOutlined,
} from '@ant-design/icons';
// ✨ 导入全局 Model
import { useModel } from '@umijs/max';
import {
  getMenuList,
  addBaseMenu,
  updateBaseMenu,
  deleteBaseMenu,
} from '@/services/system/menu';
import { getIcon } from '@/utils/iconMap';
// ✨ 导入我们刚写的组件
import IconPicker from '@/components/IconPicker';
import { clearMenuCache } from '@/routing/menuDataStore';
type MenuItem = {
  ID: number;
  parentId: number;
  path: string;
  name: string;
  nameEn?: string;
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
  const localeFormRef1 = useFormLocale();
  useI18n();
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
        message.success(
          isUpdate ? t('cms.updatedSuccessfully') : t('cms.addedSuccessfully'),
        );
        setModalVisible(false);
        actionRef.current?.reload();
        // ✨ 关键：刷新左侧全局菜单
        clearMenuCache();
        await refresh();
        console.log('update refresh success');
        return true;
      }
      message.error(res.msg || t('cms.operationFailed'));
      return false;
    } catch (error) {
      message.error(t('cms.requestFailed'));
      return false;
    }
  };
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteBaseMenu({
        id,
      });
      if (res.code === 0) {
        message.success(t('cms.deletedSuccessfully'));
        actionRef.current?.reload();
        // ✨ 关键：刷新左侧全局菜单
        clearMenuCache();
        await refresh();
      } else {
        message.error(res.msg || t('cms.deleteFailed'));
      }
    } catch (error) {
      message.error(t('cms.requestFailed'));
    }
  };
  const columns: ProColumns<MenuItem>[] = [
    {
      title: t('cms.displayName'),
      dataIndex: 'name',
      width: 200,
      fixed: 'left',
      search: false,
      // 给父节点加个文件夹图标，子节点加文件图标，更好看
      render: (text, record) => (
        <Space>
          {record.routes ? (
            <FolderOpenOutlined
              style={{
                color: '#faad14',
              }}
            />
          ) : (
            <FileOutlined
              style={{
                color: '#1890ff',
              }}
            />
          )}
          {text}
        </Space>
      ),
    },
    {
      title: t('cms.englishDisplayName'),
      dataIndex: 'nameEn',
      search: false,
      width: 180,
    },
    {
      title: t('cms.icon'),
      dataIndex: 'icon',
      width: 80,
      align: 'center',
      search: false,
      render: (text) => (
        <div
          style={{
            fontSize: 18,
            color: '#595959',
          }}
        >
          {getIcon(text as string)}
        </div>
      ),
    },
    {
      title: t('cms.routePath'),
      dataIndex: 'path',
      width: 280,
      copyable: true,
      ellipsis: true,
      search: false,
    },
    {
      title: t('cms.componentPath'),
      dataIndex: 'component',
      width: 220,
      ellipsis: true,
      search: false,
    },
    {
      title: t('cms.sortOrder'),
      dataIndex: 'sort',
      width: 80,
      align: 'center',
      search: false,
    },
    {
      title: t('cms.status'),
      // 改名，更直观
      dataIndex: 'hideInMenu',
      width: 100,
      align: 'center',
      search: false,
      // ✨ 优化：使用 Tag 渲染
      render: (_, record) =>
        record.hideInMenu ? (
          <Tag color="default">{t('cms.hidden')}</Tag>
        ) : (
          <Tag color="success">{t('cms.visible')}</Tag>
        ),
    },
    {
      title: t('cms.actions'),
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      fixed: 'right',
      // ✨ 优化：添加图标，平铺显示
      render: (_, record) => (
        <Space size="small">
          <a
            key="edit"
            onClick={() => handleEdit(record)}
            title={t('cms.edit')}
          >
            <EditOutlined />
            {t('cms.edit')}
          </a>
          <a
            key="add"
            onClick={() => handleAddChild(record)}
            title={t('cms.addSubmenu')}
          >
            <PlusOutlined />
            {t('cms.submenu')}
          </a>
          <Popconfirm
            title={t('cms.deleteThisItem')}
            description={t('cms.thisActionCannotBeUndone')}
            onConfirm={() => handleDelete(record.ID)}
            okText={t('cms.yes')}
            cancelText={t('cms.no')}
          >
            <a
              key="delete"
              style={{
                color: '#ff4d4f',
              }}
              title={t('cms.delete')}
            >
              <DeleteOutlined />
              {t('cms.delete')}
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
          const res = await getMenuList({
            pageInfo: {
              page: 1,
              pageSize: 999,
            },
          });
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
            <PlusOutlined />
            {t('cms.newRootMenu')}
          </Button>,
        ]}
        scroll={{
          x: 1600,
        }}
      />

      <ModalForm
        title={currentRow ? t('cms.editMenu') : t('cms.newMenu')}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        modalProps={{
          destroyOnClose: true,
        }}
        initialValues={currentRow}
        formRef={localeFormRef1}
      >
        <ProFormText
          name="name"
          label={t('cms.displayName')}
          placeholder={t('cms.eGWorkplace')}
          rules={[
            {
              required: true,
            },
          ]}
        />

        <ProFormText
          name="nameEn"
          label={t('cms.englishDisplayName')}
          placeholder="e.g. Workplace"
        />
        <ProFormText
          name="path"
          label={t('cms.routePath')}
          rules={[
            {
              required: true,
            },
          ]}
        />

        <ProFormText
          name="component"
          label={t('cms.componentPath')}
          rules={[
            {
              required: true,
            },
          ]}
        />

        <div
          style={{
            display: 'flex',
            gap: 16,
          }}
        >
          {/* ✨ 关键：使用自定义的 IconPicker */}
          <Form.Item
            name="icon"
            label={t('cms.icon')}
            style={{
              flex: 1,
            }}
          >
            <IconPicker />
          </Form.Item>

          <ProFormDigit
            name="sort"
            label={t('cms.sortOrder')}
            tooltip={t('cms.siblingMenusAreSortedInAscending')}
            width="xs"
            initialValue={0}
            fieldProps={{
              precision: 0,
            }}
          />
        </div>

        <div
          style={{
            display: 'flex',
            gap: 16,
          }}
        >
          <ProFormText name="access" label={t('cms.accessKey')} width="md" />
          <ProFormSwitch
            name="hideInMenu"
            label={t('cms.hideInNavigation')}
            initialValue={false}
          />
        </div>

        {/* 其他字段保持不变 */}
        <div
          style={{
            display: 'flex',
            gap: 16,
          }}
        >
          <ProFormText name="target" label={t('cms.linkTarget')} width="sm" />
          <ProFormText
            name="locale"
            label={t('cms.translationKey')}
            width="md"
            tooltip={t('cms.optionalBuiltInTranslationKeyStored')}
          />
        </div>
      </ModalForm>
    </PageContainer>
  );
};
export default MenuTableList;
