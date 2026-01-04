import React, { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import { APP_NAME, APP_VERSION, COMPANY_NAME } from '../../constants/app';
import { getAboutContent } from '../../services/content';
import Taro from '@tarojs/taro';
import ContentRenderer from '../../components/ContentRenderer';

export default function AboutPage() {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    (async () => {
      try {
        const text = await getAboutContent();
        setContent(text);
      } catch (e: any) {
        // ignore, fallback to static info
        Taro.showToast({ title: e?.message || '加载失败', icon: 'none' });
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <View style={{ padding: 16 }}>
      <Text style={{ fontSize: 20, fontWeight: 'bold' }}>{APP_NAME}</Text>
      <Text style={{ display: 'block', marginTop: 8, color: '#666' }}>版本：{APP_VERSION}</Text>
      <Text style={{ display: 'block', marginTop: 8, color: '#666' }}>出品：{COMPANY_NAME}</Text>
      <View style={{ marginTop: 16 }}>
        {!loading && (
          <ContentRenderer
            markdown={
              content ||
              `我们专注于优质茶品与相关服务，欢迎提出建议与反馈以帮助我们持续改进。\n\n> 引用示例：我们坚持精选产区与稳定工艺，持续优化用户体验。\n\n内联代码示例：\`brew 95°C 120s\`。\n\n![示例海报](local:about-demo "示例图片标题")`
            }
          />
        )}
      </View>
    </View>
  );
}
