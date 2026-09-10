import { render as rtlRender } from '@testing-library/react';
import { IntlProvider } from '@umijs/max';
import messages from '../src/locales/zh-CN';
import { createElement, type ReactNode } from 'react';

export function render(node: ReactNode) {
  return rtlRender(
    createElement(IntlProvider, { locale: 'zh-CN', messages }, node),
  );
}
