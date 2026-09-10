import { useFormLocale } from '@/i18n/useFormLocale';
import { SelectLang } from '@/components';
import { t, useI18n } from '@/i18n';
import React, { useEffect, useState } from 'react';
import { Link, request, history } from '@umijs/max';
import { Alert, App, Card, Spin } from 'antd';
import { ProForm, ProFormText } from '@ant-design/pro-components';
import { getPublicConfig } from '@/services/system/user';
export default function Register() {
  const localeFormRef1 = useFormLocale();
  useI18n();
  const [enabled, setEnabled] = useState<boolean>();
  const [failed, setFailed] = useState(false);
  const { message } = App.useApp();
  useEffect(() => {
    getPublicConfig()
      .then((r) => setEnabled(r.data.registrationEnabled))
      .catch(() => setFailed(true));
  }, []);
  return (
    <Card
      extra={<SelectLang />}
      title={t('cms.createAnAccount')}
      style={{
        maxWidth: 420,
        margin: '80px auto',
      }}
    >
      {failed ? (
        <Alert
          type="error"
          message={t('cms.registrationSettingsAreUnavailablePleaseTry')}
        />
      ) : enabled === undefined ? (
        <Spin />
      ) : !enabled ? (
        <Alert
          type="info"
          message={t('cms.registrationIsClosedContactAnAdministrator')}
        />
      ) : (
        <ProForm<{
          username: string;
          password: string;
          nickName?: string;
        }>
          submitter={{
            searchConfig: {
              submitText: t('cms.register'),
            },
            resetButtonProps: false,
          }}
          onFinish={async (values) => {
            const res = await request<{
              code: number;
              msg: string;
            }>('/api/v1/user/register', {
              method: 'POST',
              data: values,
            });
            if (res.code !== 0) {
              message.error(res.msg);
              return false;
            }
            message.success(t('cms.accountCreatedPleaseSignIn'));
            history.push('/user/login');
            return true;
          }}
          formRef={localeFormRef1}
        >
          <ProFormText
            name="username"
            label={t('cms.username')}
            rules={[
              {
                required: true,
              },
              {
                min: 3,
                max: 64,
              },
            ]}
          />
          <ProFormText name="nickName" label={t('cms.nickname')} />
          <ProFormText.Password
            name="password"
            label={t('cms.password')}
            rules={[
              {
                required: true,
              },
              {
                min: 8,
                max: 72,
              },
            ]}
          />
        </ProForm>
      )}
      <Link to="/user/login">{t('cms.backToSignIn')}</Link>
    </Card>
  );
}
