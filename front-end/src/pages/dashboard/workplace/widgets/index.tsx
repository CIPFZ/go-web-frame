import { t } from '@/i18n';
import {
  BellOutlined,
  CalendarOutlined,
  DashboardOutlined,
  UserOutlined,
} from '@ant-design/icons';
import React from 'react';
import {
  registerAppTint,
  registerRolePreset,
  registerWidget,
  registerWidgetGroup,
  setRolePriority,
  setStarterDesk,
} from '../registry';
import { GreetingWidget } from './greeting';
import { AccountWidget, NoticesWidget, ServiceWidget } from './platform';

// 框架自带的小组件与默认工作台。业务项目在 src/workplace.tsx 里注册自己的（可以覆盖这里的预置）。

registerWidgetGroup('platform', () => t('cms.desk.groupPlatform'));

registerWidget({
  key: 'greeting',
  group: 'platform',
  title: () => t('cms.widget.greeting'),
  description: () => t('cms.widget.desc.greeting'),
  icon: <CalendarOutlined />,
  tint: ['#60a5fa', '#2563eb'],
  sizes: ['w', 's', 'm'],
  bare: true,
  Body: GreetingWidget,
});
registerWidget({
  key: 'notices',
  group: 'platform',
  title: () => t('cms.myNotices'),
  description: () => t('cms.widget.desc.notices'),
  icon: <BellOutlined />,
  tint: ['#fb923c', '#f97316'],
  sizes: ['m', 's', 'l'],
  Body: NoticesWidget,
});
registerWidget({
  key: 'service',
  group: 'platform',
  title: () => t('cms.serviceOverview'),
  description: () => t('cms.widget.desc.service'),
  icon: <DashboardOutlined />,
  tint: ['#4ade80', '#16a34a'],
  sizes: ['s', 'm'],
  to: 'state',
  Body: ServiceWidget,
});
registerWidget({
  key: 'account',
  group: 'platform',
  title: () => t('cms.currentAccount'),
  description: () => t('cms.widget.desc.account'),
  icon: <UserOutlined />,
  tint: ['#94a3b8', '#64748b'],
  sizes: ['s', 'm'],
  to: 'user/info',
  public: true,
  Body: AccountWidget,
});

registerAppTint('/state', ['#4ade80', '#15803d']);
registerAppTint('/sys/', ['#94a3b8', '#475569']);

// 起步桌面与两个内置角色的预置：超级管理员（1）多一个服务概况与系统管理的常用页面
setStarterDesk({
  widgets: [
    ['greeting', 'w'],
    ['notices', 'm'],
    ['service', 's'],
    ['account', 's'],
  ],
  apps: [],
});
registerRolePreset(1, {
  widgets: [
    ['greeting', 'w'],
    ['service', 'm'],
    ['notices', 'm'],
  ],
  // 按组件标识指向页面：管理员改了菜单地址也能找到
  apps: [
    'sys/user',
    'sys/authority',
    'sys/menu',
    'sys/api',
    'sys/api-token',
    'sys/notice',
    'sys/operation',
  ],
});
registerRolePreset(888, {
  widgets: [
    ['greeting', 'w'],
    ['notices', 'm'],
    ['account', 's'],
  ],
  apps: [],
});
setRolePriority([1, 888]);
