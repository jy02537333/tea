import React, { useState } from 'react';
import './thumbnail.css';

export interface ReactThumbnailProps {
  src?: string | null;
  alt?: string;
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
  alt = '',
  width = 40,
  height = 40,
  radius = 4,
  skeleton = true,
  shimmerHighlight = '#f5f5f5',
  shimmerBg = '#eaeaea',
  shimmerSpeed = 1.2,
  loading = 'lazy',
  animateScale = true,
}: ReactThumbnailProps) {
  const [loaded, setLoaded] = useState(false);
  const style: React.CSSProperties = { width, height, borderRadius: radius, overflow: 'hidden', marginRight: 6 };
  const cssVars: React.CSSProperties = {
    ['--af-shimmer-bg' as any]: shimmerBg,
    ['--af-shimmer-highlight' as any]: shimmerHighlight,
    ['--af-shimmer-duration' as any]: `${shimmerSpeed}s`,
  };

  return (
    <div style={{ ...style, ...cssVars }} className={`af-thumbnail-container ${loaded ? 'af-loaded' : ''}`}>
      {skeleton && <div className="af-thumbnail-skeleton" />}
      {src ? (
        <img
          src={src}
          alt={alt}
          style={{ width: '100%', height: '100%', objectFit: 'cover', opacity: loaded ? 1 : 0, transform: loaded ? 'scale(1)' : (animateScale ? 'scale(1.03)' : 'scale(1)'), transition: 'opacity 320ms ease, transform 320ms ease' }}
          loading={loading}
          onLoad={() => setLoaded(true)}
        />
      ) : null}
    </div>
  );
}
