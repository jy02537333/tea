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
import Accrual from './pages/Accrual';
import Users from './pages/Users';
import Rbac from './pages/Rbac';
import Orders from './pages/Orders';

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
