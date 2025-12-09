import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, Tabs, Typography, message, Space } from 'antd';
import { LockOutlined, ReloadOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';
import { fetchCaptcha, CaptchaResponse } from '../services/auth';

export default function LoginPage() {
  const navigate = useNavigate();
  const { devLogin, login, token, loading: authLoading } = useAuthContext();
  const [formLoading, setFormLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'password' | 'dev'>('password');
  const [captcha, setCaptcha] = useState<CaptchaResponse | null>(null);
  const [captchaLoading, setCaptchaLoading] = useState(false);

  const loadCaptcha = useCallback(async () => {
    setCaptchaLoading(true);
    try {
      const data = await fetchCaptcha();
      setCaptcha(data);
    } catch (error: any) {
      message.error(error?.message || '获取验证码失败');
    } finally {
      setCaptchaLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!authLoading && token) {
      navigate('/dashboard', { replace: true });
    }
  }, [token, authLoading, navigate]);

  useEffect(() => {
    loadCaptcha();
  }, [loadCaptcha]);

  const extractErrorMessage = (error: any) => {
    if (error?.response?.data?.message) return error.response.data.message;
    if (error?.response?.data?.error) return error.response.data.error;
    if (error?.message) return error.message;
    return '登录失败';
  };

  const handlePasswordLogin = async (values: { username: string; password: string; captcha_code: string }) => {
    if (!captcha?.id) {
      message.warning('请先获取验证码');
      await loadCaptcha();
      return;
    }
    setFormLoading(true);
    try {
      await login({
        username: values.username,
        password: values.password,
        captcha_id: captcha.id,
        captcha_code: values.captcha_code,
      });
      message.success('登录成功');
      navigate('/dashboard', { replace: true });
    } catch (error: any) {
      message.error(extractErrorMessage(error));
      loadCaptcha();
    } finally {
      setFormLoading(false);
    }
  };

  const handleDevLogin = async (values: { openid: string }) => {
    setFormLoading(true);
    try {
      await devLogin(values.openid);
      message.success('Dev 登录成功');
      navigate('/dashboard', { replace: true });
    } catch (error: any) {
      message.error(extractErrorMessage(error));
    } finally {
      setFormLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg,#f5f5f5,#e0f0ea)' }}>
      <Card style={{ width: 420 }} title={<Typography.Title level={4} style={{ margin: 0 }}>茶心阁 Admin 登录</Typography.Title>}>
        <Tabs
          activeKey={activeTab}
          onChange={(key) => setActiveTab(key as 'password' | 'dev')}
          items={[
            { key: 'password', label: '账号密码登录' },
            { key: 'dev', label: 'Dev OpenID 登录' },
          ]}
        />

        {activeTab === 'password' ? (
          <Form layout="vertical" onFinish={handlePasswordLogin} initialValues={{ username: '', password: '', captcha_code: '' }}>
            <Form.Item name="username" label="账号" rules={[{ required: true, message: '请输入账号' }]}> 
              <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
            </Form.Item>
            <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}> 
              <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
            </Form.Item>
            <Form.Item label="验证码" required>
              <Space align="baseline">
                <Form.Item name="captcha_code" noStyle rules={[{ required: true, message: '请输入验证码' }]}> 
                  <Input prefix={<SafetyCertificateOutlined />} placeholder="请输入验证码" style={{ width: 160 }} />
                </Form.Item>
                {captcha?.image ? (
                  <img
                    src={captcha.image}
                    alt="验证码"
                    style={{ width: 100, height: 36, borderRadius: 4, border: '1px solid #e5e5e5', objectFit: 'cover', cursor: 'pointer' }}
                    onClick={loadCaptcha}
                  />
                ) : (
                  <div style={{ width: 100, height: 36, display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px dashed #ccc', borderRadius: 4 }}>
                    加载中
                  </div>
                )}
                <Button icon={<ReloadOutlined />} onClick={loadCaptcha} loading={captchaLoading} type="link">
                  换一张
                </Button>
              </Space>
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={formLoading}>
              登录
            </Button>
          </Form>
        ) : (
          <Form layout="vertical" onFinish={handleDevLogin} initialValues={{ openid: '' }}>
            <Form.Item name="openid" label="OpenID" rules={[{ required: true, message: '请输入 openid' }]}> 
              <Input prefix={<UserOutlined />} placeholder="admin_openid" />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={formLoading}>
              Dev 登录
            </Button>
          </Form>
        )}
      </Card>
    </div>
  );
}
