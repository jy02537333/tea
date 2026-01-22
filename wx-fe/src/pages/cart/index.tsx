import React, { useMemo, useState } from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { CartGroup, CartItem, mockCartGroups } from '../../services/mockData';
import './index.scss';

type ItemState = CartItem & { selected: boolean };

type GroupState = {
  storeId: string;
  storeName: string;
  items: ItemState[];
};

function initGroups(): GroupState[] {
  return mockCartGroups.map((group) => ({
    storeId: group.storeId,
    storeName: group.storeName,
    items: group.items.map((item) => ({ ...item, selected: true }))
  }));
}

export default function CartPage() {
  const [groups, setGroups] = useState<GroupState[]>(initGroups());

  function toggleGroup(groupIndex: number) {
    setGroups((prev) => prev.map((group, idx) => {
      if (idx !== groupIndex) return group;
      const allSelected = group.items.every((item) => item.selected);
      return {
        ...group,
        items: group.items.map((item) => ({ ...item, selected: !allSelected }))
      };
    }));
  }

  function toggleItem(groupIndex: number, itemId: string) {
    setGroups((prev) => prev.map((group, idx) => {
      if (idx !== groupIndex) return group;
      return {
        ...group,
        items: group.items.map((item) => (
          item.id === itemId ? { ...item, selected: !item.selected } : item
        ))
      };
    }));
  }

  function updateQty(groupIndex: number, itemId: string, delta: number) {
    setGroups((prev) => prev.map((group, idx) => {
      if (idx !== groupIndex) return group;
      return {
        ...group,
        items: group.items.map((item) => (
          item.id === itemId
            ? { ...item, quantity: Math.max(1, item.quantity + delta) }
            : item
        ))
      };
    }));
  }

  function toggleAll() {
    setGroups((prev) => {
      const allSelected = prev.every((group) => group.items.every((item) => item.selected));
      return prev.map((group) => ({
        ...group,
        items: group.items.map((item) => ({ ...item, selected: !allSelected }))
      }));
    });
  }

  const summary = useMemo(() => {
    let total = 0;
    let count = 0;
    groups.forEach((group) => {
      group.items.forEach((item) => {
        if (item.selected) {
          total += item.price * item.quantity;
          count += item.quantity;
        }
      });
    });
    return { total, count };
  }, [groups]);

  const allSelected = useMemo(
    () => groups.length > 0 && groups.every((group) => group.items.every((item) => item.selected)),
    [groups]
  );

  function goCheckout() {
    const selectedGroups = groups
      .map((group) => ({
        storeId: group.storeId,
        storeName: group.storeName,
        items: group.items.filter((item) => item.selected)
      }))
      .filter((group) => group.items.length > 0);
    const firstStore = selectedGroups[0];
    try {
      Taro.setStorageSync('cart_selection', JSON.stringify({
        storeId: firstStore?.storeId,
        storeName: firstStore?.storeName,
        items: selectedGroups.flatMap((group) => group.items),
        groups: selectedGroups
      }));
    } catch (_) {}
    const query = firstStore ? `?store_id=${encodeURIComponent(firstStore.storeId)}` : '';
    Taro.navigateTo({ url: `/pages/checkout/index${query}` }).catch(() => {});
  }

  return (
    <View className="page-cart">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">购物车</Text>
          </View>
          <Text className="nav-right">共 {summary.count} 件</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          {groups.map((group, groupIndex) => {
            const groupSelected = group.items.every((item) => item.selected);
            return (
              <View className="group-card" key={group.storeId}>
                <View className="group-header" onClick={() => toggleGroup(groupIndex)}>
                  <View className={`check ${groupSelected ? 'check--active' : ''}`}>
                    <Text>{groupSelected ? '✓' : ''}</Text>
                  </View>
                  <Text className="group-title">{group.storeName}</Text>
                  <Text className="group-sub">可自提</Text>
                </View>

                {group.items.map((item) => (
                  <View className="item-row" key={item.id}>
                    <View className={`check ${item.selected ? 'check--active' : ''}`} onClick={() => toggleItem(groupIndex, item.id)}>
                      <Text>{item.selected ? '✓' : ''}</Text>
                    </View>
                    <View className="item-main">
                      <Text className="item-name">{item.name}</Text>
                      <Text className="item-spec">{item.spec}</Text>
                      <View className="item-footer">
                        <Text className="item-price">¥ {item.price}</Text>
                        <View className="qty">
                          <View className="qty-btn" onClick={() => updateQty(groupIndex, item.id, -1)}>-</View>
                          <Text className="qty-value">{item.quantity}</Text>
                          <View className="qty-btn" onClick={() => updateQty(groupIndex, item.id, 1)}>+</View>
                        </View>
                      </View>
                    </View>
                  </View>
                ))}
              </View>
            );
          })}
        </ScrollView>

        <View className="bottom-bar">
          <View className="bottom-left" onClick={toggleAll}>
            <View className={`check ${allSelected ? 'check--active' : ''}`}>
              <Text>{allSelected ? '✓' : ''}</Text>
            </View>
            <Text className="bottom-label">全选</Text>
          </View>
          <View className="bottom-right">
            <Text className="bottom-total">合计：<Text className="bottom-amount">¥ {summary.total}</Text></Text>
            <View className="bottom-btn" onClick={goCheckout}>去结算</View>
          </View>
        </View>
      </View>
    </View>
  );
}
