import { api, PaginatedResult, unwrap, unwrapPagination } from './api';

export interface Ticket {
  id: number;
  uid?: string;
  type: string;
  source: string;
  user_id?: number | null;
  order_id?: number | null;
  store_id?: number | null;
  title: string;
  content?: string;
  attachments?: string;
  status: string;
  priority: string;
  assignee_id?: number | null;
  remark?: string;
  reject_reason?: string;
  created_at?: string;
  updated_at?: string;
  resolved_at?: string | null;
  closed_at?: string | null;
}

export interface TicketListParams {
  page?: number;
  limit?: number;
  status?: string;
  type?: string;
  source?: string;
  priority?: string;
  store_id?: number;
  user_id?: number;
  keyword?: string;
}

export interface CreateTicketPayload {
  type: string;
  source: string;
  user_id?: number;
  order_id?: number;
  store_id?: number;
  title: string;
  content?: string;
  attachments?: string;
  priority?: string;
}

export interface UpdateTicketPayload {
  status?: string;
  priority?: string;
  remark?: string;
  reject_reason?: string;
  assignee_id?: number;
}

export async function listTickets(params: TicketListParams): Promise<PaginatedResult<Ticket>> {
  const res = await api.get('/api/v1/admin/tickets', { params });
  return unwrapPagination<Ticket>(res);
}

export async function getTicket(id: number): Promise<Ticket> {
  const res = await api.get(`/api/v1/admin/tickets/${id}`);
  return unwrap<Ticket>(res);
}

export async function createTicket(payload: CreateTicketPayload): Promise<Ticket> {
  const res = await api.post('/api/v1/admin/tickets', payload);
  return unwrap<Ticket>(res);
}

export async function updateTicket(id: number, payload: UpdateTicketPayload): Promise<Ticket> {
  const res = await api.put(`/api/v1/admin/tickets/${id}`, payload);
  return unwrap<Ticket>(res);
}
