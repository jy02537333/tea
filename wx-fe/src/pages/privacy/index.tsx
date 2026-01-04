import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import { getPrivacyContent } from '../../services/content';
import Taro from '@tarojs/taro';
import ContentRenderer from '../../components/ContentRenderer';

export default function PrivacyPage() {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    (async () => {
      try {
        const text = await getPrivacyContent();
        setContent(text);
      } catch (e: any) {
        Taro.showToast({ title: e?.message || '加载失败', icon: 'none' });
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const fallback = `# 隐私政策\n\n我们重视您的隐私与个人信息保护。为提供服务，我们会在遵循最小必要原则的前提下收集与使用相关信息，并采取合理的安全措施进行保护。\n\n> 引用示例：我们不会出售您的个人信息，并严格限定数据访问权限。\n\n内联代码示例：如 \`cookie\`、\`session\` 等仅用于必要的身份识别与安全用途。\n\n![示例图片](local:privacy-demo "隐私政策插图示例")\n\n如需了解完整条款，请联系平台客服或查看完整版《隐私政策》。`;

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>隐私政策</Text>
      <View style={{ marginTop: 12 }}>
        {!loading && <ContentRenderer markdown={(content || fallback)} />}
      </View>
    </View>
  );
}
