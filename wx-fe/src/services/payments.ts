import api, { unwrapResponse } from './api';

export interface UnifiedOrderResult {
	payment_no: string;
	order_id: number;
	amount: string | number;
	prepay_id: string;
	nonce_str: string;
	timestamp: number;
	package: string;
	sign: string;
	pay_url?: string;
}

export async function createUnifiedOrder(orderId: number): Promise<UnifiedOrderResult> {
	const res = await api.post('/api/v1/payments/unified-order', {
		order_id: orderId,
		method: 1,
		trade_type: 'JSAPI',
		notify_url: '',
	});
	return unwrapResponse<UnifiedOrderResult>(res);
}

export async function mockPayCallback(paymentNo: string): Promise<any> {
	const res = await api.post('/api/v1/payment/mock-callback', { payment_no: paymentNo });
	return unwrapResponse<any>(res);
}
