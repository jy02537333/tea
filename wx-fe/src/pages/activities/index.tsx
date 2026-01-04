import React, { useEffect, useState } from 'react';
import { View, Text, Button, Input } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listActivities, registerActivityWithOrder } from '../../services/activities';
import { createUnifiedOrder, mockPayCallback } from '../../services/payments';
import { getStore } from '../../services/stores';
import type { Activity, Order, Store } from '../../services/types';

export default function ActivitiesPage() {
	const [storeId, setStoreId] = useState<number | undefined>(undefined);
	const [activities, setActivities] = useState<Activity[]>([]);
	const [loading, setLoading] = useState(false);
	const [name, setName] = useState('');
	const [phone, setPhone] = useState('');
	const [submittingId, setSubmittingId] = useState<number | null>(null);
	const [fee, setFee] = useState('');

	useEffect(() => {
		const router = Taro.getCurrentInstance().router;
		const storeIdParam = router?.params?.store_id;
		let id: number | undefined;
		if (storeIdParam) {
			const parsed = Number(storeIdParam);
			if (!Number.isNaN(parsed) && parsed > 0) id = parsed;
		}
		if (!id) {
			try {
				const v = Taro.getStorageSync('current_store_id');
				const n = Number(v);
				if (Number.isFinite(n) && n > 0) id = n;
			} catch (_) {}
		}
		if (id) {
			setStoreId(id);
			void fetchActivities(id);
			void fetchStoreInfo(id);
		} else {
			Taro.showToast({ title: '缺少门店信息，请从门店列表进入', icon: 'none' });
		}
	}, []);

	async function fetchActivities(id: number) {
		setLoading(true);
		try {
			const res = await listActivities({ store_id: id });
			let items: Activity[] = [];
			const maybe: any = res;
			if (Array.isArray(maybe?.data)) items = maybe.data as Activity[];
			else if (Array.isArray(maybe?.items)) items = maybe.items as Activity[];
			else if (Array.isArray(maybe)) items = maybe as Activity[];
			setActivities(items);
		} catch (e) {
			console.error('load activities failed', e);
			Taro.showToast({ title: '活动加载失败', icon: 'none' });
		} finally {
			setLoading(false);
		}
	}

	const [currentStore, setCurrentStore] = useState<Store | null>(null);
	async function fetchStoreInfo(id: number) {
		try {
			const s = await getStore(id);
			setCurrentStore(s as Store);
		} catch (_) {}
	}

	async function handleRegister(activityId: number) {
		if (!storeId) {
			Taro.showToast({ title: '缺少门店信息', icon: 'none' });
			return;
		}
		if (!name.trim()) {
			Taro.showToast({ title: '请输入姓名', icon: 'none' });
			return;
		}
		if (!phone.trim()) {
			Taro.showToast({ title: '请输入手机号', icon: 'none' });
			return;
		}
		const feeNum = fee.trim() ? Number(fee.trim()) : 0;
		if (Number.isNaN(feeNum) || feeNum < 0) {
			Taro.showToast({ title: '报名费用格式错误', icon: 'none' });
			return;
		}
		setSubmittingId(activityId);
		try {
			const res = await registerActivityWithOrder(activityId, {
				name: name.trim(),
				phone: phone.trim(),
				fee: feeNum,
			});
			const order: Order | undefined = res?.order as any;
			Taro.showToast({ title: '报名成功', icon: 'success' });
			if (order && Number(order.pay_amount) > 0) {
				try {
					const payRes = await createUnifiedOrder(order.id);
					if (payRes?.payment_no) {
						await mockPayCallback(payRes.payment_no);
						Taro.showToast({ title: '支付成功', icon: 'success' });
					}
				} catch (err) {
					console.error('pay activity order failed', err);
					Taro.showToast({ title: '支付失败，请稍后重试', icon: 'none' });
				}
			}
		} catch (e: any) {
			console.error('register activity failed', e);
			const msg = e?.response?.data?.message || '报名失败，请稍后重试';
			Taro.showToast({ title: msg, icon: 'none' });
		} finally {
			setSubmittingId(null);
		}
	}

	return (
		<View style={{ padding: 12 }}>
			{currentStore && (
				<View style={{
					marginBottom: 8,
					padding: '6px 10px',
					borderWidth: 1,
					borderStyle: 'solid',
					borderColor: '#07c160',
					borderRadius: 16,
					display: 'inline-block',
					backgroundColor: '#f6ffed',
				}}>
					<Text style={{ color: '#389e0d' }}>当前门店：{currentStore.name}</Text>
				</View>
			)}
			<View style={{ marginBottom: 8 }}>
				<Button size="mini" onClick={() => Taro.navigateTo({ url: '/pages/stores/index' })}>切换门店</Button>
			</View>
			<Text style={{ fontSize: 18, fontWeight: 'bold' }}>活动报名</Text>
			<View style={{ marginTop: 12 }}>
				<Text>姓名</Text>
				<Input
					placeholder="请输入姓名"
					value={name}
					onInput={(e) => setName((e.detail as any).value)}
				/>
			</View>
			<View style={{ marginTop: 12 }}>
				<Text>手机号</Text>
				<Input
					placeholder="请输入手机号"
					value={phone}
					onInput={(e) => setPhone((e.detail as any).value)}
				/>
			</View>
			<View style={{ marginTop: 12 }}>
				<Text>报名费用（元，可选）</Text>
				<Input
					placeholder="不填则视为免费活动"
					value={fee}
					onInput={(e) => setFee((e.detail as any).value)}
				/>
			</View>

			<View style={{ marginTop: 16 }}>
				{loading && <Text>活动加载中...</Text>}
				{!loading && activities.length === 0 && <Text>当前门店暂无可报名活动</Text>}
				{activities.map((act) => (
					<View
						key={act.id}
						style={{
							marginTop: 12,
							padding: 12,
							borderRadius: 8,
							borderWidth: 1,
							borderStyle: 'solid',
							borderColor: '#eee',
						}}
					>
						<Text style={{ fontSize: 16, fontWeight: 'bold' }}>{act.name}</Text>
						<View style={{ marginTop: 4 }}>
							<Text>
								时间：{act.start_time} - {act.end_time}
							</Text>
						</View>
						{act.description && (
							<View style={{ marginTop: 4 }}>
								<Text>说明：{act.description}</Text>
							</View>
						)}
						<Button
							style={{ marginTop: 8 }}
							type="primary"
							loading={submittingId === act.id}
							onClick={() => handleRegister(act.id)}
						>
							{submittingId === act.id ? '报名中...' : '报名参加'}
						</Button>
					</View>
				))}
			</View>
		</View>
	);
}
