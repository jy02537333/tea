import React, { useEffect, useRef, useState } from 'react';
import './thumbnail.css';

interface Props {
  src?: string | null;
  alt?: string;
  width?: number;
  height?: number;
  radius?: number;
  skeleton?: boolean;
  shimmerHighlight?: string;
  shimmerBg?: string;
  shimmerSpeed?: number;
  /** 优先使用浏览器原生 lazy loading（H5） */
  loading?: 'lazy' | 'eager';
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
}: Props) {
  const [loaded, setLoaded] = useState(false);
  const [shouldLoad, setShouldLoad] = useState(!!src && loading !== 'lazy');
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!src) return;
    if (shouldLoad) return;
    if (typeof window === 'undefined') return;
    // IntersectionObserver fallback for lazy loading
    const io = new IntersectionObserver((entries) => {
      entries.forEach(e => {
        if (e.isIntersecting) {
          setShouldLoad(true);
          io.disconnect();
        }
      });
    }, { rootMargin: '200px' });
    if (ref.current) io.observe(ref.current);
    return () => io.disconnect();
  }, [src, shouldLoad]);

  const style: React.CSSProperties = { width, height, borderRadius: radius, overflow: 'hidden', marginRight: 6 };
  const cssVars: React.CSSProperties = {
    ['--shimmer-bg' as any]: shimmerBg,
    ['--shimmer-highlight' as any]: shimmerHighlight,
    ['--shimmer-duration' as any]: `${shimmerSpeed}s`,
  };

  return (
    <div ref={ref} style={{ ...style, ...cssVars }} className="af-thumbnail-container">
      {skeleton && !loaded && <div className="af-thumbnail-skeleton" />}
      {shouldLoad && src ? (
        <img
          src={src}
          alt={alt}
          style={{ width: '100%', height: '100%', objectFit: 'cover', opacity: loaded ? 1 : 0, transform: loaded ? 'scale(1)' : 'scale(1.03)', transition: 'opacity 320ms ease, transform 320ms ease' }}
          loading={loading}
          onLoad={() => setLoaded(true)}
        />
      ) : null}
    </div>
  );
}
