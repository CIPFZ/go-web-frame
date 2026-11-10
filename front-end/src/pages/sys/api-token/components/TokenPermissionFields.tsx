import { t } from '@/i18n';
import { ProForm, ProFormDependency, ProFormRadio } from '@ant-design/pro-components';
import React from 'react';
import { ApiPermissionTransfer, type ApiPermissionOption } from './ApiPermissionTransfer';

/** Token 的权限：跟随本人（默认）或自己选接口。都不会超出本人现在的权限。 */
export const TokenPermissionFields: React.FC<{ options: ApiPermissionOption[]; loading?: boolean; followLabel?: string }> = ({ options, loading, followLabel }) => (
  <>
    <ProFormRadio.Group
      name="permissionMode"
      label={t('cms.tokenPermissionMode')}
      tooltip={t('cms.tokenPermissionModeHelp')}
      options={[
        { label: followLabel ?? t('cms.tokenPermissionFollow'), value: 'follow' },
        { label: t('cms.tokenPermissionCustom'), value: 'custom' },
      ]}
    />
    <ProFormDependency name={['permissionMode']}>
      {({ permissionMode }) =>
        permissionMode === 'custom' ? (
          <ProForm.Item name="apiIds" label={t('cms.authorizedApis.653b57')} rules={[{ required: true, message: t('cms.selectAtLeastOneApi') }]}>
            <ApiPermissionTransfer loading={loading} options={options} />
          </ProForm.Item>
        ) : null
      }
    </ProFormDependency>
  </>
);
