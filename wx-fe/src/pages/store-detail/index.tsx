import React, { useEffect, useState } from 'react';
import { View, Text, Button, Image } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { getStore } from '../../services/stores';
import { Store } from '../../services/types';
import usePermission from '../../hooks/usePermission';
import { PERM_TOAST_NO_STORE_FINANCE } from '../../constants/permission';

export default function StoreDetailPage() {
  const router = useRouter();
  const [store, setStore] = useState<Store | null>(null);
  const [loading, setLoading] = useState(false);
  const perm = usePermission();

  useEffect(() => {
    void loadStore();
  }, []);

  async function loadStore() {
    const rawId = router?.params?.store_id;
    const id = rawId ? Number(rawId) : NaN;
    if (Number.isNaN(id) || id <= 0) return;
    setLoading(true);
    try {
      const s = await getStore(id);
      setStore(s as Store);
      try { Taro.setStorageSync('current_store_id', String(id)); } catch (_) {}
    } catch (e) {
      console.error('load store failed', e);
      Taro.showToast({ title: '门店加载失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  const allowedAccounts = perm.allowedStoreAccounts;
  const allowedFinance = perm.allowedStoreFinance;

  function goNavigate() {
    if (!store) return;
    const lat = store.latitude;
    const lng = store.longitude;
    if (typeof lat === 'number' && typeof lng === 'number') {
      Taro.openLocation({ latitude: lat, longitude: lng, name: store.name, address: store.address || '' }).catch(() => {
        Taro.showToast({ title: '导航打开失败', icon: 'none' });
      });
    } else {
      Taro.showToast({ title: '该门店暂未提供定位信息', icon: 'none' });
    }
  }

  function goStoreProducts() {
    if (!store) return;
    const sid = store.id;
    Taro.navigateTo({ url: `/pages/category/index?store_id=${sid}` });
  }

  function goDial() {
    if (!store?.phone) {
      Taro.showToast({ title: '暂无联系电话', icon: 'none' });
      return;
    }
    Taro.makePhoneCall({ phoneNumber: store.phone }).catch(() => {
      Taro.showToast({ title: '拨号失败', icon: 'none' });
    });
  }

  function getLicenseUrls(): string[] {
    if (!store) return [];
    const urls: string[] = [];
    if (Array.isArray(store.licenses)) {
      for (const item of store.licenses) {
        if (typeof item === 'string') urls.push(item);
        else if (item && typeof item.url === 'string') urls.push(item.url);
      }
    }
    if (Array.isArray(store.license_images)) {
      for (const u of store.license_images) if (typeof u === 'string') urls.push(u);
    }
    return urls;
  }

  function previewLicense(url: string) {
    const urls = getLicenseUrls();
    if (!urls.length) return;
    Taro.previewImage({ current: url, urls }).catch(() => {});
  }

  if (loading && !store) return <Text>加载中...</Text>;
  if (!store) return <Text>未找到门店信息</Text>;

  return (
    <View style={{ padding: 12 }}>
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
        <Text style={{ color: '#389e0d' }}>当前门店：{store.name}</Text>
      </View>

      <View style={{ marginTop: 8 }}>
        <Text style={{ display: 'block' }}>地址：{store.address || '未设置'}</Text>
        <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>如提供定位信息，可一键导航</Text>
      </View>

      <Text style={{ display: 'block', color: '#999', fontSize: 12, marginTop: 6 }}>
        提示：可通过右上角入口查看财务流水或管理收款账户
      </Text>

      {store.phone && (
        <View style={{ marginTop: 8 }}>
          <Text style={{ display: 'block' }}>联系电话：{store.phone}</Text>
          <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>可直接拨打联系门店</Text>
        </View>
      )}

      <View style={{ marginTop: 12, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <Button size="mini" type="primary" onClick={goNavigate}>导航到门店</Button>
        <Button size="mini" onClick={goDial}>拨打电话</Button>
        <Button size="mini" onClick={goStoreProducts}>查看本店商品</Button>
        {store && (
          <Button size="mini" onClick={() => Taro.navigateTo({ url: `/pages/activities/index?store_id=${store.id}` })}>查看活动/报名</Button>
        )}
        {allowedAccounts && store && (
          <>
            <Button size="mini" onClick={() => Taro.navigateTo({ url: `/pages/store-accounts/index?store_id=${store.id}` })}>管理收款账户</Button>
          </>
        )}
        {store && (
          <>
            <Button
              size="mini"
              onClick={() => {
                if (!allowedFinance) {
                  Taro.showToast({ title: PERM_TOAST_NO_STORE_FINANCE, icon: 'none' });
                  return;
                }
                Taro.navigateTo({ url: `/pages/store-finance/index?store_id=${store.id}` });
              }}
            >
              查看财务流水
            </Button>
            {!allowedFinance && (
              <Text style={{ color: '#999', fontSize: 12 }}>（需权限）</Text>
            )}
          </>
        )}
      </View>

      {(() => {
        const licenseUrls = getLicenseUrls();
        if (!licenseUrls.length) return null;
        return (
          <View style={{ marginTop: 16 }}>
            <Text style={{ display: 'block', fontWeight: 'bold' }}>门店证照</Text>
            <View style={{ marginTop: 8, display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {licenseUrls.map((u, idx) => (
                <View key={`${u}-${idx}`} onClick={() => previewLicense(u)}>
                  <Image src={u} style={{ width: '120px', height: '90px', borderRadius: 6 }} mode="aspectFill" />
                </View>
              ))}
            </View>
          </View>
        );
      })()}
    </View>
  );
}
