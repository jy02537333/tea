import React from 'react';
import { View, Text, ScrollView } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function MembershipPage() {
  return (
    <View className="page-membership">
      <View className="app">
        <View className="nav-bar">
          <Text className="nav-title">VIP等级</Text>
          <Text className="nav-right">云也 · 会员中心</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <Text className="section-title">权益说明</Text>
          <View className="benefit-card">
            <View className="benefit-row">
              <View>
                <Text className="benefit-main">茶席服务费优惠</Text>
                <Text className="benefit-desc">到店享用茶席时，服务费 9 折</Text>
              </View>
              <Text className="benefit-right">Lv2 及以上</Text>
            </View>
            <View className="benefit-row">
              <View>
                <Text className="benefit-main">生日专属礼遇</Text>
                <Text className="benefit-desc">生日当月获赠生日茶一杯</Text>
              </View>
              <Text className="benefit-right">Lv3 及以上</Text>
            </View>
            <View className="benefit-row">
              <View>
                <Text className="benefit-main">专属会员日</Text>
                <Text className="benefit-desc">每月 18 日会员日，指定茶品买一送一</Text>
              </View>
              <Text className="benefit-right">Lv1 及以上</Text>
            </View>
          </View>

          <Text className="section-title">等级权益内容</Text>
          <View className="benefit-card">
            <View className="benefit-row">
              <Text className="benefit-main">普通用户</Text>
              <Text className="benefit-right">基础权益</Text>
            </View>
            <View className="benefit-row">
              <Text className="benefit-main">畅饮VIP</Text>
              <Text className="benefit-right">会员权益</Text>
            </View>
            <View className="benefit-row">
              <Text className="benefit-main">黄金VIP</Text>
              <Text className="benefit-right">会员权益</Text>
            </View>
            <View className="benefit-row">
              <Text className="benefit-main">初合伙人</Text>
              <Text className="benefit-right">合伙人权益</Text>
            </View>
            <View className="benefit-row">
              <Text className="benefit-main">中合伙人</Text>
              <Text className="benefit-right">合伙人权益</Text>
            </View>
            <View className="benefit-row">
              <Text className="benefit-main">高级合伙人</Text>
              <Text className="benefit-right">合伙人权益</Text>
            </View>
          </View>

          <View className="rule-card">
            <Text className="rule-item"><Text className="rule-label">升级规则：</Text>等待运营配置落地，当前仅展示权益说明。</Text>
          </View>
        </ScrollView>

        <View className="footer">
          <Text>茶心阁 · 小程序「VIP等级」页面静态稿</Text>
          <Text className="footer-sub">仅用于样式联调与布局对齐示意</Text>
        </View>

        <View className="tab-bar">
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/home/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🏠</Text></View>
              <Text className="tab-bar__label">首页</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/menu/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🍵</Text></View>
              <Text className="tab-bar__label">点单</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/discover/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>🧭</Text></View>
              <Text className="tab-bar__label">发现</Text>
            </View>
          </View>
          <View className="tab-bar__item">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>📄</Text></View>
              <Text className="tab-bar__label">订单</Text>
            </View>
          </View>
          <View className="tab-bar__item tab-bar__item--active">
            <View className="tab-bar__link" onClick={() => Taro.navigateTo({ url: '/pages/profile/index' }).catch(() => {})}>
              <View className="tab-bar__icon"><Text>👤</Text></View>
              <Text className="tab-bar__label">我的</Text>
            </View>
          </View>
        </View>
      </View>
    </View>
  );
}
