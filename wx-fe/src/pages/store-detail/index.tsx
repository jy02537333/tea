import React, { useEffect, useMemo, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { mockStores, StoreService } from '../../services/mockData';
import './index.scss';

export default function StoreDetailPage() {
  const [currentStoreName, setCurrentStoreName] = useState('');
  const router = Taro.getCurrentInstance().router;
  const storeIdParam = router?.params?.store_id;
  const storeNameParam = router?.params?.store_name;

  useEffect(() => {
    try {
      const raw = Taro.getStorageSync('current_store_name');
      let saved = '';
      if (typeof raw === 'string') {
        saved = raw;
        if (raw.trim().startsWith('{')) {
          try {
            const parsed = JSON.parse(raw);
            if (parsed && typeof parsed.data === 'string') saved = parsed.data;
          } catch (_) {}
        }
      } else if (raw && typeof raw === 'object' && typeof (raw as any).data === 'string') {
        saved = String((raw as any).data);
      }
      saved = String(saved || '').trim();
      if (saved) setCurrentStoreName(saved);
    } catch (_) {}
  }, []);

  const currentStore = useMemo(() => {
    if (storeIdParam) {
      const byId = mockStores.find((store) => store.id === String(storeIdParam));
      if (byId) return byId;
    }
    if (storeNameParam) {
      const byName = mockStores.find((store) => store.name === String(storeNameParam));
      if (byName) return byName;
    }
    if (currentStoreName) {
      const byStorage = mockStores.find((store) => store.name === currentStoreName);
      if (byStorage) return byStorage;
    }
    return mockStores[0];
  }, [storeIdParam, storeNameParam, currentStoreName]);

  const serviceFilters = useMemo(() => {
    const types = Array.from(new Set(currentStore.services.map((s) => s.type)));
    const labelMap: Record<StoreService['type'], string> = {
      tea: '茶席',
      space: '空间',
      facility: '设施',
      office: '办公'
    };
    return [
      { id: 'all', label: '全部' },
      ...types.map((type) => ({ id: type, label: labelMap[type] }))
    ];
  }, [currentStore.services]);

  const [activeServiceFilter, setActiveServiceFilter] = useState('all');

  useEffect(() => {
    setActiveServiceFilter('all');
  }, [currentStore.id]);

  const filteredServices = useMemo(() => {
    if (activeServiceFilter === 'all') return currentStore.services;
    return currentStore.services.filter((service) => service.type === activeServiceFilter);
  }, [activeServiceFilter, currentStore.services]);

  const isCurrent = currentStoreName && currentStore.name === currentStoreName;
  const statusText = currentStore.status === 'open' ? '营业中' : currentStore.status === 'busy' ? '繁忙中' : '休息中';
  const statusTagClass = currentStore.status === 'closed' ? 'store-tag-pill--closed' : 'store-tag-pill--status';

  return (
    <View className="page-store-detail">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back"><Text>‹</Text></View>
            <Text className="nav-title">门店详情</Text>
          </View>
          <Text className="nav-right">{currentStore.name}</Text>
        </View>

        <View className="scroll">
          <View className="store-header-card">
            <View className="store-header-top">
              <View>
                <Text className="store-name">{currentStore.name}</Text>
                <Text className="store-distance">{currentStore.distance} · {currentStore.meta}</Text>
              </View>
              <View className="store-tags">
                <Text className={`store-tag-pill ${statusTagClass}`}>{statusText}</Text>
                {isCurrent && <Text className="store-tag-pill">当前门店</Text>}
                {currentStore.supports.pickup && <Text className="store-tag-pill">支持自提</Text>}
                {currentStore.supports.takeout && <Text className="store-tag-pill">支持外卖</Text>}
                {currentStore.supports.member && <Text className="store-tag-pill">可享平台年卡</Text>}
              </View>
            </View>
            <View className="store-ops">
              <View
                className="btn-small"
                onClick={() => Taro.navigateTo({ url: '/pages/stores/index' }).catch(() => {})}
              >
                <Text>切换门店</Text>
              </View>
              <View className="btn-small btn-small--primary">
                <Text>呼叫门店</Text>
              </View>
            </View>
          </View>

          <View className="map-card">
            <View className="map-placeholder">
              <View className="map-pin"><Text>📍</Text></View>
            </View>
            <View className="map-footer">
              <View className="map-footer-main">
                <Text className="map-footer-title">{currentStore.mapTitle}</Text>
                <Text className="map-footer-sub">{currentStore.mapSub}</Text>
              </View>
              <View className="map-footer-btn"><Text>在地图中查看</Text></View>
            </View>
          </View>

          <View className="info-card">
            <View className="info-title-row">
              <Text className="info-title">营业时间</Text>
              <Text className="info-sub">Business Hours</Text>
            </View>
            <Text className="info-line"><Text className="info-label">营业时间</Text>{currentStore.businessHours}</Text>
            <Text className="info-line"><Text className="info-label">节假日</Text>{currentStore.holidayNote}</Text>
          </View>

          <View className="info-card">
            <View className="info-title-row">
              <Text className="info-title">门店地址</Text>
              <Text className="info-sub">Address</Text>
            </View>
            <Text className="info-line">{currentStore.address}</Text>
            {currentStore.addressHint && (
              <Text className="info-line"><Text className="info-label">停车</Text>{currentStore.addressHint}</Text>
            )}
          </View>

          <View className="info-card">
            <View className="info-title-row">
              <Text className="info-title">联系方式</Text>
              <Text className="info-sub">Contact</Text>
            </View>
            <Text className="info-line"><Text className="info-label">门店电话</Text>{currentStore.phone}</Text>
            <Text className="info-line"><Text className="info-label">微信客服</Text>可在「我的 - 联系客服」中与茶心君聊聊</Text>
          </View>

          <View className="info-card">
            <View className="info-title-row">
              <Text className="info-title">门店服务</Text>
              <Text className="info-sub">Services</Text>
            </View>
            <View className="filter-row">
              {serviceFilters.map((filter) => (
                <Text
                  key={filter.id}
                  className={`filter-pill ${activeServiceFilter === filter.id ? 'filter-pill--active' : ''}`}
                  onClick={() => setActiveServiceFilter(filter.id)}
                >
                  {filter.label}
                </Text>
              ))}
            </View>
            <View className="service-tags">
              {filteredServices.map((service) => (
                <Text className="service-tag" key={service.label}>{service.label}</Text>
              ))}
            </View>
          </View>

          <View className="info-card">
            <View className="info-title-row">
              <Text className="info-title">温馨提示</Text>
              <Text className="info-sub">Notice</Text>
            </View>
            <Text className="notice-text">
              {currentStore.notices.map((note) => `· ${note}`).join('\n')}
            </Text>
          </View>
        </View>

        <View className="footer">
          <Text>茶心阁 · 小程序「门店详情」页面静态稿 · 仅作示意</Text>
        </View>
      </View>
    </View>
  );
}
