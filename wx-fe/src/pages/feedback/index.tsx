import React, { useState } from 'react';
import { Button, Textarea, View, Input, Picker } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { createTicket } from '../../services/tickets';

const TYPES = [
  { label: '意见反馈', value: 'consult' },
  { label: '订单问题', value: 'order' },
  { label: '投诉建议', value: 'complaint' },
];

export default function FeedbackPage() {
  const router = useRouter();
  const orderIdParam = router.params.orderId ? Number(router.params.orderId) : undefined;

  const [typeIndex, setTypeIndex] = useState(0);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    if (!title.trim() || !content.trim()) {
      Taro.showToast({ title: '请填写标题和内容', icon: 'none' });
      return;
    }
    const type = TYPES[typeIndex]?.value || 'consult';
    setSubmitting(true);
    try {
      await createTicket({
        type,
        source: orderIdParam ? 'miniapp_order' : 'miniapp_feedback',
        order_id: orderIdParam,
        title: title.trim(),
        content: content.trim(),
      });
      Taro.showToast({ title: '已提交，我们会尽快处理', icon: 'none' });
      setTimeout(() => {
        Taro.navigateBack().catch(() => {
          Taro.switchTab({ url: '/pages/profile/index' }).catch(() => {});
        });
      }, 500);
    } catch (error: any) {
      Taro.showToast({ title: error?.message || '提交失败，请稍后重试', icon: 'none' });
    } finally {
      setSubmitting(false);
    }
  }

  function handleTypeChange(e: any) {
    const idx = Number(e.detail.value || 0);
    setTypeIndex(idx);
  }

  return (
    <View style={{ padding: 16 }}>
      <View style={{ marginBottom: 12 }}>
        <View>问题类型</View>
        <Picker mode="selector" range={TYPES} rangeKey="label" onChange={handleTypeChange} value={typeIndex}>
          <View
            style={{
              padding: 12,
              borderRadius: 8,
              border: '1px solid #ddd',
              marginTop: 8,
            }}
          >
            {TYPES[typeIndex]?.label}
          </View>
        </Picker>
      </View>

      <View style={{ marginBottom: 12 }}>
        <View>标题</View>
        <Input
          style={{
            marginTop: 8,
            padding: 8,
            borderRadius: 8,
            border: '1px solid #ddd',
          }}
          placeholder="请简单描述您的问题"
          value={title}
          onInput={(e) => setTitle(e.detail.value)}
        />
      </View>

      <View style={{ marginBottom: 12 }}>
        <View>详细描述</View>
        <Textarea
          style={{
            marginTop: 8,
            minHeight: 120,
            padding: 8,
            borderRadius: 8,
            border: '1px solid #ddd',
          }}
          placeholder="请详细描述您遇到的问题或建议"
          value={content}
          onInput={(e) => setContent(e.detail.value)}
        />
      </View>

      <Button type="primary" onClick={handleSubmit} loading={submitting} disabled={submitting}>
        提交
      </Button>
    </View>
  );
}
