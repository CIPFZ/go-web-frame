import React, { useState, useEffect, useRef } from 'react'; // ✨ 引入 useEffect
import { GridContent } from '@ant-design/pro-components';
import { Menu, Typography, message, List } from 'antd';
import {
  ProCard,
  ProForm,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { createStyles } from 'antd-style';
import { MobileOutlined, MailOutlined } from '@ant-design/icons';
import UploadImage from '@/components/Upload/UploadImage';
// ✨ 引入 API
import { updateSelfInfo } from '@/services/api/user';

const { Title } = Typography;

const useStyles = createStyles(({ token }) => {
  return {
    avatarWrapper: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      paddingTop: 20,
    },
    uploadOverride: {
      '.ant-upload-select': {
        width: '144px !important',
        height: '144px !important',
        borderRadius: '50% !important',
        overflow: 'hidden',
        border: `1px dashed ${token.colorBorder} !important`,
        backgroundColor: token.colorFillAlter,
        '&:hover': {
          borderColor: `${token.colorPrimary} !important`,
        },
      },
      'img': {
        width: '100%',
        height: '100%',
        objectFit: 'cover',
        borderRadius: '50%',
      }
    },
    title: {
      marginBottom: 12,
      color: token.colorTextHeading,
      fontWeight: 500,
    },
    desc: {
      marginTop: 12,
      color: token.colorTextSecondary,
      fontSize: 12,
      textAlign: 'center',
    }
  };
});

// --- 子组件：基本设置 (BaseView) ---
const BaseView: React.FC<{ currentUser: any; refresh: () => void }> = ({ currentUser, refresh }) => {
  const { styles } = useStyles();

  const handleFinish = async (values: any) => {
    try {
      // ✨ 调用后端接口更新
      const res = await updateSelfInfo({
        nickName: values.nickName,
        bio: values.bio,
      });

      if (res.code === 0) {
        message.success('更新基本信息成功');
        // ✨ 刷新全局状态，让右上角名字和页面数据同步更新
        refresh();
      } else {
        message.error(res.msg || '更新失败');
      }
    } catch (error) {
      message.error('请求失败，请稍后重试');
    }
  };

  return (
    <div style={{ display: 'flex', gap: '48px', flexDirection: 'row' }}>
      {/* 左侧表单 */}
      <div style={{ flex: 1, maxWidth: '440px' }}>
        <ProForm
          layout="vertical"
          onFinish={handleFinish}
          submitter={{
            searchConfig: {
              submitText: '更新基本信息',
            },
            render: (_, dom) => dom[1],
          }}
          // ✨ 使用 request 来回填数据，或者直接用 initialValues (这里用 initialValues 更简单，因为 currentUser 已经传进来了)
          initialValues={{
            nickName: currentUser?.nickName,
            phone: currentUser?.phone,
            email: currentUser?.email,
            bio: currentUser?.bio, // ✨ 绑定后端返回的 bio
          }}
          // 关键：当 currentUser 变化时重置表单，防止刷新后数据没变
          key={currentUser?.id}
        >
          <ProFormText
            width="md"
            name="nickName"
            label="昵称"
            rules={[{ required: true, message: '请输入您的昵称!' }]}
          />

          <ProFormTextArea
            name="bio"
            label="个人简介"
            placeholder="介绍一下自己..."
            fieldProps={{
              rows: 4,
              showCount: true,
              maxLength: 200,
            }}
          />

          <ProFormText
            width="md"
            name="phone"
            label="手机号码"
            disabled
            fieldProps={{
              prefix: <MobileOutlined style={{ color: '#999' }} />,
            }}
            extra="如需修改手机号，请前往 [安全设置]"
          />

          <ProFormText
            width="md"
            name="email"
            label="邮箱"
            disabled
            fieldProps={{
              prefix: <MailOutlined style={{ color: '#999' }} />,
            }}
            extra="如需修改邮箱，请前往 [安全设置]"
          />

        </ProForm>
      </div>

      {/* 右侧头像 */}
      <div className={styles.avatarWrapper}>
        <div className={styles.title}>头像</div>
        <div className={styles.uploadOverride}>
          {/* ✨ UploadImage 组件内部已经处理了上传逻辑，成功后会调用 onChange */}
          {/* 这里我们需要在 onChange 时不仅打印 Log，还要通知后端更新 avatar 字段 */}
          {/* 但实际上你的后端 upload 接口应该只负责上传，不更新 user 表的 avatar */}
          {/* 这里的最佳实践是：UploadImage 上传成功拿到 url -> 自动触发一个 updateUserInfo 接口更新 avatar */}
          {/* 或者让 UploadImage 只做上传，拿到 url 后，用户点击“更新基本信息”时再一起提交？*/}
          {/* 通常头像修改是即时的。我们可以复用 updateSelfInfo 或者专门写一个 updateAvatar */}

          {/* 方案：这里我们简单点，UploadImage 上传成功后，我们认为头像修改完成，刷新全局状态 */}
          {/* *注意*：你之前的 FileApi 上传后并没有更新 User 表的 Avatar 字段 */}
          {/* 你需要确认你的 UploadImage 组件是只返回 URL，还是顺便更新了用户表 */}
          {/* 如果只返回 URL，你需要在这里调用 updateSelfInfo({ ...currentUser, avatar: url }) */}

          <UploadImage
            action="/api/v1/sys/user/avatar"
            value={currentUser?.avatar}
            onChange={async (url) => {
              // 如果后端上传接口没有自动更新用户表的头像字段，需要在这里手动更新
              // 假设我们复用 updateSelfInfo，虽然它只接收 nickName/bio，你需要去扩展 DTO
              // 或者更简单的：让后端 /file/upload 成功后，如果检测到是用户头像上传，顺便更新用户表(不推荐，耦合太重)
              // 刷新页面状态
              refresh();
            }}
          />
        </div>
        <div className={styles.desc}>
          支持 jpg, png, gif 格式<br/>大小不超过 2MB
        </div>
      </div>
    </div>
  );
};

// --- 子组件：安全设置 (SecurityView) ---
const SecurityView: React.FC<{ currentUser: any }> = ({ currentUser }) => {
  const handleModify = (type: string) => {
    message.info(`即将打开 [${type}] 修改弹窗 (需验证码)`);
  };

  const data = [
    {
      title: '账户密码',
      description: '当前密码强度：强',
      actions: [<a key="Modify" onClick={() => handleModify('密码')}>修改</a>],
    },
    {
      title: '密保手机',
      description: `已绑定手机：${currentUser?.phone || '未绑定'}`,
      actions: [<a key="Modify" onClick={() => handleModify('手机')}>修改</a>],
    },
    {
      title: '密保邮箱',
      description: `已绑定邮箱：${currentUser?.email || '未绑定'}`,
      actions: [<a key="Modify" onClick={() => handleModify('邮箱')}>修改</a>],
    },
    {
      title: 'MFA 设备',
      description: '未绑定 MFA 设备，绑定后，可以进行二次确认',
      actions: [<a key="bind" onClick={() => handleModify('MFA')}>绑定</a>],
    },
  ];

  return (
    <List
      itemLayout="horizontal"
      dataSource={data}
      renderItem={(item) => (
        <List.Item actions={item.actions}>
          <List.Item.Meta title={item.title} description={item.description} />
        </List.Item>
      )}
    />
  );
};

// --- 主页面 ---
const Settings: React.FC = () => {
  const { initialState, setInitialState, refresh } = useModel('@@initialState'); // ✨ 获取 refresh 方法
  const currentUser = initialState?.currentUser;

  const [initConfig, setInitConfig] = useState<'base' | 'security'>('base');

  const menuMap: Record<'base' | 'security', string> = {
    base: '基本设置',
    security: '安全设置',
  };

  const renderChildren = () => {
    switch (initConfig) {
      case 'base':
        // ✨ 将 refresh 传给子组件，以便修改完资料后刷新全局状态
        return <BaseView currentUser={currentUser} refresh={refresh} />;
      case 'security':
        return <SecurityView currentUser={currentUser} />;
      default:
        return null;
    }
  };

  return (
    <GridContent>
      <ProCard
        style={{ height: '100%', minHeight: 600 }}
        bodyStyle={{ padding: 0, display: 'flex', flexDirection: 'row' }}
        bordered
      >
        <div style={{ width: 224, borderRight: '1px solid #f0f0f0', padding: '16px 0' }}>
          <Menu
            mode="inline"
            selectedKeys={[initConfig]}
            onClick={({ key }) => setInitConfig(key as 'base' | 'security')}
            style={{ border: 'none' }}
          >
            {(Object.keys(menuMap) as Array<'base' | 'security'>).map((item) => (
              <Menu.Item key={item}>{menuMap[item]}</Menu.Item>
            ))}
          </Menu>
        </div>

        <div style={{ flex: 1, padding: '24px 40px' }}>
          <Title level={4} style={{ marginBottom: 24 }}>
            {menuMap[initConfig]}
          </Title>
          {renderChildren()}
        </div>
      </ProCard>
    </GridContent>
  );
};

export default Settings;
