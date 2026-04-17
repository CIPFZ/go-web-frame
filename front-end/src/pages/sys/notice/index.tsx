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
import { createNotice, getNoticeList } from '@/services/api/notice';
import { getAuthorityList } from '@/services/api/authority';
import { getUserList } from '@/services/api/user';

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

const targetText: Record<string, string> = {
  all: '全体用户',
  roles: '按角色',
  users: '指定用户',
};

const flattenAuthorities = (nodes: any[] = []): { label: string; value: number }[] => {
  const result: { label: string; value: number }[] = [];
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
  const [createVisible, setCreateVisible] = useState(false);
  const [targetType, setTargetType] = useState<'all' | 'roles' | 'users'>('all');
  const [roleOptions, setRoleOptions] = useState<{ label: string; value: number }[]>([]);
  const [userOptions, setUserOptions] = useState<{ label: string; value: number }[]>([]);

  const columns: ProColumns<NoticeItem>[] = useMemo(
    () => [
      { title: 'ID', dataIndex: 'ID', width: 70, search: false },
      { title: '标题', dataIndex: 'title', ellipsis: true },
      {
        title: '级别',
        dataIndex: 'level',
        valueType: 'select',
        valueEnum: {
          info: { text: '信息' },
          warning: { text: '警告' },
          error: { text: '紧急' },
        },
        render: (_, row) => <Tag color={levelColor[row.level] || 'default'}>{row.level}</Tag>,
      },
      {
        title: '投递范围',
        dataIndex: 'targetType',
        valueType: 'select',
        valueEnum: {
          all: { text: '全体用户' },
          roles: { text: '按角色' },
          users: { text: '指定用户' },
        },
        render: (_, row) => targetText[row.targetType] || row.targetType,
      },
      { title: '接收人数', dataIndex: 'receiverCount', search: false, width: 100 },
      { title: '已读人数', dataIndex: 'readCount', search: false, width: 100 },
      { title: '创建时间', dataIndex: 'createdAt', valueType: 'dateTime', search: false, width: 180 },
    ],
    [],
  );

  const loadRoles = async () => {
    const res = await getAuthorityList({ page: 1, pageSize: 999 });
    if (res.code === 0) {
      setRoleOptions(flattenAuthorities(res.data?.list || res.data || []));
    }
  };

  const loadUsers = async () => {
    const res = await getUserList({ page: 1, pageSize: 200 });
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
        headerTitle="通知公告"
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
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={() => setCreateVisible(true)}>
            发布通知
          </Button>,
        ]}
      />

      <ModalForm
        title="发布通知"
        width={640}
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
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
            message.error(res.msg || '发布失败');
            return false;
          }
          message.success('发布成功');
          setCreateVisible(false);
          return true;
        }}
        initialValues={{ level: 'info', targetType: 'all', isPopup: false, needConfirm: false }}
      >
        <ProFormText name="title" label="标题" rules={[{ required: true }, { max: 128 }]} />
        <ProFormTextArea name="content" label="内容" rules={[{ required: true }]} fieldProps={{ rows: 4 }} />
        <ProFormSelect
          name="level"
          label="级别"
          valueEnum={{ info: '信息', warning: '警告', error: '紧急' }}
          rules={[{ required: true }]}
        />
        <ProFormSelect
          name="targetType"
          label="投递范围"
          rules={[{ required: true }]}
          valueEnum={{ all: '全体用户', roles: '按角色', users: '指定用户' }}
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
            label={targetType === 'roles' ? '角色' : '用户'}
            mode="multiple"
            rules={[{ required: true, message: '请至少选择一个目标' }]}
            options={targetType === 'roles' ? roleOptions : userOptions}
          />
        )}

        <ProFormDateTimePicker name="startAt" label="生效开始时间" />
        <ProFormDateTimePicker name="endAt" label="生效结束时间" />
        <ProFormSwitch name="isPopup" label="登录弹窗" />
        <ProFormSwitch name="needConfirm" label="需要手动确认已读" />
      </ModalForm>
    </PageContainer>
  );
};

export default NoticeAdminPage;
