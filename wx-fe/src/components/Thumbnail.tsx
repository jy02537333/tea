import { Image, View } from '@tarojs/components';
import React from 'react';

interface Props {
  src?: string | null;
  width?: number;
  height?: number;
  radius?: number;
  lazyLoad?: boolean;
}

export default function Thumbnail({ src, width = 40, height = 40, radius = 4, lazyLoad = true }: Props) {
  const style = { width: `${width}px`, height: `${height}px`, borderRadius: `${radius}px`, marginRight: '6px' } as any;
  if (!src) {
    return <View style={{ ...style, backgroundColor: '#e5e5e5' }} />;
  }
  return (
    <Image src={src} style={{ ...style, backgroundColor: '#f0f0f0' }} mode="aspectFill" lazyLoad={lazyLoad} />
  );
}
