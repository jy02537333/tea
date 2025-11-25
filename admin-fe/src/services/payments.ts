import api, { unwrapResponse } from './api';
import { PaymentIntent, PaymentCallbackResult } from './types';

export async function createPaymentIntent(payload: { amount: number; currency?: string }): Promise<PaymentIntent> {
  const res = await api.post('/api/v1/payment/intents', payload);
  return unwrapResponse<PaymentIntent>(res);
}

export async function mockPaymentCallback(payload: any): Promise<PaymentCallbackResult> {
  const res = await api.post('/api/v1/payment/mock-callback', payload);
  return unwrapResponse<PaymentCallbackResult>(res);
}
