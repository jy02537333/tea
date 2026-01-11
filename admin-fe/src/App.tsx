import React, { useEffect, Suspense } from 'react';
import { Layout, Spin } from 'antd';
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
const AppShell = React.lazy(() => import('./components/AppShell').then(m => ({ default: m.AppShell })));
import { useAuthContext } from './hooks/useAuth';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
const Products = React.lazy(() => import('./pages/Products'));
const Categories = React.lazy(() => import('./pages/Categories'));
const Stores = React.lazy(() => import('./pages/Stores'));
const StoreOrders = React.lazy(() => import('./pages/StoreOrders'));
const StoreFinance = React.lazy(() => import('./pages/StoreFinance'));
const StoreCoupons = React.lazy(() => import('./pages/StoreCoupons'));
const StoreActivities = React.lazy(() => import('./pages/StoreActivities'));
const StoreAccounts = React.lazy(() => import('./pages/StoreAccounts'));
const StoreProducts = React.lazy(() => import('./pages/StoreProducts'));
const StoreHome = React.lazy(() => import('./pages/StoreHome'));
const StoreSettings = React.lazy(() => import('./pages/StoreSettings'));
const StoreOrdersRedirect = React.lazy(() => import('./pages/StoreOrdersRedirect'));
const StoreProductsRedirect = React.lazy(() => import('./pages/StoreProductsRedirect'));
const StoreRefunds = React.lazy(() => import('./pages/StoreRefunds'));
const StoreRefundsRedirect = React.lazy(() => import('./pages/StoreRefundsRedirect'));
const StoreMall = React.lazy(() => import('./pages/StoreMall'));
const StoreMallRedirect = React.lazy(() => import('./pages/StoreMallRedirect'));
const Tickets = React.lazy(() => import('./pages/Tickets'));
const Accrual = React.lazy(() => import('./pages/Accrual'));
const CommissionRollback = React.lazy(() => import('./pages/CommissionRollback'));
const FinanceSummary = React.lazy(() => import('./pages/FinanceSummary'));
const FinanceRecords = React.lazy(() => import('./pages/FinanceRecords'));
const Users = React.lazy(() => import('./pages/Users'));
const Rbac = React.lazy(() => import('./pages/Rbac'));
const Orders = React.lazy(() => import('./pages/Orders'));
const MembershipConfigPage = React.lazy(() => import('./pages/MembershipConfig'));
const PartnersPage = React.lazy(() => import('./pages/Partners'));
const PartnerWithdrawalsPage = React.lazy(() => import('./pages/PartnerWithdrawals'));
const LogsPage = React.lazy(() => import('./pages/Logs'));
const SystemSettingsPage = React.lazy(() => import('./pages/SystemSettings'));
const BannersPage = React.lazy(() => import('./pages/Banners'));
const RechargePage = React.lazy(() => import('./pages/Recharge'));

function ProtectedRoute({ children }: { children: JSX.Element }) {
  const { token, loading, logout } = useAuthContext();
  const navigate = useNavigate();

  useEffect(() => {
    const handleUnauthorized = () => {
      logout();
      navigate('/login', { replace: true });
    };
    window.addEventListener('unauthorized', handleUnauthorized);
    return () => window.removeEventListener('unauthorized', handleUnauthorized);
  }, [logout, navigate]);

  if (loading) return <Spin style={{ width: '100%', marginTop: 120 }} />;
  if (!token) return <Navigate to="/login" replace />;
  return children;
}

function RoleGuard({ children }: { children: JSX.Element }) {
  const { user, loading } = useAuthContext();
  const location = useLocation();

  if (loading) return <Spin style={{ width: '100%', marginTop: 120 }} />;

  // 门店管理员：仅允许访问门店相关页面（临时策略）
  if (user?.role === 'store') {
    const lockedStoreId = user.store_id;
    const path = location.pathname || '';
    const allowedPrefixes = [
      '/store-home',
      '/store-mall',
      '/store-orders',
      '/store-refunds',
      '/store-products',
      '/store-settings',
      '/store-finance',
      '/store-accounts',
      '/store-coupons',
      '/store-activities',
      '/stores',
    ];
    const isAllowed = allowedPrefixes.some((p) => path === p || path.startsWith(p + '/'));
    if (!isAllowed) return <Navigate to="/store-finance" replace />;

    // 锁店：门店管理员不允许访问/切换到其他门店
    if (path === '/stores' || path === '/stores/') {
      if (!lockedStoreId) return <Navigate to="/store-finance" replace />;
      return <Navigate to={`/stores/${lockedStoreId}/orders`} replace />;
    }

    if (path.startsWith('/stores/')) {
      if (!lockedStoreId) return <Navigate to="/store-finance" replace />;
      const m = path.match(/^\/stores\/(\d+)(\/.*)?$/);
      if (m) {
        const requestedId = Number(m[1] || 0);
        const suffix = m[2] || '/orders';
        if (requestedId !== lockedStoreId) {
          return <Navigate to={`/stores/${lockedStoreId}${suffix}`} replace />;
        }
        // /stores/:id 这种未声明路由的路径，兜底跳转到订单页
        if (!m[2]) {
          return <Navigate to={`/stores/${lockedStoreId}/orders`} replace />;
        }
      }
    }
  }

  return children;
}

function DefaultRedirect() {
  const { user } = useAuthContext();
  if (user?.role === 'store') return <Navigate to="/store-home" replace />;
  return <Navigate to="/dashboard" replace />;
}

export default function App() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <RoleGuard>
                <Suspense fallback={<Spin style={{ width: '100%', marginTop: 120 }} />}>
                  <AppShell>
                    <Suspense fallback={<Spin style={{ width: '100%', marginTop: 120 }} />}>
                    <Routes>
                    <Route path="dashboard" element={<Dashboard />} />
                    <Route path="accrual" element={<Accrual />} />
                    <Route path="users" element={<Users />} />
                    <Route path="rbac" element={<Rbac />} />
                    <Route path="orders" element={<Orders />} />
                    <Route path="products" element={<Products />} />
                    <Route path="categories" element={<Categories />} />
                    <Route path="stores" element={<Stores />} />
                    <Route path="stores/:id/products" element={<StoreProducts />} />
                    <Route path="stores/:id/orders" element={<StoreOrders />} />
                    <Route path="stores/:id/refunds" element={<StoreRefunds />} />
                    <Route path="stores/:id/mall" element={<StoreMall />} />
                    <Route path="store-home" element={<StoreHome />} />
                    <Route path="store-mall" element={<StoreMallRedirect />} />
                    <Route path="store-orders" element={<StoreOrdersRedirect />} />
                    <Route path="store-refunds" element={<StoreRefundsRedirect />} />
                    <Route path="store-products" element={<StoreProductsRedirect />} />
                    <Route path="store-settings" element={<StoreSettings />} />
                    <Route path="store-finance" element={<StoreFinance />} />
                    <Route path="store-accounts" element={<StoreAccounts />} />
                    <Route path="store-coupons" element={<StoreCoupons />} />
                    <Route path="store-activities" element={<StoreActivities />} />
                    <Route path="tickets" element={<Tickets />} />
                    <Route path="commission-rollback" element={<CommissionRollback />} />
                    <Route path="finance-summary" element={<FinanceSummary />} />
                    <Route path="finance-records" element={<FinanceRecords />} />
                    <Route path="membership-config" element={<MembershipConfigPage />} />
                    <Route path="partners" element={<PartnersPage />} />
                    <Route path="partner-withdrawals" element={<PartnerWithdrawalsPage />} />
                    <Route path="logs" element={<LogsPage />} />
                    <Route path="system-settings" element={<SystemSettingsPage />} />
                    <Route path="banners" element={<BannersPage />} />
                    <Route path="recharge" element={<RechargePage />} />
                    <Route path="*" element={<DefaultRedirect />} />
                    </Routes>
                    </Suspense>
                  </AppShell>
                </Suspense>
              </RoleGuard>
            </ProtectedRoute>
          }
        />
      </Routes>
    </Layout>
  );
}
