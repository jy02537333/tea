import React, { useMemo, useState } from 'react';
import { View, Text, ScrollView, Textarea } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.scss';

export default function TeaReservationPage() {
  const timeSlots = ['本周六 · 19:00', '本周六 · 20:30', '本周日 · 14:30', '本周日 · 19:00'];
  const peopleOptions = ['1 人', '2 人', '3 人', '4 人'];
  const preferenceOptions = ['第一次来 · 店主搭配', '想喝清香乌龙', '想喝岩茶', '无酒精 · 仅茶饮'];

  const [activeTime, setActiveTime] = useState(timeSlots[0]);
  const [activePeople, setActivePeople] = useState(peopleOptions[0]);
  const [activePreference, setActivePreference] = useState(preferenceOptions[0]);
  const [remark, setRemark] = useState('可以简单写下此行的心情，例如：和朋友久别小聚 / 想安静读会儿书等…');
  const [statusText, setStatusText] = useState('');

  const summaryText = useMemo(
    () => `${activeTime.replace(' ·', '')} · ${activePeople} · ${activePreference}`,
    [activeTime, activePeople, activePreference]
  );

  function handleClearRemark() {
    setRemark('');
    setStatusText('已清空备注');
    Taro.showToast({ title: '已清空备注', icon: 'none' });
  }

  function handleSubmit() {
    setStatusText('已提交预约申请');
    Taro.showToast({ title: '已提交预约申请', icon: 'none' });
  }

  return (
    <View className="page-tea-reservation">
      <View className="app">
        <View className="nav-bar">
          <View className="nav-left">
            <View className="nav-back" onClick={() => Taro.navigateBack({ delta: 1 }).catch(() => {})}>‹</View>
            <Text className="nav-title">茶间预约</Text>
          </View>
          <Text className="nav-right">周末茶席 · 预留茶席</Text>
        </View>

        <ScrollView className="scroll" scrollY enhanced>
          <View className="tip-card">
            温馨提示：目前仅支持本店茶席预约，一客一茶，如需拼台或自带茶叶请提前在备注中说明～
          </View>

          <View className="field-card">
            <View className="field-label-row">
              <Text>选择日期</Text>
              <Text className="field-label-sub">默认展示未来 7 天可约场次</Text>
            </View>
            <View className="field-pill-row">
              {timeSlots.map((slot) => (
                <Text
                  key={slot}
                  className={`pill ${activeTime === slot ? 'pill--active' : ''}`}
                  onClick={() => setActiveTime(slot)}
                >
                  {slot}
                </Text>
              ))}
            </View>
          </View>

          <View className="field-card">
            <View className="field-label-row">
              <Text>预约人数</Text>
              <Text className="field-label-sub">单桌建议 1-4 人安静小坐</Text>
            </View>
            <View className="field-pill-row">
              {peopleOptions.map((people) => (
                <Text
                  key={people}
                  className={`pill ${activePeople === people ? 'pill--active' : ''}`}
                  onClick={() => setActivePeople(people)}
                >
                  {people}
                </Text>
              ))}
            </View>
          </View>

          <View className="field-card">
            <View className="field-label-row">
              <Text>茶席偏好</Text>
              <Text className="field-label-sub">可根据口味和场景为你搭配茶款</Text>
            </View>
            <View className="field-pill-row">
              {preferenceOptions.map((pref) => (
                <Text
                  key={pref}
                  className={`pill ${activePreference === pref ? 'pill--active' : ''}`}
                  onClick={() => setActivePreference(pref)}
                >
                  {pref}
                </Text>
              ))}
            </View>
            <Textarea
              className="remark-input"
              value={remark}
              onInput={(event) => setRemark(event.detail?.value ?? '')}
              maxlength={200}
              autoHeight
            />
            <View className="remark-actions">
              <View className="remark-btn" onClick={handleClearRemark}>清空</View>
              {statusText && <Text className="remark-status">{statusText}</Text>}
            </View>
          </View>
        </ScrollView>

        <View className="bottom-bar">
          <View className="bottom-info">
            <Text className="bottom-main">
              已选：<Text className="bottom-highlight">{summaryText}</Text>
            </Text>
            <Text className="bottom-main">具体茶款与茶席费用以到店沟通为准</Text>
          </View>
          <View className="bottom-btn" onClick={handleSubmit}>提交预约申请</View>
        </View>

        <View className="footer">
          <Text>茶心阁 · 小程序「茶间预约」页面静态稿</Text>
          <Text className="footer-sub">仅用于样式联调与布局对齐示意</Text>
        </View>
      </View>
    </View>
  );
}
