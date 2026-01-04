import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import { getTermsContent } from '../../services/content';
import Taro from '@tarojs/taro';
import ContentRenderer from '../../components/ContentRenderer';

export default function TermsPage() {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    (async () => {
      try {
        const text = await getTermsContent();
        setContent(text);
      } catch (e: any) {
        Taro.showToast({ title: e?.message || '加载失败', icon: 'none' });
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const fallback = `# 用户协议\n\n本《用户协议》描述您与平台之间的权利与义务。\n\n> 引用示例：为保障交易公平，我们可能对异常行为进行限制或验证。\n\n内联代码示例：例如 \`refund\`、\`chargeback\` 等流程以平台规则为准。\n\n![协议示意图](local:terms-demo "协议插图示例")\n\n如需了解完整条款，请在后续版本中查看正式文本或联系平台客服。`;

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>用户协议</Text>
      <View style={{ marginTop: 12 }}>
        {!loading && <ContentRenderer markdown={(content || fallback)} />}
      </View>
    </View>
  );
}
