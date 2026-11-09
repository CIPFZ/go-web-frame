import { useFormLocale } from '@/i18n/useFormLocale';
import { t, useI18n } from '@/i18n';
import type { API } from '@/services/system/types';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import {
  LoginForm,
  ProFormCheckbox,
  ProFormText,
} from '@ant-design/pro-components';
import { FormattedMessage, Helmet, useIntl, useModel } from '@umijs/max';
import { Alert, App, Button, Divider, Space, Tabs } from 'antd';
import { createStyles } from 'antd-style';
import React, { useState, useEffect, useRef } from 'react';
import { flushSync } from 'react-dom';
import { Footer, SelectLang } from '@/components';
import { extTicket, login, getPublicConfig, type PublicConfig } from '@/services/system/user';
import Settings from '../../../../config/defaultSettings';
const useStyles = createStyles(({ token }) => {
  return {
    action: {
      marginLeft: '8px',
      color: 'rgba(0, 0, 0, 0.2)',
      fontSize: '24px',
      verticalAlign: 'middle',
      cursor: 'pointer',
      transition: 'color 0.3s',
      '&:hover': {
        color: token.colorPrimaryActive,
      },
    },
    lang: {
      width: 42,
      height: 42,
      lineHeight: '42px',
      position: 'fixed',
      right: 16,
      borderRadius: token.borderRadius,
      ':hover': {
        backgroundColor: token.colorBgTextHover,
      },
    },
    container: {
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      overflow: 'auto',
      backgroundImage: "url('/V-_oS6r-i7wAAAAAAAAAAAAAFl94AQBr.png')",
      backgroundSize: '100% 100%',
    },
  };
});

// 选择语言
const Lang = () => {
  useI18n();
  const { styles } = useStyles();
  return (
    <div className={styles.lang} data-lang>
      {SelectLang && <SelectLang />}
    </div>
  );
};
const LoginMessage: React.FC<{
  content: string;
}> = ({ content }) => {
  useI18n();
  return (
    <Alert
      style={{
        marginBottom: 24,
      }}
      message={content}
      type="error"
      showIcon
    />
  );
};
// 登录页 hash 路由里的参数（外部登录回调带回 ext_ticket / ext_error / redirect）
const hashParams = () =>
  new URLSearchParams(window.location.hash.split('?')[1] ?? '');
// 登录后要去的站内页面（只接受站内路径）
const loginRedirect = () => {
  const redirect =
    hashParams().get('redirect') ??
    new URL(window.location.href).searchParams.get('redirect');
  return redirect?.startsWith('/') &&
    !redirect.startsWith('//') &&
    !redirect.includes('\\')
    ? redirect
    : '';
};
const EXT_ERRORS: Record<string, string> = {
  unknown: 'cms.extLogin.failed',
  state: 'cms.extLogin.expired',
  failed: 'cms.extLogin.failed',
  not_linked: 'cms.extLogin.notLinked',
  user_disabled: 'cms.extLogin.userDisabled',
};

const Login: React.FC = () => {
  const localeFormRef1 = useFormLocale();
  useI18n();
  const [publicConfig, setPublicConfig] = useState<PublicConfig>();
  useEffect(() => {
    getPublicConfig()
      .then((r) => setPublicConfig(r.data))
      .catch(() => setPublicConfig(undefined));
  }, []);
  const registrationEnabled = !!publicConfig?.registrationEnabled;
  const externalLogins = publicConfig?.externalLogins || [];
  const [userLoginState, setUserLoginState] = useState<API.LoginResult>({
    msg: '',
    code: -1,
  });
  const [type, setType] = useState<string>('account');
  const { initialState, setInitialState } = useModel('@@initialState');
  const { styles } = useStyles();
  const { message } = App.useApp();
  const intl = useIntl();

  // 获取用户信息
  const fetchUserInfo = async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      flushSync(() => {
        setInitialState((s) => ({
          ...s,
          currentUser: userInfo,
        }));
      });
    }
  };

  // 处理表单提交 - 登录
  const handleSubmit = async (values: API.LoginParams) => {
    try {
      // 登录
      const response = await login({
        ...values,
        type,
      });
      // 判断是否登录成功
      if (response.code === 0) {
        const defaultLoginSuccessMessage = intl.formatMessage({
          id: 'pages.login.success',
          defaultMessage: '登录成功！',
        });
        message.success(defaultLoginSuccessMessage);
        // ✨✨✨ 关键修改：保存 Token ✨✨✨
        // 假设后端返回结构是 response.data.token
        const token = response.data?.token;
        if (token) {
          localStorage.setItem('token', token);
        }
        // 获取用户具体信息
        await fetchUserInfo();
        // 路由跳转
        window.location.href = loginRedirect() || '/';
        return;
      }
      console.log(response);
      // 如果失败去设置登录失败错误信息
      setUserLoginState({
        code: response.code,
        msg: response.msg,
      });
    } catch (error) {
      const defaultLoginFailureMessage = intl.formatMessage({
        id: 'pages.login.failure',
        defaultMessage: '登录失败，请重试！',
      });
      console.log(error);
      message.error(defaultLoginFailureMessage);
    }
  };
  // 外部登录回调带回的票据换登录凭证；票据只能用一次，先从地址栏去掉
  const ticketHandled = useRef(false);
  useEffect(() => {
    if (ticketHandled.current) return;
    ticketHandled.current = true;
    const params = hashParams();
    const ticket = params.get('ext_ticket');
    const error = params.get('ext_error');
    if (!ticket && !error) return;
    const redirect = loginRedirect();
    window.history.replaceState(
      null,
      '',
      `${window.location.pathname}#/user/login${redirect ? `?redirect=${encodeURIComponent(redirect)}` : ''}`,
    );
    if (error) {
      message.error(t((EXT_ERRORS[error] || 'cms.extLogin.failed') as 'cms.extLogin.failed'));
      return;
    }
    void (async () => {
      const res = await extTicket(ticket as string).catch(() => undefined);
      if (res?.code !== 0) {
        message.error(t('cms.extLogin.failed'));
        return;
      }
      const token = (res.data as { token?: string })?.token;
      if (token) localStorage.setItem('token', token);
      message.success(intl.formatMessage({ id: 'pages.login.success', defaultMessage: '登录成功！' }));
      await fetchUserInfo();
      window.location.href = redirect || '/';
    })();
    // 只在打开登录页时处理一次
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  // 外部登录：跳到后端，由后端把浏览器送到提供方的登录页
  const startExternalLogin = (key: string) => {
    const redirect = loginRedirect();
    window.location.href = `/api/v1/user/ext-login/${encodeURIComponent(key)}${redirect ? `?redirect=${encodeURIComponent(redirect)}` : ''}`;
  };
  return (
    <div className={styles.container}>
      <Helmet>
        <title>
          {intl.formatMessage({
            id: 'menu.login',
            defaultMessage: '登录页',
          })}
          {Settings.title && ` - ${Settings.title}`}
        </title>
      </Helmet>
      <Lang />
      <div
        style={{
          flex: '1',
          padding: '32px 0',
        }}
      >
        <LoginForm
          contentStyle={{
            minWidth: 280,
            maxWidth: '75vw',
          }}
          logo={<img alt="logo" src="/logo.svg" />}
          title="Go Web Frame"
          subTitle={intl.formatMessage({
            id: 'pages.layouts.userLayout.title',
          })}
          initialValues={{
            autoLogin: true,
          }}
          onFinish={async (values) => {
            await handleSubmit(values as API.LoginParams);
          }}
          formRef={localeFormRef1}
          actions={
            externalLogins.length ? (
              <>
                <Divider plain>{t('cms.extLogin.or')}</Divider>
                <Space direction="vertical" style={{ width: '100%' }}>
                  {externalLogins.map((p) => (
                    <Button key={p.key} block onClick={() => startExternalLogin(p.key)}>
                      {t('cms.extLogin.with', { name: p.name })}
                    </Button>
                  ))}
                </Space>
              </>
            ) : undefined
          }
        >
          <Tabs
            activeKey={type}
            onChange={setType}
            centered
            items={[
              {
                key: 'account',
                label: intl.formatMessage({
                  id: 'pages.login.accountLogin.tab',
                  defaultMessage: '账户密码登录',
                }),
              },
            ]}
          />
          {/*如果登录失败显示*/}
          {userLoginState.code > 0 && (
            <LoginMessage
              content={intl.formatMessage({
                id: 'pages.login.accountLogin.errorMessage',
                defaultMessage: '账户或密码错误',
              })}
            />
          )}
          <>
            <ProFormText
              name="username"
              fieldProps={{
                size: 'large',
                prefix: <UserOutlined />,
              }}
              placeholder={intl.formatMessage({
                id: 'pages.login.username.placeholder',
                defaultMessage: '请输入用户名',
              })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.username.required"
                      defaultMessage="请输入用户名!"
                    />
                  ),
                },
              ]}
            />
            <ProFormText.Password
              name="password"
              fieldProps={{
                size: 'large',
                prefix: <LockOutlined />,
              }}
              placeholder={intl.formatMessage({
                id: 'pages.login.password.placeholder',
                defaultMessage: '请输入密码',
              })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.password.required"
                      defaultMessage="请输入密码！"
                    />
                  ),
                },
              ]}
            />
          </>
          <div
            style={{
              marginBottom: 24,
            }}
          >
            <ProFormCheckbox noStyle name="autoLogin">
              <FormattedMessage
                id="pages.login.rememberMe"
                defaultMessage="记住我"
              />
            </ProFormCheckbox>
            {registrationEnabled && (
              <a
                href="#/user/register"
                style={{
                  float: 'right',
                }}
              >
                {t('cms.createAnAccount')}
              </a>
            )}
          </div>
        </LoginForm>
      </div>
      <Footer />
    </div>
  );
};
export default Login;
