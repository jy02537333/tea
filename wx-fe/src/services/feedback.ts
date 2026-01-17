import api, { unwrapResponse } from './api';

export interface CreateFeedbackPayload {
  category?: 'consult' | 'complaint' | 'suggest';
  subject: string;
  content: string;
}

export interface CreateFeedbackResp {
  id?: number;
  status?: number;
}

// Minimal feedback create API with graceful fallback to /api/v1/tickets
export async function createFeedback(payload: CreateFeedbackPayload): Promise<CreateFeedbackResp> {
  try {
    const res = await api.post('/api/v1/feedback', payload);
    return unwrapResponse<CreateFeedbackResp>(res);
  } catch (_e) {
    const res = await api.post('/api/v1/tickets', {
      type: payload.category || 'suggest',
      source: 'miniapp_feedback',
      title: payload.subject,
      content: payload.content,
    });
    return unwrapResponse<CreateFeedbackResp>(res);
  }
}
