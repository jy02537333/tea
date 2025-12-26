import { Layout, Menu, Typography } from 'antd';
import {
  DashboardOutlined,
  EnvironmentOutlined,
  LogoutOutlined,
  ShopOutlined,
  TagsOutlined,
  BarChartOutlined,
  TeamOutlined,
  SafetyCertificateOutlined,
  ProfileOutlined,
  SettingOutlined,
  AccountBookOutlined,
  GiftOutlined,
  FireOutlined,
  CustomerServiceOutlined,
  CreditCardOutlined,
} from '@ant-design/icons';
import { PropsWithChildren } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';

const { Header, Sider, Content } = Layout;

export function AppShell({ children }: PropsWithChildren) {
  const location = useLocation();
  const navigate = useNavigate();
  const { logout } = useAuthContext();

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  let selectedKey = location.pathname;
  if (location.pathname.startsWith('/products')) selectedKey = '/products';
  else if (location.pathname.startsWith('/accrual')) selectedKey = '/accrual';
  else if (location.pathname.startsWith('/users')) selectedKey = '/users';
  else if (location.pathname.startsWith('/rbac')) selectedKey = '/rbac';
  else if (location.pathname.startsWith('/orders')) selectedKey = '/orders';
  else if (location.pathname.startsWith('/categories')) selectedKey = '/categories';
  else if (location.pathname.startsWith('/stores')) selectedKey = '/stores';
  else if (location.pathname.startsWith('/store-finance')) selectedKey = '/store-finance';
  else if (location.pathname.startsWith('/store-accounts')) selectedKey = '/store-accounts';
  else if (location.pathname.startsWith('/store-coupons')) selectedKey = '/store-coupons';
  else if (location.pathname.startsWith('/store-activities')) selectedKey = '/store-activities';
  else if (location.pathname.startsWith('/tickets')) selectedKey = '/tickets';
  else if (location.pathname.startsWith('/commission-rollback')) selectedKey = '/commission-rollback';
  else if (location.pathname.startsWith('/finance-summary')) selectedKey = '/finance-summary';
  else if (location.pathname.startsWith('/finance-records')) selectedKey = '/finance-records';
  else if (location.pathname.startsWith('/membership-config')) selectedKey = '/membership-config';
  else if (location.pathname.startsWith('/partners')) selectedKey = '/partners';
  else if (location.pathname.startsWith('/banners')) selectedKey = '/banners';
  else if (location.pathname.startsWith('/recharge')) selectedKey = '/recharge';
  else if (location.pathname.startsWith('/logs')) selectedKey = '/logs';
  else if (location.pathname.startsWith('/system-settings')) selectedKey = '/system-settings';
  else if (location.pathname === '/' || location.pathname === '') selectedKey = '/dashboard';

  return (
    <Layout>
      <Sider breakpoint="lg" collapsedWidth="0">
        <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography.Title level={4} style={{ color: '#fff', margin: 0 }}>
            茶心阁
          </Typography.Title>
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[selectedKey]}>
          <Menu.Item key="/dashboard" icon={<DashboardOutlined />}>
            <Link to="/dashboard">仪表盘</Link>
          </Menu.Item>
          <Menu.Item key="/accrual" icon={<BarChartOutlined />}>
            <Link to="/accrual">计提/报表</Link>
          </Menu.Item>
          <Menu.Item key="/finance-summary" icon={<AccountBookOutlined />}>
            <Link to="/finance-summary">财务概览</Link>
          </Menu.Item>
          <Menu.Item key="/finance-records" icon={<AccountBookOutlined />}>
            <Link to="/finance-records">财务记录</Link>
          </Menu.Item>
          <Menu.Item key="/commission-rollback" icon={<AccountBookOutlined />}>
            <Link to="/commission-rollback">佣金回滚</Link>
          </Menu.Item>
          <Menu.Item key="/users" icon={<TeamOutlined />}>
            <Link to="/users">用户管理</Link>
          </Menu.Item>
          <Menu.Item key="/rbac" icon={<SafetyCertificateOutlined />}>
            <Link to="/rbac">RBAC 权限</Link>
          </Menu.Item>
          <Menu.Item key="/logs" icon={<ProfileOutlined />}>
            <Link to="/logs">日志</Link>
          </Menu.Item>
          <Menu.Item key="/system-settings" icon={<SettingOutlined />}>
            <Link to="/system-settings">系统设置</Link>
          </Menu.Item>
          <Menu.Item key="/products" icon={<ShopOutlined />}>
            <Link to="/products">商品管理</Link>
          </Menu.Item>
          <Menu.Item key="/categories" icon={<TagsOutlined />}>
            <Link to="/categories">分类管理</Link>
          </Menu.Item>
          <Menu.Item key="/stores" icon={<EnvironmentOutlined />}>
            <Link to="/stores">门店管理</Link>
          </Menu.Item>
          <Menu.Item key="/store-finance" icon={<AccountBookOutlined />}>
            <Link to="/store-finance">门店财务</Link>
          </Menu.Item>
          <Menu.Item key="/store-accounts" icon={<CreditCardOutlined />}>
            <Link to="/store-accounts">门店收款账户</Link>
          </Menu.Item>
          <Menu.Item key="/store-coupons" icon={<GiftOutlined />}>
            <Link to="/store-coupons">门店优惠券</Link>
          </Menu.Item>
          <Menu.Item key="/store-activities" icon={<FireOutlined />}>
            <Link to="/store-activities">门店活动</Link>
          </Menu.Item>
          <Menu.Item key="/banners" icon={<TagsOutlined />}>
            <Link to="/banners">广告管理</Link>
          </Menu.Item>
          <Menu.Item key="/recharge" icon={<AccountBookOutlined />}>
            <Link to="/recharge">充值管理</Link>
          </Menu.Item>
          <Menu.Item key="/membership-config" icon={<CustomerServiceOutlined />}>
            <Link to="/membership-config">会员配置</Link>
          </Menu.Item>
          <Menu.Item key="/partners" icon={<TeamOutlined />}>
            <Link to="/partners">合伙人管理</Link>
          </Menu.Item>
          <Menu.Item key="/partner-withdrawals" icon={<AccountBookOutlined />}>
            <Link to="/partner-withdrawals">合伙人提现</Link>
          </Menu.Item>
          <Menu.Item key="/tickets" icon={<CustomerServiceOutlined />}>
            <Link to="/tickets">客服工单</Link>
          </Menu.Item>
          <Menu.Item key="/orders" icon={<ProfileOutlined />}>
            <Link to="/orders">订单管理</Link>
          </Menu.Item>
          <Menu.Item key="logout" icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Menu.Item>
        </Menu>
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px' }}>
          <Typography.Text>欢迎回来 👋</Typography.Text>
        </Header>
        <Content style={{ margin: 24 }}>
          <div style={{ padding: 24, background: '#fff', minHeight: 'calc(100vh - 160px)' }}>{children}</div>
        </Content>
      </Layout>
    </Layout>
  );
}
