import api, { unwrapResponse } from './api';
import type { Activity, Order } from './types';

export async function listActivities(params: { store_id: number }): Promise<Activity[]> {
	const res = await api.get('/api/v1/activities', { params });
	return unwrapResponse<Activity[]>(res);
}

export async function registerActivity(activityId: number, payload: { name: string; phone: string }): Promise<any> {
	const res = await api.post(`/api/v1/activities/${activityId}/register`, payload);
	return unwrapResponse<any>(res);
}

export async function registerActivityWithOrder(
	activityId: number,
	payload: { name: string; phone: string; fee: number; sharer_uid?: number; share_store_id?: number },
): Promise<{ registration: any; order: Order }> {
	const res = await api.post(`/api/v1/activities/${activityId}/register-with-order`, payload);
	return unwrapResponse<{ registration: any; order: Order }>(res);
}
