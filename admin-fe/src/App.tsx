import { Layout, Spin } from 'antd';
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { AppShell } from './components/AppShell';
import { useAuthContext } from './hooks/useAuth';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import Products from './pages/Products';
import Categories from './pages/Categories';
import Stores from './pages/Stores';
import StoreOrders from './pages/StoreOrders';
import StoreFinance from './pages/StoreFinance';
import StoreCoupons from './pages/StoreCoupons';
import StoreActivities from './pages/StoreActivities';
import StoreAccounts from './pages/StoreAccounts';
import StoreProducts from './pages/StoreProducts';
import Tickets from './pages/Tickets';
import Accrual from './pages/Accrual';
import CommissionRollback from './pages/CommissionRollback';
import FinanceSummary from './pages/FinanceSummary';
import FinanceRecords from './pages/FinanceRecords';
import Users from './pages/Users';
import Rbac from './pages/Rbac';
import Orders from './pages/Orders';
import MembershipConfigPage from './pages/MembershipConfig';
import PartnersPage from './pages/Partners';
import PartnerWithdrawalsPage from './pages/PartnerWithdrawals';
import LogsPage from './pages/Logs';
import SystemSettingsPage from './pages/SystemSettings';
import BannersPage from './pages/Banners';
import RechargePage from './pages/Recharge';

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

export default function App() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <AppShell>
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
                  <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
              </AppShell>
            </ProtectedRoute>
          }
        />
      </Routes>
    </Layout>
  );
}
