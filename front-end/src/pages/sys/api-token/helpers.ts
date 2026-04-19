import dayjs from 'dayjs';

import type { ApiTokenItem } from '@/services/api/apiToken';
import type { ApiTokenFormPayload } from '@/services/api/apiToken';

export type ApiPermissionRecord = NonNullable<ApiTokenItem['apis']>[number];

export type TokenFormValues = {
  name?: string;
  description?: string;
  maxConcurrency?: number;
  expiresAt?: string;
  apiIds?: number[];
};

export type PermissionSummaryItem = {
  key: number;
  label: string;
  method: string;
  path: string;
};

const DEFAULT_MAX_CONCURRENCY = 5;

export const API_TOKEN_TABLE_LAYOUT = {
  nameWidth: 180,
  tokenPrefixWidth: 150,
  statusWidth: 88,
  concurrencyWidth: 76,
  expiresAtWidth: 158,
  lastUsedWidth: 120,
  apisWidth: 360,
  actionWidth: 220,
  scrollX: 1360,
} as const;

export const formatApiLabel = (method?: string, path?: string) =>
  `[${method || 'GET'}] ${path || '/'}`;

export const buildPermissionSummary = (
  apis: ApiTokenItem['apis'] = [],
  maxVisible = 2,
): {
  total: number;
  visible: PermissionSummaryItem[];
  hidden: PermissionSummaryItem[];
} => {
  const items = (apis || []).map((api) => ({
    key: api.ID,
    method: api.method,
    path: api.path,
    label: formatApiLabel(api.method, api.path),
  }));

  return {
    total: items.length,
    visible: items.slice(0, maxVisible),
    hidden: items.slice(maxVisible),
  };
};

export const buildTokenFormInitialValues = (currentRow?: ApiTokenItem): TokenFormValues => {
  if (!currentRow) {
    return {
      maxConcurrency: DEFAULT_MAX_CONCURRENCY,
      apiIds: [],
    };
  }

  return {
    name: currentRow.name,
    description: currentRow.description,
    maxConcurrency: currentRow.maxConcurrency,
    expiresAt: currentRow.expiresAt,
    apiIds: (currentRow.apis || []).map((api) => api.ID),
  };
};

export const buildTokenSubmitPayload = (
  values: TokenFormValues,
  currentId?: number,
): ApiTokenFormPayload => ({
  ...(currentId ? { id: currentId } : {}),
  name: values.name || '',
  description: values.description,
  maxConcurrency: values.maxConcurrency,
  expiresAt: values.expiresAt ? dayjs(values.expiresAt).toISOString() : undefined,
  apiIds: values.apiIds || [],
});
