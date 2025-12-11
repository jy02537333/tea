import { api, unwrap } from './api';

export interface CommissionReverseOrderResult {
  order_id: number;
  processed: number;
}

export async function reverseOrderCommission(orderId: number): Promise<CommissionReverseOrderResult> {
  const res = await api.post('/api/v1/admin/finance/commission/reverse-order', { order_id: orderId });
  return unwrap<CommissionReverseOrderResult>(res);
}

export interface FinanceSummary {
  total_payments_count: number;
  total_payments_amount: string;
  total_refunds_count: number;
  total_refunds_amount: string;
  net_amount: string;
}

export interface FinanceDailyRow {
  date: string;
  pay_count: number;
  pay_amount: string;
  refund_count: number;
  refund_amount: string;
  net_amount: string;
}

export interface FinanceStoreRow {
  store_id: number;
  store_name?: string;
  pay_count: number;
  pay_amount: string;
  refund_count: number;
  refund_amount: string;
  net_amount: string;
}

export interface FinanceSummaryResponse {
  summary: FinanceSummary;
  rows?: FinanceDailyRow[] | FinanceStoreRow[];
}

export interface FinanceSummaryParams {
  start?: string;
  end?: string;
  group?: 'day' | 'store' | 'method' | '';
  store_id?: number | string;
  method?: string;
}

export async function getFinanceSummary(params: FinanceSummaryParams): Promise<FinanceSummaryResponse> {
  const res = await api.get('/api/v1/admin/finance/summary', { params });
  return unwrap<FinanceSummaryResponse>(res);
}
