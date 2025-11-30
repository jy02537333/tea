import React, { useEffect, useRef, useState } from 'react';
import { View, Image } from '@tarojs/components';

export interface TaroThumbnailProps {
  src?: string | null;
  width?: number;
  height?: number;
  radius?: number;
  skeleton?: boolean;
  shimmerHighlight?: string;
  shimmerBg?: string;
  shimmerSpeed?: number;
  loading?: 'lazy' | 'eager';
  animateScale?: boolean;
}

export default function Thumbnail({
  src,
  width = 40,
  height = 40,
  radius = 4,
  skeleton = true,
  shimmerHighlight = '#f5f5f5',
  shimmerBg = '#eaeaea',
  shimmerSpeed = 1.2,
  loading = 'lazy',
  animateScale = true,
}: TaroThumbnailProps) {
  const [loaded, setLoaded] = useState(false);
  const containerStyle: any = { width: `${width}px`, height: `${height}px`, borderRadius: `${radius}px`, marginRight: '6px', overflow: 'hidden' };
  const imgStyle: any = { width: '100%', height: '100%', opacity: loaded ? 1 : 0, transform: loaded ? 'scale(1)' : (animateScale ? 'scale(1.03)' : 'scale(1)'), transition: 'opacity 320ms ease-in-out, transform 320ms ease-in-out', backgroundColor: '#f0f0f0' };

  // css vars for global stylesheet - components can set root vars
  const cssVars: any = {
    ['--shimmer-duration']: `${shimmerSpeed}s`,
    ['--shimmer-bg']: shimmerBg,
    ['--shimmer-highlight']: shimmerHighlight,
  };

  return (
    <View style={{ ...containerStyle, ...cssVars }}>
      {skeleton ? <View className="thumbnail-skeleton" style={{ width: '100%', height: '100%' }} /> : null}
      {src ? <Image src={src} style={imgStyle} mode="aspectFill" lazyLoad={loading === 'lazy'} onLoad={() => setLoaded(true)} /> : null}
    </View>
  );
}
