import { Navigate } from 'react-router-dom';
import { useAuthContext } from '../hooks/useAuth';

export default function StoreRefundsRedirectPage() {
  const { user } = useAuthContext();
  const lockedStoreId = user?.store_id;
  if (!lockedStoreId) return <Navigate to="/store-finance" replace />;
  return <Navigate to={`/stores/${lockedStoreId}/refunds`} replace />;
}
