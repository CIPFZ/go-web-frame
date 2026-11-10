import { t } from '@/i18n';
import { getTokenLimit, updateTokenLimit, type TokenLimitDetail } from '@/services/system/apiToken';
import { ModalForm, ProFormDigit } from '@ant-design/pro-components';
import { Typography, message } from 'antd';
import React, { useState } from 'react';

type Values = { maxTokens?: number; maxTtlDays?: number; maxConcurrency?: number };

/** Token 申请限制：userId 为 0 是全局默认；单个用户的设置留空表示跟随默认。只影响之后的申请与修改。 */
export const TokenLimitForm: React.FC<{ userId: number; title: string; trigger: React.ReactElement }> = ({ userId, title, trigger }) => {
  const [detail, setDetail] = useState<TokenLimitDetail>();
  const isDefault = userId === 0;
  const hint = (value?: number | string) => (isDefault || value === undefined ? undefined : t('cms.tokenLimitFollowDefault', { value }));
  const base = detail?.default;
  return (
    <ModalForm<Values>
      title={title}
      trigger={trigger}
      width={480}
      modalProps={{ destroyOnClose: true }}
      request={async () => {
        const res = await getTokenLimit(userId);
        const d = res.data as TokenLimitDetail;
        setDetail(d);
        const from = isDefault ? d.default : d;
        return {
          maxTokens: from.maxTokens ?? undefined,
          maxTtlDays: from.maxTtlDays ?? undefined,
          maxConcurrency: from.maxConcurrency ?? undefined,
        };
      }}
      onFinish={async (v) => {
        const res = await updateTokenLimit({ userId, ...v });
        if (res.code !== 0) {
          message.error(res.msg || t('cms.saveFailed'));
          return false;
        }
        message.success(t('cms.updatedSuccessfully'));
        return true;
      }}
    >
      <Typography.Paragraph type="secondary">{t(isDefault ? 'cms.tokenLimitDefaultHint' : 'cms.tokenLimitUserHint')}</Typography.Paragraph>
      <ProFormDigit name="maxTokens" label={t('cms.tokenLimitMaxTokens')} min={0} max={1000} fieldProps={{ precision: 0, placeholder: hint(base?.maxTokens) }} />
      <ProFormDigit name="maxTtlDays" label={t('cms.tokenLimitMaxTtl')} min={1} max={3650} fieldProps={{ precision: 0, placeholder: hint(base?.maxTtlDays) }} />
      <ProFormDigit name="maxConcurrency" label={t('cms.maximumConcurrency')} min={1} max={1000} fieldProps={{ precision: 0, placeholder: hint(base?.maxConcurrency) }} />
    </ModalForm>
  );
};
