import { Avatar, Badge, Button, Dropdown, Input, Layout, Menu, Space, Typography, Tag, message } from 'antd';
import dayjs from 'dayjs';
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
  ToolOutlined,
  AppstoreOutlined,
  BellOutlined,
  SearchOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { PropsWithChildren, useMemo, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';

const { Header, Sider, Content } = Layout;

export function AppShell({ children }: PropsWithChildren) {
  const location = useLocation();
  const navigate = useNavigate();
  const { logout, user } = useAuthContext();
  const [headerSearch, setHeaderSearch] = useState('');

  // “新”标识显示规则：PR 合并/部署后 30 天自动隐藏
  // 优先读取起始日期 VITE_STORE_FINANCE_NEW_BADGE_START（格式 YYYY-MM-DD），否则以当前构建/运行时间为起点
  const envStart = (import.meta as any)?.env?.VITE_STORE_FINANCE_NEW_BADGE_START as string | undefined;
  const buildCommitDate = (import.meta as any)?.env?.VITE_BUILD_COMMIT_DATE as string | undefined;
  const startDate = envStart && String(envStart).trim()
    ? dayjs(envStart, 'YYYY-MM-DD')
    : (buildCommitDate && String(buildCommitDate).trim() ? dayjs(buildCommitDate, 'YYYY-MM-DD') : dayjs());
  const endDate = startDate.add(30, 'day').endOf('day');
  const showNewBadge = dayjs().isBefore(endDate);

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  const isStoreAdmin = user?.role === 'store';

  const urlSearchParams = useMemo(() => new URLSearchParams(location.search), [location.search]);
  const productTab = urlSearchParams.get('tab') || '';
  const orderTab = urlSearchParams.get('tab') || '';
  const systemTab = urlSearchParams.get('tab') || '';

  let selectedKey = location.pathname;
  if (location.pathname.startsWith('/products')) {
    if (productTab === 'on') selectedKey = '/products/on';
    else if (productTab === 'off') selectedKey = '/products/off';
    else if (productTab === 'draft') selectedKey = '/products/draft';
    else if (productTab === 'create') selectedKey = '/products/create';
    else selectedKey = '/products/all';
  }
  else if (location.pathname.startsWith('/accrual')) selectedKey = '/accrual';
  else if (location.pathname.startsWith('/users')) selectedKey = '/users';
  else if (location.pathname.startsWith('/rbac')) selectedKey = '/rbac';
  else if (location.pathname.startsWith('/orders')) {
    if (orderTab === 'after-sales') selectedKey = '/orders/after-sales';
    else selectedKey = '/orders/all';
  }
  else if (location.pathname.startsWith('/categories')) selectedKey = '/products/categories';
  else if (location.pathname.startsWith('/stores/')) {
    if (location.pathname.includes('/orders')) selectedKey = '/store-orders';
    else if (location.pathname.includes('/refunds')) selectedKey = '/store-refunds';
    else if (location.pathname.includes('/products')) selectedKey = '/store-products';
    else if (location.pathname.includes('/mall')) selectedKey = '/store-mall';
    else selectedKey = '/stores';
  }
  else if (location.pathname.startsWith('/stores')) selectedKey = '/stores';
  else if (location.pathname.startsWith('/store-finance')) selectedKey = '/store-finance';
  else if (location.pathname.startsWith('/store-accounts')) selectedKey = '/store-accounts';
  else if (location.pathname.startsWith('/store-coupons')) selectedKey = '/store-coupons';
  else if (location.pathname.startsWith('/store-activities')) selectedKey = '/store-activities';
  else if (location.pathname.startsWith('/store-home')) selectedKey = '/store-home';
  else if (location.pathname.startsWith('/store-orders')) selectedKey = '/store-orders';
  else if (location.pathname.startsWith('/store-refunds')) selectedKey = '/store-refunds';
  else if (location.pathname.startsWith('/store-products')) selectedKey = '/store-products';
  else if (location.pathname.startsWith('/store-mall')) selectedKey = '/store-mall';
  else if (location.pathname.startsWith('/store-settings')) selectedKey = '/store-settings';
  else if (location.pathname.startsWith('/tickets')) selectedKey = '/tickets';
  else if (location.pathname.startsWith('/commission-rollback')) selectedKey = '/commission-rollback';
  else if (location.pathname.startsWith('/finance-summary')) selectedKey = '/finance-summary';
  else if (location.pathname.startsWith('/finance-records')) selectedKey = '/finance-records';
  else if (location.pathname.startsWith('/membership-config')) selectedKey = '/membership-config';
  else if (location.pathname.startsWith('/partners')) selectedKey = '/partners';
  else if (location.pathname.startsWith('/banners')) selectedKey = '/banners';
  else if (location.pathname.startsWith('/recharge')) selectedKey = '/recharge';
  else if (location.pathname.startsWith('/logs')) selectedKey = '/logs';
  else if (location.pathname.startsWith('/system-settings')) {
    if (systemTab === 'admins') selectedKey = '/system-settings/admins';
    else selectedKey = '/system-settings/basic';
  }
  else if (location.pathname === '/' || location.pathname === '') selectedKey = '/dashboard';

  const openKeys = useMemo(() => {
    if (selectedKey.startsWith('/products/')) return ['products'];
    if (selectedKey.startsWith('/orders/')) return ['orders'];
    if (selectedKey.startsWith('/system-settings/')) return ['system-settings'];
    return [] as string[];
  }, [selectedKey]);

  const handleHeaderSearch = (raw: string) => {
    const keyword = String(raw ?? '').trim();
    if (!keyword) return;

    // 最小可用：
    // - 平台端：订单号 → 订单列表；其他 → 商品管理
    // - 门店端：优先支持订单号（跳到本店订单列表），非订单号暂不支持
    const looksLikeOrderNo = /^o/i.test(keyword) || /^\d+$/.test(keyword);
    if (isStoreAdmin) {
      if (!looksLikeOrderNo) {
        message.info('门店端搜索暂仅支持订单号');
        return;
      }
      const sid = user?.store_id;
      if (!sid) {
        message.error('门店管理员未绑定门店');
        return;
      }
      navigate(`/stores/${sid}/orders?orderNo=${encodeURIComponent(keyword)}`);
      return;
    }

    if (looksLikeOrderNo) {
      navigate(`/orders?orderNo=${encodeURIComponent(keyword)}`);
      return;
    }
    navigate(`/products?keyword=${encodeURIComponent(keyword)}`);
  };

  const quickNavItems = useMemo(() => {
    if (isStoreAdmin) {
      return [
        { key: 'nav:store-home', icon: <DashboardOutlined />, label: '首页', onClick: () => navigate('/store-home') },
        { key: 'nav:store-orders', icon: <ProfileOutlined />, label: '订单管理', onClick: () => navigate('/store-orders') },
        { key: 'nav:store-products', icon: <ShopOutlined />, label: '商品管理', onClick: () => navigate('/store-products') },
        { key: 'nav:store-mall', icon: <ShopOutlined />, label: '商家商城', onClick: () => navigate('/store-mall') },
        { key: 'nav:store-finance', icon: <AccountBookOutlined />, label: '财务管理', onClick: () => navigate('/store-finance') },
        { key: 'nav:store-coupons', icon: <GiftOutlined />, label: '优惠券', onClick: () => navigate('/store-coupons') },
        { key: 'nav:store-activities', icon: <FireOutlined />, label: '活动', onClick: () => navigate('/store-activities') },
        { key: 'nav:store-settings', icon: <SettingOutlined />, label: '门店设置', onClick: () => navigate('/store-settings') },
      ];
    }
    return [
      { key: 'nav:dashboard', icon: <DashboardOutlined />, label: '仪表盘', onClick: () => navigate('/dashboard') },
      { key: 'nav:orders', icon: <ProfileOutlined />, label: '订单列表', onClick: () => navigate('/orders') },
      { key: 'nav:afterSales', icon: <ProfileOutlined />, label: '售后订单', onClick: () => navigate('/orders?tab=after-sales') },
      { key: 'nav:products', icon: <ShopOutlined />, label: '商品管理', onClick: () => navigate('/products') },
      { key: 'nav:categories', icon: <TagsOutlined />, label: '分类', onClick: () => navigate('/categories') },
      { key: 'nav:settings', icon: <SettingOutlined />, label: '系统设置', onClick: () => navigate('/system-settings') },
    ];
  }, [isStoreAdmin, navigate]);

  return (
    <Layout>
      <Sider breakpoint="lg" collapsedWidth="0">
        <div style={{ height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography.Title level={4} style={{ color: '#fff', margin: 0 }}>
            茶心阁
          </Typography.Title>
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[selectedKey]} defaultOpenKeys={openKeys}>
          {!isStoreAdmin && (
            <>
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
              <Menu.SubMenu key="system-settings" icon={<SettingOutlined />} title="系统设置">
                <Menu.Item key="/system-settings/basic">
                  <Link to="/system-settings">基础配置</Link>
                </Menu.Item>
                <Menu.Item key="/system-settings/admins">
                  <Link to="/system-settings?tab=admins">管理员</Link>
                </Menu.Item>
              </Menu.SubMenu>

              <Menu.SubMenu key="products" icon={<ShopOutlined />} title="商品管理">
                <Menu.Item key="/products/categories" icon={<TagsOutlined />}>
                  <Link to="/categories">分类</Link>
                </Menu.Item>
                <Menu.Item key="/products/on">
                  <Link to="/products?tab=on">上架中</Link>
                </Menu.Item>
                <Menu.Item key="/products/off">
                  <Link to="/products?tab=off">已下架</Link>
                </Menu.Item>
                <Menu.Item key="/products/draft">
                  <Link to="/products?tab=draft">草稿</Link>
                </Menu.Item>
                <Menu.Item key="/products/create">
                  <Link to="/products?tab=create">新增</Link>
                </Menu.Item>
                <Menu.Item key="/products/all">
                  <Link to="/products">全部</Link>
                </Menu.Item>
              </Menu.SubMenu>
            </>
          )}

          {isStoreAdmin ? (
            <>
              <Menu.Item key="/store-home" icon={<DashboardOutlined />}>
                <Link to="/store-home">首页</Link>
              </Menu.Item>
              <Menu.Item key="/store-mall" icon={<ShopOutlined />}>
                <Link to="/store-mall">商家商城</Link>
              </Menu.Item>
              <Menu.Item key="/store-orders" icon={<ProfileOutlined />}>
                <Link to="/store-orders">订单管理</Link>
              </Menu.Item>
              <Menu.Item key="/store-refunds" icon={<ProfileOutlined />}>
                <Link to="/store-refunds">售后订单</Link>
              </Menu.Item>
              <Menu.Item key="/store-products" icon={<ShopOutlined />}>
                <Link to="/store-products">商品管理</Link>
              </Menu.Item>
              <Menu.Item
                key="/store-finance"
                icon={<AccountBookOutlined />}
                title="门店资金流水（支付/退款/提现）与提现管理"
              >
                <Link to="/store-finance">
                  财务管理 {showNewBadge && (
                    <Tag color="green" style={{ marginLeft: 8, fontSize: 12, lineHeight: '16px' }}>
                      新
                    </Tag>
                  )}
                </Link>
              </Menu.Item>
              <Menu.Item key="/store-accounts" icon={<CreditCardOutlined />}>
                <Link to="/store-accounts">收款账户</Link>
              </Menu.Item>
              <Menu.Item key="/store-coupons" icon={<GiftOutlined />}>
                <Link to="/store-coupons">优惠券</Link>
              </Menu.Item>
              <Menu.Item key="/store-activities" icon={<FireOutlined />}>
                <Link to="/store-activities">活动</Link>
              </Menu.Item>
              <Menu.Item key="/store-settings" icon={<SettingOutlined />}>
                <Link to="/store-settings">门店设置</Link>
              </Menu.Item>
            </>
          ) : (
            <>
              <Menu.Item key="/stores" icon={<EnvironmentOutlined />}>
                <Link to="/stores">门店管理</Link>
              </Menu.Item>
              <Menu.Item
                key="/store-finance"
                icon={<AccountBookOutlined />}
                title="门店资金流水（支付/退款/提现）与提现管理"
              >
                <Link to="/store-finance">
                  门店财务（资金流水） {showNewBadge && (
                    <Tag color="green" style={{ marginLeft: 8, fontSize: 12, lineHeight: '16px' }}>
                      新
                    </Tag>
                  )}
                </Link>
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
            </>
          )}

          {!isStoreAdmin && (
            <>
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
              <Menu.SubMenu key="orders" icon={<ProfileOutlined />} title="订单管理">
                <Menu.Item key="/orders/all">
                  <Link to="/orders">订单列表</Link>
                </Menu.Item>
                <Menu.Item key="/orders/after-sales">
                  <Link to="/orders?tab=after-sales">售后订单</Link>
                </Menu.Item>
              </Menu.SubMenu>
              <Menu.Item key="/dev-tools" icon={<ToolOutlined />}>
                <a href="https://github.com/jy02537333/tea#readme" target="_blank" rel="noreferrer">开发者工具</a>
              </Menu.Item>
            </>
          )}

          <Menu.Item key="logout" icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Menu.Item>
        </Menu>
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px' }}>
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space>
              <Dropdown
                menu={{
                  items: quickNavItems,
                }}
                trigger={['click']}
              >
                <Button icon={<AppstoreOutlined />}>快捷导航</Button>
              </Dropdown>
              <Input.Search
                allowClear
                value={headerSearch}
                onChange={(e) => setHeaderSearch(e.target.value)}
                onSearch={(val) => handleHeaderSearch(val)}
                placeholder={isStoreAdmin ? '搜索订单号' : '搜索订单号/商品名'}
                style={{ width: 320 }}
                enterButton={<SearchOutlined />}
              />
            </Space>

            <Space>
              <Badge count={0} size="small">
                <Button
                  icon={<BellOutlined />}
                  onClick={() => message.info('暂无消息')}
                >
                  消息
                </Button>
              </Badge>

              <Dropdown
                trigger={['click']}
                menu={{
                  items: [
                    {
                      key: 'account:info',
                      icon: <UserOutlined />,
                      label: '账户中心',
                      onClick: () => message.info('账户中心：当前仅展示基础信息'),
                    },
                    { type: 'divider' },
                    {
                      key: 'account:logout',
                      icon: <LogoutOutlined />,
                      label: '退出登录',
                      onClick: handleLogout,
                    },
                  ],
                }}
              >
                <Button>
                  <Space>
                    <Avatar size={24} icon={<UserOutlined />} />
                    <span>{user?.nickname || user?.username || `#${user?.id ?? ''}`}</span>
                  </Space>
                </Button>
              </Dropdown>
            </Space>
          </Space>
        </Header>
        <Content style={{ margin: 24 }}>
          <div style={{ padding: 24, background: '#fff', minHeight: 'calc(100vh - 160px)' }}>{children}</div>
        </Content>
      </Layout>
    </Layout>
  );
}
