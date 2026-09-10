import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
import { loadPagedOptions } from '@/utils/pagedOptions';
import React, { useMemo, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import {
  ModalForm,
  ProFormDateTimePicker,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { Button, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { createNotice, getNoticeList } from '@/services/system/notice';
import { getAuthorityList } from '@/services/system/authority';
import { getUserList } from '@/services/system/user';
type NoticeItem = {
  ID: number;
  title: string;
  content: string;
  level: 'info' | 'warning' | 'error';
  targetType: 'all' | 'roles' | 'users';
  isPopup: boolean;
  needConfirm: boolean;
  startAt?: string;
  endAt?: string;
  receiverCount: number;
  readCount: number;
  createdAt: string;
};
const levelColor: Record<string, string> = {
  info: 'blue',
  warning: 'orange',
  error: 'red',
};
const flattenAuthorities = (
  nodes: any[] = [],
): {
  label: string;
  value: number;
}[] => {
  const result: {
    label: string;
    value: number;
  }[] = [];
  const walk = (items: any[]) => {
    items.forEach((item) => {
      result.push({
        label: item.authorityName || String(item.authorityId),
        value: item.authorityId,
      });
      if (Array.isArray(item.children) && item.children.length > 0) {
        walk(item.children);
      }
    });
  };
  walk(nodes);
  return result;
};
const NoticeAdminPage: React.FC = () => {
  const localeFormRef1 = useFormLocale();
  const locale = useI18n();
  const targetText: Record<string, string> = {
    all: t('cms.allUsers'),
    roles: t('cms.byRole'),
    users: t('cms.selectedUsers'),
  };
  const [createVisible, setCreateVisible] = useState(false);
  const [targetType, setTargetType] = useState<'all' | 'roles' | 'users'>(
    'all',
  );
  const [roleOptions, setRoleOptions] = useState<
    {
      label: string;
      value: number;
    }[]
  >([]);
  const [userOptions, setUserOptions] = useState<
    {
      label: string;
      value: number;
    }[]
  >([]);
  const columns: ProColumns<NoticeItem>[] = useMemo(
    () => [
      {
        title: 'ID',
        dataIndex: 'ID',
        width: 70,
        search: false,
      },
      {
        title: t('cms.title'),
        dataIndex: 'title',
        ellipsis: true,
      },
      {
        title: t('cms.level'),
        dataIndex: 'level',
        valueType: 'select',
        valueEnum: {
          info: {
            text: t('cms.info'),
          },
          warning: {
            text: t('cms.warning'),
          },
          error: {
            text: t('cms.urgent'),
          },
        },
        render: (_, row) => (
          <Tag color={levelColor[row.level] || 'default'}>
            {
              {
                info: t('cms.info'),
                warning: t('cms.warning'),
                error: t('cms.urgent'),
              }[row.level]
            }
          </Tag>
        ),
      },
      {
        title: t('cms.recipients'),
        dataIndex: 'targetType',
        valueType: 'select',
        valueEnum: {
          all: {
            text: t('cms.allUsers'),
          },
          roles: {
            text: t('cms.byRole'),
          },
          users: {
            text: t('cms.selectedUsers'),
          },
        },
        render: (_, row) => targetText[row.targetType] || row.targetType,
      },
      {
        title: t('cms.recipients.1e1bd3'),
        dataIndex: 'receiverCount',
        search: false,
        width: 100,
      },
      {
        title: t('cms.readBy'),
        dataIndex: 'readCount',
        search: false,
        width: 100,
      },
      {
        title: t('cms.createdAt'),
        dataIndex: 'createdAt',
        valueType: 'dateTime',
        search: false,
        width: 180,
      },
    ],
    [locale],
  );
  const loadRoles = async () => {
    const res = await getAuthorityList({
      page: 1,
      pageSize: 999,
    });
    if (res.code === 0) {
      setRoleOptions(flattenAuthorities(res.data?.list || res.data || []));
    }
  };
  const loadUsers = async () => {
    const res = await loadPagedOptions(getUserList);
    if (res.code === 0) {
      const options = (res.data?.list || []).map((u: any) => ({
        label: `${u.nickName || u.username} (${u.username})`,
        value: u.ID,
      }));
      setUserOptions(options);
    }
  };
  return (
    <PageContainer title={false}>
      <ProTable<NoticeItem>
        rowKey="ID"
        headerTitle={t('cms.notices.e3d7bf')}
        columns={columns}
        request={async (params) => {
          const res = await getNoticeList({
            page: params.current,
            pageSize: params.pageSize,
            title: params.title,
            level: params.level,
          });
          return {
            success: res.code === 0,
            data: res.data?.list || [],
            total: res.data?.total || 0,
          };
        }}
        toolBarRender={() => [
          <Button
            key="new"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateVisible(true)}
          >
            {t('cms.publishNotice')}
          </Button>,
        ]}
      />

      <ModalForm
        title={t('cms.publishNotice')}
        width={640}
        open={createVisible}
        modalProps={{
          destroyOnClose: true,
          onCancel: () => setCreateVisible(false),
        }}
        onFinish={async (values) => {
          const payload = {
            title: values.title,
            content: values.content,
            level: values.level,
            targetType: values.targetType,
            targetIds: values.targetIds || [],
            isPopup: values.isPopup || false,
            needConfirm: values.needConfirm || false,
            startAt: values.startAt || undefined,
            endAt: values.endAt || undefined,
          };
          const res = await createNotice(payload);
          if (res.code !== 0) {
            message.error(res.msg || t('cms.publishFailed'));
            return false;
          }
          message.success(t('cms.publishedSuccessfully'));
          setCreateVisible(false);
          return true;
        }}
        initialValues={{
          level: 'info',
          targetType: 'all',
          isPopup: false,
          needConfirm: false,
        }}
        formRef={localeFormRef1}
      >
        <ProFormText
          name="title"
          label={t('cms.title')}
          rules={[
            {
              required: true,
            },
            {
              max: 128,
            },
          ]}
        />
        <ProFormTextArea
          name="content"
          label={t('cms.content')}
          rules={[
            {
              required: true,
            },
          ]}
          fieldProps={{
            rows: 4,
          }}
        />
        <ProFormSelect
          name="level"
          label={t('cms.level')}
          valueEnum={{
            info: t('cms.info'),
            warning: t('cms.warning'),
            error: t('cms.urgent'),
          }}
          rules={[
            {
              required: true,
            },
          ]}
        />
        <ProFormSelect
          name="targetType"
          label={t('cms.recipients')}
          rules={[
            {
              required: true,
            },
          ]}
          valueEnum={{
            all: t('cms.allUsers'),
            roles: t('cms.byRole'),
            users: t('cms.selectedUsers'),
          }}
          fieldProps={{
            onChange: async (v) => {
              const nextTargetType = v as 'all' | 'roles' | 'users';
              setTargetType(nextTargetType);
              if (nextTargetType === 'roles' && roleOptions.length === 0) {
                await loadRoles();
              }
              if (nextTargetType === 'users' && userOptions.length === 0) {
                await loadUsers();
              }
            },
          }}
        />

        {targetType !== 'all' && (
          <ProFormSelect
            name="targetIds"
            label={targetType === 'roles' ? t('cms.role') : t('cms.user')}
            mode="multiple"
            rules={[
              {
                required: true,
                message: t('cms.selectAtLeastOneRecipient'),
              },
            ]}
            options={targetType === 'roles' ? roleOptions : userOptions}
          />
        )}

        <ProFormDateTimePicker name="startAt" label={t('cms.activeFrom')} />
        <ProFormDateTimePicker name="endAt" label={t('cms.activeUntil')} />
        <ProFormSwitch name="isPopup" label={t('cms.showOnLogin')} />
        <ProFormSwitch
          name="needConfirm"
          label={t('cms.requireReadConfirmation')}
        />
      </ModalForm>
    </PageContainer>
  );
};
export default NoticeAdminPage;
