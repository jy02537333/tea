import { Navigate } from 'react-router-dom';
import { Alert } from 'antd';
import { useAuthContext } from '../hooks/useAuth';

export default function StoreProductsRedirectPage() {
  const { user } = useAuthContext();
  const lockedStoreId = user?.store_id;

  if (user?.role !== 'store') return <Navigate to="/stores" replace />;
  if (!lockedStoreId) return <Alert type="error" showIcon message="门店管理员未绑定门店（store_admins）" />;

  return <Navigate to={`/stores/${lockedStoreId}/products`} replace />;
}
