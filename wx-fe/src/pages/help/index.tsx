import React from 'react';
import { View, Text } from '@tarojs/components';

export default function HelpPage() {
  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>帮助文档</Text>
      <View style={{ marginTop: 12 }}>
        <Text style={{ display: 'block', marginBottom: 8 }}>常见问题</Text>
        <Text style={{ display: 'block', color: '#666', marginBottom: 6 }}>
          1. 如何查看订单？在「我的」→「订单」查看。
        </Text>
        <Text style={{ display: 'block', color: '#666', marginBottom: 6 }}>
          2. 如何申请售后？在「我的」→「售后服务」中查看订单并操作。
        </Text>
        <Text style={{ display: 'block', color: '#666', marginBottom: 6 }}>
          3. 如何联系人工客服？在「我的」页面点击「联系客服」。
        </Text>
        <Text style={{ display: 'block', color: '#666' }}>
          4. 如何提交意见反馈？在「我的」→「意见反馈」提交工单。
        </Text>
      </View>
    </View>
  );
}
