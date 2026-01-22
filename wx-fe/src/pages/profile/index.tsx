import React from 'react';
import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function ProfilePage() {
  return (
    <View className="page-profile">
      <View className="app">
        <View className="nav-bar">
          <Text className="nav-title">我的</Text>
          <Text className="nav-right">开通 VIP</Text>
        </View>

        <View className="scroll">
          <View className="user-card">
            <View className="user-top">
              <View className="user-main">
                <View className="user-avatar"><Text>心</Text></View>
                <View>
                  <Text className="user-name">茶心君</Text>
                  <Text className="user-sub">Lv3 · 云也·茶席常客</Text>
                </View>
              </View>
              <View className="user-tag">云也 · 会员卡</View>
            </View>
            <View className="user-grid">
              <View className="user-grid-item">
                <Text className="user-grid-label">可用积分</Text>
                <Text className="user-grid-value">1,280</Text>
                <Text className="user-grid-extra">去积分商城</Text>
              </View>
              <View className="user-grid-item">
                <Text className="user-grid-label">优惠券</Text>
                <Text className="user-grid-value">3 张</Text>
                <Text className="user-grid-extra">查看详情</Text>
              </View>
              <View className="user-grid-item" onClick={() => Taro.navigateTo({ url: '/pages/balance/index' }).catch(() => {})}>
                <Text className="user-grid-label">账户余额</Text>
                <Text className="user-grid-value">¥ 120</Text>
                <Text className="user-grid-extra">查看余额</Text>
              </View>
            </View>
            <Text className="vip-level-line">会员等级：普通用户、畅饮VIP、黄金VIP、初/中/高级合伙人</Text>
          </View>

          <View className="entry-section">
            <Text className="entry-section-title">常用功能</Text>
            <View className="entry-grid">
              <View className="entry-item" onClick={() => Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {})}>
                <View className="entry-icon-circle"><Text>单</Text></View>
                <Text>我的订单</Text>
                <Text className="entry-badge">近 3 笔</Text>
              </View>
              <View className="entry-item">
                <View className="entry-icon-circle"><Text>券</Text></View>
                <Text>我的卡券</Text>
                <Text className="entry-badge">3 张可用</Text>
              </View>
              <View className="entry-item" onClick={() => Taro.navigateTo({ url: '/pages/tea-reservation/index' }).catch(() => {})}>
                <View className="entry-icon-circle"><Text>约</Text></View>
                <Text>茶间预约</Text>
                <Text className="entry-badge">预约茶席</Text>
              </View>
              <View className="entry-item">
                <View className="entry-icon-circle"><Text>址</Text></View>
                <Text>地址管理</Text>
                <Text className="entry-badge">2 个</Text>
              </View>
              <View className="entry-item" onClick={() => Taro.navigateTo({ url: '/pages/recharge/index' }).catch(() => {})}>
                <View className="entry-icon-circle"><Text>充</Text></View>
                <Text>充值中心</Text>
              </View>
              <View className="entry-item" onClick={() => Taro.navigateTo({ url: '/pages/membership/index' }).catch(() => {})}>
                <View className="entry-icon-circle"><Text>V</Text></View>
                <Text>开通VIP</Text>
              </View>
              <View className="entry-item">
                <View className="entry-icon-circle"><Text>购</Text></View>
                <Text>购物车</Text>
              </View>
              <View className="entry-item">
                <View className="entry-icon-circle"><Text>客</Text></View>
                <Text>客服与帮助</Text>
              </View>
            </View>
          </View>

          <View className="list-card">
            <View className="list-row">
              <View>
                <Text className="list-label">会员权益说明</Text>
                <Text className="list-desc">了解不同等级对应的茶席权益</Text>
              </View>
              <View className="list-right"><Text>查看详情 ›</Text></View>
            </View>
            <View className="list-row">
              <View>
                <Text className="list-label">发票与记录</Text>
                <Text className="list-desc">查看历史发票与消费记录</Text>
              </View>
              <View className="list-right"><Text>去查看 ›</Text></View>
            </View>
            <View className="list-row">
              <View>
                <Text className="list-label">关于茶心阁</Text>
                <Text className="list-desc">品牌故事 · 门店信息</Text>
              </View>
              <View className="list-right"><Text>了解更多 ›</Text></View>
            </View>
          </View>
        </View>

        <View className="footer">
          <Text>茶心阁 · 小程序「我的」页面静态稿</Text>
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
            <View className="tab-bar__link">
              <View className="tab-bar__icon"><Text>👤</Text></View>
              <Text className="tab-bar__label">我的</Text>
            </View>
          </View>
        </View>
      </View>
    </View>
  );
}
