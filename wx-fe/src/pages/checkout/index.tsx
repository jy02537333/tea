import React, { useEffect, useState } from 'react';
import { View, Text, Input, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { listCart, removeCartItem } from '../../services/cart';
import { createOrderFromCart, getAvailableCouponsForOrder } from '../../services/orders';
import { createUnifiedOrder, mockPayCallback } from '../../services/payments';
import { CartItem, UserCoupon, Store, Product } from '../../services/types';
import { getStore } from '../../services/stores';
import { formatAddress, loadDefaultAddress, saveDefaultAddress } from '../../utils/address';
import { buildOrderShareAttributionParams } from '../../services/shareAttribution';
import { getProducts } from '../../services/products';

interface CheckoutViewItem extends CartItem {
  product?: Product;
}

export default function CheckoutPage() {
  const [items, setItems] = useState<CheckoutViewItem[]>([]);
  const [totalAmount, setTotalAmount] = useState(0);
  const [address, setAddress] = useState('');
  const [remark, setRemark] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [couponSummary, setCouponSummary] = useState('暂不使用优惠券');
  const [selectedUserCouponId, setSelectedUserCouponId] = useState<number | undefined>(undefined);
  const [availableCoupons, setAvailableCoupons] = useState<UserCoupon[]>([]);
  const [showCouponList, setShowCouponList] = useState(false);
  const [currentStore, setCurrentStore] = useState<Store | null>(null);
  const isStoreOrder = !!currentStore;
  const [storeMismatchCount, setStoreMismatchCount] = useState(0);
  const [cleaning, setCleaning] = useState(false);

  useEffect(() => {
    void fetchCart();
    void preloadAddress();
    void loadCurrentStore();
  }, []);

  async function loadCurrentStore() {
    try {
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const storeId = storeIdRaw ? Number(storeIdRaw) : NaN;
      if (!Number.isNaN(storeId) && storeId > 0) {
        const s = await getStore(storeId);
        setCurrentStore(s as Store);
      }
    } catch (_) {
      // ignore
    }
  }

  async function fetchCart() {
    try {
      const raw = await listCart();
      const cartItemsAny: any[] = (raw as any) || [];
      // 若存在门店上下文，仅用于本地校验不混入其他门店/平台商品
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const maybeStoreId = storeIdRaw ? Number(storeIdRaw) : undefined;

      let allowed: Set<number> | null = null;
      if (maybeStoreId && Number.isFinite(maybeStoreId) && maybeStoreId > 0) {
        try {
          const productsRes = await getProducts({ page: 1, limit: 1000, store_id: maybeStoreId });
          const maybe: any = productsRes;
          const storeProducts: Product[] = Array.isArray(maybe?.data)
            ? maybe.data
            : Array.isArray(maybe?.items)
            ? maybe.items
            : Array.isArray(maybe)
            ? maybe
            : [];
          allowed = new Set<number>(storeProducts.map((p: any) => Number(p.id)));
        } catch (_) {
          allowed = null; // 校验失败时不阻断渲染
        }
      }

      if (allowed) {
        const invalid = cartItemsAny.filter((c) => !allowed!.has(Number(c.product_id)));
        setStoreMismatchCount(invalid.length);
      } else {
        setStoreMismatchCount(0);
      }

      const merged: CheckoutViewItem[] = cartItemsAny.map((c: any) => {
        const embeddedProduct: Product | undefined = (c && (c.Product || c.product)) as Product | undefined;
        return {
          ...(c as CartItem),
          product: embeddedProduct,
        } as CheckoutViewItem;
      });

      setItems(merged);
      const sum = merged.reduce((acc: number, it: any) => {
        const ep = Number(it?.effective_price ?? (it.product?.price as any) ?? 0) || 0;
        return acc + ep * it.quantity;
      }, 0);
      setTotalAmount(sum);
      void refreshAvailableCoupons(sum);
    } catch (e) {
      console.error('load cart for checkout failed', e);
    }
  }

  async function refreshAvailableCoupons(sum: number) {
    try {
      if (!sum || sum <= 0) {
        setCouponSummary('暂无可用优惠券');
        setSelectedUserCouponId(undefined);
        setAvailableCoupons([]);
        return;
      }
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const maybeStoreId = storeIdRaw ? Number(storeIdRaw) : undefined;
      const payload: { order_amount: string; store_id?: number } = { order_amount: String(sum) };
      if (maybeStoreId && Number.isFinite(maybeStoreId) && maybeStoreId > 0) {
        payload.store_id = maybeStoreId;
      }
      const data = await getAvailableCouponsForOrder(payload);
      const list = data.available || [];
      setAvailableCoupons(list);
      if (list.length > 0) {
        const first = list[0];
        const label = buildCouponLabel(first);
        setSelectedUserCouponId(first.id);
        setCouponSummary(`已自动选择：${label}`);
      } else {
        setSelectedUserCouponId(undefined);
        setCouponSummary('暂无可用优惠券');
      }
    } catch (error) {
      console.error('load available coupons for order failed', error);
      setCouponSummary('优惠券加载失败，可稍后重试');
      setSelectedUserCouponId(undefined);
      setAvailableCoupons([]);
    }
  }

  async function preloadAddress() {
    try {
      const stored = await loadDefaultAddress();
      if (stored) {
        const fallback = stored.full || formatAddress(stored);
        if (fallback && !address) {
          setAddress(fallback);
        }
      }
    } catch (error) {
      console.error('load default address failed', error);
    }
  }

  function calcTotal() {
    return Number(totalAmount || 0);
  }

  function coverUrl(p?: Product): string {
    if (!p) return 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
    const raw = (p as any).image_url || (p as any).cover || p.images || '';
    const first = String(raw)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)[0];
    return first || 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
  }

  function coverUrlFromItem(it: any): string {
    // 优先商品图，其次SKU图，最后兜底
    const p = (it && (it.product || it.Product)) as Product | undefined;
    const byProduct = coverUrl(p);
    if (byProduct && !byProduct.includes('dummyimage')) return byProduct;
    const skuImg = (it && ((it.sku && it.sku.image) || (it.Sku && it.Sku.image))) || '';
    if (skuImg && String(skuImg).trim()) return String(skuImg).trim();
    return 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
  }

  // 清理非本店/平台商品，仅保留当前门店条目
  async function cleanToCurrentStore() {
    if (!currentStore) {
      Taro.showToast({ title: '未选择门店', icon: 'none', duration: 1500 });
      return;
    }
    if (!items.length) {
      Taro.showToast({ title: '购物车为空', icon: 'none', duration: 1500 });
      return;
    }
    setCleaning(true);
    try {
      const res = await getProducts({ page: 1, limit: 500, store_id: currentStore.id });
      const maybe: any = res;
      const storeProducts: Product[] = Array.isArray(maybe?.data)
        ? maybe.data
        : Array.isArray(maybe?.items)
        ? maybe.items
        : Array.isArray(maybe)
        ? maybe
        : [];
      const allowed = new Set<number>(storeProducts.map((p: any) => Number(p.id)));
      const toRemove = items.filter((it) => !allowed.has(Number(it.product_id)));
      if (!toRemove.length) {
        Taro.showToast({ title: '已全部为本店商品', icon: 'success', duration: 1200 });
        setStoreMismatchCount(0);
        return;
      }
      const mod = await Taro.showModal({
        title: '清理非本店商品',
        content: `将移除 ${toRemove.length} 件非本店/平台商品，是否继续？`,
        confirmText: '移除',
      } as any);
      if (!(mod as any)?.confirm) return;
      for (const it of toRemove) {
        try { await removeCartItem(it.id); } catch (_) { /* 忽略单个失败 */ }
      }
      await fetchCart();
      Taro.showToast({ title: '已清理', icon: 'success', duration: 1200 });
    } catch (e) {
      console.error('clean to current store failed', e);
      Taro.showToast({ title: '操作失败', icon: 'none', duration: 1500 });
    } finally {
      setCleaning(false);
    }
  }

  function buildCouponLabel(userCoupon: UserCoupon): string {
    const coupon = userCoupon.coupon;
    let label = coupon.name;
    if (coupon.type === 1 && coupon.amount) {
      label = `${coupon.name} - 满减¥${coupon.amount}`;
    } else if (coupon.type === 2 && coupon.discount) {
      label = `${coupon.name} - 折扣${coupon.discount}`;
    }
    return label;
  }

  async function handleSubmit() {
    if (!items.length) {
      Taro.showToast({ title: '购物车为空', icon: 'none', duration: 1500 });
      return;
    }
    setSubmitting(true);
    try {
      const storeIdRaw = Taro.getStorageSync('current_store_id');
      const maybeStoreId = storeIdRaw ? Number(storeIdRaw) : undefined;
      const tableIdRaw = Taro.getStorageSync('current_table_id');
      const maybeTableId = tableIdRaw ? Number(tableIdRaw) : undefined;
      const tableNoRaw = Taro.getStorageSync('current_table_no');
      const maybeTableNo = tableNoRaw ? String(tableNoRaw).trim() : undefined;
      const hasTable = (!!maybeTableNo) || (!!maybeTableId && Number.isFinite(maybeTableId) && maybeTableId > 0);
      const shareParams = buildOrderShareAttributionParams({ storeId: maybeStoreId, requireStoreId: true });
      const payload = {
        delivery_type: hasTable ? 1 : 2, // 简化：有桌号则按堂食/自取
        // 门店堂食/自取无需收货地址
        address_info: isStoreOrder ? undefined : (address || undefined),
        remark: remark || undefined,
        user_coupon_id: selectedUserCouponId,
        store_id: maybeStoreId && Number.isFinite(maybeStoreId) && maybeStoreId > 0 ? maybeStoreId : undefined,
        order_type: hasTable ? 2 : undefined,
        table_id: (maybeTableId && Number.isFinite(maybeTableId) && maybeTableId > 0) ? maybeTableId : undefined,
        table_no: maybeTableNo || undefined,
		...shareParams,
      };
      const order = await createOrderFromCart(payload as any);
      if (address.trim()) {
        await saveDefaultAddress({
          full: address.trim(),
          detail: address.trim(),
          orderId: (order as any)?.id,
          orderNo: (order as any)?.order_no,
          updatedAt: new Date().toISOString(),
          timestamp: Date.now(),
        });
      }
      // 统一下单 + 支付模拟回调
      if ((order as any)?.id) {
        try {
          const payRes = await createUnifiedOrder((order as any).id);
          await mockPayCallback(payRes.payment_no);
          Taro.showToast({ title: '支付成功（模拟）', icon: 'success', duration: 1500 });
          // 跳转到订单详情查看状态
          const sid = (maybeStoreId && Number.isFinite(maybeStoreId) && maybeStoreId > 0) ? maybeStoreId : undefined;
          const url = sid
            ? `/pages/order-detail/index?id=${(order as any).id}&store_id=${sid}`
            : `/pages/order-detail/index?id=${(order as any).id}`;
          Taro.navigateTo({ url });
        } catch (payErr) {
          console.error('unified order or callback failed', payErr);
          Taro.showToast({ title: '下单成功，支付模拟失败', icon: 'none', duration: 2000 });
          // 即使支付失败也允许查看订单详情
          const sid = (maybeStoreId && Number.isFinite(maybeStoreId) && maybeStoreId > 0) ? maybeStoreId : undefined;
          const url = sid
            ? `/pages/order-detail/index?id=${(order as any).id}&store_id=${sid}`
            : `/pages/order-detail/index?id=${(order as any).id}`;
          Taro.navigateTo({ url });
        }
      } else {
        Taro.showToast({ title: '下单成功', icon: 'success', duration: 1500 });
      }
    } catch (e: any) {
      console.error('create order failed', e);
      const msg = (e && e.response && e.response.data && ((e.response.data as any).message || (e.response.data as any).msg))
        || e?.message
        || '下单失败';
      Taro.showToast({ title: String(msg).slice(0, 24), icon: 'none', duration: 2000 });
    } finally {
      setSubmitting(false);
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
      <Text style={{ fontSize: 12, color: '#999', marginBottom: 8 }}>
        温馨提示：如购物车含平台或其他门店商品，请先回到购物车清空或按门店分单提交（当前不支持自动拆单）。
      </Text>
      {isStoreOrder && storeMismatchCount > 0 && (
        <View style={{
          marginBottom: 8,
          padding: '8px 10px',
          borderWidth: 1,
          borderStyle: 'solid',
          borderColor: '#faad14',
          borderRadius: 8,
          backgroundColor: '#fffbe6',
        }}>
          <Text style={{ color: '#ad6800' }}>
            购物车包含平台或其他门店商品 {storeMismatchCount} 件，当前仅支持在「{currentStore?.name}」下单。
          </Text>
          <View style={{ marginTop: 6, display: 'flex', flexDirection: 'row', gap: 8 }}>
            <Button size="mini" onClick={() => Taro.navigateTo({ url: '/pages/cart/index' })}>回购物车处理</Button>
            <Button size="mini" type="primary" disabled={cleaning} onClick={cleanToCurrentStore}>
              {cleaning ? '处理中...' : '仅保留本店并重试'}
            </Button>
          </View>
        </View>
      )}
      <Text>确认订单（共 {items.length} 件）</Text>

      {items.length > 0 && (
        <View style={{ marginTop: 8 }}>
          {items.map((it: any) => {
            const unitPrice = Number(it?.effective_price ?? (it.product?.price as any) ?? 0) || 0;
            const lineTotal = unitPrice * it.quantity;
            return (
              <View key={it.id} style={{ display: 'flex', flexDirection: 'row', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid #eee' }}>
                <Image src={coverUrlFromItem(it)} mode="aspectFill" style={{ width: 56, height: 56, borderRadius: 4, background: '#f5f5f5' }} />
                <View style={{ marginLeft: 8, flex: 1 }}>
                  <Text style={{ display: 'block', fontSize: 14 }}>{(it.product?.name) || (it.Product?.name) || `商品#${it.product_id}`}</Text>
                  <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>单价：¥{unitPrice.toFixed(2)} × {it.quantity}</Text>
                </View>
                <Text style={{ fontWeight: 'bold' }}>¥{lineTotal.toFixed(2)}</Text>
              </View>
            );
          })}
        </View>
      )}

      <View style={{ marginTop: 8 }}>
        <Text style={{ fontWeight: 'bold' }}>金额合计：¥{calcTotal().toFixed(2)}</Text>
      </View>

      <View style={{ marginTop: 8 }}>
        <Text>优惠券：{couponSummary}</Text>
        <Text style={{ fontSize: 12, color: '#999', marginTop: 2 }}>
          可点击下方按钮查看本单可用优惠券并手动选择
        </Text>
        <View style={{ marginTop: 4, display: 'flex', flexDirection: 'row', flexWrap: 'wrap', gap: 4 }}>
          {availableCoupons.length > 0 && (
            <Button
              size="mini"
              onClick={() => setShowCouponList((prev) => !prev)}
              type="primary"
            >
              {showCouponList
                ? '收起本单可用优惠券'
                : `本单可用优惠券（${availableCoupons.length}）`}
            </Button>
          )}
          {selectedUserCouponId ? (
            <Button
              size="mini"
              onClick={() => {
                setSelectedUserCouponId(undefined);
                setCouponSummary('暂不使用优惠券');
              }}
            >
              不使用优惠券
            </Button>
          ) : null}
        </View>

        {showCouponList && availableCoupons.length > 0 && (
          <View style={{ marginTop: 8 }}>
            {availableCoupons.map((uc) => {
              const label = buildCouponLabel(uc);
              const isSelected = uc.id === selectedUserCouponId;
              return (
                <Button
                  key={uc.id}
                  size="mini"
                  type={isSelected ? 'primary' : 'default'}
                  style={{ marginRight: 8, marginBottom: 4 }}
                  onClick={() => {
                    setSelectedUserCouponId(uc.id);
                    setCouponSummary(`已选择：${label}`);
                    setShowCouponList(false);
                  }}
                >
                  {label}
                </Button>
              );
            })}
          </View>
        )}
      </View>

      {!isStoreOrder && (
        <View style={{ marginTop: 12 }}>
          <Text>收货地址</Text>
          <Input
            type="text"
            placeholder="请输入收货地址"
            value={address}
            onInput={(e) => setAddress((e.detail as any).value)}
          />
          <Text style={{ fontSize: 12, color: '#999' }}>可在“我的-收货地址”设置默认地址</Text>
        </View>
      )}
      {isStoreOrder && (
        <View style={{ marginTop: 12 }}>
          <Text style={{ fontSize: 12, color: '#389e0d' }}>门店堂食/自取，无需填写收货地址</Text>
        </View>
      )}

      <View style={{ marginTop: 12 }}>
        <Text>备注</Text>
        <Input
          type="text"
          placeholder="可填写口味、送达时间等"
          value={remark}
          onInput={(e) => setRemark((e.detail as any).value)}
        />
      </View>

      <View style={{ marginTop: 16 }}>
        <Button type="primary" disabled={submitting || (isStoreOrder && storeMismatchCount > 0)} onClick={handleSubmit}>
          {submitting ? '提交中...' : '提交订单'}
        </Button>
      </View>
    </View>
  );
}
