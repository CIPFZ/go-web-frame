import React, { useEffect, useState } from 'react';
import { Link, request, history } from '@umijs/max';
import { Alert, App, Card, Spin } from 'antd';
import { ProForm, ProFormText } from '@ant-design/pro-components';
import { getPublicConfig } from '@/services/system/user';

export default function Register() {
  const [enabled, setEnabled] = useState<boolean>();
  const [failed, setFailed] = useState(false);
  const { message } = App.useApp();
  useEffect(() => { getPublicConfig().then(r => setEnabled(r.data.registrationEnabled)).catch(() => setFailed(true)); }, []);
  return <Card title="注册账号" style={{ maxWidth: 420, margin: '80px auto' }}>
    {failed ? <Alert type="error" message="暂时无法获取注册配置，请稍后重试" /> : enabled === undefined ? <Spin /> : !enabled ?
      <Alert type="info" message="注册已关闭，请联系管理员创建账号" /> :
      <ProForm<{username: string; password: string; nickName?: string}> submitter={{ searchConfig: { submitText: '注册' }, resetButtonProps: false }} onFinish={async values => {
        const res = await request<{code: number; msg: string}>('/api/v1/user/register', { method: 'POST', data: values });
        if (res.code !== 0) { message.error(res.msg); return false; }
        message.success('注册成功，请登录'); history.push('/user/login'); return true;
      }}>
        <ProFormText name="username" label="用户名" rules={[{required: true}, {min: 3, max: 64}]} />
        <ProFormText name="nickName" label="昵称" />
        <ProFormText.Password name="password" label="密码" rules={[{required: true}, {min: 8, max: 72}]} />
      </ProForm>}
    <Link to="/user/login">返回登录</Link>
  </Card>;
}
