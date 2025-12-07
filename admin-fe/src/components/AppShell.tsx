import { Layout, Menu, Typography } from 'antd';
import { DashboardOutlined, EnvironmentOutlined, LogoutOutlined, ShopOutlined, TagsOutlined, BarChartOutlined, TeamOutlined, SafetyCertificateOutlined, ProfileOutlined } from '@ant-design/icons';
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
          <Menu.Item key="/users" icon={<TeamOutlined />}>
            <Link to="/users">用户管理</Link>
          </Menu.Item>
          <Menu.Item key="/rbac" icon={<SafetyCertificateOutlined />}>
            <Link to="/rbac">RBAC 权限</Link>
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
