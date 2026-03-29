import React from 'react';
import * as AntdIcons from '@ant-design/icons';

const allIcons = AntdIcons as unknown as Record<string, React.ComponentType | undefined>;
const DefaultIcon = <AntdIcons.AppstoreOutlined />;

export const getIcon = (name?: string | null): React.ReactNode => {
  if (!name) {
    return null;
  }

  const IconComponent = allIcons[name];
  if (!IconComponent) {
    console.warn(`[IconMap] 未找到图标 ${name}，已使用默认图标。`);
    return DefaultIcon;
  }

  return React.createElement(IconComponent);
};
