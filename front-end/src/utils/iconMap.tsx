import React from 'react';
import * as AntdIcons from '@ant-design/icons';

const allIcons = AntdIcons as unknown as Record<string, React.ComponentType>;
const DefaultIcon = <AntdIcons.AppstoreOutlined />;

export const getIcon = (name?: string | null): React.ReactNode => {
  if (!name) {
    return null;
  }

  const IconComponent = allIcons[name];
  if (!IconComponent) {
    console.warn(`[IconMap] missing icon: ${name}, fallback to default icon.`);
    return DefaultIcon;
  }

  return React.createElement(IconComponent);
};
