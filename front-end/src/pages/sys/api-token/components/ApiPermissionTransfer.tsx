import { t, useI18n } from '@/i18n';
import React from 'react';
import { Space, Transfer, Typography } from 'antd';
import type { TransferItem } from 'antd/es/transfer';
import { formatApiLabel } from '../helpers';
export type ApiPermissionOption = {
  value: number;
  label: string;
  method?: string;
  path?: string;
  apiGroup?: string;
  description?: string;
};
type ApiPermissionTransferProps = {
  loading?: boolean;
  options: ApiPermissionOption[];
  value?: number[];
  onChange?: (value: number[]) => void;
};
type TransferDataItem = TransferItem & {
  title: string;
  description: string;
  searchText: string;
};
export const ApiPermissionTransfer: React.FC<ApiPermissionTransferProps> = ({
  loading,
  options,
  value,
  onChange,
}) => {
  useI18n();
  const dataSource: TransferDataItem[] = options.map((option) => ({
    key: String(option.value),
    title: option.label || formatApiLabel(option.method, option.path),
    description: option.description || option.apiGroup || t('cms.ungrouped'),
    searchText: `${option.label} ${option.method || ''} ${option.path || ''} ${option.apiGroup || ''}`,
  }));
  return (
    <Space
      direction="vertical"
      size={8}
      style={{
        width: '100%',
      }}
    >
      <Typography.Text type="secondary">
        {t('cms.selectApisOnTheLeftThe')}
      </Typography.Text>
      <Transfer
        oneWay
        showSearch
        dataSource={dataSource}
        targetKeys={(value || []).map(String)}
        onChange={(nextTargetKeys) =>
          onChange?.(nextTargetKeys.map((item) => Number(item)))
        }
        filterOption={(inputValue, item) =>
          (item as TransferDataItem).searchText
            .toLowerCase()
            .includes(inputValue.toLowerCase())
        }
        render={(item) => ({
          label: (
            <Space direction="vertical" size={0}>
              <Typography.Text>
                {(item as TransferDataItem).title}
              </Typography.Text>
              <Typography.Text
                type="secondary"
                style={{
                  fontSize: 12,
                }}
              >
                {(item as TransferDataItem).description}
              </Typography.Text>
            </Space>
          ),
          value: (item as TransferDataItem).title,
        })}
        disabled={loading}
        listStyle={{
          width: 280,
          height: 320,
        }}
        titles={[t('cms.availableApis'), t('cms.authorizedApis')]}
        locale={{
          itemUnit: t('cms.items'),
          itemsUnit: t('cms.items'),
          searchPlaceholder: t('cms.searchApiPathsMethodsOrGroups'),
        }}
      />
    </Space>
  );
};
