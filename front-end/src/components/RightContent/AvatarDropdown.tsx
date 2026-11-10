import { t, useI18n } from '@/i18n';
import {
  LogoutOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { history, useModel } from '@umijs/max';
import type { MenuProps } from 'antd';
import { Spin } from 'antd'; // Menu 组件不需要显式导入，HeaderDropdown 会处理
import { createStyles } from 'antd-style';
import React from 'react';
import { flushSync } from 'react-dom';
// 修正导入路径，确保你的 api 定义正确
import { outLogin } from '@/services/system/user';
import HeaderDropdown from '../HeaderDropdown';
export type GlobalHeaderRightProps = {
  menu?: boolean;
  children?: React.ReactNode;
};
export const AvatarName = () => {
  useI18n();
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
  useI18n();
  const { styles } = useStyles();
  const { initialState, setInitialState } = useModel('@@initialState');

  /**
   * 退出登录逻辑
   */
  const loginOut = async () => {
    try {
      await outLogin();
    } catch (error) {
      console.error('Logout failed:', error);
    }
    localStorage.removeItem('token');
    const { search, pathname } = window.location;
    const urlParams = new URL(window.location.href).searchParams;
    const redirect = urlParams.get('redirect');
    if (window.location.pathname !== '/user/login' && !redirect) {
      history.replace({
        pathname: '/user/login',
        search: stringify({
          redirect: pathname + search,
        }), // 需导入 stringify 或手动拼接
      });
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
        setInitialState((s) => ({
          ...s,
          currentUser: undefined,
        }));
      });
      loginOut();
      return;
    }

    // 3. 路由跳转 (个人中心/设置)
    if (key === 'settings') {
      if (settingsPath) history.push(settingsPath);
      return;
    }
  };
  const loading = (
    <span className={styles.action}>
      <Spin
        size="small"
        style={{
          marginLeft: 8,
          marginRight: 8,
        }}
      />
    </span>
  );
  if (
    !initialState ||
    !initialState.currentUser ||
    !initialState.currentUser.nickName
  ) {
    return loading;
  }

  const findSettings = (menus: any[]): string | undefined => {
    for (const item of menus) {
      if (item.component === 'user/info') return item.path;
      const child = findSettings(item.routes || []);
      if (child) return child;
    }
    return undefined;
  };
  const settingsPath = findSettings(initialState.menuData || []);

  // 组装所有菜单
  const menuItems: MenuProps['items'] = [
    {
      key: 'settings',
      disabled: !settingsPath,
      icon: <UserOutlined />,
      label: t('cms.accountSettings'),
    },
    {
      type: 'divider' as const, // ✨ 修复点 2
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('cms.signOut'),
    },
  ];
  return (
    <HeaderDropdown
      menu={{
        selectedKeys: [],
        // 不选中任何项
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
