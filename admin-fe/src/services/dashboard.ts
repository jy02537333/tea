import { api, unwrap } from './api';

export interface DashboardTodos {
  ticket_pending_count: number;
  order_to_ship_count: number;
  withdraw_pending_count: number;
}

export async function getDashboardTodos() {
  const res = await api.get('/api/v1/admin/dashboard/todos');
  return unwrap<DashboardTodos>(res);
}
