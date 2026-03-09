import React from 'react';
import * as AntdIcons from '@ant-design/icons';

// 导入所有图标，'allIcons' 将是一个包含所有图标组件的对象
const allIcons: Record<string, any> = AntdIcons;

// 默认图标，以防找不到匹配项
const DefaultIcon = <AntdIcons.AppstoreOutlined />;

/**
 * 根据后端返回的图标字符串名称 (e.g., "UserOutlined")
 * 获取 Ant Design 图标组件 (e.g., <UserOutlined />)
 * * @param name 图标的字符串名称
 * @returns ReactNode
 */
export const getIcon = (name?: React.ReactElement<unknown, string | React.JSXElementConstructor<any>> | string | number | bigint | Iterable<React.ReactNode> | React.ReactPortal | boolean | Promise<AwaitedReactNode> | null | undefined): React.ReactNode => {
  if (!name) {
    return null; // 没有图标
  }

  // 从所有图标中查找同名组件
  const IconComponent = allIcons[name];

  if (!IconComponent) {
    // 如果找不到 (e.g., 后端拼写错误)，返回一个默认图标
    console.warn(`[IconMap] 未找到图标: ${name}, 已使用默认图标。`);
    return DefaultIcon;
  }

  // 创建并返回 React 元素
  return React.createElement(IconComponent);
};
