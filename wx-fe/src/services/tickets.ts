import api, { unwrapResponse } from './api';

export interface CreateTicketPayload {
  type: string; // consult | order | refund | recharge | complaint | other
  source: string; // miniapp_feedback | miniapp_order
  order_id?: number;
  store_id?: number;
  title: string;
  content: string;
  attachments?: string; // 预留为 JSON 字符串
}

export async function createTicket(payload: CreateTicketPayload) {
  const res = await api.post('/api/v1/tickets', payload);
  return unwrapResponse<any>(res);
}
