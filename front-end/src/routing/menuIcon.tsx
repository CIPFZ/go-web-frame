import React from 'react';

type LabelProps = {
  children?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  'aria-hidden'?: boolean;
};

// ProLayout omits icons below the first level. Keep its translated label and
// collapse styles, but fill the icon slot (or add it for nested submenu titles).
export function withMenuIcon(icon: React.ReactNode, label: React.ReactNode): React.ReactNode {
  if (!icon || !React.isValidElement<LabelProps>(label)) return label;

  let hasIconSlot = false;
  const children = React.Children.map(label.props.children, child => {
    if (!React.isValidElement<LabelProps>(child) ||
      !child.props.className?.split(' ').some(name => name.endsWith('-item-icon'))) {
      return child;
    }
    hasIconSlot = true;
    return React.cloneElement(child, {
      'aria-hidden': true,
      style: { ...child.props.style, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 },
    }, icon);
  });

  const labelProps = { style: { ...label.props.style, display: 'flex', alignItems: 'center' } };
  if (hasIconSlot) return React.cloneElement(label, labelProps, children);
  return React.cloneElement(label, labelProps,
    <span className="cms-menu-icon" aria-hidden style={{ display: 'inline-flex', alignItems: 'center', flexShrink: 0 }}>
      {icon}
    </span>,
    label.props.children,
  );
}
