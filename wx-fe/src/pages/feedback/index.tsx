import React, { useState } from 'react';
import { View, Text, Input, Textarea, Picker, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { createFeedback } from '../../services/feedback';

export default function FeedbackPage() {
  const categories = ['咨询', '投诉', '建议'];
  const [categoryIndex, setCategoryIndex] = useState(0);
  const [subject, setSubject] = useState('');
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    if (!subject.trim() || !content.trim()) {
      Taro.showToast({ title: '请填写标题与内容', icon: 'none' });
      return;
    }
    setSubmitting(true);
    try {
      const cat = categoryIndex === 0 ? 'consult' : categoryIndex === 1 ? 'complaint' : 'suggest';
      const resp = await createFeedback({ category: cat as any, subject: subject.trim(), content: content.trim() });
      const ticketId = (resp as any)?.id;
      if (ticketId) {
        try {
          await Taro.showModal({ title: '提交成功', content: `工单编号：#${ticketId}`, showCancel: false, confirmText: '好的' });
        } catch (_) {}
      } else {
        Taro.showToast({ title: '提交成功', icon: 'success' });
      }
      setTimeout(() => {
        try { Taro.navigateBack({ delta: 1 }); } catch (_) {}
      }, 500);
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '提交失败', icon: 'none' });
    } finally {
      setSubmitting(false);
    }
  }

  function onCategoryChange(e: any) {
    const idx = Number(e?.detail?.value ?? 0);
    setCategoryIndex(idx);
  }

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>意见反馈</Text>
      <View style={{ marginTop: 12 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>类型</Text>
        <Picker mode="selector" range={categories} value={categoryIndex} onChange={onCategoryChange}>
          <View style={{ marginTop: 6, padding: 10, borderWidth: 1, borderStyle: 'solid', borderColor: '#ddd', borderRadius: 6 }}>
            <Text>{categories[categoryIndex]}</Text>
          </View>
        </Picker>
      </View>

      <View style={{ marginTop: 12 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>标题</Text>
        <Input
          type="text"
          placeholder="请简要描述问题或建议"
          value={subject}
          onInput={(e) => setSubject(String((e.detail as any)?.value || ''))}
        />
      </View>

      <View style={{ marginTop: 12 }}>
        <Text style={{ fontSize: 14, color: '#666' }}>内容</Text>
        <Textarea
          style={{ minHeight: '120px' }}
          placeholder="请详细描述问题、复现步骤或建议"
          value={content}
          onInput={(e) => setContent(String((e.detail as any)?.value || ''))}
        />
        <Text style={{ display: 'block', marginTop: 6, color: '#999' }}>为提高效率，请尽量提供相关订单号或页面截图说明（后续版本将支持图片上传）。</Text>
      </View>

      <View style={{ marginTop: 16 }}>
        <Button type="primary" loading={submitting} disabled={submitting} onClick={handleSubmit}>提交反馈</Button>
      </View>
    </View>
  );
}
