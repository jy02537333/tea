import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import { getHelpContent } from '../../services/content';
import Taro from '@tarojs/taro';
import ContentRenderer from '../../components/ContentRenderer';

export default function HelpPage() {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    (async () => {
      try {
        const text = await getHelpContent();
        setContent(text);
      } catch (e: any) {
        Taro.showToast({ title: e?.message || '加载失败', icon: 'none' });
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const fallback = `# 常见问题\n\n1. 如何查看订单？在「我的」→「订单」查看。\n\n2. 如何申请售后？在「我的」→「售后服务」中查看订单并操作。\n\n> 引用示例：如遇特殊问题，请优先查看帮助中心的最新公告。\n\n内联代码示例：复制订单号 \`NO123456\` 与客服沟通可提升处理效率。\n\n![操作示意](local:help-demo "帮助文档插图示例")\n\n3. 如何联系人工客服？在「我的」页面点击「联系客服」。\n\n4. 如何提交意见反馈？在「我的」→「意见反馈」提交工单。`;

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>帮助文档</Text>
      <View style={{ marginTop: 12 }}>
        {!loading && <ContentRenderer markdown={(content || fallback)} />}
      </View>
    </View>
  );
}
