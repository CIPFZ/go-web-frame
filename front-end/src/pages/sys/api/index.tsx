import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormSelect,
  ProFormTextArea,
  ProFormGroup,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Space, message, Popconfirm, Tag } from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';

// 导入 API
import {
  getApiList,
  createApi,
  updateApi,
  deleteApi,
} from '@/services/system/api';

// 定义数据类型
type ApiItem = {
  ID: number;
  path: string;
  description: string;
  apiGroup: string;
  method: string;
};
const ApiTableList: React.FC = () => {
  const localeFormRef1 = useFormLocale();
  useI18n();
  const actionRef = useRef<ActionType>(null);

  // 模态框状态
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<ApiItem>();
  // 选中的行 (用于批量删除)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // --- 操作函数 ---

  const handleAdd = () => {
    setCurrentRow(undefined);
    setModalVisible(true);
  };
  const handleEdit = (record: ApiItem) => {
    setCurrentRow(record);
    setModalVisible(true);
  };

  // 删除 (单条)
  const handleDelete = async (id: number) => {
    try {
      const res = await deleteApi({
        id,
      });
      if (res.code === 0) {
        message.success(t('cms.deletedSuccessfully'));
        actionRef.current?.reload();
      } else {
        message.error(res.msg || t('cms.deleteFailed'));
      }
    } catch (error) {
      message.error(t('cms.requestFailed'));
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (!selectedRowKeys.length) return;
    try {
      const res = await deleteApi({
        ids: selectedRowKeys as number[],
      });
      if (res.code === 0) {
        message.success(t('cms.selectedItemsDeleted'));
        setSelectedRowKeys([]); // 清空选中
        actionRef.current?.reload();
      } else {
        message.error(res.msg || t('cms.deleteFailed'));
      }
    } catch (error) {
      message.error(t('cms.requestFailed'));
    }
  };

  // 提交表单
  const handleFinish = async (values: any) => {
    const isUpdate = !!currentRow;
    const method = isUpdate ? updateApi : createApi;
    const data = {
      ...values,
      id: currentRow?.ID,
    };
    try {
      const res = await method(data);
      if (res.code === 0) {
        message.success(
          isUpdate ? t('cms.updatedSuccessfully') : t('cms.addedSuccessfully'),
        );
        setModalVisible(false);
        actionRef.current?.reload();
        return true;
      }
      message.error(res.msg || t('cms.operationFailed'));
      return false;
    } catch (error) {
      message.error(t('cms.requestFailed'));
      return false;
    }
  };

  // --- 列定义 ---
  const columns: ProColumns<ApiItem>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      width: 60,
      search: false,
      align: 'center',
    },
    {
      title: t('cms.apiPath'),
      dataIndex: 'path',
      width: 250,
      copyable: true,
      // 支持一键复制
      ellipsis: true,
    },
    {
      title: t('cms.apiGroup'),
      dataIndex: 'apiGroup',
      width: 120,
    },
    {
      title: t('cms.apiDescription'),
      dataIndex: 'description',
      ellipsis: true,
    },
    {
      title: t('cms.requestMethod'),
      dataIndex: 'method',
      width: 100,
      align: 'center',
      valueEnum: {
        POST: {
          text: 'POST',
          status: 'Processing',
        },
        // 蓝色
        GET: {
          text: 'GET',
          status: 'Success',
        },
        // 绿色
        PUT: {
          text: 'PUT',
          status: 'Warning',
        },
        // 橙色
        DELETE: {
          text: 'DELETE',
          status: 'Error',
        }, // 红色
      },
      // 自定义渲染 Tag
      render: (_, record) => {
        const colors: Record<string, string> = {
          POST: 'blue',
          GET: 'green',
          PUT: 'orange',
          DELETE: 'red',
        };
        return (
          <Tag color={colors[record.method] || 'default'}>{record.method}</Tag>
        );
      },
    },
    {
      title: t('cms.actions'),
      dataIndex: 'option',
      valueType: 'option',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <a key="edit" onClick={() => handleEdit(record)}>
            <EditOutlined />
            {t('cms.edit')}
          </a>
          <Popconfirm
            title={t('cms.deleteThisItem')}
            description={t('cms.relatedRolePermissionsWillAlsoBe')}
            onConfirm={() => handleDelete(record.ID)}
            okText={t('cms.yes')}
            cancelText={t('cms.no')}
          >
            <a
              key="delete"
              style={{
                color: '#ff4d4f',
              }}
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
      <ProTable<ApiItem>
        headerTitle={false}
        actionRef={actionRef}
        rowKey="ID"
        search={{
          labelWidth: 'auto',
        }}
        // 开启多选框
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys),
        }}
        request={async (params) => {
          // 转换分页参数并透传查询字段
          const res = await getApiList({
            page: params.current,
            pageSize: params.pageSize,
            path: params.path,
            description: params.description,
            apiGroup: params.apiGroup,
            method: params.method,
          });
          return {
            data: res.data?.list || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        }}
        columns={columns}
        scroll={{
          x: 900,
        }}
        // 工具栏
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={handleAdd}>
            <PlusOutlined />
            {t('cms.newApi')}
          </Button>,
          // 只有选中行时才显示批量删除按钮
          selectedRowKeys.length > 0 && (
            <Popconfirm
              key="batchDelete"
              title={t('cms.deleteOther', {
                value0: selectedRowKeys.length,
              })}
              onConfirm={handleBatchDelete}
              okText={t('cms.ok')}
              cancelText={t('cms.cancel')}
            >
              <Button danger>
                <DeleteOutlined />
                {t('cms.deleteSelected')}
              </Button>
            </Popconfirm>
          ),
        ]}
      />

      <ModalForm
        title={currentRow ? t('cms.editApi') : t('cms.newApi')}
        width="600px"
        open={modalVisible}
        onOpenChange={setModalVisible}
        onFinish={handleFinish}
        initialValues={currentRow}
        modalProps={{
          destroyOnClose: true,
        }}
        formRef={localeFormRef1}
      >
        <ProFormText
          name="path"
          label={t('cms.apiPath')}
          placeholder="e.g. /api/v1/user/info"
          rules={[
            {
              required: true,
              message: t('cms.enterAPath'),
            },
          ]}
        />

        {/* ✨ 2. 使用 ProFormGroup 替代 div，并调整宽度 */}
        <ProFormGroup>
          <ProFormSelect
            name="method"
            label={t('cms.requestMethod')}
            valueEnum={{
              POST: 'POST',
              GET: 'GET',
              PUT: 'PUT',
              DELETE: 'DELETE',
            }}
            placeholder={t('cms.selectAnOption')}
            // 修改为 xs (约104px)，对于 POST/GET 足够了，节省空间
            width="xs"
            rules={[
              {
                required: true,
                message: t('cms.selectAMethod'),
              },
            ]}
          />
          <ProFormText
            name="apiGroup"
            label={t('cms.apiGroup')}
            placeholder={t('cms.eGUserManagement')}
            // 保持 md (约328px)，因为前面的 xs 变小了，现在放得下了
            // xs(104) + md(328) + gap(16) = 448px < 容器宽度
            width="md"
            rules={[
              {
                required: true,
                message: t('cms.enterAGroup'),
              },
            ]}
          />
        </ProFormGroup>

        <ProFormTextArea
          name="description"
          label={t('cms.apiDescription')}
          placeholder={t('cms.describeWhatThisApiDoes')}
          rules={[
            {
              required: true,
              message: t('cms.enterADescription'),
            },
          ]}
        />
      </ModalForm>
    </PageContainer>
  );
};
export default ApiTableList;
