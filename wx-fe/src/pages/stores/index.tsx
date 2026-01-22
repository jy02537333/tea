import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import { mockStores, StoreItem } from '../../services/mockData';
import './index.scss';

export default function StoresPage() {
  const [currentStoreName, setCurrentStoreName] = useState('隽也YUYE茶馆 · 本店');

  function syncCurrentStore() {
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
  }

  useEffect(() => {
    syncCurrentStore();
  }, []);

  useDidShow(() => {
    syncCurrentStore();
  });

  function handleSelectStore(store: StoreItem) {
    try {
      Taro.setStorageSync('current_store_name', store.name);
      Taro.setStorageSync('current_store_id', store.id);
    } catch (_) {}
    setCurrentStoreName(store.name);
    Taro.showToast({ title: '已选择门店', icon: 'none' });
    Taro.navigateTo({ url: '/pages/menu/index' }).catch(() => {});
  }

  function isSelected(name: string) {
    return currentStoreName === name;
  }

  return (
    <View className="page-stores">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back"><Text>‹</Text></View>
            <Text className="nav-title">选择门店</Text>
          </View>
          <Text className="nav-right">当前：{currentStoreName.replace(' · 本店', '')}</Text>
        </View>

        <View className="scroll">
          <Text className="location-tip">基于你当前定位，为你展示附近可自提的门店～</Text>

          {mockStores.map((store) => (
            <View className="store-card" key={store.id} onClick={() => handleSelectStore(store)}>
              <View className="store-row-top">
                <View>
                  <Text className="store-name">{store.name}</Text>
                  <Text className="store-meta">{store.meta}</Text>
                </View>
                {isSelected(store.name) && <Text className="store-tag">当前门店</Text>}
              </View>
              <Text className="store-distance">{store.distance}</Text>
              <View className="store-action-row">
                <Text>{store.actionNote}</Text>
                <View
                  className={`store-select-btn ${isSelected(store.name) ? '' : 'store-select-btn--ghost'}`}
                  onClick={() => handleSelectStore(store)}
                  data-testid="store-select-btn"
                  data-store-name={store.name}
                >
                  <Text>{isSelected(store.name) ? '已选中' : '切换到这家'}</Text>
                </View>
              </View>
            </View>
          ))}
        </View>

        <View className="footer">
          <Text>茶心阁 · 小程序「选择门店」页面静态稿 · 仅作示意</Text>
        </View>
      </View>
    </View>
  );
}
