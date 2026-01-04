import React, { useMemo } from 'react';
import { View, RichText, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { markdownToHtml } from '../utils/markdown';
import { sanitizeHtml } from '../utils/sanitize';
import '../styles/markdown.scss';

export default function ContentRenderer({ markdown, html }: { markdown?: string; html?: string }) {
  const computedHtml = useMemo(() => {
    if (html && html.trim().length > 0) return html;
    if (markdown && markdown.trim().length > 0) return markdownToHtml(markdown);
    return '';
  }, [markdown, html]);

  const safeHtml = useMemo(() => sanitizeHtml(computedHtml), [computedHtml]);

  const isH5 = (Taro as any).getEnv && (Taro as any).getEnv() === 'WEB';

  if (!computedHtml) {
    return <Text style={{ color: '#999' }}>暂无内容</Text>;
  }

  if (isH5) {
    return <View className="markdown-content" dangerouslySetInnerHTML={{ __html: safeHtml }} />;
  }
  return <RichText nodes={safeHtml} />;
}
