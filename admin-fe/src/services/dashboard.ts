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

export interface OrderTrendPoint {
  date: string; // YYYY-MM-DD
  order_count: number;
  paid_order_count: number;
  sales_amount: number; // 后端以金额单位（同 pay_amount）返回
}

export async function getOrderTrends(days = 7) {
  const res = await api.get(`/api/v1/admin/dashboard/order-trends`, { params: { days } });
  return unwrap<OrderTrendPoint[]>(res);
}

export interface DashboardSummary {
  today_order_count: number;
  today_paid_order_count: number;
  today_sales_amount: number;
  today_refund_amount: number;
  yesterday_paid_order_count: number;
  yesterday_sales_amount: number;
  yesterday_refund_amount: number;
  last7d_paid_order_count: number;
  last7d_sales_amount: number;
  last7d_refund_amount: number;
  last30d_paid_order_count: number;
  last30d_sales_amount: number;
  last30d_refund_amount: number;
}

export async function getDashboardSummary() {
  const res = await api.get('/api/v1/admin/dashboard/summary');
  return unwrap<DashboardSummary>(res);
}
