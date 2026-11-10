// 项目自己的工作台扩展：小组件、组件库分组、应用图标的颜色与角标、外部链接、角色预置的默认工作台。
// 框架自带的小组件见 pages/dashboard/workplace/widgets/index.tsx；扩展点见 pages/dashboard/workplace/registry.ts。
//
// 例：
//
//   import { registerAppTint, registerRolePreset, registerWidget, registerWidgetGroup } from '@/pages/dashboard/workplace/registry';
//
//   registerWidgetGroup('orders', () => t('orders.title'));
//   registerWidget({
//     key: 'orders', group: 'orders', title: () => t('orders.title'), description: () => t('orders.widgetDesc'),
//     icon: <ShoppingOutlined />, tint: ['#60a5fa', '#2563eb'], sizes: ['m', 's', 'l'], to: '/orders/list', Body: OrdersWidget,
//   });
//   registerAppTint('/orders/', ['#60a5fa', '#2563eb']);
//   registerRolePreset(888, { widgets: [['greeting', 'w'], ['orders', 'm']], apps: ['/orders/list'] });
//
// 小组件的数据用 widgets/shared.tsx 的 useShared（同一份数据只查一次、自动刷新），状态用 StatusView 显示。

export {};
