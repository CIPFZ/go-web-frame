import {
  LogoutOutlined,
  SettingOutlined,
  UserOutlined,
  CheckOutlined,
  SwapOutlined,
} from '@ant-design/icons';
import { history, useModel } from '@umijs/max';
import type { MenuProps } from 'antd';
import { Spin, message } from 'antd'; // Menu 组件不需要显式导入，HeaderDropdown 会处理
import { createStyles } from 'antd-style';
import React from 'react';
import { flushSync } from 'react-dom';
// 修正导入路径，确保你的 api 定义正确
import { outLogin, switchAuthority } from '@/services/api/user';
import HeaderDropdown from '../HeaderDropdown';

export type GlobalHeaderRightProps = {
  menu?: boolean;
  children?: React.ReactNode;
};

export const AvatarName = () => {
  const { initialState } = useModel('@@initialState');
  const { currentUser } = initialState || {};
  return <span className="anticon">{currentUser?.nickName}</span>;
};

const useStyles = createStyles(({ token }) => {
  return {
    action: {
      display: 'flex',
      height: '48px',
      marginLeft: 'auto',
      overflow: 'hidden',
      alignItems: 'center',
      padding: '0 8px',
      cursor: 'pointer',
      borderRadius: token.borderRadius,
      '&:hover': {
        backgroundColor: token.colorBgTextHover,
      },
    },
  };
});

export const AvatarDropdown: React.FC<GlobalHeaderRightProps> = ({
                                                                   menu,
                                                                   children,
                                                                 }) => {
  const { styles } = useStyles();
  const { initialState, setInitialState } = useModel('@@initialState');

  /**
   * 退出登录逻辑
   */
  const loginOut = async () => {
    try {
      await outLogin();
    } catch (error) {
      console.error("Logout failed:", error);
    }
    localStorage.removeItem('token');

    const { search, pathname } = window.location;
    const urlParams = new URL(window.location.href).searchParams;
    const redirect = urlParams.get('redirect');

    if (window.location.pathname !== '/user/login' && !redirect) {
      history.replace({
        pathname: '/user/login',
        search: stringify({ redirect: pathname + search }), // 需导入 stringify 或手动拼接
      });
    }
  };

  /**
   * 切换角色逻辑
   */
  const handleSwitchAuthority = async (authorityId: number) => {
    try {
      const res = await switchAuthority({ authorityId });
      if (res.code === 0) {
        message.success('切换角色成功，正在刷新...');
        // 更新本地 Token
        localStorage.setItem('token', res.data.token);
        // 强制刷新页面，让应用重新加载
        window.location.href = '/';
      }
    } catch (error) {
      message.error('切换失败');
    }
  };

  /**
   * ✨ 统一的菜单点击处理
   */
  const onMenuClick: MenuProps['onClick'] = (event) => {
    const { key } = event;

    // 1. 退出登录
    if (key === 'logout') {
      flushSync(() => {
        setInitialState((s) => ({ ...s, currentUser: undefined }));
      });
      loginOut();
      return;
    }

    // 2. ✨ 切换角色 (识别 key 前缀)
    if (key.startsWith('role:')) {
      const authId = Number(key.split(':')[1]);
      handleSwitchAuthority(authId);
      return;
    }

    // 3. 路由跳转 (个人中心/设置)
    if (key === 'settings') {
      history.push(`/account/${key}`);
      return;
    }
  };

  const loading = (
    <span className={styles.action}>
      <Spin size="small" style={{ marginLeft: 8, marginRight: 8 }} />
    </span>
  );

  if (!initialState || !initialState.currentUser || !initialState.currentUser.nickName) {
    return loading;
  }

  const { currentUser } = initialState;

  // ✨ 构建角色菜单项
  const roleMenuItems: MenuProps['items'] = currentUser.authorities?.map((auth: any) => {
    const isCurrent = auth.authorityId === currentUser.authorityId;
    return {
      key: `role:${auth.authorityId}`, // 使用前缀区分
      // 当前角色显示打钩，其他显示占位符保持对齐
      icon: isCurrent ? <CheckOutlined /> : <div style={{ width: 14, display: 'inline-block' }} />,
      label: auth.authorityName,
      // ✨ 关键修改：使用 disabled 属性
      disabled: isCurrent,
    };
  }) || [];

  // 组装所有菜单
  const menuItems: MenuProps['items'] = [
    { key: 'settings', icon: <UserOutlined />, label: '账号设置' },
    {
      type: 'divider' as const, // ✨ 修复点 2
    },

    // 角色切换标题
    {
      key: 'switch-role-title',
      label: <span style={{ color: '#999', fontSize: '12px', cursor: 'default' }}><SwapOutlined /> 切换角色</span>,
      disabled: true,
    },

    ...roleMenuItems,

    {
      type: 'divider' as const, // ✨ 修复点 2
    },

    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
  ];

  return (
    <HeaderDropdown
      menu={{
        selectedKeys: [], // 不选中任何项
        onClick: onMenuClick,
        items: menuItems,
      }}
    >
      {children}
    </HeaderDropdown>
  );
};

// 补充 stringify 简单的实现，或者 import { stringify } from 'querystring'
function stringify(obj: Record<string, string>) {
  return new URLSearchParams(obj).toString();
}
