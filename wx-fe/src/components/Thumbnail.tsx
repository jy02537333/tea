import { Image, View } from '@tarojs/components';
import React, { useState, CSSProperties } from 'react';

interface Props {
  src?: string | null;
  width?: number;
  height?: number;
  radius?: number;
  lazyLoad?: boolean;
  /** 是否在加载时显示骨架（灰色占位），默认为 true */
  skeleton?: boolean;
  /** shimmer 动画高亮颜色 */
  shimmerHighlight?: string;
  /** shimmer 动画基底颜色 */
  shimmerBg?: string;
  /** shimmer 动画时长（秒），默认为 1.2 */
  shimmerSpeed?: number;
  /** 在 H5 上可选传递原生 loading 属性值（'lazy'|'eager'） */
  loading?: 'lazy' | 'eager';
}

export default function Thumbnail({
  src,
  width = 40,
  height = 40,
  radius = 4,
  lazyLoad = true,
  skeleton = true,
  shimmerHighlight = '#f5f5f5',
  shimmerBg = '#eaeaea',
  shimmerSpeed = 1.2,
  loading = 'lazy',
}: Props) {
  const [loaded, setLoaded] = useState(false);
  const containerStyle: CSSProperties = { width: `${width}px`, height: `${height}px`, borderRadius: `${radius}px`, marginRight: '6px', overflow: 'hidden' } as any;
  const imgStyle: CSSProperties = { width: '100%', height: '100%', opacity: loaded ? 1 : 0, transition: 'opacity 300ms ease-in-out', backgroundColor: '#f0f0f0' } as any;
  const skeletonStyle: CSSProperties = { width: '100%', height: '100%' } as any;

  const cssVars = {
    ['--shimmer-duration' as any]: `${shimmerSpeed}s`,
    ['--shimmer-bg' as any]: shimmerBg,
    ['--shimmer-highlight' as any]: shimmerHighlight,
  } as any;

  if (!src) {
    // 无图片时直接显示骨架占位
    return <View className={skeleton ? 'thumbnail-skeleton' : ''} style={{ ...containerStyle, ...(skeleton ? skeletonStyle : { backgroundColor: '#e5e5e5' }), ...cssVars }} />;
  }

  return (
    <View style={containerStyle}>
      {skeleton && !loaded ? <View className="thumbnail-skeleton" style={{ ...skeletonStyle, ...cssVars }} /> : null}
      <Image src={src} style={imgStyle} mode="aspectFill" lazyLoad={lazyLoad} onLoad={() => setLoaded(true)} {...({ loading } as any)} />
    </View>
  );
}
