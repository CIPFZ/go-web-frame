import React from 'react';
import { getLocale } from '@umijs/max';
import { Tag } from 'antd';

import { isEnglishLocale } from '@/utils/plugin';

const releaseStatusText: Record<number, { zh: string; en: string }> = {
  0: { zh: '草稿', en: 'Draft' },
  1: { zh: '待提交', en: 'Ready' },
  2: { zh: '待审核', en: 'Pending Review' },
  3: { zh: '已通过', en: 'Approved' },
  4: { zh: '已驳回', en: 'Rejected' },
  5: { zh: '已发布', en: 'Released' },
  6: { zh: '已下架', en: 'Offlined' },
};

const processStatusText: Record<number, { zh: string; en: string }> = {
  0: { zh: '待处理', en: 'Pending' },
  1: { zh: '处理中', en: 'Processing' },
  2: { zh: '已退回', en: 'Rejected' },
  3: { zh: '已完成', en: 'Done' },
};

const requestTypeText: Record<number, { zh: string; en: string }> = {
  1: { zh: '版本发布', en: 'Version Release' },
  2: { zh: '下架申请', en: 'Offline Request' },
};

const pickLabel = (locale: string | undefined, item?: { zh: string; en: string }, fallback?: number) => {
  if (!item) return String(fallback ?? '');
  return isEnglishLocale(locale) ? item.en : item.zh;
};

export function ReleaseStatusTag({ status, locale }: { status: number; locale?: string }) {
  const currentLocale = locale || getLocale();
  const colorMap: Record<number, string> = {
    0: 'default',
    1: 'blue',
    2: 'gold',
    3: 'processing',
    4: 'red',
    5: 'green',
    6: 'default',
  };
  return <Tag color={colorMap[status] || 'default'}>{pickLabel(currentLocale, releaseStatusText[status], status)}</Tag>;
}

export function ProcessStatusTag({ status, locale }: { status: number; locale?: string }) {
  const currentLocale = locale || getLocale();
  const colorMap: Record<number, string> = {
    0: 'gold',
    1: 'processing',
    2: 'red',
    3: 'green',
  };
  return <Tag color={colorMap[status] || 'default'}>{pickLabel(currentLocale, processStatusText[status], status)}</Tag>;
}

export function RequestTypeTag({ type, locale }: { type: number; locale?: string }) {
  const currentLocale = locale || getLocale();
  const colorMap: Record<number, string> = {
    1: 'blue',
    2: 'volcano',
  };
  return <Tag color={colorMap[type] || 'default'}>{pickLabel(currentLocale, requestTypeText[type], type)}</Tag>;
}
