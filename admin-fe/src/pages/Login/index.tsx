import React, { useState } from 'react';
import { Card, Form, Input, Button, message } from 'antd';
import { devLogin } from '../../services/auth';
import { setToken } from '../../services/api';
import { useNavigate, useLocation } from 'react-router-dom';

export default function Login() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const location = useLocation();

  async function onSubmit() {
    try {
      const values = await form.validateFields();
      setLoading(true);
      const resp = await devLogin(values.openid || 'dev-openid');
      const token = resp?.token;
      if (token) {
        setToken(token);
        message.success('登录成功');
        const params = new URLSearchParams(location.search);
        const from = params.get('from') || '/';
        navigate(from, { replace: true });
      } else {
        message.error('登录失败：无 token');
      }
    } catch (e: any) {
      message.error(e?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
      <Card title="管理后台登录" style={{ width: 360 }}>
        <Form form={form} layout="vertical" onFinish={onSubmit} initialValues={{ openid: 'dev-openid' }}>
          <Form.Item name="openid" label="OpenID" rules={[{ required: true, message: '请输入 openid（开发环境）' }]}> <Input /> </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>登录</Button>
        </Form>
      </Card>
    </div>
  );
}
